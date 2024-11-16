// Package censor - Пакет для проверки на цензуру
package censor

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
)

// Censor предоставляет методы для проверки текста на наличие оскорбительных слов
type Censor struct {
	offensiveWords []string
	//offensiveSymbols map[string][]string
}

// NewCensor создает новый экземпляр Censor и загружает оскорбительные слова из JSON файла
func NewCensor(filePathWords string) (*Censor, error) {
	c := &Censor{}
	err := c.loadOffensiveWords(filePathWords)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// loadOffensiveWords загружает оскорбительные слова из JSON файла
func (c *Censor) loadOffensiveWords(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	var data struct {
		OffensiveWords []string `json:"offensive_words"`
	}

	err = json.Unmarshal(byteValue, &data)
	if err != nil {
		return err
	}

	c.offensiveWords = data.OffensiveWords
	return nil
}

// containsOffensiveWords проверяет наличие оскорбительных слов в тексте
func (c *Censor) containsOffensiveWords(text string) bool {
	// Преобразуем текст в нижний регистр
	lowerText := strings.ToLower(text)

	for _, word := range c.offensiveWords {
		// Преобразуем каждое слово в нижний регистр
		if strings.Contains(lowerText, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

// IsOffensive проверяет текст на наличие оскорбительных слов
func (c *Censor) IsOffensive(text string) bool {
	return c.containsOffensiveWords(text)
}
