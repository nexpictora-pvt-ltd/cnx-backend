// publisher.go

package messaging

import (
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewPublisher(amqpURL string) (*Publisher, error) {
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

	return &Publisher{
		conn:    conn,
		channel: channel,
	}, nil
}

func (p *Publisher) Close() {
	// Close the channel and connection
	p.channel.Close()
	p.conn.Close()
}

func (p *Publisher) PublishMessage(queueName string, message []byte, ctx *gin.Context) error {
	// Declare a queue
	_, err := p.channel.QueueDeclare(
		queueName, // queue name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return err
	}

	// Publish a message to the queue
	err = p.channel.PublishWithContext(ctx,
		"",        // exchange
		queueName, // routing key (queue name)
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        message,
		},
	)
	return err
}
