package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// LiteNode represents a light client that syncs with the network
type LiteNode struct {
	NodeID         string
	DataDir        string
	BootstrapNodes []string
	SyncMode       string
	CurrentHeight  uint64
	PeerCount      int
	Syncing        bool
	LastSync       time.Time
}

// BootstrapNode represents a peer to sync from
type BootstrapNode struct {
	Address   string `json:"address"`
	PublicKey string `json:"public_key"`
	NodeID    string `json:"node_id"`
}

func main() {
	dataDir := flag.String("datadir", "./data/lite", "Data directory for lite node")
	configPath := flag.String("config", "config/litenode.json", "Path to lite node config")
	syncMode := flag.String("sync-mode", "light", "Sync mode: light or ultralight")
	bootstrapFile := flag.String("bootstrap-nodes", "config/bootstrap.json", "Bootstrap nodes file")
	flag.Parse()

	fmt.Println("🌐 Starting GYDS Chain Lite Node...")
	fmt.Printf("   Data Dir: %s\n", *dataDir)
	fmt.Printf("   Sync Mode: %s\n", *syncMode)

	// Load bootstrap nodes
	bootstrapNodes, err := loadBootstrapNodes(*bootstrapFile)
	if err != nil {
		log.Printf("Warning: Could not load bootstrap nodes: %v", err)
		bootstrapNodes = []BootstrapNode{}
	}

	// Create data directory
	if err := os.MkdirAll(*dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Initialize lite node
	node := &LiteNode{
		NodeID:         generateNodeID(),
		DataDir:        *dataDir,
		SyncMode:       *syncMode,
		CurrentHeight:  0,
		PeerCount:      0,
		Syncing:        false,
	}

	// Load existing state
	node.loadState()

	// Start syncing
	go node.startSync(bootstrapNodes)

	// Start health endpoint
	go node.startHealthServer()

	fmt.Println("\n========================================")
	fmt.Println("   GYDS Chain Lite Node Running")
	fmt.Println("========================================")
	fmt.Printf("   Node ID: %s\n", node.NodeID[:16]+"...")
	fmt.Printf("   Current Height: %d\n", node.CurrentHeight)
	fmt.Printf("   Bootstrap Peers: %d\n", len(bootstrapNodes))
	fmt.Println("========================================")
	fmt.Println("\nPress Ctrl+C to stop the node...")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n🛑 Shutting down Lite Node...")
	node.saveState()
	fmt.Println("✅ Lite Node stopped successfully")
	_ = configPath // config loading placeholder
}

func loadBootstrapNodes(path string) ([]BootstrapNode, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var nodes []BootstrapNode
	if err := json.Unmarshal(data, &nodes); err != nil {
		return nil, err
	}

	return nodes, nil
}

func generateNodeID() string {
	// Generate random node ID
	b := make([]byte, 32)
	for i := range b {
		b[i] = byte(time.Now().UnixNano() >> (i * 8))
	}
	return fmt.Sprintf("%x", b)
}

func (n *LiteNode) loadState() {
	statePath := n.DataDir + "/state.json"
	data, err := ioutil.ReadFile(statePath)
	if err != nil {
		return
	}

	var state struct {
		Height   uint64    `json:"height"`
		LastSync time.Time `json:"last_sync"`
	}

	if err := json.Unmarshal(data, &state); err == nil {
		n.CurrentHeight = state.Height
		n.LastSync = state.LastSync
	}
}

func (n *LiteNode) saveState() {
	state := struct {
		Height   uint64    `json:"height"`
		LastSync time.Time `json:"last_sync"`
	}{
		Height:   n.CurrentHeight,
		LastSync: n.LastSync,
	}

	data, err := json.Marshal(state)
	if err != nil {
		return
	}

	statePath := n.DataDir + "/state.json"
	ioutil.WriteFile(statePath, data, 0644)
}

func (n *LiteNode) startSync(bootstrapNodes []BootstrapNode) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		n.syncHeaders(bootstrapNodes)
	}
}

func (n *LiteNode) syncHeaders(bootstrapNodes []BootstrapNode) {
	if len(bootstrapNodes) == 0 {
		return
	}

	n.Syncing = true
	defer func() { n.Syncing = false }()

	for _, peer := range bootstrapNodes {
		// Fetch latest block height from peer
		resp, err := http.Get(fmt.Sprintf("http://%s/rpc/block/latest", peer.Address))
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		var block struct {
			Height uint64 `json:"height"`
			Hash   string `json:"hash"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&block); err != nil {
			continue
		}

		if block.Height > n.CurrentHeight {
			// Sync headers
			n.syncHeadersFromPeer(peer.Address, n.CurrentHeight, block.Height)
			n.CurrentHeight = block.Height
			n.LastSync = time.Now()
			n.PeerCount = len(bootstrapNodes)
			log.Printf("Synced to height %d from %s", block.Height, peer.Address)
		}
		break
	}
}

func (n *LiteNode) syncHeadersFromPeer(peerAddr string, from, to uint64) {
	// Light sync - only fetch block headers
	batchSize := uint64(100)
	for height := from; height < to; height += batchSize {
		end := height + batchSize
		if end > to {
			end = to
		}

		url := fmt.Sprintf("http://%s/rpc/headers?from=%d&to=%d", peerAddr, height, end)
		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		resp.Body.Close()
	}
}

func (n *LiteNode) startHealthServer() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		status := map[string]interface{}{
			"node_id":        n.NodeID[:16],
			"current_height": n.CurrentHeight,
			"peer_count":     n.PeerCount,
			"syncing":        n.Syncing,
			"last_sync":      n.LastSync,
			"sync_mode":      n.SyncMode,
		}
		json.NewEncoder(w).Encode(status)
	})

	http.ListenAndServe(":8547", nil)
}
