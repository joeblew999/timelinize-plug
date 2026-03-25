package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/settings"
	"github.com/pocketbase/pocketbase/tokens"
)

const pbAdminCookieName = "tplug_pb_admin"

type pbAdminSession struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type pbProvidersPayload struct {
	Google pbAuthProviderPayload `json:"google"`
	Github pbAuthProviderPayload `json:"github"`
}

type pbAuthProviderPayload struct {
	Enabled      bool   `json:"enabled"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type pbAdminStateResponse struct {
	NeedsBootstrap bool             `json:"needsBootstrap"`
	Session        *pbAdminSession  `json:"session,omitempty"`
	Providers      *pbProvidersView `json:"providers,omitempty"`
}

type pbProvidersView struct {
	Google      pbAuthProviderPayload `json:"google"`
	Github      pbAuthProviderPayload `json:"github"`
	RedirectURL string                `json:"redirectUrl"`
}

func registerPocketBaseAuthRoutes(r *chi.Mux, opt Options) {
	app := opt.PBApp
	if app == nil {
		return
	}

	r.Route("/api/pb/admin", func(r chi.Router) {
		r.Get("/state", func(w http.ResponseWriter, req *http.Request) {
			state, status, err := resolvePBAdminState(app, req, opt.OAuthCBAddr)
			if err != nil {
				http.Error(w, err.Error(), status)
				return
			}
			writeJSON(w, state)
		})

		r.Post("/bootstrap", func(w http.ResponseWriter, req *http.Request) {
			if status, err := ensurePBNeedsBootstrap(app); err != nil {
				http.Error(w, err.Error(), status)
				return
			}
			var payload struct {
				Email           string `json:"email"`
				Password        string `json:"password"`
				PasswordConfirm string `json:"passwordConfirm"`
			}
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			session, status, err := bootstrapPBAdmin(app, payload.Email, payload.Password, payload.PasswordConfirm)
			if err != nil {
				http.Error(w, err.Error(), status)
				return
			}
			if err := issuePBAdminCookie(w, app, session); err != nil {
				http.Error(w, "failed to issue session", http.StatusInternalServerError)
				return
			}
			writeJSON(w, map[string]any{
				"session": session,
			})
		})

		r.Post("/login", func(w http.ResponseWriter, req *http.Request) {
			var payload struct {
				Identity string `json:"identity"`
				Password string `json:"password"`
			}
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			session, status, err := authenticatePBAdmin(app, payload.Identity, payload.Password)
			if err != nil {
				http.Error(w, err.Error(), status)
				return
			}
			if err := issuePBAdminCookie(w, app, session); err != nil {
				http.Error(w, "failed to issue session", http.StatusInternalServerError)
				return
			}
			writeJSON(w, map[string]any{
				"session": session,
			})
		})

		r.Post("/logout", func(w http.ResponseWriter, req *http.Request) {
			clearPBAdminCookie(w)
			writeJSON(w, map[string]any{"ok": true})
		})

		r.Get("/providers", func(w http.ResponseWriter, req *http.Request) {
			admin, status, err := ensurePBAdminSession(app, req)
			if err != nil {
				if status == http.StatusUnauthorized {
					clearPBAdminCookie(w)
				}
				http.Error(w, err.Error(), status)
				return
			}
			_ = admin

			view, err := loadPBProviders(app, opt.OAuthCBAddr)
			if err != nil {
				log.Printf("pb auth providers load: %v", err)
				http.Error(w, "failed to load providers", http.StatusInternalServerError)
				return
			}
			writeJSON(w, map[string]any{
				"providers": view,
			})
		})

		r.Post("/providers", func(w http.ResponseWriter, req *http.Request) {
			_, status, err := ensurePBAdminSession(app, req)
			if err != nil {
				if status == http.StatusUnauthorized {
					clearPBAdminCookie(w)
				}
				http.Error(w, err.Error(), status)
				return
			}
			var payload pbProvidersPayload
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if status, err := updatePBProviders(app, payload); err != nil {
				http.Error(w, err.Error(), status)
				return
			}
			view, err := loadPBProviders(app, opt.OAuthCBAddr)
			if err != nil {
				log.Printf("pb auth providers reload: %v", err)
				http.Error(w, "failed to load providers", http.StatusInternalServerError)
				return
			}
			writeJSON(w, map[string]any{
				"providers": view,
			})
		})
	})
}

func resolvePBAdminState(app *pocketbase.PocketBase, req *http.Request, oauthAddr string) (*pbAdminStateResponse, int, error) {
	total, err := app.Dao().TotalAdmins()
	if err != nil {
		log.Printf("pocketbase admin count: %v", err)
		return nil, http.StatusInternalServerError, errors.New("failed to read admin state")
	}
	needsBootstrap := total == 0

	session, err := sessionFromRequest(app, req)
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		log.Printf("pocketbase admin session: %v", err)
	}

	state := &pbAdminStateResponse{
		NeedsBootstrap: needsBootstrap,
	}
	if session != nil {
		state.Session = session
		if view, loadErr := loadPBProviders(app, oauthAddr); loadErr == nil {
			state.Providers = view
		}
	}
	return state, http.StatusOK, nil
}

func ensurePBNeedsBootstrap(app *pocketbase.PocketBase) (int, error) {
	total, err := app.Dao().TotalAdmins()
	if err != nil {
		return http.StatusInternalServerError, errors.New("failed to check admin state")
	}
	if total > 0 {
		return http.StatusBadRequest, errors.New("admin already initialized")
	}
	return http.StatusOK, nil
}

func bootstrapPBAdmin(app *pocketbase.PocketBase, email, password, confirm string) (*pbAdminSession, int, error) {
	admin := &models.Admin{}
	form := forms.NewAdminUpsert(app, admin)
	form.Email = strings.TrimSpace(email)
	form.Password = password
	form.PasswordConfirm = confirm

	if err := form.Submit(); err != nil {
		status, formatted := classifyFormError(err)
		return nil, status, formatted
	}

	return &pbAdminSession{
		ID:    admin.Id,
		Email: admin.Email,
	}, http.StatusOK, nil
}

func authenticatePBAdmin(app *pocketbase.PocketBase, identity, password string) (*pbAdminSession, int, error) {
	form := forms.NewAdminLogin(app)
	form.Identity = strings.TrimSpace(identity)
	form.Password = password

	admin, err := form.Submit()
	if err != nil {
		status, formatted := classifyFormError(err)
		return nil, status, formatted
	}

	return &pbAdminSession{
		ID:    admin.Id,
		Email: admin.Email,
	}, http.StatusOK, nil
}

func issuePBAdminCookie(w http.ResponseWriter, app *pocketbase.PocketBase, session *pbAdminSession) error {
	if app == nil || session == nil {
		return errors.New("missing session")
	}
	admin, err := app.Dao().FindAdminById(session.ID)
	if err != nil {
		return err
	}
	token, err := tokens.NewAdminAuthToken(app, admin)
	if err != nil {
		return err
	}
	duration := time.Duration(app.Settings().AdminAuthToken.Duration) * time.Second
	cookie := &http.Cookie{
		Name:     pbAdminCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	if duration > 0 {
		cookie.Expires = time.Now().Add(duration)
		cookie.MaxAge = int(duration.Seconds())
	}
	http.SetCookie(w, cookie)
	return nil
}

func clearPBAdminCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     pbAdminCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func ensurePBAdminSession(app *pocketbase.PocketBase, req *http.Request) (*pbAdminSession, int, error) {
	session, err := sessionFromRequest(app, req)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, http.StatusUnauthorized, errors.New("admin session missing")
		}
		log.Printf("pocketbase admin session: %v", err)
		return nil, http.StatusUnauthorized, errors.New("admin session invalid")
	}
	if session == nil {
		return nil, http.StatusUnauthorized, errors.New("admin session missing")
	}
	return session, http.StatusOK, nil
}

func sessionFromRequest(app *pocketbase.PocketBase, req *http.Request) (*pbAdminSession, error) {
	if app == nil {
		return nil, errors.New("pocketbase unavailable")
	}
	cookie, err := req.Cookie(pbAdminCookieName)
	if err != nil {
		return nil, err
	}
	token := strings.TrimSpace(cookie.Value)
	if token == "" {
		return nil, http.ErrNoCookie
	}
	admin, err := app.Dao().FindAdminByToken(token, app.Settings().AdminAuthToken.Secret)
	if err != nil {
		return nil, err
	}
	return &pbAdminSession{
		ID:    admin.Id,
		Email: admin.Email,
	}, nil
}

func loadPBProviders(app *pocketbase.PocketBase, oauthAddr string) (*pbProvidersView, error) {
	settingsClone, err := app.Settings().Clone()
	if err != nil || settingsClone == nil {
		return nil, errors.New("settings not available")
	}
	redirectURL := buildDefaultRedirectURL(oauthAddr)
	return &pbProvidersView{
		Google:      providerPayloadFrom(settingsClone.GoogleAuth),
		Github:      providerPayloadFrom(settingsClone.GithubAuth),
		RedirectURL: redirectURL,
	}, nil
}

func providerPayloadFrom(cfg settings.AuthProviderConfig) pbAuthProviderPayload {
	return pbAuthProviderPayload{
		Enabled:      cfg.Enabled,
		ClientID:     cfg.ClientId,
		ClientSecret: cfg.ClientSecret,
	}
}

func updatePBProviders(app *pocketbase.PocketBase, payload pbProvidersPayload) (int, error) {
	form := forms.NewSettingsUpsert(app)
	mergeProviderPayload(&form.GoogleAuth, payload.Google)
	mergeProviderPayload(&form.GithubAuth, payload.Github)

	if err := form.Submit(); err != nil {
		return classifyFormError(err)
	}
	return http.StatusOK, nil
}

func mergeProviderPayload(dst *settings.AuthProviderConfig, src pbAuthProviderPayload) {
	if dst == nil {
		return
	}
	dst.Enabled = src.Enabled
	dst.ClientId = strings.TrimSpace(src.ClientID)
	dst.ClientSecret = strings.TrimSpace(src.ClientSecret)
}

func classifyFormError(err error) (int, error) {
	if err == nil {
		return http.StatusOK, nil
	}
	switch e := err.(type) {
	case validation.Errors:
		msg := make(map[string]string, len(e))
		for field, ferr := range e {
			if ferr != nil {
				msg[field] = ferr.Error()
			}
		}
		raw, _ := json.Marshal(msg)
		return http.StatusBadRequest, errors.New(string(raw))
	default:
		return http.StatusBadRequest, err
	}
}

func buildDefaultRedirectURL(addr string) string {
	host := strings.TrimSpace(addr)
	if host == "" {
		host = "127.0.0.1:8008"
	}
	if !strings.Contains(host, "://") {
		host = "http://" + host
	}
	if !strings.HasSuffix(host, "/oauth2-redirect") {
		host = strings.TrimRight(host, "/") + "/oauth2-redirect"
	}
	return host
}
