package command

import (
	_ "embed"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/bot"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/service"
	"github.com/mih-kopylov/our-spb-bot/internal/info"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"strings"
	"text/template"
)

//go:embed start.html
var startTextTemplate string

const (
	StartCommandName = "start"
)

type StartCommand struct {
	states  state.States
	service *service.Service
	info    *info.Info
}

func NewStartCommand(states state.States, service *service.Service, info *info.Info) bot.Command {
	return &StartCommand{
		states:  states,
		service: service,
		info:    info,
	}
}

func (c *StartCommand) Name() string {
	return StartCommandName
}

func (c *StartCommand) Description() string {
	return "Запустить бота"
}

func (c *StartCommand) Handle(message *tgbotapi.Message) error {
	userState, err := c.states.GetState(message.Chat.ID)
	if err != nil {
		if errorx.IsOfType(err, state.ErrRateLimited) {
			err = c.service.SendMessage(message.Chat, "Превышен лимит подключений к базе данных")
			if err != nil {
				return errorx.EnhanceStackTrace(err, "failed to send reply")
			}

			return nil
		}

		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	if message.Chat.IsPrivate() {
		userState.FullName = strings.TrimSpace(fmt.Sprintf("user / @%v %v %v", message.Chat.UserName, message.Chat.FirstName, message.Chat.LastName))
	} else {
		userState.FullName = strings.TrimSpace(fmt.Sprintf("%v / %v",
			message.Chat.Type, message.Chat.Title))
	}
	userState.MessageHandlerName = ""
	err = c.states.SetState(userState)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state")
	}

	parsedTemplate, err := template.New("start").Parse(startTextTemplate)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to parse template")
	}

	context := renderContext{
		Version: c.info.Version,
		Commit:  c.info.Commit,
	}

	writer := strings.Builder{}
	err = parsedTemplate.Execute(&writer, context)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to render template")
	}

	err = c.service.SendMessageCustom(message.Chat, writer.String(), func(reply *tgbotapi.MessageConfig) {
		reply.ParseMode = tgbotapi.ModeHTML
		reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
	})
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to send reply")
	}

	return nil
}

type renderContext struct {
	Version string
	Commit  string
}
