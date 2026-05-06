package services

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisServices interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error

	Get(ctx context.Context, key string) (any, error)

	Delete(ctx context.Context, key string) error

	Exists(ctx context.Context, key string) (bool, error)

	IsExistKeyValue(ctx context.Context, key string, value interface{}) (bool, error)

	MakeSubscribe(ctx context.Context, channels ...string) (*redis.PubSub, error)

	StartSubScribe(ctx context.Context, channels []string, handler func(channel, payload string)) error

	UnMakeSubscribe(ctx context.Context, subscribeName string) (bool, error)

	PublishMessage(ctx context.Context, channel string, message interface{}) (bool, error)

	GetSubChannel(ctx context.Context, key string) (any, error)
}
