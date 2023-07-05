package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
)

var (
	Errors                   = errorx.NewNamespace("Bot")
	ErrFailedToInitializeBot = Errors.NewType("FailedToInitializeBot")
	ErrFailedToDeleteMessage = Errors.NewType("FailedToDeleteMessage")
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
