package state

type States interface {
	// GetState Puts a new user into the context. If the user already exists in the context, it's kept,
	GetState(userId int64) (*UserState, error)
	SetState(state *UserState) error
}

type UserState struct {
	UserId                int64    `firestore:"userId"`
	CurrentCategoryNodeId string   `firestore:"currentCategoryNodeId"`
	MessageText           string   `firestore:"messageText"`
	Files                 []string `firestore:"files"`
	Login                 string   `firestore:"login"`
	Password              string   `firestore:"password"`
	Token                 string   `firestore:"token"`
	MessageHandlerName    string   `firestore:"messageHandlerName"`
	SentMessagesCount     int      `firestore:"sentMessagesCount"`
}
