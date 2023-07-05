package form

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/queue"
	"github.com/mih-kopylov/our-spb-bot/internal/spb"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/mih-kopylov/our-spb-bot/pkg/tgbot"
	"time"
)

const (
	PasswordFormName = "PasswordForm"
)

type PasswordForm struct {
	stateManager state.Manager
	service      *tgbot.Service
	spbClient    spb.Client
	queue        queue.MessageQueue
}

func (f *PasswordForm) Name() string {
	return PasswordFormName
}

func NewPasswordForm(stateManager state.Manager, service *tgbot.Service, spbClient spb.Client, queue queue.MessageQueue) tgbot.Form {
	return &PasswordForm{
		stateManager: stateManager,
		service:      service,
		spbClient:    spbClient,
		queue:        queue,
	}
}

func (f *PasswordForm) Handle(message *tgbotapi.Message) error {
	userState, err := f.stateManager.GetState(message.Chat.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	password := message.Text
	if password == "" {
		return f.service.SendMessage(message.Chat, "Введите пароль")
	}

	err = f.service.DeleteMessage(message)
	if err != nil {
		return err
	}

	login := userState.GetStringFormField(state.FormFieldLogin)
	if login == "" {
		return f.service.SendMessage(message.Chat, `Логин, сохранённый на предыдущем шаге, не найден.

Введите команду /login для авторизации.`)
	}

	userState.MessageHandlerName = ""
	userState.ClearForm()

	tokenResponse, err := f.spbClient.Login(login, password)
	if err != nil {
		err = f.stateManager.SetState(userState)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to set user state")
		}

		return f.service.SendMessage(message.Chat, `Не удалось авторизоваться.

Введите команду /login для авторизации.`)
	}

	userState.Accounts = append(userState.Accounts, state.Account{
		Login:            login,
		Password:         password,
		Token:            tokenResponse.AccessToken,
		RateLimitedUntil: time.Time{},
		State:            state.AccountStateEnabled,
	})

	err = f.stateManager.SetState(userState)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state")
	}

	err = f.queue.ResetAwaitingAuthorizationMessages(userState.UserId)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to reset messages that are waiting for authorization")
	}

	return f.service.SendMessage(message.Chat, `Авторизация прошла успешно. Учётные данные сохранены.

Введите команду /message для отправки обращения.`)
}
