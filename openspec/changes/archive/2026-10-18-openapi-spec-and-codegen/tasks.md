# Tasks

## 1. Спецификация `api/openapi.yaml`

- [x] 1.1 Описать `components/schemas`: `Category` (enum `food`, `transport`, `housing`, `entertainment`, `health`, `shopping`, `other`, русские названия в `description`), `Expense` (все поля `required`, `amount` int64, `updated_at` `nullable: true`, `spent_at`/`created_at`/`updated_at` — `date-time`), `ExpenseCreate` (`required: [user_id, amount, category]`, `amount` с `minimum: 1`, `spent_at` с описанием «по умолчанию — текущий момент»), `ExpenseUpdate` (`required: [amount, category, spent_at]`, `comment` необязателен, без `user_id`), `Member` (`telegram_id` int64, `username`, `first_name`, `last_name` — обязательные строки), `ErrorCode` (enum `invalid_request`, `invalid_amount`, `invalid_category`, `unknown_member`, `not_found`, `internal`) и `Error` (`code`, `message`). Проверка: каждое поле и enum совпадают с `specs/rest-api/spec.md` и `entity.Category`.
- [x] 1.2 Описать `components/responses` `BadRequest`, `NotFound`, `InternalError` (схема `Error`) и `components/parameters` `ExpenseID` (path, int64). Проверка: на них ссылаются все операции из 1.3.
- [x] 1.3 Описать `servers: [{url: /api}]` и пути: `GET /expenses` (`listExpenses`, query `user_id` int64, `from`/`to` `date-time` с описанием «включительно» / «не включительно», `category` → `$ref Category`; 200 — массив `Expense`, 400, 500), `POST /expenses` (`createExpense`, 201 `Expense`, 400, 500), `GET /expenses/{id}` (`getExpense`, 200, 400, 404, 500), `PUT /expenses/{id}` (`updateExpense`, 200, 400, 404, 500), `DELETE /expenses/{id}` (`deleteExpense`, 204 без тела, 400, 404, 500), `GET /members` (`listMembers`, 200 — массив `Member`, 500). 400 у операций `/expenses/{id}` — на некорректный `id` в пути. Проверка: шесть операций с уникальными `operationId`, у каждого ответа есть `description`.

## 2. Подключение генератора

- [x] 2.1 Выполнить `go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen` и `go get github.com/oapi-codegen/runtime`. Проверка: в `go.mod` есть директива `tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen`, `go tool oapi-codegen -version` печатает зафиксированную версию.
- [x] 2.2 Создать `internal/transport/http/openapi/oapi-codegen.yaml` (`package: openapi`, `generate: {models: true, chi-server: true, strict-server: true}`, `output: openapi.gen.go`) и `internal/transport/http/openapi/generate.go` (`package openapi` и строка `//go:generate go tool oapi-codegen -config oapi-codegen.yaml ../../../../api/openapi.yaml`, без других комментариев). Проверка: файлы на месте, других файлов, кроме перечисленных в design.md, в пакете нет.
- [x] 2.3 Добавить в `Makefile` цель `generate` (`go generate ./...`) и внести её в `.PHONY`. Проверка: `make generate` проходит.

## 3. Генерация и сборка

- [x] 3.1 Запустить `make generate`, затем `go mod tidy`. Проверка: появился `internal/transport/http/openapi/openapi.gen.go` с шапкой `Code generated ... DO NOT EDIT`, в нём есть `StrictServerInterface` с методами `ListExpenses`, `CreateExpense`, `GetExpense`, `UpdateExpense`, `DeleteExpense`, `ListMembers`, типы `Expense`, `ExpenseCreate`, `ExpenseUpdate`, `Member`, `Category`, `Error`, `ErrorCode`, `Expense.UpdatedAt` имеет тип `*time.Time`, `Amount` — `int64`. `chi/v5` и `oapi-codegen/runtime` — прямые зависимости в `go.mod`.
- [x] 3.2 Проверить воспроизводимость: закоммитить (или зафиксировать в индексе) результат, повторно запустить `go generate ./...`. Проверка: `git status --porcelain` пуст, `go mod tidy` не даёт diff.
- [x] 3.3 Проверка: `go build ./...`, `go vet ./...` и `docker compose build app` проходят, в логе сборки образа нет вызова генератора.

## 4. Валидация спецификации

- [x] 4.1 Добавить `github.com/getkin/kin-openapi` как прямую зависимость и `internal/transport/http/openapi/spec_test.go` (пакет `openapi_test`, testify `Suite`, `TestOpenAPI` с `suite.Run`, кейс `TestSpec_Valid`: `openapi3.NewLoader().LoadFromFile("../../../../api/openapi.yaml")` и `doc.Validate(ctx)` без ошибок). Проверка: `go test ./internal/transport/http/openapi/...` зелёный.
- [x] 4.2 Убедиться, что тест ловит поломку: временно заменить в спецификации один `$ref` на несуществующую схему, запустить `go test ./internal/transport/http/openapi/...` (ожидается падение), вернуть правку. Проверка: тест падал на сломанной спецификации и проходит на исходной, `git diff api/openapi.yaml` пуст.

## 5. Документация и итоговая проверка

- [x] 5.1 В `AGENTS.md` добавить в дерево структуры `internal/transport/http/openapi/` (сгенерированный код и конфиг генератора), в «Команды» — `make generate`. Проверка: описанные пути и команды существуют и работают.
- [x] 5.2 Итоговая проверка: `gofmt -l .` пуст, `go vet ./...` и `go test ./...` зелёные, повторный `go generate ./...` не меняет файлы, `openspec validate openapi-spec-and-codegen --strict` проходит.

## Workflow follow-up

- После ревью архивировать изменение (`/opsx:archive`): появятся основные спеки `rest-api` и `api-codegen`.
