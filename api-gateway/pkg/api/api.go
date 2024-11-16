package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	//"news-kafka/api-gateway/pkg/kafka"
	//"news-kafka/api-gateway/pkg/logger"
	"strconv"
	"sync"
	"text/template"
	"time"

	"github.com/IBM/sarama"
	"github.com/PolinaSvet/kafka"
	"github.com/PolinaSvet/logger"
	"github.com/gorilla/mux"
)

// service-news
// Публикация, получаемая из RSS.
type News struct {
	Id         int
	Title      string `json:"title"`
	Content    string `json:"content"`
	PublicTime int64  `json:"public_time"`
	ImageLink  string `json:"image_link"`
	Rubric     string `json:"rubric"`
	Link       string `json:"link"`
	LinkTitle  string `json:"link_title"`
}

// Пагинация.
type Paginate struct {
	PageCurr       int `json:"page_curr"`        //Номер текущей страницы
	PageCount      int `json:"page_count"`       //Количество страниц
	PageCountList  int `json:"page_count_list"`  //Количество новостей на странице
	PageCountTotal int `json:"page_count_total"` //Количество всего новостей
}

// Cтруктура для отправки данных в service
type SendMessServiceNews struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    int    `json:"status"`
	TypeQuery string `json:"type_query"`
	Rubric    string `json:"rubric"`
	CountNews int    `json:"count_news"`
	Filter    string `json:"filter"`
	Page      int    `json:"page"`
	IdNews    int    `json:"id_news"`
}

// Cтруктура для получения данных от service
type GetMessServiceNews struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Status    int      `json:"status"`
	TypeQuery string   `json:"type_query"`
	News      []News   `json:"news"`
	Paginate  Paginate `json:"paginate"`
	IdNews    int      `json:"id_news"`
}

// service-comments
// Комментарий к публикации
type Comment struct {
	Id          int    `json:"id"`
	IdNews      int    `json:"id_news"`
	CommentTime int64  `json:"comment_time"`
	UserName    string `json:"user_name"`
	Content     string `json:"content"`
}

// Cтруктура для передачи данных в service
type SendMessServiceComments struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      int    `json:"status"`
	TypeQuery   string `json:"type_query"`
	IdNews      int    `json:"id_news"`
	CommentTime int64  `json:"comment_time"`
	UserName    string `json:"user_name"`
	Content     string `json:"content"`
}

// Cтруктура для получения данных от service
type GetMessServiceComments struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    int       `json:"status"`
	TypeQuery string    `json:"type_query"`
	IdNews    int       `json:"id_news"`
	Comments  []Comment `json:"comments"`
}

// Программный интерфейс сервера GoNews
type API struct {
	producer              *kafka.Producer
	consumer              *kafka.Consumer
	configKafka           *ConfigKafka
	responseNewsCh        <-chan *sarama.ConsumerMessage
	responseOneNewsCh     <-chan *sarama.ConsumerMessage
	responseCommentsCh    <-chan *sarama.ConsumerMessage
	responseAddCommentsCh <-chan *sarama.ConsumerMessage
	responseCensorCh      <-chan *sarama.ConsumerMessage
	router                *mux.Router
	errorChannel          chan<- error
	serviveName           string
}

type ApiChannels struct {
	ResponseNewsCh        <-chan *sarama.ConsumerMessage
	ResponseOneNewsCh     <-chan *sarama.ConsumerMessage
	ResponseCommentsCh    <-chan *sarama.ConsumerMessage
	ResponseAddCommentsCh <-chan *sarama.ConsumerMessage
	ResponseCensorCh      <-chan *sarama.ConsumerMessage
	ErrorChannel          chan<- error
}

// Конструктор объекта API
func New(producer *kafka.Producer, consumer *kafka.Consumer, configKafka *ConfigKafka, apiChannels ApiChannels, serviveName string) *API {
	api := API{
		producer:              producer,
		consumer:              consumer,
		configKafka:           configKafka,
		responseNewsCh:        apiChannels.ResponseNewsCh,
		responseOneNewsCh:     apiChannels.ResponseOneNewsCh,
		responseCommentsCh:    apiChannels.ResponseCommentsCh,
		responseAddCommentsCh: apiChannels.ResponseAddCommentsCh,
		responseCensorCh:      apiChannels.ResponseCensorCh,
		errorChannel:          apiChannels.ErrorChannel,
		serviveName:           serviveName,
	}
	api.router = mux.NewRouter()
	// Добавляем middleware для request_id
	api.router.Use(RequestIDMiddleware)
	// Добавляем middleware для считывания тела запроса
	api.router.Use(ReadBodyMiddleware)
	// Добавляем middleware для логирования
	api.router.Use(func(next http.Handler) http.Handler { return LoggingMiddleware(next, api.errorChannel) })
	// Добавляем middleware для логирования ошибок сервера
	api.router.Use(func(next http.Handler) http.Handler { return ErrorHandlerMiddleware(next, api.errorChannel) })

	api.endpoints()
	return &api
}

// Получение маршрутизатора запросов.
// Требуется для передачи маршрутизатора веб-серверу.
func (api *API) Router() *mux.Router {
	return api.router
}

type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *LoggingResponseWriter) Write(b []byte) (int, error) {
	return lrw.ResponseWriter.Write(b)
}

// Middleware(1) для добавления сквозного идентификатора запроса
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Получаем ID из запроса, если он есть
		requestID := r.URL.Query().Get("request_id")

		// Если ID не передан, генерируем новый
		if requestID == "" {
			requestID = logger.GetRequestId()
		}

		// Добавляем ID в контекст
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		r = r.WithContext(ctx)

		// Вызываем следующий обработчик
		next.ServeHTTP(w, r)
	})
}

// Middleware(2) для считывания тела запроса.
func ReadBodyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Считываем тело запроса
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "ошибка чтения тела запроса", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		// Сбрасываем r.Body обратно
		r.Body = ioutil.NopCloser(bytes.NewReader(body))

		// Устанавливаем тело запроса в контекст
		r = r.WithContext(context.WithValue(r.Context(), "requestBody", body))

		// Вызываем следующий обработчик
		next.ServeHTTP(w, r)
	})
}

// Middleware(3) для логирования запросов
func LoggingMiddleware(next http.Handler, errorChan chan<- error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем тело запроса из контекста
		body, ok := r.Context().Value("requestBody").([]byte)
		if !ok {
			http.Error(w, "ошибка получения тела запроса", http.StatusInternalServerError)
			return
		}

		// Создаем кастомный ResponseWriter
		lrw := &LoggingResponseWriter{ResponseWriter: w}

		// Вызываем следующий обработчик
		next.ServeHTTP(lrw, r)

		requestID := r.Context().Value("request_id").(string)

		// Получаем HTTP-код ответа из ответа
		statusCode := lrw.statusCode

		errorChan <- fmt.Errorf("RequestID:%v, RemoteAddr:%v, StatusCode:%v, %v[%v]", requestID, r.RemoteAddr, statusCode, r.URL.Path, string(body))

	})
}

// Middleware(4) для для обработки ошибок
func ErrorHandlerMiddleware(next http.Handler, errorChan chan<- error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Создание кастомного ResponseWriter для перехвата записи ошибок
		crw := &customResponseWriter{ResponseWriter: w}
		next.ServeHTTP(crw, r)

		// Проверка на наличие ошибок
		if crw.statusCode >= 400 {
			logMessage := crw.errorMessage
			if logMessage != "" {
				errorChan <- fmt.Errorf("%v, StatusCode:%v", logMessage, crw.statusCode)
			}
		}
	})
}

// customResponseWriter для перехвата записанных данных
type customResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	errorMessage string
}

// WriteHeader для перехвата кода статуса и сообщения об ошибке
func (crw *customResponseWriter) WriteHeader(code int) {
	crw.statusCode = code
	if code >= 400 { // Если статус ошибки
		crw.errorMessage = "HTTP error occurred. Code: " + http.StatusText(code)
	}
	crw.ResponseWriter.WriteHeader(code)
}

// Регистрация обработчиков API.
func (api *API) endpoints() {

	api.router.HandleFunc("/", api.templateHandler).Methods(http.MethodGet, http.MethodOptions)
	api.router.HandleFunc("/news/{rubric}/{countNews}", api.newsHandler).Methods(http.MethodGet, http.MethodOptions)
	api.router.HandleFunc("/newsDetailed", api.newsDetailedHandler).Methods(http.MethodGet, http.MethodOptions)
	api.router.HandleFunc("/comments", api.addCommentsHandler).Methods(http.MethodPost, http.MethodOptions)

	api.router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./ui"))))
}

// Базовый маршрут.
func (api *API) templateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFiles("ui/html/base.html", "ui/html/routes.html"))

	// Отправляем HTML страницу с данными
	if err := tmpl.ExecuteTemplate(w, "base", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

// Получение всех новостей.
func (api *API) newsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodOptions {
		return
	}

	vars := mux.Vars(r)
	rubric := vars["rubric"]
	countNewsStr := vars["countNews"]
	countNews, err := strconv.Atoi(countNewsStr)
	if err != nil {
		http.Error(w, "Invalid count parameter", http.StatusBadRequest)
		return
	}

	// Получение параметров из запроса
	filter := r.URL.Query().Get("filter")
	if filter == "undefined" {
		filter = ""
	}

	pageStr := r.URL.Query().Get("page")
	pageCurr := 1
	if pageStr != "" {
		pageCurr, err = strconv.Atoi(pageStr)
		if err != nil {
			http.Error(w, "Invalid page parameter", http.StatusBadRequest)
			return
		}
	}
	request_id := r.Context().Value("request_id").(string)

	sendMessage := SendMessServiceNews{
		ID:        request_id,
		Name:      api.serviveName,
		Status:    192,
		TypeQuery: "News",
		Rubric:    rubric,
		CountNews: countNews,
		Filter:    filter,
		Page:      pageCurr,
	}

	bytesMessage, err := json.Marshal(sendMessage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Десериализация сообщения в структуру
	var serviceNews GetMessServiceNews
	serviceNews.Status = 0

	// Запуск обработки запроса в отдельной гоурутине
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(request_id string, bytesMessage []byte) {
		defer wg.Done()

		// Отправка сообщения в Kafka
		err = api.producer.SendMessage(api.configKafka.TopicResponseNews, request_id, bytesMessage)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Чтение сообщений Kafka
		select {
		case msg, ok := <-api.responseNewsCh:

			if !ok {
				api.errorChannel <- fmt.Errorf("response channel closed")
				return
			}

			if err := json.Unmarshal(msg.Value, &serviceNews); err != nil {
				api.errorChannel <- err
				return
			}

			if serviceNews.ID != request_id || serviceNews.TypeQuery != "News" {
				api.errorChannel <- fmt.Errorf("error ID and Type message")
				return
			}

		case <-time.After(3 * time.Second):
			api.errorChannel <- fmt.Errorf("timeout waiting for response")
			return
		}
	}(request_id, bytesMessage)

	// Ожидание завершения обработки запроса в гоурутине
	wg.Wait()

	if serviceNews.Status == 192 {
		// Формирование ответа JSON
		response := map[string]interface{}{
			"news":     serviceNews.News,
			"paginate": serviceNews.Paginate,
		}

		// Отправка ответа клиенту
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		http.Error(w, "server error", http.StatusInternalServerError)
	}
}

// Получение всех comments by news.
func (api *API) newsDetailedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodOptions {
		return
	}

	// Получение параметров из запроса
	id_news_str := r.URL.Query().Get("id_news")
	var err error
	id_news := 0
	if id_news_str != "" {
		id_news, err = strconv.Atoi(id_news_str)
		if err != nil {
			http.Error(w, "Invalid id_news parameter", http.StatusBadRequest)
			return
		}
	}
	request_id := r.Context().Value("request_id").(string)

	sendMessage := SendMessServiceComments{
		ID:          request_id,
		Name:        api.serviveName,
		Status:      192,
		TypeQuery:   "CommentsByIdNews",
		IdNews:      id_news,
		CommentTime: 0,
		UserName:    "",
		Content:     "",
	}

	bytesMessage, err := json.Marshal(sendMessage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendMessageNews := SendMessServiceNews{
		ID:        request_id,
		Name:      api.serviveName,
		Status:    192,
		TypeQuery: "OneNews",
		IdNews:    id_news,
		Rubric:    "",
		CountNews: 1,
		Filter:    "",
		Page:      1,
	}

	bytesMessageNews, err := json.Marshal(sendMessageNews)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправка сообщения в Kafka Comments
	err = api.producer.SendMessage(api.configKafka.TopicResponseComments, request_id, bytesMessage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправка сообщения в Kafka News
	err = api.producer.SendMessage(api.configKafka.TopicResponseNews, request_id, bytesMessageNews)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Десериализация сообщения в структуру
	var serviceComments GetMessServiceComments
	var serviceNews GetMessServiceNews

	serviceNews.Status = 0
	serviceComments.Status = 0

	// Запуск обработки запроса в отдельной гоурутине
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(request_id string, bytesMessage []byte) {
		defer wg.Done()

		select {
		case msg, ok := <-api.responseCommentsCh:

			if !ok {
				api.errorChannel <- fmt.Errorf("response channel closed")
				return
			}

			if err := json.Unmarshal(msg.Value, &serviceComments); err != nil {
				api.errorChannel <- err
				return
			}

			if serviceComments.ID != request_id || serviceComments.TypeQuery != "CommentsByIdNews" {
				api.errorChannel <- fmt.Errorf("error ID and Type message")
				return
			}

		case <-time.After(3 * time.Second):
			api.errorChannel <- fmt.Errorf("timeout waiting for response")
			return
		}
	}(request_id, bytesMessage)

	wg.Add(1)
	go func(request_id string, bytesMessageNews []byte) {
		defer wg.Done()

		select {
		case msg, ok := <-api.responseOneNewsCh:

			if !ok {
				api.errorChannel <- fmt.Errorf("response channel closed")
				return
			}

			if err := json.Unmarshal(msg.Value, &serviceNews); err != nil {
				api.errorChannel <- err
				return
			}

			if serviceNews.ID != request_id || serviceNews.TypeQuery != "OneNews" {
				api.errorChannel <- fmt.Errorf("error ID and Type message")
				return
			}

		case <-time.After(3 * time.Second):
			api.errorChannel <- fmt.Errorf("timeout waiting for response")
			return
		}
	}(request_id, bytesMessageNews)

	wg.Wait()

	if serviceNews.Status == 192 || serviceComments.Status == 192 {
		// Формирование ответа JSON
		response := map[string]interface{}{
			"comments": serviceComments.Comments,
			"news":     serviceNews.News,
			"idNews":   id_news,
		}

		// Отправка ответа клиенту
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		http.Error(w, "server error", http.StatusInternalServerError)
	}

}

// Добавление comments.
func (api *API) addCommentsHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodOptions {
		return
	}

	// Получение параметров из запроса
	id_news_str := r.URL.Query().Get("id_news")
	var err error
	id_news := 0
	if id_news_str != "" {
		id_news, err = strconv.Atoi(id_news_str)
		if err != nil {
			http.Error(w, "Invalid id_news parameter", http.StatusBadRequest)
			return
		}
	}
	request_id := r.Context().Value("request_id").(string)

	var comment Comment
	err = json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendMessage := SendMessServiceComments{
		ID:          request_id,
		Name:        api.serviveName,
		Status:      192,
		TypeQuery:   "CommentNew",
		IdNews:      id_news,
		CommentTime: comment.CommentTime,
		UserName:    comment.UserName,
		Content:     comment.Content,
	}

	bytesMessage, err := json.Marshal(sendMessage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var serviceComments GetMessServiceComments

	// 1. Отправка сообщения в Kafka Censor
	err = api.producer.SendMessage(api.configKafka.TopicResponseCensor, request_id, bytesMessage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	select {
	case msg, ok := <-api.responseCensorCh:

		if !ok {
			api.errorChannel <- fmt.Errorf("response channel closed")
			return
		}

		if err := json.Unmarshal(msg.Value, &serviceComments); err != nil {
			api.errorChannel <- err
			return
		}

		if serviceComments.ID != request_id || serviceComments.TypeQuery != "CommentNew" {
			api.errorChannel <- fmt.Errorf("error ID and Type message")
			return
		}

	case <-time.After(3 * time.Second):
		api.errorChannel <- fmt.Errorf("timeout waiting for response")
		return
	}

	if serviceComments.Status != 192 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 2. Отправка сообщения в Kafka Comments
	err = api.producer.SendMessage(api.configKafka.TopicResponseComments, request_id, bytesMessage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	select {
	case msg, ok := <-api.responseAddCommentsCh:

		if !ok {
			api.errorChannel <- fmt.Errorf("response channel closed")
			return
		}

		if err := json.Unmarshal(msg.Value, &serviceComments); err != nil {
			api.errorChannel <- err
			return
		}

		if serviceComments.ID != request_id || serviceComments.TypeQuery != "CommentNew" {
			api.errorChannel <- fmt.Errorf("error ID and Type message")
			return
		}

	case <-time.After(3 * time.Second):
		api.errorChannel <- fmt.Errorf("timeout waiting for response")
		return
	}

	// Отправка ответа клиенту
	if serviceComments.Status == 192 {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

}

// Config - структура для хранения конфигурации
type ConfigKafka struct {
	KafkaBrokers             []string `json:"kafka_brokers"`
	TopicResponseNews        string   `json:"topic_response_news"`
	TopicReceivedNews        string   `json:"topic_received_news"`
	TopicReceivedOneNews     string   `json:"topic_received_one_news"`
	TopicResponseComments    string   `json:"topic_response_comments"`
	TopicReceivedComments    string   `json:"topic_received_comments"`
	TopicReceivedAddComments string   `json:"topic_received_add_comments"`
	TopicResponseCensor      string   `json:"topic_response_censor"`
	TopicReceivedAddCensor   string   `json:"topic_received_add_censor"`
}

// readConfig - функция для чтения конфигурации из файла
func ReadConfig(filePath string) (*ConfigKafka, error) {
	// Чтение содержимого файла
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Декодирование JSON данных
	var config ConfigKafka
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config data: %w", err)
	}

	return &config, nil
}
