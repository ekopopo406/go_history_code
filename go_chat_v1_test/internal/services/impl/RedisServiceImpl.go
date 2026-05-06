package impl

import (
	"context"
	"fmt"
	"go_chat_v1_test/internal/logger"
	"go_chat_v1_test/internal/services"
	apptypes "go_chat_v1_test/internal/types"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisStore 封装 redis 客户端，方便以后加方法
type RedisServiceImpl struct {
	mainRedis apptypes.MainRedis
	mu        sync.RWMutex
	logger    *logger.Manager
}

// NewRedisStore 创建 redis 客户端
func NewRedisServiceImpl(mainRedis apptypes.MainRedis, logger *logger.Manager) services.RedisServices {
	return &RedisServiceImpl{mainRedis: mainRedis, logger: logger}
}

// 接口实现
func (s *RedisServiceImpl) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return s.mainRedis.RedisClient.Set(ctx, key, value, expiration).Err()
}

func (s *RedisServiceImpl) Get(ctx context.Context, key string) (any, error) {
	return s.mainRedis.RedisClient.Get(ctx, key).Result()
}

func (s *RedisServiceImpl) GetSubChannel(ctx context.Context, key string) (any, error) {
	existChannels := s.mainRedis.RedisClient.PubSubChannels(ctx, key)
	return existChannels, nil
}

func (s *RedisServiceImpl) Delete(ctx context.Context, key string) error {
	return s.mainRedis.RedisClient.Del(ctx, key).Err()
}

func (s *RedisServiceImpl) Exists(ctx context.Context, key string) (bool, error) {
	exists, err := s.mainRedis.RedisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// Any Key Exist Check
func (r *RedisServiceImpl) IsExistKeyValue(ctx context.Context, key string, expected interface{}) (bool, error) {
	// 检查 key 是否存在
	exists, err := r.mainRedis.RedisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if exists == 0 {
		return false, nil
	}

	// 获取 Redis 中的值（总是字符串）
	redisVal, err := r.mainRedis.RedisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	// 根据期望值的类型进行转换和比较
	switch v := expected.(type) {
	case string:
		return redisVal == v, nil
	case int, int8, int16, int32, int64:
		// 将 Redis 字符串转为 int 比较
		redisInt, err := strconv.ParseInt(redisVal, 10, 64)
		if err != nil {
			return false, nil
		}
		return redisInt == reflect.ValueOf(v).Int(), nil
	case float32, float64:
		redisFloat, err := strconv.ParseFloat(redisVal, 64)
		if err != nil {
			return false, nil
		}
		return redisFloat == reflect.ValueOf(v).Float(), nil
	case bool:
		redisBool, err := strconv.ParseBool(redisVal)
		if err != nil {
			return false, nil
		}
		return redisBool == v, nil
	default:
		// 默认转为字符串比较
		return redisVal == fmt.Sprint(v), nil
	}

}

func (r *RedisServiceImpl) MakeSubscribe(ctx context.Context, channels ...string) (*redis.PubSub, error) {
	pubsub := r.mainRedis.RedisClient.PSubscribe(ctx, channels...)
	return pubsub, nil
}

func (r *RedisServiceImpl) UnMakeSubscribe(ctx context.Context, subscribeName string) (bool, error) {

	return true, nil
}

func (r *RedisServiceImpl) PublishMessage(ctx context.Context, channel string, message interface{}) (bool, error) {
	err := r.mainRedis.RedisClient.Publish(ctx, channel, message).Err()
	if err != nil {
		r.logger.Redis.Error("Redis Publish failed",
			zap.String("channel", channel),
			zap.Error(err))
		return false, err
	}
	r.logger.Redis.Info("Message published", zap.String("channel", channel))

	return false, nil
}

func (r *RedisServiceImpl) StartSubScribe(ctx context.Context, channels []string, handler func(channel, payload string)) error {
	ticker := time.NewTicker(30 * time.Second)
	pubsub, err := r.MakeSubscribe(ctx, channels...)
	if err != nil {
		return err
	}

	r.logger.Redis.Info("Redis PubSub subscriber started", zap.Strings("channels", channels))

	go func() {
		defer func() {
			if err := pubsub.Close(); err != nil {
				r.logger.Redis.Error("PubSub close error", zap.Error(err))
			}
		}()

		ch := pubsub.Channel()

		for {
			select {
			case <-ticker.C:
				pubsub.Close()
				r.logger.Redis.Info("PubSub subscriber shutting down...")
			case <-ctx.Done():
				r.logger.Redis.Info("PubSub subscriber shutting down...")
				return
			case msg, ok := <-ch:
				if !ok {
					r.logger.Redis.Warn("PubSub channel closed")
					return
				}
				handler(msg.Channel, msg.Payload)
			}
		}
	}()

	return nil
}
