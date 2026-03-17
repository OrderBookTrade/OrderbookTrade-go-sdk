package orderbooktrade

import (
	"context"
	"fmt"
)

// GetMyBalances returns the authenticated user's balances (USDC, YES, NO tokens).
func (c *Client) GetMyBalances(ctx context.Context) ([]Balance, error) {
	var resp APIResponse[[]Balance]
	if err := c.get(ctx, "/v1/me/balances", &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetMyPositions returns the authenticated user's aggregated positions.
func (c *Client) GetMyPositions(ctx context.Context) ([]Position, error) {
	var resp APIResponse[[]Position]
	if err := c.get(ctx, "/v1/me/positions", &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetMyTrades returns the authenticated user's trade history.
func (c *Client) GetMyTrades(ctx context.Context, marketID int, limit int) ([]Trade, error) {
	path := "/v1/me/trades"
	params := ""
	if marketID > 0 {
		params += fmt.Sprintf("market_id=%d&", marketID)
	}
	if limit > 0 {
		params += fmt.Sprintf("limit=%d&", limit)
	}
	if params != "" {
		path += "?" + params[:len(params)-1]
	}

	var resp APIResponse[[]Trade]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// Deposit requests a deposit of USDC.
func (c *Client) Deposit(ctx context.Context, amount float64) (*DepositResponse, error) {
	var resp APIResponse[DepositResponse]
	if err := c.post(ctx, "/v1/me/deposit", DepositRequest{Amount: amount}, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Withdraw requests a withdrawal of USDC.
func (c *Client) Withdraw(ctx context.Context, amount float64) (*WithdrawResponse, error) {
	var resp APIResponse[WithdrawResponse]
	if err := c.post(ctx, "/v1/me/withdraw", WithdrawRequest{Amount: amount}, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
