# Tasks

## 1. Зависимость

- [x] 1.1 Добавить `github.com/openai/openai-go/v3` v3.81.0 (`go get github.com/openai/openai-go/v3@v3.81.0`). Удалить `internal/agent/.gitkeep` и `internal/llm/.gitkeep`. Проверка: модуль в блоке `require` `go.mod`, `go mod tidy` не даёт diff после использования в 4.x.

## 2. Конфиг

- [x] 2.1 `internal/config/config.go`: константа `envLLMTimeout = "LLM_TIMEOUT"`, дефолт `30 * time.Second`, поле `LLM.Timeout time.Duration`, парсинг `time.ParseDuration` с проверкой `> 0` и ошибкой `LLM_TIMEOUT: invalid value "<v>": must be a positive duration like 30s`, ошибка собирается в общий `errors.Join`; `LogValue` дополнен `llm_timeout`. В `.env.example` — `LLM_TIMEOUT=30s` с комментарием (optional, default 30s). Проверка: в `config_test.go` кейс «all values set» включает `LLM_TIMEOUT=45s` → `Timeout` 45s; кейсы — дефолт 30s при пустом значении, ошибки для `30`, `abc`, `0s`, `-5s` (сообщение содержит `LLM_TIMEOUT` и значение), `TestLogValueHidesSecrets` видит `llm_timeout`; `go test ./internal/config/...` зелёный.

## 3. Контракт в агенте

- [x] 3.1 `internal/agent/message.go` (`Role` + константы `RoleSystem`, `RoleUser`, `RoleAssistant`, `RoleTool`; `Message`; `ToolCall`) и `internal/agent/tool.go` (`Tool`) по design.md, решение 1. `internal/agent/deps.go`: `LLMClient { CreateChatCompletion(ctx context.Context, messages []Message, tools []Tool) (Message, error) }` с директивой `//go:generate go tool mockgen -source=./deps.go -destination ../../mock/agent.go -package mock`. Проверка: `go generate ./...` создаёт `mock/agent.go` с `MockLLMClient`, остальные файлы в `mock/` без diff, `go build ./...`.
- [x] 3.2 `internal/entity/errors.go`: добавить `ErrEmptyCompletion`. Проверка: `go build ./...`.

## 4. Клиент LLM

- [x] 4.1 `internal/llm/client.go`: `Client{client openai.Client; model string; timeout time.Duration}`, `New(cfg config.LLM) *Client` с `option.WithBaseURL`, `option.WithAPIKey`, `option.WithMaxRetries(2)`. Проверка: `go build ./...`.
- [x] 4.2 `internal/llm/create_chat_completion.go`: `CreateChatCompletion` по design.md, решение 2 — `context.WithTimeout(ctx, c.timeout)`; отображение ролей (неизвестная → ошибка до запроса); `Tools` остаётся `nil` при пустом списке; ошибка SDK → `fmt.Errorf("c.client.Chat.Completions.New: %w", err)`; пустой `Choices` → голый `entity.ErrEmptyCompletion`; из `Choices[0].Message` — `Content` и вызовы `Type == "function"`, пустые `Arguments` → `{}`. Без логирования, без приватных хелперов. Проверка: `go build ./...`, `go vet ./...`; соответствие `agent.LLMClient` проверяется строкой `var _ agent.LLMClient = (*llm.Client)(nil)` в `llm_test.go` (4.3).
- [x] 4.3 Тесты: `internal/llm/llm_test.go` — `LLMSuite` в пакете `llm_test` с `httptest.Server` (подменяемый обработчик, запоминает тела и число запросов), `client` через `llm.New(config.LLM{BaseURL: server.URL, APIKey: "secret-test-key", Model: "test-model", Timeout: …})`, `TestLLM`. `create_chat_completion_test.go`: `_Request` (путь `/chat/completions`, `Authorization: Bearer secret-test-key`, `model` = `test-model`, `tools[0]` типа `function` с именем, описанием, схемой), `_NoTools` (в теле нет `tools`), `_History` (assistant с `tool_calls` `call_1` и следующий `tool` с `tool_call_id` `call_1`), `_ToolCall` (разобраны `ID`, `Name`, `Arguments`), `_Text` (текст, пустые `ToolCalls`), `_EmptyArguments` (`""` → `{}`), `_EmptyChoices` (`s.Equal(entity.ErrEmptyCompletion, err)`), `_UnknownRole` (ошибка, ноль запросов), `_RetryThenSuccess` (503, 503, 200 с `Retry-After-Ms: 1` → успех, 3 запроса), `_RetryExhausted` (всегда 503 → ошибка после 3 запросов, `errors.AsType[*openai.Error]` со статусом 503, `s.Contains(err.Error(), "c.client.Chat.Completions.New")`), `_Unauthorized` (401 с телом, содержащим `secret-test-key` → 1 запрос, в `err.Error()` есть `401` и нет ключа), `_InvalidJSON` (200 с телом `not json` → ошибка), `_Timeout` (`Timeout: 100ms`, обработчик ждёт `r.Context().Done()` → `s.ErrorIs(err, context.DeadlineExceeded)`, вызов вернулся меньше чем за 1 с), `_Canceled` (контекст отменён во время ожидания → `s.ErrorIs(err, context.Canceled)`). Проверка: `go test ./internal/llm/...` зелёный и не обращается в сеть за пределы `127.0.0.1`.

## 5. Ручная проверка

- [x] 5.1 `cmd/llmcheck/main.go` по design.md, решение 6: `config.Load()` → `llm.New(cfg.LLM)` → вызов с system, user `кофе 350` и инструментом `add_expense`; печать `base_url`, `model`, каждого вызова (`id`, `name`, `arguments`) или текста; код 0 только при вызове инструмента с валидным JSON (`json.Valid`). `Makefile`: цель `llm-check` (проверка `.env`, `set -a && . ./.env && set +a && go run ./cmd/llmcheck`), добавлена в `.PHONY`. Проверка: `go build ./...`, `go vet ./...`; с заглушечным ключом `make llm-check` печатает ошибку с `401`, без значения ключа, код 1.
- [x] 5.2 Вписать в `.env` настоящие `LLM_API_KEY` и `LLM_MODEL` (сейчас там заглушки — это делает пользователь) и выполнить `make llm-check`: напечатан хотя бы один вызов `add_expense` с непустым `id` и аргументами-JSON вида `{"amount":350,…}`, код 0. Если провайдер отдаёт вызов с отклонениями (аргументы объектом, пустой `id`, вызов в тексте) — зафиксировать фактический ответ и завести отдельную задачу на нормализацию с тестом. Проверка: вывод команды приложен к отчёту по задаче.

## 6. Итоговая проверка

- [x] 6.1 `gofmt -l .` пусто, `go vet ./...`, `go test ./...`, `make lint` зелёные; повторный `go generate ./...` не даёт diff; `go mod tidy` не даёт diff.
