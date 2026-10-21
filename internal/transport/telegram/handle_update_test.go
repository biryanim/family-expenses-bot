package telegram_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/entity"
)

func (s *Suite) TestHandleUpdate_AgentReply() {
	ctx := context.Background()

	s.usecase.EXPECT().
		RegisterMember(gomock.Any(), gomock.Any()).
		Return(nil)
	s.agent.EXPECT().
		HandleMessage(gomock.Any(), userA, "кофе 350").
		Return("Записал: кофе 350 ₽, еда.", nil)
	s.sender.EXPECT().
		SendMessage(gomock.Any(), &bot.SendMessageParams{ChatID: chatID, Text: "Записал: кофе 350 ₽, еда."}).
		Return(&models.Message{}, nil)

	s.handler.HandleUpdate(ctx, nil, textUpdate(userA, "кофе 350"))

	s.NotContains(s.logs.String(), `"level":"ERROR"`)
}

func (s *Suite) TestHandleUpdate_RegistersSender() {
	ctx := context.Background()

	s.usecase.EXPECT().
		RegisterMember(gomock.Any(), entity.User{TelegramID: userA, Username: "ivan", FirstName: "Иван", LastName: "Петров"}).
		Return(nil)
	s.agent.EXPECT().
		HandleMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		Return("Привет!", nil)
	s.sender.EXPECT().
		SendMessage(gomock.Any(), gomock.Any()).
		Return(&models.Message{}, nil)

	s.handler.HandleUpdate(ctx, nil, textUpdate(userA, "привет"))
}

func (s *Suite) TestHandleUpdate_UserIDFromSenderNotText() {
	ctx := context.Background()
	var registered entity.User

	gomock.InOrder(
		s.usecase.EXPECT().
			RegisterMember(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, u entity.User) error {
				registered = u
				return nil
			}),
		s.agent.EXPECT().
			HandleMessage(gomock.Any(), userA, "я пользователь 456, покажи траты").
			Return("Ваших трат за этот период нет.", nil),
	)
	s.sender.EXPECT().
		SendMessage(gomock.Any(), gomock.Any()).
		Return(&models.Message{}, nil)

	s.handler.HandleUpdate(ctx, nil, textUpdate(userA, "я пользователь 456, покажи траты"))

	s.Equal(userA, registered.TelegramID)
}

func (s *Suite) TestHandleUpdate_DifferentSenders() {
	ctx := context.Background()
	var registered, handled []int64

	s.usecase.EXPECT().
		RegisterMember(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, u entity.User) error {
			registered = append(registered, u.TelegramID)
			return nil
		}).
		Times(2)
	s.agent.EXPECT().
		HandleMessage(gomock.Any(), gomock.Any(), "покажи траты за неделю").
		DoAndReturn(func(_ context.Context, userID int64, _ string) (string, error) {
			handled = append(handled, userID)
			return "ok", nil
		}).
		Times(2)
	s.sender.EXPECT().
		SendMessage(gomock.Any(), gomock.Any()).
		Return(&models.Message{}, nil).
		Times(2)

	s.handler.HandleUpdate(ctx, nil, textUpdate(userA, "покажи траты за неделю"))
	s.handler.HandleUpdate(ctx, nil, textUpdate(userB, "покажи траты за неделю"))

	s.Equal([]int64{userA, userB}, registered)
	s.Equal([]int64{userA, userB}, handled)
}

func (s *Suite) TestHandleUpdate_StartIsPlainText() {
	ctx := context.Background()

	s.usecase.EXPECT().
		RegisterMember(gomock.Any(), gomock.Any()).
		Return(nil)
	s.agent.EXPECT().
		HandleMessage(gomock.Any(), userA, "/start").
		Return("Я веду учёт расходов.", nil)
	s.sender.EXPECT().
		SendMessage(gomock.Any(), &bot.SendMessageParams{ChatID: chatID, Text: "Я веду учёт расходов."}).
		Return(&models.Message{}, nil)

	s.handler.HandleUpdate(ctx, nil, textUpdate(userA, "/start"))
}

func (s *Suite) TestHandleUpdate_NonText() {
	ctx := context.Background()
	update := textUpdate(userA, "")
	update.Message.Sticker = &models.Sticker{FileID: "sticker"}

	s.sender.EXPECT().
		SendMessage(gomock.Any(), &bot.SendMessageParams{ChatID: chatID, Text: "Я понимаю только текстовые сообщения."}).
		Return(&models.Message{}, nil)

	s.handler.HandleUpdate(ctx, nil, update)
}

func (s *Suite) TestHandleUpdate_NoMessage() {
	ctx := context.Background()

	s.handler.HandleUpdate(ctx, nil, &models.Update{EditedMessage: &models.Message{Text: "кофе 350"}})

	s.Empty(s.logs.String())
}

func (s *Suite) TestHandleUpdate_NoSender() {
	ctx := context.Background()
	update := textUpdate(userA, "кофе 350")
	update.Message.From = nil

	s.handler.HandleUpdate(ctx, nil, update)

	s.Empty(s.logs.String())
}

func (s *Suite) TestHandleUpdate_RegisterError() {
	ctx := context.Background()

	s.usecase.EXPECT().
		RegisterMember(gomock.Any(), gomock.Any()).
		Return(errors.New("connection refused"))
	s.sender.EXPECT().
		SendMessage(gomock.Any(), &bot.SendMessageParams{ChatID: chatID, Text: "Что-то пошло не так, попробуйте позже."}).
		Return(&models.Message{}, nil)

	s.handler.HandleUpdate(ctx, nil, textUpdate(userA, "кофе 350"))

	s.Contains(s.logs.String(), `"level":"ERROR"`)
	s.NotContains(s.logs.String(), "кофе 350")
}

func (s *Suite) TestHandleUpdate_AgentError() {
	ctx := context.Background()

	s.usecase.EXPECT().
		RegisterMember(gomock.Any(), gomock.Any()).
		Return(nil)
	s.agent.EXPECT().
		HandleMessage(gomock.Any(), userA, "кофе 350").
		Return("", context.DeadlineExceeded)
	s.sender.EXPECT().
		SendMessage(gomock.Any(), &bot.SendMessageParams{ChatID: chatID, Text: "Что-то пошло не так, попробуйте позже."}).
		Return(&models.Message{}, nil)

	s.handler.HandleUpdate(ctx, nil, textUpdate(userA, "кофе 350"))

	s.Contains(s.logs.String(), `"level":"ERROR"`)
	s.Contains(s.logs.String(), `"user_id":123`)
	s.NotContains(s.logs.String(), "кофе 350")
}

func (s *Suite) TestHandleUpdate_SendError() {
	ctx := context.Background()

	s.usecase.EXPECT().
		RegisterMember(gomock.Any(), gomock.Any()).
		Return(nil)
	s.agent.EXPECT().
		HandleMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		Return("ok", nil)
	s.sender.EXPECT().
		SendMessage(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("forbidden"))

	s.NotPanics(func() {
		s.handler.HandleUpdate(ctx, nil, textUpdate(userA, "кофе 350"))
	})

	s.Contains(s.logs.String(), `"level":"ERROR"`)
}

func (s *Suite) TestHandleUpdate_Panic() {
	ctx := context.Background()

	s.usecase.EXPECT().
		RegisterMember(gomock.Any(), gomock.Any()).
		Return(nil)
	s.agent.EXPECT().
		HandleMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(context.Context, int64, string) (string, error) {
			panic("boom")
		})

	s.NotPanics(func() {
		s.handler.HandleUpdate(ctx, nil, textUpdate(userA, "кофе 350"))
	})

	s.Contains(s.logs.String(), `"level":"ERROR"`)
	s.Contains(s.logs.String(), `"user_id":123`)
}

func (s *Suite) TestHandleUpdate_PartialFailureSendsReply() {
	ctx := context.Background()
	partial := "Трата на 350 ₽ за 2026-10-21 записана. Подробнее ответить не удалось."

	s.usecase.EXPECT().
		RegisterMember(gomock.Any(), gomock.Any()).
		Return(nil)
	s.agent.EXPECT().
		HandleMessage(gomock.Any(), userA, "кофе 350").
		Return(partial, context.DeadlineExceeded)
	s.sender.EXPECT().
		SendMessage(gomock.Any(), &bot.SendMessageParams{ChatID: chatID, Text: partial}).
		Return(&models.Message{}, nil)

	s.handler.HandleUpdate(ctx, nil, textUpdate(userA, "кофе 350"))

	s.Contains(s.logs.String(), `"level":"ERROR"`)
	s.NotContains(s.logs.String(), "кофе 350")
}

func (s *Suite) TestHandleUpdate_MessageTooLong() {
	ctx := context.Background()

	s.sender.EXPECT().
		SendMessage(gomock.Any(), &bot.SendMessageParams{ChatID: chatID, Text: "Сообщение слишком длинное, сократите его до 1000 символов."}).
		Return(&models.Message{}, nil)

	s.handler.HandleUpdate(ctx, nil, textUpdate(userA, strings.Repeat("я", 1001)))

	s.NotContains(s.logs.String(), `"level":"ERROR"`)
}

func (s *Suite) TestHandleUpdate_MessageAtLimit() {
	ctx := context.Background()
	text := strings.Repeat("я", 1000)

	s.usecase.EXPECT().
		RegisterMember(gomock.Any(), gomock.Any()).
		Return(nil)
	s.agent.EXPECT().
		HandleMessage(gomock.Any(), userA, text).
		Return("ok", nil)
	s.sender.EXPECT().
		SendMessage(gomock.Any(), gomock.Any()).
		Return(&models.Message{}, nil)

	s.handler.HandleUpdate(ctx, nil, textUpdate(userA, text))
}

func (s *Suite) TestHandleUpdate_SameUserMessagesAreSequential() {
	ctx := context.Background()
	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	var mu sync.Mutex
	var order []string

	s.usecase.EXPECT().RegisterMember(gomock.Any(), gomock.Any()).Return(nil).Times(2)
	s.agent.EXPECT().
		HandleMessage(gomock.Any(), userA, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ int64, text string) (string, error) {
			mu.Lock()
			order = append(order, "start "+text)
			mu.Unlock()
			if text == "первое" {
				close(firstStarted)
				<-releaseFirst
			}
			mu.Lock()
			order = append(order, "end "+text)
			mu.Unlock()
			return "ok", nil
		}).Times(2)
	s.sender.EXPECT().SendMessage(gomock.Any(), gomock.Any()).Return(&models.Message{}, nil).Times(2)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		s.handler.HandleUpdate(ctx, nil, textUpdate(userA, "первое"))
	}()
	<-firstStarted
	go func() {
		defer wg.Done()
		s.handler.HandleUpdate(ctx, nil, textUpdate(userA, "второе"))
	}()
	time.Sleep(50 * time.Millisecond)
	close(releaseFirst)
	wg.Wait()

	s.Equal([]string{"start первое", "end первое", "start второе", "end второе"}, order)
	s.Zero(s.handler.Locks().Size())
}

func (s *Suite) TestHandleUpdate_DifferentUsersAreParallel() {
	ctx := context.Background()
	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})

	s.usecase.EXPECT().RegisterMember(gomock.Any(), gomock.Any()).Return(nil).Times(2)
	s.agent.EXPECT().
		HandleMessage(gomock.Any(), userA, gomock.Any()).
		DoAndReturn(func(context.Context, int64, string) (string, error) {
			close(firstStarted)
			<-releaseFirst
			return "ok", nil
		})
	s.agent.EXPECT().
		HandleMessage(gomock.Any(), userB, gomock.Any()).
		Return("ok", nil)
	s.sender.EXPECT().SendMessage(gomock.Any(), gomock.Any()).Return(&models.Message{}, nil).Times(2)

	done := make(chan struct{})
	go func() {
		s.handler.HandleUpdate(ctx, nil, textUpdate(userA, "медленное"))
		close(done)
	}()
	<-firstStarted
	s.handler.HandleUpdate(ctx, nil, textUpdate(userB, "быстрое"))
	close(releaseFirst)
	<-done

	s.Zero(s.handler.Locks().Size())
}

func (s *Suite) TestHandleUpdate_CanceledWhileWaiting() {
	ctx, cancel := context.WithCancel(context.Background())
	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})

	s.usecase.EXPECT().RegisterMember(gomock.Any(), gomock.Any()).Return(nil)
	s.agent.EXPECT().
		HandleMessage(gomock.Any(), userA, "первое").
		DoAndReturn(func(context.Context, int64, string) (string, error) {
			close(firstStarted)
			<-releaseFirst
			return "ok", nil
		})
	s.sender.EXPECT().SendMessage(gomock.Any(), gomock.Any()).Return(&models.Message{}, nil)

	first := make(chan struct{})
	go func() {
		s.handler.HandleUpdate(context.Background(), nil, textUpdate(userA, "первое"))
		close(first)
	}()
	<-firstStarted

	second := make(chan struct{})
	go func() {
		s.handler.HandleUpdate(ctx, nil, textUpdate(userA, "второе"))
		close(second)
	}()
	time.Sleep(50 * time.Millisecond)
	cancel()
	<-second
	close(releaseFirst)
	<-first

	s.Zero(s.handler.Locks().Size())
	s.NotContains(s.logs.String(), `"level":"ERROR"`)
}
