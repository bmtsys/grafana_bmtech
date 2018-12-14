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
	SessionId   string
	UserId      int64
	UserAgent   string
	ClientIp    string
	RefreshedAt int64
	CreatedAt   int64
	UpdatedAt   int64
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
	SessionID string
	UserID    int64
	ClientIP  string
	UserAgent string

	Result *UserSession
}

// ---------------------
// QUERIES

type GetUserSessionQuery struct {
	SessionID string
	UserID    int64
	Result    *UserSession
}
