package main

import (
	"context"
	"log"
	"net/http"
	"ufc_stats_api/internal/config"
	"ufc_stats_api/internal/handlers"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	//api endpoints and how to handle request using a multiplexer or a mux for short
	mux := http.NewServeMux()
	mux.HandleFunc("/fighters", handlers.GetAllFighters(pool))
	mux.HandleFunc("/fighters/search", handlers.SearchFightersByName(pool))
	mux.HandleFunc("/events", handlers.GetAllEvents(pool))
	mux.HandleFunc("/events/{id}/fights", handlers.GetFightsByEventID(pool))
	mux.HandleFunc("/fighters/{id}/fights", handlers.GetFightsByFighterID(pool))
	mux.HandleFunc("/fights/{id}/stats", handlers.GetFightStatsByFightID(pool))
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))

}
