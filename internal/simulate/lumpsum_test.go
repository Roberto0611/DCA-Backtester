package simulate

import (
	"testing"

	"dca-backtester/internal/data"
)

func TestRunLumpSum_ThreeMonths(t *testing.T) {
	prices := []data.PricePoint{
		{Date: date(2024, 1, 1), Close: 100},
		{Date: date(2024, 2, 1), Close: 100},
		{Date: date(2024, 3, 1), Close: 50},
		{Date: date(2024, 4, 1), Close: 200},
	}

	sc := data.Scenario{
		StartDate:     date(2024, 1, 1),
		EndDate:       date(2024, 4, 1),
		MonthlyAmount: 100,
	}

	got := RunLumpSum(prices, sc)

	// mismo total que DCA hubiera invertido: 3 meses * 100 = 300
	wantInvested := 300.0
	// todo entra el día 1 @ 100 -> 3 shares
	wantShares := 3.0
	wantFinalValue := wantShares * 200 // valorado al último precio (abril)

	if got.TotalInvested != wantInvested {
		t.Errorf("TotalInvested = %v, want %v", got.TotalInvested, wantInvested)
	}
	if got.FinalValue != wantFinalValue {
		t.Errorf("FinalValue = %v, want %v", got.FinalValue, wantFinalValue)
	}
}
