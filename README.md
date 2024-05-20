## Prerequisites to run
- installed [docker](https://docs.docker.com/engine/install/)
- installed [go](https://go.dev/dl/)

## How to run
**Be sure to be in the root of the project**
```
docker run -d --name jaeger   -e COLLECTOR_OTLP_ENABLED=true   -p 16686:16686   -p 4317:4317   -p 4318:4318   jaegertracing/all-in-one:latest
export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT="http://localhost:4318/v1/traces"
./scripts/init_db.sh 
go run cmd/learnscape/main.go
```

## Todo list (only most important listed):
- improve tests
- add redirects on register/login
- add returning newly created models
- add proper ui
