package repository

import "errors"

var (
	ErrURLNotFound             = errors.New("url not found")
	ErrUrlExists               = errors.New("url exists")
	ErrUserExists              = errors.New("user exists")
	ErrUserNotFound            = errors.New("user not found")
	ErrUsersNotFound           = errors.New("user not found")
	ErrRefreshSessionNotFound  = errors.New("refresh session not found")
	ErrRefreshSessionsNotFound = errors.New("refresh sessions not found")
)
