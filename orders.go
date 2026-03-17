package orderbooktrade

import (
	"context"
	"fmt"
)

// CreateOrder submits a new order to the matching engine.
// The OrderRequest must include a valid EIP-712 signature.
func (c *Client) CreateOrder(ctx context.Context, req *OrderRequest) (string, error) {
	req.Action = OrderActionCreate
	var resp APIResponse[OrderIDResponse]
	if err := c.post(ctx, "/v1/orders", req, &resp); err != nil {
		return "", err
	}
	return resp.Data.OrderID, nil
}

// CancelOrder cancels an existing order by ID.
// The CancelOrderRequest must include a valid EIP-712 cancel signature.
func (c *Client) CancelOrder(ctx context.Context, orderID string, req *CancelOrderRequest) (string, error) {
	var resp APIResponse[OrderIDResponse]
	if err := c.del(ctx, fmt.Sprintf("/v1/orders/%s", orderID), req, &resp); err != nil {
		return "", err
	}
	return resp.Data.OrderID, nil
}

// GetMyOrders returns the authenticated user's orders.
// Optional filters: status, marketID, limit.
func (c *Client) GetMyOrders(ctx context.Context, status string, marketID int, limit int) ([]Order, error) {
	path := "/v1/me/orders"
	params := ""
	if status != "" {
		params += "status=" + status + "&"
	}
	if marketID > 0 {
		params += fmt.Sprintf("market_id=%d&", marketID)
	}
	if limit > 0 {
		params += fmt.Sprintf("limit=%d&", limit)
	}
	if params != "" {
		path += "?" + params[:len(params)-1] // trim trailing &
	}

	var resp APIResponse[[]Order]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// BuildSignedOrder is a helper that constructs and signs an OrderRequest using a Signer.
func BuildSignedOrder(signer *Signer, id string, marketID int, side Side, orderType OrderType, outcome Outcome, tif TimeInForce, price, size float64, salt, tokenID, makerAmount, takerAmount, feeRateBps string, expiration, nonce int64) (*OrderRequest, error) {
	sideEnum := 0
	if side == SideSell {
		sideEnum = 1
	}

	sig, err := signer.SignOrder(salt, tokenID, makerAmount, takerAmount, feeRateBps, expiration, nonce, sideEnum, 0, "")
	if err != nil {
		return nil, fmt.Errorf("sign order: %w", err)
	}

	return &OrderRequest{
		ID:            id,
		MakerAddress:  signer.Address(),
		MarketID:      marketID,
		Side:          side,
		Type:          orderType,
		Outcome:       outcome,
		TimeInForce:   tif,
		Price:         price,
		Size:          size,
		Action:        OrderActionCreate,
		Nonce:         nonce,
		Signature:     sig,
		Salt:          salt,
		Expiration:    expiration,
		TokenID:       tokenID,
		MakerAmount:   makerAmount,
		TakerAmount:   takerAmount,
		SideEnum:      sideEnum,
		SignatureType: 0,
		Signer:        signer.Address(),
		TakerAddr:     "0x0000000000000000000000000000000000000000",
		FeeRateBps:    feeRateBps,
	}, nil
}

// BuildCancelRequest is a helper that constructs and signs a CancelOrderRequest.
func BuildCancelRequest(signer *Signer, orderID string, marketID int, side string, outcome string) (*CancelOrderRequest, error) {
	sig, err := signer.SignCancel(orderID)
	if err != nil {
		return nil, fmt.Errorf("sign cancel: %w", err)
	}

	return &CancelOrderRequest{
		MakerAddress: signer.Address(),
		MarketID:     marketID,
		Side:         side,
		Outcome:      outcome,
		Signature:    sig,
	}, nil
}
