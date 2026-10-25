//go:generate oapi-codegen -o ./api.gen.go -config ./.oapi-config.yml ./openapi.yml
//go:generate sqlc -f .sqlc.yml generate
package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/krilor/solid-potato/db"
)

const (
	defaultAddr     = ":8080"
	shutdownTimeout = 10 * time.Second
)

// dbConfig is the database connection config, read from the environment.
type dbConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
}

func dbConfigFromEnv() (dbConfig, error) {
	cfg := dbConfig{
		Host:     getenv("DB_HOST", "localhost"),
		Port:     getenv("DB_PORT", "5432"),
		Name:     getenv("DB_NAME", "potato"),
		User:     getenv("DB_USER", "reader"),
		Password: os.Getenv("DB_PASSWORD"),
		SSLMode:  getenv("DB_SSLMODE", "prefer"),
	}
	if cfg.Password == "" {
		return cfg, errors.New("DB_PASSWORD must be set")
	}
	return cfg, nil
}

func (c dbConfig) connString() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.User, c.Password),
		Host:     net.JoinHostPort(c.Host, c.Port),
		Path:     "/" + c.Name,
		RawQuery: url.Values{"sslmode": {c.SSLMode}}.Encode(),
	}
	return u.String()
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	if err := run(); err != nil {
		slog.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// ctx is cancelled on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	addr := getenv("LISTEN_ADDR", defaultAddr)
	dbCfg, err := dbConfigFromEnv()
	if err != nil {
		return err
	}

	pool, err := pgxpool.New(ctx, dbCfg.connString())
	if err != nil {
		return err
	}
	defer pool.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		return err
	}
	slog.Info("connected to database")

	srv := &http.Server{
		Addr:              addr,
		Handler:           newServer(db.New(pool)),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		ErrorLog:          slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
	}

	serveErr := make(chan error, 1)
	go func() {
		slog.Info("starting server", "addr", addr)
		serveErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
		stop() // restore default signal behavior so a second interrupt kills the process
		slog.Info("shutdown signal received, shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	if err := <-serveErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	slog.Info("server stopped")
	return nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
