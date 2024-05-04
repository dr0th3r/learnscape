CREATE TYPE weekday AS ENUM ('Po', 'Út', 'St', 'Čt', 'Pá');

CREATE TABLE IF NOT EXISTS regular_timetable (
	id SERIAL PRIMARY KEY,
	period_id SERIAL REFERENCES period(id) NOT NULL,
	subject_id SERIAL REFERENCES subject(id) NOT NULL,
	room_id SERIAL REFERENCES room(id) NOT NULL,
	weekday WEEKDAY
);
