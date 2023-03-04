package state

import (
	"github.com/goioc/di"
	"github.com/samber/lo"
)

type States struct {
	states map[int64]UserState
}

const (
	BeanId = "States"
)

func RegisterBean() {
	states := States{
		states: map[int64]UserState{},
	}
	_ = lo.Must(di.RegisterBeanInstance(BeanId, &states))
}

// GetState Puts a new user into the context. If the user already exists in the context, it's kept,
func (s *States) GetState(userId int64) (UserState, error) {
	userState, exists := s.states[userId]
	if !exists {
		userState = &memoryUserState{
			userId: userId,
		}
		s.states[userId] = userState
	}
	return userState, nil
}

// todo use abstract Data structure inside, serialized to json

type UserState interface {
	GetUserId() int64
	GetCurrentCategoryNodeId() string
	SetCurrentCategoryNodeId(value string) error
	GetMessageText() string
	SetMessageText(value string) error
	GetFiles() []string
	AddFile(value string) error
	ClearFiles() error
	GetLogin() string
	SetLogin(value string) error
	GetPassword() string
	SetPassword(value string) error
	GetToken() string
	SetToken(value string) error
	GetMessageHandlerName() string
	SetMessageHandlerName(value string) error
}
