package censor

import (
	"testing"
)

// Создадим структуру для тестирования Censor с использованием брутального списка слов
func NewCensorFromWords(words []string) *Censor {
	return &Censor{offensiveWords: words}
}

// TestIsOffensive проверяет метод IsOffensive на наличие оскорбительных слов.
func TestIsOffensive(t *testing.T) {
	// Создаем Censor с оскорбительными словами
	censor := NewCensorFromWords([]string{"bad", "gross", "awful"})

	tests := []struct {
		input    string
		expected bool
	}{
		{"this is a bad example", true},
		{"this is a clean example", false},
		{"gross behavior", true},
		{"an awful situation", true},
		{"all is fine", false},
	}

	for _, test := range tests {
		result := censor.IsOffensive(test.input)
		if result != test.expected {
			t.Fatalf("for input '%s': expected %v, got %v", test.input, test.expected, result)
		}
	}
}

// TestIsOffensive_EmptyWords проверяет, что пустой список оскорбительных слов дает false.
func TestIsOffensive_EmptyWords(t *testing.T) {
	censor := NewCensorFromWords([]string{})

	tests := []struct {
		input    string
		expected bool
	}{
		{"this is a bad example", false},
		{"gross behavior", false},
		{"clean text", false},
	}

	for _, test := range tests {
		result := censor.IsOffensive(test.input)
		if result != test.expected {
			t.Fatalf("for input '%s': expected %v, got %v", test.input, test.expected, result)
		}
	}
}

// TestIsOffensive_SimilarWords проверяет некорректные случаи с похожими словами.
func TestIsOffensive_SimilarWords(t *testing.T) {
	censor := NewCensorFromWords([]string{"bad"})

	tests := []struct {
		input    string
		expected bool
	}{
		{"this is a bad example", true},
		{"this is a badd example", true},
		{"this is a BAD example", true}, // тест на регистр
	}

	for _, test := range tests {
		result := censor.IsOffensive(test.input)
		if result != test.expected {
			t.Fatalf("for input '%s': expected %v, got %v", test.input, test.expected, result)
		}
	}
}
