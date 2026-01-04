package rpc

import "encoding/json"

// Request represents a JSON-RPC request
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

// Response represents a JSON-RPC response
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      interface{} `json:"id,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Standard JSON-RPC error codes
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// Custom error codes (application-specific)
const (
	ErrBlockNotFound       = -32000
	ErrTxNotFound          = -32001
	ErrAccountNotFound     = -32002
	ErrInsufficientBalance = -32003
	ErrInvalidSignature    = -32004
	ErrNonceTooLow         = -32005
	ErrNonceTooHigh        = -32006
	ErrTxPoolFull          = -32007
	ErrValidatorNotFound   = -32008
	ErrAlreadyStaked       = -32009
	ErrNotStaked           = -32010
	ErrMinimumStake        = -32011
)

// BlockResponse represents a block in RPC responses
type BlockResponse struct {
	Number           uint64              `json:"number"`
	Hash             string              `json:"hash"`
	ParentHash       string              `json:"parentHash"`
	Timestamp        uint64              `json:"timestamp"`
	Validator        string              `json:"validator"`
	StateRoot        string              `json:"stateRoot"`
	TransactionsRoot string              `json:"transactionsRoot"`
	ReceiptsRoot     string              `json:"receiptsRoot"`
	Transactions     []string            `json:"transactions,omitempty"`
	FullTransactions []TransactionResponse `json:"fullTransactions,omitempty"`
	Size             uint64              `json:"size"`
	GasUsed          uint64              `json:"gasUsed"`
	GasLimit         uint64              `json:"gasLimit"`
}

// TransactionResponse represents a transaction in RPC responses
type TransactionResponse struct {
	Hash        string `json:"hash"`
	Nonce       uint64 `json:"nonce"`
	BlockHash   string `json:"blockHash,omitempty"`
	BlockNumber uint64 `json:"blockNumber,omitempty"`
	TxIndex     uint64 `json:"transactionIndex,omitempty"`
	From        string `json:"from"`
	To          string `json:"to,omitempty"`
	Value       string `json:"value"`
	Asset       string `json:"asset"`
	Fee         string `json:"fee"`
	Data        string `json:"data,omitempty"`
	Signature   string `json:"signature"`
	Type        string `json:"type"`
}

// TransactionReceiptResponse represents a transaction receipt
type TransactionReceiptResponse struct {
	TransactionHash string      `json:"transactionHash"`
	BlockHash       string      `json:"blockHash"`
	BlockNumber     uint64      `json:"blockNumber"`
	TxIndex         uint64      `json:"transactionIndex"`
	From            string      `json:"from"`
	To              string      `json:"to,omitempty"`
	Status          uint64      `json:"status"` // 1 = success, 0 = failure
	GasUsed         uint64      `json:"gasUsed"`
	Logs            []LogResponse `json:"logs"`
}

// LogResponse represents a log entry
type LogResponse struct {
	Address     string   `json:"address"`
	Topics      []string `json:"topics"`
	Data        string   `json:"data"`
	BlockNumber uint64   `json:"blockNumber"`
	TxHash      string   `json:"transactionHash"`
	TxIndex     uint64   `json:"transactionIndex"`
	BlockHash   string   `json:"blockHash"`
	LogIndex    uint64   `json:"logIndex"`
}

// AccountResponse represents an account in RPC responses
type AccountResponse struct {
	Address  string            `json:"address"`
	Nonce    uint64            `json:"nonce"`
	Balances map[string]string `json:"balances"` // asset -> balance
}

// ValidatorResponse represents a validator in RPC responses
type ValidatorResponse struct {
	Address          string `json:"address"`
	Stake            string `json:"stake"`
	Commission       uint64 `json:"commission"` // basis points
	Active           bool   `json:"active"`
	Jailed           bool   `json:"jailed"`
	BlocksProposed   uint64 `json:"blocksProposed"`
	BlocksSigned     uint64 `json:"blocksSigned"`
	SlashingEvents   uint64 `json:"slashingEvents"`
	DelegatorCount   uint64 `json:"delegatorCount"`
	TotalDelegations string `json:"totalDelegations"`
}

// AssetResponse represents an asset in RPC responses
type AssetResponse struct {
	ID           string `json:"id"`
	Symbol       string `json:"symbol"`
	Name         string `json:"name"`
	Decimals     uint8  `json:"decimals"`
	TotalSupply  string `json:"totalSupply"`
	MaxSupply    string `json:"maxSupply,omitempty"`
	Mintable     bool   `json:"mintable"`
	Burnable     bool   `json:"burnable"`
	Creator      string `json:"creator"`
	IsStablecoin bool   `json:"isStablecoin"`
	PegTarget    string `json:"pegTarget,omitempty"`
}

// PeerResponse represents a peer in RPC responses
type PeerResponse struct {
	ID        string `json:"id"`
	Address   string `json:"address"`
	Direction string `json:"direction"` // inbound/outbound
	Latency   uint64 `json:"latency"`   // ms
	Version   string `json:"version"`
}

// SyncStatusResponse represents sync status
type SyncStatusResponse struct {
	Syncing       bool   `json:"syncing"`
	CurrentBlock  uint64 `json:"currentBlock"`
	HighestBlock  uint64 `json:"highestBlock"`
	StartingBlock uint64 `json:"startingBlock"`
}

// MiningInfoResponse represents mining information
type MiningInfoResponse struct {
	Mining          bool   `json:"mining"`
	Hashrate        uint64 `json:"hashrate"`
	Difficulty      string `json:"difficulty"`
	CurrentBlock    uint64 `json:"currentBlock"`
	PendingTxCount  uint64 `json:"pendingTxCount"`
	MinerAddress    string `json:"minerAddress,omitempty"`
	RewardPerBlock  string `json:"rewardPerBlock"`
}

// WorkResponse represents mining work
type WorkResponse struct {
	BlockHeader string `json:"blockHeader"`
	Target      string `json:"target"`
	Height      uint64 `json:"height"`
}
