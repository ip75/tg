-- name: ListAllTopics :many
select tt.*, t.name as tag
from tg_topics tt
join tag t on t.id = tt.tag_id;

-- name: SetRecentUploadTime :exec
update tg_config 
set recent_upload_time = $1
where slug = $2;

-- name: GetRecentUploadTime :one
select recent_upload_time from tg_config where slug = $1;

-- name: ListMediaQueue :many
select
    tq.id cursor,
    tq.media_id,
    m.title,
    m.teaser,
    m.file_url,
    tt.message_thread_id,
    m.occurrence_date,
    m.issue_date,
    m.duration,
    m.size,
    t.id as tag_id,
    t.name as tag
from tg_queue tq
join tag t on t.id = tq.tag_id
join tg_topics tt on tt.id = tq.topic_id
join media m on m.id = tq.media_id
where 
    m.file_url is not null
    and tq.id > $1 -- $1 is the last id in the previous query = cursor
order by m.occurrence_date asc
limit $2;

-- name: RemoveMediaQueue :exec
delete from tg_queue where media_id = $1 and tag_id = $2;

-- name: AddMediaToFailedQueue :exec
WITH topic_lookup AS (
    SELECT id as topic_id FROM tg_topics WHERE message_thread_id = $1 LIMIT 1
)
insert into tg_queue_failed
    (topic_id, media_id, tag_id, error)
select topic_id, $2, $3, $4
from topic_lookup;

-- name: ClearFailedMediaFromQueue :exec
delete from tg_queue where media_id = $1;

-- name: MakeTopicPublished :exec
update tg_topics
set 
    message_thread_id = $1,
    created = now()
where id = $2;

-- name: GetConfig :one
select
    tc.id,
    tc.slug,
    tc.recent_upload_time,
    tc.settings
from tg_config tc
where tc.slug = $1;

-- name: LinkMediaToTelegram :exec
insert into media_data
    (media_id, data_type, value)
values ($1, 'telegram'::media_data_type, $2);


-- name: PopulateMedia :exec
insert into tg_queue (topic_id, media_id, tag_id)
SELECT tt.id, m.id, mt.tag_id
FROM media m 
JOIN media_tag mt ON m.id = mt.media_id 
JOIN tag t ON t.id = mt.tag_id 
JOIN tg_topics tt ON tt.tag_id = mt.tag_id 
left join tg_queue tq on tq.media_id = m.id
WHERE
	m.occurrence_date > $1 
	AND m.file_url IS NOT NULL
	AND tq.media_id IS NULL
ORDER BY m.occurrence_date ASC;

-- name: PopulateMediaWithTagID :exec
insert into tg_queue (topic_id, media_id, tag_id)
SELECT tt.id, m.id, mt.tag_id
FROM media m 
JOIN media_tag mt ON m.id = mt.media_id 
JOIN tag t ON t.id = mt.tag_id 
JOIN tg_topics tt ON tt.tag_id = mt.tag_id 
left join tg_queue tq on tq.media_id = m.id
WHERE
	m.occurrence_date > $1 
    AND t.id = $2
	AND m.file_url IS NOT NULL
	AND tq.media_id IS NULL
ORDER BY m.occurrence_date ASC;

-- name: GetMediaDataTelegram :one
SELECT md.media_id, md.value
FROM media_data md
WHERE
	md.media_id = $1
	AND md.data_type = 'telegram'::media_data_type
limit 1;
