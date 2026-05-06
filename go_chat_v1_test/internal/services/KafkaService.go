package services

import "context"

type KafkaServices interface {
	Publish(ctx context.Context, topic string, key []byte, value any) error
}
