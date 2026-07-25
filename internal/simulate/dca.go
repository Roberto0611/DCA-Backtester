package simulate

import (
	"dca-backtester/internal/data"
)

// RunDCA simula invertir sc.MonthlyAmount cada mes, desde StartDate hasta
// EndDate, comprando al precio de cierre disponible más cercano (>=) a cada
// fecha mensual. El valor final se calcula al último precio disponible.
func RunDCA(prices []data.PricePoint, sc data.Scenario) data.Result {
	var shares float64
	var invested float64

	currentMonth := sc.StartDate
	priceIdx := 0

	for currentMonth.Before(sc.EndDate) {
		for priceIdx < len(prices) && prices[priceIdx].Date.Before(currentMonth) {
			priceIdx++
		}
		if priceIdx >= len(prices) {
			break
		}

		price := prices[priceIdx].Close
		shares += sc.MonthlyAmount / price
		invested += sc.MonthlyAmount

		currentMonth = currentMonth.AddDate(0, 1, 0)
	}

	finalPrice := prices[len(prices)-1].Close
	finalValue := shares * finalPrice

	return data.Result{
		Scenario:      sc,
		TotalInvested: invested,
		FinalValue:    finalValue,
		ReturnPct:     (finalValue - invested) / invested * 100,
	}
}
