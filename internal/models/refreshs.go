package models

import "time"

type RefreshSession struct {
	Id           int
	UserId       int
	RefreshToken string
	ExpiresIn    time.Time
}
