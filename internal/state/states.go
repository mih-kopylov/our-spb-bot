package state

import (
	"reflect"
	"time"

	"github.com/joomcode/errorx"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type States interface {
	// GetState Reads a user from the storage
	GetState(userId int64) (*UserState, error)
	// SetState Puts a new user into the storage. If the user already exists in the context, it's kept,
	SetState(state *UserState) error
	// GetAllStates Reads all users from the storage
	GetAllStates() ([]*UserState, error)
}

var (
	Errors         = errorx.NewNamespace("States")
	ErrRateLimited = Errors.NewType("RateLimited")
)

type FormField string

const (
	FormFieldLogin               FormField = "login"
	FormFieldCurrentCategoryNode FormField = "currentCategoryNode"
	FormFieldMessageText         FormField = "messageText"
	FormFieldFiles               FormField = "files"
	FormFieldMessageIdFile       FormField = "messageIdFile"
)

type UserState struct {
	logger             *zap.Logger
	UserId             int64          `firestore:"userId"`
	FullName           string         `firestore:"fullName"`
	Accounts           []Account      `firestore:"accounts"`
	MessageHandlerName string         `firestore:"messageHandlerName"`
	SentMessagesCount  int            `firestore:"sentMessagesCount"`
	LastAccessAt       time.Time      `firestore:"lastAccessAt"`
	Form               map[string]any `firestore:"form"`
	Categories         string         `firestore:"categories"`
}

func (s *UserState) ClearForm() {
	s.Form = nil
}

func (s *UserState) SetFormField(key FormField, value any) {
	if s.Form == nil {
		s.Form = map[string]any{}
	}

	s.Form[string(key)] = value
}

func (s *UserState) GetStringFormField(key FormField) string {
	if s.Form == nil {
		return ""
	}

	value, exists := s.Form[string(key)]
	if !exists {
		return ""
	}

	stringValue, ok := value.(string)
	if !ok {
		return ""
	}

	return stringValue
}

func (s *UserState) GetIntFormField(key FormField) int {
	if s.Form == nil {
		return 0
	}

	value, exists := s.Form[string(key)]
	if !exists {
		return 0
	}

	intValue, ok := value.(int)
	if !ok {
		return 0
	}

	return intValue
}

func (s *UserState) GetStringSlice(key FormField) []string {
	if s.Form == nil {
		return nil
	}

	value, exists := s.Form[string(key)]
	if !exists {
		return nil
	}

	typeOfValue := reflect.TypeOf(value)
	if typeOfValue.Kind() == reflect.Slice && typeOfValue.Elem().Kind() == reflect.String {
		valueStringSlice, ok := value.([]string)
		if !ok {
			s.logger.Warn("failed to convert value to string array",
				zap.Int64("userId", s.UserId),
				zap.Any("key", key))
			return nil
		}

		return valueStringSlice
	}

	valueArray, ok := value.([]any)
	if !ok {
		s.logger.Warn("failed to convert value to array",
			zap.Int64("userId", s.UserId),
			zap.Any("key", key))
		return nil
	}

	var sliceStrings []string
	for _, item := range valueArray {
		itemString, ok := item.(string)
		if !ok {
			s.logger.Warn("failed to convert array value item to string",
				zap.Int64("userId", s.UserId),
				zap.Any("item", item),
				zap.Any("key", key))
		} else {
			sliceStrings = append(sliceStrings, itemString)
		}
	}

	return sliceStrings
}

func (s *UserState) AddValueToStringSlice(key FormField, value string) {
	if s.Form == nil {
		s.Form = map[string]any{}
	}

	currentValue := s.GetStringSlice(key)
	currentValue = append(currentValue, value)
	s.Form[string(key)] = currentValue
}

func (s *UserState) RemoveValueFromStringSlice(key FormField, value string) {
	if s.Form == nil {
		s.Form = map[string]any{}
	}

	currentValue := s.GetStringSlice(key)
	index := lo.IndexOf(currentValue, value)
	if index >= 0 {
		currentValue = append(currentValue[0:index], currentValue[index+1:]...)
		s.Form[string(key)] = currentValue
	}
}

func (s *UserState) GetStringMap(key FormField) map[string]string {
	if s.Form == nil {
		return map[string]string{}
	}

	value, exists := s.Form[string(key)]
	if !exists {
		return map[string]string{}
	}

	mapValue := value.(map[string]any)

	result := map[string]string{}
	for k, v := range mapValue {
		result[k] = v.(string)
	}
	return result
}

func (s *UserState) PutValueToMap(key FormField, valueKey string, value string) {
	currentValue := s.GetStringMap(key)
	currentValue[valueKey] = value
	s.Form[string(key)] = currentValue
}

type AccountState string

const (
	AccountStateEnabled  AccountState = "enabled"
	AccountStateDisabled AccountState = "disabled"
)

type Account struct {
	Login                string       `firestore:"login"`
	Password             string       `firestore:"password"`
	Token                string       `firestore:"token"`
	RateLimitedUntil     time.Time    `firestore:"rateLimitedUntil"`
	RateLimitNextDayTime time.Time    `firestore:"rateLimitNextDayTime"`
	State                AccountState `firestore:"state"`
}

func (a *Account) GetStateName() (string, error) {
	switch a.State {
	case AccountStateDisabled:
		return "Отключён", nil
	case AccountStateEnabled:
		return "Включён", nil
	default:
		return "", errorx.IllegalState.New("unsupported account state: %v", a.State)
	}

}
