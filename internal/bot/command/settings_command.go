package command

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/callback"
	"github.com/mih-kopylov/our-spb-bot/pkg/tgbot"
)

const (
	SettingsCommandName = "settings"
)

type SettingsCommand struct {
	service          *tgbot.Service
	settingsCallback *callback.SettingsCallback
}

func NewSettingsCommand(service *tgbot.Service, settingsCallback *callback.SettingsCallback) tgbot.Command {
	return &SettingsCommand{
		service:          service,
		settingsCallback: settingsCallback,
	}
}

func (c *SettingsCommand) Name() string {
	return SettingsCommandName
}

func (c *SettingsCommand) Description() string {
	return "Настройки"
}

func (c *SettingsCommand) Handle(message *tgbotapi.Message) error {
	_, err := c.service.SendMessageCustom(message.Chat, `Выберите настройку`, func(reply *tgbotapi.MessageConfig) {
		reply.ReplyMarkup = c.settingsCallback.CreateReplyMarkup()
	})

	return err
}
