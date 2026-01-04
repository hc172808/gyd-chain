package rpc

import (
	"encoding/json"
	"errors"
	"sync"
)

// MethodHandler is a function that handles an RPC method call
type MethodHandler func(params json.RawMessage) (interface{}, error)

// Methods manages registered RPC methods
type Methods struct {
	handlers map[string]MethodHandler
	mu       sync.RWMutex
}

// NewMethods creates a new Methods instance
func NewMethods() *Methods {
	m := &Methods{
		handlers: make(map[string]MethodHandler),
	}
	m.registerBuiltins()
	return m
}

// Register registers a new method handler
func (m *Methods) Register(name string, handler MethodHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[name] = handler
}

// Call calls a registered method
func (m *Methods) Call(name string, params json.RawMessage) (interface{}, error) {
	m.mu.RLock()
	handler, exists := m.handlers[name]
	m.mu.RUnlock()

	if !exists {
		return nil, errors.New("method not found: " + name)
	}

	return handler(params)
}

// registerBuiltins registers built-in RPC methods
func (m *Methods) registerBuiltins() {
	// Chain methods
	m.Register("chain_getBlockByNumber", m.getBlockByNumber)
	m.Register("chain_getBlockByHash", m.getBlockByHash)
	m.Register("chain_getLatestBlock", m.getLatestBlock)
	m.Register("chain_getBlockHeight", m.getBlockHeight)
	m.Register("chain_getChainInfo", m.getChainInfo)

	// Account methods
	m.Register("account_getBalance", m.getBalance)
	m.Register("account_getNonce", m.getNonce)
	m.Register("account_getAccount", m.getAccount)

	// Transaction methods
	m.Register("tx_sendTransaction", m.sendTransaction)
	m.Register("tx_getTransaction", m.getTransaction)
	m.Register("tx_getTransactionReceipt", m.getTransactionReceipt)
	m.Register("tx_estimateFee", m.estimateFee)
	m.Register("tx_getPendingTransactions", m.getPendingTransactions)

	// Validator methods
	m.Register("validator_getValidators", m.getValidators)
	m.Register("validator_getValidator", m.getValidator)
	m.Register("validator_stake", m.stake)
	m.Register("validator_unstake", m.unstake)

	// Asset methods
	m.Register("asset_getAsset", m.getAsset)
	m.Register("asset_getAssetBalance", m.getAssetBalance)
	m.Register("asset_transfer", m.transferAsset)

	// Network methods
	m.Register("net_getPeers", m.getPeers)
	m.Register("net_getNodeInfo", m.getNodeInfo)

	// Mining methods
	m.Register("mining_getWork", m.getWork)
	m.Register("mining_submitWork", m.submitWork)
	m.Register("mining_getMiningInfo", m.getMiningInfo)
}

// Chain method implementations
func (m *Methods) getBlockByNumber(params json.RawMessage) (interface{}, error) {
	var args struct {
		Number uint64 `json:"number"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return nil, err
	}
	// TODO: Implement block retrieval
	return nil, errors.New("not implemented")
}

func (m *Methods) getBlockByHash(params json.RawMessage) (interface{}, error) {
	var args struct {
		Hash string `json:"hash"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return nil, err
	}
	// TODO: Implement block retrieval by hash
	return nil, errors.New("not implemented")
}

func (m *Methods) getLatestBlock(params json.RawMessage) (interface{}, error) {
	// TODO: Implement latest block retrieval
	return nil, errors.New("not implemented")
}

func (m *Methods) getBlockHeight(params json.RawMessage) (interface{}, error) {
	// TODO: Implement block height retrieval
	return nil, errors.New("not implemented")
}

func (m *Methods) getChainInfo(params json.RawMessage) (interface{}, error) {
	return map[string]interface{}{
		"chainId":   "gydschain-1",
		"networkId": 1,
		"name":      "GYDS Chain",
	}, nil
}

// Account method implementations
func (m *Methods) getBalance(params json.RawMessage) (interface{}, error) {
	var args struct {
		Address string `json:"address"`
		Asset   string `json:"asset,omitempty"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return nil, err
	}
	// TODO: Implement balance retrieval
	return nil, errors.New("not implemented")
}

func (m *Methods) getNonce(params json.RawMessage) (interface{}, error) {
	var args struct {
		Address string `json:"address"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return nil, err
	}
	// TODO: Implement nonce retrieval
	return nil, errors.New("not implemented")
}

func (m *Methods) getAccount(params json.RawMessage) (interface{}, error) {
	var args struct {
		Address string `json:"address"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return nil, err
	}
	// TODO: Implement account retrieval
	return nil, errors.New("not implemented")
}

// Transaction method implementations
func (m *Methods) sendTransaction(params json.RawMessage) (interface{}, error) {
	// TODO: Implement transaction sending
	return nil, errors.New("not implemented")
}

func (m *Methods) getTransaction(params json.RawMessage) (interface{}, error) {
	var args struct {
		Hash string `json:"hash"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return nil, err
	}
	// TODO: Implement transaction retrieval
	return nil, errors.New("not implemented")
}

func (m *Methods) getTransactionReceipt(params json.RawMessage) (interface{}, error) {
	var args struct {
		Hash string `json:"hash"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return nil, err
	}
	// TODO: Implement receipt retrieval
	return nil, errors.New("not implemented")
}

func (m *Methods) estimateFee(params json.RawMessage) (interface{}, error) {
	// TODO: Implement fee estimation
	return nil, errors.New("not implemented")
}

func (m *Methods) getPendingTransactions(params json.RawMessage) (interface{}, error) {
	// TODO: Implement pending tx retrieval
	return nil, errors.New("not implemented")
}

// Validator method implementations
func (m *Methods) getValidators(params json.RawMessage) (interface{}, error) {
	// TODO: Implement validators retrieval
	return nil, errors.New("not implemented")
}

func (m *Methods) getValidator(params json.RawMessage) (interface{}, error) {
	var args struct {
		Address string `json:"address"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return nil, err
	}
	// TODO: Implement validator retrieval
	return nil, errors.New("not implemented")
}

func (m *Methods) stake(params json.RawMessage) (interface{}, error) {
	// TODO: Implement staking
	return nil, errors.New("not implemented")
}

func (m *Methods) unstake(params json.RawMessage) (interface{}, error) {
	// TODO: Implement unstaking
	return nil, errors.New("not implemented")
}

// Asset method implementations
func (m *Methods) getAsset(params json.RawMessage) (interface{}, error) {
	var args struct {
		AssetID string `json:"assetId"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return nil, err
	}
	// TODO: Implement asset retrieval
	return nil, errors.New("not implemented")
}

func (m *Methods) getAssetBalance(params json.RawMessage) (interface{}, error) {
	var args struct {
		Address string `json:"address"`
		AssetID string `json:"assetId"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return nil, err
	}
	// TODO: Implement asset balance retrieval
	return nil, errors.New("not implemented")
}

func (m *Methods) transferAsset(params json.RawMessage) (interface{}, error) {
	// TODO: Implement asset transfer
	return nil, errors.New("not implemented")
}

// Network method implementations
func (m *Methods) getPeers(params json.RawMessage) (interface{}, error) {
	// TODO: Implement peers retrieval
	return nil, errors.New("not implemented")
}

func (m *Methods) getNodeInfo(params json.RawMessage) (interface{}, error) {
	return map[string]interface{}{
		"version":  "0.1.0",
		"protocol": "gyds/1",
	}, nil
}

// Mining method implementations
func (m *Methods) getWork(params json.RawMessage) (interface{}, error) {
	// TODO: Implement mining work retrieval
	return nil, errors.New("not implemented")
}

func (m *Methods) submitWork(params json.RawMessage) (interface{}, error) {
	// TODO: Implement work submission
	return nil, errors.New("not implemented")
}

func (m *Methods) getMiningInfo(params json.RawMessage) (interface{}, error) {
	// TODO: Implement mining info retrieval
	return nil, errors.New("not implemented")
}
