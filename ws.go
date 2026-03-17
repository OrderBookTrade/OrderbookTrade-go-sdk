package orderbooktrade

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSClient manages a WebSocket connection to the matching engine.
type WSClient struct {
	url  string
	conn *websocket.Conn
	mu   sync.Mutex

	// Callbacks
	OnOrderBook func(marketID int, snapshot *OrderBookSnapshot)
	OnTrade     func(marketID int, trade *Trade)
	OnCandle    func(marketID int, candle *Candle)
	OnError     func(err error)

	done chan struct{}
}

// NewWSClient creates a new WebSocket client.
// baseURL should be like "http://localhost:8080" — it will be converted to ws://.
func NewWSClient(baseURL string) *WSClient {
	wsURL := strings.Replace(baseURL, "https://", "wss://", 1)
	wsURL = strings.Replace(wsURL, "http://", "ws://", 1)
	wsURL = strings.TrimRight(wsURL, "/") + "/ws"

	return &WSClient{
		url:  wsURL,
		done: make(chan struct{}),
	}
}

// Connect establishes the WebSocket connection and starts the read loop.
func (ws *WSClient) Connect(ctx context.Context) error {
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, ws.url, nil)
	if err != nil {
		return fmt.Errorf("websocket dial: %w", err)
	}
	ws.conn = conn

	go ws.readLoop()
	return nil
}

// Subscribe subscribes to a channel for a specific market.
// channel: "orderbook" or "trades"
func (ws *WSClient) Subscribe(channel string, marketID int) error {
	return ws.sendMessage(WSMessage{
		Type:     "subscribe",
		Channel:  channel,
		MarketID: marketID,
	})
}

// Unsubscribe unsubscribes from a channel for a specific market.
func (ws *WSClient) Unsubscribe(channel string, marketID int) error {
	return ws.sendMessage(WSMessage{
		Type:     "unsubscribe",
		Channel:  channel,
		MarketID: marketID,
	})
}

// SubscribeOrderBook is a convenience method to subscribe to orderbook updates.
func (ws *WSClient) SubscribeOrderBook(marketID int) error {
	return ws.Subscribe("orderbook", marketID)
}

// SubscribeTrades is a convenience method to subscribe to trade updates.
func (ws *WSClient) SubscribeTrades(marketID int) error {
	return ws.Subscribe("trades", marketID)
}

// Close closes the WebSocket connection.
func (ws *WSClient) Close() error {
	close(ws.done)
	ws.mu.Lock()
	defer ws.mu.Unlock()
	if ws.conn != nil {
		return ws.conn.Close()
	}
	return nil
}

// ---------- Internal ----------

func (ws *WSClient) sendMessage(msg WSMessage) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	if ws.conn == nil {
		return fmt.Errorf("websocket not connected")
	}
	return ws.conn.WriteJSON(msg)
}

func (ws *WSClient) readLoop() {
	defer func() {
		ws.mu.Lock()
		if ws.conn != nil {
			ws.conn.Close()
		}
		ws.mu.Unlock()
	}()

	for {
		select {
		case <-ws.done:
			return
		default:
		}

		_, message, err := ws.conn.ReadMessage()
		if err != nil {
			if ws.OnError != nil {
				ws.OnError(err)
			}
			return
		}

		ws.handleMessage(message)
	}
}

func (ws *WSClient) handleMessage(data []byte) {
	var event WSEvent
	if err := json.Unmarshal(data, &event); err != nil {
		if ws.OnError != nil {
			ws.OnError(fmt.Errorf("unmarshal ws event: %w", err))
		}
		return
	}

	switch event.Type {
	case "book_update":
		if ws.OnOrderBook != nil {
			raw, _ := json.Marshal(event.Data)
			var snapshot OrderBookSnapshot
			if err := json.Unmarshal(raw, &snapshot); err == nil {
				ws.OnOrderBook(event.MarketID, &snapshot)
			}
		}
	case "trade":
		if ws.OnTrade != nil {
			raw, _ := json.Marshal(event.Data)
			var trade Trade
			if err := json.Unmarshal(raw, &trade); err == nil {
				ws.OnTrade(event.MarketID, &trade)
			}
		}
	case "candle":
		if ws.OnCandle != nil {
			raw, _ := json.Marshal(event.Data)
			var candle Candle
			if err := json.Unmarshal(raw, &candle); err == nil {
				ws.OnCandle(event.MarketID, &candle)
			}
		}
	}
}

// ConnectWithReconnect connects and automatically reconnects on disconnect.
func (ws *WSClient) ConnectWithReconnect(ctx context.Context, subscriptions []WSMessage) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if err := ws.Connect(ctx); err != nil {
				if ws.OnError != nil {
					ws.OnError(fmt.Errorf("connect: %w", err))
				}
				time.Sleep(3 * time.Second)
				continue
			}

			// Re-subscribe
			for _, sub := range subscriptions {
				if err := ws.sendMessage(sub); err != nil {
					if ws.OnError != nil {
						ws.OnError(fmt.Errorf("resubscribe: %w", err))
					}
				}
			}

			// Block until connection drops
			<-ws.done
			ws.done = make(chan struct{})
			time.Sleep(3 * time.Second)
		}
	}()
}
