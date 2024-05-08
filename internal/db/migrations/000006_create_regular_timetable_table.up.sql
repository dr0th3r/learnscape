CREATE TYPE weekday AS ENUM ('Po', 'Út', 'St', 'Čt', 'Pá');

CREATE TYPE timetable_type AS ENUM ('regular', 'substitute', 'event');

CREATE TABLE IF NOT EXISTS timetable (
	id SERIAL PRIMARY KEY,
	school_id INT REFERENCES school(id) NOT NULL,
	type TIMETABLE_TYPE NOT NULL
);

 CREATE TABLE IF NOT EXISTS academic_timetable (
	id INT PRIMARY KEY REFERENCES timetable(id),
	period_id INT REFERENCES period(id) NOT NULL,
	subject_id INT REFERENCES subject(id) NOT NULL,
	room_id INT REFERENCES room(id) NOT NULL
);

CREATE TABLE IF NOT EXISTS regular_timetable (
	id INT PRIMARY KEY REFERENCES academic_timetable(id),
	weekday WEEKDAY NOT NULL
);
