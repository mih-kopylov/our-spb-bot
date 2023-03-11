package queue

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/goioc/di"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/samber/lo"
	"reflect"
	"time"
)

const (
	BeanId     = "Queue"
	collection = "messages"
)

type FirebaseQueue struct {
	states state.States      `di.inject:"States"`
	fc     *firestore.Client `di.inject:"Storage"`
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
		Where("tries", "<", MaxTries)
	snapshots, err := query.Documents(context.Background()).GetAll()
	if err != nil {
		return nil, errorx.EnhanceStackTrace(err, "failed to poll a message")
	}

	var result Message
	for _, snapshot := range snapshots {
		err := snapshot.DataTo(&result)
		if err != nil {
			return nil, errorx.EnhanceStackTrace(err, "failed to deserialize a message")
		}

		if !result.Retryable && result.Tries > 0 {
			// messages tried once or two but then failed with 400
			continue
		}
		if result.RetryAfter.After(time.Now()) {
			// will be sent later
			continue
		}

		_, err = q.fc.Collection(collection).Doc(result.Id).Delete(context.Background())
		if err != nil {
			return nil, errorx.EnhanceStackTrace(err, "failed to delete pulled message")
		}

		return &result, nil
	}

	return nil, nil

}

func (q *FirebaseQueue) WaitingCount(userId int64) (int, error) {
	query := q.fc.Collection(collection).Where("userId", "==", userId)
	documents := query.Documents(context.Background())
	snapshots, err := documents.GetAll()
	if err != nil {
		return 0, errorx.EnhanceStackTrace(err, "failed to get messages count")
	}

	return len(snapshots), nil
}
