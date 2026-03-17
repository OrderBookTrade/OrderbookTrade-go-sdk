package orderbooktrade

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// EIP712Domain holds the EIP-712 domain config (must match server).
type EIP712Domain struct {
	Name              string
	Version           string
	ChainID           int64
	VerifyingContract string
}

// DefaultDomain returns the default domain matching the server config.
func DefaultDomain() EIP712Domain {
	return EIP712Domain{
		Name:              "Polymarket CTF Exchange",
		Version:           "1",
		ChainID:           84532, // Base Sepolia
		VerifyingContract: "0x0000000000000000000000000000000000000000",
	}
}

// Signer provides EIP-712 signing capabilities using a private key.
type Signer struct {
	privateKey *ecdsa.PrivateKey
	address    common.Address
	domain     EIP712Domain
}

// NewSigner creates a Signer from a hex-encoded private key.
func NewSigner(privateKeyHex string, domain EIP712Domain) (*Signer, error) {
	key, err := crypto.HexToECDSA(stripHexPrefix(privateKeyHex))
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}
	addr := crypto.PubkeyToAddress(key.PublicKey)
	return &Signer{
		privateKey: key,
		address:    addr,
		domain:     domain,
	}, nil
}

// Address returns the signer's Ethereum address.
func (s *Signer) Address() string {
	return s.address.Hex()
}

// ---------- Login flow ----------

// GetNonce fetches a login nonce from the server.
func (c *Client) GetNonce(ctx context.Context) (string, error) {
	var resp APIResponse[NonceResponse]
	if err := c.get(ctx, "/auth/nonce", &resp); err != nil {
		return "", err
	}
	return resp.Data.Nonce, nil
}

// Login performs EIP-712 login: fetches nonce, signs it, calls /auth/login, stores JWT.
func (c *Client) Login(ctx context.Context, signer *Signer) (*LoginResponse, error) {
	nonce, err := c.GetNonce(ctx)
	if err != nil {
		return nil, fmt.Errorf("get nonce: %w", err)
	}

	timestamp := time.Now().Unix()
	sig, err := signer.SignLogin(nonce, timestamp)
	if err != nil {
		return nil, fmt.Errorf("sign login: %w", err)
	}

	req := LoginRequest{
		Account:   signer.Address(),
		Nonce:     nonce,
		Timestamp: timestamp,
		Signature: sig,
	}

	var resp APIResponse[LoginResponse]
	if err := c.post(ctx, "/auth/login", req, &resp); err != nil {
		return nil, err
	}

	c.jwtToken = resp.Data.Token
	return &resp.Data, nil
}

// SignLogin signs a login message using EIP-712.
func (s *Signer) SignLogin(nonce string, timestamp int64) (string, error) {
	loginTypes := apitypes.Types{
		"EIP712Domain": {
			{Name: "name", Type: "string"},
			{Name: "version", Type: "string"},
			{Name: "chainId", Type: "uint256"},
			{Name: "verifyingContract", Type: "address"},
		},
		"Login": {
			{Name: "account", Type: "address"},
			{Name: "nonce", Type: "string"},
			{Name: "timestamp", Type: "uint256"},
		},
	}

	typedData := apitypes.TypedData{
		Types:       loginTypes,
		PrimaryType: "Login",
		Domain:      s.buildDomain(),
		Message: apitypes.TypedDataMessage{
			"account":   s.Address(),
			"nonce":     nonce,
			"timestamp": (*math.HexOrDecimal256)(big.NewInt(timestamp)),
		},
	}

	return s.signTypedData(typedData)
}

// ---------- Order signing ----------

// SignOrder signs an order using the CTFExchange EIP-712 Order type.
func (s *Signer) SignOrder(salt, tokenID, makerAmount, takerAmount, feeRateBps string, expiration, nonce int64, sideEnum, sigType int, taker string) (string, error) {
	orderTypes := apitypes.Types{
		"EIP712Domain": {
			{Name: "name", Type: "string"},
			{Name: "version", Type: "string"},
			{Name: "chainId", Type: "uint256"},
			{Name: "verifyingContract", Type: "address"},
		},
		"Order": {
			{Name: "salt", Type: "uint256"},
			{Name: "maker", Type: "address"},
			{Name: "signer", Type: "address"},
			{Name: "taker", Type: "address"},
			{Name: "tokenId", Type: "uint256"},
			{Name: "makerAmount", Type: "uint256"},
			{Name: "takerAmount", Type: "uint256"},
			{Name: "expiration", Type: "uint256"},
			{Name: "nonce", Type: "uint256"},
			{Name: "feeRateBps", Type: "uint256"},
			{Name: "side", Type: "uint8"},
			{Name: "signatureType", Type: "uint8"},
		},
	}

	if taker == "" {
		taker = "0x0000000000000000000000000000000000000000"
	}

	typedData := apitypes.TypedData{
		Types:       orderTypes,
		PrimaryType: "Order",
		Domain:      s.buildDomain(),
		Message: apitypes.TypedDataMessage{
			"salt":          (*math.HexOrDecimal256)(bigFromStr(salt)),
			"maker":         s.Address(),
			"signer":        s.Address(),
			"taker":         taker,
			"tokenId":       (*math.HexOrDecimal256)(bigFromStr(tokenID)),
			"makerAmount":   (*math.HexOrDecimal256)(bigFromStr(makerAmount)),
			"takerAmount":   (*math.HexOrDecimal256)(bigFromStr(takerAmount)),
			"expiration":    (*math.HexOrDecimal256)(big.NewInt(expiration)),
			"nonce":         (*math.HexOrDecimal256)(big.NewInt(nonce)),
			"feeRateBps":    (*math.HexOrDecimal256)(bigFromStr(feeRateBps)),
			"side":          (*math.HexOrDecimal256)(big.NewInt(int64(sideEnum))),
			"signatureType": (*math.HexOrDecimal256)(big.NewInt(int64(sigType))),
		},
	}

	return s.signTypedData(typedData)
}

// SignCancel signs a cancel order message.
func (s *Signer) SignCancel(orderID string) (string, error) {
	cancelTypes := apitypes.Types{
		"EIP712Domain": {
			{Name: "name", Type: "string"},
			{Name: "version", Type: "string"},
			{Name: "chainId", Type: "uint256"},
			{Name: "verifyingContract", Type: "address"},
		},
		"Cancel": {
			{Name: "orderId", Type: "string"},
			{Name: "maker", Type: "address"},
		},
	}

	typedData := apitypes.TypedData{
		Types:       cancelTypes,
		PrimaryType: "Cancel",
		Domain:      s.buildDomain(),
		Message: apitypes.TypedDataMessage{
			"orderId": orderID,
			"maker":   s.Address(),
		},
	}

	return s.signTypedData(typedData)
}

// ---------- Internal ----------

func (s *Signer) buildDomain() apitypes.TypedDataDomain {
	return apitypes.TypedDataDomain{
		Name:              s.domain.Name,
		Version:           s.domain.Version,
		ChainId:           math.NewHexOrDecimal256(s.domain.ChainID),
		VerifyingContract: s.domain.VerifyingContract,
	}
}

func (s *Signer) signTypedData(typedData apitypes.TypedData) (string, error) {
	hash, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return "", fmt.Errorf("hash typed data: %w", err)
	}

	sig, err := crypto.Sign(hash, s.privateKey)
	if err != nil {
		return "", fmt.Errorf("sign: %w", err)
	}

	// Ethereum convention: v = 27 or 28
	if sig[64] < 27 {
		sig[64] += 27
	}

	return fmt.Sprintf("0x%x", sig), nil
}

func stripHexPrefix(s string) string {
	if len(s) >= 2 && (s[:2] == "0x" || s[:2] == "0X") {
		return s[2:]
	}
	return s
}

func bigFromStr(s string) *big.Int {
	if s == "" {
		return big.NewInt(0)
	}
	n, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return big.NewInt(0)
	}
	return n
}
