package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"strings"
)

type TgBot struct {
	logger    *zap.Logger
	api       *tgbotapi.BotAPI
	states    state.States
	commands  map[string]Command
	callbacks map[string]Callback
	forms     map[string]Form
}

func NewTgBot(logger *zap.Logger, api *tgbotapi.BotAPI, states state.States, commands []Command, callbacks []Callback, forms []Form) *TgBot {
	return &TgBot{
		logger: logger,
		api:    api,
		states: states,
		commands: lo.SliceToMap(commands, func(item Command) (string, Command) {
			return item.Name(), item
		}),
		callbacks: lo.SliceToMap(callbacks, func(item Callback) (string, Callback) {
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
		err := b.callHandler(update)
		if err != nil {
			err = errorx.EnhanceStackTrace(err, "failed to handle update")
			b.logger.Error("",
				zap.Int64("chat", update.FromChat().ID),
				zap.Error(err),
			)
		}
	}
}

func (b *TgBot) callHandler(update tgbotapi.Update) error {
	switch {
	case update.Message != nil:
		return b.handleMessage(update.Message)
	case update.CallbackQuery != nil:
		return b.handleCallback(update.CallbackQuery)
	default:
		return errorx.IllegalArgument.New("unsupported update type")
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
	callbackName, value, found := strings.Cut(data, CallbackSectionSeparator)
	if !found {
		return errorx.IllegalArgument.New("unsupported callback data format")
	}

	handler, exists := b.callbacks[callbackName]
	if !exists {
		return errorx.IllegalArgument.New("unsupported callback name: name=%v", callbackName)
	}

	err := handler.Handle(callbackQuery, value)
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
