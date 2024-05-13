package models

import "go.opentelemetry.io/otel"

var (
	tracer = otel.Tracer("models")
)
