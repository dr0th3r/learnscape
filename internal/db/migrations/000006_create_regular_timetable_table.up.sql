CREATE TYPE weekday AS ENUM ('Po', 'Út', 'St', 'Čt', 'Pá');

CREATE TABLE IF NOT EXISTS regular_timetable (
	id SERIAL PRIMARY KEY,
	period_id INT REFERENCES period(id) NOT NULL,
	subject_id INT REFERENCES subject(id) NOT NULL,
	school_id INT REFERENCES school(id) NOT NULL,
	room_id INT REFERENCES room(id) NOT NULL,
	weekday WEEKDAY
);
