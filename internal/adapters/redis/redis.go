package redis

import (
	"context"
	"strconv"

	"github.com/gomodule/redigo/redis"

	"github.com/Olegsandrik/Exponenta/config"
	"github.com/Olegsandrik/Exponenta/logger"
)

type Adapter struct {
	conn redis.Conn
}

func NewRedisAdapter(cfg *config.Config) (*Adapter, error) {
	logger.Info(context.Background(), cfg.RedisNetwork)
	conn, err := redis.Dial(cfg.RedisNetwork, cfg.RedisURL,
		redis.DialPassword(cfg.RedisPassword),
	)

	if err != nil {
		return nil, err
	}

	_, err = conn.Do("PING")
	if err != nil {
		return nil, err
	}

	return &Adapter{
		conn: conn,
	}, nil
}

func (a *Adapter) Close() error {
	return a.conn.Close()
}

func (a *Adapter) Get(key string) (uint, error) {
	ansBytes, err := redis.Bytes(a.conn.Do("GET", key))
	if err != nil {
		return 0, err
	}

	ans, err := strconv.ParseUint(string(ansBytes), 10, 32)
	if err != nil {
		return 0, err
	}

	return uint(ans), nil
}

func (a *Adapter) Set(key string, value uint) error {
	_, err := a.conn.Do("SET", key, value, "EX", 86400)
	if err != nil {
		return err
	}
	return nil
}

func (a *Adapter) Delete(key string) error {
	_, err := a.conn.Do("DEL", key)
	if err != nil {
		return err
	}
	return nil
}
