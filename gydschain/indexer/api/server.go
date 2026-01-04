package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gydschain/gydschain/indexer/service"
)

// Server represents the indexer API server
type Server struct {
	addr    string
	router  *mux.Router
	server  *http.Server
	db      *sql.DB
	indexer *service.Indexer
	
	// Sub-handlers
	accounts *service.AccountIndexer
	assets   *service.AssetIndexer
	txs      *service.TransactionIndexer
}

// NewServer creates a new API server
func NewServer(addr string, db *sql.DB, indexer *service.Indexer) *Server {
	s := &Server{
		addr:     addr,
		router:   mux.NewRouter(),
		db:       db,
		indexer:  indexer,
		accounts: service.NewAccountIndexer(db),
		assets:   service.NewAssetIndexer(db),
		txs:      service.NewTransactionIndexer(db),
	}
	s.setupRoutes()
	return s
}

// setupRoutes configures API routes
func (s *Server) setupRoutes() {
	// Health check
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")
	s.router.HandleFunc("/status", s.handleStatus).Methods("GET")
	
	// Blocks
	s.router.HandleFunc("/blocks", s.handleGetBlocks).Methods("GET")
	s.router.HandleFunc("/blocks/{number}", s.handleGetBlock).Methods("GET")
	s.router.HandleFunc("/blocks/{number}/transactions", s.handleGetBlockTransactions).Methods("GET")
	
	// Transactions
	s.router.HandleFunc("/transactions", s.handleGetTransactions).Methods("GET")
	s.router.HandleFunc("/transactions/{hash}", s.handleGetTransaction).Methods("GET")
	
	// Accounts
	s.router.HandleFunc("/accounts/{address}", s.handleGetAccount).Methods("GET")
	s.router.HandleFunc("/accounts/{address}/transactions", s.handleGetAccountTransactions).Methods("GET")
	s.router.HandleFunc("/accounts/{address}/balance", s.handleGetAccountBalance).Methods("GET")
	
	// Assets
	s.router.HandleFunc("/assets", s.handleGetAssets).Methods("GET")
	s.router.HandleFunc("/assets/{id}", s.handleGetAsset).Methods("GET")
	s.router.HandleFunc("/assets/{id}/holders", s.handleGetAssetHolders).Methods("GET")
	s.router.HandleFunc("/assets/{id}/transfers", s.handleGetAssetTransfers).Methods("GET")
	
	// Validators
	s.router.HandleFunc("/validators", s.handleGetValidators).Methods("GET")
	s.router.HandleFunc("/validators/{address}", s.handleGetValidator).Methods("GET")
	
	// Stats
	s.router.HandleFunc("/stats", s.handleGetStats).Methods("GET")
	s.router.HandleFunc("/stats/daily", s.handleGetDailyStats).Methods("GET")
	
	// Search
	s.router.HandleFunc("/search", s.handleSearch).Methods("GET")
	
	// Apply middleware
	s.router.Use(corsMiddleware)
	s.router.Use(loggingMiddleware)
}

// Start starts the API server
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:    s.addr,
		Handler: s.router,
	}
	fmt.Printf("Indexer API server starting on %s\n", s.addr)
	return s.server.ListenAndServe()
}

// Stop stops the API server
func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// Response helpers

func (s *Server) jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *Server) errorResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Health handlers

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.jsonResponse(w, map[string]string{"status": "healthy"})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.jsonResponse(w, map[string]interface{}{
		"status":             "running",
		"last_indexed_block": s.indexer.GetLastIndexedBlock(),
	})
}

// Block handlers

func (s *Server) handleGetBlocks(w http.ResponseWriter, r *http.Request) {
	limit := s.getIntParam(r, "limit", 20)
	offset := s.getIntParam(r, "offset", 0)
	
	rows, err := s.db.Query(`
		SELECT number, hash, parent_hash, validator, timestamp, tx_count, gas_used
		FROM blocks
		ORDER BY number DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		s.errorResponse(w, 500, err.Error())
		return
	}
	defer rows.Close()
	
	var blocks []map[string]interface{}
	for rows.Next() {
		block := make(map[string]interface{})
		var number, timestamp, txCount, gasUsed uint64
		var hash, parentHash, validator string
		rows.Scan(&number, &hash, &parentHash, &validator, &timestamp, &txCount, &gasUsed)
		block["number"] = number
		block["hash"] = hash
		block["parent_hash"] = parentHash
		block["validator"] = validator
		block["timestamp"] = timestamp
		block["tx_count"] = txCount
		block["gas_used"] = gasUsed
		blocks = append(blocks, block)
	}
	
	s.jsonResponse(w, blocks)
}

func (s *Server) handleGetBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	number, _ := strconv.ParseUint(vars["number"], 10, 64)
	
	var block map[string]interface{}
	// Query block from database
	// ...
	
	s.jsonResponse(w, block)
}

func (s *Server) handleGetBlockTransactions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	number, _ := strconv.ParseUint(vars["number"], 10, 64)
	
	txs, err := s.txs.GetTransactionsByBlock(number)
	if err != nil {
		s.errorResponse(w, 500, err.Error())
		return
	}
	
	s.jsonResponse(w, txs)
}

// Transaction handlers

func (s *Server) handleGetTransactions(w http.ResponseWriter, r *http.Request) {
	limit := s.getIntParam(r, "limit", 20)
	
	txs, err := s.txs.GetRecentTransactions(limit)
	if err != nil {
		s.errorResponse(w, 500, err.Error())
		return
	}
	
	s.jsonResponse(w, txs)
}

func (s *Server) handleGetTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]
	
	txn, err := s.txs.GetTransaction(hash)
	if err != nil {
		s.errorResponse(w, 500, err.Error())
		return
	}
	if txn == nil {
		s.errorResponse(w, 404, "transaction not found")
		return
	}
	
	s.jsonResponse(w, txn)
}

// Account handlers

func (s *Server) handleGetAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	
	account, err := s.accounts.GetAccount(address)
	if err != nil {
		s.errorResponse(w, 500, err.Error())
		return
	}
	if account == nil {
		s.errorResponse(w, 404, "account not found")
		return
	}
	
	s.jsonResponse(w, account)
}

func (s *Server) handleGetAccountTransactions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	limit := s.getIntParam(r, "limit", 20)
	offset := s.getIntParam(r, "offset", 0)
	
	txs, err := s.accounts.GetAccountTransactions(address, limit, offset)
	if err != nil {
		s.errorResponse(w, 500, err.Error())
		return
	}
	
	s.jsonResponse(w, txs)
}

func (s *Server) handleGetAccountBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	asset := r.URL.Query().Get("asset")
	if asset == "" {
		asset = "GYDS"
	}
	
	balance, err := s.accounts.GetAccountBalance(address, asset)
	if err != nil {
		s.errorResponse(w, 500, err.Error())
		return
	}
	
	s.jsonResponse(w, map[string]string{
		"address": address,
		"asset":   asset,
		"balance": balance,
	})
}

// Asset handlers

func (s *Server) handleGetAssets(w http.ResponseWriter, r *http.Request) {
	assets, err := s.assets.GetAllAssets()
	if err != nil {
		s.errorResponse(w, 500, err.Error())
		return
	}
	
	s.jsonResponse(w, assets)
}

func (s *Server) handleGetAsset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	asset, err := s.assets.GetAsset(id)
	if err != nil {
		s.errorResponse(w, 500, err.Error())
		return
	}
	if asset == nil {
		s.errorResponse(w, 404, "asset not found")
		return
	}
	
	s.jsonResponse(w, asset)
}

func (s *Server) handleGetAssetHolders(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	limit := s.getIntParam(r, "limit", 20)
	offset := s.getIntParam(r, "offset", 0)
	
	holders, err := s.assets.GetAssetHolders(id, limit, offset)
	if err != nil {
		s.errorResponse(w, 500, err.Error())
		return
	}
	
	s.jsonResponse(w, holders)
}

func (s *Server) handleGetAssetTransfers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	limit := s.getIntParam(r, "limit", 20)
	offset := s.getIntParam(r, "offset", 0)
	
	transfers, err := s.assets.GetAssetTransfers(id, limit, offset)
	if err != nil {
		s.errorResponse(w, 500, err.Error())
		return
	}
	
	s.jsonResponse(w, transfers)
}

// Validator handlers

func (s *Server) handleGetValidators(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	s.jsonResponse(w, []interface{}{})
}

func (s *Server) handleGetValidator(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	s.jsonResponse(w, nil)
}

// Stats handlers

func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	txCount, _ := s.txs.GetTransactionCount()
	
	s.jsonResponse(w, map[string]interface{}{
		"last_block":       s.indexer.GetLastIndexedBlock(),
		"total_transactions": txCount,
	})
}

func (s *Server) handleGetDailyStats(w http.ResponseWriter, r *http.Request) {
	days := s.getIntParam(r, "days", 7)
	
	stats, err := s.txs.GetDailyTransactionStats(days)
	if err != nil {
		s.errorResponse(w, 500, err.Error())
		return
	}
	
	s.jsonResponse(w, stats)
}

// Search handler

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		s.errorResponse(w, 400, "query required")
		return
	}
	
	// Try to match query to block, tx, or account
	// TODO: Implement search logic
	
	s.jsonResponse(w, map[string]interface{}{
		"query":   query,
		"results": []interface{}{},
	})
}

// Helpers

func (s *Server) getIntParam(r *http.Request, name string, defaultVal int) int {
	val := r.URL.Query().Get(name)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}

// Middleware

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
