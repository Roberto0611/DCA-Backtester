package simulate

import (
	"dca-backtester/internal/data"
)

// RunLumpSum simula invertir de una sola vez, el día de StartDate (o el
// primer precio disponible después), el total que en DCA se hubiera
// invertido a lo largo de todo el período (sc.MonthlyAmount * # de meses).
func RunLumpSum(prices []data.PricePoint, sc data.Scenario) data.Result {
	totalInvested := totalDCAInvestment(sc)

	priceIdx := 0
	for priceIdx < len(prices) && prices[priceIdx].Date.Before(sc.StartDate) {
		priceIdx++
	}
	if priceIdx >= len(prices) {
		return data.Result{Scenario: sc, TotalInvested: totalInvested}
	}

	entryPrice := prices[priceIdx].Close
	shares := totalInvested / entryPrice

	finalPrice := prices[len(prices)-1].Close
	finalValue := shares * finalPrice

	return data.Result{
		Scenario:      sc,
		TotalInvested: totalInvested,
		FinalValue:    finalValue,
		ReturnPct:     (finalValue - totalInvested) / totalInvested * 100,
	}
}

// totalDCAInvestment calcula cuánto se habría invertido en total corriendo
// DCA sobre el mismo escenario (mismo # de meses), para que la comparación
// DCA vs lump sum sea sobre el mismo monto total.
func totalDCAInvestment(sc data.Scenario) float64 {
	months := 0
	currentMonth := sc.StartDate
	for currentMonth.Before(sc.EndDate) {
		months++
		currentMonth = currentMonth.AddDate(0, 1, 0)
	}
	return float64(months) * sc.MonthlyAmount
}
