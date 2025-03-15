package config

import (
	_ "github.com/joho/godotenv/autoload"
	"os"
	"strconv"
	"time"
)

var (
	// Postgres
	POSTGRES_DRIVER_NAME    = getEnvStr("POSTGRES_DRIVER_NAME", "")
	POSTGRES_PASSWD         = getEnvStr("POSTGRES_PASSWD", "")
	POSTGRES_ENDPOINT       = getEnvStr("POSTGRES_ENDPOINT", "")
	POSTGRES_USER           = getEnvStr("POSTGRES_USER", "")
	POSTGRES_DB_NAME        = getEnvStr("POSTGRES_DB_NAME", "")
	POSTGRES_PORT           = getEnvStr("POSTGRES_PORT", "")
	POSTGRES_DISABLE        = getEnvStr("POSTGRES_DISABLE", "")
	POSTGRES_PUBLIC         = getEnvStr("POSTGRES_PUBLIC", "")
	POSTGRES_MAX_OPEN_CONN  = getEnvInt("POSTGRES_MAX_OPEN_CONN", 10)
	POSTGRES_CONN_IDLE_TIME = getEnvTime("POSTGRES_CONN_IDLE_TIME", 10*time.Second)

	// Server

	SERVER_PORT    = getEnvStr("SERVER_PORT", "")
	SERVER_TIMEOUT = getEnvTime("SERVER_TIMEOUT", 5*time.Second)
)

func getEnvStr(key, zeroKey string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return zeroKey
}

func getEnvTime(key string, defaultValue time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		timeout, err := time.ParseDuration(value)
		if err != nil {
			return defaultValue
		}
		return timeout
	}
	return defaultValue
}

func getEnvInt(key string, zeroKey int) int {
	if valueStr, ok := os.LookupEnv(key); ok {
		value, err := strconv.Atoi(valueStr)
		if err != nil {
			return zeroKey
		}
		return value
	}
	return zeroKey
}
