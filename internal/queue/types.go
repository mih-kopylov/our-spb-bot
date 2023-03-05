package queue

import "time"

type MessageQueue interface {
	Add(message *Message) error
	Poll() (*Message, error)
	SentCount(userId int64) int
	IncreaseSent(userId int64)
	WaitingCount(userId int64) int
}

type Message struct {
	Id         string
	UserId     int64
	CategoryId int64
	FileUrls   []string
	Text       string
	Location   Location
	CreatedAt  time.Time
	Tries      int
	Retryable  bool
	RetryAfter time.Time
	FailStatus FailStatus
}

type Location struct {
	Longitude float64
	Latitude  float64
}

type FailStatus int

const (
	MaxTries                     = 5
	FailStatusNoState FailStatus = iota
	FailStatusUnauthorized
	FailStatusBadRequest
	FailStatusRequestNotCreated
	FailStatusTooManyRequests
	FailStatusExpectingNotBuildingCoords
)
