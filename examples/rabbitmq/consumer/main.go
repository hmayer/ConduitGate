package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func connectWithRetry(url string, attempts int, delay time.Duration) *amqp.Connection {
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		conn, err := amqp.Dial(url)
		if err == nil {
			return conn
		}
		lastErr = err
		log.Printf("[consumer] connection attempt %d/%d failed: %v", attempt, attempts, err)
		time.Sleep(delay)
	}
	log.Fatalf("[consumer] could not connect to RabbitMQ: %v", lastErr)
	return nil
}

func main() {
	amqpURL := getenv("AMQP_URL", "amqp://guest:guest@rabbitmq:5672/")
	exchange := getenv("EXCHANGE", "webhook-exchange")
	queue := getenv("QUEUE", "webhook-events")
	routingKey := getenv("ROUTING_KEY", "webhook.events")

	conn := connectWithRetry(amqpURL, 10, 3*time.Second)
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("[consumer] failed to open a channel: %v", err)
	}
	defer ch.Close()

	// Idempotent: safe even though RabbitMQ is already provisioned via
	// definitions.json. Declaring here also lets this run standalone
	// against any broker, following the usual AMQP convention of consumers
	// owning their queue's existence.
	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		log.Fatalf("[consumer] failed to declare exchange: %v", err)
	}
	if _, err := ch.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		log.Fatalf("[consumer] failed to declare queue: %v", err)
	}
	if err := ch.QueueBind(queue, routingKey, exchange, false, nil); err != nil {
		log.Fatalf("[consumer] failed to bind queue: %v", err)
	}
	if err := ch.Qos(1, 0, false); err != nil {
		log.Fatalf("[consumer] failed to set QoS: %v", err)
	}

	deliveries, err := ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("[consumer] failed to register consumer: %v", err)
	}

	log.Printf("[consumer] waiting for messages on queue '%s' bound to '%s' (routing key '%s')", queue, exchange, routingKey)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case d, ok := <-deliveries:
			if !ok {
				return
			}
			log.Printf("[consumer] received webhook event:\n%s", prettyJSON(d.Body))
			d.Ack(false)
		case <-stop:
			log.Println("[consumer] shutting down")
			return
		}
	}
}

func prettyJSON(body []byte) []byte {
	var buf bytes.Buffer
	if err := json.Indent(&buf, body, "", "  "); err != nil {
		return body
	}
	return buf.Bytes()
}
