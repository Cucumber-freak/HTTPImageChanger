package server

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitClient struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	Queue   string
}

func ConnectRabbit(url, queueName string) *RabbitClient {
	conn, err := amqp.Dial(url)
	if err != nil {
		panic(fmt.Sprintf("RabbitMQ Error: %v", err))
	}

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		panic(err)
	}

	return &RabbitClient{Conn: conn, Channel: ch, Queue: queueName}
}

func (r *RabbitClient) Publish(ctx context.Context, taskID string) error {
	return r.Channel.PublishWithContext(
		ctx,
		"",
		r.Queue,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(taskID),
		},
	)
}

func (r *RabbitClient) Close() {
	r.Channel.Close()
	r.Conn.Close()
}
