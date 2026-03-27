package main

import (
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/go-chi/chi/v5"
	"github.com/tobyrushton/padel-stats/libs/auth"
	ss "github.com/tobyrushton/padel-stats/libs/auth/sessions"
	"github.com/tobyrushton/padel-stats/libs/config"
	"github.com/tobyrushton/padel-stats/libs/db/postgres"
	"github.com/tobyrushton/padel-stats/libs/db/postgres/sessions"
	"github.com/tobyrushton/padel-stats/libs/db/postgres/users"
	"github.com/tobyrushton/padel-stats/pkg/api/handlers"
)

func main() {
	r := chi.NewRouter()

	cfg, err := config.MustLoadConfig()
	if err != nil {
		panic(err)
	}

	db, err := postgres.NewDb(cfg.DBConnectionString)
	if err != nil {
		panic(err)
	}

	// register auth routes

	ur, err := users.NewRepository(db)
	if err != nil {
		panic(err)
	}

	sr, err := sessions.NewRepository(db)
	if err != nil {
		panic(err)
	}

	ss, err := ss.NewService(
		sr,
		cfg.JWTSecret,
		cfg.JWTIssuer,
		time.Duration(cfg.SessionTTLSeconds)*time.Second,
	)
	if err != nil {
		panic(err)
	}

	as, err := auth.NewService(ur, ss)
	if err != nil {
		panic(err)
	}

	ah := handlers.NewAuthHandler(as)
	ah.RegisterRoutes(r)

	lambda.Start(httpadapter.NewV2(r).ProxyWithContext)
}
