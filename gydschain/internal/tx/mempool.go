package tx

import (
	"container/heap"
	"errors"
	"sync"
	"time"
)

// MempoolConfig contains mempool configuration
type MempoolConfig struct {
	MaxSize       int           `json:"max_size"`
	MaxTxSize     int           `json:"max_tx_size"`
	MaxTxAge      time.Duration `json:"max_tx_age"`
	MinGasPrice   uint64        `json:"min_gas_price"`
	ReapInterval  time.Duration `json:"reap_interval"`
}

// DefaultMempoolConfig returns default configuration
func DefaultMempoolConfig() *MempoolConfig {
	return &MempoolConfig{
		MaxSize:      10000,
		MaxTxSize:    1024 * 1024, // 1MB
		MaxTxAge:     time.Hour,
		MinGasPrice:  1,
		ReapInterval: time.Minute,
	}
}

// Mempool manages pending transactions
type Mempool struct {
	mu       sync.RWMutex
	config   *MempoolConfig
	txs      map[string]*MempoolTx
	queue    *TxQueue
	nonces   map[string]uint64 // address -> highest nonce
	stopChan chan struct{}
}

// MempoolTx wraps a transaction with metadata
type MempoolTx struct {
	Tx        *Transaction
	Hash      string
	AddedAt   time.Time
	GasPrice  uint64
	Priority  int
}

// NewMempool creates a new mempool
func NewMempool(config *MempoolConfig) *Mempool {
	if config == nil {
		config = DefaultMempoolConfig()
	}
	
	mp := &Mempool{
		config:   config,
		txs:      make(map[string]*MempoolTx),
		queue:    &TxQueue{},
		nonces:   make(map[string]uint64),
		stopChan: make(chan struct{}),
	}
	
	heap.Init(mp.queue)
	
	// Start cleanup goroutine
	go mp.cleanupLoop()
	
	return mp
}

// AddTx adds a transaction to the mempool
func (mp *Mempool) AddTx(tx *Transaction) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	
	// Validate transaction
	if err := tx.Verify(); err != nil {
		return err
	}
	
	// Check size
	if tx.Size() > mp.config.MaxTxSize {
		return ErrTxTooLarge
	}
	
	// Check gas price
	gasPrice := tx.Fee / uint64(tx.Size())
	if gasPrice < mp.config.MinGasPrice {
		return ErrGasPriceTooLow
	}
	
	// Get hash
	hash, err := tx.HashHex()
	if err != nil {
		return err
	}
	
	// Check duplicate
	if _, exists := mp.txs[hash]; exists {
		return ErrDuplicateTx
	}
	
	// Check mempool size
	if len(mp.txs) >= mp.config.MaxSize {
		// Try to evict lowest priority tx
		if !mp.evictLowest(gasPrice) {
			return ErrMempoolFull
		}
	}
	
	// Check nonce
	currentNonce := mp.nonces[tx.From]
	if tx.Nonce < currentNonce {
		return ErrNonceTooLow
	}
	
	// Add to mempool
	mtx := &MempoolTx{
		Tx:       tx,
		Hash:     hash,
		AddedAt:  time.Now(),
		GasPrice: gasPrice,
		Priority: int(gasPrice),
	}
	
	mp.txs[hash] = mtx
	heap.Push(mp.queue, mtx)
	
	// Update nonce tracking
	if tx.Nonce >= mp.nonces[tx.From] {
		mp.nonces[tx.From] = tx.Nonce + 1
	}
	
	return nil
}

// RemoveTx removes a transaction from the mempool
func (mp *Mempool) RemoveTx(hash string) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	
	delete(mp.txs, hash)
}

// GetTx returns a transaction by hash
func (mp *Mempool) GetTx(hash string) *Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	
	if mtx, exists := mp.txs[hash]; exists {
		return mtx.Tx
	}
	return nil
}

// HasTx checks if a transaction exists
func (mp *Mempool) HasTx(hash string) bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.txs[hash] != nil
}

// ReapMaxTxs returns up to maxTxs transactions for block inclusion
func (mp *Mempool) ReapMaxTxs(maxTxs int) []*Transaction {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	
	if maxTxs <= 0 {
		maxTxs = mp.config.MaxSize
	}
	
	txs := make([]*Transaction, 0, maxTxs)
	
	for len(txs) < maxTxs && mp.queue.Len() > 0 {
		mtx := heap.Pop(mp.queue).(*MempoolTx)
		
		// Check if still valid
		if time.Since(mtx.AddedAt) > mp.config.MaxTxAge {
			delete(mp.txs, mtx.Hash)
			continue
		}
		
		txs = append(txs, mtx.Tx)
	}
	
	// Re-add to queue (they'll be removed after block is confirmed)
	for _, tx := range txs {
		hash, _ := tx.HashHex()
		if mtx, exists := mp.txs[hash]; exists {
			heap.Push(mp.queue, mtx)
		}
	}
	
	return txs
}

// Update removes confirmed transactions
func (mp *Mempool) Update(confirmedTxs []*Transaction) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	
	for _, tx := range confirmedTxs {
		hash, err := tx.HashHex()
		if err != nil {
			continue
		}
		delete(mp.txs, hash)
	}
	
	// Rebuild queue
	mp.rebuildQueue()
}

// rebuildQueue rebuilds the priority queue
func (mp *Mempool) rebuildQueue() {
	mp.queue = &TxQueue{}
	heap.Init(mp.queue)
	
	for _, mtx := range mp.txs {
		heap.Push(mp.queue, mtx)
	}
}

// evictLowest removes the lowest priority transaction
func (mp *Mempool) evictLowest(minGasPrice uint64) bool {
	if mp.queue.Len() == 0 {
		return false
	}
	
	// Find lowest priority (at end of queue when sorted)
	lowest := (*mp.queue)[mp.queue.Len()-1]
	if lowest.GasPrice >= minGasPrice {
		return false
	}
	
	delete(mp.txs, lowest.Hash)
	mp.rebuildQueue()
	return true
}

// cleanupLoop periodically removes expired transactions
func (mp *Mempool) cleanupLoop() {
	ticker := time.NewTicker(mp.config.ReapInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-mp.stopChan:
			return
		case <-ticker.C:
			mp.cleanup()
		}
	}
}

// cleanup removes expired transactions
func (mp *Mempool) cleanup() {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	
	now := time.Now()
	for hash, mtx := range mp.txs {
		if now.Sub(mtx.AddedAt) > mp.config.MaxTxAge {
			delete(mp.txs, hash)
		}
	}
	
	mp.rebuildQueue()
}

// Size returns the number of transactions
func (mp *Mempool) Size() int {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return len(mp.txs)
}

// TotalBytes returns approximate total size
func (mp *Mempool) TotalBytes() int {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	
	total := 0
	for _, mtx := range mp.txs {
		total += mtx.Tx.Size()
	}
	return total
}

// GetPending returns all pending transactions for an address
func (mp *Mempool) GetPending(address string) []*Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	
	var txs []*Transaction
	for _, mtx := range mp.txs {
		if mtx.Tx.From == address {
			txs = append(txs, mtx.Tx)
		}
	}
	return txs
}

// Stop stops the mempool
func (mp *Mempool) Stop() {
	close(mp.stopChan)
}

// TxQueue implements heap.Interface for priority queue
type TxQueue []*MempoolTx

func (q TxQueue) Len() int { return len(q) }

func (q TxQueue) Less(i, j int) bool {
	return q[i].GasPrice > q[j].GasPrice // Higher gas price = higher priority
}

func (q TxQueue) Swap(i, j int) { q[i], q[j] = q[j], q[i] }

func (q *TxQueue) Push(x interface{}) {
	*q = append(*q, x.(*MempoolTx))
}

func (q *TxQueue) Pop() interface{} {
	old := *q
	n := len(old)
	item := old[n-1]
	*q = old[0 : n-1]
	return item
}

// Mempool errors
var (
	ErrTxTooLarge     = errors.New("transaction too large")
	ErrGasPriceTooLow = errors.New("gas price too low")
	ErrDuplicateTx    = errors.New("duplicate transaction")
	ErrMempoolFull    = errors.New("mempool full")
	ErrNonceTooLow    = errors.New("nonce too low")
)
