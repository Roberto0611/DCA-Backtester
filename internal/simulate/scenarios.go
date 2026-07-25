package simulate

import "dca-backtester/internal/data"

// GenerateScenarios crea un escenario de duración fija (durationYears) por
// cada día de trading disponible como posible fecha de inicio, descartando
// los que se saldrían del rango de datos disponible.
func GenerateScenarios(prices []data.PricePoint, durationYears int, monthlyAmount float64) []data.Scenario {
	if len(prices) == 0 {
		return nil
	}

	lastDate := prices[len(prices)-1].Date
	scenarios := make([]data.Scenario, 0, len(prices))

	for _, p := range prices {
		end := p.Date.AddDate(durationYears, 0, 0)
		if end.After(lastDate) {
			continue
		}
		scenarios = append(scenarios, data.Scenario{
			StartDate:     p.Date,
			EndDate:       end,
			MonthlyAmount: monthlyAmount,
		})
	}

	return scenarios
}
