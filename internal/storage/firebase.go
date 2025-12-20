package storage

import (
	"context"
	"encoding/base64"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/mih-kopylov/our-spb-bot/internal/config"
	"google.golang.org/api/option"
)

func NewFirebaseStorage(conf *config.Config) (*firestore.Client, error) {
	fbConfig := firebase.Config{
		ProjectID: "ourspbbot",
	}
	serviceAccountJson, err := base64.StdEncoding.DecodeString(conf.FirebaseServiceAccount)
	if err != nil {
		return nil, err
	}

	serviceAccountOption := option.WithAuthCredentialsJSON(option.ServiceAccount, serviceAccountJson)
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, &fbConfig, serviceAccountOption)
	if err != nil {
		return nil, err
	}

	return app.Firestore(ctx)
}
