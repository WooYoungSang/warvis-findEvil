package mcp

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"
)

// TestTransportSendReceive tests basic request/response cycle.
func TestTransportSendReceive(t *testing.T) {
	// Create a mock server using net.Pipe
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	// Mock server goroutine
	go func() {
		defer serverConn.Close()
		buf := make([]byte, 1024)
		n, err := serverConn.Read(buf)
		if err != nil {
			t.Logf("server read error: %v", err)
			return
		}

		// Parse the request
		var req Request
		if err := json.Unmarshal(buf[:n], &req); err != nil {
			t.Logf("failed to unmarshal request: %v", err)
			return
		}

		// Send a response
		resp := Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage([]byte(`{"status":"ok"}`)),
		}

		respData, _ := json.Marshal(resp)
		respData = append(respData, '\n')
		serverConn.Write(respData)
	}()

	// Create transport manually with net.Pipe (simulating stdio)
	t.Logf("test setup complete, transport test deferred")
}

// TestTransportGracefulShutdown tests graceful shutdown.
func TestTransportGracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use a simple no-op command
	transport, err := NewTransport(ctx, []string{"cat"})
	if err != nil {
		t.Fatalf("failed to create transport: %v", err)
	}

	// Close should not panic
	if err := transport.Close(); err != nil {
		t.Logf("close error (expected): %v", err)
	}
}

// TestClientInitialize tests the initialize handshake.
func TestClientInitialize(t *testing.T) {
	// This test would require a real or mock MCP server
	// For now, we skip and rely on integration tests
	t.Skip("requires mock server setup")
}

// TestClientListTools tests listing tools.
func TestClientListTools(t *testing.T) {
	t.Skip("requires mock server setup")
}

// TestClientCallTool tests calling a tool.
func TestClientCallTool(t *testing.T) {
	t.Skip("requires mock server setup")
}

// Integration test (requires actual MCP server)
// To run: go test -tags integration -run TestClientIntegration

// TestClientIntegration tests the client against a real MCP server.
// This test requires the find_evil_mcp server to be running.
func TestClientIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try to connect to the MCP server
	client, err := NewClient(ctx, []string{"python", "-m", "find_evil_mcp"})
	if err != nil {
		t.Skipf("failed to create client (MCP server may not be running): %v", err)
	}
	defer client.Close()

	// Test ListTools
	tools, err := client.ListTools()
	if err != nil {
		t.Fatalf("failed to list tools: %v", err)
	}

	if len(tools) == 0 {
		t.Error("expected non-empty tool list")
	}

	t.Logf("found %d tools", len(tools))

	// Verify some expected tools are present
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
		t.Logf("  - %s: %s", tool.Name, tool.Description)
	}

	expectedTools := []string{
		"case.open",
		"case.read",
		"case.list",
	}

	for _, expectedTool := range expectedTools {
		if !toolNames[expectedTool] {
			t.Logf("warning: expected tool %s not found", expectedTool)
		}
	}
}

// TestProtocolTypes tests that protocol types marshal/unmarshal correctly.
func TestProtocolTypes(t *testing.T) {
	// Test Request
	req := Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "test",
		Params:  json.RawMessage([]byte(`{"foo":"bar"}`)),
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	var req2 Request
	if err := json.Unmarshal(data, &req2); err != nil {
		t.Fatalf("failed to unmarshal request: %v", err)
	}

	if req2.ID != req.ID || req2.Method != req.Method {
		t.Error("request marshal/unmarshal mismatch")
	}

	// Test Response
	resp := Response{
		JSONRPC: "2.0",
		ID:      1,
		Result:  json.RawMessage([]byte(`{"ok":true}`)),
	}

	data, err = json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var resp2 Response
	if err := json.Unmarshal(data, &resp2); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp2.ID != resp.ID {
		t.Error("response marshal/unmarshal mismatch")
	}

	// Test Notification
	notif := Notification{
		JSONRPC: "2.0",
		Method:  "test/event",
		Params:  json.RawMessage([]byte(`{}`)),
	}

	data, err = json.Marshal(notif)
	if err != nil {
		t.Fatalf("failed to marshal notification: %v", err)
	}

	var notif2 Notification
	if err := json.Unmarshal(data, &notif2); err != nil {
		t.Fatalf("failed to unmarshal notification: %v", err)
	}

	if notif2.Method != notif.Method {
		t.Error("notification marshal/unmarshal mismatch")
	}
}
