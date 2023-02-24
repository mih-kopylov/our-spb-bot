package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/config"
	"github.com/mih-kopylov/our-spb-bot/internal/handler"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

func main() {
	conf, err := config.ReadConfig()
	if err != nil {
		logrus.Fatal(err)
	}

	//startTelebot(conf)
	err = startApiBot(conf)
	if err != nil {
		logrus.Fatal(err)
	}
}

func startApiBot(conf *config.Config) error {
	states := state.NewStates()

	bot, err := tgbotapi.NewBotAPI(conf.Token)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to create bot")
	}

	bot.Debug = true

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	commands := handler.GetCommands()
	setMyCommandsConfig := tgbotapi.NewSetMyCommands(lo.MapToSlice(commands, func(commandName string, comm handler.CommandConfiguration) tgbotapi.BotCommand {
		return tgbotapi.BotCommand{
			Command:     comm.Name,
			Description: comm.Description,
		}
	})...)
	_, err = bot.Request(setMyCommandsConfig)
	if err != nil {
		logrus.Fatal(errorx.EnhanceStackTrace(err, "failed to register bot commands"))
	}

	updates := bot.GetUpdatesChan(updateConfig)
	for update := range updates {
		var stack *errorx.Error
		var chat *tgbotapi.Chat

		if update.CallbackQuery != nil {
			err = handler.CallbackHandler(bot, update.CallbackQuery, states)
			if err != nil {
				chat = update.CallbackQuery.Message.Chat
				stack = errorx.EnhanceStackTrace(err, "failed to handle callback")
			}

		} else if update.Message != nil {
			message := update.Message
			command := message.Command()
			chat = message.Chat

			if command != "" {
				commandConfiguration, exists := commands[command]
				if !exists {
					stack = errorx.IllegalArgument.New("unsupported command")
				} else {
					err = commandConfiguration.Handler(bot, message, states)
					if err != nil {
						stack = errorx.EnhanceStackTrace(err, "failed to handle command")
					}
				}
			} else {
				err = handler.SimpleMessageHandler(bot, message, states)
				if err != nil {
					stack = errorx.EnhanceStackTrace(err, "failed to handle message")
				}
			}
		} else {
			logrus.Info("unsupported update type")
		}

		if stack != nil {
			logrus.Error(stack)
			reply := tgbotapi.NewMessage(chat.ID, stack.Message())
			_, err = bot.Send(reply)
			if err != nil {
				logrus.Error(err)
			}
		}

	}

	return nil
}
