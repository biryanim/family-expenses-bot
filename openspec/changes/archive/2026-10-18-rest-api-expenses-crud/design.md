# Design

## Context

Мотивация — см. proposal.md, требования — `specs/*/spec.md` этого изменения.

Текущее состояние:
- `internal/transport/http/openapi/openapi.gen.go` — сгенерированные модели, `StrictServerInterface` (`ListExpenses`, `CreateExpense`, `GetExpense`, `UpdateExpense`, `DeleteExpense`, `ListMembers`), `NewStrictHandlerWithOptions`, `HandlerWithOptions` под chi. По умолчанию ошибки разбора параметров (`ChiServerOptions.ErrorHandlerFunc`) и тела (`StrictHTTPServerOptions.RequestErrorHandlerFunc`) отдаются через `http.Error` — `text/plain`, что не совпадает с контрактом.
- Сгенерированный код не проверяет `enum` и `minimum`: `openapi.Category` — просто `string`, `amount` — `int64`. Обязательные поля тела без указателей при отсутствии декодируются в нулевые значения. Дробное `amount` и невалидный JSON дают ошибку декодирования.
- `internal/usecase` содержит только `deps.go` с интерфейсом `ExpenseRepository`. Реализации, часов и тестов нет.
- Репозиторий: CRUD, per-user методы и `ListUserIDs`. Профили членов семьи не отдаются. Доменные sentinel-ошибки — `ErrNotFound`, `ErrUnknownUser`, `ErrInvalidAmount`, `ErrInvalidCategory`; репозиторий возвращает их и по ограничениям БД.
- `cmd/app/main.go` подключается к БД и ждёт сигнала, HTTP-сервера нет. `config.HTTP.Port` уже есть.
- В `go.mod` есть chi, oapi-codegen runtime, kin-openapi (для теста спецификации), testify, gomock.

## Goals / Non-Goals

**Goals:**
- Хендлеры — тонкий перевод `openapi` ↔ `entity` и ошибок в ответы. Все правила — в `usecase`, бот потом использует те же методы или их per-user двойники.
- Каждый ответ API, включая ошибки разбора, соответствует `api/openapi.yaml`; это проверяется в тестах по самой спецификации.

**Non-Goals:**
- Per-user методы `usecase` для бота (`…ForUser`) — в задаче бота. Правила оттуда повторят проверки этого изменения (повтор кода допустим по `AGENTS.md`).
- Middleware валидации запросов по спецификации (`kin-openapi` `openapi3filter` на входе) — проверки делает `usecase`, дублировать на транспорте не нужно (как и решено в задаче 04).
- Регистрация членов семьи — их создаёт бот.
- nginx-прокси и фронтенд.

## Decisions

### Слой `usecase`

`usecase.Usecase` с конструктором `New(repo ExpenseRepository)`. В `internal/usecase/deps.go` в `ExpenseRepository` появляется `ListUsers` вместо `ListUserIDs`. Отдельного интерфейса часов нет: дату по умолчанию `usecase` берёт из `time.Now()`, а тест проверяет, что она попадает между моментами до и после вызова. Абстракция ради одного вызова не окупается.

Методы, по файлу на каждый:
- `CreateExpense(ctx, e entity.Expense)` — проверка `Amount > 0` → `ErrInvalidAmount`, `Category.Valid()` → `ErrInvalidCategory`; нулевой `SpentAt` → `time.Now()`; `repo.CreateExpense`. `ErrUnknownUser` приходит из репозитория по FK — отдельной проверки члена семьи не делаем: лишний запрос и гонка между проверкой и вставкой.
- `GetExpense(ctx, id)`, `DeleteExpense(ctx, id)` — прямой вызов репозитория.
- `ListExpenses(ctx, f entity.ExpenseFilter)` — непустая невалидная категория → `ErrInvalidCategory`; `From` и `To` заданы и `From.After(To)` → `ErrInvalidDate`.
- `UpdateExpense(ctx, e entity.Expense)` — сумма, категория, нулевой `SpentAt` → `ErrInvalidDate`.
- `ListMembers(ctx)` — `repo.ListUsers`.

Любая ошибка репозитория, включая доменные sentinel (`ErrNotFound`, `ErrUnknownUser`, …), оборачивается именем упавшего вызова — `fmt.Errorf("u.repo.CreateExpense: %w", err)`, без отдельной ветки для sentinel. Голым sentinel возвращается только там, где он рождается (репозиторий, проверки самого `usecase`). Хендлер различает ошибки через `errors.Is`.

Валидация до репозитория, хотя у БД есть те же CHECK-ограничения: ограничения БД — последняя линия обороны, а `usecase` даёт ошибку без похода в базу и одинаково для бота.

Почему `ErrInvalidDate` при пустой дате изменения, а не «оставить прежнюю»: `PUT` — полная замена, `spent_at` в `ExpenseUpdate` обязателен. Сгенерированная модель не отличает «не передан» от нулевого времени, поэтому нулевое время трактуем как отсутствующее.

`from > to` — ошибка, а не пустой список: почти наверняка перепутаны границы в UI, молчаливый пустой ответ скроет баг. `from == to` — валидный пустой полуинтервал.

Альтернатива для `invalid_date` — отдавать `invalid_request`. Отказались: `invalid_request` означает «запрос не разобран», а тут запрос корректен синтаксически и нарушает доменное правило — как `invalid_amount`. Панели нужен отдельный ключ для сообщения у поля даты.

### Контракт: `invalid_date`

`api/openapi.yaml`: в `ErrorCode.enum` добавляется `invalid_date`. Затем `go generate ./...` — появится константа `openapi.InvalidDate`. Это единственная правка контракта.

### Транспорт `internal/transport/http`

Пакет `http` (как каталог). Внутри импорт `net/http` работает без конфликта; в `cmd/app` пакет импортируется под алиасом `httpapi`.

- `deps.go` — интерфейс `Usecase` с шестью методами `usecase.Usecase` и `//go:generate go tool mockgen -source=./deps.go -destination ../../../mock/http.go -package mock`.
- `server.go` — `type Server struct { usecase Usecase; logger *slog.Logger }`, `New(u Usecase, logger *slog.Logger) *Server`, compile-time проверка `var _ openapi.StrictServerInterface = (*Server)(nil)`.
- По файлу на метод: `list_expenses.go`, `create_expense.go`, `get_expense.go`, `update_expense.go`, `delete_expense.go`, `list_members.go`. Каждый сам переводит `openapi` → `entity`, вызывает `usecase`, маппит ошибку через `switch` на sentinel'ах в свой типизированный ответ (`CreateExpense400JSONResponse{BadRequestJSONResponse{Code: openapi.InvalidAmount, …}}` и т.д.) и переводит результат обратно. Общего хелпера маппинга нет — правило `AGENTS.md`, повтор допустим.
- Хендлер никогда не возвращает `error` в strict-обёртку: на неизвестную ошибку он сам логирует её (`logger.ErrorContext` с операцией) и возвращает `…500JSONResponse` с `code=internal` и общим `message`. Так соблюдаются «кто возвращает — не логирует» (хендлер не возвращает) и «500 без деталей».
- `handle_request_error.go` — `HandleRequestError(w, r, err)`: 400, `application/json`, `{"code":"invalid_request","message":<err.Error()>}`. Подставляется и в `ChiServerOptions.ErrorHandlerFunc` (параметры пути/query), и в `StrictHTTPServerOptions.RequestErrorHandlerFunc` (тело). Текст ошибки разбора безопасно показывать: он про ввод клиента, не про внутренности.
- `handle_response_error.go` — метод `(*Server).HandleResponseError`: 500 `{"code":"internal",…}`. Срабатывает только при сбое кодирования ответа; логирует через логгер `Server`, поэтому это метод, а не свободная функция.
- `new_router.go` — `NewRouter(s *Server, allowedOrigins []string) http.Handler` (логгер берётся из `Server`, отдельный параметр был бы лишним): `chi.NewRouter()`, `middleware.Recoverer`, CORS, затем `openapi.HandlerWithOptions(openapi.NewStrictHandlerWithOptions(s, nil, …), openapi.ChiServerOptions{BaseURL: "/api", BaseRouter: r, ErrorHandlerFunc: HandleRequestError})`. `NotFound`/`MethodNotAllowed` оставляем дефолтными chi: спецификация не описывает их тела.

Почему фильтр `user_id` идёт в `entity.ExpenseFilter.UserID` указателем без проверки существования: неизвестный член семьи в фильтре — просто пустой результат, это поведение из `rest-api` («Нет подходящих трат»).

Перевод времени: `entity` → `openapi` без смены часового пояса (как вернул pgx). `ExpenseCreate.SpentAt == nil` → нулевое `time.Time` в entity, дефолт ставит `usecase`. `Comment == nil` → `""`.

### CORS: `github.com/go-chi/cors`

Новая зависимость (v1.2.x, последняя на момент планирования — v1.2.2). Альтернатива — свой middleware на ~30 строк. Отказались: preflight, `Vary: Origin`, сравнение origin и кэширование preflight легко сделать неправильно, а go-chi/cors — тот же мейнтейнер, что и chi, без транзитивных зависимостей. Настройки: `AllowedOrigins` из конфига, `AllowedMethods: GET, POST, PUT, DELETE, OPTIONS`, `AllowedHeaders: Content-Type`, `AllowCredentials: false`, `MaxAge: 300`.

Дефолт `http://localhost:5173`, а не `*`. Авторизации нет, поэтому `*` позволил бы любой открытой в браузере странице читать и удалять траты семьи через `localhost:8080`. В compose фронтенд ходит через nginx на тот же origin, CORS там не нужен; список нужен только для dev-сервера Vite и нестандартных установок.

### Конфиг

`config.HTTP.AllowedOrigins []string` из `CORS_ALLOWED_ORIGINS`: split по `,`, `TrimSpace`, пустые элементы выбрасываются; если в итоге пусто — `[]string{"http://localhost:5173"}`. Без валидации URL: неверный origin просто ни с чем не совпадёт. В `LogValue` добавляется `cors_allowed_origins` (не секрет). `.env.example` — новая строка с комментарием.

### Запуск и graceful shutdown в `cmd/app`

В `run` после подключения к БД:
```
repo := repository.New(pool)
uc := usecase.New(repo)
handler := httpapi.NewRouter(httpapi.New(uc, logger), cfg.HTTP.AllowedOrigins)
srv := &http.Server{Addr: ":<port>", Handler: handler, ReadHeaderTimeout: 5s}
```
`srv.ListenAndServe` в горутине, ошибка (кроме `http.ErrServerClosed`) — в канал. `select` на `ctx.Done()` и канал: ошибка запуска → `run` возвращает `fmt.Errorf("srv.ListenAndServe: %w", err)` (лог и exit 1 делает `main`); сигнал → `srv.Shutdown` с `context.WithTimeout(context.Background(), 10*time.Second)`, ошибка шатдауна возвращается из `run`. Чтобы «порт занят» ловился до записи «сервер запущен», используем `net.ListenConfig.Listen(ctx, …)` + `srv.Serve(ln)` (линтер `noctx` требует вариант с контекстом): ошибка `lc.Listen` возвращается синхронно, лог `http server started` пишется после успешного `Listen`.


### Тесты

- `usecase`: suite в `usecase_test` с `MockExpenseRepository` (мок генерируется в `mock/usecase.go`), файл на метод, кейсы `_Success`, `_InvalidAmount`, `_InvalidCategory`, `_InvalidDate`, `_DefaultSpentAt`, `_NotFound`, `_RepositoryError` и т.д. Для невалидного ввода — `EXPECT` на репозиторий не ставится, gomock упадёт при вызове.
- `transport/http`: suite в `http_test` с `MockUsecase`. Запросы идут через `httptest` в настоящий `NewRouter` — так проверяются пути с `/api`, разбор параметров, коды и тела ошибок. Каждый ответ дополнительно проверяется по `api/openapi.yaml` через `openapi3filter.ValidateResponse` с роутером `routers/legacy` из kin-openapi (уже в `go.mod`; `routers/gorillamux` потянул бы `gorilla/mux`) — это закрывает критерий «ответы совпадают со спецификацией» без новой зависимости. Валидация ответа — через `openapi3filter` внутри теста, а не общий хелпер в продовом коде; в тестах suite-метод для валидации допустим.
- Комбинации фильтров — таблица с `s.Run`: query-строка → ожидаемый `entity.ExpenseFilter` в `EXPECT().ListExpenses(gomock.Any(), want)`.
- `repository`: `list_users_test.go` вместо `list_user_ids_test.go`.
- `config`: кейсы `CORS_ALLOWED_ORIGINS`.

## Risks / Trade-offs

- [Обязательные поля тела без указателей: отсутствующий `user_id` превращается в `0`] → `0` не существует как Telegram ID, FK даёт `unknown_member`; отсутствующие `amount`/`category` дают `invalid_amount`/`invalid_category`. Ответ 400 корректен, хотя код ошибки описывает следствие, а не «поле не передано». Принимаем: отдельная проверка required требует request-validator middleware.
- [Дефолтные 404/405 chi отдают `text/plain`] → спецификация описывает только 404 для существующих путей с несуществующим `id`; на неизвестные пути формат не обещан.
- [`ListUserIDs` удаляется] → потребителей в коде нет (проверено), бот в будущем возьмёт `ListUsers`.
- [Нет request-validator: расхождение спецификации и кода по enum/minimum возможно] → тесты хендлеров с невалидным вводом и валидация ответов по спецификации.

## Migration Plan

Миграций БД нет. Развёртывание — `docker compose up --build`; порт уже проброшен. Откат — предыдущий образ.
