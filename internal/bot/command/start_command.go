package command

import (
	_ "embed"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/category"
	"github.com/mih-kopylov/our-spb-bot/internal/info"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/mih-kopylov/our-spb-bot/pkg/tgbot"
	"strings"
	"text/template"
)

//go:embed start.html
var startTextTemplate string

const (
	StartCommandName = "start"
)

type StartCommand struct {
	stateManager state.Manager
	service      *tgbot.Service
	info         *info.Info
}

func NewStartCommand(stateManager state.Manager, service *tgbot.Service, info *info.Info) tgbot.Command {
	return &StartCommand{
		stateManager: stateManager,
		service:      service,
		info:         info,
	}
}

func (c *StartCommand) Name() string {
	return StartCommandName
}

func (c *StartCommand) Description() string {
	return "Запустить бота"
}

func (c *StartCommand) Handle(message *tgbotapi.Message) error {
	userState, err := c.stateManager.GetState(message.Chat.ID)
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
	if userState.Categories == "" {
		userState.Categories = string(category.DefaultCategoriesText)
	}
	err = c.stateManager.SetState(userState)
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

	_, err = c.service.SendMessageCustom(message.Chat, writer.String(), func(reply *tgbotapi.MessageConfig) {
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
