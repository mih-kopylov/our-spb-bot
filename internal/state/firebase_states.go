package state

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/goioc/di"
	"github.com/joomcode/errorx"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"reflect"
	"strconv"
)

const (
	BeanId = "States"
)

type FirebaseStates struct {
	fc *firestore.Client `di.inject:"Storage"`
}

func RegisterBean() {
	_ = lo.Must(di.RegisterBean(BeanId, reflect.TypeOf((*FirebaseStates)(nil))))
}

func (f *FirebaseStates) GetState(userId int64) (*UserState, error) {
	doc := f.fc.Collection("states").Doc(strconv.FormatInt(userId, 10))
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
		return nil, errorx.EnhanceStackTrace(err, "failed to get state document snapshot")
	}

	var userState UserState
	err = snapshot.DataTo(&userState)
	if err != nil {
		return nil, errorx.EnhanceStackTrace(err, "failed to deserialize user state data: userId=%v", userId)
	}

	return &userState, nil
}

func (f *FirebaseStates) SetState(state *UserState) error {
	_, err := f.fc.Collection("states").Doc(strconv.FormatInt(state.UserId, 10)).Set(context.Background(), state)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state: userId=%v", state.UserId)
	}

	return nil
}
