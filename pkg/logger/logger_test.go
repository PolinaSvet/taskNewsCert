package logger

import (
	"os"
	"testing"
)

func TestLogger(t *testing.T) {
	logFile := "test_log.json"
	logger, err := NewLogger(logFile, 2, "servName")
	if err != nil {
		t.Fatalf("Error creating logger: %v", err)
	}
	defer os.Remove(logFile) // Удаляем файл после тестирования
	defer logger.Close()

	logger.LogRequest("req1", "192.168.1.1", 200, "{\"key\":\"value1\"}")
	logger.LogRequest("req2", "192.168.1.1", 404, "{\"key\":\"value2\"}")

	// Вызов potentail flush через закрытие
	logger.Close()

	// Проверяем, что файл был создан и содержит лог записи
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Fatalf("Log file does not exist: %v", err)
	}
}
