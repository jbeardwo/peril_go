package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	connectionString := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer connection.Close()

	cha, err := connection.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	fmt.Println("Peril connected to RabbitMQ server")

	err = pubsub.SubscribeGob(connection,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.Durable,
		func(gl routing.GameLog) pubsub.Acktype {
			defer fmt.Print("> ")
			err = gamelogic.WriteLog(gl)
			if err != nil {
				fmt.Printf("problem with declaring/binding: %v", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		},
	)
	if err != nil {
		log.Fatalf("could not subscribe to logs: %v", err)
	}

	gamelogic.PrintServerHelp()
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "pause":
			val := routing.PlayingState{IsPaused: true}
			err = pubsub.PublishJSON(cha, routing.ExchangePerilDirect, routing.PauseKey, val)
			if err != nil {
				log.Fatalf("problem publishing Json: %v", err)
			}

		case "resume":
			val := routing.PlayingState{IsPaused: false}
			err = pubsub.PublishJSON(cha, routing.ExchangePerilDirect, routing.PauseKey, val)
			if err != nil {
				log.Fatalf("problem publishing Json: %v", err)
			}

		case "quit":
			fmt.Println("exiting server...")
			return

		default:
			fmt.Println("invalid command")
		}

	}

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("RabbitMQ connection closed")
}
