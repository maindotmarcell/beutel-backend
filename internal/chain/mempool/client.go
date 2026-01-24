package mempool

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/maindotmarcell/beutel-backend/internal/chain"
	"github.com/maindotmarcell/beutel-backend/internal/logging"
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

// doGet performs a GET request and records upstream call details in the log context
func (c *Client) doGet(logCtx *logging.LogContext, url string) (*http.Response, error) {
	start := time.Now()

	resp, err := c.httpClient.Get(url)
	duration := time.Since(start).Milliseconds()

	// Record upstream call details for canonical logging
	if logCtx != nil {
		logCtx.Add("upstream_url", url)
		logCtx.Add("upstream_method", "GET")
		logCtx.Add("upstream_duration_ms", duration)
		if resp != nil {
			logCtx.Add("upstream_status", resp.StatusCode)
		}
		if err != nil {
			logCtx.Add("upstream_error", err.Error())
		}
	}

	return resp, err
}

// doPost performs a POST request and records upstream call details in the log context
func (c *Client) doPost(logCtx *logging.LogContext, url string, contentType string, body io.Reader) (*http.Response, error) {
	start := time.Now()

	resp, err := c.httpClient.Post(url, contentType, body)
	duration := time.Since(start).Milliseconds()

	// Record upstream call details for canonical logging
	if logCtx != nil {
		logCtx.Add("upstream_url", url)
		logCtx.Add("upstream_method", "POST")
		logCtx.Add("upstream_duration_ms", duration)
		if resp != nil {
			logCtx.Add("upstream_status", resp.StatusCode)
		}
		if err != nil {
			logCtx.Add("upstream_error", err.Error())
		}
	}

	return resp, err
}

func (c *Client) GetBalance(logCtx *logging.LogContext, address string) (*types.Balance, error) {
	url := fmt.Sprintf("%s/api/address/%s", c.baseURL, address)

	resp, err := c.doGet(logCtx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch address: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("mempool API error (status %d): %s", resp.StatusCode, string(body))
		if logCtx != nil {
			logCtx.Add("upstream_error", errMsg)
		}
		return nil, fmt.Errorf("%s", errMsg)
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

func (c *Client) GetUTXOs(logCtx *logging.LogContext, address string) ([]types.UTXO, error) {
	url := fmt.Sprintf("%s/api/address/%s/utxo", c.baseURL, address)

	resp, err := c.doGet(logCtx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch UTXOs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("mempool API error (status %d): %s", resp.StatusCode, string(body))
		if logCtx != nil {
			logCtx.Add("upstream_error", errMsg)
		}
		return nil, fmt.Errorf("%s", errMsg)
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

// mempoolTxInput represents a transaction input from mempool.space
type mempoolTxInput struct {
	Prevout *struct {
		ScriptpubkeyAddress string `json:"scriptpubkey_address"`
		Value               int64  `json:"value"`
	} `json:"prevout"`
}

// mempoolTxOutput represents a transaction output from mempool.space
type mempoolTxOutput struct {
	ScriptpubkeyAddress string `json:"scriptpubkey_address"`
	Value               int64  `json:"value"`
}

// mempoolTx represents a full transaction from mempool.space
type mempoolTx struct {
	Txid   string `json:"txid"`
	Status struct {
		Confirmed   bool  `json:"confirmed"`
		BlockHeight int64 `json:"block_height"`
		BlockTime   int64 `json:"block_time"`
	} `json:"status"`
	Vin  []mempoolTxInput  `json:"vin"`
	Vout []mempoolTxOutput `json:"vout"`
	Fee  int64             `json:"fee"`
}

func (c *Client) GetTransactions(logCtx *logging.LogContext, address string) ([]types.Transaction, error) {
	url := fmt.Sprintf("%s/api/address/%s/txs", c.baseURL, address)

	resp, err := c.doGet(logCtx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("mempool API error (status %d): %s", resp.StatusCode, string(body))
		if logCtx != nil {
			logCtx.Add("upstream_error", errMsg)
		}
		return nil, fmt.Errorf("%s", errMsg)
	}

	var mempoolTxs []mempoolTx
	if err := json.NewDecoder(resp.Body).Decode(&mempoolTxs); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	txs := make([]types.Transaction, len(mempoolTxs))
	for i, tx := range mempoolTxs {
		txs[i] = enrichTransaction(tx, address)
	}

	return txs, nil
}

// enrichTransaction calculates send/receive direction and amounts for a transaction
func enrichTransaction(tx mempoolTx, address string) types.Transaction {
	// Check if address appears in inputs (sender) or outputs (receiver)
	isInInputs := false
	for _, vin := range tx.Vin {
		if vin.Prevout != nil && vin.Prevout.ScriptpubkeyAddress == address {
			isInInputs = true
			break
		}
	}

	// Determine transaction type
	// If address is in inputs, it's a send (we're spending)
	// If only in outputs, it's a receive
	txType := "receive"
	if isInInputs {
		txType = "send"
	}

	// Calculate amount
	var amountSats int64
	if txType == "send" {
		// For sends, sum all outputs to other addresses (excluding change back to us)
		for _, vout := range tx.Vout {
			if vout.ScriptpubkeyAddress != address {
				amountSats += vout.Value
			}
		}
	} else {
		// For receives, sum all outputs to our address
		for _, vout := range tx.Vout {
			if vout.ScriptpubkeyAddress == address {
				amountSats += vout.Value
			}
		}
	}

	// Get the other party's address
	var otherAddr string
	if txType == "send" {
		// For sends, get the recipient address (first output that's not us)
		for _, vout := range tx.Vout {
			if vout.ScriptpubkeyAddress != address {
				otherAddr = vout.ScriptpubkeyAddress
				break
			}
		}
	} else {
		// For receives, get the sender address (first input that's not us)
		for _, vin := range tx.Vin {
			if vin.Prevout != nil && vin.Prevout.ScriptpubkeyAddress != address {
				otherAddr = vin.Prevout.ScriptpubkeyAddress
				break
			}
		}
	}

	// Fallback: if we couldn't find the other address, use first available
	if otherAddr == "" {
		if txType == "send" && len(tx.Vout) > 0 {
			otherAddr = tx.Vout[0].ScriptpubkeyAddress
		} else if txType == "receive" && len(tx.Vin) > 0 && tx.Vin[0].Prevout != nil {
			otherAddr = tx.Vin[0].Prevout.ScriptpubkeyAddress
		}
	}

	return types.Transaction{
		Txid:        tx.Txid,
		Type:        txType,
		AmountSats:  amountSats,
		OtherAddr:   otherAddr,
		Confirmed:   tx.Status.Confirmed,
		BlockHeight: tx.Status.BlockHeight,
		BlockTime:   tx.Status.BlockTime,
		FeeSats:     tx.Fee,
	}
}

func (c *Client) GetFeeRates(logCtx *logging.LogContext) (*types.FeeRates, error) {
	url := fmt.Sprintf("%s/api/v1/fees/recommended", c.baseURL)

	resp, err := c.doGet(logCtx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch fee rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("mempool API error (status %d): %s", resp.StatusCode, string(body))
		if logCtx != nil {
			logCtx.Add("upstream_error", errMsg)
		}
		return nil, fmt.Errorf("%s", errMsg)
	}

	var fees types.FeeRates
	if err := json.NewDecoder(resp.Body).Decode(&fees); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &fees, nil
}

func (c *Client) BroadcastTx(logCtx *logging.LogContext, txHex string) (string, error) {
	url := fmt.Sprintf("%s/api/tx", c.baseURL)

	resp, err := c.doPost(logCtx, url, "text/plain", strings.NewReader(txHex))
	if err != nil {
		return "", fmt.Errorf("failed to broadcast transaction: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("broadcast failed (status %d): %s", resp.StatusCode, string(body))
		if logCtx != nil {
			logCtx.Add("upstream_error", errMsg)
		}
		return "", fmt.Errorf("%s", errMsg)
	}

	// mempool.space returns the txid as plain text
	txid := strings.TrimSpace(string(body))
	if logCtx != nil {
		logCtx.Add("txid", txid)
	}
	return txid, nil
}

// Ensure Client implements Provider
var _ chain.Provider = (*Client)(nil)
