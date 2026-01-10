package mempool

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/maindotmarcell/beutel-backend/internal/chain"
	"github.com/maindotmarcell/beutel-backend/pkg/types"
)

// Client implements chain.Provider using mempool.space API
type Client struct {
	httpClient *http.Client
	network    chain.Network
	baseURL    string
}

// NewClient creates a new mempool.space client for the given network
func NewClient(network chain.Network) *Client {
	baseURL := "https://mempool.space"
	switch network {
	case chain.Testnet3:
		baseURL = "https://mempool.space/testnet"
	case chain.Testnet4:
		baseURL = "https://mempool.space/testnet4"
	case chain.Signet:
		baseURL = "https://mempool.space/signet"
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		network: network,
		baseURL: baseURL,
	}
}

// Network returns the configured network
func (c *Client) Network() chain.Network {
	return c.network
}

func (c *Client) GetBalance(address string) (*types.Balance, error) {
	url := fmt.Sprintf("%s/api/address/%s", c.baseURL, address)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch address: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("mempool API error (status %d): %s", resp.StatusCode, string(body))
	}

	var data struct {
		ChainStats struct {
			FundedTxoSum int64 `json:"funded_txo_sum"`
			SpentTxoSum  int64 `json:"spent_txo_sum"`
		} `json:"chain_stats"`
		MempoolStats struct {
			FundedTxoSum int64 `json:"funded_txo_sum"`
			SpentTxoSum  int64 `json:"spent_txo_sum"`
		} `json:"mempool_stats"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	confirmed := data.ChainStats.FundedTxoSum - data.ChainStats.SpentTxoSum
	unconfirmed := data.MempoolStats.FundedTxoSum - data.MempoolStats.SpentTxoSum

	return &types.Balance{
		Confirmed:   confirmed,
		Unconfirmed: unconfirmed,
		Total:       confirmed + unconfirmed,
	}, nil
}

func (c *Client) GetUTXOs(address string) ([]types.UTXO, error) {
	url := fmt.Sprintf("%s/api/address/%s/utxo", c.baseURL, address)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch UTXOs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("mempool API error (status %d): %s", resp.StatusCode, string(body))
	}

	var mempoolUTXOs []struct {
		Txid   string `json:"txid"`
		Vout   int    `json:"vout"`
		Value  int64  `json:"value"`
		Status struct {
			Confirmed   bool  `json:"confirmed"`
			BlockHeight int64 `json:"block_height"`
		} `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&mempoolUTXOs); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	utxos := make([]types.UTXO, len(mempoolUTXOs))
	for i, u := range mempoolUTXOs {
		utxos[i] = types.UTXO{
			Txid:        u.Txid,
			Vout:        u.Vout,
			Value:       u.Value,
			Confirmed:   u.Status.Confirmed,
			BlockHeight: u.Status.BlockHeight,
		}
	}

	return utxos, nil
}

func (c *Client) GetTransactions(address string) ([]types.Transaction, error) {
	url := fmt.Sprintf("%s/api/address/%s/txs", c.baseURL, address)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("mempool API error (status %d): %s", resp.StatusCode, string(body))
	}

	var mempoolTxs []struct {
		Txid   string `json:"txid"`
		Status struct {
			Confirmed   bool  `json:"confirmed"`
			BlockHeight int64 `json:"block_height"`
			BlockTime   int64 `json:"block_time"`
		} `json:"status"`
		Fee int64 `json:"fee"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&mempoolTxs); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	txs := make([]types.Transaction, len(mempoolTxs))
	for i, tx := range mempoolTxs {
		txs[i] = types.Transaction{
			Txid:        tx.Txid,
			Confirmed:   tx.Status.Confirmed,
			BlockHeight: tx.Status.BlockHeight,
			BlockTime:   tx.Status.BlockTime,
			Fee:         tx.Fee,
		}
	}

	return txs, nil
}

func (c *Client) GetFeeRates() (*types.FeeRates, error) {
	url := fmt.Sprintf("%s/api/v1/fees/recommended", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch fee rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("mempool API error (status %d): %s", resp.StatusCode, string(body))
	}

	var fees types.FeeRates
	if err := json.NewDecoder(resp.Body).Decode(&fees); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &fees, nil
}

func (c *Client) BroadcastTx(txHex string) (string, error) {
	url := fmt.Sprintf("%s/api/tx", c.baseURL)

	resp, err := c.httpClient.Post(url, "text/plain", strings.NewReader(txHex))
	if err != nil {
		return "", fmt.Errorf("failed to broadcast transaction: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("broadcast failed: %s", string(body))
	}

	// mempool.space returns the txid as plain text
	return strings.TrimSpace(string(body)), nil
}

// Ensure Client implements Provider
var _ chain.Provider = (*Client)(nil)
