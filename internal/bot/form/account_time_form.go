package form

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/bot"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/service"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/mih-kopylov/our-spb-bot/internal/util"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

const (
	AccountTimeFormName = "AccountTimeForm"
)

type AccountTimeForm struct {
	logger  *zap.Logger
	states  state.States
	service *service.Service
}

func (f *AccountTimeForm) Name() string {
	return AccountTimeFormName
}

func NewAccountTimeForm(logger *zap.Logger, states state.States, service *service.Service) bot.Form {
	return &AccountTimeForm{
		logger:  logger,
		states:  states,
		service: service,
	}
}

func (f *AccountTimeForm) Handle(message *tgbotapi.Message) error {
	userState, err := f.states.GetState(message.Chat.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	accountLogin := userState.GetStringFormField(state.FormFieldLogin)
	if accountLogin == "" {
		return f.service.SendMessage(message.Chat, `Логин, сохранённый на предыдущем шаге, не найден.

Введите команду /settings чтобы попробовать снова.`)
	}

	_, accountIndex, found := lo.FindIndexOf(userState.Accounts, func(item state.Account) bool {
		return item.Login == accountLogin
	})
	if !found {
		replyText := fmt.Sprintf(`Не удалось найти аккаунт по логину %v`, accountLogin)
		return f.service.SendMessage(message.Chat, replyText)
	}

	timeString := message.Text
	if timeString == "" {
		return f.service.SendMessage(message.Chat, "Введите время в формате ЧЧ:ММ")
	}

	timeValue, err := time.Parse("15:04", timeString)
	if err != nil {
		f.logger.Error("failed to parse user input time", zap.Error(err))
		return f.service.SendMessage(message.Chat, "Не удалось прочитать время. Убедитесь, что формат соответствует ЧЧ:ММ")
	}

	hour, min, _ := timeValue.Clock()
	year, month, day := util.DefaultSendTime.Date()
	newTime := time.Date(year, month, day, hour, min, 0, 0, util.SpbLocation)
	userState.Accounts[accountIndex].RateLimitNextDayTime = newTime
	userState.MessageHandlerName = ""
	userState.ClearForm()

	err = f.states.SetState(userState)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state")
	}

	replyText := fmt.Sprintf(`Время отправки сообщений для аккаунта %v сохранено.`, accountLogin)
	return f.service.SendMessage(message.Chat, replyText)
}
