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
	addr   string
	domain string

	s      *web.Server
	conf   *oauth2.Config
	states map[string]string

	tokenAuth *jwtauth.JWTAuth
	discord   *discord.Client
	db        *sql.DB
}

func New(opts ...Option) (*Server, error) {
	s := &Server{
		states: make(map[string]string),
	}

	for _, o := range opts {
		o(s)
	}

	if err := s.verifyDefaults(); err != nil {
		return nil, err
	}

	dsc, err := discord.New(discord.WithOAuth2Config(s.conf))
	if err != nil {
		return nil, err
	}
	s.discord = dsc

	srv, err := web.New(s.addr)
	if err != nil {
		return nil, err
	}

	s.s = srv
	return s, nil
}

func (s *Server) verifyDefaults() error {
	if s.addr == "" {
		return fmt.Errorf("%w: missing addr", ErrInvalidOption)
	}

	if s.db == nil {
		return fmt.Errorf("%w: missing db", ErrInvalidOption)
	}

	if s.conf == nil {
		return fmt.Errorf("%w: missing oauth2 config", ErrInvalidOption)
	}

	if s.tokenAuth == nil {
		return fmt.Errorf("%w: missing jwtauth", ErrInvalidOption)
	}

	return nil
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
