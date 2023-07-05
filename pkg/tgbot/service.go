package tgbot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/imroc/req/v3"
	"net/http"
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
		return nil, ErrFailedToSendReply.WrapWithNoMessage(err)
	}

	return &result, nil
}

func (s *Service) Send(chattable tgbotapi.Chattable) error {
	_, err := s.api.Send(chattable)
	if err != nil {
		return ErrFailedToSendReply.WrapWithNoMessage(err)
	}

	return nil
}

func (s *Service) DeleteMessage(message *tgbotapi.Message) error {
	return s.DeleteMessageById(message.Chat.ID, message.MessageID)
}

func (s *Service) DeleteMessageById(chatId int64, messageId int) error {
	deleteMessage := tgbotapi.NewDeleteMessage(chatId, messageId)
	resp, err := s.api.Request(deleteMessage)
	if err != nil {
		return ErrFailedToDeleteMessage.Wrap(err, "chat=%v, message=%v", chatId, messageId)
	}

	if !resp.Ok {
		return ErrFailedToDeleteMessage.New("resp=%v", resp)
	}

	return nil
}

func (s *Service) GetFileDirectUrl(fileId string) (string, error) {
	return s.api.GetFileDirectURL(fileId)
}

func (s *Service) SendDocument(chat *tgbotapi.Chat, bytes []byte, name string) error {
	file := tgbotapi.FileBytes{
		Name:  name,
		Bytes: bytes,
	}
	document := tgbotapi.NewDocument(chat.ID, file)
	_, err := s.api.Send(document)
	if err != nil {
		return ErrFailedToUploadFile.WrapWithNoMessage(err)
	}

	return nil
}

func (s *Service) DownloadFile(fileId string) ([]byte, error) {
	fileUrl, err := s.api.GetFileDirectURL(fileId)
	if err != nil {
		return nil, ErrFailedToDownloadFile.Wrap(err, "failed to get file url")
	}

	response, err := req.R().Get(fileUrl)
	if err != nil {
		return nil, ErrFailedToDownloadFile.WrapWithNoMessage(err)
	}

	fileBytes, err := response.ToBytes()
	if err != nil {
		return nil, ErrFailedToDownloadFile.Wrap(err, "failed to get response bytes: code=%v", response.StatusCode)
	}

	if response.StatusCode != http.StatusOK {
		return nil, ErrFailedToDownloadFile.New("fileId=%v, fileUrl=%v, response=%v", fileId, fileUrl, fileBytes)
	}

	return fileBytes, nil
}
