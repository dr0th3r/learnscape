CREATE TABLE IF NOT EXISTS grade (
	student_id UUID REFERENCES users(id) NOT NULL,
	report_id INT REFERENCES report(id) NOT NULL,
	value SMALLINT NOT NULL CHECK (value >= 1 AND value <= 5),
	weight SMALLINT NOT NULL CHECK (weight >= 1 AND weight <= 10)
)
