# Tasks

## 1. Хранилище истории

- [x] 1.1 `internal/agent/deps.go`: добавить интерфейс `ConversationStore` с методами `ListMessagesForUser(userID int64, now time.Time) []Message` и `AppendMessagesForUser(userID int64, now time.Time, messages ...Message)` (design.md, решение 2). Запустить `go generate ./...`, в `mock/agent.go` должен появиться `MockConversationStore`. Проверка: `go build ./...`.
- [x] 1.2 Пакет `internal/agent/history`. В `history.go`: тип `History` с полями `mu sync.Mutex` и `dialogs map[int64][]entry`, тип `entry{message agent.Message; at time.Time}`, `New()`, константы `maxMessages = 6` и `ttl = 15 * time.Minute`. Методы `list_messages_for_user.go` и `append_messages_for_user.go` по design.md, решение 2: копия под мьютексом, отбрасывание записей с `now.Sub(at) > ttl`, обрезка до последних `maxMessages`. Комментариев нет. Проверка: `go vet ./internal/agent/...`.
- [x] 1.3 Тесты `history_test` (`Suite` в `history_test.go`, `TestHistory`). Файл `list_messages_for_user_test.go`: `_Empty`, `_ReturnsAppendedInOrder`, `_Expired` (запись 16 минут назад не возвращается, 15 минут — возвращается), `_OtherUser` (после записи для `userB` у `userA` пусто), `_ReturnsCopy` (изменение результата не меняет хранилище). Файл `append_messages_for_user_test.go`: `_KeepsLastSix` (после 4 пар остаются 6 последних, первая — реплика пользователя), `_DropsExpired`, `_OtherUserUntouched`, `_Concurrent` (параллельная запись из горутин для разных пользователей без потерь). Проверка: `go test -race ./internal/agent/history/...`.

## 2. История в агенте

- [x] 2.1 `internal/agent/agent.go`: поле `history ConversationStore`, сигнатура `New(llm LLMClient, history ConversationStore, tools ...ToolExecutor)`. `cmd/app/main.go`: передать `history.New()`. `internal/agent/agent_test.go`: в `SetupTest` передавать свежий `history.New()` в `s.agent` и `s.isolated`. Проверка: `go build ./...`, существующие тесты `go test ./internal/agent/...` зелёные.
- [x] 2.2 `internal/agent/handle_message.go`: собирать `messages` как `[system] + ListMessagesForUser(userID, now) + [user]`, после непустого финального ответа вызывать `AppendMessagesForUser(userID, now, user, assistant)`. На ветках ошибок не писать (design.md, решение 3). Проверка: `go test ./internal/agent/...`.
- [x] 2.3 Тесты в `handle_message_test.go`:
  - `_HistoryPassedToModel`: второе сообщение получает `[system, user1, assistant1, user2]`.
  - `_HistoryNotSavedOnError`: первое сообщение завершилось `CreateChatCompletion` с ошибкой, второе получает `[system, user2]`. То же для `ErrEmptyAgentReply`.
  - `_HistoryIsolatedBetweenUsers`: после диалога `userB` с уникальным текстом сообщение `userA` получает `[system, userA text]`, ни одно сообщение не содержит текст B.
  - `_HistoryExcludesToolMessages`: после диалога с вызовом инструмента в истории следующего запроса нет `RoleTool` и `ToolCalls`.
  Проверка: `go test ./internal/agent/...`.

## 3. Правила промпта и сценарии неоднозначных случаев

Сценарии из пункта 3.2 идут на `s.isolated` со строгим `s.repo`: ожидаются только перечисленные обращения, любое лишнее (особенно `UpdateExpenseForUser`/`DeleteExpenseForUser`/`CreateExpense`) роняет тест (design.md, решение 5). Сообщения модели для второго шага захватываются через `DoAndReturn`.

- [x] 3.1 `internal/agent/agent.go`: удалить из `systemPrompt` правила «…выбери самую позднюю…» и «Уточняющих вопросов не задавай…», добавить пять правил из design.md, решение 4. В правило про ответ на уточнение добавить сверку выбранной траты с суммой и датой из вопроса (design.md, Risks). `TestHandleMessage_SystemPrompt`: добавить `s.Contains` для фраз «ничего не меняй и не удаляй», «спроси сумму», «записал на сегодня», «не вызывай инструменты» и `s.NotContains` для «самую позднюю» и «Уточняющих вопросов не задавай». Проверка: `go test ./internal/agent/...`.
- [x] 3.2 Сценарии в `handle_message_test.go`, по одному методу на случай:
  - `_DeleteLastExpense`: `удали последнее` → `get_last_expense` → `delete_expense`. `s.repo`: `GetLastExpenseForUser(_, userA)` возвращает трату с ранней `SpentAt`, но наибольшим `id`, затем `GetExpenseForUser` и `DeleteExpenseForUser(_, userA, этот id)`.
  - `_DeleteLastExpenseNone`: `GetLastExpenseForUser` → `ErrNotFound`, tool-сообщение `{"error":"expense not found"}`, удалений нет, текст ответа передан пользователю.
  - `_AmbiguousDeleteAsks`: `удали такси` → `list_expenses` возвращает два такси (500 ₽ 15.10 и 700 ₽ 17.10) → модель отвечает вопросом. На `s.repo` только `ListExpensesForUser(_, userA, _)`.
  - `_AmbiguousDeleteThenChoice`: два сообщения подряд. Второе (`второе`) получает в истории вопрос бота. Модель вызывает `list_expenses`, затем `delete_expense` с `id` такси от 15.10. `s.repo` ожидает `DeleteExpenseForUser(_, userA, id такси 15.10)` и не ожидает удаления такси 17.10.
  - `_AmbiguousUpdateAsks`: `кофе было 400` при двух кофе → вопрос, `UpdateExpenseForUser` не ожидается.
  - `_DeleteNothingMatches`: `удали такси` → `list_expenses` возвращает пусто (`count 0`) → текстовый ответ, удалений нет.
  - `_UpdateNothingMatches`: `такси было 900` → пустой список → ответ, ни `UpdateExpenseForUser`, ни `CreateExpense` не ожидаются.
  - `_MissingAmountAsks`: `кофе` → модель сразу отвечает вопросом, на `s.repo` ожиданий нет.
  - `_MissingAmountThenAmount`: после `кофе` и вопроса приходит `350`. В истории есть `кофе` и вопрос. Модель вызывает `add_expense` с `{"amount":350,"category":"food","comment":"кофе"}`, `s.repo.CreateExpense` получает `UserID == userA`, `Amount == 35000`, `Category == food`.
  - `_UnrecognizedDate`: `кофе 350 на днях` → `add_expense` без `date`. `CreateExpense` получает `SpentAt` в пределах секунды от `time.Now()` (`s.WithinDuration`), ответ модели передаётся без изменений.
  - `_NotAboutExpenses`: `расскажи анекдот` → модель отвечает текстом, на `s.repo` ожиданий нет, ответ передан пользователю.
  Проверка: `go test ./internal/agent/...`.

## 4. README

- [x] 4.1 Создать `README.md` (design.md, решение 6): что делает сервис, запуск через `docker compose up --build` с `.env.example`, раздел «Неоднозначные случаи» по каждому пункту задачи 10. Отдельный абзац: нераспознанная дата записывается на сегодня с предупреждением, потому что добавление обратимо, и как её исправить. Описать короткую историю диалога (6 реплик, 15 минут, теряется при перезапуске). Проверка: каждый из шести пунктов задачи есть в README, команды из README совпадают с `Makefile` и `docker-compose.yml`.

## 5. Итоговая проверка

- [x] 5.1 `gofmt -l .` пусто, `go vet ./...`, `go test -race ./...`, `make lint` зелёные, `go generate ./...` не даёт diff. В новом коде нет комментариев, кроме `//go:generate`. `openspec validate ambiguous-cases --strict` проходит.
- [x] 5.2 Реальный диалог: `docker compose up --build` с настоящими `LLM_*`, сценарии из `specs/expense-agent/spec.md`:
  - `такси 500` за 15.10 и `такси 700` за 17.10, затем `удали такси` → вопрос с обоими вариантами, ничего не удалено; затем `второе` → удалено названное вторым.
  - `кофе` → вопрос о сумме; `350` → трата добавлена.
  - `кофе 350 на днях` → записано на сегодня с предупреждением.
  - `удали самокат` → «не нашёл».
  - `удали последнее` → удалена последняя созданная.
  - `расскажи анекдот` → короткий ответ без новых трат.
  - Второй пользователь сразу пишет `второе` → бот не видит чужой диалог и переспрашивает.

  Проверять по веб-панели, что удалено и что добавлено. Если модель отвечает плохо, подправить формулировку правил и повторить. Проверка: переписка приложена к отчёту. Результат: прогон на `nvidia/nemotron-3-ultra-550b-a55b:free` (прежняя `nvidia/nemotron-3-super-120b-a12b:free` отвечала 404 со стороны провайдера), пользователь подтвердил, что всё работает; переписка не приложена.

## Workflow follow-up

- Желательно заархивировать `expense-user-isolation` до архивации этого change.
- После ревью и прохождения 5.2 — `/opsx:archive`.
