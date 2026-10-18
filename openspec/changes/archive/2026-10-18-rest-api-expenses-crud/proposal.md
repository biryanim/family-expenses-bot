# Proposal

## Why

Контракт REST API (`api/openapi.yaml`) и сгенерированный strict-интерфейс сервера уже есть, но их никто не реализует: HTTP-сервер не поднимается, слоя `usecase` фактически нет (только `deps.go`), а доменные проверки суммы, категории и даты нигде не выполняются до базы. Без этого веб-панели не с чем работать, а боту (следующие задачи) — не на что опереться: правила должны жить в `usecase` и быть общими для REST и бота.

## What Changes

- Слой `usecase` для трат и членов семьи: создание, получение, список с фильтрами, изменение, удаление трат, список членов семьи. Валидация входных данных здесь: сумма > 0, категория из фиксированного списка, дата траты задана; при создании без даты — текущий момент.
- Новая доменная ошибка «некорректная дата» (`entity.ErrInvalidDate`): пустая `spent_at` при изменении и период фильтра, где начало позже конца.
- Реализация `StrictServerInterface` в `internal/transport/http`: хендлеры вызывают только `usecase`, переводят типы `openapi` ↔ `entity` и маппят ошибки в ответы спецификации (`not_found` → 404, доменные → 400, прочее → 500 без деталей).
- Ошибки разбора запроса (некорректные query/path-параметры, невалидный JSON, дробная сумма) отдаются в едином формате `{"code":"invalid_request","message":…}` со статусом 400 вместо текстовых ответов генератора по умолчанию.
- Роутер chi с префиксом `/api`, CORS для фронтенда, HTTP-сервер в `cmd/app` с graceful shutdown по `SIGINT`/`SIGTERM`.
- Новая необязательная переменная окружения `CORS_ALLOWED_ORIGINS` (по умолчанию `http://localhost:5173` — dev-сервер Vite).
- Контракт: в `ErrorCode` добавляется код `invalid_date` (правка `api/openapi.yaml` → `go generate` → код).
- Репозиторий: `ListUserIDs` заменяется на `ListUsers`, который отдаёт профили членов семьи — их требует `GET /members`. **BREAKING** для внутреннего интерфейса `usecase.ExpenseRepository` (внешних потребителей нет).

## Capabilities

### New Capabilities
- `expense-rules`: бизнес-правила трат, общие для REST API и бота: валидация суммы, категории и даты, дата по умолчанию, проверка периода фильтра, различимые доменные ошибки.

### Modified Capabilities
- `rest-api`: маппинг доменных ошибок в коды ответов, код `invalid_date`, единый формат для ошибок разбора запроса, CORS.
- `app-bootstrap`: приложение поднимает HTTP-сервер на `HTTP_PORT` и корректно его останавливает; новая переменная `CORS_ALLOWED_ORIGINS`.
- `expense-repository`: список членов семьи возвращает профили (`telegram_id`, `username`, `first_name`, `last_name`), а не только идентификаторы.

## Impact

- Код: `internal/usecase/*` (новые методы и тесты), `internal/entity/errors.go`, `internal/transport/http/*` (хендлеры, роутер, обработчики ошибок, тесты), `internal/database/repository/list_users.go` (вместо `list_user_ids.go`), `internal/config`, `cmd/app/main.go`, `mock/*`.
- API: `api/openapi.yaml` (+`invalid_date` в `ErrorCode`), перегенерация `openapi.gen.go`. Пути, параметры и схемы не меняются.
- Зависимости: `github.com/go-chi/cors` (обоснование в design.md).
- Конфиг: `CORS_ALLOWED_ORIGINS` в `.env.example`.
- Не входит: бот и per-user методы `usecase` для бота, фронтенд и nginx, авторизация (по ТЗ её нет).
