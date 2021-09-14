package main

import (
	"errors"
	"fmt"

	"github.com/BurntSushi/toml"
)

var errMissingOption = errors.New("auth: missing option")

func fmtError(msg string) error {
	return fmt.Errorf("%w: %s", errMissingOption, msg)
}

type config struct {
	Addr   string `toml:"addr"`
	Domain string `toml:"domain"`
	Secret string `toml:"secret"`
	DSN    string `toml:"dsn"`

	OAuth2 *oauth2Config `toml:"oauth2"`
}

type oauth2Config struct {
	ClientID     string   `toml:"client_id"`
	ClientSecret string   `toml:"client_secret"`
	RedirectURI  string   `toml:"redirect_uri"`
	Scopes       []string `toml:"scopes"`
}

func (c *config) check() error {
	if c.Addr == "" {
		return fmtError("addr")
	}

	if c.Domain == "" {
		return fmtError("domain")
	}

	if c.Secret == "" {
		return fmtError("secret")
	}

	if c.DSN == "" {
		return fmtError("dsn")
	}

	if c.OAuth2 == nil {
		return fmtError("oauth2 config")
	}

	if c.OAuth2.ClientID == "" {
		return fmtError("oauth2 client id")
	}

	if c.OAuth2.ClientSecret == "" {
		return fmtError("oauth2 client secret")
	}

	if c.OAuth2.RedirectURI == "" {
		return fmtError("oauth2 redirect uri")
	}

	if c.OAuth2.Scopes == nil {
		return fmtError("oauth2 scopes")
	}

	return nil
}

func loadConfig(path string) (*config, error) {
	c := new(config)
	if _, err := toml.DecodeFile(path, c); err != nil {
		return nil, err
	}

	if err := c.check(); err != nil {
		return nil, err
	}

	return c, nil
}
