-- INITIALIZE tg_topics table
insert into tg_topics ( message_thread_id, tag_id, name, icon_custom_emoji_id, created) values
    (5, 1, 'Бхагавад-гита', '5377317729109811382', now()),
    (4, 2, 'Шримад Бхагаватам', '5377317729109811382', now()),
    (21, 8, 'Школа джапа-медитации', '5377317729109811382', now()),
    (22, 2134, 'Шикшаштака', '5377317729109811382', now()),
    (23, 91, 'Шаранагати', '5377317729109811382', now()),
    (24, 1447, 'Ведическая психология', '5377317729109811382', now()),
    (25, 2055, 'Ступени бхакти', '5377317729109811382', now()),
    (27, 202, 'Нектар преданности', '5377317729109811382', now()),
    (28, 484, 'Нектар наставлений', '5377317729109811382', now()),
    (29, 2020, 'Паломничества', '5377317729109811382', now()),
    (30, 20148, 'СИНДУ', '5377317729109811382', now()),
    (31, 10879, 'Миссия Господа Чайтаньи', '5377317729109811382', now()),
    (32, 244, 'Шрила Прабхупада', '5377317729109811382', now()),
    (33, 1536, 'Мадхурья-кадамбини', '5377317729109811382', now()),
    (34, 1582, 'Обзор Шримад-Бхагаватам', '5377317729109811382', now()),
    (26, 1419, 'Чайтанья Чаритамрита', '5377317729109811382', now());


-- TODO: main goal is to propagate existent media to telegram
-- 1. add tags to media_tag based on scripture_media through manual linking table mentioned below.
-- 2. forbid to update media_tag table from jira until 2025



INSERT INTO media_tag (media_id, tag_id)
WITH ts AS (
	SELECT tag_scripture.*, t.name, tt.name FROM
	(
		VALUES 
			(1,1),			--Бхагавад-гита
			(2,2),			--Шримад Бхагаватам
			(2134,NULL),	--Шикшаштака
			(91,NULL),		--Шаранагати
			(1447,NULL),	--Ведическая психология
			(2055,NULL),	--Ступени бхакти
			(202,4),		--Нектар преданности
			(484,5),		--Нектар наставлений
			(2020,NULL),	--Паломничества
			(20148,NULL),	--СИНДУ
			(10879,NULL),	--Миссия Господа Чайтаньи
			(244,NULL),		--Шрила Прабхупада
			(1536,7),		--Мадхурья-кадамбини
			(1582,NULL),	--Обзор Шримад-Бхагаватам
			(1419,3),		--Чайтанья Чаритамрита
			(8,NULL),		--Ведическая культура
			(2844,NULL)		--Школа джапа-медитации
	) AS tag_scripture(tag_id, scripture_id)
	JOIN tag t ON t.id = tag_scripture.tag_id
	JOIN tg_topics tt ON tt.tag_id = tag_scripture.tag_id
	JOIN scripture s ON s.id = tag_scripture.scripture_id
	WHERE tag_scripture.scripture_id IS NOT NULL
)
SELECT
	DISTINCT 
	m.id media_id, t.id tag_id
--	s.name AS scripture,
--	m.title, 
--	m.file_url,
--	t.name AS tag,
--	EXISTS (SELECT mt.tag_id FROM media_tag mt WHERE mt.media_id = m.id AND mt.tag_id = t.id)
FROM media_scripture ms
JOIN ts ON ms.scripture_id = ts.scripture_id
JOIN media m ON m.id = ms.media_id
JOIN scripture s ON s.id = ms.scripture_id
--JOIN media_tag mt ON mt.tag_id = ts.tag_id
JOIN tag t ON t.id = ts.tag_id
WHERE
	NOT EXISTS (SELECT mt.tag_id FROM media_tag mt WHERE mt.media_id = m.id AND mt.tag_id = t.id)
--	m.file_url = '2021.02.19 - The glory of Advaita Acharya.mp3'
--	m.file_url = '2022.02.22 - Use viral effect to organically spread sankirtana.mp3'
--GROUP BY m.id, m.title, m.file_url
;