package hcheck

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func HandleHealthCheck() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, span := otel.Tracer("health_chechk").Start(r.Context(), "check")
			defer span.End()
			span.SetAttributes(attribute.String("test", "idk"))
			w.WriteHeader(http.StatusOK)
		},
	)
}
