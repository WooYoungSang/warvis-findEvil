package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

// Client is an MCP JSON-RPC 2.0 client.
type Client struct {
	transport *Transport
	tools     []Tool // Cached tools
}

// NewClient creates a new MCP client and establishes connection.
func NewClient(ctx context.Context, serverCmd []string) (*Client, error) {
	// Create and start transport
	transport, err := NewTransport(ctx, serverCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	c := &Client{
		transport: transport,
	}

	// Initialize handshake
	if err := c.initialize(); err != nil {
		transport.Close()
		return nil, fmt.Errorf("initialization failed: %w", err)
	}

	// Send notifications/initialized
	notif := Notification{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
		Params:  json.RawMessage([]byte("{}")),
	}
	if err := transport.SendNotification(notif); err != nil {
		transport.Close()
		return nil, fmt.Errorf("failed to send initialized notification: %w", err)
	}

	// Cache tools
	tools, err := c.ListTools()
	if err != nil {
		transport.Close()
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}
	c.tools = tools

	return c, nil
}

// initialize performs the MCP initialize handshake.
func (c *Client) initialize() error {
	params := InitializeParams{
		ProtocolVersion: "2024-11-05",
		ClientInfo: ClientInfo{
			Name:    "warvis",
			Version: "0.1.0",
		},
		Capabilities: ClientCapabilities{},
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal initialize params: %w", err)
	}

	req := Request{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params:  paramsJSON,
	}

	resp, err := c.transport.Send(req)
	if err != nil {
		return fmt.Errorf("failed to send initialize request: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("initialize request failed: %s", resp.Error.Message)
	}

	// Parse the response
	var result InitializeResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return fmt.Errorf("failed to unmarshal initialize result: %w", err)
	}

	// Validate protocol version
	if result.ProtocolVersion != "2024-11-05" {
		return fmt.Errorf("unsupported protocol version: %s", result.ProtocolVersion)
	}

	return nil
}

// ListTools returns the list of available tools.
func (c *Client) ListTools() ([]Tool, error) {
	// Return cached tools if available
	if len(c.tools) > 0 {
		return c.tools, nil
	}

	req := Request{
		JSONRPC: "2.0",
		Method:  "tools/list",
		Params:  json.RawMessage([]byte("{}")),
	}

	resp, err := c.transport.Send(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send list_tools request: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("list_tools request failed: %s", resp.Error.Message)
	}

	var result ListToolsResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal list_tools result: %w", err)
	}

	return result.Tools, nil
}

// CallTool invokes a tool with the given arguments.
func (c *Client) CallTool(name string, args interface{}) (*CallToolResult, error) {
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tool arguments: %w", err)
	}

	params := CallToolParams{
		Name:      name,
		Arguments: argsJSON,
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal call_tool params: %w", err)
	}

	req := Request{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params:  paramsJSON,
	}

	resp, err := c.transport.Send(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send call_tool request: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("call_tool request failed: %s", resp.Error.Message)
	}

	var result CallToolResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal call_tool result: %w", err)
	}

	return &result, nil
}

// Close closes the client and transport connection.
func (c *Client) Close() error {
	return c.transport.Close()
}
