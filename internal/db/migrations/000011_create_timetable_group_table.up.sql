CREATE TABLE IF NOT EXISTS timetable_group (
	timetable_id INT REFERENCES timetable(id) NOT NULL,
	group_id INT REFERENCES "group"(id) NOT NULL,
	PRIMARY KEY (timetable_id, group_id)
);
