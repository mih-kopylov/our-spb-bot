package state

import (
	"cloud.google.com/go/firestore"
	"context"
	"encoding/base64"
	firebase "firebase.google.com/go/v4"
	"github.com/goioc/di"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/config"
	"github.com/samber/lo"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
)

const (
	BeanId = "States"
)

type firebaseStates struct {
	ctx context.Context
	fc  *firestore.Client
}

func RegisterBean(conf *config.Config) {
	ctx := context.Background()

	fbConfig := firebase.Config{
		ProjectID: "ourspbbot",
	}
	serviceAccountJson := lo.Must(base64.StdEncoding.DecodeString(conf.FirebaseServiceAccount))
	serviceAccountOption := option.WithCredentialsJSON(serviceAccountJson)
	app := lo.Must(firebase.NewApp(ctx, &fbConfig, serviceAccountOption))
	fc := lo.Must(app.Firestore(ctx))
	repo := firebaseStates{
		fc:  fc,
		ctx: ctx,
	}
	_ = lo.Must(di.RegisterBeanInstance(BeanId, &repo))
}

func (f *firebaseStates) GetState(userId int64) (*UserState, error) {
	doc := f.fc.Collection("states").Doc(strconv.FormatInt(userId, 10))
	snapshot, err := doc.Get(f.ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			newState := UserState{
				UserId: userId,
			}
			_, err := doc.Create(f.ctx, &newState)
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

func (f *firebaseStates) SetState(state *UserState) error {
	_, err := f.fc.Collection("states").Doc(strconv.FormatInt(state.UserId, 10)).Set(f.ctx, state)
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to set user state: userId=%v", state.UserId)
	}

	return nil
}
