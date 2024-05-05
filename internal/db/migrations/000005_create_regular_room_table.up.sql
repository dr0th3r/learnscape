CREATE TABLE IF NOT EXISTS room (
	id SERIAL PRIMARY KEY,
	school_id INT REFERENCES school(id) NOT NULL,
	name VARCHAR(255),
	teacher_id UUID REFERENCES users(id)
);
