package middleware

import (
	"strconv"
	"time"

	"gopkg.in/macaron.v1"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/components/apikeygen"
	"github.com/grafana/grafana/pkg/log"
	"github.com/grafana/grafana/pkg/login"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/session"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/util"
)

var (
	ReqGrafanaAdmin = Auth(&AuthOptions{ReqSignedIn: true, ReqGrafanaAdmin: true})
	ReqSignedIn     = Auth(&AuthOptions{ReqSignedIn: true})
	ReqEditorRole   = RoleAuth(m.ROLE_EDITOR, m.ROLE_ADMIN)
	ReqOrgAdmin     = RoleAuth(m.ROLE_ADMIN)
)

func GetContextHandler() macaron.Handler {
	return func(c *macaron.Context) {
		ctx := &m.ReqContext{
			Context:        c,
			SignedInUser:   &m.SignedInUser{},
			Session:        session.GetSession(),
			IsSignedIn:     false,
			AllowAnonymous: false,
			SkipCache:      false,
			Logger:         log.New("context"),
		}

		orgId := int64(0)
		orgIdHeader := ctx.Req.Header.Get("X-Grafana-Org-Id")
		if orgIdHeader != "" {
			orgId, _ = strconv.ParseInt(orgIdHeader, 10, 64)
		}

		// the order in which these are tested are important
		// look for api key in Authorization header first
		// then init session and look for userId in session
		// then look for api key in session (special case for render calls via api)
		// then test if anonymous access is enabled
		switch {
		case initContextWithRenderAuth(ctx):
		case initContextWithApiKey(ctx):
		case initContextWithBasicAuth(ctx, orgId):
		case initContextWithAuthProxy(ctx, orgId):
		case initContextWithTokenCookie(ctx, orgId):
		// case initContextWithUserSessionCookie(ctx, orgId):
		case initContextWithAnonymousUser(ctx):
		}

		ctx.Logger = log.New("context", "userId", ctx.UserId, "orgId", ctx.OrgId, "uname", ctx.Login)
		ctx.Data["ctx"] = ctx

		c.Map(ctx)

		// update last seen every 5min
		if ctx.ShouldUpdateLastSeenAt() {
			ctx.Logger.Debug("Updating last user_seen_at", "user_id", ctx.UserId)
			if err := bus.Dispatch(&m.UpdateUserLastSeenAtCommand{UserId: ctx.UserId}); err != nil {
				ctx.Logger.Error("Failed to update last_seen_at", "error", err)
			}
		}
	}
}

func initContextWithAnonymousUser(ctx *m.ReqContext) bool {
	if !setting.AnonymousEnabled {
		return false
	}

	orgQuery := m.GetOrgByNameQuery{Name: setting.AnonymousOrgName}
	if err := bus.Dispatch(&orgQuery); err != nil {
		log.Error(3, "Anonymous access organization error: '%s': %s", setting.AnonymousOrgName, err)
		return false
	}

	ctx.IsSignedIn = false
	ctx.AllowAnonymous = true
	ctx.SignedInUser = &m.SignedInUser{IsAnonymous: true}
	ctx.OrgRole = m.RoleType(setting.AnonymousOrgRole)
	ctx.OrgId = orgQuery.Result.Id
	ctx.OrgName = orgQuery.Result.Name
	return true
}

func initContextWithTokenCookie(ctx *m.ReqContext, orgId int64) bool {
	sessionCookieName := "session"
	sessionToken := ctx.GetCookie(sessionCookieName)
	if sessionToken == "" {
		return false
	}

	tokenAuthenticator, err := login.NewTokenAuthenticator()
	if err != nil {
		ctx.Logger.Error("Failed to create topen authenticator", "error", err)
		return false
	}

	sessionID, userID, err := tokenAuthenticator.Validate(sessionToken)
	if err != nil && err != m.ErrSessionTokenExpired {
		ctx.Logger.Error("Failed to validate session token", "error", err)
		ctx.Resp.Header().Del("Set-Cookie")
		ctx.SetCookie(sessionCookieName, "", -1, setting.AppSubUrl+"/", setting.Domain, false, true)

		return false
	}

	if err != nil && err == m.ErrSessionTokenExpired {
		cmd := &m.RefreshUserSessionCommand{
			SessionID: sessionID,
			UserID:    userID,
			ClientIP:  ctx.Req.RemoteAddr,
			UserAgent: ctx.Req.UserAgent(),
		}

		if err := bus.Dispatch(cmd); err != nil {
			if err == m.ErrSessionNotFound {
				ctx.Logger.Error("Session not found")
				ctx.Resp.Header().Del("Set-Cookie")
				ctx.SetCookie(sessionCookieName, "", -1, setting.AppSubUrl+"/", setting.Domain, false, true)
				return false
			}

			ctx.Logger.Error("Error while trying to refresh user session", "error", err)
			return false
		}

		serializedToken, err := tokenAuthenticator.RefreshToken(cmd.SessionID, cmd.UserID, time.Unix(cmd.Result.CreatedAt, 0).UTC())
		if err != nil {
			ctx.Logger.Error("Error while trying to refresh session token", "error", err)
			return false
		}

		ctx.Resp.Header().Del("Set-Cookie")
		ctx.SetCookie(sessionCookieName, serializedToken, 86400, setting.AppSubUrl+"/", setting.Domain, false, true, time.Now().UTC().Add(24*time.Hour))
	}

	query := m.GetSignedInUserQuery{UserId: userID, OrgId: orgId}
	if err := bus.Dispatch(&query); err != nil {
		ctx.Logger.Error("Failed to get user with id", "userId", userID, "error", err)
		return false
	}

	ctx.SignedInUser = query.Result
	ctx.IsSignedIn = true
	return true
}

func initContextWithUserSessionCookie(ctx *m.ReqContext, orgId int64) bool {
	// initialize session
	if err := ctx.Session.Start(ctx.Context); err != nil {
		ctx.Logger.Error("Failed to start session", "error", err)
		return false
	}

	var userId int64
	if userId = getRequestUserId(ctx); userId == 0 {
		return false
	}

	query := m.GetSignedInUserQuery{UserId: userId, OrgId: orgId}
	if err := bus.Dispatch(&query); err != nil {
		ctx.Logger.Error("Failed to get user with id", "userId", userId, "error", err)
		return false
	}

	ctx.SignedInUser = query.Result
	ctx.IsSignedIn = true
	return true
}

func initContextWithApiKey(ctx *m.ReqContext) bool {
	var keyString string
	if keyString = getApiKey(ctx); keyString == "" {
		return false
	}

	// base64 decode key
	decoded, err := apikeygen.Decode(keyString)
	if err != nil {
		ctx.JsonApiErr(401, "Invalid API key", err)
		return true
	}

	// fetch key
	keyQuery := m.GetApiKeyByNameQuery{KeyName: decoded.Name, OrgId: decoded.OrgId}
	if err := bus.Dispatch(&keyQuery); err != nil {
		ctx.JsonApiErr(401, "Invalid API key", err)
		return true
	}

	apikey := keyQuery.Result

	// validate api key
	if !apikeygen.IsValid(decoded, apikey.Key) {
		ctx.JsonApiErr(401, "Invalid API key", err)
		return true
	}

	ctx.IsSignedIn = true
	ctx.SignedInUser = &m.SignedInUser{}
	ctx.OrgRole = apikey.Role
	ctx.ApiKeyId = apikey.Id
	ctx.OrgId = apikey.OrgId
	return true
}

func initContextWithBasicAuth(ctx *m.ReqContext, orgId int64) bool {

	if !setting.BasicAuthEnabled {
		return false
	}

	header := ctx.Req.Header.Get("Authorization")
	if header == "" {
		return false
	}

	username, password, err := util.DecodeBasicAuthHeader(header)
	if err != nil {
		ctx.JsonApiErr(401, "Invalid Basic Auth Header", err)
		return true
	}

	loginQuery := m.GetUserByLoginQuery{LoginOrEmail: username}
	if err := bus.Dispatch(&loginQuery); err != nil {
		ctx.JsonApiErr(401, "Basic auth failed", err)
		return true
	}

	user := loginQuery.Result

	loginUserQuery := m.LoginUserQuery{Username: username, Password: password, User: user}
	if err := bus.Dispatch(&loginUserQuery); err != nil {
		ctx.JsonApiErr(401, "Invalid username or password", err)
		return true
	}

	query := m.GetSignedInUserQuery{UserId: user.Id, OrgId: orgId}
	if err := bus.Dispatch(&query); err != nil {
		ctx.JsonApiErr(401, "Authentication error", err)
		return true
	}

	ctx.SignedInUser = query.Result
	ctx.IsSignedIn = true
	return true
}

func AddDefaultResponseHeaders() macaron.Handler {
	return func(ctx *m.ReqContext) {
		if ctx.IsApiRequest() && ctx.Req.Method == "GET" {
			ctx.Resp.Header().Add("Cache-Control", "no-cache")
			ctx.Resp.Header().Add("Pragma", "no-cache")
			ctx.Resp.Header().Add("Expires", "-1")
		}
	}
}
