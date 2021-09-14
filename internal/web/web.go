package web

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth/v5"
	"github.com/holedaemon/discord"
	"github.com/jaunty/web"
	"golang.org/x/oauth2"
)

var ErrInvalidOption = errors.New("auth: invalid option")

type Server struct {
	domain string

	s      *web.Server
	conf   *oauth2.Config
	states map[string]string

	tokenAuth *jwtauth.JWTAuth
	discord   *discord.Client
	db        *sql.DB
}

func New(addr string, opts ...Option) (*Server, error) {
	s := &Server{
		states: make(map[string]string),
	}

	for _, o := range opts {
		o(s)
	}

	if s.conf == nil {
		return nil, fmt.Errorf("%w: an OAuth2 config is required", ErrInvalidOption)
	}

	if s.tokenAuth == nil {
		return nil, fmt.Errorf("%w: a TokenAuth is required", ErrInvalidOption)
	}

	srv, err := web.New(addr)
	if err != nil {
		return nil, err
	}

	s.s = srv
	return s, nil
}

func (s *Server) router() *chi.Mux {
	r := chi.NewRouter()

	r.Get("/", s.handleAuth)
	r.Get("/callback", s.handleCallback)
	return r
}

func (s *Server) Start(ctx context.Context) error {
	return s.s.Start(ctx,
		web.Router(s.router()),
	)
}
