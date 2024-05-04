package internal

import (
	"net/http"

	hcheck "github.com/dr0th3r/learnscape/internal/healthCheck"
	"github.com/dr0th3r/learnscape/internal/period"
	"github.com/dr0th3r/learnscape/internal/room"
	"github.com/dr0th3r/learnscape/internal/school"
	"github.com/dr0th3r/learnscape/internal/subject"
	"github.com/dr0th3r/learnscape/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func NewServer(db *pgxpool.Pool) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, db)
	var handler http.Handler = mux
	handler = otelhttp.NewHandler(handler, "server")
	return handler
}

func addRoutes(
	mux *http.ServeMux,
	db *pgxpool.Pool,
) {
	mux.Handle("/health_check", hcheck.HandleHealthCheck())
	mux.Handle("POST /register_user", user.HandleRegisterUser(db))
	mux.Handle("POST /login", user.HandleLogin(db))
	mux.Handle("POST /register_school", school.HandleRegisterSchool(db))
	mux.Handle("POST /period", period.HandleCreatePeriod(db))
	mux.Handle("POST /room", room.HandleCreateRoom(db))
	mux.Handle("POST /subject", subject.HandleCreateSubject(db))
}
