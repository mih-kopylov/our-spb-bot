package queue

import (
	"time"
)

type MessageQueue interface {
	Add(message *Message) error
	Poll() (*Message, error)
	UserMessagesCount(userId int64) (map[Status]int, error)
	ResetAwaitingAuthorizationMessages(userId int64) error
	UpdateEachMessage(userId int64, updater func(*Message)) error
	GetMessage(id string) (*Message, error)
	DeleteMessage(message *Message) error
}

type Message struct {
	Id              string    `firestore:"id"`
	UserId          int64     `firestore:"userId"`
	CategoryId      int64     `firestore:"categoryId"`
	Files           []string  `firestore:"files"`
	Text            string    `firestore:"text"`
	Longitude       float64   `firestore:"longitude"`
	Latitude        float64   `firestore:"latitude"`
	CreatedAt       time.Time `firestore:"createdAt"`
	LastTriedAt     time.Time `firestore:"lastTriedAt"`
	Tries           int       `firestore:"tries"`
	RetryAfter      time.Time `firestore:"retryAfter"`
	FailDescription string    `firestore:"failDescription"`
	Status          Status    `firestore:"status"`
}

const (
	MaxTries = 5
)

type Status string

const (
	// StatusCreated for messages that are awaiting to be sent
	StatusCreated Status = "created"
	// StatusFailed for messages that failed to be sent and need to be investigated
	StatusFailed Status = "failed"
	// StatusAwaitingAuthorization for messages that are awaiting user's authorization
	StatusAwaitingAuthorization Status = "awaiting_authorization"
)
