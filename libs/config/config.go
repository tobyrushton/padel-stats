package config

import (
	"errors"

	"github.com/sst/sst/v3/sdk/golang/resource"
)

type Config struct {
	DBConnectionString string
}

func MustLoadConfig() (*Config, error) {
	neonConnection, err := resource.Get("NeonDB", "connectionString")
	if err != nil {
		return nil, err
	}
	if neonConnection.(string) == "" {
		return nil, errors.New("neon db connection string must be set")
	}

	return &Config{
		DBConnectionString: neonConnection.(string),
	}, nil
}
