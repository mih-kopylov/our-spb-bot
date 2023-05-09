package queue

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/joomcode/errorx"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"
	"time"
)

const (
	collection = "messages"
)

type FirebaseQueue struct {
	logger *zap.Logger
	fc     *firestore.Client
}

func NewFirebaseQueue(logger *zap.Logger, storage *firestore.Client) *FirebaseQueue {
	return &FirebaseQueue{
		logger: logger,
		fc:     storage,
	}
}

func (q *FirebaseQueue) Add(message *Message) error {
	q.debugMessage(message, "adding message to queue")

	_, err := q.fc.Collection(collection).Doc(message.Id).Create(context.Background(), message)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to put message to queue")
	}

	return nil
}

func (q *FirebaseQueue) Poll() (*Message, error) {
	query := q.fc.Collection(collection).
		Where("status", "==", StatusCreated).
		Where("retryAfter", "<=", time.Now()).
		Limit(1)
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

		q.debugMessage(&result, "message polled")

		return &result, nil
	}

	q.logger.Debug("no appropriate messages found")
	return nil, nil
}

func (q *FirebaseQueue) UpdateEachMessage(userId int64, updater func(*Message)) error {
	query := q.fc.Collection(collection).Where("userId", "==", userId)
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

		updater(&message)
		_, err := q.fc.Collection(collection).Doc(snapshot.Ref.ID).Set(context.Background(), message)
		if err != nil {
			return errorx.EnhanceStackTrace(err, "failed to store message: id=%v", snapshot.Ref.ID)
		}
	}

	return nil

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

func (q *FirebaseQueue) GetMessage(id string) (*Message, error) {
	snapshot, err := q.fc.Collection(collection).Doc(id).Get(context.Background())
	if status.Code(err) == codes.NotFound {
		return nil, errorx.EnhanceStackTrace(err, "message not found")
	}

	var message Message
	err = snapshot.DataTo(&message)
	if err != nil {
		return nil, errorx.EnhanceStackTrace(err, "failed to read a message")
	}

	return &message, nil
}

func (q *FirebaseQueue) DeleteMessage(message *Message) error {
	_, err := q.fc.Collection(collection).Doc(message.Id).Delete(context.Background())
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to delete a message")
	}

	return nil
}

func (q *FirebaseQueue) debugMessage(message *Message, text string) {
	if ce := q.logger.Check(zap.DebugLevel, text); ce != nil {
		messageYaml, err := yaml.Marshal(message)
		if err != nil {
			ce.Write(zap.Int64("userId", message.UserId), zap.Error(err))
		} else {
			ce.Write(zap.String("message", string(messageYaml)))
		}
	}
}
