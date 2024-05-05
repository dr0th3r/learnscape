CREATE TABLE IF NOT EXISTS regular_report (
	id SERIAL PRIMARY KEY,
	regular_timetable_id INT REFERENCES regular_timetable(id),
	reported_by UUID REFERENCES users(id),
	reported_at TIMESTAMP DEFAULT NOW(),
	topic_covered VARCHAR(255)
);
