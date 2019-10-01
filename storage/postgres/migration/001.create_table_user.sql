CREATE TABLE IF NOT EXISTS users (
    id serial PRIMARY KEY,
    email varchar(255) NOT NULL,
    fullname varchar(255) NOT NULL,
    phone varchar(255),
    location varchar(255),
    bio varchar(255),
    weburl varchar(255),
    pictureurl varchar(255),
    tfaenabled boolean,
    password TEXT NOT NULL,
    createdAt TIMESTAMP,
    updatedAt TIMESTAMP,

	CONSTRAINT users_unique_email UNIQUE (email)
);

CREATE INDEX IF NOT EXISTS index_users_on_email ON public.users USING btree (email);
