package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Server represents the JSON-RPC server
type Server struct {
	addr       string
	router     *mux.Router
	httpServer *http.Server
	methods    *Methods
	subs       *SubscriptionManager
	upgrader   websocket.Upgrader
	mu         sync.RWMutex
}

// NewServer creates a new RPC server
func NewServer(addr string) *Server {
	s := &Server{
		addr:    addr,
		router:  mux.NewRouter(),
		methods: NewMethods(),
		subs:    NewSubscriptionManager(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
		},
	}
	s.setupRoutes()
	return s
}

// setupRoutes configures HTTP routes
func (s *Server) setupRoutes() {
	s.router.HandleFunc("/", s.handleRPC).Methods("POST")
	s.router.HandleFunc("/ws", s.handleWebSocket)
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")
}

// Start starts the RPC server
func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:    s.addr,
		Handler: s.router,
	}
	fmt.Printf("RPC server starting on %s\n", s.addr)
	return s.httpServer.ListenAndServe()
}

// Stop gracefully stops the server
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// handleRPC handles JSON-RPC requests
func (s *Server) handleRPC(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, nil, -32700, "Parse error")
		return
	}

	result, err := s.methods.Call(req.Method, req.Params)
	if err != nil {
		s.writeError(w, req.ID, -32601, err.Error())
		return
	}

	s.writeResult(w, req.ID, result)
}

// handleWebSocket handles WebSocket connections for subscriptions
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	clientID := s.subs.AddClient(conn)
	defer s.subs.RemoveClient(clientID)

	for {
		var req Request
		if err := conn.ReadJSON(&req); err != nil {
			break
		}

		switch req.Method {
		case "subscribe":
			s.handleSubscribe(clientID, req)
		case "unsubscribe":
			s.handleUnsubscribe(clientID, req)
		default:
			result, err := s.methods.Call(req.Method, req.Params)
			if err != nil {
				conn.WriteJSON(Response{
					JSONRPC: "2.0",
					ID:      req.ID,
					Error:   &RPCError{Code: -32601, Message: err.Error()},
				})
			} else {
				conn.WriteJSON(Response{
					JSONRPC: "2.0",
					ID:      req.ID,
					Result:  result,
				})
			}
		}
	}
}

// handleSubscribe handles subscription requests
func (s *Server) handleSubscribe(clientID string, req Request) {
	// Parse subscription type from params
	// Add subscription for client
}

// handleUnsubscribe handles unsubscription requests
func (s *Server) handleUnsubscribe(clientID string, req Request) {
	// Remove subscription for client
}

// handleHealth returns server health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

// writeResult writes a successful response
func (s *Server) writeResult(w http.ResponseWriter, id interface{}, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

// writeError writes an error response
func (s *Server) writeError(w http.ResponseWriter, id interface{}, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &RPCError{Code: code, Message: message},
	})
}

// RegisterMethod registers a new RPC method
func (s *Server) RegisterMethod(name string, handler MethodHandler) {
	s.methods.Register(name, handler)
}

// BroadcastBlock broadcasts a new block to subscribers
func (s *Server) BroadcastBlock(block interface{}) {
	s.subs.Broadcast("newBlock", block)
}

// BroadcastTransaction broadcasts a new transaction to subscribers
func (s *Server) BroadcastTransaction(tx interface{}) {
	s.subs.Broadcast("newTransaction", tx)
}
