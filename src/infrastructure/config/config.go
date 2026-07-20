package config

import (
	"fmt"
	"time"

	e "github.com/ChatDetectiveORG/shared/errors"

	"github.com/spf13/viper"
)

const PodType = "event_loop"

// DefaultTokenBucketRate is the global Telegram Bot API rate budget shared by every chat-export and
// message-sender pod. Telegram's documented hard limit for bots is 30 messages per second;
// keeping that as the default ensures we never need a separate per-chat limiter on top.
const DefaultTokenBucketRate = 30

const DefaultReferralRewardManagerInterval = 12 * time.Hour

const DefaultLevelTerminationInterval = 12 * time.Hour

type Config struct {
	RuntimeConfig  *RuntimeConfig
	RedisConfig    *RedisConfig
	RabbitMQConfig *RabbitMQConfig
}

type RabbitMQConfig struct {
	Host     string
	Port     string
	Username string
	Password string
}

func (cfg *RabbitMQConfig) URL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port)
}

type RuntimeConfig struct {
	PodID                         string
	PodType                       string
	TokenBucketRate               int
	ReferralRewardManagerInterval time.Duration
	LevelTerminationInterval      time.Duration
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
	viper.SetDefault("REFERRAL_REWARD_MANAGER_INTERVAL", DefaultReferralRewardManagerInterval)
	viper.SetDefault("LEVEL_TERMINATION_INTERVAL", DefaultLevelTerminationInterval)

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
		RabbitMQConfig: &RabbitMQConfig{
			Host:     viper.GetString("RABBITMQ_HOST"),
			Port:     viper.GetString("RABBITMQ_PORT"),
			Username: viper.GetString("RABBITMQ_USERNAME"),
			Password: viper.GetString("RABBITMQ_PASSWORD"),
		},
		RuntimeConfig: &RuntimeConfig{
			PodID:                         viper.GetString("POD_ID"),
			PodType:                       PodType,
			TokenBucketRate:               viper.GetInt("TOKEN_BUCKET_RATE"),
			ReferralRewardManagerInterval: viper.GetDuration("REFERRAL_REWARD_MANAGER_INTERVAL"),
			LevelTerminationInterval:      viper.GetDuration("LEVEL_TERMINATION_INTERVAL"),
		},
	}
	return cfg, e.Nil()
}
