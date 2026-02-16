package config

import (
	"os"
	"strconv"
)

type Config struct {
	JWTSecret string
	JWTIssuer string
	JWTTTL    int64 // seconds
}

func Load() Config {
	return Config{
		JWTSecret: os.Getenv("JWT_SECRET"),
		JWTIssuer: os.Getenv("JWT_ISSUER"),
		JWTTTL:    readInt64("JWT_TTL", 3600),
	}
}

func (c Config) Valid() bool {
	return c.JWTSecret != "" && c.JWTIssuer != "" && c.JWTTTL > 0
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
