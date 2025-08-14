package kafka_subscriber

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/Util787/order-base/internal/common"
	"github.com/Util787/order-base/internal/models"
)

func (k *KafkaSubscriber) fetcher(ctx context.Context) {
	log := k.log.With(slog.String("op", common.GetOperationName()))

	for {
		message, err := k.kafkaReader.FetchMessage(ctx)
		log.Info("fetching message from Kafka", slog.Any("message", message))
		if err != nil {
			log.Error("failed to fetch message from Kafka", slog.String("error", err.Error()))
			continue
		}

		k.messageCh <- message
	}
}

func (k *KafkaSubscriber) saveOrderHandler(ctx context.Context) {
	op := common.GetOperationName()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-k.messageCh:
			start := time.Now()
			// making request_id = msg.Key for every msg to track in logs
			ctx = context.WithValue(ctx, common.ContextKey("request_id"), msg.Key)
			log := common.LogOpAndReqId(ctx, op, k.log)
			log.Info("start handling message", slog.Time("start", start))

			var order models.Order
			err := json.Unmarshal(msg.Value, &order)
			if err != nil {
				log.Error("failed to unmarshal order", slog.String("error", err.Error()))
				continue
			}

			err = k.orderUsecase.SaveOrder(ctx, order)
			if err != nil {
				log.Error("failed to save order", slog.String("error", err.Error()))
				continue
			}

			if err := k.kafkaReader.CommitMessages(ctx, msg); err != nil {
				log.Error("failed to commit message", slog.String("error", err.Error()))
			} else {
				log.Debug("order saved successfully", slog.String("order_id", order.OrderUID), slog.Int64("duration", time.Since(start).Milliseconds()))
			}
		}
	}
}
