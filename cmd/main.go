package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/Danztee/shazam-build/internal/env"
	"github.com/jackc/pgx/v5"
)

func main() {

	// if err := godotenv.Load(); err != nil {
	// 	slog.Error("failed to load .env file", "error", err)
	// 	os.Exit(1)
	// }

	ctx := context.Background()

	cfg := Config{
		addr: ":" + env.GetString("PORT", ""),
		db: dbConfig{
			dsn: env.GetString("DATABASE_URL", ""),
		},
	}

	// logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Database
	conn, err := pgx.Connect(ctx, cfg.db.dsn)
	if err != nil {
		logger.Error("failed to connect to database", "error", err, "dsn", cfg.db.dsn)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	logger.Info("Connected to database")

	api := application{
		config: cfg,
		db:     conn,
	}

	if err := api.run(api.mount()); err != nil {
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}

}
