package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
)

const (
	StartCommand   = "start"
	MessageCommand = "message"
	StatusCommand  = "status"
	LoginCommand   = "login"
)

type CommandConfiguration struct {
	Name        string
	Description string
	Handler     func(bot *tgbotapi.BotAPI, message *tgbotapi.Message, states *state.States) error
}

func GetCommands() map[string]CommandConfiguration {
	return map[string]CommandConfiguration{StartCommand: {
		Name:        StartCommand,
		Description: "Запустить бота",
		Handler:     StartCommandHandler,
	}, StatusCommand: {
		Name:        StatusCommand,
		Description: "Статус обращений",
		Handler:     StatusCommandHandler,
	}, MessageCommand: {
		Name:        MessageCommand,
		Description: "Отправить обращение",
		Handler:     MessageCommandHandler,
	}, LoginCommand: {
		Name:        LoginCommand,
		Description: "Авторизация на портале",
		Handler:     LoginCommandHandler,
	}}
}
