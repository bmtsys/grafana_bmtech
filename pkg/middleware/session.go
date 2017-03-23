package middleware

import (
	"net/http"
	"time"

	"github.com/go-macaron/session"
	_ "github.com/go-macaron/session/memcache"
	_ "github.com/go-macaron/session/mysql"
	_ "github.com/go-macaron/session/postgres"
	_ "github.com/go-macaron/session/redis"
	"github.com/grafana/grafana/pkg/log"
	"github.com/grafana/grafana/pkg/setting"
	"gopkg.in/macaron.v1"
)

const (
	SESS_KEY_USERID       = "uid"
	SESS_KEY_OAUTH_STATE  = "state"
	SESS_KEY_APIKEY       = "apikey_id" // used for render requests with api keys
	SESS_KEY_LASTLDAPSYNC = "last_ldap_sync"
)

var sessionManager *session.Manager
var sessionOptions *session.Options
var startSessionGC func()
var getSessionCount func() int
var sessionLogger = log.New("session")

func Sessioner(options *session.Options) macaron.Handler {
	return func(ctx *Context) {
		ctx.Next()

		if err := ctx.Session.Release(); err != nil {
			panic("session(release): " + err.Error())
		}
	}
}

func GetSession() SessionStore {
	return &CookieSessionStore{
		cookieName:   setting.SessionOptions.CookieName,
		cookieSecure: setting.SessionOptions.Secure,
		cookieMaxAge: setting.SessionOptions.CookieLifeTime,
		cookieDomain: setting.SessionOptions.Domain,
		cookiePath:   setting.SessionOptions.CookiePath,
	}
}

type SessionStore interface {
	// Set sets value to given key in session.
	Set(string, string) error
	// Get gets value by given key in session.
	Get(string) string
	// ID returns current session ID.
	ID() string
	// Release releases session resource and save data to provider.
	Release() error
	// Destory deletes a session.
	Destory(*Context) error
	// init
	Start(*Context) error
}

type CookieSessionStore struct {
	cookieName   string
	cookieSecure bool
	cookiePath   string
	cookieDomain string
	cookieMaxAge int
	id           string
	data         map[string]string
}

func (s *CookieSessionStore) Start(ctx *Context) error {
	cookieString := ctx.GetCookie(s.cookieName)
	if len(cookieString) > 0 {
		sessionLogger.Info("Found session cookie", "cookie", cookieString)
		return nil
	}

	s.id = "my id"

	cookie := &http.Cookie{
		Name:     s.cookieName,
		Value:    "session cookie",
		Path:     s.cookiePath,
		HttpOnly: true,
		Secure:   s.cookieSecure,
		Domain:   s.cookieDomain,
	}

	if s.cookieMaxAge >= 0 {
		cookie.MaxAge = s.cookieMaxAge
	}

	http.SetCookie(ctx.Resp, cookie)
	ctx.Req.AddCookie(cookie)
	return nil
}

func (s *CookieSessionStore) Set(k string, v string) error {
	sessionLogger.Info("Set", "key", k, "value", v)
	s.data[k] = v
	return nil
}

func (s *CookieSessionStore) Get(k string) string {
	sessionLogger.Info("Get", "key", k)
	value, _ := s.data[k]
	return value
}

func (s *CookieSessionStore) ID() string {
	return s.id
}

func (s *CookieSessionStore) Release() error {
	return nil
}

func (s *CookieSessionStore) Destory(c *Context) error {
	cookieString := c.GetCookie(s.cookieName)
	if len(cookieString) == 0 {
		return nil
	}

	cookie := &http.Cookie{
		Name:     s.cookieName,
		Path:     s.cookiePath,
		HttpOnly: true,
		Expires:  time.Now(),
		MaxAge:   -1,
	}

	http.SetCookie(c.Resp, cookie)
	return nil
}
