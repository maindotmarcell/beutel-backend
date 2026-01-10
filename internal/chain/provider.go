package chain

import "github.com/maindotmarcell/beutel-backend/pkg/types"

// Network represents a Bitcoin network
type Network string

const (
	Mainnet  Network = "mainnet"
	Testnet3 Network = "testnet3"
	Testnet4 Network = "testnet4"
	Signet   Network = "signet"
)

// ParseNetwork converts a string to a Network, defaulting to Mainnet
func ParseNetwork(s string) Network {
	switch s {
	case "testnet3":
		return Testnet3
	case "testnet4":
		return Testnet4
	case "signet":
		return Signet
	default:
		return Mainnet
	}
}

// Provider abstracts blockchain data access.
// Implementations: mempool.space, Electrum, own indexer, etc.
type Provider interface {
	// GetBalance returns confirmed and unconfirmed balance for an address
	GetBalance(address string) (*types.Balance, error)

	// GetUTXOs returns unspent transaction outputs for an address
	GetUTXOs(address string) ([]types.UTXO, error)

	// GetTransactions returns transaction history for an address
	GetTransactions(address string) ([]types.Transaction, error)

	// GetFeeRates returns recommended fee rates
	GetFeeRates() (*types.FeeRates, error)

	// BroadcastTx broadcasts a signed transaction hex and returns txid
	BroadcastTx(txHex string) (string, error)

	// Network returns the configured network
	Network() Network
}
