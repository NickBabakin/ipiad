package rabbitmqgo

import (
	"context"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func initRabbit() (*amqp.Channel, *amqp.Connection) {
	//conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/") // with docker
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/") // without docker
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
	}
	failOnError(err, "Failed to open a channel")

	return ch, conn
}

func Send(body []byte, queueName string) {
	ch, conn := initRabbit()
	defer conn.Close()
	defer ch.Close()

	q, err := ch.QueueDeclare(
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

	err = ch.PublishWithContext(ctx,
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

func Receive(queueName string, wg_ext *sync.WaitGroup, chv chan *amqp.Delivery) {
	defer wg_ext.Done()

	ch, conn := initRabbit()
	defer conn.Close()
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for msg := range msgs {
			chv <- &msg
			if string(msg.Body) == "stop" {
				log.Printf(" [*] Stopped receiving from %s\n", queueName)
				break
			}
		}
	}()

	log.Printf(" [*] Waiting for messages from %s\n", queueName)
	wg.Wait()
}
