package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

// Transport handles JSON-RPC 2.0 communication over stdio.
type Transport struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	stderr bytes.Buffer

	mu      sync.Mutex
	pending map[int]chan Response
	nextID  int

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

// NewTransport creates a new stdio transport and starts the server process.
func NewTransport(ctx context.Context, serverCmd []string) (*Transport, error) {
	if len(serverCmd) == 0 {
		return nil, fmt.Errorf("serverCmd cannot be empty")
	}

	cmd := exec.CommandContext(ctx, serverCmd[0], serverCmd[1:]...)
	cmd.Stderr = &bytes.Buffer{} // Capture stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		stdin.Close()
		return nil, fmt.Errorf("failed to start server process: %w", err)
	}

	tctx, cancel := context.WithCancel(ctx)

	t := &Transport{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  bufio.NewReader(stdout),
		pending: make(map[int]chan Response),
		nextID:  1,
		ctx:     tctx,
		cancel:  cancel,
		done:    make(chan struct{}),
	}

	// Copy stderr to our buffer
	if cmdStderr, err := cmd.StderrPipe(); err == nil {
		go func() {
			io.Copy(&t.stderr, cmdStderr)
		}()
	}

	// Start the read loop
	go t.readLoop()

	return t, nil
}

// Send sends a JSON-RPC 2.0 request and waits for a response.
func (t *Transport) Send(req Request) (Response, error) {
	// Allocate an ID and create response channel
	t.mu.Lock()
	req.ID = t.nextID
	t.nextID++
	respChan := make(chan Response, 1)
	t.pending[req.ID] = respChan
	t.mu.Unlock()

	// Encode and send the request
	data, err := json.Marshal(req)
	if err != nil {
		t.mu.Lock()
		delete(t.pending, req.ID)
		t.mu.Unlock()
		return Response{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	data = append(data, '\n')

	if _, err := t.stdin.Write(data); err != nil {
		t.mu.Lock()
		delete(t.pending, req.ID)
		t.mu.Unlock()
		return Response{}, fmt.Errorf("failed to write to stdin: %w", err)
	}

	// Wait for response with timeout
	select {
	case resp := <-respChan:
		return resp, nil
	case <-t.ctx.Done():
		t.mu.Lock()
		delete(t.pending, req.ID)
		t.mu.Unlock()
		return Response{}, fmt.Errorf("transport context cancelled")
	case <-time.After(30 * time.Second):
		t.mu.Lock()
		delete(t.pending, req.ID)
		t.mu.Unlock()
		return Response{}, fmt.Errorf("request timeout after 30s")
	}
}

// SendNotification sends a notification (no ID, no response expected).
func (t *Transport) SendNotification(notif Notification) error {
	data, err := json.Marshal(notif)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	data = append(data, '\n')

	if _, err := t.stdin.Write(data); err != nil {
		return fmt.Errorf("failed to write notification: %w", err)
	}

	return nil
}

// readLoop reads JSON-RPC 2.0 messages from stdout.
func (t *Transport) readLoop() {
	defer close(t.done)

	for {
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		line, err := t.stdout.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return
			}
			// Log error but continue
			fmt.Fprintf(os.Stderr, "failed to read from stdout: %v\n", err)
			return
		}

		// Parse as Response (includes both responses and notifications)
		var resp Response
		if err := json.Unmarshal(line, &resp); err != nil {
			// Might be a notification or something else
			fmt.Fprintf(os.Stderr, "failed to unmarshal response: %v\n", err)
			continue
		}

		// Route to pending request
		if resp.ID > 0 {
			t.mu.Lock()
			if ch, ok := t.pending[resp.ID]; ok {
				ch <- resp
				delete(t.pending, resp.ID)
			}
			t.mu.Unlock()
		}
	}
}

// Close gracefully shuts down the transport.
func (t *Transport) Close() error {
	t.cancel()

	// Try SIGTERM first
	if t.cmd.Process != nil {
		t.cmd.Process.Signal(os.Interrupt)
	}

	// Wait up to 2 seconds for graceful shutdown
	done := make(chan error, 1)
	go func() {
		done <- t.cmd.Wait()
	}()

	select {
	case <-done:
		return nil
	case <-time.After(2 * time.Second):
		// Kill if still running
		if t.cmd.Process != nil {
			t.cmd.Process.Kill()
		}
		<-done
		return nil
	}
}
