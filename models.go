package orderbooktrade

import "time"

// ---------- Enums ----------

type Side string

const (
	SideBuy  Side = "buy"
	SideSell Side = "sell"
)

type OrderType string

const (
	OrderTypeLimit  OrderType = "limit"
	OrderTypeMarket OrderType = "market"
)

type OrderStatus string

const (
	StatusOpen      OrderStatus = "OPEN"
	StatusPartial   OrderStatus = "PARTIAL"
	StatusFilled    OrderStatus = "FILLED"
	StatusCancelled OrderStatus = "CANCELLED"
)

type TimeInForce string

const (
	TimeInForceGTC TimeInForce = "GTC"
	TimeInForceIOC TimeInForce = "IOC"
	TimeInForceFOK TimeInForce = "FOK"
)

type Outcome string

const (
	OutcomeYes Outcome = "YES"
	OutcomeNo  Outcome = "NO"
)

type OrderAction string

const (
	OrderActionCreate OrderAction = "create"
	OrderActionCancel OrderAction = "cancel"
)

// ---------- API Response Wrapper ----------

type APIResponse[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

// ---------- Auth ----------

type NonceResponse struct {
	Nonce string `json:"nonce"`
}

type LoginRequest struct {
	Account   string `json:"account"`
	Nonce     string `json:"nonce"`
	Timestamp int64  `json:"timestamp"`
	Signature string `json:"signature"`
}

type LoginResponse struct {
	Token   string `json:"token"`
	Address string `json:"address"`
	Expires int64  `json:"expires"`
}

// ---------- Market ----------

type Market struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	BaseAsset   string    `json:"base_asset"`
	QuoteAsset  string    `json:"quote_asset"`
	MinPrice    string    `json:"min_price"`
	MaxPrice    string    `json:"max_price"`
	MinSize     string    `json:"min_size"`
	MaxSize     string    `json:"max_size"`
	TickSize    string    `json:"tick_size"`
	Category    string    `json:"category"`
	ClosesAt    time.Time `json:"closes_at"`
	ConditionID string    `json:"condition_id"`
	TokenIDYes  string    `json:"token_id_yes"`
	TokenIDNo   string    `json:"token_id_no"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ---------- Order ----------

type Order struct {
	ID            string      `json:"id"`
	MarketID      int         `json:"market_id"`
	MakerAddress  string      `json:"maker_address"`
	Side          Side        `json:"side"`
	Type          OrderType   `json:"type"`
	Outcome       Outcome     `json:"outcome"`
	TimeInForce   TimeInForce `json:"time_in_force"`
	Price         float64     `json:"price"`
	Size          float64     `json:"size"`
	RemainingSize float64     `json:"remaining_size"`
	Status        OrderStatus `json:"status"`
	Timestamp     int64       `json:"timestamp"`

	// EIP-712 fields
	Salt       string `json:"salt"`
	Signature  string `json:"signature"`
	Expiration int64  `json:"expiration"`
	Nonce      int64  `json:"nonce"`

	// Contract fields
	TokenID     string `json:"token_id"`
	MakerAmount string `json:"maker_amount"`
	TakerAmount string `json:"taker_amount"`
	SideEnum    int    `json:"side_enum"`
	Signer      string `json:"signer"`
	Taker       string `json:"taker"`
	FeeRateBps  string `json:"fee_rate_bps"`
}

// OrderRequest is the payload sent to POST /v1/orders.
type OrderRequest struct {
	ID           string      `json:"id"`
	MakerAddress string      `json:"maker_address"`
	MarketID     int         `json:"market_id"`
	Side         Side        `json:"side"`
	Type         OrderType   `json:"type"`
	Outcome      Outcome     `json:"outcome"`
	TimeInForce  TimeInForce `json:"time_in_force,omitempty"`
	Price        float64     `json:"price"`
	Size         float64     `json:"size"`
	Action       OrderAction `json:"action"`

	// EIP-712 signature fields
	Nonce      int64  `json:"nonce,omitempty"`
	Signature  string `json:"signature,omitempty"`
	Salt       string `json:"salt,omitempty"`
	Expiration int64  `json:"expiration,omitempty"`

	// Contract-level fields
	TokenID       string `json:"token_id,omitempty"`
	MakerAmount   string `json:"maker_amount,omitempty"`
	TakerAmount   string `json:"taker_amount,omitempty"`
	SideEnum      int    `json:"side_enum,omitempty"`
	SignatureType int    `json:"signature_type,omitempty"`
	Signer        string `json:"signer,omitempty"`
	TakerAddr     string `json:"taker,omitempty"`
	FeeRateBps    string `json:"fee_rate_bps,omitempty"`
}

// CancelOrderRequest is the payload sent to DELETE /v1/orders/{order_id}.
type CancelOrderRequest struct {
	MakerAddress string `json:"maker_address"`
	MarketID     int    `json:"market_id"`
	Side         string `json:"side"`
	Outcome      string `json:"outcome"`
	Signature    string `json:"signature"`
}

type OrderIDResponse struct {
	OrderID string `json:"order_id"`
}

// ---------- OrderBook ----------

type Level struct {
	Price float64 `json:"price"`
	Size  float64 `json:"size"`
}

type OrderBookSnapshot struct {
	MarketID  int     `json:"market_id"`
	YesBids   []Level `json:"yes_bids"`
	YesAsks   []Level `json:"yes_asks"`
	NoBids    []Level `json:"no_bids"`
	NoAsks    []Level `json:"no_asks"`
	LastPrice float64 `json:"last_price"`
	MidPrice  float64 `json:"mid_price"`
}

// ---------- Trade ----------

type Trade struct {
	ID           string  `json:"id"`
	MarketID     int     `json:"market_id"`
	MakerOrderID string  `json:"maker_order_id"`
	TakerOrderID string  `json:"taker_order_id"`
	Side         Side    `json:"side"`
	Price        float64 `json:"price"`
	Size         float64 `json:"size"`
	Timestamp    int64   `json:"timestamp"`
}

// ---------- Candle ----------

type Candle struct {
	StartTime int64   `json:"startTime"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}

// ---------- Balance ----------

type Balance struct {
	ID          int64   `json:"id"`
	UserAddress string  `json:"user_address"`
	TokenType   string  `json:"token_type"`
	MarketID    *int    `json:"market_id"`
	Available   float64 `json:"available"`
	Locked      float64 `json:"locked"`
}

// ---------- Position ----------

type Position struct {
	MarketID   int     `json:"market_id"`
	MarketName string  `json:"market_name"`
	Outcome    string  `json:"outcome"`
	TotalSize  float64 `json:"total_size"`
	AvgPrice   float64 `json:"avg_price"`
	TotalCost  float64 `json:"total_cost"`
}

// ---------- Deposit / Withdrawal ----------

type DepositRequest struct {
	Amount float64 `json:"amount"`
}

type DepositResponse struct {
	UserAddress string  `json:"user_address"`
	Amount      float64 `json:"amount"`
	RefID       string  `json:"ref_id"`
}

type WithdrawRequest struct {
	Amount float64 `json:"amount"`
}

type WithdrawResponse struct {
	WithdrawalID string  `json:"withdrawal_id"`
	Amount       float64 `json:"amount"`
	Status       string  `json:"status"`
}

// ---------- WebSocket ----------

type WSMessage struct {
	Type     string `json:"type"`      // subscribe | unsubscribe
	Channel  string `json:"channel"`   // orderbook | trades
	MarketID int    `json:"market_id"`
}

type WSEvent struct {
	Type     string      `json:"type"`      // book_update | trade | candle
	MarketID int         `json:"market_id"`
	Data     interface{} `json:"data"`
}
