CREATE TABLE IF NOT EXISTS report (
	id SERIAL PRIMARY KEY,
	timetable_id INT REFERENCES timetable(id) NOT NULL,
	reported_by UUID REFERENCES users(id) NOT NULL,
	reported_at TIMESTAMP DEFAULT NOW(),
	topic_covered VARCHAR(255) NOT NULL
);
