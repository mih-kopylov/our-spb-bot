package handler

import (
	_ "embed"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"strings"
	"text/template"
)

//go:embed start.html
var startTextTemplate string

func StartCommandHandler(bot *tgbotapi.BotAPI, message *tgbotapi.Message, states *state.States) error {
	_, err := states.NewStateIfNotExists(message.From.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to crate user state")
	}

	parsedTemplate, err := template.New("start").Parse(startTextTemplate)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to parse template")
	}

	commands := GetCommands()
	data := struct {
		Commands map[string]CommandConfiguration
	}{
		Commands: commands,
	}

	writer := strings.Builder{}
	err = parsedTemplate.Execute(&writer, data)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to render template")
	}

	reply := tgbotapi.NewMessage(message.Chat.ID, writer.String())
	reply.ParseMode = tgbotapi.ModeHTML
	reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)

	_, err = bot.Send(reply)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to send reply")
	}

	return nil
}
