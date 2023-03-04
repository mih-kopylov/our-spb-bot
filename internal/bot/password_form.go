package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/goioc/di"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/queue"
	"github.com/mih-kopylov/our-spb-bot/internal/spb"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/samber/lo"
	"reflect"
)

const (
	PasswordFormBeanId = "PasswordForm"
)

type PasswordForm struct {
	states       *state.States      `di.inject:"States"`
	tgbot        *TgBot             `di.inject:"TgBot"`
	messageQueue queue.MessageQueue `di.inject:"Queue"`
	spbClient    spb.Client         `di.inject:"SpbClient"`
}

func RegisterPasswordFormBean() {
	_ = lo.Must(di.RegisterBean(PasswordFormBeanId, reflect.TypeOf((*PasswordForm)(nil))))
}

func (f *PasswordForm) Handle(message *tgbotapi.Message) error {
	userState, err := f.states.GetState(message.Chat.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	if message.Text == "" {
		return f.tgbot.SendMessage(message.Chat, "Введите пароль")
	}

	err = userState.SetPassword(message.Text)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set password")
	}

	err = userState.SetMessageHandlerName("")
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set message handler")
	}

	tokenResponse, err := f.spbClient.Login(userState.GetLogin(), userState.GetPassword())
	if err != nil {
		err = userState.SetLogin("")
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to reset login")
		}

		err = userState.SetPassword("")
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to reset password")
		}

		return f.tgbot.SendMessage(message.Chat, `Не удалось авторизоваться.

Введите команду /login, чтобы залогиниться снова`)
	}

	err = userState.SetLogin(tokenResponse.AccessToken)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set token")
	}

	return f.tgbot.SendMessage(message.Chat, `Авторизация прошла успешно. Учётные данные сохранены.

Введите команду /message для отправки обращения`)
}
