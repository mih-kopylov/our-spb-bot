package queue

import (
	"time"
)

type MessageQueue interface {
	Add(message *Message) error
	Poll() (*Message, error)
	WaitingCount(userId int64) (int, error)
}

type Message struct {
	Id          string     `firestore:"id"`
	UserId      int64      `firestore:"userId"`
	CategoryId  int64      `firestore:"categoryId"`
	FileUrls    []string   `firestore:"fileUrls"`
	Text        string     `firestore:"text"`
	Longitude   float64    `firestore:"longitude"`
	Latitude    float64    `firestore:"latitude"`
	CreatedAt   time.Time  `firestore:"createdAt"`
	LastTriedAt time.Time  `firestore:"lastTriedAt"`
	Tries       int        `firestore:"tries"`
	Retryable   bool       `firestore:"retryable"`
	RetryAfter  time.Time  `firestore:"retryAfter"`
	FailStatus  FailStatus `firestore:"failStatus"`
}

type FailStatus int

const (
	MaxTries = 5
)

const (
	FailStatusNone FailStatus = iota
	FailStatusNoState
	FailStatusTokenExpired
	FailStatusUnauthorized
	FailStatusBadRequest
	FailStatusRequestNotCreated
	FailStatusTooManyRequests
	FailStatusExpectingNotBuildingCoords
)
