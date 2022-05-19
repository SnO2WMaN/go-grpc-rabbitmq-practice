package main

import (
	"encoding/json"
	"flag"
	"log"
	"time"

	"github.com/streadway/amqp"
)

type SendProtocol struct {
	FeedId    string `json:"feed_id"`
	Link      string `json:"link"`
	Timestamp int64  `json:"timestamp"`
}

var (
	rabbitmqUrl = flag.String(
		"rabbitmqUrl",
		"",
		"The RabbitMQ URL",
	)
)

func main() {
	flag.Parse()
	if *rabbitmqUrl == "" {
		log.Fatalf("RabbitMQ URL not provided")
	}

	conn, err := amqp.Dial(*rabbitmqUrl)

	if err != nil {
		log.Fatalf("Failed to connect to channel: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"rss", // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // no args
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	p := SendProtocol{
		FeedId:    "1",
		Link:      "https://jser.info/rss/",
		Timestamp: time.Now().UnixNano(),
	}
	bytes, err := json.Marshal(p)
	if err != nil {
		log.Fatalf("Failed to marshal message: %v", err)
	}

	if err = ch.Publish(
		"",     // exchange
		q.Name, // key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        bytes,
		},
	); err != nil {
		log.Fatalf("Failed to publish a message: %v", err)
	}

	log.Printf("Sent successfully")
}
