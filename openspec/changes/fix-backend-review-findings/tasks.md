# Tasks

## 1. Контракт API (openapi → generate)

- [x] 1.1 В `api/openapi.yaml`: параметры `limit` (1–500, default 100) и `offset` (≥0, default 0) у `listExpenses`, заголовок `X-Total-Count` в ответе 200, коды `invalid_pagination` и `invalid_comment` в `ErrorCode`, `maxLength: 200` у `comment`. Проверка: `go generate ./...` проходит
- [x] 1.2 `go generate ./...` и `make generate-web`, закоммитить сгенерированные `openapi.gen.go` и `schema.gen.ts`. Проверка: `go build ./...` падает только на нереализованных методах `StrictServerInterface`, `git diff --stat` показывает только `*.gen.*`

## 2. Слой данных: страница и атомарный патч

- [x] 2.1 Миграция `00003_add_expenses_category_index.sql` (up: индекс `(category, spent_at DESC)`, down: `DROP INDEX`). Проверка: `make migrate-up` и `make migrate-down` на чистой БД проходят, `\d expenses` показывает индекс
- [x] 2.2 `entity`: `ExpensePage{Items, Count, Total}`, `Limit`/`Offset` в `ExpenseFilter`, `ExpensePatch`, `ErrInvalidPagination`, `ErrInvalidComment`, константы лимитов. Проверка: `go build ./internal/entity/...`
- [x] 2.3 Репозиторий `ListExpenses` и `ListExpensesForUser`: страница + количество + сумма одним запросом (D1). Тесты для каждого: без фильтров, фильтры, границы периода, 45 трат с `limit 30 offset 30`, смещение за пределами (пустая страница, `Count` и `Total` корректны), пустая выборка, ошибки `Query`/`Scan`; для `ForUser` — A не получает траты B. Проверка: `go test ./internal/database/repository/...`
- [x] 2.4 Репозиторий `UpdateExpense` и `UpdateExpenseForUser` принимают `ExpensePatch` и делают один `UPDATE ... COALESCE` (D3). Тесты: изменение одного поля, пустой комментарий, несуществующая трата, нарушение ограничений, A не меняет трату B, параллельное изменение разных полей двумя горутинами не теряет ни одно (если в проекте нет интеграционного слоя — проверка вручную в 9.2). Проверка: `go test ./internal/database/repository/...`

## 3. Usecase

- [x] 3.1 `ListExpenses` и `ListExpensesForUser` проверяют категорию, период и параметры страницы (`ErrInvalidPagination`). Тесты: лимит 0/501/500, смещение -1/0, неверный период, неизвестная категория. Проверка: `go test ./internal/usecase/...`
- [x] 3.2 `CreateExpense`, `UpdateExpense`, `UpdateExpenseForUser`: проверка длины комментария (`ErrInvalidComment`), принимают `ExpensePatch` там, где это изменение. Тесты: комментарий 200/201 рун (кириллица), пустой патч, нулевая сумма в патче, чужая трата → `ErrNotFound`, ошибка репозитория оборачивается. Проверка: `go test ./internal/usecase/...`
- [x] 3.3 Обновить `deps.go` репозитория/usecase и перегенерировать моки (`go generate ./...`). Проверка: `go build ./...` проходит

## 4. Инструменты агента

- [x] 4.1 `list_expenses`: вызывать `ListExpensesForUser` с `Limit=30`; `Count`/`Total` из страницы, `Truncated = Count > 30`. Тесты: 1000 трат → запрошена одна страница из 30, сумма всех; изоляция A/B; ошибки usecase. Проверка: `go test ./internal/agent/tools/...`
- [x] 4.2 `update_expense` без предварительного чтения: собрать `ExpensePatch` из переданных аргументов, неизвестная категория → «другое» (как раньше), комментарий >200 рун → ошибка инструмента. `add_expense`: заменить сравнение строк с `time.Now().Format(...)` на сравнение календарных дат, поведение «сегодня = текущий момент» сохраняется (D8), комментарий >200 → ошибка. Тесты: частичные патчи, изоляция A/B для add/update/delete, дата сегодня/вчера/без даты. Проверка: `go test ./internal/agent/tools/...`

## 5. Агент: сбой после инструментов и очистка истории

- [x] 5.1 `HandleMessage`: запоминать результаты изменяющих инструментов; при ошибке модели или `ErrAgentStepLimit` после них записывать в историю сообщение и реплику «выполнено ...», возвращать детерминированный ответ и обёрнутую ошибку (D4). Дополнить системный промпт правилом о повторе. Тесты: сбой после `add_expense` (запись в историю, ответ с суммой и датой, возвращённая ошибка), сбой после `list_expenses` (история не меняется), лимит шагов после `update_expense`, изоляция A/B историй, промпт содержит правило. Проверка: `go test ./internal/agent/...`
- [x] 5.2 `history.AppendMessagesForUser`: периодическая очистка карты не чаще раза в `ttl` (D6). Тесты: пользователь, ушедший 16 минут назад, удалён при записи другого; активный не удаляется; повторная очистка внутри `ttl` не выполняется. Проверка: `go test ./internal/agent/history/...`

## 6. Telegram-транспорт

- [x] 6.1 Тип сериализации по `userID` с учётом `ctx` и освобождением записей (D5), использовать в `HandleUpdate`. Тесты: два сообщения одного пользователя идут последовательно (порядок вызовов агента), разные пользователи — параллельно, отмена `ctx` при ожидании, карта пуста после обработки. Проверка: `go test -race ./internal/transport/telegram/...`
- [x] 6.2 Лимит 1000 рун: длинное сообщение отбивается до `RegisterMember` и агента. Непустой ответ вместе с ошибкой агента отправляется пользователю, ошибка логируется. Тесты: 1000 и 1001 руна, частичный сбой, сбой отправки, паника. Проверка: `go test ./internal/transport/telegram/...`

## 7. HTTP-транспорт

- [x] 7.1 `ListExpenses`: параметры `limit`/`offset`, заголовок `X-Total-Count`, `invalid_pagination`; `CreateExpense`/`UpdateExpense` — `invalid_comment`, `PUT` собирает полный патч. Тесты: все сценарии rest-api из delta (лимит по умолчанию, вторая страница, смещение за пределами, 0/501/-1, комментарий 200/201). Проверка: `go test ./internal/transport/http/...`
- [x] 7.2 CORS `ExposedHeaders: X-Total-Count`, `http.MaxBytesHandler(1 MiB)`. Тесты: тело 5 МиБ → 400 `invalid_request`, заголовок в ответе. Проверка: `go test ./internal/transport/http/...`

## 8. Приложение и деплой

- [x] 8.1 `cmd/app/main.go`: `import _ "time/tzdata"`, таймауты `http.Server` (5/10/30/60 с), `TZ=Mars/Olympus` → UTC и `warn`. Проверка: `go test ./...`
- [x] 8.2 Обновить README (пагинация). Проверка: `docker compose config` валиден
- [x] 8.3 Проверить tzdata в образе: запуск с `TZ=Europe/Moscow` показывает `Europe/Moscow` в логе старта. Проверка: запись в логе `timezone`

## 9. Веб-панель и сквозная проверка

- [x] 9.1 `web`: типы из `make generate-web`, чтение `X-Total-Count`, параметр страницы в адресе, размер страницы 50, кнопки «Назад/Вперёд», подпись «51–100 из 120», сброс страницы при смене фильтров, переход назад при удалении последней строки страницы, запасной режим без заголовка (полная страница = есть следующая). Unit-тесты на чистую логику по примеру существующих. Проверка: `npm --prefix web test`, `npm --prefix web run build`
- [ ] 9.2 `gofmt -l .` пуст, `go vet ./...`, `go test -race ./...`. Запуск `docker compose up --build`: 150 трат через REST → `GET /api/expenses` отдаёт 100 и `X-Total-Count: 150`; параллельные `PUT` и сообщение боту не затирают поля. Проверка: результаты зафиксированы в описании PR

## Workflow follow-up

- Заархивировать change (`/opsx:archive`) после ревью; убедиться, что main-спеки `rest-api`, `expense-repository`, `expense-rules`, `expense-agent`, `telegram-bot`, `app-bootstrap`, `web-panel` обновились.
