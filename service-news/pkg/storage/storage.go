package storage

// Публикация, получаемая из RSS.
type News struct {
	Id         int
	Title      string `json:"title"`
	Content    string `json:"content"`
	PublicTime int64  `json:"public_time"`
	ImageLink  string `json:"image_link"`
	Rubric     string `json:"rubric"`
	Link       string `json:"link"`
	LinkTitle  string `json:"link_title"`
}

// Пагинация.
type Paginate struct {
	PageCurr       int `json:"page_curr"`        //Номер текущей страницы
	PageCount      int `json:"page_count"`       //Количество страниц
	PageCountList  int `json:"page_count_list"`  //Количество новостей на странице
	PageCountTotal int `json:"page_count_total"` //Количество всего новостей
}

// Interface задаёт контракт на работу с БД.
type Interface interface {
	GetInform() string
	Close()

	News(rubric string, countNews int, filter string, pageCurr int) ([]News, Paginate, error) // News возвращает последние новости из БД.
	NewsOne(id int) (News, error)                                                             // News возвращает новость по ID.
	AddNew(news []News) error                                                                 // Добавляем новость в БД.
}
