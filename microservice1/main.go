package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	requestID := uuid.New().String()

	log := log.New(os.Stdout, "[m1] ", log.LstdFlags)

	// Подключение к RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Fatalf("Ошибка подключения к RabbitMQ:\n %s", err)
	}
	defer conn.Close()

	chnl, err := conn.Channel()
	if err != nil {
		log.Fatalf("Ошибка создания канала:\n %s", err)
	}
	defer chnl.Close()

	q, err := chnl.QueueDeclare(
		"myQueue", // Имя очереди
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Fatalf("Ошибка объявления очереди:\n %s", err)
	}

	log.Printf("routing key %s", q.Name)

	// Создание экземпляра логгера для Fluentd
	logger, err := fluent.New(fluent.Config{
		FluentPort: 24224,     // Порт Fluentd в контейнере
		FluentHost: "fluentd", // Имя сервиса Fluentd в Docker Compose
		// Укажите ваш порт для Fluentd
	})
	if err != nil {
		log.Fatalf("Ошибка создания логгера: %v", err)
	}
	defer logger.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body := "Hello World!"

	forever := true

	for forever {
		for i := 0; i < 10; i++ {
			go func() {
				// Отправка сообщения в RabbitMQ
				err = chnl.PublishWithContext(ctx, "", q.Name, false, false, amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(body + "_" + string(requestID)),
				})
				if err != nil {
					log.Fatalf("Ошибка публикации сообщения в RabbitMQ:\n %s", err)
				}

				log.Printf("[Request ID: %s] Sending log from microservice1", requestID)

				// Отправка лога в Fluentd с использованием тега 'your_microservice_logs'
				err = logger.Post("[m1]", map[string]string{
					"request_id": requestID,
					"message":    body,
				})
				if err != nil {
					log.Fatalf("Ошибка отправки лога в Fluentd: %v", err)
				}

				log.Println("Сообщение успешно отправлено в RabbitMQ и Fluentd")

			}()
		}
		time.Sleep(5e9)
	}

}
