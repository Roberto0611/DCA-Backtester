# DCA Backtester — Motor de backtesting concurrente en Go
 
## Idea del proyecto
 
Herramienta que responde una pregunta simple pero útil: **"Si hubiera invertido $X cada mes en el SP500 empezando en [fecha], ¿cuánto tendría hoy? ¿Me hubiera ido mejor metiendo todo de golpe (lump sum)?"**
 
No es un tracker de gastos ni un CRUD más — el objetivo es que la **concurrencia de Go sea el punto central** del proyecto, no un detalle decorativo. Se corren miles de simulaciones históricas en paralelo (distintas fechas de inicio, distintos montos) usando goroutines, channels y worker pools, algo que en lenguajes como Python sin vectorización sería mucho más lento.
 
Es un proyecto genuinamente útil: mucha gente se pregunta si conviene invertir todo de una vez o poco a poco (DCA vs lump sum), y no abundan herramientas open source que lo respondan con datos históricos reales.
 
### Por qué este proyecto y no otro
- No es el típico "hola mundo" de AWS + Go (como una app de registro de finanzas con Lambda).
- Aprovecha justo lo que hace especial a Go: goroutines, WaitGroups, channels, `select`, `context.Context`, worker pools.
- Es autocontenido: los datos históricos del SP500 no cambian, así que no depende de que una API externa esté disponible en tiempo real.
- Se puede tener un MVP funcional en un fin de semana.
---
 
## MVP (versión mínima, 100% local, sin AWS)
 
**Alcance:**
- Un solo ticker (SPY o VOO, proxy del SP500).
- Input: monto mensual, fecha de inicio, fecha de fin.
- Output: cuánto tendrías hoy con DCA vs metiendo todo de un jalón (lump sum) el día 1.
**Fuente de datos:** [Stooq](https://stooq.com), CSV directo sin necesidad de API key:
```bash
curl "https://stooq.com/q/d/l/?s=spy.us&i=d" -o spy.csv
```
 
### Estructura del proyecto
 
```
dca-backtester/
├── go.mod
├── cmd/
│   ├── backtester/
│   │   └── main.go        # CLI entry point
│   └── lambda/
│       └── main.go        # handler para AWS Lambda
├── internal/
│   ├── data/
│   │   ├── loader.go       # lee el CSV, parsea a []PricePoint
│   │   └── types.go        # struct PricePoint, Scenario, Result
│   ├── simulate/
│   │   ├── dca.go          # lógica de simular DCA
│   │   ├── lumpsum.go      # lógica de lump sum
│   │   └── worker_pool.go  # concurrencia para múltiples escenarios
│   └── report/
│       └── report.go       # formatea resultados (tabla en terminal)
├── data/
│   └── spy.csv
├── template.yaml            # AWS SAM (infraestructura como código)
└── README.md
```
 
### Tipos base
 
```go
// internal/data/types.go
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
```
 
### Lógica central (simulación DCA)
 
```go
// internal/simulate/dca.go
package simulate
 
import (
    "dca-backtester/internal/data"
)
 
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
```
 
### Roadmap después del MVP local
1. **Lump sum comparison** — mismo patrón, todo el dinero invertido el día 1.
2. **Múltiples escenarios en paralelo** — aquí entran las goroutines: correr "empezando cada mes posible entre 2000-2024" y ver en qué % de los casos DCA le gana a lump sum.
3. **CLI con flags** (`flag` o `cobra`) para no tener parámetros hardcodeados.
4. **Output bonito** — tabla en terminal (`tablewriter`) o export a CSV/JSON.
5. Envolver la misma lógica en un servidor HTTP / Lambda.
---
 
## Concurrencia: dónde y por qué
 
Cada "escenario" (ej. empezar en enero 2005, empezar en marzo 2020...) es independiente entre sí — candidato perfecto para paralelizar:
 
```go
type tickerResult struct {
    Scenario data.Scenario
    Result   data.Result
}
 
func RunBacktests(scenarios []data.Scenario, prices []data.PricePoint) []data.Result {
    resultsCh := make(chan data.Result, len(scenarios))
    var wg sync.WaitGroup
 
    for _, s := range scenarios {
        wg.Add(1)
        go func(sc data.Scenario) {
            defer wg.Done()
            resultsCh <- simulate.RunDCA(prices, sc)
        }(s)
    }
 
    go func() {
        wg.Wait()
        close(resultsCh)
    }()
 
    var results []data.Result
    for r := range resultsCh {
        results = append(results, r)
    }
    return results
}
```
 
Si el número de escenarios es muy grande, se puede acotar con un worker pool de tamaño fijo (canal semáforo) en vez de lanzar una goroutine por escenario sin límite.
 
**El "wow" del proyecto:** correr 10,000+ simulaciones (ej. "empezando cada día posible de los últimos 20 años") en menos de un segundo, con un benchmark en el README comparando tiempo con-goroutines vs sin-goroutines.
 
---
 
## Frontend
 
Enfoque por etapas — no meterle complejidad antes de tener el backend probado.
 
1. **CLI con buen output en terminal** — demostrable en un GIF para el README, ya vale por sí solo.
2. **API HTTP simple** sobre la misma lógica (`net/http`, sin frameworks):
```go
func handleBacktest(w http.ResponseWriter, r *http.Request) {
    var req struct {
        StartDate     string  `json:"start_date"`
        EndDate       string  `json:"end_date"`
        MonthlyAmount float64 `json:"monthly_amount"`
    }
    json.NewDecoder(r.Body).Decode(&req)
    // parseo de fechas + simulate.RunDCA(...)
    json.NewEncoder(w).Encode(result)
}
```
 
3. **HTML + JS vanilla con Chart.js** — un solo `index.html` con formulario (fechas, monto mensual) y gráfica comparando DCA vs lump sum. Sin React: es un formulario y una gráfica, no justifica el peso de un framework. Además el propio servidor Go puede servir este HTML (`http.FileServer`), sin build steps ni node_modules.
   *(Alternativa: React + Recharts si se quiere practicar el stack completo — en ese caso se despliega aparte, en S3+CloudFront o Vercel. Solo vale la pena para una v2 más ambiciosa.)*
---
 
## Arquitectura en AWS (MVP)
 
```
┌─────────────┐      ┌──────────────┐      ┌───────────────────┐
│   S3 (static │      │ API Gateway  │      │  Lambda (Go)       │
│   website:   │─────▶│  HTTP API    │─────▶│  handler +         │
│   HTML/JS)   │      │  /api/*      │      │  simulate.RunDCA   │
└─────────────┘      └──────────────┘      └─────────┬──────────┘
                                                        │
                                              ┌─────────▼──────────┐
                                              │  S3 (spy.csv,      │
                                              │  datos históricos) │
                                              └─────────────────────┘
```
 
### Por qué cada pieza
- **Lambda en Go**: binario nativo, cold start bajísimo comparado con Python/Node. Con el free tier (1M requests/mes) el proyecto sale prácticamente gratis.
- **S3 para el CSV histórico**: se sube una vez, la Lambda lo lee de ahí en vez de bajarlo en cada corrida. Se puede actualizar con un job manual o programado una vez al mes.
- **S3 static website** para el frontend: sin servidor, casi gratis. CloudFront opcional si se quiere HTTPS + dominio propio.
- **API Gateway (HTTP API, no REST API)**: más barato y simple para este caso de uso — solo conecta frontend con Lambda.
### Handler de Lambda (wrapper delgado sobre la misma lógica del CLI)
 
```go
// cmd/lambda/main.go
package main
 
import (
    "context"
    "encoding/json"
    "dca-backtester/internal/data"
    "dca-backtester/internal/simulate"
 
    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
)
 
var prices []data.PricePoint // se carga una vez por cold start, se reusa en warm starts
 
func init() {
    var err error
    prices, err = data.LoadPricesFromS3("dca-backtester-data", "spy.csv")
    if err != nil {
        panic(err)
    }
}
 
func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
    var body struct {
        StartDate     string  `json:"start_date"`
        MonthlyAmount float64 `json:"monthly_amount"`
    }
    json.Unmarshal([]byte(req.Body), &body)
 
    // parseo de fechas + simulate.RunDCA(prices, sc)
 
    respBody, _ := json.Marshal(result)
    return events.APIGatewayV2HTTPResponse{
        StatusCode: 200,
        Headers:    map[string]string{"Content-Type": "application/json"},
        Body:       string(respBody),
    }, nil
}
 
func main() {
    lambda.Start(handler)
}
```
 
El `init()` que carga los precios una sola vez (reusado en warm starts) es un detalle de la Lambda lifecycle que vale la pena mencionar en el README.
 
### Infraestructura como código (AWS SAM)
 
```yaml
# template.yaml
Resources:
  BacktestFunction:
    Type: AWS::Serverless::Function
    Properties:
      CodeUri: cmd/lambda/
      Handler: bootstrap
      Runtime: provided.al2023
      Architectures: [arm64]
      Events:
        Api:
          Type: HttpApi
          Properties:
            Path: /api/backtest
            Method: post
```
 
```bash
sam build
sam deploy --guided
```
 
*(Alternativa: AWS CDK con Go si se quiere practicar IaC en Go — vale más la pena cuando la infraestructura crece y necesita lógica condicional.)*
 
### Costos esperados
Con el tráfico típico de un proyecto de portafolio, se mantiene dentro del **free tier**:
- Lambda: 1M requests/mes gratis.
- API Gateway HTTP API: 1M requests/mes gratis el primer año.
- S3: centavos por el tamaño del CSV + HTML.
---
 
## Roadmap de infraestructura (crecimiento futuro)
 
1. **Ahora (MVP)**: S3 (frontend) + API Gateway + Lambda + S3 (datos históricos).
2. **Caché de resultados**: DynamoDB, partition key = hash de los parámetros del escenario, para no recalcular lo mismo dos veces.
3. **CI/CD**: GitHub Actions corriendo `sam build && sam deploy` en cada push a `main`, con badge de deploy status en el README.
4. **Extensión natural — vigilante de mercado**: reutilizando el mismo módulo de fetch/paralelización, un servicio que jale precios de varios tickers en paralelo (worker pool + semáforo + `context.WithTimeout`), calcule indicadores simples, y mande alertas por SNS/Discord cuando se cruce un umbral. Se dispara con EventBridge Scheduled Rule.
5. **Extensión natural — portfolio tracker**: mismo módulo base de precios, pero para dar seguimiento a inversiones reales con métricas como TWR vs MWR y comparación contra benchmark.
---
 
## Checklist de arranque
 
- [ ] `go mod init dca-backtester`
- [ ] Descargar `spy.csv` de Stooq y colocarlo en `data/`
- [ ] Implementar `data.LoadPrices` y probarlo con un `fmt.Println` de los primeros registros
- [ ] Implementar `simulate.RunDCA` y verificar el cálculo a mano con un caso chico (ej. 3 meses)
- [ ] `main.go` del CLI corriendo un escenario fijo
- [ ] Agregar `simulate.RunLumpSum` y comparar contra DCA
- [ ] Agregar paralelización con goroutines para múltiples escenarios
- [ ] Exponer como servidor HTTP local (`net/http`)
- [ ] Probar el endpoint con `curl`/Postman
- [ ] `index.html` + Chart.js consumiendo el endpoint local
- [ ] Migrar el handler a Lambda (`cmd/lambda/main.go`)
- [ ] `template.yaml` de SAM + primer `sam deploy`
- [ ] Subir frontend a S3 static website
- [ ] README con GIF de demo + benchmark de concurrencia
 