CREATE TABLE IF NOT EXISTS users (
    id serial PRIMARY KEY,
    email varchar(255) NOT NULL,
    fullname varchar(255) NOT NULL,
    phone varchar(255),
    location varchar(255),
    bio varchar(255),
    web_url varchar(255),
    picture_url varchar(255),
    tfa_enabled boolean,
    verified boolean,
    password TEXT NOT NULL,
    backup_codes text[],
    tfa_enabled_at TIMESTAMP,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,

	CONSTRAINT users_unique_email UNIQUE (email)
);

CREATE INDEX IF NOT EXISTS index_users_on_email ON public.users USING btree (email);