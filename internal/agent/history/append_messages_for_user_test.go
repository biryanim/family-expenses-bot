package history_test

import (
	"strconv"
	"sync"
	"time"

	"github.com/biryanim/family-expenses-bot/internal/agent"
)

func (s *Suite) TestAppendMessagesForUser_KeepsLastSix() {
	for i := range 4 {
		at := now.Add(time.Duration(i) * time.Minute)
		s.history.AppendMessagesForUser(userA, at, userText("вопрос "+strconv.Itoa(i)), assistantText("ответ "+strconv.Itoa(i)))
	}

	messages := s.history.ListMessagesForUser(userA, now.Add(4*time.Minute))

	s.Equal([]agent.Message{
		userText("вопрос 1"),
		assistantText("ответ 1"),
		userText("вопрос 2"),
		assistantText("ответ 2"),
		userText("вопрос 3"),
		assistantText("ответ 3"),
	}, messages)
	s.Equal(agent.RoleUser, messages[0].Role)
}

func (s *Suite) TestAppendMessagesForUser_DropsExpired() {
	s.history.AppendMessagesForUser(userA, now, userText("вопрос 0"), assistantText("ответ 0"))
	for i := 1; i <= 3; i++ {
		at := now.Add(20*time.Minute + time.Duration(i)*time.Minute)
		s.history.AppendMessagesForUser(userA, at, userText("вопрос "+strconv.Itoa(i)), assistantText("ответ "+strconv.Itoa(i)))
	}

	messages := s.history.ListMessagesForUser(userA, now.Add(24*time.Minute))

	s.Equal([]agent.Message{
		userText("вопрос 1"),
		assistantText("ответ 1"),
		userText("вопрос 2"),
		assistantText("ответ 2"),
		userText("вопрос 3"),
		assistantText("ответ 3"),
	}, messages)
}

func (s *Suite) TestAppendMessagesForUser_OtherUserUntouched() {
	s.history.AppendMessagesForUser(userA, now, userText("кофе"), assistantText("Сколько стоил кофе?"))

	for i := range 4 {
		s.history.AppendMessagesForUser(userB, now, userText("такси "+strconv.Itoa(i)), assistantText("Записал."))
	}

	s.Equal([]agent.Message{userText("кофе"), assistantText("Сколько стоил кофе?")}, s.history.ListMessagesForUser(userA, now))
	s.Len(s.history.ListMessagesForUser(userB, now), 6)
}

func (s *Suite) TestAppendMessagesForUser_Concurrent() {
	const users = 50
	var wg sync.WaitGroup

	for i := range users {
		wg.Go(func() {
			userID := int64(i)
			s.history.AppendMessagesForUser(userID, now, userText(strconv.Itoa(i)), assistantText("ok"))
			s.history.ListMessagesForUser(userID, now)
		})
	}
	wg.Wait()

	for i := range users {
		s.Equal([]agent.Message{userText(strconv.Itoa(i)), assistantText("ok")}, s.history.ListMessagesForUser(int64(i), now))
	}
}
