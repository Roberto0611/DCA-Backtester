package data

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

// LoadPrices lee un CSV con columnas Date,Open,High,Low,Close,Volume
// y devuelve los puntos de precio ordenados por fecha ascendente.
func LoadPrices(path string) ([]PricePoint, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("abriendo %s: %w", path, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parseando csv: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("csv vacío o sin datos: %s", path)
	}

	prices := make([]PricePoint, 0, len(rows)-1)
	for i, row := range rows[1:] { // rows[0] es el header
		date, err := time.Parse("2006-01-02", row[0])
		if err != nil {
			return nil, fmt.Errorf("fila %d: fecha inválida %q: %w", i+2, row[0], err)
		}
		close, err := strconv.ParseFloat(row[4], 64)
		if err != nil {
			return nil, fmt.Errorf("fila %d: close inválido %q: %w", i+2, row[4], err)
		}
		prices = append(prices, PricePoint{Date: date, Close: close})
	}

	return prices, nil
}
