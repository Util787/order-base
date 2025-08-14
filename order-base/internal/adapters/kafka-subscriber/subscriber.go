package kafka_subscriber

import (
	"context"
	"log/slog"

	"github.com/Util787/order-base/internal/config"
	"github.com/Util787/order-base/internal/models"
	"github.com/segmentio/kafka-go"
)

type OrderUsecase interface {
	SaveOrder(ctx context.Context, order models.Order) error
}

type KafkaSubscriber struct {
	log          *slog.Logger
	kafkaReader  *kafka.Reader
	orderUsecase OrderUsecase
	messageCh    chan kafka.Message
}

func NewKafkaSubscriber(log *slog.Logger, cfg config.KafkaConfig, orderUsecase OrderUsecase, msgChanBuf uint) *KafkaSubscriber {
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		GroupID:  cfg.GroupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		MaxWait:  cfg.MaxWait,
	})

	return &KafkaSubscriber{
		log:          log,
		kafkaReader:  kafkaReader,
		orderUsecase: orderUsecase,
		messageCh:    make(chan kafka.Message, msgChanBuf),
	}
}

func (k *KafkaSubscriber) Shutdown() error {
	k.kafkaReader.Close()
	return nil
}

func (k *KafkaSubscriber) Subscribe(ctx context.Context, numFetchers int, numHandlers int) {
	for i := 0; i < numFetchers; i++ {
		go k.fetcher(ctx)
	}

	for i := 0; i < numHandlers; i++ {
		go k.saveOrderHandler(ctx)
	}
}
