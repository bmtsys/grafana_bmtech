package sqlstore

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/grafana/grafana/pkg/bus"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/util"
)

func init() {
	bus.AddHandler("sql", CreateUserSession)
	bus.AddHandler("sql", LookupUserSessionByToken)
	bus.AddHandler("sql", RefreshUserSession)
}

var now = time.Now

func CreateUserSession(cmd *m.CreateUserSessionCommand) error {
	return inTransaction(func(sess *DBSession) error {
		clientIP := parseIPAddress(cmd.ClientIP)
		token, err := util.RandomHex(16)
		if err != nil {
			return err
		}

		hashBytes := sha256.Sum256([]byte(token + setting.SecretKey))
		hashedToken := hex.EncodeToString(hashBytes[:])

		userSession := &m.UserSession{
			UserId:        cmd.UserID,
			AuthToken:     hashedToken,
			PrevAuthToken: hashedToken,
			ClientIp:      clientIP,
			UserAgent:     cmd.UserAgent,
			RotatedAt:     now().Unix(),
			CreatedAt:     now().Unix(),
			UpdatedAt:     now().Unix(),
			SeenAt:        0,
		}
		_, err = sess.Insert(userSession)
		if err != nil {
			return err
		}

		cmd.Result = userSession

		return nil
	})
}

func LookupUserSessionByToken(query *m.LookupUserSessionByTokenQuery) error {
	// var userSession m.UserSession
	// exists, err := x.Where("session_id = ? AND user_id = ?", query.SessionID, query.UserID).Get(&userSession)
	// if err != nil {
	// 	return err
	// }

	// if !exists {
	// 	return m.ErrSessionNotFound
	// }

	// query.Result = &userSession
	return nil
}

func RefreshUserSession(cmd *m.RefreshUserSessionCommand) error {
	return inTransaction(func(sess *DBSession) error {
		var userSession m.UserSession
		exists, err := sess.Where("auth_token = ? OR prev_auth_token = ?", cmd.AuthToken, cmd.AuthToken).Get(&userSession)
		if err != nil {
			return err
		}

		if !exists {
			fmt.Println("miss token", "authToken", cmd.AuthToken, "clientIP", cmd.ClientIP, "userAgent", cmd.UserAgent)
			return m.ErrSessionNotFound
		}

		if userSession.AuthToken != cmd.AuthToken && userSession.PrevAuthToken == cmd.AuthToken && userSession.AuthTokenSeen {
			userSession.AuthTokenSeen = false
			expireBefore := now().Add(-2 * time.Minute).Unix()
			affectedRows, err := sess.Where("id = ? AND prev_auth_token = ? AND rotated_at < ?", userSession.Id, userSession.PrevAuthToken, expireBefore).AllCols().Update(&userSession)
			if err != nil {
				return err
			}

			if affectedRows == 0 {
				fmt.Println("prev seen token unchanged", "userSessionId", userSession.Id, "userId", userSession.UserId, "authToken", userSession.AuthToken, "clientIP", userSession.ClientIp, "userAgent", userSession.UserAgent)
			} else {
				fmt.Println("prev seen token", "userSessionId", userSession.Id, "userId", userSession.UserId, "authToken", userSession.AuthToken, "clientIP", userSession.ClientIp, "userAgent", userSession.UserAgent)
			}
		}

		if !userSession.AuthTokenSeen && userSession.AuthToken == cmd.AuthToken {
			userSessionCopy := userSession
			userSessionCopy.AuthTokenSeen = true
			userSessionCopy.SeenAt = now().Unix()
			affectedRows, err := sess.Where("id = ? AND auth_token = ?", userSessionCopy.Id, userSessionCopy.AuthToken).AllCols().Update(&userSessionCopy)
			if err != nil {
				return err
			}

			if affectedRows == 1 {
				userSession = userSessionCopy
			}

			if affectedRows == 0 {
				fmt.Println("seen wrong token", "userSessionId", userSession.Id, "userId", userSession.UserId, "authToken", userSession.AuthToken, "clientIP", userSession.ClientIp, "userAgent", userSession.UserAgent)
			} else {
				fmt.Println("seen token", "userSessionId", userSession.Id, "userId", userSession.UserId, "authToken", userSession.AuthToken, "clientIP", userSession.ClientIp, "userAgent", userSession.UserAgent)
			}
		}

		clientIP := parseIPAddress(cmd.ClientIP)
		token, err := util.RandomHex(16)
		if err != nil {
			return err
		}

		hashBytes := sha256.Sum256([]byte(token + setting.SecretKey))
		hashedToken := hex.EncodeToString(hashBytes[:])

		userSessionCopy := userSession
		if userSessionCopy.AuthTokenSeen {
			userSessionCopy.PrevAuthToken = userSessionCopy.AuthToken
		}
		userSessionCopy.AuthTokenSeen = false
		userSessionCopy.SeenAt = 0
		userSessionCopy.ClientIp = clientIP
		userSessionCopy.UserAgent = cmd.UserAgent
		userSessionCopy.AuthToken = hashedToken
		userSessionCopy.RotatedAt = now().Unix()
		userSessionCopy.UpdatedAt = now().Unix()

		safeguardTime := now().Add(-90 * time.Second).Unix()
		affectedRows, err := sess.Where("id = ? AND (auth_token_seen = ? OR rotated_at < ?)", userSessionCopy.Id, dialect.BooleanStr(true), safeguardTime).AllCols().Update(&userSessionCopy)
		if err != nil {
			return err
		}

		if affectedRows == 1 {
			userSession = userSessionCopy
			cmd.Refreshed = true
		}

		cmd.Result = &userSession
		return nil
	})
}

func parseIPAddress(input string) string {
	var s string
	lastIndex := strings.LastIndex(input, ":")

	if lastIndex != -1 {
		s = input[:lastIndex]
	}

	s = strings.Replace(s, "[", "", -1)
	s = strings.Replace(s, "]", "", -1)

	ip := net.ParseIP(s)

	if ip.IsLoopback() {
		return "127.0.0.1"
	}

	return ip.String()
}
