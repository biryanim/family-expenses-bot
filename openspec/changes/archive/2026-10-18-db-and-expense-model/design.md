# Design

## Context

Мотивация — см. proposal.md, требования — `specs/expense-storage/spec.md` и `specs/app-bootstrap/spec.md`.

Текущее состояние:
- `cmd/app/main.go` загружает конфиг, настраивает `slog`, ждёт сигнала. `run(ctx, cfg, logger)` пока ничего не возвращает. В дизайне каркаса было заложено `run(...) error`, но так и не реализовано.
- `internal/config` уже требует `DATABASE_URL`, а `Config.LogValue()` не выводит его в лог.
- В `docker-compose.yml` есть `postgres:17-alpine` с healthcheck, и `app` стартует только после него. Порт Postgres наружу не проброшен.
- `Dockerfile` — multi-stage: `golang:1.27-alpine` собирает бинарь, финальный образ — `distroless/static-debian12:nonroot`.
- `migrations/`, `internal/entity/`, `internal/repository/` пусты (только `.gitkeep`).

## Goals / Non-Goals

**Goals:**
- Инварианты из спеки (категория, `amount > 0`, обязательные поля, FK) гарантируются самой БД. Код приложения их дублирует, но не служит единственной защитой.
- Ограничения схемы имеют стабильные имена, чтобы репозиторий (задача 03) мог превращать нарушения в доменные ошибки по имени constraint.
- Схемой управляют явные команды; приложение схему не трогает.

**Non-Goals:**
- Методы репозитория (CRUD, фильтры, список членов семьи, upsert пользователя, выставление `updated_at`) — задачи 03 и 06.
- Автоматические тесты на реальной БД: в этом изменении миграции и ограничения проверяются вручную командами (см. «Проверка»).
- Часовой пояс семьи и вычисление границ «сегодня/вчера» — см. Open Questions.
- Конфигурация пула (размер, таймауты) через env: хватает значений `pgxpool` по умолчанию.

## Decisions

### Схема

`users`:
- `telegram_id BIGINT PRIMARY KEY` — Telegram ID помещается в int64. Ключ естественный: суррогатный `id` ничего не даёт, а `expenses.user_id` по смыслу и есть Telegram ID (так сформулировано в `AGENTS.md`).
- `username`, `first_name`, `last_name` — `TEXT NOT NULL DEFAULT ''`. Пустая строка вместо NULL: в Go это просто `string`, а для отображения «нет username» и «пустой username» — одно и то же.
- `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`, `updated_at TIMESTAMPTZ` (nullable, без дефолта).

`expenses`:
- `id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY`. `GENERATED ALWAYS` не даёт вставить свой id в обход последовательности (в отличие от `BY DEFAULT`/`serial`).
- `user_id BIGINT NOT NULL`, constraint `expenses_user_id_fkey` → `users(telegram_id)`. Поведение `ON DELETE` по умолчанию (`NO ACTION`): удалить члена семьи с тратами нельзя. Альтернатива `CASCADE` отвергнута — молча теряет траты, а удаления членов семьи в продукте нет.
- `amount BIGINT NOT NULL`, constraint `expenses_amount_positive CHECK (amount > 0)`.
- `category TEXT NOT NULL`, constraint `expenses_category_valid CHECK (category IN ('food','transport','housing','entertainment','health','shopping','other'))`.
- `spent_at TIMESTAMPTZ NOT NULL DEFAULT now()` — тип выбран пользователем. Дефолт в БД — страховка; правило «по умолчанию сегодня» живёт в `usecase`.
- `comment TEXT NOT NULL DEFAULT ''` — по той же причине, что и текстовые поля `users`.
- `created_at`, `updated_at` — как у `users`.
- Индексы: `(user_id, spent_at DESC)` — основной запрос бота «мои траты за период, свежие сверху», он же покрывает фильтр панели по члену семьи; `(spent_at DESC)` — список панели по всем за период. По категории индекс не нужен: семейный объём данных маленький, а категорий всего семь.

### Категории: английские коды + CHECK

Коды вместо русских строк: идентификаторы в коде на английском (`AGENTS.md`), в REST API и JSON-схемах инструментов агента — ASCII-значения без проблемы «жильё/жилье». Русские названия — забота отображения: бот, промпт агента, фронтенд.

CHECK вместо `ENUM` и справочной таблицы:
- `CREATE TYPE ... AS ENUM` — значение нельзя удалить или переименовать без пересоздания типа, а `ALTER TYPE ... ADD VALUE` имеет ограничения в транзакциях. Выигрыша нет.
- Справочная таблица `categories` с FK — лишний join и лишние данные ради фиксированного списка из семи значений.
- CHECK меняется одной миграцией (`DROP CONSTRAINT` + `ADD CONSTRAINT`).

В Go (`internal/entity/category.go`): тип `Category string`, семь констант, `Categories() []Category`, `(Category) Valid() bool`. Сопоставление русских названий и «нераспознанная → `other`» сюда не входят. Это правило `usecase`, и появится оно вместе с первым потребителем.

### Временные метки: без триггеров, `updated_at` nullable

`created_at` заполняет `DEFAULT now()`. `updated_at` при вставке остаётся `NULL`: это даёт бесплатный признак «запись ни разу не изменялась». При изменении `updated_at = now()` явно ставит SQL репозитория (задача 03, у неё есть критерий «`updated_at` обновляется при изменении» и тесты на него).

Триггер `BEFORE UPDATE` отвергнут по решению пользователя. Он скрывает поведение от читателя SQL репозитория, а все `UPDATE` пишутся вручную в одном пакете, так что забытое поле ловится тестом репозитория. Цена — дисциплина: любой новый `UPDATE` обязан ставить `updated_at`.

### Сущности

`internal/entity/expense.go`:
```go
type Expense struct {
    ID        int64
    UserID    int64
    Amount    int64
    Category  Category
    SpentAt   time.Time
    Comment   string
    CreatedAt time.Time
    UpdatedAt *time.Time
}
```
`internal/entity/user.go`: `User{TelegramID int64; Username, FirstName, LastName string; CreatedAt time.Time; UpdatedAt *time.Time}`.

`*time.Time` для `updated_at` прямо отражает nullable-колонку; pgx сканирует `NULL` в `nil`. Альтернатива `pgtype.Timestamptz` отвергнута — сущности не должны зависеть от драйвера. Struct-тегов нет: сериализация в API — забота `transport/http` (свои DTO), сканирование — `repository`.

### Миграции: goose CLI, а не код приложения

Файлы в `migrations/` нумеруются последовательно (`goose create -s`), а не по timestamp: изменения идут строго по порядку, и последовательные номера проще читать.
- `00001_create_users.sql` — таблица `users`.
- `00002_create_expenses.sql` — таблица `expenses`, constraints, индексы.

goose выполняет каждую миграцию в транзакции; DDL в PostgreSQL транзакционный, так что частично применённой миграции не остаётся.

Применение — goose CLI по решению пользователя. Приложение миграции не запускает и goose как Go-модуль не импортирует. Плюсы: схема меняется только явной командой, откат — тоже команда, бинарь приложения не тащит goose и SQL.

**Образ.** В `Dockerfile` добавляется stage `migrate`:
- в builder-stage на `golang:1.27-alpine` собирается `go install github.com/pressly/goose/v3/cmd/goose@v3.29.0` с `CGO_ENABLED=0` и тегами `no_clickhouse no_libsql no_mssql no_mysql no_sqlite3 no_vertica no_ydb`, чтобы в бинарь попал только драйвер postgres;
- финальный образ stage `migrate` — `distroless/static-debian12:nonroot` с `/goose` и `/migrations`, `ENTRYPOINT ["/goose"]`.

Stage `migrate` объявляется до финального stage приложения, чтобы `docker build .` без `--target` по-прежнему собирал приложение. Сторонние образы с goose (`kukymbr/goose-docker` и т.п.) отвергнуты: неофициальные, а версия CLI должна быть зафиксирована нами.

**Compose.** Сервис `migrate`: `build: {context: ., target: migrate}`, переменные `GOOSE_DRIVER=postgres`, `GOOSE_DBSTRING=${DATABASE_URL}`, `GOOSE_MIGRATION_DIR=/migrations` (goose CLI читает их сам), команда `up`, `depends_on: postgres: service_healthy`, `restart: "no"`. У `app` добавляется `depends_on: migrate: condition: service_completed_successfully`. Так `docker compose up` по-прежнему одна команда, а упавшая миграция не даёт стартовать приложению.

**Makefile.** Цели работают через тот же сервис, чтобы не требовать локального goose и проброса порта Postgres:
- `migrate-up` → `docker compose run --rm migrate up`;
- `migrate-down` → `... migrate down` (откат одной миграции);
- `migrate-reset` → `... migrate reset` (полный откат);
- `migrate-status` → `... migrate status`;
- `migrate-create name=<name>` → `go run github.com/pressly/goose/v3/cmd/goose@v3.29.0 -dir migrations -s create $(name) sql` — файл создаётся на хосте, Docker не нужен.

Версия goose задаётся одной переменной `GOOSE_VERSION` в Makefile и build-arg в Dockerfile (значение по умолчанию то же).

### Подключение в коде

`internal/repository/postgres.go`: `NewPool(ctx, url) (*pgxpool.Pool, error)` — `pgxpool.New` + `Ping`, потому что `pgxpool.New` ленивый и сам по себе недоступность БД не обнаружит. При ошибке Ping пул закрывается. Место — `repository`, это код доступа к PostgreSQL; отдельный пакет `internal/db` отвергнут — его нет в утверждённой структуре.

### Старт в `main`

`run(cfg, logger) error`: `signal.NotifyContext` → `NewPool` → `defer pool.Close()` → лог `info` «database connected» → запись о старте → ожидание сигнала. Контекст сигналов создаётся внутри `run`, а не в `main`: так в `main` нет defer'ов, которые `os.Exit(1)` пропустил бы (gocritic `exitAfterDefer`). `main` при ошибке пишет `logger.Error` и вызывает `os.Exit(1)`.

Ошибка оборачивается с контекстом (`connect to database: %w`). `DATABASE_URL` в текст ошибок не подставляется. pgx в ошибках разбора конфига сам маскирует пароль, а в ошибках подключения выводит только host/user/database. Это закрепляется unit-тестом без БД (недоступный порт и битый URL — пароля в ошибке нет).

### Проверка

Автоматических тестов на реальной БД в этом изменении нет — по решению пользователя. Критерии готовности проверяются вручную на compose-Postgres:
- `make migrate-up` на чистой базе → `make migrate-status` → `make migrate-reset` (таблиц нет) → `make migrate-up` снова → повторный `make migrate-up` (no-op);
- `psql` через `docker compose exec postgres`: вставки с недопустимой категорией, `amount <= 0`, неизвестным `user_id`, `NULL` в обязательных полях отклоняются; `created_at` заполнен, `updated_at` — `NULL`.

Unit-тесты: `internal/entity` (категории) и `NewPool` (пароль не утекает).

## Risks / Trade-offs

- [`updated_at` не выставлен в каком-нибудь новом `UPDATE`] → все `UPDATE` пишутся в `internal/repository`, тест репозитория (задача 03) проверяет `updated_at` после изменения.
- [Без автотестов на БД ограничения схемы могут сломаться незаметно при будущей миграции] → ручная проверка из «Проверка» в tasks.md. Набор категорий в Go и в CHECK сверяется вручную. Тесты репозитория в задаче 03 потребуют реальную БД — способ решается там.
- [Миграции и приложение не синхронизированы: `app` запущен на старой схеме] → в compose `app` зависит от успешного `migrate`; локально `make run` без `make migrate-up` упадёт на первом запросе, а не на старте.
- [`TIMESTAMPTZ` хранит момент, а «сегодня», «вчера» и границы периода зависят от часового пояса] → в этой задаче это не проявляется. Задачам 03/08 нужен явный часовой пояс семьи (см. Open Questions).
- [Версии goose в Makefile и Dockerfile разойдутся] → обе берут одно значение по умолчанию; при обновлении меняются вместе.
- [`docker compose run --rm migrate` поднимает `postgres`, если он не запущен] → это ожидаемо и удобно; `make down` его останавливает.

## Migration Plan

Первая схема — переносить нечего. На существующем volume `pgdata` (после задачи 01 в нём пустая база) миграции применяются при первом `docker compose up` сервисом `migrate`. Откат: `make migrate-reset` или `docker compose down -v`.

## Open Questions

- Как задаётся часовой пояс семьи для «сегодня/вчера» и фильтров по периоду (например, env `APP_TIMEZONE` с дефолтом `Europe/Moscow`)? Схему не меняет, решается в задачах 03/08.
