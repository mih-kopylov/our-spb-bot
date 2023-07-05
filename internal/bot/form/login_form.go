package form

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/mih-kopylov/our-spb-bot/pkg/tgbot"
	"github.com/samber/lo"
	"strings"
)

const (
	LoginFormName = "LoginForm"
)

type LoginForm struct {
	stateManager state.Manager
	service      *tgbot.Service
}

func NewLoginForm(stateManager state.Manager, service *tgbot.Service) tgbot.Form {
	return &LoginForm{
		stateManager: stateManager,
		service:      service,
	}
}

func (f *LoginForm) Name() string {
	return LoginFormName
}

func (f *LoginForm) Handle(message *tgbotapi.Message) error {
	userState, err := f.stateManager.GetState(message.Chat.ID)
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
	err = f.stateManager.SetState(userState)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state")
	}

	return f.service.SendMessage(message.Chat, "Введите пароль")
}
