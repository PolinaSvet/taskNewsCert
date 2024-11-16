package storage

// Комментарий к публикации
type Comment struct {
	Id          int
	IdNews      int    `json:"id_news"`
	CommentTime int64  `json:"comment_time"`
	UserName    string `json:"user_name"`
	Content     string `json:"content"`
}

// Interface задаёт контракт на работу с БД.
type Interface interface {
	GetInform() string
	Close()

	CommentsByIdNews(idNews int) ([]Comment, error) // Возвращает комментарии по статье.
	CommentNew(comment Comment) (int, error)        // Добавляем комментарий в БД.
}
