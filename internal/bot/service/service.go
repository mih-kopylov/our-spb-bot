package service

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/bot"
)

type Service struct {
	api *tgbotapi.BotAPI
}

func NewService(api *tgbotapi.BotAPI) *Service {
	return &Service{
		api: api,
	}
}

func (s *Service) SendMessage(chat *tgbotapi.Chat, text string) error {
	_, err := s.SendMessageCustom(chat, text, func(reply *tgbotapi.MessageConfig) {})
	return err
}

func (s *Service) SendMessageCustom(chat *tgbotapi.Chat, text string, messageAdjuster func(reply *tgbotapi.MessageConfig)) (*tgbotapi.Message, error) {
	message := tgbotapi.NewMessage(chat.ID, text)
	messageAdjuster(&message)
	result, err := s.api.Send(message)
	if err != nil {
		return nil, errorx.EnhanceStackTrace(err, "failed to send reply")
	}

	return &result, nil
}

func (s *Service) Send(chattable tgbotapi.Chattable) error {
	_, err := s.api.Send(chattable)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to send reply")
	}

	return nil
}

func (s *Service) DeleteMessage(message *tgbotapi.Message) error {
	deleteMessage := tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID)
	resp, err := s.api.Request(deleteMessage)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to delete message: chat=%v, message=%v", message.Chat.ID, message.MessageID)
	}

	if !resp.Ok {
		return bot.ErrFailedToDeleteMessage.New("resp=%v", resp)
	}

	return nil
}

func (s *Service) GetFileDirectUrl(fileId string) (string, error) {
	return s.api.GetFileDirectURL(fileId)
}
