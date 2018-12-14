package models

import (
	"errors"
)

// Typed errors
var (
	ErrSessionNotFound     = errors.New("User session not found")
	ErrSessionExpired      = errors.New("User session expired")
	ErrSessionTokenExpired = errors.New("User session token expired")
)

type UserSession struct {
	Id            int64
	UserId        int64
	AuthToken     string
	PrevAuthToken string
	UserAgent     string
	ClientIp      string
	AuthTokenSeen bool
	SeenAt        int64
	RotatedAt     int64
	CreatedAt     int64
	UpdatedAt     int64
}

// ---------------------
// COMMANDS

type CreateUserSessionCommand struct {
	UserID    int64
	ClientIP  string
	UserAgent string

	Result *UserSession
}

type RefreshUserSessionCommand struct {
	AuthToken string
	ClientIP  string
	UserAgent string

	Result    *UserSession
	Refreshed bool
}

// ---------------------
// QUERIES

type LookupUserSessionByTokenQuery struct {
	Token  string
	Result *UserSession
}
