package simulate

import (
	"sync"

	"dca-backtester/internal/data"
)

// ScenarioResult empareja el resultado de DCA y lump sum para el mismo
// escenario, para poder compararlos directamente.
type ScenarioResult struct {
	Scenario data.Scenario
	DCA      data.Result
	LumpSum  data.Result
}

func runBoth(prices []data.PricePoint, sc data.Scenario) ScenarioResult {
	return ScenarioResult{
		Scenario: sc,
		DCA:      RunDCA(prices, sc),
		LumpSum:  RunLumpSum(prices, sc),
	}
}

// RunBacktests lanza una goroutine POR escenario. Simple y rápido para
// cientos o pocos miles de escenarios, pero sin límite de concurrencia:
// con demasiados escenarios puede saturar CPU/memoria.
func RunBacktests(scenarios []data.Scenario, prices []data.PricePoint) []ScenarioResult {
	resultsCh := make(chan ScenarioResult, len(scenarios))
	var wg sync.WaitGroup

	for _, sc := range scenarios {
		wg.Add(1)
		go func(sc data.Scenario) {
			defer wg.Done()
			resultsCh <- runBoth(prices, sc)
		}(sc)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	results := make([]ScenarioResult, 0, len(scenarios))
	for r := range resultsCh {
		results = append(results, r)
	}
	return results
}

// RunBacktestsPooled usa un worker pool de tamaño fijo: `workers` goroutines
// fijas leen escenarios de un channel compartido hasta que se acaban. Es el
// patrón correcto cuando el número de escenarios es grande (miles+): acota
// la concurrencia en vez de lanzar una goroutine por tarea.
func RunBacktestsPooled(scenarios []data.Scenario, prices []data.PricePoint, workers int) []ScenarioResult {
	scenarioCh := make(chan data.Scenario)
	resultsCh := make(chan ScenarioResult, len(scenarios))
	var wg sync.WaitGroup

	// workers goroutines fijas, todas leyendo del mismo channel de entrada.
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for sc := range scenarioCh {
				resultsCh <- runBoth(prices, sc)
			}
		}()
	}

	// alimenta el channel de entrada; se cierra cuando ya no hay más escenarios.
	go func() {
		for _, sc := range scenarios {
			scenarioCh <- sc
		}
		close(scenarioCh)
	}()

	// cuando todos los workers terminan, cerramos el channel de salida.
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	results := make([]ScenarioResult, 0, len(scenarios))
	for r := range resultsCh {
		results = append(results, r)
	}
	return results
}
