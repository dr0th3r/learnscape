package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func createNewDB(url string, db_name string) error {
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(),
		"CREATE DATABASE "+db_name,
	)
	if err != nil {
		return err
	}

	return nil
}

func dropDB(url string, db_name string) error {
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(),
		"DROP DATABASE "+db_name+" WITH (FORCE)",
	)
	if err != nil {
		return err
	}

	return nil
}

func waitForReady(ctx context.Context) error {
	startTime := time.Now()
	client := &http.Client{}
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/health_check", nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		res, err := client.Do(req)
		if err != nil {
			//fmt.Printf("Error making request %s\n", err)
			continue
		}
		if res.StatusCode == http.StatusOK {
			fmt.Println("Endpoint is ready")
			res.Body.Close()
			return nil
		}
		res.Body.Close()

		select {
		case <-ctx.Done():
			ctx.Err()
		default:
			if time.Since(startTime) > time.Second*5 {
				return fmt.Errorf("Timeout reached while waiting for endpoint")
			}
			time.Sleep(250 * time.Millisecond)
		}
	}
}

func createSchool(db *pgx.Conn) (string, error) {
	id := fmt.Sprint(rand.Intn(10000))
	_, err := db.Exec(context.Background(), "insert into school (id, name, city, zip_code, street_address) values ($1, $2, $3, $4, $5)",
		id, "test", "test city", "123 45", "street 7",
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func createUser(db *pgx.Conn) (string, error) {
	id := uuid.NewString()
	_, err := db.Exec(context.Background(), "insert into users (id, name, surname, email, password) values ($1, $2, $3, $4, $5)",
		id, "test", "idk", "test@idk.com", "testidk123",
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func createPeriod(db *pgx.Conn, schoolId string) (string, error) {
	id := fmt.Sprint(rand.Intn(10000))

	_, err := db.Exec(context.Background(), "insert into period (id, school_id, span) values ($1, $2, $3)", id, schoolId, "[8:00:00, 8:45:00]")
	if err != nil {
		return "", err
	}

	return id, nil
}

func createSubject(db *pgx.Conn) (string, error) {
	id := fmt.Sprint(rand.Intn(10000))

	_, err := db.Exec(context.Background(), "insert into subject (id, name) values ($1, $2)", id, "Math")
	if err != nil {
		return "", err
	}

	return id, nil
}

func createRoom(db *pgx.Conn, teacherId string, schoolId string) (string, error) {
	id := fmt.Sprint(rand.Intn(10000))

	_, err := db.Exec(context.Background(), "insert into room (id, name, teacher_id, school_id) values ($1, $2, $3, $4)", id, "Labs", teacherId, schoolId)
	if err != nil {
		return "", err
	}

	return id, nil
}

func createRegularTimetable(db *pgx.Conn, periodId, subjectId, schoolId, roomId string) (string, error) {
	id := fmt.Sprint(rand.Intn(10000))

	_, err := db.Exec(context.Background(),
		`
		WITH inserted_timetable AS (
		    INSERT INTO timetable (id, school_id, type) 
		    VALUES ($1, $2, $3)
		    RETURNING id
		),
		inserted_academic_timetable AS (
		    INSERT INTO academic_timetable (id, period_id, subject_id, room_id)
		    SELECT id, $4, $5, $6
		    FROM inserted_timetable
		)
		INSERT INTO regular_timetable (id, weekday)
		SELECT id, $7
		FROM inserted_timetable
		`,
		id, schoolId, "regular", periodId, subjectId, roomId, "Po",
	)

	if err != nil {
		return "", err
	}

	return id, nil
}

func createClass(db *pgx.Conn, teacherId string) (string, error) {
	id := fmt.Sprint(rand.Intn(10000))

	_, err := db.Exec(context.Background(),
		"insert into class (id, name, year, class_teacher_id) values ($1, $2, $3, $4)",
		id, "test", 1, teacherId,
	)
	if err != nil {
		return "", err
	}

	return id, nil
}

func createGroup(db *pgx.Conn) (string, error) {
	id := fmt.Sprint(rand.Intn(10000))

	_, err := db.Exec(context.Background(),
		`insert into "group" (id, name) values ($1, $2)`,
		id, "test_group",
	)
	if err != nil {
		return "", err
	}

	return id, nil
}

func createReport(db *pgx.Conn, reportedBy, timetableId string) (string, error) {
	id := fmt.Sprint(rand.Intn(10000))

	_, err := db.Exec(context.Background(),
		"insert into report (id, timetable_id, reported_by, topic_covered) values ($1, $2, $3, $4)",
		id, timetableId, reportedBy, "linear algebra",
	)
	if err != nil {
		return "", err
	}

	return id, nil
}
