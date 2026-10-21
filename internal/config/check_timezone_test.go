package config_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/biryanim/family-expenses-bot/internal/config"
)

type CheckTimezoneSuite struct {
	suite.Suite
}

func TestCheckTimezone(t *testing.T) {
	suite.Run(t, new(CheckTimezoneSuite))
}

func (s *CheckTimezoneSuite) TestCheckTimezone_Empty() {
	s.NoError(config.CheckTimezone(""))
}

func (s *CheckTimezoneSuite) TestCheckTimezone_Known() {
	s.NoError(config.CheckTimezone("Europe/Moscow"))
}

func (s *CheckTimezoneSuite) TestCheckTimezone_Unknown() {
	err := config.CheckTimezone("Mars/Olympus")

	s.Error(err)
	s.Contains(err.Error(), "time.LoadLocation")
}
