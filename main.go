package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/holedaemon/discord"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jaunty/auth/internal/web"
	"golang.org/x/oauth2"
)

func main() {
	confPath := flag.String("c", "auth.toml", "Path to config file.")
	dbTimeout := flag.Duration("t", time.Second*20, "Time between DB connection attempts")
	dbMax := flag.Int("m", 20, "The number of times to attempt a DB connection")

	flag.Parse()

	conf, err := loadConfig(*confPath)
	if err != nil {
		log.Fatalln("Unable to load config file:", err)
	}

	var (
		db        *sql.DB
		connected = false
	)
	for i := 0; i < *dbMax && !connected; i++ {
		db, err = sql.Open("pgx", conf.DSN)
		if err != nil {
			log.Printf("Unable to open connection to DB, trying again in %s: %s\n", dbTimeout, err)
			time.Sleep(*dbTimeout)
			continue
		}

		if err = db.Ping(); err != nil {
			log.Printf("Unable to ping DB, trying again in %s: %s\n", dbTimeout, err)
			time.Sleep(*dbTimeout)
			continue
		}

		connected = true
	}

	oa2 := &oauth2.Config{
		ClientID:     conf.OAuth2.ClientID,
		ClientSecret: conf.OAuth2.ClientSecret,
		Endpoint:     discord.Endpoint,
		RedirectURL:  conf.OAuth2.RedirectURI,
		Scopes:       conf.OAuth2.Scopes,
	}
	ta := jwtauth.New("HS256", []byte(conf.Secret), nil)
	srv, err := web.New(
		web.WithAddr(conf.Addr),
		web.WithDB(db),
		web.WithDomain(conf.Domain),
		web.WithTokenAuth(ta),
		web.WithOAuth2Config(oa2),
	)
	if err != nil {
		log.Fatalln("Unable to create server:", err)
	}

	ctx := context.Background()
	if err := srv.Start(ctx); err != nil {
		log.Fatalln("Unable to start server:", err)
	}
}
