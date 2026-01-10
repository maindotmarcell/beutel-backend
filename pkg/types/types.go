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

// Transaction represents an enriched transaction with send/receive info
type Transaction struct {
	Txid        string `json:"txid"`
	Type        string `json:"type"`                  // "send" or "receive"
	AmountSats  int64  `json:"amountSats"`            // net amount for this address
	OtherAddr   string `json:"otherAddr"`             // recipient (for send) or sender (for receive)
	Confirmed   bool   `json:"confirmed"`
	BlockHeight int64  `json:"blockHeight,omitempty"`
	BlockTime   int64  `json:"blockTime,omitempty"`
	FeeSats     int64  `json:"feeSats"`
}

// FeeRates represents recommended fee rates in sat/vB
type FeeRates struct {
	FastestFee  int `json:"fastestFee"`
	HalfHourFee int `json:"halfHourFee"`
	HourFee     int `json:"hourFee"`
	EconomyFee  int `json:"economyFee"`
	MinimumFee  int `json:"minimumFee"`
}
