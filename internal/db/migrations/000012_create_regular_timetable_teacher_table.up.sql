CREATE TABLE IF NOT EXISTS regular_timetable_teacher (
	regular_timetable_id INT REFERENCES regular_timetable(id) NOT NULL,
	teacher_id UUID REFERENCES users(id) NOT NULL,
	PRIMARY KEY (regular_timetable_id, teacher_id)
)
