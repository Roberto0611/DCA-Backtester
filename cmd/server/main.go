package main

import (
	"log"
	"net/http"

	"dca-backtester/internal/api"
	"dca-backtester/internal/data"
)

func main() {
	prices, err := data.LoadPrices("data/spy.csv")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("cargados %d precios históricos\n", len(prices))

	srv := api.NewServer(prices)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/backtest", srv.HandleBacktest)
	mux.Handle("/", http.FileServer(http.Dir("web")))

	addr := ":8080"
	log.Printf("escuchando en http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
