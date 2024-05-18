package internal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"

	u "github.com/dr0th3r/learnscape/internal/utils"
)

func initDB(config u.DBConfig) (*pgxpool.Pool, error) {
	connectionUrl := config.GetConnectionUrl()

	//migrate database
	m, err := migrate.New(
		fmt.Sprintf("file://%s", config.MigrationsDir),
		connectionUrl,
	)
	if err != nil {
		return nil, err
	}
	if err := m.Up(); err != nil && err.Error() != "no change" {
		return nil, err
	}

	//return dbpool
	return pgxpool.New(context.Background(), connectionUrl)
}

func Run(ctx context.Context, config *u.Config) (err error) {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	db, err := initDB(config.DB)
	if err != nil {
		return errors.New("error connecting to database: " + err.Error())
	}
	defer db.Close()

	otelShutdown, err := u.SetupOTelSDK(ctx)
	if err != nil {
		return errors.New("error setting up otel " + err.Error())
	}

	defer func() {
		if err := otelShutdown(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down otel: %s\n", err)
		}
	}()

	srv := NewServer(db, config.App)
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(config.Server.Host, fmt.Sprint(config.Server.Port)),
		Handler: srv,
	}

	go func() {
		log.Printf("listening on %s\n", httpServer.Addr)
		fmt.Println("listening")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()

		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()
	return nil
}
