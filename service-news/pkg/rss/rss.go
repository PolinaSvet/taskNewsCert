package rss

import (
	"context"
	"strings"
	"time"

	"news-kafka/service-news/pkg/storage"

	"github.com/mmcdole/gofeed"
)

func GetNewsFromRss(url string, rubric string, image string) ([]storage.News, error) {

	// Создаем новый парсер RSS
	feedParser := gofeed.NewParser()

	// Устанавливаем таймаут для запроса RSS
	timeout := 3 * time.Second

	// Парсим RSS-канал с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	//Обрабатываем ошибки
	feed, err := feedParser.ParseURLWithContext(url, ctx)
	if err != nil {
		return nil, err
	}

	// Выводим элементы RSS-канала
	var data []storage.News
	for _, item := range feed.Items {
		var dataItem storage.News
		dataItem.Title = item.Title
		dataItem.Content = item.Description
		dataItem.Link = item.Link
		dataItem.Rubric = rubric
		dataItem.LinkTitle = feed.Title

		//date public
		var itemPublicTime = item.Published
		itemPublicTime = strings.ReplaceAll(itemPublicTime, ",", "")
		t, err := time.Parse("Mon 2 Jan 2006 15:04:05 -0700", itemPublicTime)
		if err != nil {
			t, err = time.Parse("Mon 2 Jan 2006 15:04:05 GMT", itemPublicTime)
		}
		if err == nil {
			dataItem.PublicTime = t.Unix()
		}

		// Получаем изображение из RSS-канала
		dataItem.ImageLink = image
		if len(item.Enclosures) > 0 {
			for _, enclosure := range item.Enclosures {
				if enclosure.Type == "image/jpeg" || enclosure.Type == "image/png" {
					dataItem.ImageLink = enclosure.URL
					break
				}
			}
		}

		data = append(data, dataItem)
	}
	return data, nil

}
