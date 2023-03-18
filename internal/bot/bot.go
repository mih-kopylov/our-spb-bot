package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/goioc/di"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"reflect"
	"strings"
)

type TgBot struct {
	api    *tgbotapi.BotAPI `di.inject:"TgApi"`
	states state.States     `di.inject:"States"`
}

const (
	TgBotBeanId      = "TgBot"
	SectionSeparator = "."
)

func RegisterBotBean() {
	_ = lo.Must(di.RegisterBean(TgBotBeanId, reflect.TypeOf((*TgBot)(nil))))

	_ = lo.Must(di.RegisterBean(StartCommandName, reflect.TypeOf((*StartCommand)(nil))))
	_ = lo.Must(di.RegisterBean(StatusCommandName, reflect.TypeOf((*StatusCommand)(nil))))
	_ = lo.Must(di.RegisterBean(MessageCommandName, reflect.TypeOf((*MessageCommand)(nil))))
	_ = lo.Must(di.RegisterBean(LoginCommandName, reflect.TypeOf((*LoginCommand)(nil))))
	_ = lo.Must(di.RegisterBean(ResetStatusCommandName, reflect.TypeOf((*ResetStatusCommand)(nil))))

	RegisterMessageFormBean()
	RegisterLoginFormBean()
	RegisterPasswordFormBean()
}

func (b *TgBot) RegisterBotCommands() {
	setMyCommandsConfig := tgbotapi.NewSetMyCommands(
		lo.Map(
			b.GetCommands(), func(commandName string, _ int) tgbotapi.BotCommand {
				comm := di.GetInstance(commandName).(Command)
				return tgbotapi.BotCommand{
					Command:     comm.Name(),
					Description: comm.Description(),
				}
			},
		)...,
	)
	_, err := b.api.Request(setMyCommandsConfig)
	if err != nil {
		logrus.Fatal(errorx.EnhanceStackTrace(err, "failed to register bot commands"))
	}

}
func (b *TgBot) GetCommands() []string {
	return []string{StartCommandName, LoginCommandName, MessageCommandName, StatusCommandName, ResetStatusCommandName}
}

func (b *TgBot) ProcessUpdates() {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := b.api.GetUpdatesChan(updateConfig)
	for update := range updates {
		var stack *errorx.Error
		var chat *tgbotapi.Chat

		if update.CallbackQuery != nil {
			err := b.handleCallback(update.CallbackQuery)
			if err != nil {
				chat = update.CallbackQuery.Message.Chat
				stack = errorx.EnhanceStackTrace(err, "failed to handle callback")
			}
		} else if update.Message != nil {
			err := b.handleMessage(update.Message)
			if err != nil {
				chat = update.Message.Chat
				stack = errorx.EnhanceStackTrace(err, "failed to handle message")
			}
		} else {
			logrus.Info("unsupported update type")
		}

		if stack != nil {
			logrus.WithField("chat", chat.ID).Error(stack)
		}
	}
}

func (b *TgBot) SendMessage(chat *tgbotapi.Chat, text string) error {
	return b.SendMessageCustom(chat, text, func(reply *tgbotapi.MessageConfig) {})
}

func (b *TgBot) SendMessageCustom(chat *tgbotapi.Chat, text string, messageAdjuster func(reply *tgbotapi.MessageConfig)) error {
	message := tgbotapi.NewMessage(chat.ID, text)
	messageAdjuster(&message)
	_, err := b.api.Send(message)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to send reply")
	}

	return nil
}

func (b *TgBot) handleCallback(callbackQuery *tgbotapi.CallbackQuery) error {
	data := callbackQuery.Data
	commandName, value, found := strings.Cut(data, SectionSeparator)
	if !found {
		return errorx.IllegalArgument.New("unsupported callback data format")
	}

	comm, err := di.GetInstanceSafe(commandName)
	if err != nil {
		return errorx.IllegalArgument.New("unsupported command name: name=%v", commandName)
	}

	err = comm.(Command).Callback(callbackQuery, value)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to handle callback")
	}

	return nil
}

func (b *TgBot) handleMessage(message *tgbotapi.Message) error {
	commandName := message.Command()

	if commandName != "" {
		comm, err := di.GetInstanceSafe(commandName)
		if err != nil {
			return errorx.IllegalArgument.New("unsupported command name: name=%v", commandName)
		}

		err = comm.(Command).Handle(message)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to handle command")
		}
	} else {
		userState, err := b.states.GetState(message.Chat.ID)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to get user state")
		}

		formInstance, err := di.GetInstanceSafe(userState.MessageHandlerName)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "no message handler is waiting for a message")
		}

		err = formInstance.(Form).Handle(message)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to handle message")
		}
	}

	return nil
}
