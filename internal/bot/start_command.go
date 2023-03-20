package bot

import (
	_ "embed"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/goioc/di"
	"github.com/joomcode/errorx"
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
	states state.States `di.inject:"States"`
	tgbot  *TgBot       `di.inject:"TgBot"`
	info   *info.Info   `di.inject:"Info"`
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
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	userState.UserName = message.Chat.UserName
	userState.FullName = message.Chat.FirstName + " " + message.Chat.LastName
	err = c.states.SetState(userState)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state")
	}

	parsedTemplate, err := template.New("start").Parse(startTextTemplate)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to parse template")
	}

	commands := c.tgbot.GetCommands()
	context := renderContext{
		Commands: nil,
		Version:  c.info.Version,
		Commit:   c.info.Commit,
	}
	for _, commandName := range commands {
		comm := di.GetInstance(commandName).(Command)
		context.Commands = append(
			context.Commands, commandDescription{
				Name:        comm.Name(),
				Description: comm.Description(),
			},
		)
	}

	writer := strings.Builder{}
	err = parsedTemplate.Execute(&writer, context)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to render template")
	}

	reply := tgbotapi.NewMessage(message.Chat.ID, writer.String())
	reply.ParseMode = tgbotapi.ModeHTML
	reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)

	_, err = c.tgbot.api.Send(reply)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to send reply")
	}

	return nil
}

func (c *StartCommand) Callback(_ *tgbotapi.CallbackQuery, _ string) error {
	return nil
}

type renderContext struct {
	Commands []commandDescription
	Version  string
	Commit   string
}

type commandDescription struct {
	Name        string
	Description string
}
