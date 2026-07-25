package api

import (
	"encoding/json"
	"net/http"
	"time"

	"dca-backtester/internal/data"
	"dca-backtester/internal/simulate"
)

// Server guarda las dependencias que los handlers necesitan (por ahora, los
// precios históricos, cargados una sola vez al arrancar).
type Server struct {
	prices []data.PricePoint
}

func NewServer(prices []data.PricePoint) *Server {
	return &Server{prices: prices}
}

type backtestRequest struct {
	StartDate     string  `json:"start_date"`
	EndDate       string  `json:"end_date"`
	MonthlyAmount float64 `json:"monthly_amount"`
}

type backtestResponse struct {
	DCA     data.Result `json:"dca"`
	LumpSum data.Result `json:"lump_sum"`
}

// HandleBacktest procesa POST /api/backtest con un JSON
// {"start_date": "2015-01-01", "end_date": "2025-01-01", "monthly_amount": 500}
// y devuelve la comparación DCA vs lump sum para ese escenario.
func (s *Server) HandleBacktest(w http.ResponseWriter, r *http.Request) {
	var req backtestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "json inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		http.Error(w, "start_date inválida (usa YYYY-MM-DD): "+err.Error(), http.StatusBadRequest)
		return
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		http.Error(w, "end_date inválida (usa YYYY-MM-DD): "+err.Error(), http.StatusBadRequest)
		return
	}
	if !endDate.After(startDate) {
		http.Error(w, "end_date debe ser posterior a start_date", http.StatusBadRequest)
		return
	}
	if req.MonthlyAmount <= 0 {
		http.Error(w, "monthly_amount debe ser mayor a 0", http.StatusBadRequest)
		return
	}

	sc := data.Scenario{
		StartDate:     startDate,
		EndDate:       endDate,
		MonthlyAmount: req.MonthlyAmount,
	}

	resp := backtestResponse{
		DCA:     simulate.RunDCA(s.prices, sc),
		LumpSum: simulate.RunLumpSum(s.prices, sc),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
