# Design

## Context

Мотивация — см. proposal.md, требования — `specs/expense-repository/spec.md`.

Текущее состояние:
- `internal/repository/postgres.go` содержит только `NewPool(ctx, dsn)` (`pgxpool.New` + `Ping`), `postgres_test.go` проверяет, что пароль не попадает в ошибку. `cmd/app/main.go` вызывает `repository.NewPool` и сам пишет `info` «database connected».
- Схема (`migrations/00002_create_expenses.sql`) уже даёт стабильные имена ограничений: `expenses_user_id_fkey`, `expenses_amount_positive`, `expenses_category_valid`, а также индексы `(user_id, spent_at DESC)` и `(spent_at DESC)`. Триггеров нет, `updated_at` выставляет SQL репозитория.
- `internal/entity` содержит `Expense`, `User`, `Category`. `internal/usecase` пуст.
- В проекте нет комментариев в коде и директив вроде `//nolint`: это решение пользователя. Исключение — `//go:generate` для моков.

## Goals / Non-Goals

**Goals:**
- Методы для бота физически не могут вернуть или изменить чужую трату: `user_id` входит в `WHERE` каждого их запроса.
- Ошибки, на которые будет реагировать `usecase` («не найдено», нарушения ограничений), различимы через `errors.Is` и не зависят от pgx.
- Репозиторий тестируется без БД и Docker — по решению пользователя, через gomock.

**Non-Goals:**
- Валидация до запроса в БД (сумма, категория, дата по умолчанию) — это `usecase`. Репозиторий полагается на ограничения схемы и только переводит их нарушения в доменные ошибки.
- Пагинация, сортировка по выбору вызывающего, «последняя трата», агрегаты.
- Транзакции между несколькими операциями: каждая операция — один SQL-запрос.
- Подключение репозитория в `main`: потребителя пока нет, а неиспользуемая переменная не нужна.

## Decisions

### Раскладка пакетов: `internal/database/{postgres,repository}`

По требованию пользователя `internal/repository` заменяется на:
- `internal/database/postgres` — только подключение. В `deps.go` интерфейс `Logger { Info(msg string, args ...any) }`, в `postgres.go`:
  ```go
  func Connect(ctx context.Context, logger Logger, connString string) (*pgxpool.Pool, error)
  ```
  Тело — прежнее `NewPool` (`pgxpool.New` + `Ping`, при ошибке Ping пул закрывается, ошибки `pgxpool.New: %w` и `pool.Ping: %w`). Плюс после успешного Ping — `logger.Info("database connected")`, эта запись переезжает из `main`. Интерфейс `Logger` объявлен у потребителя и минимален, `*slog.Logger` ему соответствует. Ошибки `Connect` не логирует, только возвращает: иначе `main` залогирует их второй раз. Тест «пароль не утекает» переезжает сюда с логгером на `slog.DiscardHandler`.
- `internal/database/repository` — операции. Тип `Repository` (`New(pool Pool)`), по одному файлу на метод с именем метода в snake_case (`create_expense.go`, `get_expense_for_user.go`, …). Все интерфейсы-зависимости пакета лежат в `deps.go`:
  ```go
  type Pool interface {
      Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
      Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
      QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
  }
  ```
  `*pgxpool.Pool` удовлетворяет `Pool` (сигнатуры совпадают), подойдёт и `pgx.Tx`, если понадобятся транзакции.

Это расходится со структурой в `AGENTS.md` (`internal/repository/`), поэтому `AGENTS.md` обновляется: дерево каталогов и правила зависимостей (`database/repository` реализует интерфейсы из `usecase`; транспорт и агент к `database/*` не обращаются).

### Интерфейс в `usecase`, типы в `entity`

`internal/usecase/deps.go` (тем же именем, что `deps.go` в репозитории, — «зависимости потребителя»):
```go
type ExpenseRepository interface {
    CreateExpense(ctx context.Context, e entity.Expense) (entity.Expense, error)
    GetExpense(ctx context.Context, id int64) (entity.Expense, error)
    ListExpenses(ctx context.Context, f entity.ExpenseFilter) ([]entity.Expense, error)
    UpdateExpense(ctx context.Context, e entity.Expense) (entity.Expense, error)
    DeleteExpense(ctx context.Context, id int64) error

    GetExpenseForUser(ctx context.Context, userID, id int64) (entity.Expense, error)
    ListExpensesForUser(ctx context.Context, userID int64, f entity.ExpenseFilter) ([]entity.Expense, error)
    UpdateExpenseForUser(ctx context.Context, userID int64, e entity.Expense) (entity.Expense, error)
    DeleteExpenseForUser(ctx context.Context, userID, id int64) error

    ListUserIDs(ctx context.Context) ([]int64, error)
}
```
Имена методов называют действие и сущность (`CreateExpense`, а не `Create`): в тот же `Repository` позже добавятся методы пользователей (`CreateUser` и т.п.), и имена не должны конфликтовать.

`CreateExpense` отдельного варианта «для пользователя» не имеет: `user_id` новой траты и так передаётся кодом в `e.UserID`.

Фильтр и ошибки — в `entity`, чтобы `database/repository` зависел только от `entity` и не импортировал `usecase` (Go-интерфейсы реализуются неявно):
```go
type ExpenseFilter struct {
    UserID   *int64
    From     time.Time
    To       time.Time
    Category Category
}
```
Нулевые `From`/`To` и пустая `Category` означают «без условия». `UserID` — указатель, чтобы «не задан» не путалось с ID. Граница `[From, To)`: так соседние периоды («неделя», «месяц») не пересекаются и не теряют траты на границе.

`internal/entity/errors.go`: `ErrNotFound` (общий, пригодится и для других сущностей), `ErrUnknownUser`, `ErrInvalidAmount`, `ErrInvalidCategory` (`errors.New`).

**Альтернатива: один набор методов с необязательным `userID *int64`.** Отвергнута. `nil` по ошибке в коде бота молча снимает изоляцию, а отдельные методы с `int64` делают ошибку видимой в сигнатуре. Каждый метод — самостоятельная функция со своим SQL, сканированием и обработкой ошибок, без общих приватных хелперов и без вызова одного метода из другого (решение пользователя: читаемость одного файла важнее отсутствия повторов).

**Альтернатива: `ListExpensesForUser(ctx, f)` с проверкой, что `f.UserID` задан.** Отвергнута: это ошибка времени выполнения вместо гарантии в типах. `ListExpensesForUser` всегда ставит условие `user_id = $1` с переданным `userID` и не читает `f.UserID`, поэтому указанный в фильтре чужой ID ни на что не влияет (сценарий в спеке).

### SQL

Общая только константа списка колонок `id, user_id, amount, category, spent_at, comment, created_at, updated_at` (и константы имён constraint) в `repository.go`. Сканирование пишется в каждом методе. `updated_at` сканируется в `&e.UpdatedAt` (`**time.Time`), `NULL` → `nil`.

- `CreateExpense`: `INSERT INTO expenses (user_id, amount, category, spent_at, comment) VALUES ($1..$5) RETURNING <cols>`. `spent_at` передаётся всегда, правило «по умолчанию сегодня» — в `usecase`.
- `GetExpense`/`GetExpenseForUser`: `SELECT <cols> FROM expenses WHERE id = $1 [AND user_id = $2]`. `pgx.ErrNoRows` → `ErrNotFound`.
- `ListExpenses`/`ListExpensesForUser`: условия собираются слайсом строк и аргументов: `user_id = $n`, `spent_at >= $n`, `spent_at < $n`, `category = $n`. Порядок условий фиксированный, номера плейсхолдеров идут подряд. `WHERE` добавляется, только если условия есть. Хвост — `ORDER BY spent_at DESC, id DESC`, совпадает с индексами. Пустой результат — `[]entity.Expense{}`, не `nil`, чтобы JSON в REST был `[]`. Обход — `for rows.Next()`, затем `rows.Err()` и `defer rows.Close()`. `pgx.CollectRows` не используется: с мок-строками ему нужны `FieldDescriptions`/`Values`, а ручной цикл мокается просто.
- `UpdateExpense`/`UpdateExpenseForUser`: `UPDATE expenses SET amount = $1, category = $2, spent_at = $3, comment = $4, updated_at = now() WHERE id = $5 [AND user_id = $6] RETURNING <cols>`. Один запрос, без предварительного `SELECT`. `pgx.ErrNoRows` → `ErrNotFound`. Время берётся из `now()` БД, а не из часов приложения, — так же, как `created_at` (`DEFAULT now()`), чтобы метки одной записи не расходились из-за рассинхрона часов. Полная замена изменяемых полей: частичное изменение («поменяй только сумму») собирает `usecase` из прочитанной траты.
- `DeleteExpense`/`DeleteExpenseForUser`: `Exec("DELETE FROM expenses WHERE id = $1 [AND user_id = $2]")`, `RowsAffected() == 0` → `ErrNotFound`.
- `ListUserIDs`: `SELECT telegram_id FROM users ORDER BY telegram_id`. Пустой результат → `[]int64{}`.

### Ошибки

Каждый метод обрабатывает ошибки сам:
- «не найдено» и нарушения ограничений возвращаются голым sentinel из `entity/errors.go`, без обёртки: `pgx.ErrNoRows` (и `RowsAffected() == 0` в удалении) → `ErrNotFound`; `*pgconn.PgError` (через `if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok`) по `ConstraintName`: `expenses_user_id_fkey` → `ErrUnknownUser`, `expenses_amount_positive` → `ErrInvalidAmount`, `expenses_category_valid` → `ErrInvalidCategory`;
- всё остальное оборачивается именем упавшего вызова (не метода репозитория): `fmt.Errorf("r.pool.QueryRow.Scan: %w", err)`, `r.pool.Query`, `r.pool.Exec`, `rows.Scan`, `rows.Err`.

Сопоставление по имени constraint, а не по SQLSTATE: у `amount` и `category` один код `23514`, различает их только имя. Имена закреплены в design задачи 02 ровно для этого.

### Тесты: gomock-моки `Pool`, `pgx.Rows`, `pgx.Row`

По решению пользователя тесты репозитория — только unit на моках, без реальной Postgres.
- Библиотека `go.uber.org/mock` v0.6.0 (поддерживаемый форк `golang/mock`, архивного с 2023 года).
- Моки генерирует `go generate ./...` по директивам в `internal/database/repository/deps.go` в пакет корневой `mock`:
  ```go
  //go:generate go tool mockgen -source=./deps.go -destination ../../../mock/repository.go -package mock
  //go:generate go tool mockgen -destination ../../../mock/row.go -package mock github.com/jackc/pgx/v5 Row
  //go:generate go tool mockgen -destination ../../../mock/rows.go -package mock github.com/jackc/pgx/v5 Rows
  ```
  `go tool mockgen`, а не `go run ...@latest`: версия зафиксирована в `go.mod` (`tool go.uber.org/mock/mockgen`, v0.6.0), генерация воспроизводима и не ходит в сеть. Мок `pgx.Tx` не генерируется — транзакций нет. Директива `//go:generate` — единственное исключение из правила «без комментариев в коде». Сгенерированные файлы коммитятся и руками не правятся.
- Тесты — testify `suite` в пакете `repository_test`: `Suite` с `ctrl`, `pool *mock.MockPool`, `repo *repository.Repository`; `SetupTest`/`TearDownTest` создают и завершают контроллер (`SetupSubTest`/`TearDownSubTest` — для таблицы фильтров). Конструктор — `repository.New(pool)`. Стиль: один кейс — один метод `TestМетод_Кейс` (`_Success`, `_NotFound`, `_OtherUsersExpense`, `_InvalidAmount`, `_ScanError`, …), внутри Arrange (`ctx`, `mock.NewMockRow`/`NewMockRows`, `EXPECT()` цепочкой) → Act → Assert по полям результата. Хелперы: `mockAnyArgs(n)`, `expenseScanArgs()` (8 × `gomock.Any()`), `testExpense()`, `populateExpenseScanDest(e, dest...)`. Sentinel проверяется `s.Equal(entity.ErrNotFound, err)` и нулевым результатом, прочие ошибки — `s.ErrorIs` и `s.Contains(err.Error(), "r.pool.QueryRow.Scan")`. Текст SQL сверяется точно там, где он и есть проверяемое поведение (фильтры, `*ForUser`, `updated_at = now()`), в остальных кейсах — `gomock.Any()`. Единственная таблица — комбинации фильтров (16 для `ListExpenses`, 16 для `ListExpensesForUser`, включая чужой `f.UserID`), через `s.Run`.
- Тестовые хелперы: `scanValues(vals ...any)` — `DoAndReturn` для `Scan`, раскладывающий значения в указатели (`*int64`, `*entity.Category`, `*time.Time`, `**time.Time` и т.д.); `expectRows(rows, n)` — последовательность `Next()` → true×n, false, затем `Err()`, `Close()`.
- Что проверяется: точный текст SQL и аргументы каждой операции (`gomock.Eq`); сканирование всех полей, включая `updated_at` `nil` и не-`nil`; маппинг ошибок: `ErrNoRows`, три constraint, неизвестный `PgError`, произвольная ошибка, `RowsAffected() == 0`, ошибка `rows.Err()`; пустой список — не `nil`.
- Фильтры: table-driven по всем 16 комбинациям (`UserID`, `From`, `To`, `Category` — задан/не задан) для `ListExpenses` и по 8 комбинациям `From`/`To`/`Category` для `ListExpensesForUser`, в том числе с чужим `f.UserID` → в аргументах переданный `userID`.
- Тесты лежат по файлу на метод (`create_expense_test.go`, …), `Suite` и хелперы — в `repository_test.go`.
- Изоляция A/B (требование `AGENTS.md`): для `GetExpenseForUser`, `UpdateExpenseForUser`, `DeleteExpenseForUser` проверяется, что `user_id` A есть в `WHERE` и аргументах, и что при пустом ответе БД (так она ответит на трату B) возвращается `ErrNotFound`.
- `updated_at`: SQL `UpdateExpense`/`UpdateExpenseForUser` содержит `updated_at = now()`, а значение из `RETURNING` попадает в результат.

## Risks / Trade-offs

- [Моки проверяют текст SQL, а не его выполнение: опечатка в имени колонки, неверная граница `<`/`<=` или тип плейсхолдера, который Postgres не примет, пройдут тесты и сломаются на первом реальном запросе] → это осознанный выбор пользователя. Смягчение: SQL короткий и собран в одном файле. Ожидаемые строки в тестах пишутся по спеке (граница `[From, To)`, порядок), а не копируются из реализации. Первый потребитель (REST в задаче 04/05) даст сквозную проверку на compose-Postgres. Предлагаю позже добавить интеграционные тесты на testcontainers.
- [Тесты с точным текстом SQL хрупкие: любая переформатировка запроса ломает тесты] → SQL хранится в константах/собирается в одном месте; тест сравнивает ту же строку, переформатирование — это осознанная правка в двух местах.
- [Критерий «`updated_at` обновляется при изменении» на моках проверяется только как наличие `updated_at = now()` в SQL и проброс значения из `RETURNING`] → см. первый пункт.
- [Изоляция A/B на моках проверяется по `WHERE ... AND user_id = $n`, а не по реальному отсутствию строки] → то же. Условие в SQL — единственный механизм изоляции, и оно проверяется точным сравнением.
- [`UpdateExpense` перезаписывает все изменяемые поля: два одновременных частичных изменения одной траты потеряют одно из них] → для семейного бота риск ничтожен; блокировки и версии не вводятся.
- [Методы без общих хелперов повторяют сканирование и маппинг ошибок; правка схемы (новая колонка) затрагивает каждый файл] → колонки в одной константе, тесты на каждый метод сверяют число и порядок полей в `Scan`.
- [Переименование пакета ломает импорты] → потребитель один (`cmd/app/main.go`), правится в этом же изменении.

## Migration Plan

Схема не меняется. Деплой — обычный `docker compose up --build`. Откат — revert коммита.
