# Design

## Context

Мотивация — см. proposal.md, требования — `specs/*/spec.md` этого изменения.

Текущее состояние:
- `internal/transport/telegram/` пуст (`.gitkeep`). `config.Telegram.Token` уже читается из обязательной `TELEGRAM_BOT_TOKEN`.
- `cmd/app/main.go`: конфиг → пул БД → `usecase.New(repository.New(pool))` → HTTP-сервер, ожидание сигнала через `signal.NotifyContext`, graceful shutdown HTTP с таймаутом 10 с.
- Таблица `users (telegram_id PK, username, first_name, last_name, created_at, updated_at)` есть, но в неё никто не пишет; репозиторий умеет только `ListUsers`. `expenses.user_id` ссылается на `users.telegram_id` внешним ключом.
- `mock/` — один пакет `mock`. В нём уже есть `MockUsecase` (из `internal/transport/http/deps.go`).

Факты о `github.com/go-telegram/bot` v1.27.0, проверенные по исходникам модуля:
- `bot.New(token, opts...)` по умолчанию вызывает `getMe` (таймаут инициализации внутри библиотеки) и возвращает ошибку при отказе.
- `Start(ctx)` запускает long polling и воркеры и блокируется до отмены `ctx`. Сетевые ошибки `getUpdates` библиотека сама передаёт в `ErrorsHandler` и повторяет запрос с нарастающей паузой.
- Каждый апдейт обрабатывается хендлером в отдельной горутине (`go r(ctx, b, upd)`) **без `recover`** — паника в хендлере роняет весь процесс.
- `ErrorsHandler` и debug-хендлер по умолчанию пишут через стандартный `log`. Токен в URL ошибок `net/http` библиотека маскирует (`***`).
- `HandlerFunc` — `func(ctx context.Context, b *bot.Bot, update *models.Update)`; `*bot.Bot` — конкретный тип.

## Goals / Non-Goals

**Goals:**
- Транспорт отвечает только за перевод апдейта в `(user_id, профиль, текст)` и ответ в чат. Регистрация члена семьи — через `usecase`, к `database/*` транспорт не обращается.
- Обработчик тестируется без сети: Telegram-отправка и `usecase` подменяются gomock-моками.
- Точка расширения для задачи 08 очевидна: эхо в обработчике заменяется вызовом агента с тем же `user_id`.

**Non-Goals:**
- Агент, LLM, инструменты над тратами. Поэтому тест «A не видит/не меняет/не удаляет траты B» в этом изменении не появляется: бот здесь к тратам не обращается. Вместо него — тесты, что `user_id` берётся только из `from.id` и не зависит от текста; полноценный тест изоляции придёт с инструментами агента.
- Webhook, `allowed_updates` кроме `message`, редактирование сообщений, inline-режим, callback-и.
- Ожидание завершения уже начатых обработчиков при остановке (см. риски).

## Decisions

### Слой данных и `usecase`

- Репозиторий: `UpsertUser(ctx, u entity.User) error` в `upsert_user.go`, один `r.pool.Exec`:
  `INSERT INTO users (telegram_id, username, first_name, last_name) VALUES ($1, $2, $3, $4) ON CONFLICT (telegram_id) DO UPDATE SET username = EXCLUDED.username, first_name = EXCLUDED.first_name, last_name = EXCLUDED.last_name, updated_at = now() WHERE (users.username, users.first_name, users.last_name) IS DISTINCT FROM (EXCLUDED.username, EXCLUDED.first_name, EXCLUDED.last_name)`.
  `WHERE` в `DO UPDATE` даёт «`updated_at` только при реальном изменении» без отдельного `SELECT` и без гонки между проверкой и вставкой. Ошибка — `fmt.Errorf("r.pool.Exec: %w", err)`.
  Альтернатива `INSERT … ON CONFLICT DO NOTHING` (запомнить только при первом сообщении) отвергнута: в панели остались бы устаревшие имена после смены профиля в Telegram.
- Возвращаемого значения кроме ошибки нет: обработчику профиль не нужен, а `RETURNING` при сработавшем `WHERE` вернёт ноль строк и усложнит контракт.
- `usecase.ExpenseRepository` в `deps.go` получает `UpsertUser`. `usecase.RegisterMember(ctx, u entity.User) error` в `register_member.go` — вызов репозитория, ошибка оборачивается `u.repo.UpsertUser`. Доменной валидации нет: `telegram_id` приходит из Telegram, а не от пользователя.
- Регистрация при **каждом** текстовом сообщении, а не по признаку «первое»: признак пришлось бы где-то хранить или проверять отдельным запросом. Upsert идемпотентен и дешёвый (один запрос по первичному ключу).

### Пакет `internal/transport/telegram`

`deps.go`:
- `Usecase { RegisterMember(ctx, u entity.User) error }`.
- `Sender { SendMessage(ctx, params *bot.SendMessageParams) (*models.Message, error) }` — подмножество `*bot.Bot`, ради мока.
- Директива `//go:generate go tool mockgen -source=./deps.go -destination ../../../mock/telegram.go -package mock -mock_names Usecase=MockTelegramUsecase`. Имя `Usecase` повторяет соглашение `transport/http`, а `-mock_names` снимает коллизию с уже существующим `mock.MockUsecase`. Альтернатива — назвать интерфейс иначе (`MemberRegistrar`) — отвергнута: в задаче 08 сюда добавятся новые зависимости, и имя по одному методу устареет.

`handler.go`: `Handler { usecase Usecase; sender Sender; logger *slog.Logger }`, `New(u Usecase, s Sender, logger *slog.Logger) *Handler`.

`handle_update.go`: `(h *Handler) HandleUpdate(ctx, _ *bot.Bot, update *models.Update)` — сигнатура `bot.HandlerFunc`. Параметр `*bot.Bot` игнорируется, ответы идут через `h.sender`: так обработчик тестируется с `nil` вместо бота. Порядок:
1. `defer` с `recover()`: паника логируется `error` с `user_id`, если он уже известен, и проглатывается. Это единственное место, где её можно поймать — библиотека запускает хендлер в своей горутине без `recover`.
2. `update.Message == nil` или `update.Message.From == nil` → `return` молча.
3. `update.Message.Text == ""` → ответ-подсказка «Я понимаю только текстовые сообщения.», без регистрации. Фото с подписью тоже сюда: подпись — `Caption`, не `Text`.
4. `userID := update.Message.From.ID`; `RegisterMember(ctx, entity.User{TelegramID: userID, Username, FirstName, LastName})` из `From`. При ошибке — лог `error` и ответ «Что-то пошло не так, попробуйте позже.», дальше не идём.
5. Ответ `SendMessage{ChatID: update.Message.Chat.ID, Text: update.Message.Text}` — эхо. В задаче 08 эта строка станет вызовом агента с `userID` и текстом.
6. Ошибка любого `SendMessage` — лог `error`, без повтора.

Тексты ответов — константы в пакете. Текст сообщения пользователя в лог не пишется, только `user_id` и `chat_id`. Ответ отправляется в `Chat.ID`, а не в `From.ID`: в личке они совпадают, в группе ответ уйдёт в группу.

Обработчик нарушает «кто возвращает ошибку — не логирует» не по-настоящему: он ничего не возвращает и является конечной точкой, как хендлеры HTTP.

### Сборка в `cmd/app/main.go`

После создания `uc` и до HTTP-сервера:
- `b, err := bot.New(cfg.Telegram.Token, bot.WithErrorsHandler(...), bot.WithAllowedUpdates(bot.AllowedUpdates{"message"}))`; ошибка → `fmt.Errorf("bot.New: %w", err)`, `run` возвращает её, `main` логирует и выходит с 1 — как с БД. `getMe` не отключаем (`WithSkipGetMe` не используем): неверный токен должен ронять старт, а не молча давать неработающего бота.
- `WithErrorsHandler` — замыкание, пишущее `logger.Error("telegram bot error", ...)`, чтобы ошибки библиотеки шли в JSON-лог, а не в стандартный `log`.
- `b.RegisterHandlerMatchFunc(func(*models.Update) bool { return true }, telegram.New(uc, b, logger).HandleUpdate)`. Хендлер регистрируется после `bot.New`, потому что ему нужен `b` как `Sender`; `WithDefaultHandler` требует хендлер до создания бота. Обработчик по всем апдейтам заодно гарантирует, что дефолтный хендлер библиотеки (печатает апдейт целиком через `log`) не вызывается.
- `go func() { b.Start(ctx); close(botDone) }()` и `logger.Info("telegram bot started")`.
- При остановке: после `srv.Shutdown` ждём `<-botDone`. `ctx` из `signal.NotifyContext` уже отменён, `Start` возвращается сам. На ветке ошибки `srv.Serve` `defer stop()` отменит `ctx` и бот тоже остановится.

### Тесты

- `internal/database/repository/upsert_user_test.go`: `_Success` (точный SQL и аргументы в `Exec`), `_ExecError` (`s.ErrorIs`, `s.Contains("r.pool.Exec")`).
- `internal/usecase/register_member_test.go`: `_Success`, `_RepositoryError`.
- `internal/transport/telegram/` — suite в пакете `telegram_test` с `MockTelegramUsecase`, `MockSender`, логгером в буфер (проверка, что есть запись `error`). Кейсы в `handle_update_test.go`: эхо; профиль из `From` уходит в `RegisterMember`; `user_id` из `From.ID`, а текст `я пользователь 456` его не меняет; два разных отправителя — два разных `user_id`; ответ в `Chat.ID`; нетекстовое сообщение — подсказка без `RegisterMember`; нет `Message`/нет `From` — ни одного вызова моков; ошибка `RegisterMember` — ответ об ошибке, эха нет; ошибка `SendMessage` — без паники; паника в `RegisterMember` (через `DoAndReturn`) — `HandleUpdate` возвращается нормально, в логе `error`.
- Проверка в реальном Telegram — ручная, см. tasks.md.

### Зависимость

`github.com/go-telegram/bot` v1.27.0 (последняя на момент планирования) — выбрана в стеке `AGENTS.md`, отдельного обоснования не требует.

## Risks / Trade-offs

- [Обработчики, начатые до `SIGTERM`, получают отменённый `ctx` и не дожидаются завершения: запрос в БД или ответ может оборваться] → для эха и регистрации это безвредно (upsert идемпотентен). Ожидание горутин библиотека не даёт; если для агента это станет проблемой, решать в задаче 08 (например, собственный `WaitGroup` вокруг обработки).
- [У бота уже настроен webhook — `getUpdates` будет получать 409 Conflict и бесконечно повторять] → ошибки видны в логе через `ErrorsHandler`. Автоматически удалять webhook не стали: это неявное изменение настроек бота у Telegram. Описать в README (задача 15).
- [Регистрация на каждом сообщении — лишний запрос в БД] → один upsert по первичному ключу, для семейного бота нагрузка несущественна.
- [Хендлеры работают параллельно, два сообщения одного пользователя гоняются в upsert] → `ON CONFLICT` атомарен, итог — одна запись.
- [Неверный токен роняет всё приложение, включая REST API] → осознанно: `TELEGRAM_BOT_TOKEN` обязательный, бот — основная часть продукта. Compose перезапустит `app` (`restart: unless-stopped`), ошибка будет видна в логах.

## Migration Plan

Схема БД не меняется. Развёртывание — обычный `docker compose up --build`. Откат — предыдущий образ; записи в `users`, созданные ботом, совместимы со старой версией.
