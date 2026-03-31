package main

import (
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/go-chi/chi/v5"
	"github.com/tobyrushton/padel-stats/libs/auth"
	ss "github.com/tobyrushton/padel-stats/libs/auth/sessions"
	"github.com/tobyrushton/padel-stats/libs/config"
	"github.com/tobyrushton/padel-stats/libs/db/postgres"
	dbgames "github.com/tobyrushton/padel-stats/libs/db/postgres/games"
	dbleaderboard "github.com/tobyrushton/padel-stats/libs/db/postgres/leaderboard"
	"github.com/tobyrushton/padel-stats/libs/db/postgres/sessions"
	"github.com/tobyrushton/padel-stats/libs/db/postgres/users"
	gamelib "github.com/tobyrushton/padel-stats/libs/games"
	leaderboardlib "github.com/tobyrushton/padel-stats/libs/leaderboard"
	"github.com/tobyrushton/padel-stats/pkg/api/handlers"
)

func main() {
	r := chi.NewRouter()
	r.Use(corsMiddleware)

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

	gr, err := dbgames.NewRepository(db)
	if err != nil {
		panic(err)
	}

	gs, err := gamelib.NewService(gr)
	if err != nil {
		panic(err)
	}

	gh := handlers.NewGamesHandler(gs, ss)
	gh.RegisterRoutes(r)

	lr, err := dbleaderboard.NewRepository(db)
	if err != nil {
		panic(err)
	}

	ls, err := leaderboardlib.NewService(lr)
	if err != nil {
		panic(err)
	}

	lh := handlers.NewLeaderboardHandler(ls)
	lh.RegisterRoutes(r)

	lambda.Start(httpadapter.NewV2(r).ProxyWithContext)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
