package types

// Balance represents confirmed and unconfirmed satoshi amounts
type Balance struct {
	Confirmed   int64 `json:"confirmed"`
	Unconfirmed int64 `json:"unconfirmed"`
	Total       int64 `json:"total"`
}

// UTXO represents an unspent transaction output
type UTXO struct {
	Txid        string `json:"txid"`
	Vout        int    `json:"vout"`
	Value       int64  `json:"value"`
	Confirmed   bool   `json:"confirmed"`
	BlockHeight int64  `json:"blockHeight,omitempty"`
}

// Transaction represents a transaction in the address history
type Transaction struct {
	Txid        string `json:"txid"`
	Confirmed   bool   `json:"confirmed"`
	BlockHeight int64  `json:"blockHeight,omitempty"`
	BlockTime   int64  `json:"blockTime,omitempty"`
	Fee         int64  `json:"fee"`
}

// FeeRates represents recommended fee rates in sat/vB
type FeeRates struct {
	FastestFee  int `json:"fastestFee"`
	HalfHourFee int `json:"halfHourFee"`
	HourFee     int `json:"hourFee"`
	EconomyFee  int `json:"economyFee"`
	MinimumFee  int `json:"minimumFee"`
}
