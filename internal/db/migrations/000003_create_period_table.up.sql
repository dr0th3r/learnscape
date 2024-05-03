CREATE FUNCTION time_subtype_diff(x time, y time) RETURNS float8 AS
'SELECT EXTRACT (EPOCH FROM (x - y))' LANGUAGE sql STRICT IMMUTABLE;

CREATE TYPE timerange AS RANGE (
	subtype = time,
	subtype_diff = time_subtype_diff
);

CREATE EXTENSION btree_gist;

CREATE TABLE IF NOT EXISTS period (
	id SERIAL PRIMARY KEY,
	school_id UUID REFERENCES school(id),
	span TIMERANGE NOT NULL,
	EXCLUDE USING gist (school_id WITH =, span WITH &&)
);


