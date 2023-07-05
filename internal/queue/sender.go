package queue

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/config"
	"github.com/mih-kopylov/our-spb-bot/internal/spb"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/mih-kopylov/our-spb-bot/internal/util"
	"github.com/mih-kopylov/our-spb-bot/pkg/tgbot"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"math"
	"time"
)

type MessageSender struct {
	logger        *zap.Logger
	stateManager  state.Manager
	queue         MessageQueue
	spbClient     spb.Client
	api           *tgbotapi.BotAPI
	service       *tgbot.Service
	enabled       bool
	sleepDuration time.Duration
}

var (
	Errors                    = errorx.NewNamespace("Sender")
	ErrNoAccounts             = Errors.NewType("NoAccounts")
	ErrAllAccountsDisabled    = Errors.NewType("AllAccountsDisabled")
	ErrAllAccountsRateLimited = Errors.NewType("AllAccountsRateLimited")
)

func NewMessageSender(logger *zap.Logger, conf *config.Config, stateManager state.Manager, queue MessageQueue, spbClient spb.Client,
	api *tgbotapi.BotAPI, service *tgbot.Service) *MessageSender {
	return &MessageSender{
		logger:        logger,
		stateManager:  stateManager,
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
		s.logger.Info("starting sender")
		go func() {
			for {
				s.sendNextMessage()
			}
		}()
	} else {
		s.logger.Warn("sender is disabled")
	}
	return nil
}

func (s *MessageSender) sendNextMessage() {
	s.logger.Debug("polling messages")
	message, err := s.queue.Poll()
	if err != nil {
		s.logger.Error("failed to poll next message",
			zap.Duration("sleep", s.sleepDuration),
			zap.Error(err))
		time.Sleep(s.sleepDuration)
		return
	}
	if message == nil {
		s.logger.Debug("no messages found, sleeping for " + s.sleepDuration.String())
		time.Sleep(s.sleepDuration)
		return
	}

	s.logger.Debug("message found", zap.String("id", message.Id))

	userState, err := s.stateManager.GetState(message.UserId)
	if err != nil {
		s.logger.Error("failed to get user state",
			zap.Error(err))
		s.returnMessage(message, StatusFailed, "failed to get user state")
		return
	}

	account, appropriateAccountsCount, err := s.chooseAccount(userState, message)
	if err != nil {
		if errorx.IsOfType(err, ErrNoAccounts) || errorx.IsOfType(err, ErrAllAccountsDisabled) {
			s.logger.Error("failed to choose an account",
				zap.Error(err))
			s.returnMessage(message, StatusFailed, "no authorized accounts found")
			return
		}
		if errorx.IsOfType(err, ErrAllAccountsRateLimited) {
			s.logger.Info("user is rate limited",
				zap.String("id", message.Id))

			enabledAccounts := lo.Filter(userState.Accounts, func(item state.Account, _ int) bool {
				return item.State == state.AccountStateEnabled
			})
			message.RetryAfter = lo.MinBy(enabledAccounts, func(a state.Account, b state.Account) bool {
				return a.RateLimitedUntil.Before(b.RateLimitedUntil)
			}).RateLimitedUntil
			s.returnMessage(message, StatusCreated, "user is rate limited")
			return
		}
	}

	s.logger.Debug("creating a request",
		zap.String("id", message.Id))
	request, err := s.spbClient.CreateSendProblemRequest(message.CategoryId, message.Text, message.Latitude, message.Longitude)
	if err != nil {
		s.logger.Error("failed to create a request",
			zap.Error(err))
		s.returnMessage(message, StatusFailed, "failed to create a request: "+err.Error())
		return
	}

	s.logger.Debug("getting files",
		zap.String("id", message.Id))
	files, err := s.getFiles(message)
	if err != nil {
		s.logger.Error("failed to get message files",
			zap.Error(err))
		s.returnMessage(message, StatusFailed, "failed to get messages files "+err.Error())
		return
	}

	s.logger.Debug("sending message",
		zap.String("id", message.Id))
	sentMessageResponse, err := s.spbClient.Send(account.Token, request, files)
	if err != nil {
		s.logger.Warn("failed to send message",
			zap.String("id", message.Id),
			zap.Error(err))
		s.handleMessageSendingError(err, userState, account, appropriateAccountsCount, message)
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
		s.logger.Warn("failed to send reply",
			zap.String("id", message.Id),
			zap.Error(err))
		return
	}

	userState.SentMessagesCount++
	err = s.stateManager.SetState(userState)
	if err != nil {
		s.logger.Error("failed to set user state",
			zap.Error(err))
	}

	s.logger.Debug("message sent",
		zap.String("id", message.Id))
}

func (s *MessageSender) handleMessageSendingError(err error, userState *state.UserState, account *state.Account, appropriateAccountsCount int, message *Message) {
	if errorx.IsOfType(err, spb.ErrUnauthorized) {
		account.Token = ""
		err = s.stateManager.SetState(userState)
		if err != nil {
			s.logger.Error("failed to set user state",
				zap.Error(err))
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
		year, month, day := time.Now().In(util.SpbLocation).AddDate(0, 0, 1).Date()
		hour, min, _ := account.RateLimitNextDayTime.In(util.SpbLocation).Clock()
		nextTryTime := time.Date(year, month, day, hour, min, 0, 0, util.SpbLocation)

		account.RateLimitedUntil = nextTryTime
		err = s.stateManager.SetState(userState)
		if err != nil {
			s.logger.Error("failed to set user state",
				zap.Error(err))
			s.returnMessageIncreaseTries(message, StatusFailed, "failed to set user state: "+err.Error())
		} else {
			if appropriateAccountsCount == 1 {
				//delay message only in case there are no other accounts that may be used to sent it
				message.RetryAfter = nextTryTime
			}
			s.returnMessage(message, StatusCreated, "too many requests")
		}
	} else {
		s.returnMessageIncreaseTries(message, StatusFailed, "failed to send a message: "+err.Error())
	}
}

func (s *MessageSender) tryReauthorize(userState *state.UserState, message *Message, account *state.Account) error {
	if account.Login == "" {
		account.State = state.AccountStateDisabled
		err := s.stateManager.SetState(userState)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to set user state")
		}

		return errorx.IllegalState.New("user not authorized")
	}

	s.logger.Info("refreshing token",
		zap.String("id", message.Id),
		zap.String("login", account.Login))
	tokenResponse, err := s.spbClient.Login(account.Login, account.Password)
	if err != nil {
		account.Login = ""
		account.Password = ""
		account.State = state.AccountStateDisabled
		err2 := s.stateManager.SetState(userState)
		if err2 != nil {
			return errorx.EnhanceStackTrace(err2, "failed to set user state")
		}

		return errorx.EnhanceStackTrace(err, "failed to reauthorize")
	}

	s.logger.Info("new token obtained",
		zap.String("id", message.Id))
	account.Token = tokenResponse.AccessToken
	err = s.stateManager.SetState(userState)
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
		s.logger.Error("failed to return a failed message back to queue")
	} else {
		s.logger.Info("message returned to the queue",
			zap.String("id", message.Id),
			zap.Any("status", message.Status),
			zap.String("failDescription", message.FailDescription))
	}
}

func (s *MessageSender) getFiles(message *Message) (map[string][]byte, error) {
	result := map[string][]byte{}
	for i, fileId := range message.Files {
		fileBytes, err := s.service.DownloadFile(fileId)
		if err != nil {
			return nil, err
		}

		result[fmt.Sprintf("file_%v.jpg", i)] = fileBytes
	}
	return result, nil
}

func (s *MessageSender) shiftLongitudeMeters(latitude float64, longitude float64, meters int) float64 {
	metersFloat := float64(meters)
	earthRadius := 6378.137                                      //radius of the Earth in kilometers
	oneMeter := (1 / ((2 * math.Pi / 360) * earthRadius)) / 1000 //1 meter in degrees

	return longitude - (metersFloat*oneMeter)/math.Cos(latitude*(math.Pi/180))
}

func (s *MessageSender) chooseAccount(userState *state.UserState, message *Message) (*state.Account, int, error) {
	if len(userState.Accounts) == 0 {
		return nil, 0, ErrNoAccounts.New("no accounts found")
	}

	if len(lo.Filter(userState.Accounts, func(item state.Account, index int) bool {
		return item.State == state.AccountStateEnabled
	})) == 0 {
		return nil, 0, ErrAllAccountsDisabled.New("all accounts disabled")
	}

	var appropriateAccounts []*state.Account

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
				s.logger.Warn("failed to authorize with account",
					zap.String("id", message.Id),
					zap.String("login", account.Login))
				continue
			}
		}

		appropriateAccounts = append(appropriateAccounts, &userState.Accounts[i])
	}

	if len(appropriateAccounts) == 0 {
		return nil, 0, ErrAllAccountsRateLimited.New("all accounts are rate limited")
	}

	return appropriateAccounts[0], len(appropriateAccounts), nil
}
