package config

import (
	"os"
	"strconv"
	"time"

	// Используется для загрузки переменных окружения из .env файла.
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	// Postgres
	PostgresDriverName   string
	PostgresPasswd       string
	PostgresEndpoint     string
	PostgresUser         string
	PostgresDBName       string
	PostgresPort         string
	PostgresDisable      string
	PostgresPublic       string
	PostgresMaxOpenConn  int
	PostgresConnIdleTime time.Duration

	// Server
	ServerTimeout time.Duration
	Port          string

	// Elasticsearch
	ElasticsearchAddress  string
	ElasticsearchUsername string
	ElasticsearchPassword string

	// Minio

	MinioUser     string
	MinioPassword string
	MinioEndpoint string
	MinioBucket   string
}

func NewConfig() *Config {
	return &Config{
		PostgresDriverName:   getEnvStr("POSTGRES_DRIVER_NAME", ""),
		PostgresPasswd:       getEnvStr("POSTGRES_PASSWD", ""),
		PostgresEndpoint:     getEnvStr("POSTGRES_ENDPOINT", ""),
		PostgresUser:         getEnvStr("POSTGRES_USER", ""),
		PostgresDBName:       getEnvStr("POSTGRES_DB_NAME", ""),
		PostgresPort:         getEnvStr("POSTGRES_PORT", ""),
		PostgresDisable:      getEnvStr("POSTGRES_DISABLE", ""),
		PostgresPublic:       getEnvStr("POSTGRES_PUBLIC", ""),
		PostgresMaxOpenConn:  getEnvInt("POSTGRES_MAX_OPEN_CONN", 10),
		PostgresConnIdleTime: getEnvTime("POSTGRES_CONN_IDLE_TIME", 10*time.Second),

		ServerTimeout: getEnvTime("SERVER_TIMEOUT", 5*time.Second),
		Port:          getEnvStr("SERVER_PORT", ":8080"),

		ElasticsearchAddress:  getEnvStr("ELASTIC_ADDRESS", ""),
		ElasticsearchUsername: getEnvStr("ELASTIC_USERNAME", ""),
		ElasticsearchPassword: getEnvStr("ELASTIC_PASSWORD", ""),

		MinioUser:     getEnvStr("MINIO_USER", ""),
		MinioPassword: getEnvStr("MINIO_PASSWD", ""),
		MinioEndpoint: getEnvStr("MINIO_ENDPOINT", ""),
		MinioBucket:   getEnvStr("MINIO_BUCKET_NAME", ""),
	}
}

func getEnvStr(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
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

func getEnvInt(key string, defaultValue int) int {
	if valueStr, ok := os.LookupEnv(key); ok {
		value, err := strconv.Atoi(valueStr)
		if err != nil {
			return defaultValue
		}
		return value
	}
	return defaultValue
}
