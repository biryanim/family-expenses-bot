# Design

## Context

- `internal/agent` и `internal/llm` пустые (`.gitkeep`). Конфиг уже читает `LLM_BASE_URL`, `LLM_API_KEY`, `LLM_MODEL` в `config.LLM`, но никто их не использует.
- `AGENTS.md` (уже обновлён в рабочей копии) фиксирует: SDK `github.com/openai/openai-go/v3`, ретраи SDK, таймаут через `context`, интерфейс клиента — в `agent/deps.go`, реализация — в `llm`.
- В кэше модулей есть `openai-go/v3@v3.81.0`; решения ниже сверены с его API: `openai.NewClient(opts...)`, `client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams)`, `openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam)`, `ChatCompletionMessageParamUnion{OfAssistant: …}` с `ToolCalls []ChatCompletionMessageToolCallUnionParam{OfFunction: …}`, `openai.ToolMessage(content, toolCallID)`, ответ — `ChatCompletion.Choices[i].Message.ToolCalls[j].Function.{Name,Arguments}`, ошибка API — `*openai.Error` (алиас `apierror.Error`).
- `*openai.Error.Error()` в этой версии выдаёт только `OpenAI API error: <код> <текст статуса>` — без URL, заголовков и тела. Ключ в текст ошибки не попадает.
- `openai.NewClient` сначала применяет `DefaultClientOptions()` — переменные `OPENAI_BASE_URL`, `OPENAI_API_KEY`, `OPENAI_ORG_ID`, `OPENAI_PROJECT_ID`, `OPENAI_CUSTOM_HEADERS` и т.п., затем наши опции. Наши `WithBaseURL` и `WithAPIKey` их перекрывают.
- В `.env` сейчас заглушки `LLM_API_KEY` и `LLM_MODEL`, для ручной проверки нужны настоящие значения.

## Goals / Non-Goals

**Goals:**
- Контракт клиента, по которому агента можно тестировать gomock-моком, не зная про SDK.
- Реализация на SDK, где провайдер и модель меняются только через env.
- Время одного обращения к LLM, включая повторы SDK, ограничено `LLM_TIMEOUT`.
- Повторяемая ручная проверка tool calling на реальной модели.

**Non-Goals:**
- Цикл агента, промпт, инструменты трат и тексты ответов пользователю — следующее изменение.
- Подключение клиента в `cmd/app/main.go`: потребителя ещё нет, а неиспользуемая зависимость в `main` — мёртвый код.
- Стриминг, structured outputs (`strict`), параллельные вызовы как настройка, учёт токенов.
- Свои ретраи, circuit breaker, rate limiting.

## Decisions

### 1. Свои типы сообщений в `agent`, а не типы SDK в интерфейсе

`internal/agent/deps.go`:

```go
//go:generate go tool mockgen -source=./deps.go -destination ../../mock/agent.go -package mock
type LLMClient interface {
	CreateChatCompletion(ctx context.Context, messages []Message, tools []Tool) (Message, error)
}
```

Типы — в `internal/agent/message.go` и `internal/agent/tool.go`:
- `Role string` с константами `RoleSystem`, `RoleUser`, `RoleAssistant`, `RoleTool`;
- `Message{Role, Content string; ToolCalls []ToolCall; ToolCallID string}`;
- `ToolCall{ID, Name, Arguments string}` — `Arguments` строкой JSON, разбирает её конкретный инструмент;
- `Tool{Name, Description string; Parameters map[string]any}` — JSON Schema параметров.

`llm` импортирует `agent` ради этих типов; `agent` не импортирует `llm` (их связывает `main`), цикла нет.

Альтернатива — интерфейс прямо на `openai.ChatCompletionNewParams` / `*openai.ChatCompletion`. Отвергнута: агент и его тесты привязываются к SDK (тест строит громоздкие union-типы SDK), а разбор отклонений провайдеров расползается в агент. Цена своих типов — одно отображение в `llm`, и оно всё равно нужно для нормализации.

Типы не кладутся в `entity`: это не доменные сущности, а протокол общения агента с моделью.

### 2. Реализация `internal/llm`

- `client.go`: `Client{client openai.Client; model string; timeout time.Duration}`, `New(cfg config.LLM) *Client` → `openai.NewClient(option.WithBaseURL(cfg.BaseURL), option.WithAPIKey(cfg.APIKey), option.WithMaxRetries(2))`. `WithMaxRetries(2)` задаётся явно, хотя совпадает с дефолтом SDK: спека требует ровно 2 повтора, и смена дефолта в будущей версии SDK не должна менять поведение молча.
- `create_chat_completion.go`: `(c *Client) CreateChatCompletion(ctx, messages, tools) (agent.Message, error)`:
  1. `ctx, cancel := context.WithTimeout(ctx, c.timeout)`, `defer cancel()`. Срок общий для всех попыток SDK. `option.WithRequestTimeout` не используется — он ограничивает одну попытку, а нам нужен предел ожидания пользователя.
  2. Отображение сообщений по `Role`: system → `openai.SystemMessage`, user → `openai.UserMessage`, tool → `openai.ToolMessage(content, ToolCallID)`, assistant → `ChatCompletionMessageParamUnion{OfAssistant: &ChatCompletionAssistantMessageParam{…}}`, где текст ставится только если не пуст, а `ToolCalls` — `OfFunction` с `ID`, `Function.Name`, `Function.Arguments`. Неизвестная роль → `fmt.Errorf("unknown message role %q", role)` до запроса.
  3. Инструменты → `openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{Name, Description: openai.String(…), Parameters: shared.FunctionParameters(t.Parameters)})`. При пустом списке `Tools` остаётся `nil` и поле не уходит в запрос (`omitzero`) — некоторые провайдеры отвергают `tools: []`.
  4. `c.client.Chat.Completions.New(ctx, params)`, ошибка → `fmt.Errorf("c.client.Chat.Completions.New: %w", err)`.
  5. Пустой `Choices` → голый `entity.ErrEmptyCompletion` (новый sentinel в `internal/entity/errors.go`, все sentinel-ы проекта живут там).
  6. Из `Choices[0].Message` берутся `Content` и `ToolCalls`; берутся только вызовы с `Type == "function"` (custom-инструменты мы не объявляем). `Arguments`, пустые после `strings.TrimSpace`, заменяются на `{}`.
- Отображение делается внутри метода, без приватных хелперов — по правилам `AGENTS.md`.

### 3. Распознавание таймаута и ошибок провайдера

Ошибка оборачивается через `%w`, цепочка не рвётся: `errors.Is(err, context.DeadlineExceeded)` / `context.Canceled` работают, `errors.AsType[*openai.Error](err)` доступен тому, кому нужен статус. Отдельных sentinel-ов «таймаут» / «ошибка провайдера» не вводим: агенту достаточно `context.DeadlineExceeded`, чтобы ответить «модель не ответила вовремя», и любой другой ошибки — для «сервис недоступен». Клиент ничего не логирует: возвращает ошибку — значит, не логирует (`AGENTS.md`).

### 4. `LLM_TIMEOUT` в конфиге

`config.LLM` получает `Timeout time.Duration`. Парсинг — `time.ParseDuration`, значение должно быть > 0, пустое → `30s`. Ошибка в стиле существующих: `LLM_TIMEOUT: invalid value "<v>": must be a positive duration like 30s`. `LogValue` дополняется `llm_timeout`. 30 секунд: модель с tool calling обычно отвечает за единицы секунд, бюджет покрывает 2 повтора с бэкоффом SDK (0.5 с, 1 с) и не держит пользователя Telegram дольше разумного. Альтернатива — таймаут на весь апдейт в Telegram-транспорте; это решение агентского изменения (одно сообщение пользователя может дать несколько обращений к LLM), здесь ограничивается одно обращение.

### 5. Тесты клиента — против `httptest.Server`

Unit-тесты `internal/llm` поднимают `httptest.Server` и передают его URL как `BaseURL` — работает настоящий SDK (сериализация, ретраи, разбор), но к провайдеру запросов нет. Suite `LLMSuite` в пакете `llm_test`; моков нет, в `SetupTest` создаётся сервер с подменяемым обработчиком. Чтобы тесты повторов не ждали бэкофф SDK, сервер отвечает заголовком `Retry-After-Ms: 1` (SDK его учитывает). Таймаут проверяется с `Timeout: 100ms` и обработчиком, который ждёт `r.Context().Done()`.

Альтернатива — `option.WithHTTPClient` с подменённым транспортом. Отвергнута: `httptest.Server` проще и проверяет реальный HTTP, включая путь `/chat/completions` и заголовок `Authorization`.

Тесты агента в будущем — на `mock.MockLLMClient`.

### 6. Ручная проверка — `cmd/llmcheck`

`cmd/llmcheck/main.go`: `config.Load()` → `llm.New(cfg.LLM)` → один вызов с system-сообщением, user-сообщением `кофе 350` и инструментом `add_expense` (`amount` number, `category` string, `description` string, все required). Печатает `base_url`, `model`, затем каждый вызов `id`, `name`, `arguments` или текст ответа. Код 0 — есть хотя бы один вызов инструмента с валидным JSON в аргументах, иначе 1. Ключ не печатается. `make llm-check` запускает её с `.env` так же, как `make run`.

Почему команда, а не тест с build-тегом: правило «реальных запросов в тестах нет», а команда остаётся в репозитории для проверки при смене провайдера (критерий «смена через конфиг»). `config.Load()` требует и остальные обязательные переменные — в `.env` они есть, отдельный загрузчик не вводим.

## Risks / Trade-offs

- [Совместимый провайдер отдаёт `arguments` объектом, а не строкой, или не отдаёт `id` вызова] → SDK не разберёт поле в строку. На OpenAI такого нет; проверяем на реальной модели через `make llm-check`. Если всплывёт — нормализация добавляется в `create_chat_completion.go` отдельной задачей с тестом, по фактическому ответу, а не наугад.
- [`OPENAI_ORG_ID` / `OPENAI_PROJECT_ID` / `OPENAI_CUSTOM_HEADERS` в окружении процесса молча добавят заголовки] → в compose и `.env.example` их нет; base URL и ключ всё равно берутся из `LLM_*`. Принимаем.
- [Повторы SDK съедают бюджет таймаута] → это и нужно: пользователь ждёт не дольше `LLM_TIMEOUT` суммарно; последний повтор может оборваться по сроку.
- [`llm` импортирует `agent`] → типы протокола объявлены потребителем вместе с интерфейсом; это осознанная плата за отсутствие SDK-типов в агенте.
- [Модель вернула невалидный JSON в `arguments`] → клиент его не проверяет и отдаёт как есть: решение, что делать (вернуть модели ошибку инструмента), — за агентом.

## Migration Plan

Новая необязательная переменная с дефолтом — существующие `.env` продолжают работать. Откат — revert изменения; БД и API не затронуты.
