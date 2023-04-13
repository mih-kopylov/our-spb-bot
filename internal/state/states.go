package state

import (
	"github.com/joomcode/errorx"
	"github.com/sirupsen/logrus"
	"time"
)

type States interface {
	// GetState Puts a new user into the context. If the user already exists in the context, it's kept,
	GetState(userId int64) (*UserState, error)
	SetState(state *UserState) error
}

var (
	Errors         = errorx.NewNamespace("States")
	ErrRateLimited = Errors.NewType("RateLimited")
)

const (
	MessagePriorityNormal int = iota
	MessagePriorityHigh
)

type FormField string

const (
	FormFieldLogin               FormField = "login"
	FormFieldCurrentCategoryNode FormField = "currentCategoryNode"
	FormFieldMessageText         FormField = "messageText"
	FormFieldMessagePriority     FormField = "messagePriority"
	FormFieldFiles               FormField = "files"
)

type UserState struct {
	UserId             int64          `firestore:"userId"`
	UserName           string         `firestore:"userName"`
	FullName           string         `firestore:"fullName"`
	Accounts           []Account      `firestore:"accounts"`
	MessageHandlerName string         `firestore:"messageHandlerName"`
	SentMessagesCount  int            `firestore:"sentMessagesCount"`
	Form               map[string]any `firestore:"form"`
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

	valueArray, ok := value.([]any)
	if !ok {
		logrus.WithField("userId", s.UserId).
			WithField("key", key).
			Warn("failed to convert value to array")
		return nil
	}

	var sliceStrings []string
	for _, item := range valueArray {
		itemString, ok := item.(string)
		if !ok {
			logrus.WithField("userId", s.UserId).
				WithField("key", key).
				WithField("item", item).
				Warn("failed to convert array value item to string")
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

	storedValue, exists := s.Form[string(key)]
	if !exists {
		storedValue = []string{}
		s.Form[string(key)] = storedValue
	}

	slice, ok := storedValue.([]string)
	if !ok {
		slice = []string{}
	}

	slice = append(slice, value)

	s.Form[string(key)] = slice
}

type AccountState string

const (
	AccountStateEnabled  AccountState = "enabled"
	AccountStateDisabled AccountState = "disabled"
)

type Account struct {
	Login            string       `firestore:"login"`
	Password         string       `firestore:"password"`
	Token            string       `firestore:"token"`
	RateLimitedUntil time.Time    `firestore:"rateLimitedUntil"`
	State            AccountState `firestore:"state"`
}
