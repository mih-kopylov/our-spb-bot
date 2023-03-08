package storage

import (
	"context"
	"encoding/base64"
	firebase "firebase.google.com/go/v4"
	"github.com/goioc/di"
	"github.com/mih-kopylov/our-spb-bot/internal/config"
	"github.com/samber/lo"
	"google.golang.org/api/option"
)

const (
	BeanId = "Storage"
)

func RegisterBean(conf *config.Config) {
	fbConfig := firebase.Config{
		ProjectID: "ourspbbot",
	}
	serviceAccountJson := lo.Must(base64.StdEncoding.DecodeString(conf.FirebaseServiceAccount))
	serviceAccountOption := option.WithCredentialsJSON(serviceAccountJson)
	ctx := context.Background()
	app := lo.Must(firebase.NewApp(ctx, &fbConfig, serviceAccountOption))
	store := lo.Must(app.Firestore(ctx))
	_ = lo.Must(di.RegisterBeanInstance(BeanId, store))
}
