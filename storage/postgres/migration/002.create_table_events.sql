CREATE TABLE IF NOT EXISTS events (
    id serial PRIMARY KEY,
    userid int NOT NULL,
    event varchar(255) NOT NULL,
    useragent TEXT,
    ip TEXT,
    client_id int,
    client_name TEXT,
    timestamp TIMESTAMP,
    createdAt TIMESTAMP
);

CREATE INDEX IF NOT EXISTS index_events_on_userid ON public.events USING btree (userid);
CREATE INDEX IF NOT EXISTS index_events_on_event ON public.events USING btree (event);