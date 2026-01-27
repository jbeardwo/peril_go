package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerMove(gs *gamelogic.GameState, conn *amqp.Connection) func(gamelogic.ArmyMove) pubsub.Acktype {
	return func(move gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(move)
		switch outcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack

		case gamelogic.MoveOutcomeMakeWar:
			cha, err := conn.Channel()
			if err != nil {
				log.Printf("could not create channel: %v", err)
				return pubsub.NackDiscard
			}
			defer cha.Close()

			key := routing.WarRecognitionsPrefix + "." + gs.GetPlayerSnap().Username
			val := gamelogic.RecognitionOfWar{
				Attacker: move.Player,
				Defender: gs.GetPlayerSnap(),
			}
			err = pubsub.PublishJSON(cha, routing.ExchangePerilTopic, key, val)
			if err != nil {
				log.Printf("problem publishing Json: %v", err)
			}

			return pubsub.NackRequeue

		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard

		default:
			return pubsub.NackDiscard
		}
	}
}
