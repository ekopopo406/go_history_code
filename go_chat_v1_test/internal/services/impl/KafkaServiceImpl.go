package impl

import (
	"context"
	"encoding/json"

	"go_chat_v1_test/internal/services"
	apptypes "go_chat_v1_test/internal/types"
	"log"

	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaService struct {
	mainKafkaProducer apptypes.MainKafkaProducer
	mainKafkaConsumer apptypes.MainKafkaConsumer
}

func NewKafkaServiceImpl(mainKafkaProducer apptypes.MainKafkaProducer, mainKafkaConsumer apptypes.MainKafkaConsumer) services.KafkaServices {
	return &KafkaService{mainKafkaProducer: mainKafkaProducer, mainKafkaConsumer: mainKafkaConsumer}
}

func (p *KafkaService) Publish(ctx context.Context, topic string, key []byte, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}

	record := &kgo.Record{
		Topic: topic,
		Key:   key,
		Value: b,
	}

	// 异步发，带回调
	p.mainKafkaProducer.ProducerClient.Produce(ctx, record, func(r *kgo.Record, err error) {
		if err != nil {
			log.Printf("produce fail topic=%s key=%s err=%v", topic, string(key), err)
			// 可以在这里埋 metric / alert
		}
	})

	return nil
}
