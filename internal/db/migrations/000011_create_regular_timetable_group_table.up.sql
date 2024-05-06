CREATE TABLE IF NOT EXISTS regular_timetable_group (
	regular_timetable_id INT REFERENCES regular_timetable(id),
	group_id INT REFERENCES "group"(id),
	PRIMARY KEY (regular_timetable_id, group_id)
);
