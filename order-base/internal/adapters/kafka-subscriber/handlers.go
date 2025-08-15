package kafka_subscriber

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/Util787/order-base/internal/common"
	"github.com/Util787/order-base/internal/models"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type message struct {
	content *kafka.Message
	UID     string
}

func (k *KafkaSubscriber) fetcher(ctx context.Context) {
	log := k.log.With(slog.String("op", common.GetOperationName()))

	for {
		kafkaMsg, err := k.kafkaReader.FetchMessage(ctx)
		msgUID := uuid.NewString()

		log := log.With(slog.String("message_uid", msgUID))
		log.Debug("fetching message from Kafka", slog.Any("message", kafkaMsg))

		if err != nil {
			log.Error("failed to fetch message from Kafka", slog.String("error", err.Error()))
			continue
		}

		k.messageCh <- message{
			content: &kafkaMsg,
			UID:     msgUID,
		}
	}
}

func (k *KafkaSubscriber) saveOrderHandler(ctx context.Context) {
	log := k.log.With(slog.String("op", common.GetOperationName()))

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-k.messageCh:
			start := time.Now()
			log := log.With(slog.String("message_uid", msg.UID))
			log.Info("start handling message", slog.Time("start", start))

			var order models.Order
			err := json.Unmarshal(msg.content.Value, &order)
			if err != nil {
				log.Error("failed to unmarshal order", slog.String("error", err.Error()))
				continue
			}

			err = k.orderUsecase.SaveOrder(ctx, order)
			if err != nil {
				log.Error("failed to save order", slog.String("error", err.Error()))
				continue
			}

			if err := k.kafkaReader.CommitMessages(ctx, *msg.content); err != nil {
				log.Error("failed to commit message", slog.String("error", err.Error()))
			} else {
				log.Debug("order saved successfully", slog.String("order_id", order.OrderUID), slog.Int64("duration", time.Since(start).Milliseconds()))
			}
		}
	}
}
