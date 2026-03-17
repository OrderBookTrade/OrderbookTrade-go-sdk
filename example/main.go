package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	sdk "github.com/orderbooktrade/orderbooktrade-go-sdk"
)

func main() {
	baseURL := envOrDefault("ORDERBOOK_URL", "http://localhost:8080")
	privateKey := os.Getenv("PRIVATE_KEY") // hex without 0x prefix

	ctx := context.Background()

	// 1. Create client
	client := sdk.NewClient(baseURL)

	// 2. Login with EIP-712
	if privateKey != "" {
		domain := sdk.DefaultDomain()
		// Override chain/contract if needed
		// domain.ChainID = 84532
		// domain.VerifyingContract = "0x..."

		signer, err := sdk.NewSigner(privateKey, domain)
		if err != nil {
			log.Fatalf("create signer: %v", err)
		}

		loginResp, err := client.Login(ctx, signer)
		if err != nil {
			log.Fatalf("login: %v", err)
		}
		fmt.Printf("Logged in as %s, token expires at %d\n", loginResp.Address, loginResp.Expires)

		// --- Place an order example ---
		// orderReq, err := sdk.BuildSignedOrder(
		// 	signer,
		// 	"order-uuid-123",   // id
		// 	1,                  // market_id
		// 	sdk.SideBuy,        // side
		// 	sdk.OrderTypeLimit, // type
		// 	sdk.OutcomeYes,     // outcome
		// 	sdk.TimeInForceGTC, // time_in_force
		// 	0.55,               // price
		// 	100.0,              // size
		// 	"12345",            // salt
		// 	"token-id",         // token_id
		// 	"55000000",         // maker_amount (USDC 6 decimals)
		// 	"100000000",        // taker_amount
		// 	"0",                // fee_rate_bps
		// 	time.Now().Add(24*time.Hour).Unix(), // expiration
		// 	1,                  // nonce
		// )
		// if err != nil {
		// 	log.Fatalf("build order: %v", err)
		// }
		// orderID, err := client.CreateOrder(ctx, orderReq)
		// if err != nil {
		// 	log.Fatalf("create order: %v", err)
		// }
		// fmt.Printf("Order created: %s\n", orderID)
		_ = signer
		_ = time.Now // suppress unused import
	}

	// 3. Fetch markets (public, no auth needed)
	markets, err := client.GetMarkets(ctx, "active")
	if err != nil {
		log.Fatalf("get markets: %v", err)
	}
	fmt.Printf("Found %d active markets\n", len(markets))
	for _, m := range markets {
		fmt.Printf("  [%d] %s (%s)\n", m.ID, m.Name, m.Status)
	}

	// 4. Get orderbook for first market
	if len(markets) > 0 {
		ob, err := client.GetOrderBook(ctx, markets[0].ID)
		if err != nil {
			log.Printf("get orderbook: %v", err)
		} else {
			fmt.Printf("OrderBook market=%d: %d yes_bids, %d yes_asks, mid=%.4f\n",
				ob.MarketID, len(ob.YesBids), len(ob.YesAsks), ob.MidPrice)
		}
	}

	// 5. WebSocket streaming
	ws := sdk.NewWSClient(baseURL)
	ws.OnOrderBook = func(marketID int, snapshot *sdk.OrderBookSnapshot) {
		fmt.Printf("[WS] book_update market=%d bids=%d asks=%d mid=%.4f\n",
			marketID, len(snapshot.YesBids), len(snapshot.YesAsks), snapshot.MidPrice)
	}
	ws.OnTrade = func(marketID int, trade *sdk.Trade) {
		fmt.Printf("[WS] trade market=%d price=%.4f size=%.4f side=%s\n",
			marketID, trade.Price, trade.Size, trade.Side)
	}
	ws.OnError = func(err error) {
		log.Printf("[WS] error: %v", err)
	}

	if err := ws.Connect(ctx); err != nil {
		log.Printf("ws connect: %v", err)
	} else {
		if len(markets) > 0 {
			ws.SubscribeOrderBook(markets[0].ID)
			ws.SubscribeTrades(markets[0].ID)
			fmt.Printf("Subscribed to WS for market %d\n", markets[0].ID)
		}
	}

	// Wait for interrupt
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Println("\nShutting down...")
	ws.Close()
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
