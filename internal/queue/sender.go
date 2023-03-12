package queue

import (
	"fmt"
	"github.com/goioc/di"
	"github.com/imroc/req/v3"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/spb"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"math"
	"reflect"
	"time"
)

type MessageSender struct {
	states    state.States `di.inject:"States"`
	queue     MessageQueue `di.inject:"Queue"`
	spbClient spb.Client   `di.inject:"SpbClient"`
}

const (
	SenderBeanId = "Sender"
)

var (
	spbLocation = time.FixedZone("UTC+3", 3*60*60)
)

func RegisterSenderBean() {
	_ = lo.Must(di.RegisterBean(SenderBeanId, reflect.TypeOf((*MessageSender)(nil))))

	lo.Must0(di.RegisterBeanPostprocessor(reflect.TypeOf((*MessageSender)(nil)), func(sender any) error {
		err := sender.(*MessageSender).Start()
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to start sender")
		}

		return nil
	}))
}

func (s *MessageSender) Start() error {
	go func() {
		for {
			logrus.Debug("polling messages")
			message, err := s.queue.Poll()
			if err != nil {
				logrus.Error(errorx.EnhanceStackTrace(err, "failed to poll next message"))
				continue
			}
			if message == nil {
				logrus.Debug("no messages found, sleeping")
				time.Sleep(10 * time.Minute)
				continue
			}

			logrus.WithField("id", message.Id).Debug("message found")
			userState, err := s.states.GetState(message.UserId)
			if err != nil {
				logrus.Error(errorx.EnhanceStackTrace(err, "failed to get user state"))
				s.returnMessage(message, StatusFailed, "failed to get user state")
				continue
			}

			if userState.Token == "" {
				logrus.WithField("id", message.Id).Debug("no token found")
				err = s.tryReauthorize(userState, message)
				if err != nil {
					logrus.WithField("id", message.Id).Error(err)
					continue
				}
			}

			logrus.WithField("id", message.Id).Debug("creating a request")
			request, err := s.spbClient.CreateSendProblemRequest(message.CategoryId, message.Text, message.Latitude, message.Longitude)
			if err != nil {
				logrus.Error(errorx.EnhanceStackTrace(err, "failed to create a request"))
				s.returnMessage(message, StatusFailed, "failed to create a request")
				continue
			}

			logrus.WithField("id", message.Id).Debug("getting files")
			files, err := s.getFiles(message)
			if err != nil {
				logrus.Error(errorx.EnhanceStackTrace(err, "failed to get message files"))
				s.returnMessage(message, StatusFailed, "failed to get messages files")
				continue
			}

			logrus.WithField("id", message.Id).Debug("sending message")
			err = s.spbClient.Send(userState.Token, request, files)
			if err != nil {
				logrus.WithField("id", message.Id).Warn(errorx.EnhanceStackTrace(err, "failed to send message"))
				s.handleMessageSendingError(err, userState, message)
				continue
			}

			userState.SentMessagesCount++
			err = s.states.SetState(userState)
			if err != nil {
				logrus.Error(errorx.EnhanceStackTrace(err, "failed to set user state"))
			}

			logrus.WithField("id", message.Id).Debug("message sent")
		}
	}()
	return nil
}

func (s *MessageSender) handleMessageSendingError(err error, userState *state.UserState, message *Message) {
	if errorx.IsOfType(err, spb.ErrUnauthorized) {
		userState.Token = ""
		err = s.states.SetState(userState)
		if err != nil {
			logrus.Error(errorx.EnhanceStackTrace(err, "failed to set user state"))
			s.returnMessage(message, StatusFailed, "failed to set user state")
		} else {
			message.RetryAfter = time.Now()
			s.returnMessage(message, StatusCreated, "token expired")
		}
	}
	if errorx.IsOfType(err, spb.ErrExpectingNotBuildingCoords) {
		message.RetryAfter = time.Now()
		message.Longitude = s.shiftLongitudeMeters(message.Latitude, message.Longitude, 50)
		s.returnMessage(message, StatusCreated, "service expects not building coordinates")
	}
	if errorx.IsOfType(err, spb.ErrBadRequest) {
		s.returnMessage(message, StatusFailed, err.Error())
	}
	if errorx.IsOfType(err, spb.ErrTooManyRequests) {
		year, month, day := time.Now().AddDate(0, 0, 1).Date()
		message.RetryAfter = time.Date(year, month, day, 1, 0, 0, 0, spbLocation)
		s.returnMessage(message, StatusCreated, "too many requests")
	}
}

func (s *MessageSender) tryReauthorize(userState *state.UserState, message *Message) error {
	if userState.Login == "" {
		s.returnMessage(message, StatusAwaitingAuthorization, "user not authorized")
		return errorx.IllegalState.New("user not authorized")
	}

	logrus.WithField("id", message.Id).Info("refreshing token")
	tokenResponse, err := s.spbClient.Login(userState.Login, userState.Password)
	if err != nil {
		userState.Login = ""
		userState.Password = ""
		err2 := s.states.SetState(userState)
		if err2 != nil {
			s.returnMessage(message, StatusFailed, "failed to set user state")
			return errorx.EnhanceStackTrace(err2, "failed to set user state")
		}

		s.returnMessage(message, StatusAwaitingAuthorization, "failed to reauthorize")
		return errorx.EnhanceStackTrace(err, "failed to reauthorize")
	}

	logrus.WithField("id", message.Id).Info("new token obtained")
	userState.Token = tokenResponse.AccessToken
	err = s.states.SetState(userState)
	if err != nil {
		s.returnMessage(message, StatusFailed, "failed to set user state")
		return errorx.EnhanceStackTrace(err, "failed to set user state")
	}

	return nil
}

func (s *MessageSender) returnMessage(message *Message, status Status, description string) {
	message.Tries++
	if message.Tries >= MaxTries {
		message.Status = StatusFailed
	} else {
		message.Status = status
	}
	message.LastTriedAt = time.Now()
	message.FailDescription = description
	err := s.queue.Add(message)
	if err != nil {
		logrus.Error(errorx.IllegalState.New("failed to return a failed message back to queue"))
	}
}

func (s *MessageSender) getFiles(message *Message) (map[string][]byte, error) {
	result := map[string][]byte{}
	for i, fileUrl := range message.FileUrls {
		response, err := req.R().Get(fileUrl)
		if err != nil {
			return nil, errorx.EnhanceStackTrace(err, "failed to donwload file")
		}

		fileBytes, err := response.ToBytes()
		if err != nil {
			return nil, errorx.EnhanceStackTrace(err, "failed to get response bytes")
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
