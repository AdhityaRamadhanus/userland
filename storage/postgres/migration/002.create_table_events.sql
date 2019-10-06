CREATE TABLE IF NOT EXISTS events (
    id serial PRIMARY KEY,
    user_id int NOT NULL,
    event varchar(255) NOT NULL,
    user_agent TEXT,
    ip TEXT,
    client_id int,
    client_name TEXT,
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS index_events_on_user_id ON public.events USING btree (user_id);
CREATE INDEX IF NOT EXISTS index_events_on_event ON public.events USING btree (event);