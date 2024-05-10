CREATE TABLE IF NOT EXISTS absence (
	id SERIAL PRIMARY KEY,
	user_id UUID NOT NULL REFERENCES users(id),
	span TSRANGE
);
