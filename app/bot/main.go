package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/config"
	"github.com/mih-kopylov/our-spb-bot/internal/handler"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
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
	updateConfig.Timeout = 10

	setMyCommandsConfig := tgbotapi.NewSetMyCommands(tgbotapi.BotCommand{
		Command:     handler.CommandStart,
		Description: "Показать это сообщение1",
	}, tgbotapi.BotCommand{
		Command:     handler.CommandCreateMessage,
		Description: "Создать новое сообщение",
	}, tgbotapi.BotCommand{
		Command:     handler.CommandStatus,
		Description: "Статус сообщений",
	})
	_, err = bot.Request(setMyCommandsConfig)
	if err != nil {
		logrus.Fatal(errorx.EnhanceStackTrace(err, "failed to register bot commands"))
	}

	updates := bot.GetUpdatesChan(updateConfig)
	for update := range updates {
		if update.CallbackQuery != nil {
			err = handler.CallbackHandler(bot, update.CallbackQuery, states)
			if err != nil {
				stack := errorx.EnhanceStackTrace(err, "failed to handle callback")
				logrus.WithField("data", update.CallbackData()).Error(stack)
				reply := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, stack.Message())
				_, err = bot.Send(reply)
				if err != nil {
					logrus.WithField("callback", update.CallbackData()).Error(err)
				}
			}

		} else if update.Message != nil {
			message := update.Message
			command := message.Command()

			switch command {
			case handler.CommandStart:
				err = handler.StartHandlerApi(bot, message, states)
			case handler.CommandStatus:
				err = handler.StatusHandlerApi(bot, message, states)
			case handler.CommandCreateMessage:
				err = handler.CreateMessageHandlerApi(bot, message, states)
			default:
				err = handler.FillMessageApi(bot, message, states)
			}
			if err != nil {
				stack := errorx.EnhanceStackTrace(err, "failed to handle command")
				logrus.WithField("command", command).Error(stack)
				reply := tgbotapi.NewMessage(message.Chat.ID, stack.Message())
				_, err = bot.Send(reply)
				if err != nil {
					logrus.WithField("command", command).Error(err)
				}
			}
		} else {
			logrus.Info("unsupported update type")
			continue
		}

	}

	return nil
}
