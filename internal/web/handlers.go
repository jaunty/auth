package web

import (
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jaunty/database/models"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/zikaeroh/ctxlog"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const cookieName = "jwt"

var (
	conflictColumns = []string{"sf"}
	updateColumns   = boil.Whitelist("access_token", "token_type", "refresh_token", "expiry")
)

func (s *Server) handleAuth(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	redir := q.Get("redirect")
	state := uuid.Must(uuid.NewV4()).String()

	s.states[state] = redir
	url := s.discord.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (s *Server) handleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	state := r.FormValue("state")
	if state == "" {
		ctxlog.Debug(ctx, "empty state was passed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	redir, ok := s.states[state]
	if !ok {
		ctxlog.Debug(ctx, "wrong state was passed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		ctxlog.Error(ctx, "error beginning transaction", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer tx.Rollback()

	tok, err := s.discord.Exchange(ctx, r.FormValue("code"))
	if err != nil {
		ctxlog.Error(ctx, "error exchanging code to Discord", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := s.discord.User(ctx, tok.AccessToken)
	if err != nil {
		ctxlog.Error(ctx, "error getting discord user with associated token", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sf := user.ID

	exists, err := models.Users(qm.Where("sf = ?", sf)).Exists(ctx, tx)
	if err != nil {
		ctxlog.Error(ctx, "error getting user from database", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !exists {
		nu := &models.User{
			SF: sf,
		}

		if err := nu.Insert(ctx, tx, boil.Infer()); err != nil {
			ctxlog.Error(ctx, "error inserting new user into database", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	dt := &models.DiscordToken{
		SF:           sf,
		AccessToken:  tok.AccessToken,
		TokenType:    tok.TokenType,
		RefreshToken: tok.RefreshToken,
		Expiry:       tok.Expiry,
	}

	if err := dt.Upsert(ctx, tx, true, conflictColumns, updateColumns, boil.Infer()); err != nil {
		ctxlog.Error(ctx, "error inserting/upserting into database", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	claims := map[string]interface{}{
		"user_id": sf,
	}

	_, ts, err := s.tokenAuth.Encode(claims)
	if err != nil {
		ctxlog.Error(ctx, "error encoding claims", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		ctxlog.Error(ctx, "error committing transaction", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	expireDur := (time.Hour * 24) * 30
	expire := time.Now().Add(expireDur)

	cook := &http.Cookie{
		Name:     cookieName,
		Value:    ts,
		Path:     "/",
		Domain:   s.domain,
		Expires:  expire,
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
	}

	if redir == "" {
		redir = "https://" + s.domain
	}

	if !strings.HasPrefix(redir, "https://") {
		redir = "https://" + redir
	}

	http.SetCookie(w, cook)
	w.Header().Set("Location", redir)
	w.WriteHeader(http.StatusSeeOther)
}
