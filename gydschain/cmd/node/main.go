package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gydschain/gydschain/internal/chain"
	"github.com/gydschain/gydschain/internal/config"
	"github.com/gydschain/gydschain/internal/consensus/pos"
	"github.com/gydschain/gydschain/internal/p2p"
	"github.com/gydschain/gydschain/internal/rpc"
	"github.com/gydschain/gydschain/internal/state"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.json", "Path to configuration file")
	genesisPath := flag.String("genesis", "genesis.json", "Path to genesis file")
	dataDir := flag.String("data", "./data", "Data directory")
	rpcAddr := flag.String("rpc", "0.0.0.0:8545", "RPC listen address")
	p2pAddr := flag.String("p2p", "0.0.0.0:26656", "P2P listen address")
	flag.Parse()

	fmt.Println("🚀 Starting GYDS Chain Node...")
	fmt.Printf("   Config: %s\n", *configPath)
	fmt.Printf("   Genesis: %s\n", *genesisPath)
	fmt.Printf("   Data Dir: %s\n", *dataDir)
	fmt.Printf("   RPC: %s\n", *rpcAddr)
	fmt.Printf("   P2P: %s\n", *p2pAddr)

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Printf("Warning: Could not load config, using defaults: %v", err)
		cfg = config.DefaultConfig()
	}

	// Override with command line flags
	cfg.RPC.ListenAddr = *rpcAddr
	cfg.P2P.ListenAddr = *p2pAddr
	cfg.DataDir = *dataDir

	// Initialize state database
	stateDB := state.NewStateDB()
	fmt.Println("✅ State database initialized")

	// Initialize blockchain
	chainConfig := chain.DefaultConfig()
	blockchain, err := chain.NewChain(chainConfig, stateDB)
	if err != nil {
		log.Fatalf("Failed to create chain: %v", err)
	}

	// Load genesis
	genesis, err := chain.LoadGenesis(*genesisPath)
	if err != nil {
		log.Printf("Warning: Could not load genesis, using default: %v", err)
		genesis = chain.DefaultGenesis()
	}

	if err := blockchain.InitGenesis(genesis); err != nil {
		log.Fatalf("Failed to initialize genesis: %v", err)
	}
	fmt.Println("✅ Genesis block initialized")

	// Initialize consensus engine
	posEngine := pos.NewEngine(
		cfg.Consensus.MinStake,
		uint32(cfg.Consensus.MaxValidators),
		cfg.Consensus.BlockTime,
	)
	fmt.Println("✅ PoS consensus engine initialized")

	// Initialize P2P node
	p2pConfig := &p2p.NodeConfig{
		ListenAddr:   cfg.P2P.ListenAddr,
		MaxPeers:     cfg.P2P.MaxPeers,
		DialTimeout:  cfg.P2P.DialTimeout,
		PingInterval: cfg.P2P.PingInterval,
		Seeds:        cfg.P2P.Seeds,
		NetworkID:    cfg.NetworkID,
	}

	p2pNode, err := p2p.NewNode(p2pConfig)
	if err != nil {
		log.Fatalf("Failed to create P2P node: %v", err)
	}

	if err := p2pNode.Start(); err != nil {
		log.Fatalf("Failed to start P2P node: %v", err)
	}
	fmt.Printf("✅ P2P node started on %s\n", cfg.P2P.ListenAddr)

	// Initialize RPC server
	rpcConfig := &rpc.Config{
		ListenAddr:     cfg.RPC.ListenAddr,
		EnableWS:       cfg.RPC.EnableWebSocket,
		MaxConnections: cfg.RPC.MaxConnections,
	}

	rpcServer := rpc.NewServer(rpcConfig, blockchain, posEngine, stateDB)
	if err := rpcServer.Start(); err != nil {
		log.Fatalf("Failed to start RPC server: %v", err)
	}
	fmt.Printf("✅ RPC server started on %s\n", cfg.RPC.ListenAddr)

	// Print node info
	fmt.Println("\n========================================")
	fmt.Println("   GYDS Chain Node Running")
	fmt.Println("========================================")
	fmt.Printf("   Chain ID: %s\n", chainConfig.ChainID)
	fmt.Printf("   Network ID: %d\n", cfg.NetworkID)
	fmt.Printf("   Block Height: %d\n", blockchain.Height())
	fmt.Printf("   Validators: %d\n", posEngine.ValidatorCount())
	fmt.Printf("   Peers: %d\n", p2pNode.PeerCount())
	fmt.Println("========================================")
	fmt.Println("\nPress Ctrl+C to stop the node...")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n🛑 Shutting down GYDS Chain Node...")

	// Graceful shutdown
	rpcServer.Stop()
	p2pNode.Stop()

	fmt.Println("✅ Node stopped successfully")
}
