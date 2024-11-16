--++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
--1) create tables
--++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

DROP TABLE IF EXISTS comments;

CREATE TABLE comments (
    id BIGSERIAL PRIMARY KEY,
    id_news BIGSERIAL,
    comment_time INTEGER DEFAULT 0,
    user_name TEXT NOT NULL,
    content TEXT NOT NULL
);

-- Заполнение таблицы данными
INSERT INTO comments (id_news, comment_time,user_name,content) VALUES
(1, 1730100873, 'user_name_001', 'content_001'),
(1, 1730100874, 'user_name_002', 'content_002'),
(1, 1730100875, 'user_name_003', 'content_003'),
(2, 1730100876, 'user_name_004', 'content_004'),
(2, 1730100877, 'user_name_005', 'content_005'),
(2, 1730100878, 'user_name_006', 'content_006'),
(3, 1730100879, 'user_name_007', 'content_007');
