package queue

import (
	"github.com/goioc/di"
	"github.com/samber/lo"
)

const (
	BeanId = "Queue"
)

type MemoryQueue struct {
	messages map[int64]*userQueue
}

func RegisterBean() {
	_ = lo.Must(di.RegisterBeanInstance(BeanId, &MemoryQueue{map[int64]*userQueue{}}))
}

func (q *MemoryQueue) getOrCreate(userId int64) *userQueue {
	queue, exists := q.messages[userId]
	if !exists {
		queue = &userQueue{}
		q.messages[userId] = queue
	}
	return queue
}

func (q *MemoryQueue) Add(userId int64, message *Message) error {
	queue := q.getOrCreate(userId)
	queue.messages = append(queue.messages, message)
	return nil
}

func (q *MemoryQueue) Poll(userId int64) (*Message, error) {
	queue := q.getOrCreate(userId)
	if len(queue.messages) == 0 {
		return nil, nil
	}
	return queue.messages[0], nil
}

func (q *MemoryQueue) SentCount(userId int64) int {
	return q.getOrCreate(userId).sent
}

func (q *MemoryQueue) WaitingCount(userId int64) int {
	return len(q.getOrCreate(userId).messages)
}

type userQueue struct {
	messages []*Message
	sent     int
}
