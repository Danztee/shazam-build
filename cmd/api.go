package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Danztee/shazam-build/internal/audio"
	repo "github.com/Danztee/shazam-build/internal/database/queries"
	"github.com/Danztee/shazam-build/internal/download"
	"github.com/Danztee/shazam-build/internal/json"
	"github.com/Danztee/shazam-build/internal/songs"
	"github.com/Danztee/shazam-build/internal/spotify"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

func (app *application) mount() http.Handler {

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		json.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	spotifyService := spotify.NewService()
	downloadService := download.NewService(slog.Default())
	audioService := audio.NewService(slog.Default())
	songsService := songs.NewService(repo.New(app.db), app.db, spotifyService, downloadService, audioService)
	songsHandler := songs.NewHandler(songsService)
	r.Post("/songs", songsHandler.AddSong)
	r.Post("/songs/match", songsHandler.MatchSong)

	return r

}

func (app *application) run(h http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      h,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	slog.Info("starting server", "addr", app.config.addr)

	return srv.ListenAndServe()
}

type application struct {
	config Config
	db     *pgx.Conn
}

type Config struct {
	addr string
	db   dbConfig
}

type dbConfig struct {
	dsn string
}
