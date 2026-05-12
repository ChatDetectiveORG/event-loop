package config

import (
	"time"

	e "github.com/ChatDetectiveORG/shared/errors"

	"github.com/spf13/viper"
)

const PodType = "event_loop"

// DefaultTokenBucketRate is the global Telegram Bot API rate budget shared by every chat-export and
// message-sender pod. Telegram's documented hard limit for bots is 30 messages per second;
// keeping that as the default ensures we never need a separate per-chat limiter on top.
const DefaultTokenBucketRate = 30

type Config struct {
	RuntimeConfig *RuntimeConfig
	RedisConfig   *RedisConfig
}

type RuntimeConfig struct {
	PodID           string
	PodType         string
	TokenBucketRate int
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	Database int

	MaxIdle     int
	MaxActive   int
	IdleTimeout time.Duration
	Wait        bool

	ConnectionTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
}

func FetchConfig() (*Config, *e.ErrorInfo) {
	viper.AutomaticEnv()
	viper.SetDefault("TOKEN_BUCKET_RATE", DefaultTokenBucketRate)

	cfg := &Config{
		RedisConfig: &RedisConfig{
			Host:              viper.GetString("REDIS_HOST"),
			Port:              viper.GetString("REDIS_PORT"),
			Password:          viper.GetString("REDIS_PASSWORD"),
			Database:          viper.GetInt("REDIS_DB"),
			MaxIdle:           viper.GetInt("REDIS_MAX_IDLE"),
			MaxActive:         viper.GetInt("REDIS_MAX_ACTIVE"),
			IdleTimeout:       viper.GetDuration("REDIS_IDLE_TIMEOUT"),
			Wait:              viper.GetBool("REDIS_WAIT"),
			ConnectionTimeout: viper.GetDuration("REDIS_CONNECTION_TIMEOUT"),
			ReadTimeout:       viper.GetDuration("REDIS_READ_TIMEOUT"),
			WriteTimeout:      viper.GetDuration("REDIS_WRITE_TIMEOUT"),
		},
		RuntimeConfig: &RuntimeConfig{
			PodID:           viper.GetString("POD_ID"),
			PodType:         PodType,
			TokenBucketRate: viper.GetInt("TOKEN_BUCKET_RATE"),
		},
	}
	return cfg, e.Nil()
}
