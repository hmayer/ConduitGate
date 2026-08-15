package main

import (
	"log/slog"
	"os"

	"conduitgate/internal/config"
	"conduitgate/internal/forwarder"
	"conduitgate/internal/listener"
	"conduitgate/internal/router"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	slog.Info("Initializing ConduitGate...")

	cfg, err := config.LoadConfig("routes.json")
	if err != nil {
		slog.Error("Failed to load config", "path", "routes.json", "error", err.Error())
		os.Exit(1)
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
		slog.Error("Listener failed", "error", err.Error())
		os.Exit(1)
	}
}
