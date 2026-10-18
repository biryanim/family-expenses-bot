# Tasks

## 1. Зависимости

- [x] 1.1 Оставить в `go.mod` только `github.com/jackc/pgx/v5` из новых модулей; goose и testcontainers как Go-модули не нужны. Проверка: после кода групп 2–5 `go mod tidy` оставляет `pgx/v5` и не содержит `pressly/goose` и `testcontainers`, `go build ./...` проходит.

## 2. Сущности

- [x] 2.1 Добавить `internal/entity/category.go`: тип `Category string`, семь констант (`food`, `transport`, `housing`, `entertainment`, `health`, `shopping`, `other`), `Categories() []Category` (новый слайс при каждом вызове) и `(Category) Valid() bool`. Удалить `internal/entity/.gitkeep`. Проверка: `go build ./...`.
- [x] 2.2 Добавить `internal/entity/expense.go` (`Expense`: `ID`, `UserID`, `Amount int64`, `Category`, `SpentAt`, `Comment`, `CreatedAt time.Time`, `UpdatedAt *time.Time`) и `internal/entity/user.go` (`User`: `TelegramID`, `Username`, `FirstName`, `LastName`, `CreatedAt time.Time`, `UpdatedAt *time.Time`) без struct-тегов и импортов проекта. Проверка: `go build ./...`, `go vet ./internal/entity/...`.
- [x] 2.3 Table-driven тест `internal/entity/category_test.go`: каждая категория из `Categories()` валидна; `Categories()` возвращает ровно 7 уникальных значений; `travel`, `еда`, `Food`, пустая строка невалидны; изменение возвращённого слайса не влияет на следующий вызов. Проверка: `go test ./internal/entity/...` зелёный.

## 3. Миграции

- [x] 3.1 `migrations/00001_create_users.sql`. Up: таблица `users` (`telegram_id BIGINT PRIMARY KEY`, `username`/`first_name`/`last_name TEXT NOT NULL DEFAULT ''`, `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`, `updated_at TIMESTAMPTZ` без дефолта). Down: удаление таблицы. Без триггеров и функций, без комментариев, кроме директив goose. Проверка: 4.4.
- [x] 3.2 `migrations/00002_create_expenses.sql`. Up: таблица `expenses` по design.md (`id BIGINT GENERATED ALWAYS AS IDENTITY`, `user_id` с `expenses_user_id_fkey` → `users(telegram_id)`, `amount BIGINT` с `expenses_amount_positive`, `category TEXT` с `expenses_category_valid`, `spent_at TIMESTAMPTZ NOT NULL DEFAULT now()`, `comment TEXT NOT NULL DEFAULT ''`, `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`, `updated_at TIMESTAMPTZ`), индексы `(user_id, spent_at DESC)` и `(spent_at DESC)`. Down: удаление таблицы. Проверка: 4.4, 4.5.
- [x] 3.3 Удалить `migrations/migrations.go` (embed не нужен). Проверка: в `migrations/` только `.sql`-файлы, `go build ./...` проходит.

## 4. Запуск миграций командами

- [x] 4.1 В `Dockerfile` добавить build-arg `GOOSE_VERSION=v3.29.0`, stage сборки goose CLI (`CGO_ENABLED=0`, теги `no_clickhouse no_libsql no_mssql no_mysql no_sqlite3 no_vertica no_ydb`) и stage `migrate` на `distroless/static-debian12:nonroot` с `/goose`, `/migrations`, `ENTRYPOINT ["/goose"]`; финальным остаётся stage приложения. Проверка: `docker build --target migrate .` и `docker build .` проходят, `docker run --rm <migrate-image> --version` печатает `v3.29.0`.
- [x] 4.2 В `docker-compose.yml` добавить сервис `migrate` (`target: migrate`, `GOOSE_DRIVER=postgres`, `GOOSE_DBSTRING=${DATABASE_URL}`, `GOOSE_MIGRATION_DIR=/migrations`, `command: ["up"]`, `depends_on: postgres: service_healthy`, `restart: "no"`), а у `app` — `depends_on: migrate: condition: service_completed_successfully`. Проверка: `docker compose config` валиден; проверка поведения — 5.4.
- [x] 4.3 В `Makefile` добавить `GOOSE_VERSION ?= v3.29.0` и цели `migrate-up`, `migrate-down`, `migrate-reset`, `migrate-status` (через `docker compose run --rm migrate ...`) и `migrate-create` (`go run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION) -dir migrations -s create $(name) sql`, без `name` — ошибка), добавить их в `.PHONY`. Проверка: `make migrate-create name=tmp_check` создаёт `migrations/00003_tmp_check.sql` (затем файл удалить), `make migrate-create` без `name` завершается ошибкой.
- [x] 4.4 Ручная проверка миграций на чистой базе (`docker compose down -v`, затем `docker compose up -d postgres`): `make migrate-up` применяет 2 миграции; `make migrate-status` показывает обе применёнными; повторный `make migrate-up` — `no migrations to run`; `make migrate-reset` проходит, `\dt` в `psql` не показывает `users` и `expenses`; `make migrate-up` снова проходит.
- [x] 4.5 Ручная проверка ограничений через `docker compose exec postgres psql`: вставка траты с категорией `travel`, `еда`, `''` → ошибка `expenses_category_valid`; `amount` `0` и `-100` → `expenses_amount_positive`; `UPDATE` категории на `travel` и `amount` на `0` → ошибка, строка не изменилась; неизвестный `user_id` → `expenses_user_id_fkey`; `DELETE` пользователя с тратами → FK-ошибка; `NULL` в `amount`/`category`/`user_id` → not-null; повтор `telegram_id` → unique; каждая из 7 категорий вставляется; у новой траты и пользователя `created_at` заполнен, `updated_at IS NULL`, `comment = ''`. Тестовые строки после проверки удалить.

## 5. Подключение при старте

- [x] 5.1 `internal/repository/postgres.go` оставить только с `NewPool(ctx, url)` (`pgxpool.New` + `Ping`, пул закрывается при ошибке Ping, ошибка `connect to database: %w`); убрать `Migrate` и импорт goose. Удалить `internal/repository/.gitkeep`. Проверка: `go build ./...`, `go vet ./...`.
- [x] 5.2 Unit-тест `internal/repository/postgres_test.go`: URL с паролем `s3cr3t-pass` на недоступный порт (`127.0.0.1:1`) и синтаксически битый URL с тем же паролем → ошибка не пустая и не содержит `s3cr3t-pass`. Без Docker. Проверка: `go test ./internal/repository/...` зелёный.
- [x] 5.3 Изменить `cmd/app/main.go`: `run(cfg, logger) error` создаёт контекст сигналов, вызывает `repository.NewPool` (с `defer pool.Close()`), пишет `info` «database connected», затем запись о старте и ждёт сигнала; `main` при ошибке `run` пишет `logger.Error` и выходит с кодом 1. Проверка: `go build ./...`, `go vet ./...`.
- [x] 5.4 Ручная проверка через compose (спека `app-bootstrap`): (а) `docker compose down -v && docker compose up --build` — `migrate` применяет миграции и завершается с кодом 0, затем `app` пишет «database connected» и запись о старте; (б) повторный `docker compose up` — `migrate` пишет `no migrations to run`, `app` стартует; (в) `make run` с `DATABASE_URL` на недоступный порт — код выхода 1, JSON-запись уровня `error` без пароля; (г) с временно сломанной миграцией `migrate` завершается с ошибкой, `app` не стартует (сломанную правку откатить).

## 6. Документация и итоговая проверка

- [x] 6.1 В `AGENTS.md` заменить «применяются при старте» на «применяются goose CLI: в compose — сервис `migrate` до старта `app`, локально — `make migrate-*`» и добавить команды миграций в раздел «Команды». Проверка: текст совпадает с design.md.
- [x] 6.2 `gofmt -l .` (пусто), `go vet ./...`, `make lint`, `go build ./...`, `go test ./...` зелёные; в новом коде и SQL нет комментариев, кроме директив `-- +goose`; `git status` не показывает `.env` и бинарей.
