package state

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/joomcode/errorx"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"
	"strconv"
	"time"
)

const (
	collection = "states"
)

type FirebaseManager[S BaseUserState] struct {
	logger   *zap.Logger
	storage  *firestore.Client
	newState S
}

func NewFirebaseManager[S BaseUserState](logger *zap.Logger, storage *firestore.Client, newState S) *FirebaseManager[S] {
	return &FirebaseManager[S]{
		logger:   logger,
		storage:  storage,
		newState: newState,
	}
}

func (f *FirebaseManager[S]) GetState(userId int64) (*S, error) {
	doc := f.storage.Collection(collection).Doc(strconv.FormatInt(userId, 10))
	snapshot, err := doc.Get(context.Background())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			result := f.newState
			result.UserId = userId
			return &result, nil
		}

		if status.Code(err) == codes.ResourceExhausted {
			return nil, ErrRateLimited.New("failed to get user state")
		}

		return nil, errorx.EnhanceStackTrace(err, "failed to get state document snapshot")
	}

	//var state UserState
	err = snapshot.DataTo(&newState)
	if err != nil {
		return nil, errorx.EnhanceStackTrace(err, "failed to deserialize user state data: userId=%v", userId)
	}

	newState.logger = f.logger
	f.debugUserState(&newState, "read user state")

	return &state, nil
}

func (f *FirebaseManager[S]) SetState(state *UserState) error {
	state.LastAccessAt = time.Now()
	f.debugUserState(state, "saving user state")

	wr, err := f.storage.Collection(collection).Doc(strconv.FormatInt(state.UserId, 10)).Set(context.Background(), state)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state: userId=%v", state.UserId)
	} else {
		f.logger.Debug("user state saved",
			zap.Int64("userId", state.UserId),
			zap.Time("updateTime", wr.UpdateTime))
	}

	return nil
}

func (f *FirebaseManager[S]) GetAllStates() ([]*UserState, error) {
	snapshots, err := f.storage.Collection(collection).Documents(context.Background()).GetAll()
	if err != nil {
		return nil, errorx.EnhanceStackTrace(err, "failed to get all user states")
	}

	var states []*UserState
	for _, snapshot := range snapshots {
		var state UserState
		err = snapshot.DataTo(&state)
		if err != nil {
			return nil, errorx.EnhanceStackTrace(err, "failed to deserialize user state data: userId=%v", snapshot.Ref.ID)
		}
		states = append(states, &state)
	}

	return states, nil
}

func (f *FirebaseManager[S]) debugUserState(state *S, message string) {
	if ce := f.logger.Check(zap.DebugLevel, message); ce != nil {
		stateYaml, err := yaml.Marshal(state)
		if err != nil {
			ce.Write(zap.Int64("userId", state.UserId), zap.Error(err))
		} else {
			ce.Write(zap.String("state", string(stateYaml)))
		}
	}
}
