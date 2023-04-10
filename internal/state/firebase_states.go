package state

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/goioc/di"
	"github.com/joomcode/errorx"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"
	"reflect"
	"strconv"
	"time"
)

const (
	BeanId     = "States"
	collection = "states"
)

type FirebaseStates struct {
	fc *firestore.Client `di.inject:"Storage"`
}

func RegisterBean() {
	_ = lo.Must(di.RegisterBean(BeanId, reflect.TypeOf((*FirebaseStates)(nil))))
}

func (f *FirebaseStates) GetState(userId int64) (*UserState, error) {
	doc := f.fc.Collection(collection).Doc(strconv.FormatInt(userId, 10))
	snapshot, err := doc.Get(context.Background())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			newState := UserState{
				UserId: userId,
			}
			_, err := doc.Create(context.Background(), &newState)
			if err != nil {
				return nil, errorx.EnhanceStackTrace(err, "failed to create user state: userId=%v", userId)
			}
			return &newState, nil
		}

		if status.Code(err) == codes.ResourceExhausted {
			return nil, ErrRateLimited.New("failed to get user state")
		}

		return nil, errorx.EnhanceStackTrace(err, "failed to get state document snapshot")
	}

	var state UserState
	err = snapshot.DataTo(&state)
	if err != nil {
		return nil, errorx.EnhanceStackTrace(err, "failed to deserialize user state data: userId=%v", userId)
	}

	if state.Login != "" {
		_, exists := lo.Find(state.Accounts, func(item Account) bool {
			return item.Login == state.Login
		})
		if !exists {
			state.Accounts = []Account{{
				Login:            state.Login,
				Password:         state.Password,
				Token:            state.Token,
				RateLimitedUntil: time.Now(),
				State:            AccountStateEnabled,
			}}
		}
	}

	f.debugUserState(&state, "read user state")

	return &state, nil
}

func (f *FirebaseStates) SetState(state *UserState) error {
	f.debugUserState(state, "saving user state")

	wr, err := f.fc.Collection(collection).Doc(strconv.FormatInt(state.UserId, 10)).Set(context.Background(), state)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state: userId=%v", state.UserId)
	} else {
		logrus.WithField("userId", state.UserId).
			WithField("updateTime", wr.UpdateTime.Format(time.RFC3339Nano)).
			Debug("user state saved")
	}

	return nil
}

func (f *FirebaseStates) debugUserState(state *UserState, message string) {
	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		stateYaml, err := yaml.Marshal(state)
		if err != nil {
			logrus.WithField("userId", state.UserId).Error("failed to serialize user state")
		} else {
			logrus.WithField("state", string(stateYaml)).Debug(message)
		}
	}
}
