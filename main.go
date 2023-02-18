package main

import (
	"github.com/joomcode/errorx"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
	"time"
)

func main() {
	config, err := ReadConfig()
	if err != nil {
		logrus.Fatal(err)
	}

	settings := telebot.Settings{
		Token:   config.Token,
		Poller:  &telebot.LongPoller{Timeout: 30 * time.Second},
		Verbose: true,
	}
	bot, err := telebot.NewBot(settings)
	if err != nil {
		logrus.Fatal(errorx.EnhanceStackTrace(err, "failed to create bot api"))
	}

	logrus.WithField("username", bot.Me.Username).
		WithField("id", bot.Me.ID).
		WithField("firstName", bot.Me.FirstName).
		WithField("lastName", bot.Me.LastName).
		WithField("languageCode", bot.Me.LanguageCode).
		WithField("canJoinGroups", bot.Me.CanJoinGroups).
		WithField("canReadMessages", bot.Me.CanReadMessages).
		WithField("supportsInline", bot.Me.SupportsInline).
		Info("started")

}
