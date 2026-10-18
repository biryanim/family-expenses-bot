# Design

## Context

- `internal/agent` содержит только контракт с моделью: `LLMClient`, `Message`, `ToolCall`, `Tool` (задача 07). `internal/llm.Client` его реализует, но в `cmd/app/main.go` не подключён.
- Репозиторий уже умеет операции в пределах члена семьи: `GetExpenseForUser`, `ListExpensesForUser`, `UpdateExpenseForUser`, `DeleteExpenseForUser` — они объявлены в `usecase.ExpenseRepository`, но методов usecase над ними нет. «Последней траты» нет ни на одном уровне.
- `usecase.CreateExpense` берёт `UserID` из переданной траты и ставит `time.Now()`, если дата пустая. Валидация суммы и категории живёт там же и возвращает голые sentinel-ы. Ошибки репозитория usecase оборачивает через `%w`, поэтому `errors.Is(err, entity.ErrNotFound)` работает.
- `telegram.Handler.HandleUpdate` берёт `userID` из `msg.From.ID`, регистрирует члена семьи и отправляет эхо. Это эхо и заменяем.
- Образ приложения — `gcr.io/distroless/static-debian12`, процесс работает в UTC. `.env` передаётся в `app` целиком через `env_file`.
- Ограничения `AGENTS.md`: интерфейсы — в `deps.go` потребителя; один метод — один файл; приватных хелперов нет, методы друг друга не вызывают; комментариев нет; ошибки оборачиваются именем упавшего вызова; моки — gomock в `mock/`.

## Goals / Non-Goals

**Goals:**
- Цикл агента, который тестируется моками `LLMClient` и инструментов, без сети.
- Каждый инструмент тестируется отдельно моком usecase, включая изоляцию пользователя.
- Деньги нигде не проходят через `float64`.

**Non-Goals:**
- Стриминг ответа, индикатор «печатает…», параллельное выполнение вызовов инструментов.
- Поиск по тексту комментария в БД: модель выбирает трату из результата `list_expenses`.
- Изменения REST API и веб-панели.

## Decisions

### 1. Инструменты — отдельный пакет `internal/agent/tools`, цикл видит их через интерфейс

`internal/agent/deps.go` дополняется:

```go
type ToolExecutor interface {
	Definition() Tool
	Execute(ctx context.Context, userID int64, arguments string) (any, error)
}
```

`Execute` возвращает Go-значение результата, а не JSON-строку: в JSON его сериализует цикл (`json.Marshal`, ошибка → `fmt.Errorf("json.Marshal: %w", err)`). Иначе каждой ветке ошибки в каждом инструменте понадобился бы свой `json.Marshal` со своей обработкой ошибки, а тесты инструментов разбирали бы JSON вместо сравнения структур. Ошибка, которую должна увидеть модель, — значение `agent.ToolError{Error string}` (`internal/agent/tool.go`), сериализуется как `{"error":"…"}`.

`Agent{llm LLMClient; tools []ToolExecutor}`, `New(llm LLMClient, tools ...ToolExecutor) *Agent`. Цикл ищет исполнитель по `Definition().Name` и вызывает `Execute` с `userID`, который получил сам. `userID` в инструмент передаёт только код агента — у модели нет канала, чтобы его задать.

`internal/agent/tools` — пять типов: `AddExpense`, `ListExpenses`, `GetLastExpense`, `UpdateExpense`, `DeleteExpense`. У каждого: `<tool>.go` (структура с полем `usecase Usecase` и конструктор `New<Tool>(uc Usecase)`), `<tool>_definition.go` (`Definition()`), `<tool>_execute.go` (`Execute()`). `tools/deps.go` объявляет `Usecase` только с нужными методами, мок — `-mock_names Usecase=MockToolsUsecase` (имя `MockUsecase` уже занято моком HTTP). `tools` импортирует `agent` ради типа `Tool`; `agent` не импортирует `tools`, их связывает `main`.

Почему так, а не методы `Agent` по инструменту со `switch` в цикле: цикл, вызывающий `a.AddExpense(...)`, нарушает правило «методы не вызывают друг друга», а единый метод со всеми инструментами внутри получится на сотни строк, и его не протестировать по частям. Интерфейс `ToolExecutor` развязывает цикл и инструменты: цикл тестируется моком `ToolExecutor`, инструменты — моком usecase.

Отступление от «Конструктор — `New`»: в пакете пять типов, поэтому `NewAddExpense` и т.д. Альтернатива — пакет на инструмент — даёт пять пакетов по три файла и пять `deps.go` с одинаковыми интерфейсами. Отвергнута.

### 2. Цикл агента — `HandleMessage`

`internal/agent/handle_message.go`: `(a *Agent) HandleMessage(ctx context.Context, userID int64, text string) (string, error)`.

1. `now := time.Now()`; системный промпт — `fmt.Sprintf(systemPrompt, now.Format("2006-01-02"), weekday)`, где `weekday` берётся из локального массива русских названий по `now.Weekday()`. Константа `systemPrompt` и `maxSteps = 5` — в `agent.go`.
2. Из `a.tools` собираются `[]Tool` для запроса и `map[string]ToolExecutor` по имени.
3. История: `system`, `user`. До `maxSteps` раз: `a.llm.CreateChatCompletion` (ошибка → `fmt.Errorf("a.llm.CreateChatCompletion: %w", err)`). Нет `ToolCalls` → вернуть `strings.TrimSpace(Content)`; пустой текст → голый `entity.ErrEmptyAgentReply`. Есть вызовы → ответ ассистента в историю, затем для каждого вызова по порядку: неизвестное имя → результат `ToolError{Error: "unknown tool"}`; иначе `Execute(ctx, userID, call.Arguments)` — ошибка → `fmt.Errorf("executor.Execute: %w", err)`, обработка прекращается; результат → `json.Marshal` → сообщение `tool` с `ToolCallID`.
4. Шаги исчерпаны → голый `entity.ErrAgentStepLimit`.

Новые sentinel-ы — в `internal/entity/errors.go`. Агент ничего не логирует: ошибку логирует Telegram-транспорт.

Память диалога не хранится (решение пользователя): история живёт только внутри одного `HandleMessage`. Хранилища состояния нет, рестарт ничего не ломает, конкурентные апдейты от одного пользователя не делят общее состояние.

### 3. Системный промпт

Текст на русском, черновик (финальная формулировка уточняется на реальном диалоге в задаче 8.2):

```
Ты помощник семейного учёта расходов в Telegram. Сегодня %s, %s.
Категории (код — название): food — еда, transport — транспорт, housing — жильё,
entertainment — развлечения, health — здоровье, shopping — покупки, other — другое.
Правила:
- Сообщение о трате («кофе 350», «такси 1200 вчера») — вызови add_expense. Сумма в рублях.
  Категорию выбери сам; если не подходит ни одна — other. В comment — что купили, 1–3 слова.
- Даты передавай как YYYY-MM-DD. Без даты — сегодня. «Вчера» — сегодня минус день.
  «В этом месяце» — с 1-го числа по сегодня. «За неделю» — последние 7 дней включая сегодня.
- Вопросы о тратах — list_expenses; сумму бери из поля total, сам не складывай.
- «Последняя/последнее» — get_last_expense, затем update_expense или delete_expense по id.
- Изменение или удаление по описанию («удали такси») — list_expenses, выбери самую позднюю
  подходящую трату. Если подходящей нет — так и скажи. Уточняющих вопросов не задавай.
- Отвечай по-русски одним-двумя предложениями, без Markdown. Суммы — «350 ₽»,
  категории — русскими названиями. Называй дату и сумму изменённой или удалённой траты.
- Не о тратах — коротко скажи, что ведёшь учёт расходов.
```

Почему дата и день недели в промпте, а не инструмент «который час»: на каждый относительный запрос это экономит шаг цикла, а день недели помогает модели с «в понедельник».

### 4. Сумма: рубли строкой → копейки

Схема: `amount` — `{"type": "number"}`. Аргументы разбираются через `json.Decoder` с `UseNumber()` и `DisallowUnknownFields()` в структуру с полем `Amount json.Number`. `json.Number` принимает и число, и строку с числом, если модель пришлёт `"350"`. Дальше `entity.ParseAmount(string) (int64, error)` (`internal/entity/parse_amount.go`): целая часть и не более двух цифр после точки, без экспоненты и знака, результат > 0 и без переполнения `int64`; иначе голый `entity.ErrInvalidAmount`. Обратно — `entity.FormatAmount(int64) string` (`format_amount.go`): `35000` → `"350"`, `19990` → `"199.90"`.

Функции живут в `entity`: перевод рублей в копейки — доменное правило денег, и оно нужно и в `add_expense`, и в `update_expense`, и в форматировании результатов всех инструментов. Это экспортируемые доменные функции, а не приватные хелперы.

Альтернатива — модель сама передаёт `amount_kopecks` целым. Отвергнута: умножение на 100 модель иногда делает с ошибкой, а ошибка в 100 раз хуже отказа.

### 5. Схемы и поведение инструментов

| Инструмент | Параметры (required жирным) | Usecase | Результат |
|---|---|---|---|
| `add_expense` | **amount**, **category** (enum кодов), **comment**, date | `CreateExpense` с `UserID: userID` | трата |
| `list_expenses` | from, to, category (enum) | `ListExpensesForUser` | `count`, `total`, `expenses` (≤30), `truncated` |
| `get_last_expense` | — | `GetLastExpenseForUser` | трата |
| `update_expense` | **id**, amount, category, date, comment | `GetExpenseForUser` → слияние → `UpdateExpenseForUser` | трата после изменения |
| `delete_expense` | **id** | `GetExpenseForUser` → `DeleteExpenseForUser` | удалённая трата |

- Трата в результате: `{"id":7,"amount":"350","category":"food","date":"2026-10-18","comment":"кофе"}`, дата — `SpentAt.In(time.Local)`.
- Категория в `add`/`update`: не `Valid()` или не передана → `entity.CategoryOther`. В `list_expenses` неизвестная категория уходит в usecase и даёт `ErrInvalidCategory` → ошибка модели. Подмена фильтра на `other` дала бы неправильный ответ на вопрос.
- Даты: `time.ParseInLocation("2006-01-02", s, time.Local)` — строгий формат, `2026-02-30` отвергается самим `time`. В `add_expense` дата, равная сегодняшней, превращается в нулевой `SpentAt`, и usecase ставит текущий момент. В `list_expenses` `to` превращается в начало следующего дня, потому что репозиторий фильтрует `spent_at < To`.
- `count` и `total` считаются по всему списку из usecase, `expenses` — первые 30 (список уже отсортирован от поздних к ранним). Сумма считается кодом: модель складывает ненадёжно.
- Доменные ошибки (`ErrNotFound`, `ErrInvalidAmount`, `ErrInvalidCategory`, `ErrInvalidDate`, проверка через `errors.Is`) и ошибки разбора аргументов → `(agent.ToolError{Error: "<текст>"}, nil)`. Модель видит причину и отвечает пользователю. Любая другая ошибка usecase → `fmt.Errorf("t.usecase.<Метод>: %w", err)` (получатель инструмента — `t`), цикл прерывается.
- Гонка «get → update» в `update_expense` и `delete_expense` не страшна: обе операции ограничены `userID`, и худший случай — «трата не найдена».

### 6. Usecase и репозиторий

Новые методы usecase, каждый в своём файле, с тестом: `GetExpenseForUser`, `ListExpensesForUser` (те же проверки категории и периода, что в `ListExpenses`), `UpdateExpenseForUser` (те же проверки, что в `UpdateExpense`), `DeleteExpenseForUser`, `GetLastExpenseForUser`. Проверки дублируются, а не выносятся, по правилу «повтор кода допустим».

Репозиторий: `GetLastExpenseForUser(ctx, userID)` — `SELECT … FROM expenses WHERE user_id = $1 ORDER BY id DESC LIMIT 1`, `pgx.ErrNoRows` → голый `entity.ErrNotFound`. Сортировка по `id`, а не по `created_at`: identity монотонен, а у `created_at` бывают равные значения. Индекс не нужен: выборка идёт по `expenses_user_id_spent_at_idx` с фильтром по `user_id`, трат у одного человека мало.

### 7. Telegram-транспорт и сборка

`telegram/deps.go`: `Agent interface { HandleMessage(ctx context.Context, userID int64, text string) (string, error) }` (мок `MockAgent`). `New(u Usecase, a Agent, s Sender, logger)`. В `HandleUpdate` эхо заменяется так: после `RegisterMember` вызывается `h.agent.HandleMessage(ctx, userID, msg.Text)`. Ошибка → лог `error` с `user_id` и ошибкой (без текста сообщения) и ответ `replyFailure`. Иначе отправляется ответ агента.

`cmd/app/main.go`: `llm.New(cfg.LLM)` → `agent.New(llmClient, tools.NewAddExpense(uc), tools.NewListExpenses(uc), tools.NewGetLastExpense(uc), tools.NewUpdateExpense(uc), tools.NewDeleteExpense(uc))` → `telegram.New(uc, ag, b, logger)`. В лог `app started` добавляется `timezone` (`time.Local.String()`), чтобы ошибка в `TZ` была видна сразу: при неизвестном значении Go молча берёт UTC.

### 8. Часовой пояс — стандартный `TZ`

Свой конфиг не вводим: Go сам читает `TZ` при старте и выставляет `time.Local`, а весь код уже работает через `time.Now()` и `time.Local`. В `.env.example` добавляется `TZ=Europe/Moscow` с комментарием, compose передаёт её через `env_file`. Если в distroless-образе нет `/usr/share/zoneinfo` (проверяется в задаче 7.2), сборка получает `-tags timetzdata`, и база поясов вшивается в бинарь.

### 9. Тесты

- `entity`: `parse_amount_test.go`, `format_amount_test.go` — таблицы в стиле `category_test.go`.
- `usecase`, `repository`: файл на новый метод, как у существующих.
- `agent`: suite с `MockLLMClient` и двумя `MockToolExecutor`. Кейсы: текстовый ответ; один вызов → результат в истории с `ToolCallID` → ответ; два вызова в одном ответе; неизвестный инструмент; лимит 5 шагов; ошибка LLM; ошибка `Execute` прерывает цикл; пустой ответ; `Execute` получает `userID` из аргумента `HandleMessage`; системное сообщение содержит сегодняшнюю дату и все коды категорий; в запрос уходят определения всех инструментов.
- `tools`: suite с `MockToolsUsecase`, по файлу тестов на `Execute` и `Definition` каждого инструмента. Обязательные кейсы изоляции: usecase получает `userID` из аргумента `Execute`; аргументы с `user_id` отклоняются без вызова usecase; `ErrNotFound` на чужой `id` даёт ошибку модели, `Update`/`Delete` не вызываются. Ещё: определения без пользовательских полей, разбор суммы, `travel` → `other`, даты, обрезка до 30 с полными `count` и `total`, сбой usecase → ошибка.
- `telegram`: эхо-тест заменяется проверкой ответа агента; агент получает `userA` при тексте «я пользователь 456»; ошибка агента → `replyFailure` и лог без текста; ошибка регистрации → агент не вызывается.

## Risks / Trade-offs

- [Модель выбирает не ту трату для «удали такси» или ошибается в датах] → правила и дата с днём недели в промпте, итоги считает код, в ответе называется сумма и дата изменённой траты. Проверка — реальный диалог (задача 8.2). Отката удаления нет, это осознанная цена отказа от подтверждений.
- [Без памяти нельзя продолжить диалог («а вчерашнее?»)] → решение пользователя. Промпт запрещает уточняющие вопросы, чтобы бот не обещал продолжения, которого не будет.
- [До 5 обращений к LLM на сообщение, каждое до `LLM_TIMEOUT`] → в худшем случае пользователь ждёт долго. Типичный путь — 2 обращения. Общий таймаут не вводим: отмену при остановке приложения даёт `ctx` бота.
- [Модель отвечает Markdown] → бот шлёт текст без `parse_mode`, звёздочки останутся видны, но сообщение дойдёт. Промпт запрещает Markdown.
- [`list_expenses` без периода по всей истории] → отдаём модели не больше 30 трат, итоги считаются по всем. Для семейного бота объём выборки из БД некритичен.
- [Неверное значение `TZ` даёт молчаливый UTC] → `timezone` в логе старта.
- [Отступление от «Конструктор — `New`» в `tools`] → см. решение 1. Если не устраивает, альтернатива — пакет на инструмент.

## Migration Plan

Миграций БД и изменений API нет. После деплоя бот вместо эха отвечает агентом. В существующий `.env` нужно добавить `TZ=Europe/Moscow`, иначе даты считаются по UTC. Откат — revert изменения.
