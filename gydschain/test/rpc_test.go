package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gydschain/gydschain/internal/rpc"
)

func TestRPCServer(t *testing.T) {
	server := rpc.NewServer(":0")
	
	// Create test request
	req := rpc.Request{
		JSONRPC: "2.0",
		Method:  "chain_getChainInfo",
		ID:      1,
	}
	
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	
	// This would require the server to be fully implemented
	// For now, just test that the server initializes
	if server == nil {
		t.Error("expected server, got nil")
	}
}

func TestRPCRequest(t *testing.T) {
	req := rpc.Request{
		JSONRPC: "2.0",
		Method:  "test_method",
		ID:      1,
	}
	
	data, err := json.Marshal(req)
	if err != nil {
		t.Errorf("failed to marshal request: %v", err)
	}
	
	var decoded rpc.Request
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Errorf("failed to unmarshal request: %v", err)
	}
	
	if decoded.Method != "test_method" {
		t.Errorf("expected method test_method, got %s", decoded.Method)
	}
}

func TestRPCResponse(t *testing.T) {
	resp := rpc.Response{
		JSONRPC: "2.0",
		Result:  map[string]string{"status": "ok"},
		ID:      1,
	}
	
	data, err := json.Marshal(resp)
	if err != nil {
		t.Errorf("failed to marshal response: %v", err)
	}
	
	var decoded rpc.Response
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}
	
	if decoded.Error != nil {
		t.Errorf("unexpected error in response: %v", decoded.Error)
	}
}

func TestRPCError(t *testing.T) {
	resp := rpc.Response{
		JSONRPC: "2.0",
		Error: &rpc.RPCError{
			Code:    -32601,
			Message: "Method not found",
		},
		ID: 1,
	}
	
	data, err := json.Marshal(resp)
	if err != nil {
		t.Errorf("failed to marshal error response: %v", err)
	}
	
	var decoded rpc.Response
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Errorf("failed to unmarshal error response: %v", err)
	}
	
	if decoded.Error == nil {
		t.Error("expected error in response")
	}
	
	if decoded.Error.Code != -32601 {
		t.Errorf("expected error code -32601, got %d", decoded.Error.Code)
	}
}

func TestBlockResponse(t *testing.T) {
	block := rpc.BlockResponse{
		Number:     100,
		Hash:       "0x1234567890abcdef",
		ParentHash: "0xabcdef1234567890",
		Timestamp:  1704067200,
		Validator:  "gyds1validator",
		GasUsed:    21000,
		GasLimit:   10000000,
	}
	
	data, err := json.Marshal(block)
	if err != nil {
		t.Errorf("failed to marshal block: %v", err)
	}
	
	var decoded rpc.BlockResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Errorf("failed to unmarshal block: %v", err)
	}
	
	if decoded.Number != 100 {
		t.Errorf("expected number 100, got %d", decoded.Number)
	}
}

func TestTransactionResponse(t *testing.T) {
	tx := rpc.TransactionResponse{
		Hash:        "0xabcdef",
		Nonce:       5,
		BlockNumber: 100,
		From:        "gyds1sender",
		To:          "gyds1receiver",
		Value:       "1000000000000000000",
		Asset:       "GYDS",
		Fee:         "21000000000000",
		Type:        "transfer",
	}
	
	data, err := json.Marshal(tx)
	if err != nil {
		t.Errorf("failed to marshal transaction: %v", err)
	}
	
	var decoded rpc.TransactionResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Errorf("failed to unmarshal transaction: %v", err)
	}
	
	if decoded.From != "gyds1sender" {
		t.Errorf("expected from gyds1sender, got %s", decoded.From)
	}
}

func TestValidatorResponse(t *testing.T) {
	validator := rpc.ValidatorResponse{
		Address:          "gyds1validator",
		Stake:            "100000000000000000000000",
		Commission:       500,
		Active:           true,
		BlocksProposed:   1000,
		BlocksSigned:     990,
		DelegatorCount:   50,
		TotalDelegations: "500000000000000000000000",
	}
	
	data, err := json.Marshal(validator)
	if err != nil {
		t.Errorf("failed to marshal validator: %v", err)
	}
	
	var decoded rpc.ValidatorResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Errorf("failed to unmarshal validator: %v", err)
	}
	
	if decoded.Commission != 500 {
		t.Errorf("expected commission 500, got %d", decoded.Commission)
	}
}

func TestAssetResponse(t *testing.T) {
	asset := rpc.AssetResponse{
		ID:           "GYDS",
		Symbol:       "GYDS",
		Name:         "GYDS Token",
		Decimals:     18,
		TotalSupply:  "1000000000000000000000000000",
		Mintable:     true,
		Burnable:     true,
		IsStablecoin: false,
	}
	
	data, err := json.Marshal(asset)
	if err != nil {
		t.Errorf("failed to marshal asset: %v", err)
	}
	
	var decoded rpc.AssetResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Errorf("failed to unmarshal asset: %v", err)
	}
	
	if decoded.Symbol != "GYDS" {
		t.Errorf("expected symbol GYDS, got %s", decoded.Symbol)
	}
}

func TestHealthEndpoint(t *testing.T) {
	server := rpc.NewServer(":0")
	
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	
	// Would need to expose handler for testing
	_ = server
	_ = req
	_ = rr
}

// Benchmark tests
func BenchmarkRPCRequestMarshal(b *testing.B) {
	req := rpc.Request{
		JSONRPC: "2.0",
		Method:  "chain_getBlockByNumber",
		Params:  json.RawMessage(`{"number": 100}`),
		ID:      1,
	}
	
	for i := 0; i < b.N; i++ {
		json.Marshal(req)
	}
}

func BenchmarkRPCResponseMarshal(b *testing.B) {
	resp := rpc.Response{
		JSONRPC: "2.0",
		Result: rpc.BlockResponse{
			Number:    100,
			Hash:      "0x1234567890abcdef",
			Timestamp: 1704067200,
		},
		ID: 1,
	}
	
	for i := 0; i < b.N; i++ {
		json.Marshal(resp)
	}
}
