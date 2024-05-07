CREATE TABLE IF NOT EXISTS users_group (
	user_id UUID REFERENCES users(id) NOT NULL,
	group_id INT REFERENCES "group"(id) NOT NULL,
	PRIMARY KEY (user_id, group_id)
);
