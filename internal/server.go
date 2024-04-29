package internal

import (
	"net/http"

	hcheck "github.com/dr0th3r/learnscape/internal/healthCheck"
	"github.com/dr0th3r/learnscape/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewServer(db *pgxpool.Pool) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, db)
	var handler http.Handler = mux
	return handler
}

func addRoutes(
	mux *http.ServeMux,
	db *pgxpool.Pool,
) {
	mux.Handle("/health_check", hcheck.HandleHealthCheck())
	mux.Handle("POST /register_user", user.HandleRegisterUser(db))
}
