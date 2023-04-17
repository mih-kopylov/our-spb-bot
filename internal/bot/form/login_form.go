package form

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/bot"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/service"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/samber/lo"
	"strings"
)

const (
	LoginFormName = "LoginForm"
)

type LoginForm struct {
	states  state.States
	service *service.Service
}

func NewLoginForm(states state.States, service *service.Service) bot.Form {
	return &LoginForm{
		states:  states,
		service: service,
	}
}

func (f *LoginForm) Name() string {
	return LoginFormName
}

func (f *LoginForm) Handle(message *tgbotapi.Message) error {
	userState, err := f.states.GetState(message.Chat.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	login := message.Text
	if login == "" {
		return f.service.SendMessage(message.Chat, "Введите логин")
	}

	err = f.service.DeleteMessage(message)
	if err != nil {
		return err
	}

	if lo.CountBy(userState.Accounts, func(item state.Account) bool {
		return strings.EqualFold(item.Login, login)
	}) > 0 {
		return f.service.SendMessage(message.Chat, "Этот логин уже используется, введите новый")
	}

	userState.SetFormField(state.FormFieldLogin, login)
	userState.MessageHandlerName = PasswordFormName
	err = f.states.SetState(userState)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state")
	}

	return f.service.SendMessage(message.Chat, "Введите пароль")
}
