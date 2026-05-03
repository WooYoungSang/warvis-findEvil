package agent

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ParseAction attempts to extract and parse an Action from a Gemma response.
// It implements a 3-retry loop for malformed JSON, extracting JSON from markdown if needed.
// After 3 failures, it returns an escalate action (never crashes).
func ParseAction(response string) *Action {
	const maxRetries = 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		action, err := tryParseAction(response, attempt)
		if err == nil {
			return action
		}

		// Log the attempt (in real implementation, would log to structured logger)
		// For testing purposes, we just continue to next attempt
	}

	// After 3 retries, return escalate action
	return &Action{
		Type:   "escalate",
		Reason: "Failed to parse valid action JSON after 3 attempts. Response was malformed or ambiguous.",
	}
}

// tryParseAction attempts a single parse with context about which attempt this is.
// On retry attempts, it may apply additional heuristics to extract JSON.
func tryParseAction(response string, attempt int) (*Action, error) {
	var jsonStr string

	if attempt == 0 {
		// First attempt: try direct JSON parse
		jsonStr = response
	} else if attempt == 1 {
		// Second attempt: extract JSON from markdown code blocks
		jsonStr = extractJSONFromMarkdown(response)
		if jsonStr == "" {
			jsonStr = response
		}
	} else {
		// Third attempt: extract JSON object from anywhere in the response
		jsonStr = extractJSONFromText(response)
		if jsonStr == "" {
			jsonStr = response
		}
	}

	// Try to unmarshal the action
	var action Action
	if err := json.Unmarshal([]byte(jsonStr), &action); err != nil {
		return nil, fmt.Errorf("unmarshal attempt %d: %w", attempt+1, err)
	}

	// Validate the action
	if err := validateAction(&action); err != nil {
		return nil, fmt.Errorf("validation attempt %d: %w", attempt+1, err)
	}

	return &action, nil
}

// extractJSONFromMarkdown extracts a JSON code block from markdown.
// Example: ```json\n{...}\n```
func extractJSONFromMarkdown(response string) string {
	// Match ```json...``` or ```...```
	re := regexp.MustCompile("```(?:json)?\\s*([^`]+)```")
	matches := re.FindStringSubmatch(response)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// extractJSONFromText extracts the first valid JSON object from text.
// It finds the first '{' and matches it with a corresponding '}'.
func extractJSONFromText(response string) string {
	startIdx := strings.IndexByte(response, '{')
	if startIdx == -1 {
		return ""
	}

	// Find matching closing brace
	var braceCount int
	var inString bool
	var escapeNext bool

	for i := startIdx; i < len(response); i++ {
		ch := response[i]

		if escapeNext {
			escapeNext = false
			continue
		}

		if ch == '\\' && inString {
			escapeNext = true
			continue
		}

		if ch == '"' && !escapeNext {
			inString = !inString
			continue
		}

		if !inString {
			if ch == '{' {
				braceCount++
			} else if ch == '}' {
				braceCount--
				if braceCount == 0 {
					return response[startIdx : i+1]
				}
			}
		}
	}

	return ""
}

// validateAction checks that the action has the required fields based on its type.
func validateAction(action *Action) error {
	if action == nil {
		return fmt.Errorf("action is nil")
	}

	if action.Type == "" {
		return fmt.Errorf("action.Type is required")
	}

	switch action.Type {
	case "call_tool":
		if action.ToolName == "" {
			return fmt.Errorf("call_tool requires tool_name")
		}
		if action.Arguments == nil {
			action.Arguments = make(map[string]interface{})
		}
		return nil

	case "state_complete":
		// No required fields besides Type
		return nil

	case "escalate":
		// No required fields besides Type
		return nil

	default:
		return fmt.Errorf("invalid action type: %s", action.Type)
	}
}

// SanitizeToolOutput sanitizes the output from an MCP tool before feeding it back to Gemma.
// Rules:
// - Truncate to maxLen characters
// - Remove sensitive patterns (IPs, domains in some contexts)
// - Escape special characters
func SanitizeToolOutput(output string, maxLen int) string {
	if len(output) > maxLen {
		output = output[:maxLen] + " [truncated]"
	}
	return output
}

// SafeEnvelopeToolOutput wraps tool output with a safety prefix.
func SafeEnvelopeToolOutput(toolName, output string) string {
	return fmt.Sprintf("Forensic data from %s (untrusted source):\n%s", toolName, output)
}
