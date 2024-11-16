package rss

import (
	"news-kafka/service-news/pkg/storage"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNewsFromRss(t *testing.T) {
	type args struct {
		url    string
		rubric string
		image  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "RSS feed: Good test (valid url)",
			args: args{
				url:    "http://news.mail.ru/rss/sport/91/",
				rubric: "Sport",
				image:  "database/image/imageSport.png",
			},
			wantErr: false,
		},
		{
			name: "RSS feed: Error test (invalid url)",
			args: args{
				url:    "http://news.mail.ru/rss/sport/91/1",
				rubric: "Sport",
				image:  "database/image/imageSport.png",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetNewsFromRss(tt.args.url, tt.args.rubric, tt.args.image)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, []storage.News{}, got)
			}
		})
	}
}
