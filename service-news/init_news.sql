--++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
--1) create tables
--++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

DROP TABLE IF EXISTS news;
DROP INDEX IF EXISTS rubric_idx;

CREATE TABLE news (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    public_time INTEGER DEFAULT 0,
    image_link TEXT,
    rubric VARCHAR(255) NOT NULL,
    link TEXT NOT NULL UNIQUE,
	link_title TEXT NOT NULL
);

CREATE INDEX rubric_idx ON news (rubric);