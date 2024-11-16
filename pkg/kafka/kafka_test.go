package kafka

import (
	"fmt"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Producer
type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) SendMessage(topic string, key string, value []byte) error {
	args := m.Called(topic, key, value)
	return args.Error(0)
}

func (m *MockProducer) Close() error {
	return m.Called().Error(0)
}

// Test Producer
func TestSendMessage(t *testing.T) {
	mockProd := new(MockProducer)

	// Настройка ожиданий
	mockProd.On("SendMessage", "test_topic", "test_key", []byte("test_value")).Return(nil)

	// Выполнение теста
	err := mockProd.SendMessage("test_topic", "test_key", []byte("test_value"))

	// Проверка
	assert.NoError(t, err)
	mockProd.AssertExpectations(t) // Проверка выполнения всех ожидаемых вызовов
}

func TestSendMessageError(t *testing.T) {
	mockProd := new(MockProducer)

	// Настройка ожиданий на ошибку
	mockProd.On("SendMessage", "test_topic", "test_key", []byte("test_value")).Return(fmt.Errorf("some error"))

	// Выполнение теста
	err := mockProd.SendMessage("test_topic", "test_key", []byte("test_value"))

	// Проверка
	assert.Error(t, err)
	assert.EqualError(t, err, "some error")
	mockProd.AssertExpectations(t) // Проверка выполнения всех ожидаемых вызовов
}

func TestProducerClose(t *testing.T) {
	mockProd := new(MockProducer)

	// Настройка ожиданий
	mockProd.On("Close").Return(nil)

	// Выполнение теста
	err := mockProd.Close()

	// Проверка
	assert.NoError(t, err)
	mockProd.AssertExpectations(t) // Проверка выполнения всех ожидаемых вызовов
}
