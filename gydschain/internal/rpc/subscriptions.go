package rpc

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// SubscriptionType represents different types of subscriptions
type SubscriptionType string

const (
	SubNewBlock       SubscriptionType = "newBlock"
	SubNewTransaction SubscriptionType = "newTransaction"
	SubPendingTx      SubscriptionType = "pendingTransaction"
	SubLogs           SubscriptionType = "logs"
	SubSyncing        SubscriptionType = "syncing"
)

// Subscription represents an active subscription
type Subscription struct {
	ID       string
	Type     SubscriptionType
	ClientID string
	Filter   interface{} // Optional filter criteria
}

// Client represents a connected WebSocket client
type Client struct {
	ID            string
	Conn          *websocket.Conn
	Subscriptions map[string]*Subscription
	mu            sync.RWMutex
}

// SubscriptionManager manages WebSocket subscriptions
type SubscriptionManager struct {
	clients map[string]*Client
	subs    map[SubscriptionType]map[string]*Subscription // type -> subID -> sub
	mu      sync.RWMutex
}

// NewSubscriptionManager creates a new subscription manager
func NewSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		clients: make(map[string]*Client),
		subs:    make(map[SubscriptionType]map[string]*Subscription),
	}
}

// AddClient adds a new WebSocket client
func (sm *SubscriptionManager) AddClient(conn *websocket.Conn) string {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	clientID := uuid.New().String()
	sm.clients[clientID] = &Client{
		ID:            clientID,
		Conn:          conn,
		Subscriptions: make(map[string]*Subscription),
	}

	return clientID
}

// RemoveClient removes a client and all its subscriptions
func (sm *SubscriptionManager) RemoveClient(clientID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	client, exists := sm.clients[clientID]
	if !exists {
		return
	}

	// Remove all subscriptions for this client
	for subID, sub := range client.Subscriptions {
		if typeSubs, ok := sm.subs[sub.Type]; ok {
			delete(typeSubs, subID)
		}
	}

	delete(sm.clients, clientID)
}

// Subscribe creates a new subscription
func (sm *SubscriptionManager) Subscribe(clientID string, subType SubscriptionType, filter interface{}) (string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	client, exists := sm.clients[clientID]
	if !exists {
		return "", nil
	}

	subID := uuid.New().String()
	sub := &Subscription{
		ID:       subID,
		Type:     subType,
		ClientID: clientID,
		Filter:   filter,
	}

	// Add to client's subscriptions
	client.mu.Lock()
	client.Subscriptions[subID] = sub
	client.mu.Unlock()

	// Add to type-based index
	if _, ok := sm.subs[subType]; !ok {
		sm.subs[subType] = make(map[string]*Subscription)
	}
	sm.subs[subType][subID] = sub

	return subID, nil
}

// Unsubscribe removes a subscription
func (sm *SubscriptionManager) Unsubscribe(clientID string, subID string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	client, exists := sm.clients[clientID]
	if !exists {
		return false
	}

	client.mu.Lock()
	sub, exists := client.Subscriptions[subID]
	if exists {
		delete(client.Subscriptions, subID)
	}
	client.mu.Unlock()

	if !exists {
		return false
	}

	// Remove from type-based index
	if typeSubs, ok := sm.subs[sub.Type]; ok {
		delete(typeSubs, subID)
	}

	return true
}

// Broadcast sends data to all subscribers of a specific type
func (sm *SubscriptionManager) Broadcast(subType string, data interface{}) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	typeSubs, exists := sm.subs[SubscriptionType(subType)]
	if !exists {
		return
	}

	for _, sub := range typeSubs {
		client, exists := sm.clients[sub.ClientID]
		if !exists {
			continue
		}

		// Send notification
		notification := map[string]interface{}{
			"jsonrpc": "2.0",
			"method":  "subscription",
			"params": map[string]interface{}{
				"subscription": sub.ID,
				"result":       data,
			},
		}

		client.Conn.WriteJSON(notification)
	}
}

// BroadcastToClient sends data to a specific client
func (sm *SubscriptionManager) BroadcastToClient(clientID string, subID string, data interface{}) {
	sm.mu.RLock()
	client, exists := sm.clients[clientID]
	sm.mu.RUnlock()

	if !exists {
		return
	}

	notification := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "subscription",
		"params": map[string]interface{}{
			"subscription": subID,
			"result":       data,
		},
	}

	client.Conn.WriteJSON(notification)
}

// GetSubscriptionCount returns the number of active subscriptions
func (sm *SubscriptionManager) GetSubscriptionCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	count := 0
	for _, typeSubs := range sm.subs {
		count += len(typeSubs)
	}
	return count
}

// GetClientCount returns the number of connected clients
func (sm *SubscriptionManager) GetClientCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.clients)
}
