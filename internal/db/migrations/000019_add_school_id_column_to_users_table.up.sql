CREATE INDEX idx_users_school_id ON users (school_id);

ALTER TABLE users
ADD COLUMN school_id INT NOT NULL,
ADD CONSTRAINT fk_users_school_id
    FOREIGN KEY (school_id)
    REFERENCES school(id);

