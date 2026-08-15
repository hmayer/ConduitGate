# RabbitMQ Example

This example runs the full ConduitGate flow end-to-end with Docker Compose: a webhook is sent to ConduitGate, routed to a RabbitMQ exchange, and consumed by a small example service.

## What's Included

- **rabbitmq** — RabbitMQ broker with the management plugin enabled, pre-provisioned with the `webhook-exchange` topic exchange, the `webhook-events` queue, and the binding between them.
- **conduitgate** — Built from the repository's `Dockerfile`, configured with `examples/rabbitmq/routes.json` to forward requests on `/webhook-test` to RabbitMQ.
- **consumer** — A minimal Python service that connects to RabbitMQ, consumes messages from `webhook-events`, and prints each payload.
- **producer** — A one-shot `curl` container that sends a sample webhook to ConduitGate once it's reachable, so the flow runs automatically without a manual step.

## Files

```
docker-compose.yml
examples/rabbitmq/
├── routes.json
├── rabbitmq/
│   ├── rabbitmq.conf
│   └── definitions.json
└── consumer/
    ├── Dockerfile
    ├── requirements.txt
    └── consumer.py
```

## 1. Run the Example

From the repository root:

```bash
docker compose up --build
```

This builds the ConduitGate and consumer images, starts RabbitMQ and waits for it to become healthy, starts ConduitGate and the consumer, then sends a sample webhook through the `producer` container.

## 2. Observe the Flow

Watch the terminal output:

- `producer` logs its `curl` request and ConduitGate's `202 Accepted` response, then exits.
- `conduitgate` logs that it published a message to the `webhook-exchange` exchange.
- `consumer` logs the JSON payload it received from the `webhook-events` queue.

The `rabbitmq`, `conduitgate`, and `consumer` services keep running after `producer` exits — stop them with `Ctrl+C` or `docker compose down` (see step 4).

You can also inspect the exchange, queue, and message rates in the RabbitMQ management UI at http://localhost:15672 (login `guest` / `guest`).

## 3. Send Additional Webhooks

With the stack running, send more events manually:

```bash
curl -X POST http://localhost:8080/webhook-test \
     -H "Content-Type: application/json" \
     -d '{"event": "order.updated", "data": {"id": 67890}}'
```

Each request appears in the `consumer` logs a few moments later.

## 4. Stop the Example

```bash
docker compose down
```
