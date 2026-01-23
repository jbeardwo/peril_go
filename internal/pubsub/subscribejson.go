package pubsub

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T),
) error {
	ch, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("error binding queue: %v", err)
	}

	deliveries, err := ch.Consume(
		queueName,
		"",
		false,
		false, false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("error consuming queue: %v", err)
	}

	go func() {
		for delivery := range deliveries {
			var body T
			err := json.Unmarshal(delivery.Body, &body)
			if err != nil {
				log.Printf("failed to unmarshal: %v", err)
				continue
			}
			handler(body)
			err = delivery.Ack(false)
			if err != nil {
				log.Printf("failed to ack: %v", err)
				continue
			}
		}
	}()

	return nil
}
