package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"

	"github.com/stayradiated/zwolf-signal/signal"
)

var logger = watermill.NewStdLogger(
	true,  // debug
	false, // trace
)

func main() {
	amqpAddress := os.Getenv("AMQP_ADDRESS")
	queueConfig := amqp.NewDurableQueueConfig(amqpAddress)

	// publisher, err := amqp.NewPublisher(queueConfig, logger)
	// if err != nil {
	// 	panic(err)
	// }

	subscriber, err := amqp.NewSubscriber(queueConfig, logger)
	if err != nil {
		panic(err)
	}

	messages, err := subscriber.Subscribe(context.Background(), "received-messages")
	if err != nil {
		panic(err)
	}

	for msg := range messages {
		fmt.Printf("received message: %s, payload: %s\n", msg.UUID, string(msg.Payload))
		msg.Ack()

		smsg := signal.Message{}
		err := json.Unmarshal(msg.Payload, &smsg)
		if err != nil {
			log.Print(err)
			continue
		}

		fmt.Printf("received signal from %s: '%s'", smsg.Username, smsg.Text)
	}
}
