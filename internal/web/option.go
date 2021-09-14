package web

import (
	"database/sql"

	"github.com/go-chi/jwtauth/v5"
	"golang.org/x/oauth2"
)

type Option func(s *Server)

func WithOAuth2Config(oa2 *oauth2.Config) Option {
	return func(s *Server) {
		s.conf = oa2
	}
}

func WithTokenAuth(ta *jwtauth.JWTAuth) Option {
	return func(s *Server) {
		s.tokenAuth = ta
	}
}

func WithDomain(d string) Option {
	return func(s *Server) {
		s.domain = d
	}
}

func WithDB(db *sql.DB) Option {
	return func(s *Server) {
		s.db = db
	}
}

func WithAddr(addr string) Option {
	return func(s *Server) {
		s.addr = addr
	}
}
