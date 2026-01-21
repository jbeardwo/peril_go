package pubsub

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType string

const (
	Durable   SimpleQueueType = "durable"
	Transient SimpleQueueType = "transient"
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	cha, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	q, err := cha.QueueDeclare(
		queueName,
		queueType == Durable,
		queueType == Transient,
		queueType == Transient,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("could not declare queue: %v", err)
	}

	err = cha.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		log.Fatalf("problem binding queue to exchange: %v", err)
	}

	return cha, q, nil
}
