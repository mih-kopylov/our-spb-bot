package queue

import "time"

type MessageQueue interface {
	Add(message *Message) error
	Poll() (*Message, error)
	WaitingCount(userId int64) int
}

type Message struct {
	Id          string
	UserId      int64
	CategoryId  int64
	FileUrls    []string
	Text        string
	Location    Location
	CreatedAt   time.Time
	LastTriedAt time.Time
	Tries       int
	Retryable   bool
	RetryAfter  time.Time
	FailStatus  FailStatus
}

type Location struct {
	Longitude float64
	Latitude  float64
}

type FailStatus int

const (
	MaxTries = 5
)

const (
	FailStatusNone FailStatus = iota
	FailStatusNoState
	FailStatusUnauthorized
	FailStatusBadRequest
	FailStatusRequestNotCreated
	FailStatusTooManyRequests
	FailStatusExpectingNotBuildingCoords
)
