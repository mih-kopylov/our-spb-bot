package queue

import "time"

type MessageQueue interface {
	Add(userId int64, message *Message) error
	Poll(userId int64) (*Message, error)
	SentCount(userId int64) int
	WaitingCount(userId int64) int
}

type Message struct {
	CategoryId int
	FileUrls   []string
	Text       string
	Location   Location
	SentAt     time.Time
}

type Location struct {
	Longitude float64
	Latitude  float64
}
