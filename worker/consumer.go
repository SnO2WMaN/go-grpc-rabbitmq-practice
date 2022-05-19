package main

import (
	"flag"
	"log"

	"github.com/streadway/amqp"
)

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

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Recieved message: %s", d.Body)
		}
	}()

	log.Printf("Ctrl+C to interrupt")
	<-forever
}
