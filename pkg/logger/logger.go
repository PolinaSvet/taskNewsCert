// Package logger - Пакет для логирования.

package logger

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RequestLog структура для лога запроса
type RequestLog struct {
	Timestamp   time.Time `json:"timestamp"`
	ServiceID   string    `json:"service_id"`
	RequestID   string    `json:"request_id"`
	RemoteAddr  string    `json:"remote_addr"`
	StatusCode  int       `json:"status_code"`
	DataRequest string    `json:"data_request"`
}

// Logger для записи запросов
type Logger struct {
	mu         sync.Mutex
	logs       []RequestLog
	bufferSize int
	file       *os.File
	serverName string
}

// NewLogger создает новый экземпляр логгера
func NewLogger(filePath string, bufferSize int, envName string) (*Logger, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		logs:       make([]RequestLog, 0, bufferSize),
		bufferSize: bufferSize,
		file:       file,
		serverName: GetServiceName(envName),
	}, nil
}

// LogRequest логирует запрос
func (l *Logger) LogRequest(requestID, remoteAddr string, statusCode int, dataRequest string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	logEntry := RequestLog{
		Timestamp:   time.Now(),
		ServiceID:   l.serverName, //GetServiceName(),
		RequestID:   requestID,
		RemoteAddr:  remoteAddr,
		StatusCode:  statusCode,
		DataRequest: dataRequest,
	}

	l.logs = append(l.logs, logEntry)

	// Проверяем, если буфер заполнен
	if len(l.logs) >= l.bufferSize {
		l.flush()
	}
}

// flush записывает логи в файл
func (l *Logger) flush() {
	if len(l.logs) == 0 {
		return
	}

	// Преобразуем логи в JSON и записываем в файл
	encoder := json.NewEncoder(l.file)
	for _, log := range l.logs {
		if err := encoder.Encode(log); err != nil {
			// Здесь можно обработать ошибку, например, записать в stderr
			continue
		}
	}

	// Очищаем буфер
	l.logs = l.logs[:0]
}

// Close закрывает логгер и записывает оставшиеся логи
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.flush() // Записываем оставшиеся записи
	return l.file.Close()
}

// GetRequestId возвращает id запроса
func GetRequestId() string {
	return uuid.New().String()
}

// GetServiceName возвращает имя сервиса
func GetServiceName(envname string) string {
	//Переменные окружения
	//os.Setenv("NEWSNAMESERVISE", "service-news-001")
	//return os.Getenv("NEWSNAMESERVISE")
	return os.Getenv(envname)
}

// GetLocalIP возвращает локальный IP-адрес в виде строки
func GetLocalIP() string {
	// Получаем список всех адаптеров
	interfaces, err := net.Interfaces()
	if err != nil {
		return fmt.Sprintf("%v", err)
	}

	for _, iface := range interfaces {
		// Игнорируем отключенные интерфейсы и петлевые интерфейсы
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Получаем адреса интерфейса
		addrs, err := iface.Addrs()
		if err != nil {
			return fmt.Sprintf("%v", err)
		}

		for _, addr := range addrs {
			var ip net.IP
			// Получаем IP адрес
			switch v := addr.(type) {
			case *net.IPAddr:
				ip = v.IP
			case *net.IPNet:
				ip = v.IP
			}

			if ip != nil && ip.To4() != nil { // Проверка на IPv4
				return ip.String()
			}
		}
	}
	return fmt.Sprintf("%v", "no valid IP address found")
}
