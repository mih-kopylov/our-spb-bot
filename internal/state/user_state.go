package state

type UserState struct {
	BaseUserState

	Accounts          []Account `firestore:"accounts"`
	Categories        string    `firestore:"categories"`
	SentMessagesCount int       `firestore:"sentMessagesCount"`
}
