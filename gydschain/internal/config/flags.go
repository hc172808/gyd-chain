package config

import (
	"flag"
	"fmt"
	"os"
)

// Flags represents command-line flags
type Flags struct {
	// General
	ConfigFile string
	DataDir    string
	LogLevel   string
	Version    bool
	Help       bool

	// Network
	ListenAddr     string
	BootstrapPeers string
	MaxPeers       int

	// RPC
	RPCEnabled  bool
	RPCAddr     string
	RPCPort     int
	WSEnabled   bool
	WSAddr      string
	WSPort      int
	CORSOrigins string

	// Mining
	MiningEnabled bool
	MinerAddress  string
	MiningThreads int

	// Validator
	ValidatorEnabled bool
	ValidatorKey     string
	Commission       uint64

	// Chain
	ChainID     string
	NetworkID   uint64
	GenesisFile string
}

// ParseFlags parses command-line flags
func ParseFlags() *Flags {
	f := &Flags{}

	// General flags
	flag.StringVar(&f.ConfigFile, "config", "", "Path to configuration file")
	flag.StringVar(&f.DataDir, "datadir", "./data", "Data directory path")
	flag.StringVar(&f.LogLevel, "loglevel", "info", "Log level (debug, info, warn, error)")
	flag.BoolVar(&f.Version, "version", false, "Print version and exit")
	flag.BoolVar(&f.Help, "help", false, "Print help and exit")

	// Network flags
	flag.StringVar(&f.ListenAddr, "listen", "0.0.0.0:30303", "P2P listen address")
	flag.StringVar(&f.BootstrapPeers, "bootstrap", "", "Comma-separated bootstrap peer addresses")
	flag.IntVar(&f.MaxPeers, "maxpeers", 50, "Maximum number of peers")

	// RPC flags
	flag.BoolVar(&f.RPCEnabled, "rpc", true, "Enable HTTP-RPC server")
	flag.StringVar(&f.RPCAddr, "rpcaddr", "127.0.0.1", "HTTP-RPC listen address")
	flag.IntVar(&f.RPCPort, "rpcport", 8545, "HTTP-RPC port")
	flag.BoolVar(&f.WSEnabled, "ws", true, "Enable WebSocket server")
	flag.StringVar(&f.WSAddr, "wsaddr", "127.0.0.1", "WebSocket listen address")
	flag.IntVar(&f.WSPort, "wsport", 8546, "WebSocket port")
	flag.StringVar(&f.CORSOrigins, "cors", "*", "Comma-separated CORS origins")

	// Mining flags
	flag.BoolVar(&f.MiningEnabled, "mine", false, "Enable mining")
	flag.StringVar(&f.MinerAddress, "miner", "", "Miner address for rewards")
	flag.IntVar(&f.MiningThreads, "threads", 1, "Number of mining threads")

	// Validator flags
	flag.BoolVar(&f.ValidatorEnabled, "validator", false, "Run as validator")
	flag.StringVar(&f.ValidatorKey, "validatorkey", "", "Path to validator key file")
	flag.Uint64Var(&f.Commission, "commission", 500, "Validator commission in basis points")

	// Chain flags
	flag.StringVar(&f.ChainID, "chainid", "gydschain-1", "Chain ID")
	flag.Uint64Var(&f.NetworkID, "networkid", 1, "Network ID")
	flag.StringVar(&f.GenesisFile, "genesis", "./genesis.json", "Path to genesis file")

	flag.Parse()

	return f
}

// PrintVersion prints version information
func PrintVersion() {
	fmt.Println("GYDS Chain Node")
	fmt.Println("Version: 0.1.0")
	fmt.Println("Protocol: gyds/1")
}

// PrintUsage prints usage information
func PrintUsage() {
	fmt.Println("Usage: gydschain [options]")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  gydschain --datadir ./node1 --rpcport 8545")
	fmt.Println("  gydschain --validator --validatorkey ./validator.key")
	fmt.Println("  gydschain --mine --miner 0x1234... --threads 4")
}

// ApplyToConfig applies flags to a configuration
func (f *Flags) ApplyToConfig(c *Config) {
	if f.DataDir != "" {
		c.DataDir = f.DataDir
	}
	if f.LogLevel != "" {
		c.LogLevel = f.LogLevel
	}

	// Network
	if f.ListenAddr != "" {
		c.Network.ListenAddr = f.ListenAddr
	}
	if f.MaxPeers > 0 {
		c.Network.MaxPeers = f.MaxPeers
	}

	// RPC
	c.RPC.Enabled = f.RPCEnabled
	if f.RPCAddr != "" {
		c.RPC.HTTPAddr = f.RPCAddr
	}
	if f.RPCPort > 0 {
		c.RPC.HTTPPort = f.RPCPort
	}
	if f.WSAddr != "" {
		c.RPC.WSAddr = f.WSAddr
	}
	if f.WSPort > 0 {
		c.RPC.WSPort = f.WSPort
	}

	// Mining
	c.Mining.Enabled = f.MiningEnabled
	if f.MinerAddress != "" {
		c.Mining.MinerAddress = f.MinerAddress
	}
	if f.MiningThreads > 0 {
		c.Mining.Threads = f.MiningThreads
	}

	// Validator
	c.Validator.Enabled = f.ValidatorEnabled
	if f.ValidatorKey != "" {
		c.Validator.ValidatorKey = f.ValidatorKey
	}
	if f.Commission > 0 {
		c.Validator.Commission = f.Commission
	}

	// Chain
	if f.ChainID != "" {
		c.Chain.ChainID = f.ChainID
	}
	if f.NetworkID > 0 {
		c.Chain.NetworkID = f.NetworkID
	}
	if f.GenesisFile != "" {
		c.Chain.GenesisFile = f.GenesisFile
	}
}

// Validate validates the flags
func (f *Flags) Validate() error {
	if f.MiningEnabled && f.MinerAddress == "" {
		return fmt.Errorf("miner address required when mining is enabled")
	}
	if f.ValidatorEnabled && f.ValidatorKey == "" {
		return fmt.Errorf("validator key required when validator mode is enabled")
	}
	return nil
}

// HandleExit handles version and help flags
func (f *Flags) HandleExit() {
	if f.Version {
		PrintVersion()
		os.Exit(0)
	}
	if f.Help {
		PrintUsage()
		os.Exit(0)
	}
}
