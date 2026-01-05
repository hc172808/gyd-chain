package state

import (
	"encoding/json"
	"sync"
)

// StateDB manages the world state
type StateDB struct {
	mu       sync.RWMutex
	accounts map[string]*Account
	assets   map[string]*Asset
	dirty    map[string]bool
	root     string
}

// NewStateDB creates a new state database
func NewStateDB() *StateDB {
	return &StateDB{
		accounts: make(map[string]*Account),
		assets:   make(map[string]*Asset),
		dirty:    make(map[string]bool),
	}
}

// GetAccount returns an account by address
func (s *StateDB) GetAccount(address string) *Account {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	account, exists := s.accounts[address]
	if !exists {
		return nil
	}
	
	return account.Copy()
}

// SetAccount updates or creates an account
func (s *StateDB) SetAccount(address string, account *Account) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.accounts[address] = account.Copy()
	s.dirty[address] = true
}

// DeleteAccount removes an account
func (s *StateDB) DeleteAccount(address string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.accounts, address)
	s.dirty[address] = true
}

// GetBalance returns the balance for an address and asset
func (s *StateDB) GetBalance(address, asset string) uint64 {
	account := s.GetAccount(address)
	if account == nil {
		return 0
	}
	return account.GetBalance(asset)
}

// Transfer moves tokens between accounts
func (s *StateDB) Transfer(from, to, asset string, amount uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Get or create accounts
	sender := s.accounts[from]
	if sender == nil {
		return ErrAccountNotFound
	}
	
	receiver := s.accounts[to]
	if receiver == nil {
		receiver = NewAccount(to)
		s.accounts[to] = receiver
	}
	
	// Check balance
	if sender.Balances[asset] < amount {
		return ErrInsufficientBalance
	}
	
	// Transfer
	sender.Balances[asset] -= amount
	receiver.Balances[asset] += amount
	
	s.dirty[from] = true
	s.dirty[to] = true
	
	return nil
}

// GetAsset returns an asset by ID
func (s *StateDB) GetAsset(id string) *Asset {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.assets[id]
}

// SetAsset updates or creates an asset
func (s *StateDB) SetAsset(id string, asset *Asset) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.assets[id] = asset
}

// Commit finalizes state changes
func (s *StateDB) Commit() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Calculate new state root
	root, err := s.calculateRoot()
	if err != nil {
		return "", err
	}
	
	s.root = root
	s.dirty = make(map[string]bool)
	
	return root, nil
}

// Root returns the current state root
func (s *StateDB) Root() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.root
}

// Snapshot creates a copy of the current state
func (s *StateDB) Snapshot() *StateDB {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	snapshot := NewStateDB()
	
	for addr, account := range s.accounts {
		snapshot.accounts[addr] = account.Copy()
	}
	
	for id, asset := range s.assets {
		snapshot.assets[id] = asset.Copy()
	}
	
	snapshot.root = s.root
	
	return snapshot
}

// Revert restores state from a snapshot
func (s *StateDB) Revert(snapshot *StateDB) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.accounts = snapshot.accounts
	s.assets = snapshot.assets
	s.root = snapshot.root
	s.dirty = make(map[string]bool)
}

// calculateRoot computes the state root hash
func (s *StateDB) calculateRoot() (string, error) {
	// Serialize accounts
	data, err := json.Marshal(s.accounts)
	if err != nil {
		return "", err
	}
	
	// Calculate merkle root (simplified)
	return CalculateMerkleRoot(data), nil
}

// AccountCount returns the number of accounts
func (s *StateDB) AccountCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.accounts)
}

// AssetCount returns the number of assets
func (s *StateDB) AssetCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.assets)
}

// AllAccounts returns all account addresses
func (s *StateDB) AllAccounts() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	addresses := make([]string, 0, len(s.accounts))
	for addr := range s.accounts {
		addresses = append(addresses, addr)
	}
	return addresses
}

// TotalSupply calculates total supply of an asset
func (s *StateDB) TotalSupply(asset string) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var total uint64
	for _, account := range s.accounts {
		total += account.Balances[asset]
		if asset == "GYDS" {
			total += account.Staked
			for _, delegated := range account.Delegated {
				total += delegated
			}
		}
	}
	return total
}

// Export exports the entire state
func (s *StateDB) Export() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	export := struct {
		Accounts map[string]*Account `json:"accounts"`
		Assets   map[string]*Asset   `json:"assets"`
		Root     string              `json:"root"`
	}{
		Accounts: s.accounts,
		Assets:   s.assets,
		Root:     s.root,
	}
	
	return json.Marshal(export)
}

// Errors
var (
	ErrAccountNotFound     = &StateError{"account not found"}
	ErrInsufficientBalance = &StateError{"insufficient balance"}
	ErrAssetNotFound       = &StateError{"asset not found"}
)

type StateError struct {
	msg string
}

func (e *StateError) Error() string {
	return e.msg
}
