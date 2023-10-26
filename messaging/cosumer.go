// consumer.go

package messaging

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewConsumer(amqpURL, queueName string) (*Consumer, error) {
	// Connect to RabbitMQ server
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	// Create a channel
	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Declare a queue
	_, err = channel.QueueDeclare(
		queueName, // queue name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		conn:    conn,
		channel: channel,
	}, nil
}

func (c *Consumer) Close() {
	// Close the channel and connection
	c.channel.Close()
	c.conn.Close()
}

func (c *Consumer) ConsumeMessages(queueName string) (<-chan amqp.Delivery, error) {
	// Consume messages from the queue
	messages, err := c.channel.Consume(
		queueName, // queue name
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return nil, err
	}

	return messages, nil
}
