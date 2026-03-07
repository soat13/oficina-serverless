package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DBDSN     string
	JWTSecret string
	JWTIssuer string
	JWTTTL    int64 // seconds
}

func Load() (Config, error) {
	conf := Config{
		DBDSN:     os.Getenv("DB_DSN"),
		JWTSecret: os.Getenv("JWT_SECRET"),
		JWTIssuer: os.Getenv("JWT_ISSUER"),
		JWTTTL:    readInt64("JWT_TTL", 3600),
	}

	if err := conf.validate(); err != nil {
		return Config{}, err
	}

	return conf, nil
}

func (c Config) validate() error {
	var missing []string

	if c.DBDSN == "" {
		missing = append(missing, "DB_DSN")
	}
	if c.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if c.JWTIssuer == "" {
		missing = append(missing, "JWT_ISSUER")
	}
	if c.JWTTTL <= 0 {
		missing = append(missing, "JWT_TTL (must be > 0)")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing or invalid environment variables: %s",
			strings.Join(missing, ", "))
	}

	return nil
}

func readInt64(key string, def int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	valueAsInt, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return valueAsInt
}
