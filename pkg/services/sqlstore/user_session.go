package sqlstore

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/grafana/grafana/pkg/bus"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/util"
)

func init() {
	bus.AddHandler("sql", CreateUserSession)
	bus.AddHandler("sql", GetUserSession)
	bus.AddHandler("sql", RefreshUserSession)
}

var now = time.Now

func CreateUserSession(cmd *m.CreateUserSessionCommand) error {
	return inTransaction(func(sess *DBSession) error {
		parts := strings.Split(cmd.ClientIP, ":")
		clientIP := parts[0]

		key, err := util.RandomHex(16)
		if err != nil {
			return err
		}

		hashBytes := sha256.Sum256([]byte(key + setting.SecretKey))
		sessionID := hex.EncodeToString(hashBytes[:])

		userSession := &m.UserSession{
			SessionId:   sessionID,
			UserId:      cmd.UserID,
			ClientIp:    clientIP,
			UserAgent:   cmd.UserAgent,
			RefreshedAt: now().Unix(),
			CreatedAt:   now().Unix(),
			UpdatedAt:   now().Unix(),
		}
		_, err = sess.Insert(userSession)
		if err != nil {
			return err
		}

		cmd.Result = userSession

		return nil
	})
}

func GetUserSession(query *m.GetUserSessionQuery) error {
	var userSession m.UserSession
	exists, err := x.Where("session_id = ? AND user_id = ?", query.SessionID, query.UserID).Get(&userSession)
	if err != nil {
		return err
	}

	if !exists {
		return m.ErrSessionNotFound
	}

	query.Result = &userSession
	return nil
}

func RefreshUserSession(cmd *m.RefreshUserSessionCommand) error {
	return inTransaction(func(sess *DBSession) error {
		var userSession m.UserSession
		exists, err := x.Where("session_id = ? AND user_id = ?", cmd.SessionID, cmd.UserID).Get(&userSession)
		if err != nil {
			return err
		}

		if !exists {
			return m.ErrSessionNotFound
		}

		// Doesn't work
		// if userSession.RefreshedAt != cmd.LastRefreshedAt {
		// 	if time.Unix(cmd.LastRefreshedAt, 0).Sub(time.Unix(userSession.RefreshedAt, 0)) <= 30*time.Second {
		// 		cmd.Result = &userSession
		// 		return nil
		// 	}

		// 	return fmt.Errorf("Too old token. Failed to refresh user session")
		// }

		userSession.RefreshedAt = now().Unix()
		userSession.UpdatedAt = now().Unix()

		rowsAffected, err := sess.Where("session_id = ? AND user_id = ?", cmd.SessionID, cmd.UserID).Update(&userSession)
		if err != nil {
			return err
		}

		if rowsAffected == 0 {
			return fmt.Errorf("Failed to refresh user session")
		}

		cmd.Result = &userSession

		return nil
	})
}
