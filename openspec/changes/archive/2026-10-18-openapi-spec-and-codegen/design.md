# Design

## Context

Мотивация — см. proposal.md, требования — `specs/rest-api/spec.md` и `specs/api-codegen/spec.md`.

Текущее состояние:
- `api/openapi.yaml` — заглушка: `info` и `paths: {}`.
- `internal/transport/http/` пуст. chi и oapi-codegen в `go.mod` не подключены. Единственный tool — `go.uber.org/mock/mockgen`, моки генерируются директивами `//go:generate go tool mockgen` в `deps.go`.
- Домен: `entity.Category` — коды `food` … `other`, `entity.Expense` с `UpdatedAt *time.Time`, `entity.ExpenseFilter` с `From` включительно и `To` не включительно. Sentinel-ошибки: `ErrNotFound`, `ErrUnknownUser`, `ErrInvalidAmount`, `ErrInvalidCategory`.
- Репозиторий умеет `ListUserIDs`, но не отдаёт профили членов семьи.
- `Dockerfile` собирает `go build ./cmd/app` из исходников и генератор не вызывает. Это и требуется.

## Goals / Non-Goals

**Goals:**
- Контракт однозначно соответствует существующему домену: те же коды категорий, та же семантика периода, сумма в копейках `int64`. Хендлерам задачи 05 не нужно ничего переводить, кроме типов.
- Имена методов сгенерированного `StrictServerInterface` следуют правилу проекта «действие + сущность».
- Генерация воспроизводима и не требует ничего, кроме Go из `go.mod`.

**Non-Goals:**
- Реализация `StrictServerInterface`, маппинг доменных ошибок в коды `Error.code`, обработчики ошибок разбора запроса, подключение роутера в `main` — задача 05.
- Middleware валидации запросов по спецификации (`kin-openapi` request validator). Проверки суммы и категории живут в `usecase` (правило `AGENTS.md`), дублировать их на транспорте не нужно.
- Встраивание спецификации в бинарь (`embedded-spec`), Swagger UI.
- Генерация TS-типов — задача 11.

## Decisions

### Ресурсы и операции

`servers: [{url: /api}]`, пути без префикса: `/expenses`, `/expenses/{id}`, `/members`. Префикс `/api` задаётся при монтировании (`HandlerWithOptions` с `BaseURL`) в задаче 05, а nginx фронтенда проксирует `/api` на `app`. Так пути в спецификации не дублируют префикс, а клиент на TS получает тот же `baseUrl`.

`operationId`: `listExpenses`, `createExpense`, `getExpense`, `updateExpense`, `deleteExpense`, `listMembers`. oapi-codegen делает из них методы `ListExpenses`, `CreateExpense`, … — они совпадают с правилом именования и с методами репозитория.

### Фильтр периода — `date-time`, `to` не включительно

Альтернатива — `format: date` с обеими границами включительно. Тогда сервер должен решать, в каком часовом поясе начинается и заканчивается день. Часового пояса семьи в конфиге нет, а угадывать его по серверу — источник ошибок на границах суток. С `date-time` фронтенд сам переводит выбранные даты в моменты в локальном поясе браузера (`from` = начало первого дня, `to` = начало дня после последнего), а семантика один в один совпадает с `entity.ExpenseFilter`.

### Изменение — `PUT` с полной заменой, а не `PATCH`

Форма редактирования в панели всегда отправляет все поля. `repository.UpdateExpense` тоже заменяет все изменяемые поля. `PATCH` потребовал бы различать «поле не передано» и «передан `null`», то есть `nullable`-обёрток в сгенерированных моделях и merge в коде. `ExpenseUpdate`: `amount`, `category`, `spent_at` обязательны, `comment` необязателен (нет → пустая строка). `user_id` в `ExpenseUpdate` нет: переносить трату на другого члена семьи панель не умеет, и репозиторий `user_id` при изменении не трогает.

### Схемы

- `Category` — `type: string`, `enum` из семи кодов, в `description` — русские названия. Генерируется `openapi.Category` с константами. Она отдельна от `entity.Category`, конвертация — в задаче 05.
- `Expense` — все поля `required`, `updated_at` — `nullable: true` (генерируется `*time.Time`, совпадает с entity). `comment` — обязательная строка, может быть пустой: клиенту не нужно различать «нет поля» и `""`.
- `ExpenseCreate` — `required: [user_id, amount, category]`, `spent_at` и `comment` необязательны (генерируются указателями). «По умолчанию — сейчас» фиксируется в `description` поля и реализуется в задаче 05.
- `amount` — `type: integer, format: int64, minimum: 1` везде, где поле встречается во входных данных. `minimum` документирует контракт и попадёт в TS-типы. На сервере его применяет `usecase` (`ErrInvalidAmount`). Дробное число не декодируется в `int64`, и это ошибка разбора тела, т.е. 400.
- `Member` — `telegram_id` (int64), `username`, `first_name`, `last_name` — обязательные строки. Совпадает с `entity.User` и колонками `NOT NULL DEFAULT ''`.
- `Error` — `code` (`ErrorCode`: `invalid_request`, `invalid_amount`, `invalid_category`, `unknown_member`, `not_found`, `internal`) и `message`. Коды соответствуют sentinel-ошибкам `entity` плюс ошибке разбора запроса и внутренней ошибке. Это даёт панели стабильные ключи для сообщений вместо разбора текста.
- Ответы 400/404/500 вынесены в `components/responses` (`BadRequest`, `NotFound`, `InternalError`) со схемой `Error`. В strict-режиме oapi-codegen генерирует общие типы ответов, и операции на них ссылаются. Так формат ошибки один на весь API.

### Раскладка генерации

```
internal/transport/http/openapi/
  generate.go            package openapi + //go:generate go tool oapi-codegen -config oapi-codegen.yaml ../../../../api/openapi.yaml
  oapi-codegen.yaml      package: openapi; generate: models, chi-server, strict-server; output: openapi.gen.go
  openapi.gen.go         сгенерировано
  spec_test.go           валидация api/openapi.yaml
```

- Отдельный пакет `openapi`, а не код прямо в `internal/transport/http`. Сгенерированные типы (`Expense`, `Category`, `Error`) не смешиваются с хендлерами и не конфликтуют с их именами. Задача 05 импортирует пакет и реализует `openapi.StrictServerInterface`.
- Файл с директивой называется `generate.go`. `deps.go` здесь не подходит: зависимостей-интерфейсов у пакета нет. Правило «нет комментариев» соблюдается, потому что `//go:generate` — разрешённое исключение.
- oapi-codegen подключается через `go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen`, как и mockgen. Версия фиксируется в `go.mod`. Строка версии в шапке `.gen.go` поэтому стабильна, и повторный запуск даёт тот же файл.
- Runtime: `github.com/oapi-codegen/runtime` (разбор параметров) и `github.com/go-chi/chi/v5` (сгенерированный роутер) становятся прямыми зависимостями после `go mod tidy`.
- В `Makefile` — цель `generate: go generate ./...` и запись в `.PHONY`. Она запускает и моки, и OpenAPI: одна команда на всю генерацию.

### Валидация спецификации — Go-тест на kin-openapi

`spec_test.go` (пакет `openapi_test`, testify suite, кейс `TestSpec_Valid`) загружает `../../../../api/openapi.yaml` через `openapi3.NewLoader().LoadFromFile` и вызывает `doc.Validate(ctx)`.

Почему не `npx @redocly/cli lint` или `swagger-cli`: для них нужен Node и сеть, их нет в `go test ./...`, и про них легко забыть. `github.com/getkin/kin-openapi` — тот же парсер, которым пользуется oapi-codegen, он уже есть в графе модулей как зависимость генератора. Тест лишь делает его прямой зависимостью. Это новая прямая библиотека сверх списка в `AGENTS.md`, причина — эта проверка.

## Risks / Trade-offs

- [Strict-сервер не проверяет `enum` и `minimum` в теле и query] → проверка в `usecase` (`ErrInvalidCategory`, `ErrInvalidAmount`) с маппингом в 400 в задаче 05. Контракт для клиента от этого не меняется.
- [`GET /members` обещает имена, репозиторий отдаёт только ID] → чтение профилей (`ListUsers` или аналог) добавляется в задаче 05/06. До этого реализация эндпоинта невозможна, а контракт уже зафиксирован.
- [Telegram ID и `id` — int64, в JS `number` точен до 2^53] → реальные Telegram ID и identity-ключи далеко от этого предела. Строковое представление усложнило бы и контракт, и фильтры.
- [Обновление oapi-codegen меняет сгенерированный код] → версия зафиксирована в `go.mod`. Обновление — отдельный осознанный шаг с перегенерацией и коммитом.
- [Пакет `internal/transport/http` в будущем называется так же, как `net/http`] → это решение задачи 05, сгенерированный пакет `openapi` от него не зависит.

## Migration Plan

Не требуется: API ещё нет, клиентов нет, сервер в `main` не подключается.
