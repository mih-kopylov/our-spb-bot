package state

import "golang.org/x/exp/slices"

type memoryUserState struct {
	userId                int64
	currentCategoryNodeId string
	messageText           string
	files                 []string
	login                 string
	password              string
	token                 string
	messageHandlerName    string
}

func (s *memoryUserState) GetUserId() int64 {
	return s.userId
}

func (s *memoryUserState) GetCurrentCategoryNodeId() string {
	return s.currentCategoryNodeId
}

func (s *memoryUserState) SetCurrentCategoryNodeId(value string) error {
	s.currentCategoryNodeId = value
	return nil
}

func (s *memoryUserState) GetMessageText() string {
	return s.messageText
}

func (s *memoryUserState) SetMessageText(value string) error {
	s.messageText = value
	return nil
}

func (s *memoryUserState) GetFiles() []string {
	return slices.Clone(s.files)
}

func (s *memoryUserState) AddFile(file string) error {
	s.files = append(s.files, file)
	return nil
}

func (s *memoryUserState) ClearFiles() error {
	s.files = nil
	return nil
}

func (s *memoryUserState) GetLogin() string {
	return s.login
}

func (s *memoryUserState) SetLogin(value string) error {
	s.login = value
	return nil
}

func (s *memoryUserState) GetPassword() string {
	return s.password
}

func (s *memoryUserState) SetPassword(value string) error {
	s.password = value
	return nil
}

func (s *memoryUserState) GetToken() string {
	return s.token
}

func (s *memoryUserState) SetToken(value string) error {
	s.token = value
	return nil
}

func (s *memoryUserState) GetMessageHandlerName() string {
	return s.messageHandlerName
}

func (s *memoryUserState) SetMessageHandlerName(value string) error {
	s.messageHandlerName = value
	return nil
}
