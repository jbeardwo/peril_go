package pubsub

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeGob[T any](
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
			buf := bytes.NewBuffer(delivery.Body)
			dec := gob.NewDecoder(buf)
			err := dec.Decode(&body)
			if err != nil {
				log.Printf("failed to gob decode: %v", err)
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
