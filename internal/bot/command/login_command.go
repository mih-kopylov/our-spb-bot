package command

import (
	_ "embed"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/form"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/service"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/mih-kopylov/our-spb-bot/pkg/bot"
)

const (
	LoginCommandName = "login"
)

type LoginCommand struct {
	states  state.States
	service *service.Service
}

func NewLoginCommand(states state.States, service *service.Service) bot.Command {
	return &LoginCommand{
		states:  states,
		service: service,
	}
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

	userState.MessageHandlerName = form.LoginFormName
	err = c.states.SetState(userState)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state")
	}

	err = c.service.SendMessage(message.Chat, "Введите логин от аккаунта на портале")
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to send reply")
	}

	return nil
}
