CREATE TABLE IF NOT EXISTS substitute_timetable (
	id INT PRIMARY KEY REFERENCES academic_timetable(id),
	date DATE NOT NULL
);
