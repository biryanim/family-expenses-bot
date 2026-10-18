# Tasks

## 1. Контракт

- [x] 1.1 В `api/openapi.yaml` добавить `invalid_date` в `components.schemas.ErrorCode.enum` (после `invalid_category`). Запустить `go generate ./...`. Проверка: в `openapi.gen.go` появилась константа для `invalid_date`, `go test ./internal/transport/http/openapi/...` (валидация спецификации) зелёный, повторный `go generate ./...` не даёт diff.

## 2. Домен и репозиторий

- [x] 2.1 В `internal/entity/errors.go` добавить `ErrInvalidDate = errors.New("invalid date")`. Проверка: `go build ./...`.
- [x] 2.2 Заменить `list_user_ids.go` на `internal/database/repository/list_users.go`: `ListUsers(ctx) ([]entity.User, error)`, `SELECT telegram_id, username, first_name, last_name, created_at, updated_at FROM users ORDER BY telegram_id`, пустой результат — слайс длины 0, ошибки `r.pool.Query`, `rows.Scan`, `rows.Err` обёрнуты. В `usecase.ExpenseRepository` заменить `ListUserIDs` на `ListUsers`. `go generate ./...` для моков. Проверка: `list_users_test.go` (suite, кейсы `_Success` с несколькими членами семьи и пустыми строками, `_Empty`, `_QueryError`, `_ScanError`, `_RowsError`) вместо `list_user_ids_test.go`; `go test ./internal/database/repository/...` зелёный.

## 3. Слой usecase

- [x] 3.1 В `internal/usecase/deps.go` добавить директиву `//go:generate go tool mockgen -source=./deps.go -destination ../../mock/usecase.go -package mock`. Создать `internal/usecase/usecase.go`: `type Usecase struct { repo ExpenseRepository }`, `New(repo ExpenseRepository) *Usecase`. Удалить `.gitkeep`, если есть. Проверка: `go generate ./...` создаёт `mock/usecase.go` с `MockExpenseRepository`, `go build ./...`.
- [x] 3.2 `usecase_test.go`: testify `Suite` в пакете `usecase_test` (`ctrl`, `repo`, `uc`), `SetupTest`/`TearDownTest`, `TestUsecase`. Проверка: `go test ./internal/usecase/...` запускает suite.
- [x] 3.3 `create_expense.go`: `Amount <= 0` → `ErrInvalidAmount`, `!Category.Valid()` → `ErrInvalidCategory` (без вызова репозитория); нулевой `SpentAt` → `time.Now()`; любая ошибка репозитория — `fmt.Errorf("u.repo.CreateExpense: %w", err)`. Проверка: `create_expense_test.go` — `_Success`, `_DefaultSpentAt` (дата между моментами до и после вызова), `_InvalidAmount` (0 и -100), `_InvalidCategory` (`travel`, `""`), `_UnknownUser`, `_RepositoryError` (`s.ErrorIs` + `s.Contains(..., "u.repo.CreateExpense")`), зелёный.
- [x] 3.4 `get_expense.go` и `delete_expense.go`: прямой вызов репозитория, любая ошибка (включая `ErrNotFound`) обёрнута именем вызова. Проверка: `get_expense_test.go`, `delete_expense_test.go` — `_Success`, `_NotFound`, `_RepositoryError`, зелёные.
- [x] 3.5 `list_expenses.go`: непустая невалидная категория → `ErrInvalidCategory`; `From` и `To` заданы и `From.After(To)` → `ErrInvalidDate`; иначе `repo.ListExpenses` с тем же фильтром. Проверка: `list_expenses_test.go` — `_Success` (фильтр передаётся без изменений), `_EqualBounds` (вызов репозитория есть), `_InvalidCategory`, `_InvalidPeriod`, `_RepositoryError`, зелёный.
- [x] 3.6 `update_expense.go`: проверки суммы, категории, нулевой `SpentAt` → `ErrInvalidDate`, затем `repo.UpdateExpense`. Проверка: `update_expense_test.go` — `_Success`, `_InvalidAmount`, `_InvalidCategory`, `_InvalidDate`, `_NotFound`, `_RepositoryError`, зелёный.
- [x] 3.7 `list_members.go`: `ListMembers(ctx) ([]entity.User, error)` → `repo.ListUsers`. Проверка: `list_members_test.go` — `_Success`, `_Empty`, `_RepositoryError`, зелёный.

## 4. Конфиг

- [x] 4.1 В `internal/config` добавить `HTTP.AllowedOrigins []string` из `CORS_ALLOWED_ORIGINS` (split по `,`, `TrimSpace`, без пустых; пусто → `["http://localhost:5173"]`), в `LogValue` — `cors_allowed_origins`. В `.env.example` — переменная с комментарием. Проверка: `config_test.go` — дефолт, одно значение, несколько с пробелами и пустыми элементами; `go test ./internal/config/...` зелёный.

## 5. HTTP-транспорт

- [x] 5.1 Добавить `github.com/go-chi/cors` v1.2.2 (`go get`). Проверка: `go mod tidy` не даёт diff после использования в 5.4.
- [x] 5.2 `internal/transport/http/deps.go`: интерфейс `Usecase` (`CreateExpense`, `GetExpense`, `ListExpenses`, `UpdateExpense`, `DeleteExpense`, `ListMembers` с сигнатурами `usecase.Usecase`) и директива mockgen → `mock/http.go`. `server.go`: `Server{usecase, logger}`, `New`, `var _ openapi.StrictServerInterface = (*Server)(nil)`. Удалить `.gitkeep`. Проверка: `go generate ./...` создаёт `mock/http.go` с `MockUsecase`; сборка пройдёт после 5.5.
- [x] 5.3 `handle_request_error.go` (`HandleRequestError`: 400, `application/json`, `code=invalid_request`, `message` — текст ошибки разбора) и `handle_response_error.go` (метод `(*Server).HandleResponseError`: логирует, 500, `code=internal`, общее сообщение). Проверка: покрывается тестами 5.6.
- [x] 5.4 `new_router.go`: `NewRouter(s *Server, allowedOrigins []string) http.Handler` — chi, `middleware.Recoverer`, `cors.Handler` (методы GET/POST/PUT/DELETE/OPTIONS, заголовок `Content-Type`, без credentials, `MaxAge` 300), `openapi.HandlerWithOptions` с `BaseURL: "/api"`, `ErrorHandlerFunc: HandleRequestError` и strict-обёрткой с `RequestErrorHandlerFunc: HandleRequestError`, `ResponseErrorHandlerFunc: HandleResponseError`. Проверка: покрывается тестами 5.6.
- [x] 5.5 Хендлеры по файлу на метод: `list_expenses.go`, `create_expense.go`, `get_expense.go`, `update_expense.go`, `delete_expense.go`, `list_members.go`. Каждый переводит `openapi` → `entity` (`SpentAt == nil` → нулевое время, `Comment == nil` → `""`, параметры фильтра → `entity.ExpenseFilter`), вызывает `usecase`, маппит ошибки `switch`'ем с `errors.Is`: `ErrInvalidAmount` → 400 `invalid_amount`, `ErrInvalidCategory` → 400 `invalid_category`, `ErrInvalidDate` → 400 `invalid_date`, `ErrUnknownUser` → 400 `unknown_member`, `ErrNotFound` → 404 `not_found`, прочее → `logger.ErrorContext` + 500 `internal` с общим сообщением. `error` в strict-обёртку не возвращает. Проверка: `go build ./...`, `go vet ./...`.
- [x] 5.6 Тесты `transport/http`: `server_test.go` — suite в `http_test` (`ctrl`, `MockUsecase`, `router` из `NewRouter` с origin `http://localhost:5173`, загруженная `api/openapi.yaml` и `openapi3filter`-роутер), метод suite для выполнения запроса через `httptest` и проверки ответа по спецификации `openapi3filter.ValidateResponse`. Файлы на метод: `list_expenses_test.go` (`_Success`, `_Empty` — `[]`, а не `null`, `_InvalidUserID`, `_InvalidFrom`, `_InvalidCategory`, `_InvalidPeriod`, `_InternalError` — тело без текста исходной ошибки; таблица `_Filters` с `s.Run`: каждая комбинация `user_id`/`from`/`to`/`category` → ожидаемый `entity.ExpenseFilter`), `create_expense_test.go` (`_Success` 201, `_DefaultSpentAt` — `SpentAt` нулевой в вызове usecase, `_InvalidJSON`, `_FractionalAmount`, `_InvalidAmount`, `_InvalidCategory`, `_UnknownMember`, `_InternalError`), `get_expense_test.go` (`_Success`, `_InvalidID` — `/api/expenses/abc`, `_NotFound`, `_InternalError`), `update_expense_test.go` (`_Success`, `_EmptyComment`, `_InvalidID`, `_InvalidJSON`, `_InvalidAmount`, `_InvalidCategory`, `_InvalidDate`, `_NotFound`, `_InternalError`), `delete_expense_test.go` (`_Success` 204 без тела, `_InvalidID`, `_NotFound`, `_InternalError`), `list_members_test.go` (`_Success`, `_Empty`, `_InternalError`), `new_router_test.go` (`_PathWithoutPrefix` → 404, `_CORSPreflightAllowed`, `_CORSForeignOrigin` — нет `Access-Control-Allow-Origin`). Проверка: `go test ./internal/transport/http/...` зелёный.

## 6. Запуск приложения

- [x] 6.1 `cmd/app/main.go`: в `run` собрать `repository.New(pool)` → `usecase.New(repo)` → `httpapi.NewRouter(httpapi.New(uc, logger), cfg.HTTP.AllowedOrigins)`; `net.ListenConfig.Listen(ctx, "tcp", ":<port>")` (ошибка → `fmt.Errorf("lc.Listen: %w", err)`), лог `http server started` с адресом, `srv.Serve(ln)` в горутине с `ReadHeaderTimeout` 5s; `select` на `ctx.Done()` и ошибку `Serve` (кроме `http.ErrServerClosed`); на сигнал — `srv.Shutdown` с таймаутом 10s, затем `app stopped`. Проверка: `go build ./...`, `go vet ./...`; `make run` с рабочей БД — `curl -i localhost:8080/api/expenses` отдаёт 200 `[]` или список, `curl -i localhost:8080/api/expenses/abc` — 400 JSON `invalid_request`; `kill -TERM` — записи об остановке и код выхода 0; запуск при занятом порте — `error` в логе и код выхода 1.

## 7. Итоговая проверка

- [x] 7.1 `gofmt -l .` пуст, `go vet ./...`, `go test ./...`, `make lint` зелёные; `go generate ./...` повторно не даёт diff (`git status --porcelain` показывает только изменения этой задачи).
- [x] 7.2 `docker compose up --build`: создать члена семьи вручную в БД (`INSERT INTO users`), через `curl` пройти `POST` → `GET /api/expenses?user_id=…&category=food` → `PUT` → `DELETE` → повторный `DELETE` (404 `not_found`), `GET /api/members`. Проверка: коды и тела совпадают со спецификацией.

## Workflow follow-up

- После ревью — `/opsx:archive`: дельты `rest-api`, `app-bootstrap`, `expense-repository` вливаются в основные спеки, создаётся `openspec/specs/expense-rules/spec.md`.
