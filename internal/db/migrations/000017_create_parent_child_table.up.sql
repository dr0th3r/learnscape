CREATE TABLE IF NOT EXISTS parent_child (
	parent_id UUID REFERENCES users(id),
	child_id UUID REFERENCES users(id),
	PRIMARY KEY (parent_id, child_id)
);
