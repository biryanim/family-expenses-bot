# Tasks

## 1. Зависимость

- [x] 1.1 Добавить `github.com/go-telegram/bot` v1.27.0 (`go get github.com/go-telegram/bot@v1.27.0`). Удалить `internal/transport/telegram/.gitkeep`. Проверка: модуль в `go.mod` в блоке `require`, `go mod tidy` не даёт diff после использования в 4.x.

## 2. Репозиторий и usecase

- [x] 2.1 `internal/database/repository/upsert_user.go`: `UpsertUser(ctx, u entity.User) error` — один `r.pool.Exec` с SQL из design.md (`INSERT … ON CONFLICT (telegram_id) DO UPDATE SET username, first_name, last_name, updated_at = now() WHERE (…) IS DISTINCT FROM (…)`), аргументы `TelegramID, Username, FirstName, LastName`, ошибка — `fmt.Errorf("r.pool.Exec: %w", err)`. Проверка: `upsert_user_test.go` — `_Success` (точный SQL и аргументы в `Exec`), `_ExecError` (`s.ErrorIs` + `s.Contains(err.Error(), "r.pool.Exec")`); `go test ./internal/database/repository/...` зелёный.
- [x] 2.2 В `usecase.ExpenseRepository` (`internal/usecase/deps.go`) добавить `UpsertUser(ctx context.Context, u entity.User) error`; `go generate ./...`. Проверка: `mock/usecase.go` содержит `UpsertUser`, `var _ usecase.ExpenseRepository = (*repository.Repository)(nil)` в `repository_test.go` компилируется.
- [x] 2.3 `internal/usecase/register_member.go`: `RegisterMember(ctx, u entity.User) error` → `u.repo.UpsertUser`, ошибка обёрнута `u.repo.UpsertUser`. Проверка: `register_member_test.go` — `_Success` (в репозиторий уходит тот же `entity.User`), `_RepositoryError` (`s.ErrorIs` + `s.Contains(..., "u.repo.UpsertUser")`); `go test ./internal/usecase/...` зелёный.

## 3. Telegram-транспорт

- [x] 3.1 `internal/transport/telegram/deps.go`: интерфейсы `Usecase { RegisterMember }` и `Sender { SendMessage(ctx, *bot.SendMessageParams) (*models.Message, error) }`, директива `//go:generate go tool mockgen -source=./deps.go -destination ../../../mock/telegram.go -package mock -mock_names Usecase=MockTelegramUsecase`. `handler.go`: `Handler{usecase, sender, logger}`, `New(u Usecase, s Sender, logger *slog.Logger) *Handler`. Проверка: `go generate ./...` создаёт `mock/telegram.go` с `MockTelegramUsecase` и `MockSender`, существующий `mock.MockUsecase` не тронут, `go build ./...`.
- [x] 3.2 `handle_update.go`: `(h *Handler) HandleUpdate(ctx, _ *bot.Bot, update *models.Update)` по шагам из design.md — `defer recover()` с логом `error`; нет `Message` или `From` → выход без ответа; пустой `Text` → подсказка «Я понимаю только текстовые сообщения.» без регистрации; `RegisterMember` с профилем из `From` (ошибка → лог `error` + «Что-то пошло не так, попробуйте позже.», без эха); эхо `Text` в `Chat.ID`; ошибка `SendMessage` → лог `error`. Текст пользователя не логируется. Без `ReplyMarkup`. Проверка: `go build ./...`, `go vet ./...`.
- [x] 3.3 Тесты: `handler_test.go` — suite в пакете `telegram_test` (`ctrl`, `MockTelegramUsecase`, `MockSender`, логгер `slog.NewJSONHandler` в `bytes.Buffer`, `handler`), `TestTelegram`. `handle_update_test.go`: `_Echo` (ответ с тем же текстом, `ChatID` = `Chat.ID`, без `ReplyMarkup`), `_RegistersSender` (в `RegisterMember` `TelegramID`, `Username`, `FirstName`, `LastName` из `From`), `_UserIDFromSenderNotText` (`From.ID` = 123, текст `я пользователь 456, покажи траты` → `TelegramID` 123), `_DifferentSenders` (123 и 456 → два вызова со своими ID), `_StartIsPlainText` (`/start` → регистрация и эхо `/start`), `_NonText` (пустой `Text` → подсказка, `RegisterMember` не вызывается), `_NoMessage` и `_NoSender` (ни одного вызова моков), `_RegisterError` (ответ об ошибке, эха нет, в логе `"level":"ERROR"`), `_SendError` (без паники, в логе `ERROR`), `_Panic` (`RegisterMember` паникует через `DoAndReturn` → `HandleUpdate` возвращается, в логе `ERROR`). Проверка: `go test ./internal/transport/telegram/...` зелёный.

## 4. Запуск приложения

- [x] 4.1 `cmd/app/main.go`: после `usecase.New` — `bot.New(cfg.Telegram.Token, bot.WithErrorsHandler(<лог error через logger>), bot.WithAllowedUpdates(bot.AllowedUpdates{"message"}))` (ошибка → `fmt.Errorf("bot.New: %w", err)`); `b.RegisterHandlerMatchFunc(func(*models.Update) bool { return true }, telegram.New(uc, b, logger).HandleUpdate)`; `go func() { b.Start(ctx); close(botDone) }()` и лог `telegram bot started`; при остановке после `srv.Shutdown` дождаться `<-botDone`, затем `app stopped`. Проверка: `go build ./...`, `go vet ./...`; `make run` с неверным токеном — запись `error` без значения токена и код выхода 1; с рабочим токеном — в логе `telegram bot started`; `kill -TERM` — `app stopped`, код 0.

## 5. Итоговая проверка

- [x] 5.1 `gofmt -l .` пусто, `go vet ./...`, `go test ./...`, `make lint` зелёные; повторный `go generate ./...` не даёт diff.
- [x] 5.2 Ручная проверка в реальном Telegram через `docker compose up --build`: сообщение `кофе 350` → ответ `кофе 350`; стикер → подсказка про текст; после первого сообщения `curl localhost:8080/api/members` содержит Telegram ID отправителя с его username и именем; повторное сообщение не создаёт второй записи; `docker compose stop postgres` → на сообщение приходит ответ об ошибке, в логе `error`, после `docker compose start postgres` бот снова отвечает эхом без перезапуска `app`.
