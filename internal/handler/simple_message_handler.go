package handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/spb"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/samber/lo"
	"golang.org/x/exp/slices"
	"time"
)

type PersonalSimpleMessageHandler func(bot *tgbotapi.BotAPI, message *tgbotapi.Message, userState *state.UserState) error

const (
	MessageFormHandlerName  = "MessageForm"
	LoginFormHandlerName    = "LoginForm"
	PasswordFormHandlerName = "PasswordForm"
)

var (
	MessageHandlers = map[string]PersonalSimpleMessageHandler{
		MessageFormHandlerName:  MessageFormPersonalSimpleMessageHandler,
		LoginFormHandlerName:    LoginFormPersonalSimpleMessageHandler,
		PasswordFormHandlerName: PasswordFormPersonalSimpleMessageHandler,
	}
)

func SimpleMessageHandler(bot *tgbotapi.BotAPI, message *tgbotapi.Message, states *state.States) error {
	userState, err := states.GetState(message.Chat.ID)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to get user state")
	}

	handler, found := MessageHandlers[userState.MessageHandlerName]
	if !found {
		return errorx.IllegalState.New("message handler not found in the current state")
	}

	return handler(bot, message, userState)
}

func MessageFormPersonalSimpleMessageHandler(bot *tgbotapi.BotAPI, message *tgbotapi.Message, userState *state.UserState) error {
	if message.Text != "" {
		userState.OverrideText = message.Text
		reply := tgbotapi.NewMessage(message.Chat.ID, "Текст сообщения заменён")
		reply.ReplyToMessageID = message.MessageID
		_, err := bot.Send(reply)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to send reply")
		}
	}
	if len(message.Photo) > 0 {
		maxPhotoSize := lo.MaxBy(message.Photo, func(a tgbotapi.PhotoSize, b tgbotapi.PhotoSize) bool {
			return a.FileSize > b.FileSize
		})
		userState.Files = append(userState.Files, maxPhotoSize.FileID)

		reply := tgbotapi.NewMessage(message.Chat.ID, "Фото прикреплено")
		reply.ReplyToMessageID = message.MessageID
		reply.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButtonLocation("Отправить сообщение"),
			),
		)
		_, err := bot.Send(reply)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to send reply")
		}
	}
	if message.Location != nil {
		text := userState.CurrentCategory.Message
		if userState.OverrideText != "" {
			text = userState.OverrideText
		}

		reply := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(`
Сообщение добавлено в очередь и будет отправлено при первой возможности.

Пользователь: @%v
Категория: %v
Текст: %v
Локация: %v %v
Файлы: %v
`, message.Chat.UserName, userState.CurrentCategory.Category.Id, text,
			message.Location.Longitude, message.Location.Latitude, userState.Files))
		reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)

		userState.Queue = append(userState.Queue, state.QueueMessage{
			CategoryId: userState.CurrentCategory.Category.Id,
			FileUrls:   slices.Clone(userState.Files),
			Text:       text,
			SentAt:     time.Now(),
		})
		userState.Files = []string{}
		userState.OverrideText = ""
		userState.ResetCurrentCategory()
		userState.MessageHandlerName = ""

		_, err := bot.Send(reply)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to send reply")
		}
	}

	return nil
}

func LoginFormPersonalSimpleMessageHandler(bot *tgbotapi.BotAPI, message *tgbotapi.Message, userState *state.UserState) error {
	if message.Text == "" {
		return errorx.IllegalArgument.New("login expected")
	}

	userState.Credentials.Login = message.Text
	userState.MessageHandlerName = PasswordFormHandlerName

	reply := tgbotapi.NewMessage(message.Chat.ID, "Введите пароль")
	_, err := bot.Send(reply)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to send reply")
	}

	return nil
}

func PasswordFormPersonalSimpleMessageHandler(bot *tgbotapi.BotAPI, message *tgbotapi.Message, userState *state.UserState) error {
	if message.Text == "" {
		return errorx.IllegalArgument.New("password expected")
	}

	userState.Credentials.Password = message.Text
	userState.MessageHandlerName = ""

	client, err := spb.NewReqClient()
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to initialize http client")
	}

	tokenResponse, err := client.Login(userState.Credentials.Login, userState.Credentials.Password)
	if err != nil {
		userState.Credentials = nil
		return errorx.EnhanceStackTrace(err, "failed to login")
	}

	userState.Token = tokenResponse.AccessToken

	reply := tgbotapi.NewMessage(message.Chat.ID, "Авторизация прошла успешно. Учётные данные сохранены")
	_, err = bot.Send(reply)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to send reply")
	}

	return nil
}
