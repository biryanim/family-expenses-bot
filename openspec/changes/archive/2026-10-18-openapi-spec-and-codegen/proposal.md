# Proposal

## Why

Веб-панели (задача 11) и реализации REST API (задача 05) нужен общий контракт. Проект spec-first: сначала `api/openapi.yaml`, из него генерируются Go-интерфейс сервера и модели, потом пишется код. Сейчас `api/openapi.yaml` — заглушка с `paths: {}`, генератор не подключён, поэтому задачи 05 и 11 начать нельзя.

## What Changes

- `api/openapi.yaml` (OpenAPI 3.0.3) описывает API панели:
  - `GET /expenses` — список трат с необязательными фильтрами: член семьи (`user_id`), период по `spent_at` (`from` включительно, `to` не включительно, обе границы — `date-time`), категория;
  - `POST /expenses` — создание траты члена семьи;
  - `GET /expenses/{id}`, `PUT /expenses/{id}`, `DELETE /expenses/{id}` — получение, полное изменение и удаление траты;
  - `GET /members` — список членов семьи для фильтра панели (Telegram ID и отображаемые данные);
  - схемы `Expense`, `ExpenseCreate`, `ExpenseUpdate`, `Member`, `Category` (enum кодов `food`, `transport`, `housing`, `entertainment`, `health`, `shopping`, `other`, как в `entity.Category` и CHECK в БД), `Error`;
  - сумма — `integer` / `int64` в копейках, строго больше нуля;
  - единый формат ошибки `{code, message}` и ответы 400, 404, 500.
- Подключается `github.com/oapi-codegen/oapi-codegen/v2` как `tool` в `go.mod` и runtime `github.com/oapi-codegen/runtime`.
- Генерация: модели, chi-сервер и strict-интерфейс в пакет `internal/transport/http/openapi`, файл `openapi.gen.go`. Конфиг генератора лежит рядом, запуск — директивой `//go:generate`.
- Валидация спецификации — Go-тест, который загружает `api/openapi.yaml` и валидирует его через `kin-openapi` (эта библиотека уже нужна генератору). Тест запускается в `go test ./...`.
- В `Makefile` появляется цель `generate` (`go generate ./...`).
- Сгенерированный код коммитится: `docker build` не запускает генератор.

## Capabilities

### New Capabilities
- `rest-api`: контракт REST API веб-панели. Ресурсы и операции, фильтры списка трат, схемы данных, сумма в копейках, фиксированный enum категорий, единый формат ошибок и коды ответов.
- `api-codegen`: генерация серверного кода из спецификации. Валидность спецификации, воспроизводимость `go generate`, сборка со сгенерированным кодом, коммит сгенерированных файлов, запрет ручной правки.

### Modified Capabilities
Нет. Требования `expense-repository` и `expense-storage` не меняются: контракт на них опирается, но их не меняет.

## Impact

- Файлы: `api/openapi.yaml`, новый пакет `internal/transport/http/openapi/` (конфиг генератора, файл с `//go:generate`, `openapi.gen.go`, тест валидации спецификации), `Makefile`, `go.mod`/`go.sum`.
- Зависимости: `github.com/oapi-codegen/oapi-codegen/v2` (tool), `github.com/oapi-codegen/runtime`, `github.com/go-chi/chi/v5` (нужен сгенерированному коду; по `AGENTS.md` он и так HTTP-роутер проекта), `github.com/getkin/kin-openapi` (прямая зависимость теста валидации).
- Вне скоупа: реализация хендлеров, маппинг доменных ошибок в HTTP и подключение сервера в `main` (задача 05); генерация TS-типов фронтенда (задача 11); пагинация; авторизация.
- Известный разрыв: `GET /members` возвращает имена членов семьи, а репозиторий пока отдаёт только `ListUserIDs`. Чтение профилей добавляется в задаче 05/06, контракт этого не ждёт.
