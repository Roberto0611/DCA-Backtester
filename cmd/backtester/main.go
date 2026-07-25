package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"dca-backtester/internal/data"
	"dca-backtester/internal/simulate"
)

func main() {
	prices, err := data.LoadPrices("data/spy.csv")
	if err != nil {
		log.Fatal(err)
	}

	sc := data.Scenario{
		StartDate:     time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:       time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		MonthlyAmount: 500,
	}

	dcaResult := simulate.RunDCA(prices, sc)
	lumpSumResult := simulate.RunLumpSum(prices, sc)

	fmt.Printf("Escenario: %s -> %s, $%.2f/mes\n\n",
		sc.StartDate.Format("2006-01-02"), sc.EndDate.Format("2006-01-02"), sc.MonthlyAmount)

	fmt.Println("DCA:")
	fmt.Printf("  Invertido total: $%.2f\n", dcaResult.TotalInvested)
	fmt.Printf("  Valor final:     $%.2f\n", dcaResult.FinalValue)
	fmt.Printf("  Retorno:         %.2f%%\n\n", dcaResult.ReturnPct)

	fmt.Println("Lump sum:")
	fmt.Printf("  Invertido total: $%.2f\n", lumpSumResult.TotalInvested)
	fmt.Printf("  Valor final:     $%.2f\n", lumpSumResult.FinalValue)
	fmt.Printf("  Retorno:         %.2f%%\n\n", lumpSumResult.ReturnPct)

	if lumpSumResult.FinalValue > dcaResult.FinalValue {
		fmt.Println("-> Lump sum le hubiera ganado a DCA en este período.")
	} else {
		fmt.Println("-> DCA le hubiera ganado a lump sum en este período.")
	}

	fmt.Println()
	runMassBacktest(prices)
}

// runMassBacktest corre un escenario de 5 años empezando en cada día de
// trading disponible, en paralelo con un worker pool, y reporta en qué
// porcentaje de los casos DCA le ganó a lump sum.
func runMassBacktest(prices []data.PricePoint) {
	scenarios := simulate.GenerateScenarios(prices, 1, 500)

	start := time.Now()
	results := simulate.RunBacktestsPooled(scenarios, prices, runtime.NumCPU())
	elapsed := time.Since(start)

	dcaWins := 0
	for _, r := range results {
		if r.DCA.FinalValue > r.LumpSum.FinalValue {
			dcaWins++
		}
	}

	fmt.Printf("Simulación masiva: %d escenarios (ventanas de 1 año, cada día de trading como inicio)\n", len(scenarios))
	fmt.Printf("Tiempo: %s (con %d workers)\n", elapsed, runtime.NumCPU())
	fmt.Printf("DCA le ganó a lump sum en %d/%d casos (%.1f%%)\n",
		dcaWins, len(results), float64(dcaWins)/float64(len(results))*100)
}
