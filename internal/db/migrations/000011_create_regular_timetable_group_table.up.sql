CREATE TABLE IF NOT EXISTS regular_timetable_group (
	regular_timetable_id INT REFERENCES regular_timetable(id) NOT NULL,
	group_id INT REFERENCES "group"(id) NOT NULL,
	PRIMARY KEY (regular_timetable_id, group_id)
);
