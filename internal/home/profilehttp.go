package home

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"

	"github.com/AdguardTeam/AdGuardHome/internal/aghhttp"
	"github.com/AdguardTeam/AdGuardHome/internal/aghuser"
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
	if !allowedLanguages.Has(lang) {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "unknown language: %q", lang)

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
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusForbidden, "no users configured")

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
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusForbidden, "current password is incorrect")

		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
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
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusForbidden, "no users configured")

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

	if req.NewUsername == "" {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "username must not be empty")

		return
	}

	oldLogin := u.Login
	u.Login = aghuser.Login(req.NewUsername)

	func() {
		config.Lock()
		defer config.Unlock()

		for i, wu := range config.Users {
			if aghuser.Login(wu.Name) == oldLogin {
				config.Users[i].Name = req.NewUsername

				break
			}
		}
	}()

	web.confModifier.Apply(ctx)

	l.InfoContext(ctx, "username changed", "old", oldLogin, "new", req.NewUsername)

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

	func() {
		config.Lock()
		defer config.Unlock()

		addr := config.HTTPConfig.Address.Addr()
		config.HTTPConfig.Address = netip.AddrPortFrom(addr, req.Port)
	}()

	web.confModifier.Apply(ctx)

	l.InfoContext(ctx, "http port changed", "port", req.Port)

	aghhttp.OK(ctx, l, w)
}
