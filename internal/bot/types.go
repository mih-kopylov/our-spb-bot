package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Command interface {
	Name() string
	Description() string
	Handle(message *tgbotapi.Message) error
}

type Form interface {
	Name() string
	Handle(message *tgbotapi.Message) error
}

type Callback interface {
	Name() string
	Handle(callbackQuery *tgbotapi.CallbackQuery, data string) error
}

const (
	CallbackSectionSeparator = "."
)
