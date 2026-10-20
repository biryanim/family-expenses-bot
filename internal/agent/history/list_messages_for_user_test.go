package history_test

import (
	"time"

	"github.com/biryanim/family-expenses-bot/internal/agent"
)

func (s *Suite) TestListMessagesForUser_Empty() {
	messages := s.history.ListMessagesForUser(userA, now)

	s.Empty(messages)
}

func (s *Suite) TestListMessagesForUser_ReturnsAppendedInOrder() {
	s.history.AppendMessagesForUser(userA, now, userText("кофе"), assistantText("Сколько стоил кофе?"))
	s.history.AppendMessagesForUser(userA, now.Add(time.Minute), userText("350"), assistantText("Записал: кофе 350 ₽."))

	messages := s.history.ListMessagesForUser(userA, now.Add(2*time.Minute))

	s.Equal([]agent.Message{
		userText("кофе"),
		assistantText("Сколько стоил кофе?"),
		userText("350"),
		assistantText("Записал: кофе 350 ₽."),
	}, messages)
}

func (s *Suite) TestListMessagesForUser_Expired() {
	s.history.AppendMessagesForUser(userA, now, userText("кофе"), assistantText("Сколько стоил кофе?"))
	s.history.AppendMessagesForUser(userA, now.Add(time.Minute), userText("чай"), assistantText("Сколько стоил чай?"))

	messages := s.history.ListMessagesForUser(userA, now.Add(16*time.Minute))

	s.Equal([]agent.Message{userText("чай"), assistantText("Сколько стоил чай?")}, messages)
}

func (s *Suite) TestListMessagesForUser_ExactlyTTL() {
	s.history.AppendMessagesForUser(userA, now, userText("кофе"), assistantText("Сколько стоил кофе?"))

	messages := s.history.ListMessagesForUser(userA, now.Add(15*time.Minute))

	s.Equal([]agent.Message{userText("кофе"), assistantText("Сколько стоил кофе?")}, messages)
}

func (s *Suite) TestListMessagesForUser_OtherUser() {
	s.history.AppendMessagesForUser(userB, now, userText("такси 1200"), assistantText("Записал: такси 1200 ₽."))

	messages := s.history.ListMessagesForUser(userA, now)

	s.Empty(messages)
}

func (s *Suite) TestListMessagesForUser_ReturnsCopy() {
	s.history.AppendMessagesForUser(userA, now, userText("кофе"), assistantText("Сколько стоил кофе?"))

	messages := s.history.ListMessagesForUser(userA, now)
	messages[0].Content = "изменено"

	s.Equal([]agent.Message{userText("кофе"), assistantText("Сколько стоил кофе?")}, s.history.ListMessagesForUser(userA, now))
}
