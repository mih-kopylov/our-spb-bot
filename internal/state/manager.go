package state

import (
	"github.com/joomcode/errorx"
)

type Manager[S BaseUserState] interface {
	// GetState Reads a user from the storage
	GetState(userId int64) (*S, error)
	// SetState Puts a new user into the storage. If the user already exists in the context, it's kept,
	SetState(state *S) error
	// GetAllStates Reads all users from the storage
	GetAllStates() ([]*S, error)
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
