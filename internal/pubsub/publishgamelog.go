package pubsub

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishGameLog(ch *amqp.Channel, gl routing.GameLog) error {
	err := PublishGob(ch,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+gl.Username,
		gl,
	)
	if err != nil {
		return fmt.Errorf("error publishing gamelog: %v", err)
	}
	return nil
}
