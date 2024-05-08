CREATE OR REPLACE FUNCTION validate_timetable_does_not_have_date()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM timetable
        WHERE id = NEW.timetable_id
            AND type IN ('regular', 'substitute')
    ) THEN
        RAISE EXCEPTION 'Invalid timetable type - should not have type ''regular'' or ''event''';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION validate_timetable_has_date()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM timetable
        WHERE id = NEW.timetable_id
            AND type NOT IN ('regular', 'substitute')
    ) THEN
        RAISE EXCEPTION 'Invalid timetable type - should have type ''regular'' or ''event''';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TYPE note_type AS ENUM ('homework', 'test');

-- This triggers are possible because in PostgreSQL the inheritance does not apply to triggers

CREATE TABLE IF NOT EXISTS note (
        id SERIAL PRIMARY KEY,
        type NOTE_TYPE NOT NULL,
        content TEXT NOT NULL,
        timetable_id INT NOT NULL REFERENCES timetable(id)
);

CREATE TRIGGER note_references_valid_timetable
        BEFORE INSERT OR UPDATE 
        ON note
        FOR EACH ROW
        EXECUTE FUNCTION validate_timetable_does_not_have_date();

CREATE TABLE IF NOT EXISTS note_with_date (
        date DATE NOT NULL
) INHERITS (note);

CREATE TRIGGER note_with_date_references_valid_timetable
        BEFORE INSERT OR UPDATE 
        ON note_with_date
        FOR EACH ROW
        EXECUTE FUNCTION validate_timetable_has_date();
