package bot

import (
	_ "embed"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
)

const (
	LoginCommandName = "login"
)

type LoginCommand struct {
	states state.States `di.inject:"States"`
	tgbot  *TgBot       `di.inject:"TgBot"`
}

func (c *LoginCommand) Name() string {
	return LoginCommandName
}

func (c *LoginCommand) Description() string {
	return "Авторизация на портале"
}

func (c *LoginCommand) Handle(message *tgbotapi.Message) error {
	userState, err := c.states.GetState(message.Chat.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	userState.MessageHandlerName = LoginFormBeanId
	err = c.states.SetState(userState)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state")
	}

	reply := tgbotapi.NewMessage(message.Chat.ID, "Введите логин от аккаунта на портале")
	_, err = c.tgbot.api.Send(reply)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to send reply")
	}

	return nil
}

func (c *LoginCommand) Callback(_ *tgbotapi.CallbackQuery, _ string) error {
	return nil
}
