package simulate

import (
	"testing"
	"time"

	"dca-backtester/internal/data"
)

func TestRunDCA_ThreeMonths(t *testing.T) {
	prices := []data.PricePoint{
		{Date: date(2024, 1, 1), Close: 100},
		{Date: date(2024, 2, 1), Close: 100}, // sube 0% -> compra 1 share
		{Date: date(2024, 3, 1), Close: 50},  // baja a 50 -> compra 2 shares
		{Date: date(2024, 4, 1), Close: 200}, // sube a 200 -> compra 0.5 shares
	}

	sc := data.Scenario{
		StartDate:     date(2024, 1, 1),
		EndDate:       date(2024, 4, 1),
		MonthlyAmount: 100,
	}

	got := RunDCA(prices, sc)

	wantInvested := 300.0
	// compras: ene @100 -> 1 share, feb @100 -> 1 share, mar @50 -> 2 shares
	// (abril no se compra: EndDate=1-abril no es "before" abril)
	wantShares := 1 + 1 + 2.0 // 4 shares
	wantFinalValue := wantShares * 200 // valorado al último precio del dataset (abril)

	if got.TotalInvested != wantInvested {
		t.Errorf("TotalInvested = %v, want %v", got.TotalInvested, wantInvested)
	}
	if got.FinalValue != wantFinalValue {
		t.Errorf("FinalValue = %v, want %v", got.FinalValue, wantFinalValue)
	}
}

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
