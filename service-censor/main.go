package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"news-kafka/service-censor/pkg/censor"

	//"news-kafka/service-censor/pkg/kafka"
	//"news-kafka/service-censor/pkg/logger"

	"sync"

	"fmt"
	"log"

	"github.com/IBM/sarama"
	"github.com/PolinaSvet/kafka"
	"github.com/PolinaSvet/logger"
)

// Сервер
type server struct {
	censor *censor.Censor
}

// Config - структура для хранения конфигурации
type ConfigKafka struct {
	KafkaBrokers           []string `json:"kafka_brokers"`
	TopicResponse          string   `json:"topic_response"`
	TopicReceivedAddCensor string   `json:"topic_received_add_censor"`
}

// Комментарий к публикации
type Comment struct {
	Id          int    `json:"id"`
	IdNews      int    `json:"id_news"`
	CommentTime int64  `json:"comment_time"`
	UserName    string `json:"user_name"`
	Content     string `json:"content"`
}

// Cтруктура для передачи данных в api-gateway
type SendMessServiceComments struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    int       `json:"status"`
	TypeQuery string    `json:"type_query"`
	IdNews    int       `json:"id_news"`
	Comments  []Comment `json:"comments"`
}

// Cтруктура для получения данных от api-gateway
type GetMessServiceComments struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      int    `json:"status"`
	TypeQuery   string `json:"type_query"`
	IdNews      int    `json:"id_news"`
	CommentTime int64  `json:"comment_time"`
	UserName    string `json:"user_name"`
	Content     string `json:"content"`
}

const (
	VAR_SERVICENAMAE    = "CENSORNAMESERVISE"
	VAR_CONFIGOFFENSIVE = "./configOffensive.json"
	VAR_CONFIGKAFKA     = "./configKafka.json"
	VAR_LOG             = "./logs.json"
)

func main() {

	fmt.Println("service-censor:", logger.GetServiceName(VAR_SERVICENAMAE))
	fmt.Println("service-censor:", logger.GetLocalIP())

	//==============================================
	//Logger
	//==============================================
	logs, err := logger.NewLogger(VAR_LOG, 50, VAR_SERVICENAMAE)
	if err != nil {
		fmt.Printf("Error creating logger: %v", err)
	}
	defer logs.Close()

	//==============================================
	//Kafka
	//==============================================
	// чтение и раскодирование файла конфигурации
	data, err := ioutil.ReadFile(VAR_CONFIGKAFKA)
	if err != nil {
		log.Fatal(err)
	}
	var configKafka ConfigKafka
	err = json.Unmarshal(data, &configKafka)
	if err != nil {
		log.Fatal(err)
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

	// канал для потребления сообщений
	responseCh, err := kafkaConsumer.Consume(configKafka.TopicResponse, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Failed to consume partition: %v", err)
	}

	//==============================================
	//Censor
	//==============================================
	// Создаём объект сервера.
	var srv server

	// Инициализируем пакет
	c, err := censor.NewCensor(VAR_CONFIGOFFENSIVE) // Замените на актуальный путь
	if err != nil {
		log.Fatalf("Ошибка при создании Censor: %v", err)
	}
	srv.censor = c

	errorChannel := make(chan error, 100)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel when we are finished consuming integers

	var wg sync.WaitGroup
	wg.Add(2)

	// обрабатываем данные полученные из kafak
	go checkCensor(ctx, srv.censor, kafkaProducer, configKafka, responseCh, errorChannel)
	// выводим ошибки
	go handleErrors(ctx, errorChannel, logs)

	wg.Wait()
	//select {}
}

func checkCensor(ctx context.Context, censor *censor.Censor, producer *kafka.Producer, config ConfigKafka, responseCh <-chan *sarama.ConsumerMessage, errs chan<- error) {
	for msg := range responseCh {
		select {
		case <-ctx.Done():
			return
		default:

			// Обработка входящего сообщения
			var receivedMessage GetMessServiceComments
			err := json.Unmarshal(msg.Value, &receivedMessage)
			if err != nil {
				errs <- err
			}

			//пишем запрос данных в лог
			var errMsg error = receivedMessage
			errs <- errMsg

			responseMessage := SendMessServiceComments{
				ID:        receivedMessage.ID,
				Name:      logger.GetServiceName(VAR_SERVICENAMAE),
				TypeQuery: receivedMessage.TypeQuery,
				Status:    0,
				IdNews:    receivedMessage.IdNews,
				Comments:  nil,
			}

			switch receivedMessage.TypeQuery {
			case "CommentNew":

				// Проверка комментария
				if !censor.IsOffensive(receivedMessage.UserName) && !censor.IsOffensive(receivedMessage.Content) {
					responseMessage.Status = 192
				} else {
					responseMessage.Status = 0
					errs <- fmt.Errorf("comment not valid")
				}

				bytesMessage, err := json.Marshal(responseMessage)
				if err != nil {
					errs <- err
				}

				err = producer.SendMessage(config.TopicReceivedAddCensor, responseMessage.ID, bytesMessage)
				if err != nil {
					errs <- err
				}
			}

		}
	}
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

// Метод для реализации интерфейса error
func (g GetMessServiceComments) Error() string {
	jsonData, err := json.Marshal(g)
	if err != nil {
		return "error convert to JSON"
	}
	return string(jsonData)
}
