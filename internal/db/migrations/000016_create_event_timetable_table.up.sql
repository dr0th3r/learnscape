CREATE TABLE IF NOT EXISTS event_timetable (
	id INT PRIMARY KEY REFERENCES timetable(id),
	span TSRANGE NOT NULL,
	name VARCHAR(255) NOT NULL,
	description TEXT
)
