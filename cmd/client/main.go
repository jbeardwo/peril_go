package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
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

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("could not get username: %v", err)
	}

	gameState := gamelogic.NewGameState(username)

	queueName := routing.PauseKey + "." + username
	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gameState),
	)
	if err != nil {
		log.Fatalf("problem subscribing: %v", err)
	}

	moveQueueName := "army_moves." + username
	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		moveQueueName,
		"army_moves.*",
		pubsub.Transient,
		handlerMove(gameState, connection),
	)
	if err != nil {
		log.Fatalf("problem subscribing: %v", err)
	}

	warKey := routing.WarRecognitionsPrefix + ".*"
	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		"war", // shared durable queue
		warKey,
		pubsub.Durable, // or whatever durable queue config you’re using
		handlerWar(gameState, cha),
	)
	if err != nil {
		log.Fatalf("problem subscribing to war: %v", err)
	}

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "spawn":
			err = gameState.CommandSpawn(words)
			if err != nil {
				fmt.Printf("Problem spawning units: %v", err)
			}
		case "move":
			move, err := gameState.CommandMove(words)
			if err != nil {
				fmt.Printf("Problem moving units: %v", err)
			} else {
				fmt.Printf("Successful move: %v", move)
				err = pubsub.PublishJSON(cha, routing.ExchangePerilTopic, moveQueueName, move)
				if err != nil {
					log.Fatalf("problem publishing Json: %v", err)
				}
				fmt.Println("Move published successfully")
			}

		case "status":
			gameState.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			if len(words) != 2 {
				fmt.Println("spam requires one argument, ex: spam 2")
			} else {
				n, err := strconv.Atoi(words[1])
				if err != nil {
					fmt.Println("Invalid input for n, ex: spam 2")
				}
				for range n {
					msg := gamelogic.GetMaliciousLog()
					gl := routing.GameLog{
						CurrentTime: time.Now(),
						Message:     msg,
						Username:    username,
					}
					if err := pubsub.PublishGameLog(cha, gl); err != nil {
						fmt.Printf("failed to publish log: %v", err)
					}
				}

			}

		case "quit":
			fmt.Println("disconnecting...")
			return

		default:
			fmt.Println("unknown command")
		}
	}

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("Peril connection closed")
}
