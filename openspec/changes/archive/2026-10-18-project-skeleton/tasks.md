# Tasks

## 1. Структура и служебные файлы

- [x] 1.1 Создать папки `cmd/app`, `api`, `migrations`, `internal/{entity,usecase,repository,transport/http,transport/telegram,agent,llm,config}`, `web`; в папки, которые останутся без файлов после этого изменения, положить `.gitkeep`. Проверка: `find . -path ./.git -prune -o -type d -print` совпадает со схемой из `AGENTS.md`.
- [x] 1.2 Добавить `.gitignore` (`.env`, бинарь `/app` и `bin/`, `web/node_modules`, `web/dist`, файлы IDE/OS). Проверка: после создания тестового `.env` команда `git check-ignore .env` его выводит.
- [x] 1.3 Добавить `.env.example` со всеми переменными из спеки (`TELEGRAM_BOT_TOKEN`, `LLM_BASE_URL`, `LLM_API_KEY`, `LLM_MODEL`, `DATABASE_URL` с хостом `postgres`, `HTTP_PORT=8080`, `LOG_LEVEL=info`) и плейсхолдерами вместо секретов, с короткими комментариями. Проверка: файл в git, реальных ключей нет.
- [x] 1.4 Добавить `api/openapi.yaml` — OpenAPI 3.0 с `info` (название, версия) и `paths: {}`. Проверка: файл парсится как YAML и содержит `openapi: 3.0.x`.

## 2. Конфиг

- [x] 2.1 Реализовать `internal/config`: структура `Config` с группами `Telegram`, `LLM`, `DB`, `HTTP`, `Log`; `Load()` поверх `load(getenv)`; обязательные значения с `strings.TrimSpace`, сбор всех ошибок через `errors.Join` с перечислением незаданных переменных; дефолты `HTTP_PORT=8080`, `LOG_LEVEL=info`; валидация порта (1–65535) и уровня (`debug|info|warn|error`, без учёта регистра). Проверка: `go build ./...`.
- [x] 2.2 Реализовать `Config.LogValue()`, отдающий только несекретные поля (порт, уровень, `LLM_BASE_URL`, `LLM_MODEL`). Проверка: тест из 2.3 на отсутствие секретов.
- [x] 2.3 Table-driven тесты `internal/config/config_test.go`: все значения заданы; дефолты порта и уровня; одна незаданная переменная; несколько незаданных — все в одном сообщении; пустая строка и строка из пробелов считаются незаданными; некорректный `HTTP_PORT` (`abc`, `0`, `70000`) и `LOG_LEVEL` (`verbose`) — сообщение содержит имя и значение; уровень без учёта регистра (`WARN`); вывод `slog` с `Config` не содержит значений токена, ключа и `DATABASE_URL`. Проверка: `go test ./internal/config/...` зелёный.

## 3. Точка входа и логирование

- [x] 3.1 Реализовать `cmd/app/main.go`: `main` загружает конфиг, при ошибке пишет её в stderr и выходит с кодом 1; создаёт `slog` с `JSONHandler` в stdout и уровнем из конфига, ставит `slog.SetDefault`; вызывает `run(ctx, cfg, logger) error` с контекстом от `signal.NotifyContext(SIGINT, SIGTERM)`. `run` пишет запись о старте (с `cfg` через `LogValue`), ждёт `ctx.Done()`, пишет запись об остановке. Проверка: `go build ./...`, `go vet ./...`.
- [x] 3.2 Ручная проверка бинаря: (а) запуск без переменных — код выхода 1, в stderr перечислены все пять обязательных переменных; (б) запуск с заполненными переменными и `LOG_LEVEL=debug` — JSON-запись о старте в stdout без секретов; (в) `kill -TERM` — запись об остановке, код выхода 0.

## 4. Docker

- [x] 4.1 Добавить `Dockerfile`: stage сборки на `golang:<версия не ниже go.mod>` с `CGO_ENABLED=0 go build -o /app ./cmd/app`, финальный образ `gcr.io/distroless/static-debian12:nonroot`; перед этим убедиться, что выбранные теги существуют. Проверка: `docker build .` проходит.
- [x] 4.2 Добавить `docker-compose.yml`: `postgres` (`postgres:17-alpine`, именованный volume, healthcheck `pg_isready`, учётные данные из env) и `app` (`build: .`, `env_file: .env`, `depends_on` с `condition: service_healthy`, проброс порта `HTTP_PORT`). Проверка: `cp .env.example .env && docker compose up --build` — postgres healthy, в логах `app` JSON-запись о старте; `docker compose down` останавливает `app` с записью об остановке.

## 5. Итоговая проверка

- [x] 5.1 Прогнать `gofmt -l .` (пусто), `go vet ./...`, `go build ./...`, `go test ./...` — всё проходит; `git status` не показывает `.env` и бинарей; структура папок соответствует схеме из `AGENTS.md`.
