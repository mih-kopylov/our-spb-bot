package queue

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/goioc/di"
	"github.com/joomcode/errorx"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"reflect"
	"time"
)

const (
	BeanId     = "Queue"
	collection = "messages"
)

type FirebaseQueue struct {
	fc *firestore.Client `di.inject:"Storage"`
}

func RegisterQueueBean() {
	_ = lo.Must(di.RegisterBean(BeanId, reflect.TypeOf((*FirebaseQueue)(nil))))
}

func (q *FirebaseQueue) Add(message *Message) error {
	_, err := q.fc.Collection(collection).Doc(message.Id).Create(context.Background(), message)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to put message to queue")
	}

	return nil
}

func (q *FirebaseQueue) Poll() (*Message, error) {
	query := q.fc.Collection(collection).
		Where("status", "==", StatusCreated).
		Where("retryAfter", "<=", time.Now())
	snapshots, err := query.Documents(context.Background()).GetAll()
	if err != nil {
		return nil, errorx.EnhanceStackTrace(err, "failed to poll a message")
	}

	var result Message
	for _, snapshot := range snapshots {
		err := snapshot.DataTo(&result)
		if err != nil {
			return nil, errorx.EnhanceStackTrace(err, "failed to deserialize message: id=%v", snapshot.Ref.ID)
		}

		if result.RetryAfter.After(time.Now()) {
			// will be sent later
			continue
		}

		_, err = q.fc.Collection(collection).Doc(snapshot.Ref.ID).Delete(context.Background())
		if err != nil {
			return nil, errorx.EnhanceStackTrace(err, "failed to delete pulled message")
		}

		return &result, nil
	}

	logrus.Debug("no appropriate messages found")
	return nil, nil

}

func (q *FirebaseQueue) ResetAwaitingAuthorizationMessages(userId int64) error {
	query := q.fc.Collection(collection).
		Where("userId", "==", userId).
		Where("status", "==", StatusAwaitingAuthorization)
	documents := query.Documents(context.Background())
	snapshots, err := documents.GetAll()
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to filter messages")
	}

	var message Message
	for _, snapshot := range snapshots {
		err = snapshot.DataTo(&message)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to deserialize message: id=%v", snapshot.Ref.ID)
		}

		message.Status = StatusCreated
		_, err := q.fc.Collection(collection).Doc(snapshot.Ref.ID).Set(context.Background(), message)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to store message: id=%v", snapshot.Ref.ID)
		}
	}

	return nil
}

func (q *FirebaseQueue) UserMessagesCount(userId int64) (map[Status]int, error) {
	query := q.fc.Collection(collection).Where("userId", "==", userId)
	documents := query.Documents(context.Background())
	snapshots, err := documents.GetAll()
	if err != nil {
		return nil, errorx.EnhanceStackTrace(err, "failed to filter messages")
	}

	result := map[Status]int{}
	var message Message
	for _, snapshot := range snapshots {
		err := snapshot.DataTo(&message)
		if err != nil {
			return nil, errorx.EnhanceStackTrace(err, "failed to read a message")
		}
		result[message.Status]++
	}

	return result, nil
}
