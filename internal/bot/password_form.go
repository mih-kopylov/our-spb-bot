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
	"time"
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

	password := message.Text
	if password == "" {
		return f.tgbot.SendMessage(message.Chat, "Введите пароль")
	}

	err = f.tgbot.DeleteMessage(message)
	if err != nil {
		return err
	}

	login := userState.GetStringFormField(state.FormFieldLogin)
	if login == "" {
		return f.tgbot.SendMessage(message.Chat, `Логин, сохранённый на предыдущем шаге, не найден.

Введите команду /login для авторизации.`)
	}

	userState.MessageHandlerName = ""
	userState.ClearForm()

	tokenResponse, err := f.spbClient.Login(login, password)
	if err != nil {
		err = f.states.SetState(userState)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to set user state")
		}

		return f.tgbot.SendMessage(message.Chat, `Не удалось авторизоваться.

Введите команду /login для авторизации.`)
	}

	userState.Accounts = append(userState.Accounts, state.Account{
		Login:            login,
		Password:         password,
		Token:            tokenResponse.AccessToken,
		RateLimitedUntil: time.Time{},
		State: state.AccountStateEnabled,
	})

	err = f.states.SetState(userState)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state")
	}

	err = f.queue.ResetAwaitingAuthorizationMessages(userState.UserId)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to reset messages that are waiting for authorization")
	}

	return f.tgbot.SendMessage(message.Chat, `Авторизация прошла успешно. Учётные данные сохранены.

Введите команду /message для отправки обращения.`)
}
