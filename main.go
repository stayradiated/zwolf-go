package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
)

var amqpAddress = "amqp://guest:guest@rabbitmq:5672/"

var (
	consumeTopic = "events"
	publishTopic = "events-processed"

	logger = watermill.NewStdLogger(
		true,  // debug
		false, // trace
	)
)

type event struct {
	ID int `json:"id"`
}

type processedEvent struct {
	ProcessedID int       `json:"processed_id"`
	Time        time.Time `json:"time"`
}

func main() {
	commandsAMQPConfig := amqp.NewDurableQueueConfig(amqpAddress)

	publisher, err := amqp.NewPublisher(commandsAMQPConfig, logger)
	if err != nil {
		panic(err)
	}

	subscriber, err := amqp.NewSubscriber(commandsAMQPConfig, logger)
	if err != nil {
		panic(err)
	}

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		panic(err)
	}

	router.AddPlugin(plugin.SignalsHandler)
	router.AddMiddleware(middleware.Recoverer)

	// Adding a handler (multiple handlers can be added)
	router.AddHandler(
		"handler_1",  // handler name, must be unique
		consumeTopic, // topic from which messages should be consumed
		subscriber,
		publishTopic, // topic to which messages should be published
		publisher,
		func(msg *message.Message) ([]*message.Message, error) {
			consumedPayload := event{}
			err := json.Unmarshal(msg.Payload, &consumedPayload)
			if err != nil {
				// When a handler returns an error, the default behavior is to send a Nack (negative-acknowledgement).
				// The message will be processed again.
				//
				// You can change the default behaviour by using middlewares, like Retry or PoisonQueue.
				// You can also implement your own middleware.
				return nil, err
			}

			log.Printf("received event %+v", consumedPayload)

			newPayload, err := json.Marshal(processedEvent{
				ProcessedID: consumedPayload.ID,
				Time:        time.Now(),
			})
			if err != nil {
				return nil, err
			}

			newMessage := message.NewMessage(watermill.NewUUID(), newPayload)

			return []*message.Message{newMessage}, nil
		},
	)

	// Simulate incoming events in the background
	go simulateEvents(publisher)

	if err := router.Run(context.Background()); err != nil {
		panic(err)
	}
}

// simulateEvents produces events that will be later consumed.
func simulateEvents(publisher message.Publisher) {
	i := 0
	for {
		e := event{
			ID: i,
		}

		payload, err := json.Marshal(e)
		if err != nil {
			panic(err)
		}

		err = publisher.Publish(consumeTopic, message.NewMessage(
			watermill.NewUUID(), // internal uuid of the message, useful for debugging
			payload,
		))
		if err != nil {
			panic(err)
		}

		i++

		time.Sleep(time.Second)
	}
}
