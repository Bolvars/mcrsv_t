package main

import (
	"log"
	"os"
	"strings"

	"github.com/fluent/fluent-logger-golang/fluent"
	//"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	//requestID := uuid.New().String()

	log := log.New(os.Stdout, "[m2] ", log.LstdFlags)

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

	msgs, err := chnl.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Ошибка подписки на очередь: %s", err)
	}

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

	log.Println("Ожидание сообщений. Для выхода нажмите CTRL+C")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			// Отправка логов в Fluentd с использованием тега 'your_microservice_logs'

			err = logger.Post("[m2]", map[string]string{
				"request_id": strings.Split(string(d.Body), "_")[1],
				"message":    strings.Split(string(d.Body), "_")[0],
			})
			if err != nil {
				log.Fatalf("Ошибка отправки лога в Fluentd: %v", err)
			}

			log.Printf("Получено сообщение: %s", d.Body)
		}
	}()

	<-forever
}
