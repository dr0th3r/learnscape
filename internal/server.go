package internal

import (
	"net/http"

	hcheck "github.com/dr0th3r/learnscape/internal/healthCheck"
	"github.com/dr0th3r/learnscape/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func NewServer(db *pgxpool.Pool, rdb *redis.Client) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, db, rdb)
	var handler http.Handler = mux
	return handler
}

func addRoutes(
	mux *http.ServeMux,
	db *pgxpool.Pool,
	rdb *redis.Client,
) {
	mux.Handle("/health_check", hcheck.HandleHealthCheck())
	mux.Handle("POST /register_user", user.HandleRegisterUser(db, rdb))
}
