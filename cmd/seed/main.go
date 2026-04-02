package main

import (
	"context"
	"time"

	"github.com/tobyrushton/padel-stats/libs/auth"
	sessionsService "github.com/tobyrushton/padel-stats/libs/auth/sessions"
	"github.com/tobyrushton/padel-stats/libs/config"
	"github.com/tobyrushton/padel-stats/libs/db/models"
	"github.com/tobyrushton/padel-stats/libs/db/postgres"
	"github.com/tobyrushton/padel-stats/libs/db/postgres/sessions"
	"github.com/tobyrushton/padel-stats/libs/db/postgres/users"
)

// very basic seeding command for development, seeds an admin user
func main() {
	appConfig, err := config.MustLoadConfig()
	if err != nil {
		panic(err)
	}

	db, err := postgres.NewDb(appConfig.DBConnectionString)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ur, err := users.NewRepository(db)
	if err != nil {
		panic(err)
	}

	sr, err := sessions.NewRepository(db)
	if err != nil {
		panic(err)
	}

	ss, err := sessionsService.NewService(
		sr,
		appConfig.JWTSecret,
		appConfig.JWTIssuer,
		time.Duration(appConfig.SessionTTLSeconds)*time.Second,
	)

	us, err := auth.NewService(ur, ss)
	if err != nil {
		panic(err)
	}

	r, err := us.Signup(
		context.Background(),
		&auth.SignupInput{
			FirstName: "Toby",
			LastName:  "Rushton",
			Password:  "12345678",
			Username:  "toby",
		},
	)
	if err != nil {
		panic(err)
	}

	_, err = db.NewUpdate().
		Model((*models.User)(nil)).
		Set("is_admin = ?", true).
		Set("is_accepted_by_admin = ?", true).
		Where("id = ?", r.User.ID).
		Exec(context.Background())
}
