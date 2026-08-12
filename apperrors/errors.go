package apperrors

import (
	"errors"
)

var (
	ErrEmptyTitle         = errors.New("Title cannot be empty")
	ErrEmptyEmail         = errors.New("Email cannot be empty")
	ErrEmptyPassword      = errors.New("Password cannot be empty")
	PasswordTooShort      = errors.New("Password must be at least 6 characters")
	ErrInvalidCredentials = errors.New("Invalid Credentials")
	ExistEmailError       = errors.New("email already exists")
	ErrTaskNotFound       = errors.New("Task not found")
)
