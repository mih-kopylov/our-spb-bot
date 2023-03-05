package queue

import (
	"github.com/goioc/di"
	"github.com/samber/lo"
	"time"
)

const (
	MemoryQueueBeanId = "Queue"
)

type MemoryQueue struct {
	messages  []*Message
	sentCount map[int64]int
}

func RegisterQueueBean() {
	_ = lo.Must(di.RegisterBeanInstance(MemoryQueueBeanId, &MemoryQueue{
		messages:  []*Message{},
		sentCount: map[int64]int{},
	}))
}

func (q *MemoryQueue) Add(message *Message) error {
	q.messages = append(q.messages, message)
	return nil
}

func (q *MemoryQueue) Poll() (*Message, error) {
	message, index, found := lo.FindIndexOf(q.messages, func(item *Message) bool {
		return item.Tries == 0 || (item.Retryable && item.Tries <= MaxTries && item.RetryAfter.Before(time.Now()))
	})

	if !found {
		return nil, nil
	}

	q.messages = append(q.messages[0:index], q.messages[index+1:]...)
	return message, nil
}

func (q *MemoryQueue) SentCount(userId int64) int {
	return q.sentCount[userId]
}

func (q *MemoryQueue) IncreaseSent(userId int64) {
	q.sentCount[userId] = q.sentCount[userId] + 1
}

func (q *MemoryQueue) WaitingCount(userId int64) int {
	return lo.CountBy(q.messages, func(item *Message) bool {
		return item.UserId == userId
	})
}
