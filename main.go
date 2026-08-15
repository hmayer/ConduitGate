package main

import (
	"log"

	"conduitgate/internal/config"
	"conduitgate/internal/forwarder"
	"conduitgate/internal/listener"
	"conduitgate/internal/router"
)

func main() {
	log.Println("Initializing ConduitGate...")

	cfg, err := config.LoadConfig("routes.json")
	if err != nil {
		log.Fatalf("Failed to load config from routes.json: %v", err)
	}
	r := router.NewRouter(cfg)

	mf := forwarder.NewMultiForwarder()
	mf.Register("http", forwarder.NewHTTPForwarder())
	mf.Register("https", forwarder.NewHTTPForwarder())
	mf.Register("amqp", forwarder.NewAMQPForwarder())
	mf.Register("rabbitmq", forwarder.NewAMQPForwarder())
	mf.Register("sns", forwarder.NewSNSForwarder())
	mf.Register("sqs", forwarder.NewSQSForwarder())

	l := listener.NewListener(cfg, r, mf)

	if err := l.Start(); err != nil {
		log.Fatalf("Listener failed: %v", err)
	}
}
