# Chevron Router

Chevron Router is a lightweight, high-performance webhook router and multiplexer written in Go. It allows you to receive incoming webhooks at a single endpoint and route them to multiple destinations based on the request path.

## Objectives

- **Webhook Multiplexing**: Forward a single incoming webhook to multiple backend services simultaneously.
- **Protocol Translation**: Receive HTTP/HTTPS webhooks and forward them to various message brokers and cloud services like RabbitMQ, AWS SNS, and AWS SQS.
- **Path-Based Routing**: Define routing rules based on the request URI path to direct traffic to the correct destinations.
- **Asynchronous Forwarding**: Ensure fast response times for incoming requests by processing outgoing forwards in the background.

## Configuration Guide

The application is configured using a `routes.json` file located in the root directory. This file defines the listening port and the routing rules.

### Basic Structure

```json
{
  "port": 8080,
  "rules": [
    {
      "sourcePath": "/path/to/webhook",
      "destinations": [
        {
          "protocol": "http",
          "url": "http://example.com/endpoint"
        }
      ]
    }
  ]
}
```

### Protocol-Specific Configuration

Each destination in the `destinations` array must specify a `protocol` and relevant connection details.

#### 1. HTTP / HTTPS
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

To start Chevron Router, run:

```bash
go run main.go
```

The server will start listening on the port specified in `routes.json` (default is `8080`).

### 3. Sending Webhooks

Chevron Router accepts `POST` requests at any of the paths defined in your `rules`.

Example using `curl`:

```bash
curl -X POST http://localhost:8080/webhook-test \
     -H "Content-Type: application/json" \
     -d '{"event": "order.created", "data": {"id": 12345}}'
```

The router will acknowledge receipt with an HTTP `202 Accepted` status and then asynchronously forward the payload to all configured destinations.

### 4. Running Tests

To run the unit tests:

```bash
go test ./...
```
