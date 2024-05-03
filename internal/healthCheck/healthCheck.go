package hcheck

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func HandleHealthCheck() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Println(r.Context())
			_, span := otel.Tracer("health_chechk").Start(r.Context(), "check")
			defer span.End()
			span.SetAttributes(attribute.String("test", "idk"))
			w.WriteHeader(http.StatusOK)
		},
	)
}
