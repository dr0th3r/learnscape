package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"testing"

	i "github.com/dr0th3r/learnscape/internal"
)

func TestRegistration(t *testing.T) {
	db_url := os.Getenv("DATABASE_URL")
	db_name := "test_" + fmt.Sprint(rand.Int())
	t.Setenv("DATABASE_NAME", db_name)

	if err := createNewDB(db_url, db_name); err != nil {
		t.Error(err)
		return
	}

	t.Cleanup(func() {
		if err := dropDB(db_url, db_name); err != nil {
			fmt.Println(err)
		}
	})

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)
	go i.Run(ctx)

	client := &http.Client{}
	waitForReady(ctx, client)

	t.Run("incomplete body returns 400 bad request", func(t *testing.T) {
		res, err := http.PostForm("http://localhost:8080/register_user", url.Values{
			"name":    {"test"},
			"surname": {"idk"},
			"email":   {"test@idk.com"},
			//password is missing
		})
		if err != nil {
			t.Error(err)
		}
		defer res.Body.Close()

		got := res.StatusCode
		want := http.StatusBadRequest
		if got != want {
			t.Errorf("Got %d, want %d", got, want)
		}
	})

	t.Run("invalid email is rejected", func(t *testing.T) {
		res, err := http.PostForm("http://localhost:8080/register_user", url.Values{
			"name":     {"test"},
			"surname":  {"idk"},
			"email":    {"invalid"},
			"password": {"test123456"},
		})
		if err != nil {
			t.Error(err)
		}
		defer res.Body.Close()

		got := res.StatusCode
		want := http.StatusBadRequest
		if got != want {
			t.Errorf("Got %d, want %d", got, want)
		}
	})
}
