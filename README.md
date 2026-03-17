# OrderbookTrade Go SDK

Go client SDK for the OrderbookTrade matching engine — a prediction market orderbook exchange with EIP-712 authentication and real-time WebSocket streaming.

## Install

```bash
go get github.com/orderbooktrade/orderbooktrade-go-sdk
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    sdk "github.com/orderbooktrade/orderbooktrade-go-sdk"
)

func main() {
    ctx := context.Background()
    client := sdk.NewClient("http://localhost:8080")

    // Login with EIP-712 signature
    domain := sdk.DefaultDomain()
    signer, _ := sdk.NewSigner("your-private-key-hex", domain)
    loginResp, _ := client.Login(ctx, signer)
    fmt.Printf("Logged in as %s\n", loginResp.Address)

    // Fetch markets
    markets, _ := client.GetMarkets(ctx, "active")
    for _, m := range markets {
        fmt.Printf("[%d] %s\n", m.ID, m.Name)
    }

    // Get orderbook
    ob, _ := client.GetOrderBook(ctx, 1)
    fmt.Printf("Mid price: %.4f\n", ob.MidPrice)
}
```

## Features

| Module | Methods |
|--------|---------|
| **Auth** | `NewSigner`, `Login`, `GetNonce` |
| **Markets** | `GetMarkets`, `GetMarket`, `GetOrderBook`, `GetMarketTrades`, `GetCandles` |
| **Orders** | `CreateOrder`, `CancelOrder`, `GetMyOrders`, `BuildSignedOrder`, `BuildCancelRequest` |
| **User** | `GetMyBalances`, `GetMyPositions`, `GetMyTrades`, `Deposit`, `Withdraw` |
| **WebSocket** | `Subscribe`, `Unsubscribe`, `ConnectWithReconnect` |

## Authentication

The SDK supports two auth mechanisms used by the matching engine:

- **EIP-712 Signature** — required for placing/cancelling orders. The `Signer` handles all signing.
- **JWT Token** — returned after login, automatically attached to authenticated requests (`/me/*` endpoints).

```go
// Create a signer from your Ethereum private key
domain := sdk.EIP712Domain{
    Name:              "Polymarket CTF Exchange",
    Version:           "1",
    ChainID:           84532,                                      // Base Sepolia
    VerifyingContract: "0x0000000000000000000000000000000000000000", // your CTFExchange address
}
signer, err := sdk.NewSigner("abcdef1234...", domain)

// Login — fetches nonce, signs it, calls /auth/login, stores JWT automatically
loginResp, err := client.Login(ctx, signer)
```

## Placing Orders

```go
import (
    "time"
    "github.com/google/uuid"
)

// Build a signed order
orderReq, err := sdk.BuildSignedOrder(
    signer,
    uuid.New().String(),  // order ID
    1,                    // market_id
    sdk.SideBuy,          // side
    sdk.OrderTypeLimit,   // type
    sdk.OutcomeYes,       // outcome
    sdk.TimeInForceGTC,   // time_in_force
    0.55,                 // price (0-1)
    100.0,                // size (shares)
    "12345",              // salt
    "token-id-yes",       // token_id
    "55000000",           // maker_amount
    "100000000",          // taker_amount
    "0",                  // fee_rate_bps
    time.Now().Add(24*time.Hour).Unix(), // expiration
    1,                    // nonce
)

// Submit to matching engine
orderID, err := client.CreateOrder(ctx, orderReq)
```

## Cancelling Orders

```go
cancelReq, err := sdk.BuildCancelRequest(
    signer,
    "order-id-to-cancel",
    1,     // market_id
    "buy", // side
    "YES", // outcome
)
_, err = client.CancelOrder(ctx, "order-id-to-cancel", cancelReq)
```

## Querying User Data

```go
// Balances (USDC, YES/NO tokens)
balances, _ := client.GetMyBalances(ctx)

// Positions (aggregated by market + outcome)
positions, _ := client.GetMyPositions(ctx)

// Trade history
trades, _ := client.GetMyTrades(ctx, 1, 50) // market_id=1, limit=50

// Order history
orders, _ := client.GetMyOrders(ctx, "", 0, 50) // all statuses, all markets, limit=50
```

## WebSocket Streaming

```go
ws := sdk.NewWSClient("http://localhost:8080")

// Set callbacks
ws.OnOrderBook = func(marketID int, snapshot *sdk.OrderBookSnapshot) {
    fmt.Printf("Book update: market=%d mid=%.4f\n", marketID, snapshot.MidPrice)
}
ws.OnTrade = func(marketID int, trade *sdk.Trade) {
    fmt.Printf("Trade: price=%.4f size=%.4f\n", trade.Price, trade.Size)
}
ws.OnError = func(err error) {
    log.Printf("WS error: %v", err)
}

// Connect and subscribe
ws.Connect(ctx)
ws.SubscribeOrderBook(1) // market_id=1
ws.SubscribeTrades(1)

// Or use auto-reconnect
ws.ConnectWithReconnect(ctx, []sdk.WSMessage{
    {Type: "subscribe", Channel: "orderbook", MarketID: 1},
    {Type: "subscribe", Channel: "trades", MarketID: 1},
})
```

## Client Options

```go
// Custom HTTP client
client := sdk.NewClient("http://localhost:8080",
    sdk.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
)

// Pre-set JWT token (skip login)
client := sdk.NewClient("http://localhost:8080",
    sdk.WithJWTToken("eyJhbGci..."),
)
```

## API Reference

### Public Endpoints (no auth)

| Method | Endpoint |
|--------|----------|
| `GetNonce` | `GET /auth/nonce` |
| `Login` | `POST /auth/login` |
| `GetMarkets` | `GET /v1/markets` |
| `GetMarket` | `GET /v1/markets/{id}` |
| `GetOrderBook` | `GET /v1/markets/{id}/orderbook` |
| `GetMarketTrades` | `GET /v1/markets/{id}/trades` |
| `GetCandles` | `GET /v1/markets/{id}/candles` |

### Authenticated Endpoints (EIP-712 signature)

| Method | Endpoint |
|--------|----------|
| `CreateOrder` | `POST /v1/orders` |
| `CancelOrder` | `DELETE /v1/orders/{id}` |

### Authenticated Endpoints (JWT)

| Method | Endpoint |
|--------|----------|
| `GetMyOrders` | `GET /v1/me/orders` |
| `GetMyBalances` | `GET /v1/me/balances` |
| `GetMyPositions` | `GET /v1/me/positions` |
| `GetMyTrades` | `GET /v1/me/trades` |
| `Deposit` | `POST /v1/me/deposit` |
| `Withdraw` | `POST /v1/me/withdraw` |

### WebSocket

| Channel | Events |
|---------|--------|
| `orderbook` | `book_update` — full orderbook snapshot |
| `trades` | `trade` — individual trade execution |
| `trades` | `candle` — 1-minute OHLCV candle update |

## Data Types

### OrderBookSnapshot

```go
type OrderBookSnapshot struct {
    MarketID  int     // Market identifier
    YesBids   []Level // YES outcome buy orders (price descending)
    YesAsks   []Level // YES outcome sell orders (price ascending)
    NoBids    []Level // NO outcome buy orders
    NoAsks    []Level // NO outcome sell orders
    LastPrice float64 // Last traded price
    MidPrice  float64 // Mid price between best bid/ask
}
```

### Enums

```go
// Side
sdk.SideBuy   // "buy"
sdk.SideSell  // "sell"

// OrderType
sdk.OrderTypeLimit   // "limit"
sdk.OrderTypeMarket  // "market"

// Outcome
sdk.OutcomeYes  // "YES"
sdk.OutcomeNo   // "NO"

// TimeInForce
sdk.TimeInForceGTC  // Good Til Cancelled
sdk.TimeInForceIOC  // Immediate Or Cancel
sdk.TimeInForceFOK  // Fill Or Kill

// OrderStatus
sdk.StatusOpen       // "OPEN"
sdk.StatusPartial    // "PARTIAL"
sdk.StatusFilled     // "FILLED"
sdk.StatusCancelled  // "CANCELLED"
```

## Local Development

If the SDK is not published yet, use `replace` in your bot's `go.mod`:

```go
require github.com/orderbooktrade/orderbooktrade-go-sdk v0.0.0

replace github.com/orderbooktrade/orderbooktrade-go-sdk => /path/to/OrderbookTrade-go-sdk
```

## License

MIT
