package redis

import (
	"sync"
	"time"

	"github.com/ChatDetectiveORG/event-loop/src/infrastructure/config"
	e "github.com/ChatDetectiveORG/shared/errors"

	"github.com/gomodule/redigo/redis"
)

var (
	poolOnce sync.Once
	pool     *redis.Pool
)

func newPool(cfg *config.Config) *redis.Pool {
	dial := func() (redis.Conn, error) {
		opts := []redis.DialOption{
			redis.DialConnectTimeout(cfg.RedisConfig.ConnectionTimeout),
			redis.DialReadTimeout(cfg.RedisConfig.ReadTimeout),
			redis.DialWriteTimeout(cfg.RedisConfig.WriteTimeout),
		}
		if cfg.RedisConfig.Password != "" {
			opts = append(opts, redis.DialPassword(cfg.RedisConfig.Password))
		}
		opts = append(opts, redis.DialDatabase(cfg.RedisConfig.Database))
		return redis.Dial("tcp", cfg.RedisConfig.Host+":"+cfg.RedisConfig.Port, opts...)
	}

	return &redis.Pool{
		MaxIdle:     cfg.RedisConfig.MaxIdle,
		MaxActive:   cfg.RedisConfig.MaxActive,
		IdleTimeout: cfg.RedisConfig.IdleTimeout,
		Wait:        cfg.RedisConfig.Wait,
		Dial:        dial,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < 30*time.Second {
				return nil
			}
			_, err := redis.String(c.Do("PING"))
			return err
		},
	}
}

func GetPool(cfg *config.Config) (*redis.Pool, *e.ErrorInfo) {
	poolOnce.Do(func() { pool = newPool(cfg) })
	return pool, e.Nil()
}

func InitRedis(cfg *config.Config) *e.ErrorInfo {
	p, err := GetPool(cfg)
	if !err.IsNil() {
		return err
	}
	conn := p.Get()
	defer func() { _ = conn.Close() }()
	if conn.Err() != nil {
		return e.FromError(conn.Err(), "failed to get redis connection").WithSeverity(e.Critical)
	}
	if _, pingErr := redis.String(conn.Do("PING")); pingErr != nil {
		return e.FromError(pingErr, "redis ping failed").WithSeverity(e.Critical)
	}
	return e.Nil()
}
