package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"strings"
)

const (
	SectionSeparator = "."
)

var (
	Errors                   = errorx.NewNamespace("Bot")
	ErrFailedToDeleteMessage = Errors.NewType("FailedToDeleteMessage")
)

type TgBot struct {
	api      *tgbotapi.BotAPI
	states   state.States
	commands map[string]Command
	forms    map[string]Form
}

func NewTgBot(api *tgbotapi.BotAPI, states state.States, commands []Command, forms []Form) *TgBot {
	return &TgBot{
		api:    api,
		states: states,
		commands: lo.SliceToMap(commands, func(item Command) (string, Command) {
			return item.Name(), item
		}),
		forms: lo.SliceToMap(forms, func(item Form) (string, Form) {
			return item.Name(), item
		}),
	}
}

func (b *TgBot) Start() error {
	err := b.registerCommands()
	if err != nil {
		return err
	}

	go func() {
		b.processUpdates()
	}()

	return nil
}

func (b *TgBot) processUpdates() {
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

func (b *TgBot) registerCommands() error {
	setMyCommandsConfig := tgbotapi.NewSetMyCommands(
		lo.Map(
			maps.Values(b.commands), func(command Command, _ int) tgbotapi.BotCommand {
				return tgbotapi.BotCommand{
					Command:     command.Name(),
					Description: command.Description(),
				}
			},
		)...,
	)
	_, err := b.api.Request(setMyCommandsConfig)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to register bot commands")
	}

	return nil
}

func (b *TgBot) handleCallback(callbackQuery *tgbotapi.CallbackQuery) error {
	data := callbackQuery.Data
	commandName, value, found := strings.Cut(data, SectionSeparator)
	if !found {
		return errorx.IllegalArgument.New("unsupported callback data format")
	}

	comm, exists := b.commands[commandName]
	if !exists {
		return errorx.IllegalArgument.New("unsupported command name: name=%v", commandName)
	}

	err := comm.Callback(callbackQuery, value)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to handle callback")
	}

	return nil
}

func (b *TgBot) handleMessage(message *tgbotapi.Message) error {
	commandName := message.Command()

	if commandName != "" {
		comm, exists := b.commands[commandName]
		if !exists {
			return errorx.IllegalArgument.New("unsupported command name: name=%v", commandName)
		}

		err := comm.Handle(message)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to handle command")
		}
	} else {
		userState, err := b.states.GetState(message.Chat.ID)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to get user state")
		}

		form, exists := b.forms[userState.MessageHandlerName]
		if !exists {
			return errorx.IllegalState.New("no message handler is waiting for a message")
		}

		err = form.Handle(message)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to handle message")
		}
	}

	return nil
}
