package state

import (
	"github.com/joomcode/errorx"
)

type Manager interface {
	// GetState Reads a user from the storage
	GetState(userId int64) (*UserState, error)
	// SetState Puts a new user into the storage. If the user already exists in the context, it's kept,
	SetState(state *UserState) error
	// GetAllStates Reads all users from the storage
	GetAllStates() ([]*UserState, error)
}

var (
	Errors         = errorx.NewNamespace("State")
	ErrRateLimited = Errors.NewType("RateLimited")
)

const (
	FormFieldLogin               FormField = "login"
	FormFieldCurrentCategoryNode FormField = "currentCategoryNode"
	FormFieldMessageText         FormField = "messageText"
	FormFieldFiles               FormField = "files"
	FormFieldMessageIdFile       FormField = "messageIdFile"
)
