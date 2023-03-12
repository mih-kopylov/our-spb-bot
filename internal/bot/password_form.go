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
	states    state.States       `di.inject:"States"`
	tgbot     *TgBot             `di.inject:"TgBot"`
	spbClient spb.Client         `di.inject:"SpbClient"`
	queue     queue.MessageQueue `di.inject:"Queue"`
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

	userState.Password = message.Text
	userState.MessageHandlerName = ""

	tokenResponse, err := f.spbClient.Login(userState.Login, userState.Password)
	if err != nil {
		userState.Login = ""
		userState.Password = ""
		err = f.states.SetState(userState)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to set user state")
		}

		return f.tgbot.SendMessage(message.Chat, `Не удалось авторизоваться.

Введите команду /login, чтобы залогиниться снова`)
	}

	userState.Token = tokenResponse.AccessToken
	err = f.states.SetState(userState)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state")
	}

	err = f.queue.ResetAwaitingAuthorizationMessages(userState.UserId)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to reset messages that are waiting for authorization")
	}

	return f.tgbot.SendMessage(message.Chat, `Авторизация прошла успешно. Учётные данные сохранены.

Введите команду /message для отправки обращения`)
}
