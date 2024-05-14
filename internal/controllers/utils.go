package controllers

import "go.opentelemetry.io/otel"

var (
	tracer = otel.Tracer("controllers")
)
