package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dr0th3r/learnscape/internal/healthCheck"
)

func TestHealthCheck(t *testing.T) {
	t.Run("health check returns 200", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/health_check", nil)
		res := httptest.NewRecorder()

		hcheck.HandleHealthCheck().ServeHTTP(res, req)

		statusGot := res.Result().StatusCode
		statusWant := 200
		if statusGot != statusWant {
			t.Errorf("got %q, want %q", statusGot, statusWant)
		}

		bodyGot := res.Body.String()
		bodyWant := ""
		if bodyGot != bodyWant {
			t.Errorf("got %q, want %q", bodyGot, bodyWant)
		}

	})
}
