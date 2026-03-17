package orderbooktrade

import (
	"context"
	"fmt"
)

// GetMarkets returns all markets, optionally filtered by status.
func (c *Client) GetMarkets(ctx context.Context, status string) ([]Market, error) {
	path := "/v1/markets"
	if status != "" {
		path += "?status=" + status
	}
	var resp APIResponse[[]Market]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetMarket returns a single market by ID.
func (c *Client) GetMarket(ctx context.Context, marketID int) (*Market, error) {
	var resp APIResponse[Market]
	if err := c.get(ctx, fmt.Sprintf("/v1/markets/%d", marketID), &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// GetOrderBook returns the current orderbook snapshot for a market.
func (c *Client) GetOrderBook(ctx context.Context, marketID int) (*OrderBookSnapshot, error) {
	var snapshot OrderBookSnapshot
	if err := c.get(ctx, fmt.Sprintf("/v1/markets/%d/orderbook", marketID), &snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}

// GetMarketTrades returns recent trades for a market.
func (c *Client) GetMarketTrades(ctx context.Context, marketID int, limit int) ([]Trade, error) {
	path := fmt.Sprintf("/v1/markets/%d/trades", marketID)
	if limit > 0 {
		path += fmt.Sprintf("?limit=%d", limit)
	}
	var resp APIResponse[[]Trade]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetCandles returns candlestick data for a market.
func (c *Client) GetCandles(ctx context.Context, marketID int, limit int) ([]Candle, error) {
	path := fmt.Sprintf("/v1/markets/%d/candles", marketID)
	if limit > 0 {
		path += fmt.Sprintf("?limit=%d", limit)
	}
	var resp APIResponse[[]Candle]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
