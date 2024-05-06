CREATE TABLE IF NOT EXISTS users_group (
	user_id UUID REFERENCES users(id),
	group_id INT REFERENCES "group"(id),
	PRIMARY KEY (user_id, group_id)
);
