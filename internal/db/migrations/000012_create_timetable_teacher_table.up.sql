CREATE TABLE IF NOT EXISTS timetable_teacher (
	timetable_id INT REFERENCES timetable(id) NOT NULL,
	teacher_id UUID REFERENCES users(id) NOT NULL,
	PRIMARY KEY (timetable_id, teacher_id)
)
