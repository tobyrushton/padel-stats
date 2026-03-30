package config

import (
	"errors"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/sst/sst/v3/sdk/golang/resource"
)

type Config struct {
	DBConnectionString string

	JWTSecret         string `envconfig:"JWT_SECRET" required:"true"`
	JWTIssuer         string `envconfig:"JWT_ISSUER" required:"true"`
	SessionTTLSeconds int    `envconfig:"SESSION_TTL_SECONDS" default:"86400"`
}

func MustLoadConfig() (*Config, error) {
	neonConnection, err := resource.Get("NeonDB", "connectionString")
	if err != nil {
		return nil, err
	}
	if neonConnection.(string) == "" {
		return nil, errors.New("neon db connection string must be set")
	}

	cfg := &Config{
		DBConnectionString: neonConnection.(string),
	}

	godotenv.Load()
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
