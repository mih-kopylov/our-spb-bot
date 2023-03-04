package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/goioc/di"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/category"
	"github.com/mih-kopylov/our-spb-bot/internal/queue"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/samber/lo"
	"reflect"
)

const (
	LoginFormBeanId = "LoginForm"
)

type LoginForm struct {
	states         *state.States                  `di.inject:"States"`
	tgbot          *TgBot                         `di.inject:"TgBot"`
	messageQueue   queue.MessageQueue             `di.inject:"Queue"`
	cateogiresTree *category.UserCategoryTreeNode `di.inject:"Categories"`
}

func RegisterLoginFormBean() {
	_ = lo.Must(di.RegisterBean(LoginFormBeanId, reflect.TypeOf((*LoginForm)(nil))))
}

func (f *LoginForm) Handle(message *tgbotapi.Message) error {
	userState, err := f.states.GetState(message.Chat.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	if message.Text == "" {
		return f.tgbot.SendMessage(message.Chat, "Введите логин")
	}

	err = userState.SetLogin(message.Text)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set login")
	}

	err = userState.SetMessageHandlerName(PasswordFormBeanId)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set message handler")
	}

	return f.tgbot.SendMessage(message.Chat, "Введите пароль")
}
