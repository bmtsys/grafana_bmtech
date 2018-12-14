package login

import (
	"fmt"
	"strconv"
	"time"

	"github.com/grafana/grafana/pkg/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
	jose "gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

const sessionCookieName = "session"

var now = time.Now

type TokenAuthenticator interface {
	CreateToken(authToken string, userID int64) (string, error)
	RefreshToken(authToken string, userID int64, issuedAt time.Time) (string, error)
	Validate(serializedToken string) (authToken string, userID int64, err error)
}

type tokenAuthenticator struct {
	jwtSigner       jose.Signer
	sessionLifetime time.Duration
	tokenLifeTime   time.Duration
	logger          log.Logger
}

func NewTokenAuthenticator() (TokenAuthenticator, error) {
	sig, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.HS256, Key: []byte(setting.SecretKey)}, (&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		return nil, err
	}

	return &tokenAuthenticator{
		jwtSigner:       sig,
		sessionLifetime: 10 * time.Minute,
		tokenLifeTime:   1 * time.Minute,
		logger:          log.New("tokenauth"),
	}, nil
}

func (ta *tokenAuthenticator) CreateToken(authToken string, userID int64) (string, error) {
	return ta.issueAndSignToken(authToken, userID, now())
}

func (ta *tokenAuthenticator) RefreshToken(authToken string, userID int64, issuedAt time.Time) (string, error) {
	return ta.issueAndSignToken(authToken, userID, issuedAt)
}

func (ta *tokenAuthenticator) Validate(serializedToken string) (authToken string, userID int64, err error) {
	if len(serializedToken) == 0 {
		return "", 0, fmt.Errorf("Invalid token")
	}

	tok, err := jwt.ParseSigned(serializedToken)
	if err != nil {
		return "", 0, err
	}

	claims := jwt.Claims{}
	if err := tok.Claims([]byte(setting.SecretKey), &claims); err != nil {
		return "", 0, err
	}

	err = claims.Validate(jwt.Expected{
		Issuer:   setting.AppUrl,
		Audience: jwt.Audience{setting.AppUrl},
		Time:     time.Now().UTC(),
	})

	if err != nil && err != jwt.ErrExpired {
		return "", 0, err
	}

	jwtExpired := err == jwt.ErrExpired
	authToken = claims.ID
	userID, err = strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("Invalid user id in subject claim, error=%v", err)
	}
	if userID <= 0 {
		return "", 0, fmt.Errorf("Invalid user id in subject claim")
	}

	if jwtExpired {
		ta.logger.Debug("JWT expired", "jti", claims.ID, "sub", claims.Subject, "iss", claims.IssuedAt.Time(), "exp", claims.Expiry.Time())

		sessionExpirationTime := claims.IssuedAt.Time().UTC().Add(ta.sessionLifetime)
		if now().After(sessionExpirationTime) {
			ta.logger.Debug("Session expired", "jti", claims.ID, "sub", claims.Subject, "sessionExp", sessionExpirationTime)
			return "", 0, models.ErrSessionExpired
		}

		return authToken, userID, models.ErrSessionTokenExpired
	}

	ta.logger.Debug("Valid JWT", "jti", claims.ID, "sub", claims.Subject, "iss", claims.IssuedAt.Time(), "exp", claims.Expiry.Time())

	return authToken, userID, nil
}

func (ta *tokenAuthenticator) issueAndSignToken(authToken string, userID int64, issuedAt time.Time) (string, error) {
	if len(authToken) == 0 {
		return "", fmt.Errorf("Invalid session id")
	}

	if userID <= 0 {
		return "", fmt.Errorf("Invalid user id")
	}

	claims := jwt.Claims{
		ID:        authToken,
		Issuer:    setting.AppUrl,
		Subject:   strconv.FormatInt(userID, 10),
		Audience:  jwt.Audience{setting.AppUrl},
		IssuedAt:  jwt.NewNumericDate(issuedAt),
		Expiry:    jwt.NewNumericDate(now().Add(ta.tokenLifeTime)),
		NotBefore: jwt.NewNumericDate(now()),
	}

	ta.logger.Debug("Issuing JWT", "jti", claims.ID, "sub", claims.Subject, "iss", claims.IssuedAt.Time(), "exp", claims.Expiry.Time())

	serializedToken, err := jwt.Signed(ta.jwtSigner).Claims(claims).CompactSerialize()
	if err != nil {
		return "", err
	}

	return serializedToken, nil
}
