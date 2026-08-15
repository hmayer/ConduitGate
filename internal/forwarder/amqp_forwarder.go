package forwarder

import (
	"conduitgate/internal/config"
	"context"
	"fmt"
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type amqpClient struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

type AMQPForwarder struct {
	mu      sync.RWMutex
	clients map[string]*amqpClient
}

func NewAMQPForwarder() *AMQPForwarder {
	return &AMQPForwarder{
		clients: make(map[string]*amqpClient),
	}
}

func (f *AMQPForwarder) getClient(url string) (*amqpClient, error) {
	f.mu.RLock()
	client, ok := f.clients[url]
	f.mu.RUnlock()

	if ok && client.conn != nil && !client.conn.IsClosed() && client.ch != nil && !client.ch.IsClosed() {
		return client, nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// Check again in case another goroutine already created it
	client, ok = f.clients[url]
	if ok && client.conn != nil && !client.conn.IsClosed() && client.ch != nil && !client.ch.IsClosed() {
		return client, nil
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AMQP: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	client = &amqpClient{
		conn: conn,
		ch:   ch,
	}
	f.clients[url] = client

	return client, nil
}

func (f *AMQPForwarder) Forward(ctx context.Context, payload []byte, destination config.Destination) error {
	if destination.Protocol != "amqp" && destination.Protocol != "rabbitmq" {
		return fmt.Errorf("unsupported protocol: %s", destination.Protocol)
	}

	client, err := f.getClient(destination.URL)
	if err != nil {
		return err
	}

	err = client.ch.PublishWithContext(ctx,
		destination.Exchange,   // exchange
		destination.RoutingKey, // routing key
		false,                  // mandatory
		false,                  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payload,
		})

	if err != nil {
		// If publishing fails, we might want to invalidate the client so it's recreated next time
		f.mu.Lock()
		delete(f.clients, destination.URL)
		f.mu.Unlock()
		client.ch.Close()
		client.conn.Close()
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	log.Printf("Message published to AMQP exchange: %s", destination.Exchange)
	return nil
}
