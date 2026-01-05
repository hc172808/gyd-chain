package state

import (
	"encoding/json"
	"sync"
)

// Account represents a blockchain account
type Account struct {
	mu        sync.RWMutex
	Address   string            `json:"address"`
	Nonce     uint64            `json:"nonce"`
	Balances  map[string]uint64 `json:"balances"`
	Staked    uint64            `json:"staked"`
	Delegated map[string]uint64 `json:"delegated"`
	Code      []byte            `json:"code,omitempty"`
	Storage   map[string][]byte `json:"storage,omitempty"`
	CreatedAt int64             `json:"created_at"`
	UpdatedAt int64             `json:"updated_at"`
}

// NewAccount creates a new account
func NewAccount(address string) *Account {
	return &Account{
		Address:   address,
		Balances:  make(map[string]uint64),
		Delegated: make(map[string]uint64),
		Storage:   make(map[string][]byte),
	}
}

// GetBalance returns the balance for a specific asset
func (a *Account) GetBalance(asset string) uint64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Balances[asset]
}

// SetBalance sets the balance for a specific asset
func (a *Account) SetBalance(asset string, amount uint64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Balances[asset] = amount
}

// AddBalance adds to the balance for a specific asset
func (a *Account) AddBalance(asset string, amount uint64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Balances[asset] += amount
}

// SubBalance subtracts from the balance for a specific asset
func (a *Account) SubBalance(asset string, amount uint64) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if a.Balances[asset] < amount {
		return false
	}
	
	a.Balances[asset] -= amount
	return true
}

// GetNonce returns the current nonce
func (a *Account) GetNonce() uint64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Nonce
}

// IncrementNonce increases the nonce by 1
func (a *Account) IncrementNonce() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Nonce++
}

// Stake locks tokens for staking
func (a *Account) Stake(amount uint64) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if a.Balances["GYDS"] < amount {
		return false
	}
	
	a.Balances["GYDS"] -= amount
	a.Staked += amount
	return true
}

// Unstake unlocks tokens from staking
func (a *Account) Unstake(amount uint64) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if a.Staked < amount {
		return false
	}
	
	a.Staked -= amount
	a.Balances["GYDS"] += amount
	return true
}

// GetStaked returns the staked amount
func (a *Account) GetStaked() uint64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Staked
}

// Delegate delegates stake to a validator
func (a *Account) Delegate(validator string, amount uint64) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if a.Balances["GYDS"] < amount {
		return false
	}
	
	a.Balances["GYDS"] -= amount
	a.Delegated[validator] += amount
	return true
}

// Undelegate removes delegation from a validator
func (a *Account) Undelegate(validator string, amount uint64) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if a.Delegated[validator] < amount {
		return false
	}
	
	a.Delegated[validator] -= amount
	a.Balances["GYDS"] += amount
	return true
}

// GetDelegation returns the delegated amount to a validator
func (a *Account) GetDelegation(validator string) uint64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Delegated[validator]
}

// TotalDelegated returns total delegated amount
func (a *Account) TotalDelegated() uint64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	var total uint64
	for _, amount := range a.Delegated {
		total += amount
	}
	return total
}

// IsContract returns true if account has code
func (a *Account) IsContract() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.Code) > 0
}

// SetCode sets contract code
func (a *Account) SetCode(code []byte) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Code = code
}

// GetCode returns contract code
func (a *Account) GetCode() []byte {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Code
}

// SetStorage sets a storage value
func (a *Account) SetStorage(key string, value []byte) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Storage[key] = value
}

// GetStorage returns a storage value
func (a *Account) GetStorage(key string) []byte {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Storage[key]
}

// Copy creates a deep copy of the account
func (a *Account) Copy() *Account {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	copy := &Account{
		Address:   a.Address,
		Nonce:     a.Nonce,
		Staked:    a.Staked,
		Balances:  make(map[string]uint64),
		Delegated: make(map[string]uint64),
		Storage:   make(map[string][]byte),
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
	
	for k, v := range a.Balances {
		copy.Balances[k] = v
	}
	
	for k, v := range a.Delegated {
		copy.Delegated[k] = v
	}
	
	for k, v := range a.Storage {
		copy.Storage[k] = append([]byte{}, v...)
	}
	
	if a.Code != nil {
		copy.Code = append([]byte{}, a.Code...)
	}
	
	return copy
}

// Serialize converts account to bytes
func (a *Account) Serialize() ([]byte, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return json.Marshal(a)
}

// Deserialize creates an account from bytes
func Deserialize(data []byte) (*Account, error) {
	var account Account
	if err := json.Unmarshal(data, &account); err != nil {
		return nil, err
	}
	
	if account.Balances == nil {
		account.Balances = make(map[string]uint64)
	}
	if account.Delegated == nil {
		account.Delegated = make(map[string]uint64)
	}
	if account.Storage == nil {
		account.Storage = make(map[string][]byte)
	}
	
	return &account, nil
}
