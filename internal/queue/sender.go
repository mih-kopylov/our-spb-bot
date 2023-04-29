package queue

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/imroc/req/v3"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/bot/service"
	"github.com/mih-kopylov/our-spb-bot/internal/config"
	"github.com/mih-kopylov/our-spb-bot/internal/spb"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"math"
	"net/http"
	"time"
)

type MessageSender struct {
	states        state.States
	queue         MessageQueue
	spbClient     spb.Client
	api           *tgbotapi.BotAPI
	service       *service.Service
	enabled       bool
	sleepDuration time.Duration
}

var (
	spbLocation = time.FixedZone("UTC+3", 3*60*60)
)

var (
	Errors                    = errorx.NewNamespace("Sender")
	ErrNoAccounts             = Errors.NewType("NoAccounts")
	ErrAllAccountsDisabled    = Errors.NewType("AllAccountsDisabled")
	ErrAllAccountsRateLimited = Errors.NewType("AllAccountsRateLimited")
)

func NewMessageSender(conf *config.Config, states state.States, queue MessageQueue, spbClient spb.Client,
	api *tgbotapi.BotAPI, service *service.Service) *MessageSender {
	return &MessageSender{
		states:        states,
		queue:         queue,
		spbClient:     spbClient,
		api:           api,
		service:       service,
		enabled:       conf.SenderEnabled,
		sleepDuration: conf.SenderSleepDuration,
	}
}

func (s *MessageSender) Start() error {
	if s.enabled {
		logrus.Info("starting sender")
		go func() {
			for {
				s.sendNextMessage()
			}
		}()
	} else {
		logrus.Warn("sender is disabled")
	}
	return nil
}

func (s *MessageSender) sendNextMessage() {
	logrus.Debug("polling messages")
	message, err := s.queue.Poll()
	if err != nil {
		logrus.Error(errorx.EnhanceStackTrace(err, "failed to poll next message, sleeping for "+s.sleepDuration.String()))
		time.Sleep(s.sleepDuration)
		return
	}
	if message == nil {
		logrus.Debug("no messages found, sleeping for " + s.sleepDuration.String())
		time.Sleep(s.sleepDuration)
		return
	}

	logrus.WithField("id", message.Id).Debug("message found")
	userState, err := s.states.GetState(message.UserId)
	if err != nil {
		logrus.Error(errorx.EnhanceStackTrace(err, "failed to get user state"))
		s.returnMessage(message, StatusFailed, "failed to get user state")
		return
	}

	account, err := s.chooseAccount(userState, message)
	if err != nil {
		if errorx.IsOfType(err, ErrNoAccounts) || errorx.IsOfType(err, ErrAllAccountsDisabled) {
			logrus.Error(errorx.EnhanceStackTrace(err, "failed to choose an account"))
			s.returnMessage(message, StatusFailed, "no authorized accounts found")
			return
		}
		if errorx.IsOfType(err, ErrAllAccountsRateLimited) {
			logrus.WithField("id", message.Id).Info("user is rate limited")
			message.RetryAfter = userState.Accounts[0].RateLimitedUntil
			s.returnMessage(message, StatusCreated, "user is rate limited")
			return
		}
	}

	logrus.WithField("id", message.Id).Debug("creating a request")
	request, err := s.spbClient.CreateSendProblemRequest(message.CategoryId, message.Text, message.Latitude, message.Longitude)
	if err != nil {
		logrus.Error(errorx.EnhanceStackTrace(err, "failed to create a request"))
		s.returnMessage(message, StatusFailed, "failed to create a request: "+err.Error())
		return
	}

	logrus.WithField("id", message.Id).Debug("getting files")
	files, err := s.getFiles(message)
	if err != nil {
		logrus.Error(errorx.EnhanceStackTrace(err, "failed to get message files"))
		s.returnMessage(message, StatusFailed, "failed to get messages files "+err.Error())
		return
	}

	logrus.WithField("id", message.Id).Debug("sending message")
	sentMessageResponse, err := s.spbClient.Send(account.Token, request, files)
	if err != nil {
		logrus.WithField("id", message.Id).Warn(errorx.EnhanceStackTrace(err, "failed to send message"))
		s.handleMessageSendingError(err, userState, account, message)
		return
	}

	err = s.service.SendMessage(&tgbotapi.Chat{ID: message.UserId}, fmt.Sprintf(`Обращение отправлено.
Пользователь: %v
Id: %v
Ссылка: https://gorod.gov.spb.ru/problems/%v/`,
		account.Login,
		message.Id,
		sentMessageResponse.Id,
	))
	if err != nil {
		logrus.WithField("id", message.Id).Warn(errorx.EnhanceStackTrace(err, "failed to send reply"))
		return
	}

	userState.SentMessagesCount++
	err = s.states.SetState(userState)
	if err != nil {
		logrus.Error(errorx.EnhanceStackTrace(err, "failed to set user state"))
	}

	logrus.WithField("id", message.Id).Debug("message sent")
}

func (s *MessageSender) handleMessageSendingError(err error, userState *state.UserState, account *state.Account, message *Message) {
	if errorx.IsOfType(err, spb.ErrUnauthorized) {
		account.Token = ""
		err = s.states.SetState(userState)
		if err != nil {
			logrus.Error(errorx.EnhanceStackTrace(err, "failed to set user state"))
			s.returnMessageIncreaseTries(message, StatusFailed, "failed to set user state: "+err.Error())
		} else {
			message.RetryAfter = time.Now()
			s.returnMessageIncreaseTries(message, StatusCreated, "token expired")
		}
	} else if errorx.IsOfType(err, spb.ErrExpectingNotBuildingCoords) {
		message.RetryAfter = time.Now()
		message.Longitude = s.shiftLongitudeMeters(message.Latitude, message.Longitude, 50)
		s.returnMessageIncreaseTries(message, StatusCreated, "service expects coordinates outside a building")
	} else if errorx.IsOfType(err, spb.ErrBadRequest) {
		s.returnMessageIncreaseTries(message, StatusFailed, err.Error())
	} else if errorx.IsOfType(err, spb.ErrTooManyRequests) {
		year, month, day := time.Now().In(spbLocation).AddDate(0, 0, 1).Date()
		nextTryTime := time.Date(year, month, day, 1, 0, 0, 0, spbLocation)

		account.RateLimitedUntil = nextTryTime
		err = s.states.SetState(userState)
		if err != nil {
			logrus.Error(errorx.EnhanceStackTrace(err, "failed to set user state"))
			s.returnMessageIncreaseTries(message, StatusFailed, "failed to set user state: "+err.Error())
		} else {
			message.RetryAfter = nextTryTime
			s.returnMessage(message, StatusCreated, "too many requests")
		}
	} else {
		s.returnMessageIncreaseTries(message, StatusFailed, "failed to send a message: "+err.Error())
	}
}

func (s *MessageSender) tryReauthorize(userState *state.UserState, message *Message, account *state.Account) error {
	if account.Login == "" {
		account.State = state.AccountStateDisabled
		err := s.states.SetState(userState)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to set user state")
		}

		return errorx.IllegalState.New("user not authorized")
	}

	logrus.WithField("id", message.Id).
		WithField("login", account.Login).
		Info("refreshing token")
	tokenResponse, err := s.spbClient.Login(account.Login, account.Password)
	if err != nil {
		account.Login = ""
		account.Password = ""
		account.State = state.AccountStateDisabled
		err2 := s.states.SetState(userState)
		if err2 != nil {
			return errorx.EnhanceStackTrace(err2, "failed to set user state")
		}

		return errorx.EnhanceStackTrace(err, "failed to reauthorize")
	}

	logrus.WithField("id", message.Id).Info("new token obtained")
	account.Token = tokenResponse.AccessToken
	err = s.states.SetState(userState)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state")
	}

	return nil
}

func (s *MessageSender) returnMessageIncreaseTries(message *Message, status Status, description string) {
	message.Tries++
	if message.Tries >= MaxTries {
		status = StatusFailed
	}
	s.returnMessage(message, status, description)
}

func (s *MessageSender) returnMessage(message *Message, status Status, description string) {
	message.LastTriedAt = time.Now()
	message.Status = status
	message.FailDescription = description
	err := s.queue.Add(message)
	if err != nil {
		logrus.Error(errorx.IllegalState.New("failed to return a failed message back to queue"))
	} else {
		logrus.WithField("id", message.Id).
			WithField("status", message.Status).
			WithField("failDescription", message.FailDescription).
			Info("message returned to the queue")
	}
}

func (s *MessageSender) getFiles(message *Message) (map[string][]byte, error) {
	result := map[string][]byte{}
	for i, fileId := range message.Files {
		fileUrl, err := s.api.GetFileDirectURL(fileId)
		if err != nil {
			return nil, errorx.EnhanceStackTrace(err, "failed to get file url")
		}

		response, err := req.R().Get(fileUrl)
		if err != nil {
			return nil, errorx.EnhanceStackTrace(err, "failed to donwload file")
		}

		responseBytes, err := response.ToBytes()
		if err != nil {
			return nil, errorx.EnhanceStackTrace(err, "failed to get response bytes: code=%v", response.StatusCode)
		}

		if response.StatusCode != http.StatusOK {
			return nil, errorx.IllegalArgument.New("failed to download file: fileId=%v, fileUrl=%v, response=%v", fileId, fileUrl, responseBytes)
		}

		result[fmt.Sprintf("file_%v.jpg", i)] = responseBytes
	}
	return result, nil
}

func (s *MessageSender) shiftLongitudeMeters(latitude float64, longitude float64, meters int) float64 {
	metersFloat := float64(meters)
	earthRadius := 6378.137                                      //radius of the Earth in kilometers
	oneMeter := (1 / ((2 * math.Pi / 360) * earthRadius)) / 1000 //1 meter in degrees

	return longitude - (metersFloat*oneMeter)/math.Cos(latitude*(math.Pi/180))
}

func (s *MessageSender) chooseAccount(userState *state.UserState, message *Message) (*state.Account, error) {
	if len(userState.Accounts) == 0 {
		return nil, ErrNoAccounts.New("no accounts found")
	}

	if len(lo.Filter(userState.Accounts, func(item state.Account, index int) bool {
		return item.State == state.AccountStateEnabled
	})) == 0 {
		return nil, ErrAllAccountsDisabled.New("all accounts disabled")
	}

	for i, account := range userState.Accounts {
		if account.State == state.AccountStateDisabled {
			continue
		}

		if account.RateLimitedUntil.After(time.Now()) {
			continue
		}

		if account.Token == "" {
			err := s.tryReauthorize(userState, message, &userState.Accounts[i])
			if err != nil {
				logrus.WithField("id", message.Id).
					WithField("login", account.Login).
					Warn("failed to authorize with account")
				continue
			}
		}

		return &userState.Accounts[i], nil
	}

	return nil, ErrAllAccountsRateLimited.New("all accounts are rate limited")
}
