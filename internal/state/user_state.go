package state

import (
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/category"
	"time"
)

type States struct {
	states map[int64]*UserState
}

// NewStateIfNotExists Puts a new user into the context. If the user already exists in the context, it's kept,
func (s *States) NewStateIfNotExists(userId int64) (*UserState, error) {
	userState, exists := s.states[userId]

	if !exists {
		tree, err := category.CreateUserCategoryTree()
		if err != nil {
			return nil, errorx.EnhanceStackTrace(err, "failed to create user category tree")
		}
		userState = &UserState{
			UserId:          userId,
			Queue:           []QueueMessage{},
			CurrentCategory: tree,
		}
		s.states[userId] = userState
	}

	return userState, nil
}

func (s *States) GetState(userId int64) (*UserState, error) {
	userState, exists := s.states[userId]
	if !exists {
		return nil, errorx.AssertionFailed.New("no state found: userId=%v", userId)
	}
	return userState, nil
}

func NewStates() *States {
	return &States{
		states: map[int64]*UserState{},
	}
}

type UserState struct {
	UserId             int64
	Queue              []QueueMessage
	SentCount          int
	CurrentCategory    *category.UserCategoryTreeNode
	OverrideText       string
	Files              []string
	Credentials        *Credentials
	Token              string
	MessageHandlerName string
}

type Credentials struct {
	Login    string
	Password string
}

func (s *UserState) ResetCurrentCategory() {
	for {
		if s.CurrentCategory.Parent == nil {
			break
		}
		s.CurrentCategory = s.CurrentCategory.Parent
	}
}

type QueueMessage struct {
	CategoryId int
	FileUrls   []string
	Text       string
	SentAt     time.Time
}
