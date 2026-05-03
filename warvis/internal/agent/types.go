package agent

import (
	"time"
)

// Action represents an action parsed from Gemma's response.
type Action struct {
	Type      string                 `json:"action"`                 // "call_tool" | "state_complete" | "escalate"
	ToolName  string                 `json:"tool_name"`             // Required if Type == "call_tool"
	Arguments map[string]interface{} `json:"arguments,omitempty"`   // Required if Type == "call_tool"
	Reason    string                 `json:"reason,omitempty"`      // Optional explanation
}

// ConversationTurn represents a single exchange in the conversation.
type ConversationTurn struct {
	Timestamp time.Time
	Role      string // "user" or "assistant"
	Content   string
	Metadata  map[string]interface{} // For tool calls, metadata
}

// ToolInfo represents metadata about an available tool.
type ToolInfo struct {
	Name        string
	Description string
	Parameters  map[string]interface{} // JSON schema
}
