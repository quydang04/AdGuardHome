package home

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/aghhttp"
	"github.com/AdguardTeam/AdGuardHome/internal/aghuser"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"golang.org/x/crypto/bcrypt"
)

// Theme is an enum of all allowed UI themes.
type Theme string

// Allowed [Theme] values.
//
// Keep in sync with client/src/helpers/constants.ts.
const (
	ThemeAuto  Theme = "auto"
	ThemeLight Theme = "light"
	ThemeDark  Theme = "dark"
)

// UnmarshalText implements [encoding.TextUnmarshaler] interface for *Theme.
func (t *Theme) UnmarshalText(b []byte) (err error) {
	switch string(b) {
	case "auto":
		*t = ThemeAuto
	case "dark":
		*t = ThemeDark
	case "light":
		*t = ThemeLight
	default:
		return fmt.Errorf("invalid theme %q, supported: %q, %q, %q", b, ThemeAuto, ThemeDark, ThemeLight)
	}

	return nil
}

// profileJSON is an object for /control/profile and /control/profile/update
// endpoints.
type profileJSON struct {
	Name     string `json:"name"`
	Language string `json:"language"`
	Theme    Theme  `json:"theme"`
}

// changePasswordJSON is the JSON structure for the password change request.
type changePasswordJSON struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// changeUsernameJSON is the JSON structure for the username change request.
type changeUsernameJSON struct {
	NewUsername string `json:"new_username"`
}

// changePortJSON is the JSON structure for the port change request.
type changePortJSON struct {
	Port uint16 `json:"port"`
}

// handleGetProfile is the handler for GET /control/profile endpoint.
func (web *webAPI) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var name string

	if !web.auth.isGLiNet && !web.auth.isUserless {
		u, ok := webUserFromContext(ctx)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		name = string(u.Login)
	}

	var resp profileJSON
	func() {
		config.RLock()
		defer config.RUnlock()

		resp = profileJSON{
			Name:     name,
			Language: config.Language,
			Theme:    config.Theme,
		}
	}()

	aghhttp.WriteJSONResponseOK(ctx, web.logger, w, r, resp)
}

// handlePutProfile is the handler for PUT /control/profile/update endpoint.
func (web *webAPI) handlePutProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := web.logger

	if aghhttp.WriteTextPlainDeprecated(ctx, l, w, r) {
		return
	}

	profileReq := &profileJSON{}
	err := json.NewDecoder(r.Body).Decode(profileReq)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "reading req: %s", err)

		return
	}

	lang := profileReq.Language
	err = validateLang(lang, false)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "%s", err)

		return
	}

	theme := profileReq.Theme

	changed := false
	func() {
		config.Lock()
		defer config.Unlock()

		if config.Language == lang && config.Theme == theme {
			l.DebugContext(ctx, "updating profile; no changes")

			return
		}

		changed = true
		config.Language = lang
		config.Theme = theme
		l.InfoContext(ctx, "profile updated", "lang", lang, "theme", theme)
	}()

	if changed {
		web.confModifier.Apply(ctx)
	}

	aghhttp.OK(ctx, l, w)
}

// handleChangePassword is the handler for POST /control/profile/password
// endpoint.
func (web *webAPI) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := web.logger

	if web.auth.isUserless {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "no users configured")

		return
	}

	u, ok := webUserFromContext(ctx)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	req := &changePasswordJSON{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "reading req: %s", err)

		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "passwords must not be empty")

		return
	}

	if !u.Password.Authenticate(ctx, req.CurrentPassword) {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "current password is incorrect")

		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "password must be at most 72 bytes")

			return
		}

		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusInternalServerError, "generating hash: %s", err)

		return
	}

	u.Password = aghuser.NewDefaultPassword(string(hash))

	func() {
		config.Lock()
		defer config.Unlock()

		for i, wu := range config.Users {
			if aghuser.Login(wu.Name) == u.Login {
				config.Users[i].PasswordHash = string(hash)

				break
			}
		}
	}()

	web.confModifier.Apply(ctx)

	l.InfoContext(ctx, "password changed", "login", u.Login)

	aghhttp.OK(ctx, l, w)
}

// handleChangeUsername is the handler for POST /control/profile/username
// endpoint.
func (web *webAPI) handleChangeUsername(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := web.logger

	if web.auth.isUserless {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "no users configured")

		return
	}

	u, ok := webUserFromContext(ctx)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	req := &changeUsernameJSON{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "reading req: %s", err)

		return
	}

	newUsername := strings.TrimSpace(req.NewUsername)
	if newUsername == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "username must not be empty")

		return
	}

	oldLogin := u.Login
	newLogin := aghuser.Login(newUsername)
	if newLogin == oldLogin {
		aghhttp.OK(ctx, l, w)

		return
	}

	err = web.auth.users.UpdateLogin(ctx, oldLogin, newLogin)
	if err != nil {
		if errors.Is(err, errors.ErrDuplicated) {
			aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusConflict, "username already exists")

			return
		}

		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusInternalServerError, "updating username: %s", err)

		return
	}

	err = web.auth.sessions.UpdateLogin(ctx, oldLogin, newLogin)
	if err != nil {
		_ = web.auth.users.UpdateLogin(ctx, newLogin, oldLogin)
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusInternalServerError, "updating sessions: %s", err)

		return
	}

	func() {
		config.Lock()
		defer config.Unlock()

		for i, wu := range config.Users {
			if aghuser.Login(wu.Name) == oldLogin {
				config.Users[i].Name = newUsername

				break
			}
		}
	}()

	web.confModifier.Apply(ctx)

	l.InfoContext(ctx, "username changed", "old", oldLogin, "new", newUsername)

	aghhttp.OK(ctx, l, w)
}

// handleChangePort is the handler for POST /control/change_port endpoint.
func (web *webAPI) handleChangePort(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := web.logger

	req := &changePortJSON{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "reading req: %s", err)

		return
	}

	if req.Port == 0 {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "port must not be zero")

		return
	}

	var (
		oldAddr     netip.AddrPort
		tlsEnabled  bool
		tlsPortHTTP uint16
	)
	func() {
		config.RLock()
		defer config.RUnlock()

		oldAddr = config.HTTPConfig.Address
		tlsEnabled = config.TLS.Enabled
		tlsPortHTTP = config.TLS.PortHTTPS
	}()

	if oldAddr.Port() == req.Port {
		aghhttp.OK(ctx, l, w)

		return
	}

	if tlsEnabled && tlsPortHTTP == req.Port {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "port conflicts with HTTPS port")

		return
	}

	newAddr := netip.AddrPortFrom(oldAddr.Addr(), req.Port)
	if newAddr.IsValid() {
		var ln net.Listener
		ln, err = net.Listen("tcp", newAddr.String())
		if err != nil {
			aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "port is unavailable: %s", err)

			return
		}
		_ = ln.Close()
	}

	func() {
		config.Lock()
		defer config.Unlock()

		config.HTTPConfig.Address = newAddr
	}()

	if web.conf != nil {
		web.conf.BindAddr = newAddr
	}

	web.confModifier.Apply(ctx)

	l.InfoContext(ctx, "http port changed", "port", req.Port)

	aghhttp.OK(ctx, l, w)

	if web.httpServer == nil {
		return
	}

	rc := http.NewResponseController(w)
	err = rc.Flush()
	if err != nil {
		l.WarnContext(ctx, "flushing response", slogutil.KeyError, err)
	}

	// Shut down the HTTP server asynchronously so the accept loop in
	// [webAPI.start] rebinds to the updated port without blocking this
	// handler.
	go func(timeout time.Duration, srv *http.Server) {
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
		defer slogutil.RecoverAndLog(shutdownCtx, l)
		defer cancel()

		shutdownSrv(shutdownCtx, l, srv)
	}(shutdownTimeout, web.httpServer)
}
