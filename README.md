# ConduitGate

ConduitGate is an HTTP event gateway that receives webhooks and routes them to queues, services, or other endpoints.

Receive once. Route anywhere.

It decouples webhook ingestion from processing, allowing you to scale and evolve your architecture without changing your integrations.

## Example
A single webhook can be routed to multiple destinations:

```json
{
  "port": 8080,
  "rules": [
    {
      "sourcePath": "/stripe/payment_succeeded",
      "destinations": [
        {
          "protocol": "sqs",
          "region": "us-east-1",
          "queueURL": "https://sqs.us-east-1.amazonaws.com/123456789/PaymentSucceededQueue"
        },
        {
          "protocol": "http",
          "url": "http://internal-api/payments"
        }
      ]
    }
  ]
}
```
Send one webhook → route to multiple backends.

## Why ConduitGate?

Handling webhooks directly inside your application leads to:

- Tight coupling between external services and internal logic
- Hard-to-maintain integrations
- Difficult scaling and retry strategies
- Vendor lock-in to specific messaging systems

ConduitGate solves this by acting as an ingestion layer that routes events to the right destination without embedding business logic.

This keeps your system flexible, scalable, and easier to evolve over time.

## Features

- **Webhook Multiplexing**: Receive once, route to multiple destinations
- **Protocol Translation**: HTTP → AMQP, SQS, SNS, or other HTTP endpoints
- **Path-Based Routing**: Route events based on request paths
- **Asynchronous Processing**: Immediate response with background delivery
- **Decoupled Architecture**: Separate external integrations from internal systems

## Configuration Guide

## Configuration Guide

ConduitGate is configured using a `routes.json` file located in the root directory.

Each rule defines a source path and one or more destinations.

### Protocol-Specific Configuration

Each destination in the `destinations` array must specify a `protocol` and relevant connection details.

#### 1. HTTP
- `protocol`: `"http"` or `"https"`
- `url`: The destination URL.
- `headers` (optional): A map of custom headers to include in the request.
- `username` (optional): Basic Auth username.
- `password` (optional): Basic Auth password.

#### 2. RabbitMQ (AMQP)
- `protocol`: `"amqp"` or `"rabbitmq"`
- `url`: The AMQP connection string (e.g., `amqp://guest:guest@localhost:5672/`).
- `exchange`: The exchange name to publish the message to.
- `routingKey`: The routing key for the message.

#### 3. AWS SNS
- `protocol`: `"sns"`
- `region`: The AWS region (e.g., `us-east-1`).
- `topicARN`: The Amazon Resource Name (ARN) of the SNS topic.
- `accessKey`: AWS access key.
- `secretKey`: AWS secret key.

#### 4. AWS SQS
- `protocol`: `"sqs"`
- `region`: The AWS region (e.g., `us-east-1`).
- `queueURL`: The URL of the SQS queue.
- `accessKey`: AWS access key.
- `secretKey`: AWS secret key.

## Project Usage

### 1. Prepare Configuration

Create a `routes.json` file in the root directory. You can use `routes.sample.json` as a template.

```bash
cp routes.sample.json routes.json
```

Edit `routes.json` to match your requirements.

### 2. Run the Application

To start ConduitGate, run:

```bash
go run main.go
```

The server will start listening on the port specified in `routes.json` (default is `8080`).

### 3. Sending Webhooks

ConduitGate accepts `POST` requests at any of the paths defined in your `rules`.

Example using `curl`:

```bash
curl -X POST http://localhost:8080/webhook-test \
     -H "Content-Type: application/json" \
     -d '{"event": "order.created", "data": {"id": 12345}}'
```

The request is immediately acknowledged with HTTP `202 Accepted`, while delivery happens asynchronously.

This ensures external services are never blocked by internal processing delays.

### 4. Running Tests

To run the unit tests:

```bash
go test ./...
```

## Docker

ConduitGate ships an official multi-stage `Dockerfile` that builds a small, static binary and runs it in a minimal, non-root image.

### 1. Build the Image

```bash
docker build -t conduitgate .
```

### 2. Run the Container

The image bundles `routes.sample.json` as its default `routes.json`, so it starts out of the box:

```bash
docker run -p 8080:8080 conduitgate
```

To use your own configuration, mount it over the default at `/app/routes.json`:

```bash
docker run -p 8080:8080 -v $(pwd)/routes.json:/app/routes.json conduitgate
```

If your `routes.json` sets a different `"port"`, update the `-p` mapping (`-p <host-port>:<port>`) to match.

## Philosophy

ConduitGate treats webhooks as events, not HTTP requests.

Instead of handling them directly, it routes them to systems designed to process them — queues, services, or other endpoints.