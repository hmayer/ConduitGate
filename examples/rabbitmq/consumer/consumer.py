import json
import os
import time

import pika

AMQP_URL = os.environ.get("AMQP_URL", "amqp://guest:guest@rabbitmq:5672/")
EXCHANGE = os.environ.get("EXCHANGE", "webhook-exchange")
QUEUE = os.environ.get("QUEUE", "webhook-events")
ROUTING_KEY = os.environ.get("ROUTING_KEY", "webhook.events")


def connect_with_retry(url, attempts=10, delay=3):
    params = pika.URLParameters(url)
    last_err = None
    for attempt in range(1, attempts + 1):
        try:
            return pika.BlockingConnection(params)
        except pika.exceptions.AMQPConnectionError as err:
            last_err = err
            print(f"[consumer] connection attempt {attempt}/{attempts} failed: {err}", flush=True)
            time.sleep(delay)
    raise SystemExit(f"[consumer] could not connect to RabbitMQ: {last_err}")


def on_message(channel, method, _properties, body):
    try:
        payload = json.loads(body)
    except ValueError:
        payload = body.decode("utf-8", errors="replace")

    print("[consumer] received webhook event:", flush=True)
    if isinstance(payload, (dict, list)):
        print(json.dumps(payload, indent=2), flush=True)
    else:
        print(payload, flush=True)

    channel.basic_ack(delivery_tag=method.delivery_tag)


def main():
    connection = connect_with_retry(AMQP_URL)
    channel = connection.channel()

    # Idempotent: safe even though RabbitMQ is already provisioned via
    # definitions.json. Declaring here also lets this script run standalone
    # against any broker, following the usual AMQP convention of consumers
    # owning their queue's existence.
    channel.exchange_declare(exchange=EXCHANGE, exchange_type="topic", durable=True)
    channel.queue_declare(queue=QUEUE, durable=True)
    channel.queue_bind(queue=QUEUE, exchange=EXCHANGE, routing_key=ROUTING_KEY)

    channel.basic_qos(prefetch_count=1)
    channel.basic_consume(queue=QUEUE, on_message_callback=on_message)

    print(
        f"[consumer] waiting for messages on queue '{QUEUE}' "
        f"bound to '{EXCHANGE}' (routing key '{ROUTING_KEY}')",
        flush=True,
    )
    try:
        channel.start_consuming()
    except KeyboardInterrupt:
        channel.stop_consuming()
    finally:
        connection.close()


if __name__ == "__main__":
    main()
