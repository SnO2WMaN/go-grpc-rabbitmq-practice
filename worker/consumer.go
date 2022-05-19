package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"time"

	"github.com/mmcdole/gofeed"
	pb "github.com/sno2wman/go-rabbitmq-grpc-practice/sayhello"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	rabbitmqUrl = flag.String(
		"rabbitmqUrl",
		"",
		"The RabbitMQ URL",
	)
	grpcUrl = flag.String(
		"grpcUrl",
		"",
		"The gRPC URL",
	)
)

type RecvProtocol struct {
	FeedId    string `json:"feed_id"`
	Link      string `json:"link"`
	Timestamp int64  `json:"timestamp"`
}

type FeedPayload struct {
	Id string
}

func buildPayload(feed *gofeed.Feed, feedId string) FeedPayload {
	return FeedPayload{Id: feedId}
}

func main() {
	flag.Parse()

	if *rabbitmqUrl == "" {
		log.Fatalf("RabbitMQ URL not provided")
	}
	if *grpcUrl == "" {
		log.Fatalf("gRPC URL not provided")
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
		"rss", // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
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

	grpcConn, err := grpc.Dial(*grpcUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}
	defer grpcConn.Close()

	c := pb.NewGreeterClient(grpcConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() {
		for d := range msgs {
			var pr RecvProtocol
			if err := json.Unmarshal(d.Body, &pr); err != nil {
				log.Fatalf("Failed to unmarshal: %v", err)
			}
			var link = pr.Link
			var feedId = pr.FeedId

			fp := gofeed.NewParser()
			feed, err := fp.ParseURL(link)
			if err != nil {
				log.Fatalf("Failed to fetch RSS: %v", err)
			}

			payload := buildPayload(feed, feedId)
			r, err := c.UpdateFeed(ctx, &pb.UpdateFeedRequest{FeedId: payload.Id})
			if err != nil {
				log.Fatalf("could not greet: %v", err)
			}
			log.Printf("Status: %v", r.GetOk())
		}
	}()

	log.Printf("Ctrl+C to interrupt")
	<-forever
}
