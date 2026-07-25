# DCA Backtester

Motor de backtesting en Go que responde: *"Si hubiera invertido $X cada mes en el SP500 (SPY) desde [fecha], ¿cuánto tendría hoy? ¿Me hubiera ido mejor invirtiendo todo de golpe (lump sum)?"*

Corre miles de simulaciones históricas en paralelo con goroutines y un worker pool, en milisegundos. Incluye CLI, API HTTP y un frontend simple con gráfica.

## Uso

### CLI

```bash
go run ./cmd/backtester
```

Corre un escenario fijo (DCA vs lump sum, $500/mes, 2015–2025) y luego una simulación masiva: un escenario de 1 año por cada día de trading disponible como fecha de inicio, en paralelo.

### Servidor web (API + frontend)

```bash
go run ./cmd/server
```

Levanta un servidor en `http://localhost:8080` que sirve:
- `POST /api/backtest` — recibe `{"start_date", "end_date", "monthly_amount"}` y devuelve la comparación DCA vs lump sum en JSON.
- `/` — el frontend (`web/index.html`), un formulario + gráfica (Chart.js) que consume el endpoint anterior.

```bash
curl -X POST http://localhost:8080/api/backtest \
  -d '{"start_date":"2015-01-01","end_date":"2025-01-01","monthly_amount":500}'
```

## Estructura

```
cmd/backtester/   # CLI, punto de entrada
cmd/server/       # servidor HTTP, punto de entrada
internal/data/    # tipos + carga del CSV histórico
internal/simulate/# lógica de negocio: DCA, lump sum, concurrencia
internal/api/     # handlers HTTP sobre la misma lógica de simulate
web/              # frontend estático (index.html + Chart.js vía CDN)
data/spy.csv      # precios históricos de SPY (Stooq)
```

`internal/data` no depende de nada más; `internal/simulate` depende de `internal/data`; `internal/api` depende de ambos. `cmd/backtester` y `cmd/server` son dos puntos de entrada distintos que reusan exactamente la misma lógica de negocio. `internal/` es una carpeta especial de Go: nada fuera del módulo puede importarla.

## Concurrencia

- `simulate.RunBacktests` — una goroutine por escenario.
- `simulate.RunBacktestsPooled` — worker pool de tamaño fijo (`runtime.NumCPU()` workers) leyendo de un channel compartido; es lo que usa el CLI para miles de escenarios.

**Benchmark:** ~3700 simulaciones (DCA + lump sum cada una) en **~9ms** con worker pool.

## Hallazgo

Sobre el dataset 2010–2026 (mercado alcista casi ininterrumpido):
- Ventanas de **5 años**: lump sum le gana a DCA en el **100%** de los casos.
- Ventanas de **1 año**: DCA gana en el **~17%** de los casos (típicamente cuando el inicio coincide con una caída, ej. covid).

## Tests

```bash
go test ./...
```

## Roadmap

- [x] Servidor HTTP (`net/http`) sobre la misma lógica
- [x] Frontend simple (HTML + Chart.js)
- [ ] Deploy a AWS Lambda
- [ ] Infraestructura como código (AWS SAM)
