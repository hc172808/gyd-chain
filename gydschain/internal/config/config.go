package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the node configuration
type Config struct {
	// Node identity
	NodeID   string `json:"node_id"`
	DataDir  string `json:"data_dir"`
	LogLevel string `json:"log_level"`

	// Network configuration
	Network NetworkConfig `json:"network"`

	// Chain configuration
	Chain ChainConfig `json:"chain"`

	// RPC configuration
	RPC RPCConfig `json:"rpc"`

	// Mining configuration
	Mining MiningConfig `json:"mining"`

	// Validator configuration
	Validator ValidatorConfig `json:"validator"`

	// Database configuration
	Database DatabaseConfig `json:"database"`
}

// NetworkConfig contains P2P network settings
type NetworkConfig struct {
	ListenAddr     string   `json:"listen_addr"`
	ExternalAddr   string   `json:"external_addr"`
	BootstrapPeers []string `json:"bootstrap_peers"`
	MaxPeers       int      `json:"max_peers"`
	MinPeers       int      `json:"min_peers"`
	EnableNAT      bool     `json:"enable_nat"`
	EnableUPnP     bool     `json:"enable_upnp"`
}

// ChainConfig contains blockchain settings
type ChainConfig struct {
	ChainID         string `json:"chain_id"`
	NetworkID       uint64 `json:"network_id"`
	GenesisFile     string `json:"genesis_file"`
	BlockTime       uint64 `json:"block_time"`       // seconds
	BlockGasLimit   uint64 `json:"block_gas_limit"`
	MinGasPrice     string `json:"min_gas_price"`
	MaxTxPerBlock   int    `json:"max_tx_per_block"`
}

// RPCConfig contains RPC server settings
type RPCConfig struct {
	Enabled       bool     `json:"enabled"`
	HTTPAddr      string   `json:"http_addr"`
	HTTPPort      int      `json:"http_port"`
	WSAddr        string   `json:"ws_addr"`
	WSPort        int      `json:"ws_port"`
	CORSOrigins   []string `json:"cors_origins"`
	EnabledAPIs   []string `json:"enabled_apis"`
	RateLimit     int      `json:"rate_limit"`      // requests per second
	MaxBatchSize  int      `json:"max_batch_size"`
}

// MiningConfig contains mining settings
type MiningConfig struct {
	Enabled      bool   `json:"enabled"`
	MinerAddress string `json:"miner_address"`
	Threads      int    `json:"threads"`
	ExtraData    string `json:"extra_data"`
	PoolMode     bool   `json:"pool_mode"`
	PoolAddr     string `json:"pool_addr"`
}

// ValidatorConfig contains validator settings
type ValidatorConfig struct {
	Enabled        bool   `json:"enabled"`
	ValidatorKey   string `json:"validator_key"`
	Commission     uint64 `json:"commission"` // basis points (100 = 1%)
	MinStake       string `json:"min_stake"`
	AutoCompound   bool   `json:"auto_compound"`
}

// DatabaseConfig contains database settings
type DatabaseConfig struct {
	Engine      string `json:"engine"` // leveldb, badger, rocksdb
	Path        string `json:"path"`
	CacheSize   int    `json:"cache_size"` // MB
	Compression bool   `json:"compression"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		NodeID:   "",
		DataDir:  "./data",
		LogLevel: "info",
		Network: NetworkConfig{
			ListenAddr:     "0.0.0.0:30303",
			ExternalAddr:   "",
			BootstrapPeers: []string{},
			MaxPeers:       50,
			MinPeers:       10,
			EnableNAT:      true,
			EnableUPnP:     true,
		},
		Chain: ChainConfig{
			ChainID:       "gydschain-1",
			NetworkID:     1,
			GenesisFile:   "./genesis.json",
			BlockTime:     5,
			BlockGasLimit: 10000000,
			MinGasPrice:   "1000000000", // 1 gwei
			MaxTxPerBlock: 1000,
		},
		RPC: RPCConfig{
			Enabled:      true,
			HTTPAddr:     "127.0.0.1",
			HTTPPort:     8545,
			WSAddr:       "127.0.0.1",
			WSPort:       8546,
			CORSOrigins:  []string{"*"},
			EnabledAPIs:  []string{"chain", "account", "tx", "net"},
			RateLimit:    100,
			MaxBatchSize: 100,
		},
		Mining: MiningConfig{
			Enabled:      false,
			MinerAddress: "",
			Threads:      1,
			ExtraData:    "",
			PoolMode:     false,
			PoolAddr:     "",
		},
		Validator: ValidatorConfig{
			Enabled:      false,
			ValidatorKey: "",
			Commission:   500, // 5%
			MinStake:     "10000000000000000000000", // 10000 GYDS
			AutoCompound: true,
		},
		Database: DatabaseConfig{
			Engine:      "leveldb",
			Path:        "./data/db",
			CacheSize:   256,
			Compression: true,
		},
	}
}

// LoadConfig loads configuration from a file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := DefaultConfig()
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// SaveConfig saves configuration to a file
func (c *Config) SaveConfig(path string) error {
	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// TODO: Add validation logic
	return nil
}

// GetDataPath returns the full path for a data subdirectory
func (c *Config) GetDataPath(subdir string) string {
	return filepath.Join(c.DataDir, subdir)
}

// GetDatabasePath returns the database path
func (c *Config) GetDatabasePath() string {
	if filepath.IsAbs(c.Database.Path) {
		return c.Database.Path
	}
	return filepath.Join(c.DataDir, c.Database.Path)
}
