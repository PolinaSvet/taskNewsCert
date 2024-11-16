package postgres

import (
	"context"
	"fmt"
	"news-kafka/service-comments/pkg/storage"

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

// CommentsByIdNews возвращает комментарии к статье из БД.
func (s *Store) CommentsByIdNews(idNews int) ([]storage.Comment, error) {

	rows, err := s.db.Query(context.Background(), `
	 SELECT id, id_news,comment_time, user_name, content 
	 FROM comments
	 WHERE id_news = $1
	 ORDER BY comment_time DESC
	`, idNews)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	var comments []storage.Comment
	for rows.Next() {
		var p storage.Comment
		err = rows.Scan(
			&p.Id,
			&p.IdNews,
			&p.CommentTime,
			&p.UserName,
			&p.Content,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		comments = append(comments, p)
	}

	return comments, nil
}

// CommentNew добавляем комментарий в БД.
func (s *Store) CommentNew(comment storage.Comment) (int, error) {

	var id_rec int
	err := s.db.QueryRow(context.Background(), `
		INSERT INTO comments(id_news, comment_time, user_name, content)
		VALUES ($1, $2, $3, $4) RETURNING id;`,
		comment.IdNews,
		comment.CommentTime,
		comment.UserName,
		comment.Content,
	).Scan(&id_rec)

	if err != nil {
		return 0, fmt.Errorf("failed to insert row: %w", err)
	}

	return id_rec, nil
}
