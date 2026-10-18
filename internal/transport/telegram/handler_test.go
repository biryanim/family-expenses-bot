package telegram_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/go-telegram/bot/models"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/biryanim/family-expenses-bot/internal/transport/telegram"
	"github.com/biryanim/family-expenses-bot/mock"
)

const (
	userA  int64 = 123
	userB  int64 = 456
	chatID int64 = 777
)

type Suite struct {
	suite.Suite

	ctrl *gomock.Controller

	usecase *mock.MockTelegramUsecase
	agent   *mock.MockAgent
	sender  *mock.MockSender
	logs    *bytes.Buffer
	handler *telegram.Handler
}

func (s *Suite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.usecase = mock.NewMockTelegramUsecase(s.ctrl)
	s.agent = mock.NewMockAgent(s.ctrl)
	s.sender = mock.NewMockSender(s.ctrl)
	s.logs = &bytes.Buffer{}
	s.handler = telegram.New(s.usecase, s.agent, s.sender, slog.New(slog.NewJSONHandler(s.logs, nil)))
}

func (s *Suite) TearDownTest() {
	s.ctrl.Finish()
}

func TestTelegram(t *testing.T) {
	suite.Run(t, new(Suite))
}

func textUpdate(from int64, text string) *models.Update {
	return &models.Update{
		Message: &models.Message{
			From: &models.User{ID: from, Username: "ivan", FirstName: "Иван", LastName: "Петров"},
			Chat: models.Chat{ID: chatID},
			Text: text,
		},
	}
}
