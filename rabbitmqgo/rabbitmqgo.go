package rabbitmqgo

import (
	"context"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Rabbit struct {
	Conn *amqp.Connection
	Ch   *amqp.Channel
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func InitRabbit() *Rabbit {
	rabbit := Rabbit{nil, nil}
	var err error
	rabbit.Conn, err = amqp.Dial("amqp://guest:guest@localhost:5672/") // without docker
	//rabbit.Conn, err = amqp.Dial("amqp://guest:guest@rabbitmq:5672/") // with docker
	failOnError(err, "Failed to connect to RabbitMQ")

	rabbit.Ch, err = rabbit.Conn.Channel()
	if err != nil {
		rabbit.Conn.Close()
	}
	failOnError(err, "Failed to open a channel")
	return &rabbit
}

func Send(body []byte, queueName string) {
	rabbit := InitRabbit()

	q, err := rabbit.Ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = rabbit.Ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		})
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s\n", body)
}

func Receive(queueName string, wg_ext *sync.WaitGroup, rabbit *Rabbit) <-chan amqp.Delivery {
	defer wg_ext.Done()

	q, err := rabbit.Ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := rabbit.Ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	return msgs
}
