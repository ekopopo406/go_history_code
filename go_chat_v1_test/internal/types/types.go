package types

import (
	"github.com/redis/go-redis/v9"
	"github.com/twmb/franz-go/pkg/kgo"
	"gorm.io/gorm"
)

type MainDB struct{ Db *gorm.DB }

type MainRedis struct{ RedisClient *redis.Client }

type MainKafkaProducer struct{ ProducerClient *kgo.Client }

type MainKafkaConsumer struct{ ConsumerClient *kgo.Client }
