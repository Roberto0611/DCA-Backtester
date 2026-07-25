// Package data contains the data types for the investment simulator.
package data

import "time"

type PricePoint struct {
	Date  time.Time
	Close float64
}

type Scenario struct {
	StartDate     time.Time
	EndDate       time.Time
	MonthlyAmount float64
}

type Result struct {
	Scenario      Scenario
	TotalInvested float64
	FinalValue    float64
	ReturnPct     float64
}
