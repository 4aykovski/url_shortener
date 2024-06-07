package services

import "errors"

var (
	ErrAliasAlreadyExists = errors.New("alias already exists")
	ErrURLNotFound        = errors.New("url not found")
	ErrUserHasNoUrls      = errors.New("user has no urls")
)
