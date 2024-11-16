package postgres

import (
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/mock"
)

// MockStore — это мок-реализация интерфейса Storage
type MockStore struct {
	mock.Mock
}

func (m *MockStore) GetInform() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockStore) Close() {
	m.Called()
}

func TestStore_GetInform(t *testing.T) {
	// Создание mock хранилища
	mockStore := new(MockStore)

	mockStore.On("GetInform").Return("PostgreSQL")

	result := mockStore.GetInform()
	assert.Equal(t, "PostgreSQL", result)

	// Проверка вызова метода
	mockStore.AssertExpectations(t)
}

func TestStore_Close(t *testing.T) {
	// Создание mock хранилища
	mockStore := new(MockStore)

	mockStore.On("Close").Return()

	mockStore.Close()

	// Проверка вызова метода
	mockStore.AssertExpectations(t)
}
