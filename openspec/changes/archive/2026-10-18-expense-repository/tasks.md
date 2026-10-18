# Tasks

## 1. Пакет подключения `internal/database/postgres`

- [x] 1.1 Создать `internal/database/postgres/deps.go` с интерфейсом `Logger { Info(msg string, args ...any) }` и `internal/database/postgres/postgres.go` с `Connect(ctx, logger, connString) (*pgxpool.Pool, error)` — тело прежнего `NewPool` (`pgxpool.New` + `Ping`, закрытие пула при ошибке Ping, ошибки `pgxpool.New: %w` / `pool.Ping: %w`), после успешного Ping `logger.Info("database connected")`. Ошибки не логировать. Проверка: `go build ./...`.
- [x] 1.2 Перенести тест «пароль не утекает» в `internal/database/postgres/postgres_test.go` (логгер `slog.New(slog.DiscardHandler)`, кейсы «недоступный порт» и «битый URL» с паролем `s3cr3t-pass`). Проверка: `go test ./internal/database/postgres/...` зелёный без Docker.
- [x] 1.3 В `cmd/app/main.go` заменить `repository.NewPool` на `postgres.Connect(ctx, logger, cfg.DB.DSN)` и убрать из `run` собственную запись «database connected». Удалить `internal/repository/`. Проверка: `go build ./...`, `go vet ./...`; `make run` с рабочей БД пишет одну запись `info` «database connected», с недоступной — `error` без пароля и код выхода 1.

## 2. Доменные типы и интерфейс

- [x] 2.1 Добавить `internal/entity/errors.go` (`ErrNotFound`, `ErrUnknownUser`, `ErrInvalidAmount`, `ErrInvalidCategory`) и `internal/entity/expense_filter.go` (`ExpenseFilter{UserID *int64; From, To time.Time; Category Category}`) без импортов проекта. Проверка: `go build ./internal/entity/...`.
- [x] 2.2 Добавить `internal/usecase/deps.go` с интерфейсом `ExpenseRepository` из design.md (10 методов). Удалить `internal/usecase/.gitkeep`. Проверка: `go build ./...`.

## 3. Моки

- [x] 3.1 Добавить в `go.mod` `go.uber.org/mock` v0.6.0 (`go get -tool go.uber.org/mock/mockgen@v0.6.0`) и `github.com/stretchr/testify`. Проверка: `go mod tidy` не даёт diff.
- [x] 3.2 Создать `internal/database/repository/deps.go` с интерфейсом `Pool` (`Exec`, `Query`, `QueryRow`, сигнатуры pgx v5) и тремя директивами `//go:generate go tool mockgen` → `mock/repository.go` (`-source=./deps.go`), `mock/row.go` (`pgx.Row`), `mock/rows.go` (`pgx.Rows`), пакет `mock`. Проверка: `go generate ./...` создаёт три файла, повторный запуск не даёт diff.

## 4. Репозиторий трат

- [x] 4.1 `internal/database/repository/repository.go`: тип `Repository`, конструктор `New(pool Pool)`, константы списка колонок и имён constraint. Методы — по файлу на метод (`create_expense.go`, `get_expense.go`, `get_expense_for_user.go`, `list_expenses.go`, `list_expenses_for_user.go`, `update_expense.go`, `update_expense_for_user.go`, `delete_expense.go`, `delete_expense_for_user.go`, `list_user_ids.go`), каждый самостоятельный, без общих хелперов. «Не найдено» и нарушения constraint → голый sentinel из `entity`, прочие ошибки → `fmt.Errorf("<упавший вызов>: %w", err)` (`r.pool.QueryRow.Scan`, `r.pool.Query`, `r.pool.Exec`, `rows.Scan`, `rows.Err`). Проверка на этапе компиляции, что `*Repository` реализует `usecase.ExpenseRepository`, — в тестовом файле. Проверка: `go build ./...`.
- [x] 4.2 `internal/database/repository/repository_test.go`: testify `Suite` (`ctrl`, `pool *mock.MockPool`, `repo`), `SetupTest`/`TearDownTest` (+ `SubTest` для таблицы фильтров), `TestRepository`, хелперы `mockAnyArgs`, `expenseScanArgs`, `testExpense`, `populateExpenseScanDest`. Тесты 4.3–4.8 — по методу `Suite` на кейс (`TestМетод_Кейс`) в файлах `<метод>_test.go`. Проверка: `go test ./internal/database/repository/...`.
- [x] 4.3 `CreateExpense` (`create_expense_test.go`): точный SQL и аргументы, возвращённая трата с `id`, `created_at` и `updated_at == nil`; три constraint → голые sentinel; прочий constraint и ошибка соединения обёрнуты. Проверка: `go test ./internal/database/repository/...`.
- [x] 4.4 `GetExpense`/`GetExpenseForUser`: SQL и аргументы обоих (у `ForUser` — `AND user_id = $2` с ID пользователя A); сканирование с `updated_at` `nil` и не-`nil`; `ErrNoRows` → `ErrNotFound` (в т.ч. «A запрашивает трату B»); ошибка соединения обёрнута. Проверка: `go test ./internal/database/repository/...`.
- [x] 4.5 `ListExpenses`/`ListExpensesForUser`: table-driven по 16 комбинациям фильтра для каждого (включая чужой `f.UserID`, который игнорируется); несколько строк сканируются по порядку; пустой результат — слайс длины 0; ошибки обёрнуты как `r.pool.Query: `, `rows.Scan: `, `rows.Err: `. Проверка: `go test ./internal/database/repository/...`.
- [x] 4.6 `UpdateExpense`/`UpdateExpenseForUser`: SQL содержит `updated_at = now()`, аргументы обоих, `updated_at` из `RETURNING` в результате; `ErrNoRows` → `ErrNotFound` (несуществующая и «A меняет трату B»); constraint суммы и категории → sentinel; ошибка соединения обёрнута. Проверка: `go test ./internal/database/repository/...`.
- [x] 4.7 `DeleteExpense`/`DeleteExpenseForUser`: `"DELETE 1"` → нет ошибки, `"DELETE 0"` → `ErrNotFound` (несуществующая и «A удаляет трату B»), ошибка `Exec` обёрнута. Проверка: `go test ./internal/database/repository/...`.
- [x] 4.8 `ListUserIDs`: `SELECT telegram_id FROM users ORDER BY telegram_id`; несколько ID по порядку, пустой результат — слайс длины 0, ошибки `r.pool.Query`, `rows.Scan`, `rows.Err` обёрнуты. Проверка: `go test ./internal/database/repository/...`.

## 5. Документация и итоговая проверка

- [x] 5.1 Обновить `AGENTS.md`: в дереве `internal/repository/` заменить на `internal/database/postgres/` (подключение) и `internal/database/repository/` (доступ к данным); в архитектурных правилах `repository` → `database/repository`, транспорт и агент не обращаются к `database/*`; в разделе «Команды» указать, что `go generate ./...` генерирует и моки. Проверка: текст совпадает с design.md, упоминаний `internal/repository` не осталось (`grep -rn "internal/repository" --exclude-dir=openspec .` пусто).
- [x] 5.2 `gofmt -l .` (пусто), `go vet ./...`, `make lint`, `go build ./...`, `go test ./...` зелёные; в рукописном коде нет комментариев и директив, кроме `//go:generate`; `go mod tidy` не даёт diff.
