package main

import (
	"context"
	"news-kafka/api-gateway/pkg/api"

	//"news-kafka/api-gateway/pkg/kafka"
	//"news-kafka/api-gateway/pkg/logger"

	"fmt"
	"log"
	"net/http"

	"github.com/IBM/sarama"
	"github.com/PolinaSvet/kafka"
	"github.com/PolinaSvet/logger"
)

// tasknews> go test ./... -coverprofile=coverage.out
// tasknews> go test ./... -v -coverprofile=coverage.out

// http://127.0.0.1:8080/news/{rubric}/{countNews}
// http://127.0.0.1:8080/newsDetailed?id_news=1

//https://localhost:9443
//psql -U postgres -d prgComments
//psql -U postgres -d prgNews
//\dt+

// Сервер
type server struct {
	api *api.API
}

const (
	VAR_SERVICENAMAE = "NEWSNAMESERVISE"
	VAR_CONFIGKAFKA  = "./configKafka.json"
	VAR_LOG          = "./logs.json"
)

func main() {

	fmt.Println("api-gateway:", logger.GetServiceName(VAR_SERVICENAMAE))
	fmt.Println("api-gateway:", logger.GetLocalIP())

	//==============================================
	//Logger
	//==============================================
	logs, err := logger.NewLogger(VAR_LOG, 50, VAR_SERVICENAMAE)
	if err != nil {
		fmt.Printf("Error creating logger: %v", err)
	}
	defer logs.Close()

	errorChannel := make(chan error, 100)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel when we are finished consuming integers

	// выводим ошибки
	go handleErrors(ctx, errorChannel, logs)

	// Создаём объект сервера.
	var srv server

	//==============================================
	//Kafka
	//==============================================
	// Чтение конфигурации (предположим, что конфигурация хранится в файле config.json)
	configKafka, err := api.ReadConfig(VAR_CONFIGKAFKA)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	// Создание Kafka Producer и Consumer
	kafkaProducer, err := kafka.NewProducer(configKafka.KafkaBrokers)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	kafkaConsumer, err := kafka.NewConsumer(configKafka.KafkaBrokers)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer kafkaConsumer.Close()

	// Запуск гоурутины для потребления сообщений service-news
	responseNewsCh, err := kafkaConsumer.Consume(configKafka.TopicReceivedNews, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Failed to consume partition News: %v", err)
	}

	// Запуск гоурутины для потребления сообщений service-news
	responseOneNewsCh, err := kafkaConsumer.Consume(configKafka.TopicReceivedOneNews, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Failed to consume partition News: %v", err)
	}

	// Запуск гоурутины для потребления сообщений service-comments
	responseCommentsCh, err := kafkaConsumer.Consume(configKafka.TopicReceivedComments, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Failed to consume partition Comments: %v", err)
	}

	// Запуск гоурутины для потребления сообщений service-comments
	responseAddCommentsCh, err := kafkaConsumer.Consume(configKafka.TopicReceivedAddComments, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Failed to consume partition Comments: %v", err)
	}

	// Запуск гоурутины для потребления сообщений service-censor
	responseCensorCh, err := kafkaConsumer.Consume(configKafka.TopicReceivedAddCensor, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Failed to consume partition Censor: %v", err)
	}

	apiChannels := api.ApiChannels{
		ResponseNewsCh:        responseNewsCh,
		ResponseOneNewsCh:     responseOneNewsCh,
		ResponseCommentsCh:    responseCommentsCh,
		ResponseAddCommentsCh: responseAddCommentsCh,
		ResponseCensorCh:      responseCensorCh,
		ErrorChannel:          errorChannel,
	}

	srv.api = api.New(kafkaProducer, kafkaConsumer, configKafka, apiChannels, logger.GetServiceName(VAR_SERVICENAMAE))

	fmt.Println("Запуск веб-сервера на http://127.0.0.1:8080 ...")
	http.ListenAndServe(":8080", srv.api.Router())
}

func handleErrors(ctx context.Context, errs <-chan error, logs *logger.Logger) {
	for {
		select {
		case err, ok := <-errs:
			if !ok {
				return
			}
			logs.LogRequest(logger.GetRequestId(), logger.GetLocalIP(), 500, err.Error())
		case <-ctx.Done():
			return
		}
	}
}
