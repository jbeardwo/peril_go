package pubsub

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Acktype int

const (
	Ack Acktype = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) Acktype,
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
			ack := handler(body)
			if ack == Ack {
				err = delivery.Ack(false)
				if err != nil {
					log.Printf("failed to ack: %v", err)
					continue
				}
				log.Printf("Acked")
				continue
			} else if ack == NackRequeue {
				err = delivery.Nack(false, true)
				if err != nil {
					log.Printf("failed to nack: %v", err)
					continue
				}
				log.Printf("Nack Requeued")
				continue
			} else if ack == NackDiscard {
				err = delivery.Nack(false, false)
				if err != nil {
					log.Printf("failed to nack: %v", err)
					continue
				}
				log.Printf("Nack Discarded")
				continue
			}
		}
	}()

	return nil
}
