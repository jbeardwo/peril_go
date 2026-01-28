package main

import (
	"fmt"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.Acktype {
	return func(row gamelogic.RecognitionOfWar) pubsub.Acktype {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(row)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue

		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard

		case gamelogic.WarOutcomeOpponentWon:
			message := winner + " won a war against " + loser
			gl := routing.GameLog{
				CurrentTime: time.Now(),
				Message:     message,
				Username:    row.Attacker.Username,
			}

			err := pubsub.PublishGameLog(ch, gl)
			if err != nil {
				return pubsub.NackRequeue
			}

			return pubsub.Ack

		case gamelogic.WarOutcomeYouWon:
			message := winner + " won a war against " + loser
			gl := routing.GameLog{
				CurrentTime: time.Now(),
				Message:     message,
				Username:    row.Attacker.Username,
			}

			err := pubsub.PublishGameLog(ch, gl)
			if err != nil {
				return pubsub.NackRequeue
			}

			return pubsub.Ack

		case gamelogic.WarOutcomeDraw:
			message := fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
			gl := routing.GameLog{
				CurrentTime: time.Now(),
				Message:     message,
				Username:    row.Attacker.Username,
			}

			err := pubsub.PublishGameLog(ch, gl)
			if err != nil {
				return pubsub.NackRequeue
			}

			return pubsub.Ack

		default:
			fmt.Print("war handler error")
			return pubsub.NackDiscard
		}
	}
}
