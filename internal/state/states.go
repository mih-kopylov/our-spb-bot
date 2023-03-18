package state

import "time"

type States interface {
	// GetState Puts a new user into the context. If the user already exists in the context, it's kept,
	GetState(userId int64) (*UserState, error)
	SetState(state *UserState) error
}

type UserState struct {
	UserId                int64     `firestore:"userId"`
	UserName              string    `firestore:"userName"`
	FullName              string    `firestore:"fullName"`
	CurrentCategoryNodeId string    `firestore:"currentCategoryNodeId"`
	MessageText           string    `firestore:"messageText"`
	Files                 []string  `firestore:"files"`
	Login                 string    `firestore:"login"`
	Password              string    `firestore:"password"`
	Token                 string    `firestore:"token"`
	RateLimitedUntil      time.Time `firestore:"rateLimitedUntil"`
	MessageHandlerName    string    `firestore:"messageHandlerName"`
	SentMessagesCount     int       `firestore:"sentMessagesCount"`
}
