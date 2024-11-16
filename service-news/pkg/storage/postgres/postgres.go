package postgres

import (
	"context"
	"fmt"
	"news-kafka/service-news/pkg/storage"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Хранилище данных.
type Store struct {
	db *pgxpool.Pool
}

func (s *Store) GetInform() string {
	return "PostgreSQL"
}

// Конструктор объекта хранилища.
func New(constr string) (*Store, error) {
	db, err := pgxpool.Connect(context.Background(), constr)
	if err != nil {
		return nil, err
	}
	s := Store{
		db: db,
	}

	fmt.Println("Loaded bd: ", s.GetInform())

	return &s, nil
}

func (s *Store) Close() {
	s.db.Close()
}

// News возвращает последние новости из БД.
func (s *Store) News(rubric string, countNews int, filter string, pageCurr int) ([]storage.News, storage.Paginate, error) {
	if countNews <= 0 {
		countNews = 10
	}

	// Получаем общее количество новостей с учетом фильтра
	var totalCount int
	err := s.db.QueryRow(context.Background(), `
	 SELECT COUNT(*) FROM news
	 WHERE rubric LIKE $1 AND title ILIKE $2
	`, rubric, "%"+filter+"%").Scan(&totalCount)
	if err != nil {
		return nil, storage.Paginate{}, fmt.Errorf("failed to get total count: %w", err)
	}

	// Рассчитываем количество страниц
	pageCount := totalCount / int(countNews)
	if totalCount%int(countNews) != 0 {
		pageCount++
	}

	// Выполняем запрос с пагинацией
	rows, err := s.db.Query(context.Background(), `
	 SELECT id, title, content, public_time, image_link, rubric, link, link_title 
	 FROM news
	 WHERE rubric LIKE $1 AND title ILIKE $2
	 ORDER BY public_time DESC
	 LIMIT $3 OFFSET $4
	`, rubric, "%"+filter+"%", countNews, (pageCurr-1)*int(countNews))
	if err != nil {
		return nil, storage.Paginate{}, fmt.Errorf("failed to query news: %w", err)
	}
	defer rows.Close()

	// Декодируем результаты
	var news []storage.News
	for rows.Next() {
		var p storage.News
		err = rows.Scan(
			&p.Id,
			&p.Title,
			&p.Content,
			&p.PublicTime,
			&p.ImageLink,
			&p.Rubric,
			&p.Link,
			&p.LinkTitle,
		)
		if err != nil {
			return nil, storage.Paginate{}, fmt.Errorf("failed to scan news row: %w", err)
		}
		news = append(news, p)
	}
	if err := rows.Err(); err != nil {
		return nil, storage.Paginate{}, fmt.Errorf("failed to iterate news rows: %w", err)
	}

	// Возвращаем новости и объект пагинации
	return news, storage.Paginate{
		PageCurr:       pageCurr,
		PageCount:      pageCount,
		PageCountList:  int(countNews),
		PageCountTotal: int(totalCount),
	}, nil
}

// News возвращает последние новости из БД.
func (s *Store) NewsOne(id int) (storage.News, error) {

	rows, err := s.db.Query(context.Background(), `
	SELECT id, title, content, public_time, image_link, rubric, link, link_title FROM news
	WHERE id = $1
	`,
		id,
	)
	if err != nil {
		return storage.News{}, err
	}

	// Декодируем результаты
	var news storage.News
	for rows.Next() {
		var p storage.News
		err = rows.Scan(
			&p.Id,
			&p.Title,
			&p.Content,
			&p.PublicTime,
			&p.ImageLink,
			&p.Rubric,
			&p.Link,
			&p.LinkTitle,
		)
		if err != nil {
			return storage.News{}, fmt.Errorf("failed to scan news row: %w", err)
		}
		news = p
	}
	if err := rows.Err(); err != nil {
		return storage.News{}, fmt.Errorf("failed to iterate news rows: %w", err)
	}

	return news, rows.Err()
}

// Добавляем новость в БД.
func (s *Store) AddNew(news []storage.News) error {
	for _, newsRec := range news {
		_, err := s.db.Exec(context.Background(), `
		INSERT INTO news(title, content, public_time, image_link, rubric, link, link_title)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			newsRec.Title,
			newsRec.Content,
			newsRec.PublicTime,
			newsRec.ImageLink,
			newsRec.Rubric,
			newsRec.Link,
			newsRec.LinkTitle,
		)
		if err != nil && !strings.Contains(err.Error(), "news_link_key") {
			return err
		}
	}
	return nil
}
