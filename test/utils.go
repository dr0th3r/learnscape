package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

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
