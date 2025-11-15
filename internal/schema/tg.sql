-- topics in tg group
create table tg_topics (
    id bigserial primary key,
    message_thread_id bigint not null,
    tag_id integer references tag(id) not null,
    name text unique not null,
    icon_custom_emoji_id varchar(128),
    created timestamp default NULL,
    CONSTRAINT tg_unique_topic UNIQUE(message_thread_id, tag_id)
);

create table tg_config (
    id bigserial primary key,
    slug text unique not null,
    recent_upload_time timestamp not null,
    settings jsonb not null
);
COMMENT ON TABLE tg_config IS 'Config for sending messages. For future use when will be many chatbots on one server';
COMMENT ON COLUMN tg_config.slug IS 'Unique slug for config. Used for getting config by slug';
COMMENT ON COLUMN tg_config.recent_upload_time IS 'Last time updated topics for telegram, updated when recent audio sent to topic.';
COMMENT ON COLUMN tg_config.settings IS 'Bot settings for sending messages.';

-- function to fill tg_queue on inserting data into media_tag
create or replace function copy_media_tag_to_queue() returns trigger
AS $copy_media_tag_to_queue$
    begin
        with topic as (
            select t.id from tg_topics t where t.tag_id = new.tag_id
        )
        insert into tg_queue  
            (topic_id, media_id, tag_id)
        select
            topic.id,
            NEW.media_id,
            NEW.tag_id
        from topic;
        return NEW;
    end;
$copy_media_tag_to_queue$ language plpgsql;

-- trigger to add records to the queue after inserting record to the media_tag table
create or replace trigger media_tags_after_insert 
    after insert on media_tag
    for each row execute function copy_media_tag_to_queue (); 

-- queue to send messages to telegram topic
create table tg_queue (
    id bigserial primary key,
    topic_id bigint references tg_topics(id) not null,
    media_id integer references media(id) not null,
    tag_id integer references tag(id) not null
);
create unique index tg_queue_unique_idx on tg_queue (topic_id, media_id);

create table tg_queue_failed (
    id bigserial primary key,
    topic_id bigint references tg_topics(id) not null,
    media_id integer references media(id) not null,
    tag_id integer references tag(id) not null,
    error text not null
);
create unique index tg_queue_failed_unique_idx on tg_queue (topic_id, media_id);

-- tables from main schema (DO NOT CREATE IT) it's for sqlc only

CREATE TABLE tag (
	id serial primary key,
	name varchar(128) NOT NULL,
	CONSTRAINT tag_name_key UNIQUE (name)
);
CREATE INDEX tag_name_idx ON tag USING btree (name);
COMMENT ON TABLE tag IS 'Справочник ключевых слов';

CREATE TABLE media (
	id serial NOT NULL,
	title varchar(256) NOT NULL,
	teaser text NULL,
	"text" text NULL,
	occurrence_date date NOT NULL,
	issue_date timestamp NULL,
	file_url text NULL,
	visible bool DEFAULT true NULL,
	duration interval NULL,
	"size" int4 NULL
);
COMMENT ON TABLE media IS 'Лекция, книга, статья';

CREATE TYPE media_data_type AS ENUM (
	'video',
	'image',
	'telegram');

CREATE TABLE media_data (
	id serial4 NOT NULL,
	media_id int4 NOT NULL,
	data_type media_data_type NOT NULL,
	value text NOT NULL,
	CONSTRAINT media_data_pkey PRIMARY KEY (id),
	CONSTRAINT media_data_media_id_fkey FOREIGN KEY (media_id) REFERENCES media(id) ON DELETE CASCADE
);
CREATE INDEX media_data_media_idx ON media_data USING btree (media_id);
COMMENT ON TABLE media_data IS 'Дополнительные атрибуты объекта';

CREATE TABLE media_tag (
	media_id int4 NOT NULL,
	tag_id int4 NOT NULL,
	CONSTRAINT media_tag_pk PRIMARY KEY (media_id, tag_id),
	CONSTRAINT media_tag_media_id_fkey FOREIGN KEY (media_id) REFERENCES public.media(id) ON DELETE CASCADE,
	CONSTRAINT media_tag_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES public.tag(id) ON DELETE CASCADE
);
CREATE INDEX media_tag_media_idx ON public.media_tag USING btree (media_id);
CREATE INDEX media_tag_tag_idx ON public.media_tag USING btree (tag_id);
COMMENT ON TABLE public.media_tag IS 'Связка лекции и ключевых слов';


-------- START MIGRATION ---------

-- ALTER TYPE media_data_type ADD VALUE IF NOT EXISTS 'telegram';
-- ALTER TABLE media_data ALTER COLUMN media_id SET NOT NULL;
-- ALTER TABLE media_data ALTER COLUMN value SET NOT NULL;
-- ALTER TABLE media_data ALTER COLUMN data_type SET NOT NULL;

-- insert into
-- tg_config (slug, recent_upload_time, settings)
-- values (
--     'goswami.ru',
--     now() - INTERVAL '1 day', 
--     '{
--         "bot_token": "8410539636:AAFvJ4FAvXiyfaNasYBpG_agemgOGIbCGgo",
--         "app_id": 21417045,
--         "app_hash": "fa0dba19e334f1e5e41a1b265f1d2767",
--         "group_id": -1002586736000,
--         "mtproto_group_id": 2586736000,
--         "access_hash": -6294104672070874117,
--         "upload_threads": 2,
--         "audio": "/crate/audio",
--         "assets": "/srv/www/goswami.ru/assets",
--         "chunk_size": 30,
--         "update_interval": 60,
--         "jobs": 2,
--         "performer": "Бхакти Вигьяна Госвами"
--     }'
-- );
-- ALTER TABLE tg_config  OWNER TO www;
-- ALTER TABLE tg_queue  OWNER TO www;
-- ALTER TABLE tg_queue_failed  OWNER TO www;
-- ALTER TABLE tg_topics  OWNER TO www;