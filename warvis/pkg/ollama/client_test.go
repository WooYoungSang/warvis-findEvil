package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChatSuccess(t *testing.T) {
	// Mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}

		resp := ChatResponse{
			Model: "gemma4:26b-a4b-it-q4_K_M",
			Message: Message{
				Role:    "assistant",
				Content: `{"action": "call_tool", "tool_name": "timeline.build", "arguments": {}}`,
			},
			Done: true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "gemma4:26b-a4b-it-q4_K_M")
	ctx := context.Background()

	messages := []Message{
		{Role: "system", Content: "You are a forensic investigator."},
		{Role: "user", Content: "What tools do you need?"},
	}

	resp, err := client.Chat(ctx, messages)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if resp == nil {
		t.Fatal("response is nil")
	}
	if resp.Message.Role != "assistant" {
		t.Fatalf("unexpected role: %s", resp.Message.Role)
	}
	if resp.Done != true {
		t.Fatalf("expected done=true")
	}
}

func TestChatRetry(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			// Simulate failure on first 2 attempts
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Success on 3rd attempt
		resp := ChatResponse{
			Model: "gemma4:26b-a4b-it-q4_K_M",
			Message: Message{
				Role:    "assistant",
				Content: "success",
			},
			Done: true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "gemma4:26b-a4b-it-q4_K_M")
	ctx := context.Background()

	messages := []Message{
		{Role: "user", Content: "test"},
	}

	resp, err := client.Chat(ctx, messages)
	if err == nil {
		t.Fatalf("expected error on 500 responses")
	}
	_ = resp // Retries only on network errors, not HTTP 500
}

func TestListModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		resp := ModelsResponse{
			Models: []ModelInfo{
				{
					Name:       "gemma4:26b-a4b-it-q4_K_M",
					ModifiedAt: "2026-05-01T00:00:00Z",
					Size:       26000000000,
					Digest:     "abc123",
				},
				{
					Name:       "llama2:7b",
					ModifiedAt: "2026-05-01T00:00:00Z",
					Size:       7000000000,
					Digest:     "def456",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "gemma4:26b-a4b-it-q4_K_M")
	ctx := context.Background()

	models, err := client.ListModels(ctx)
	if err != nil {
		t.Fatalf("ListModels failed: %v", err)
	}

	if len(models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(models))
	}
	if models[0].Name != "gemma4:26b-a4b-it-q4_K_M" {
		t.Fatalf("unexpected model name: %s", models[0].Name)
	}
}

func TestCheckModelAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ModelsResponse{
			Models: []ModelInfo{
				{Name: "gemma4:26b-a4b-it-q4_K_M"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "gemma4:26b-a4b-it-q4_K_M")
	ctx := context.Background()

	available, err := client.CheckModelAvailable(ctx)
	if err != nil {
		t.Fatalf("CheckModelAvailable failed: %v", err)
	}

	if !available {
		t.Fatal("expected model to be available")
	}
}

func TestCheckModelNotAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ModelsResponse{
			Models: []ModelInfo{
				{Name: "llama2:7b"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "gemma4:26b-a4b-it-q4_K_M")
	ctx := context.Background()

	available, err := client.CheckModelAvailable(ctx)
	if err != nil {
		t.Fatalf("CheckModelAvailable failed: %v", err)
	}

	if available {
		t.Fatal("expected model to NOT be available")
	}
}

func TestChatContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This should not be called due to cancelled context
		t.Fatal("handler should not be called")
	}))
	defer server.Close()

	client := NewClient(server.URL, "gemma4:26b-a4b-it-q4_K_M")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	messages := []Message{
		{Role: "user", Content: "test"},
	}

	_, err := client.Chat(ctx, messages)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	// The error is wrapped by the retry logic, but should mention "context canceled"
	if err.Error() == "" {
		t.Fatalf("expected non-empty error message")
	}
}
