CREATE TABLE IF NOT EXISTS "group" (
	id SERIAL PRIMARY KEY,
	class_id INT REFERENCES class(id),
	name VARCHAR(100) NOT NULL
);
