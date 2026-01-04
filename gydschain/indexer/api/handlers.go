package api

import (
	"encoding/json"
	"net/http"
)

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(data interface{}, total int64, page, perPage int) *PaginatedResponse {
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}
	
	return &PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// WriteError writes an error response
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Code:    status,
		Message: message,
	})
}

// WriteSuccess writes a success response
func WriteSuccess(w http.ResponseWriter, data interface{}, message string) {
	WriteJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// BlockResponse represents a block in the API
type BlockResponse struct {
	Number           uint64   `json:"number"`
	Hash             string   `json:"hash"`
	ParentHash       string   `json:"parent_hash"`
	StateRoot        string   `json:"state_root"`
	TransactionsRoot string   `json:"transactions_root"`
	Validator        string   `json:"validator"`
	Timestamp        uint64   `json:"timestamp"`
	GasUsed          uint64   `json:"gas_used"`
	GasLimit         uint64   `json:"gas_limit"`
	TxCount          int      `json:"tx_count"`
	Size             uint64   `json:"size"`
	Transactions     []string `json:"transactions,omitempty"`
}

// TransactionResponse represents a transaction in the API
type TransactionResponse struct {
	Hash        string  `json:"hash"`
	BlockNumber uint64  `json:"block_number"`
	BlockHash   string  `json:"block_hash"`
	TxIndex     int     `json:"tx_index"`
	From        string  `json:"from"`
	To          *string `json:"to,omitempty"`
	Value       string  `json:"value"`
	Asset       string  `json:"asset"`
	Fee         string  `json:"fee"`
	Nonce       uint64  `json:"nonce"`
	Type        string  `json:"type"`
	Status      string  `json:"status"`
	Timestamp   uint64  `json:"timestamp"`
}

// AccountResponse represents an account in the API
type AccountResponse struct {
	Address        string            `json:"address"`
	Nonce          uint64            `json:"nonce"`
	TxCount        uint64            `json:"tx_count"`
	Balances       map[string]string `json:"balances"`
	FirstSeenBlock uint64            `json:"first_seen_block"`
	LastSeenBlock  uint64            `json:"last_seen_block"`
}

// AssetResponse represents an asset in the API
type AssetResponse struct {
	ID           string `json:"id"`
	Symbol       string `json:"symbol"`
	Name         string `json:"name"`
	Decimals     uint8  `json:"decimals"`
	TotalSupply  string `json:"total_supply"`
	MaxSupply    string `json:"max_supply,omitempty"`
	Creator      string `json:"creator"`
	IsNative     bool   `json:"is_native"`
	IsStablecoin bool   `json:"is_stablecoin"`
	PegTarget    string `json:"peg_target,omitempty"`
	HolderCount  int    `json:"holder_count"`
}

// ValidatorResponse represents a validator in the API
type ValidatorResponse struct {
	Address          string `json:"address"`
	Stake            string `json:"stake"`
	Commission       uint64 `json:"commission"`
	Active           bool   `json:"active"`
	Jailed           bool   `json:"jailed"`
	BlocksProposed   uint64 `json:"blocks_proposed"`
	BlocksSigned     uint64 `json:"blocks_signed"`
	UptimePercentage float64 `json:"uptime_percentage"`
	DelegatorCount   int    `json:"delegator_count"`
	TotalDelegations string `json:"total_delegations"`
}

// StatsResponse represents chain statistics
type StatsResponse struct {
	LastBlock         uint64  `json:"last_block"`
	TotalTransactions uint64  `json:"total_transactions"`
	TotalAccounts     uint64  `json:"total_accounts"`
	TotalValidators   int     `json:"total_validators"`
	TotalStaked       string  `json:"total_staked"`
	GYDSCirculating   string  `json:"gyds_circulating"`
	GYDCirculating    string  `json:"gyd_circulating"`
	AvgBlockTime      float64 `json:"avg_block_time"`
	TPS               float64 `json:"tps"`
}

// SearchResult represents a search result
type SearchResult struct {
	Type    string      `json:"type"` // block, transaction, account, asset
	ID      string      `json:"id"`
	Preview interface{} `json:"preview"`
}

// SearchResponse represents a search response
type SearchResponse struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
	Total   int            `json:"total"`
}
