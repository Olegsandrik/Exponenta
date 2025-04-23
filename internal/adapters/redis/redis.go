package redis

import (
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/Olegsandrik/Exponenta/config"
)

type Adapter struct {
	pool *redis.Pool
}

func NewRedisAdapter(cfg *config.Config) (*Adapter, error) {
	pool := &redis.Pool{
		MaxIdle:     10,
		MaxActive:   100,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial(cfg.RedisNetwork, cfg.RedisURL,
				redis.DialPassword(cfg.RedisPassword),
			)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	conn := pool.Get()
	defer conn.Close()

	_, err := conn.Do("PING")
	if err != nil {
		return nil, err
	}

	return &Adapter{
		pool: pool,
	}, nil
}

func (a *Adapter) GetConn() redis.Conn {
	return a.pool.Get()
}

func (a *Adapter) Get(key string) (uint, error) {
	conn := a.pool.Get()
	defer conn.Close()
	ansBytes, err := redis.Bytes(conn.Do("GET", key))
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
	conn := a.pool.Get()
	defer conn.Close()
	_, err := conn.Do("SET", key, value, "EX", 86400)
	if err != nil {
		return err
	}
	return nil
}

func (a *Adapter) Delete(key string) error {
	conn := a.pool.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", key)
	if err != nil {
		return err
	}
	return nil
}
