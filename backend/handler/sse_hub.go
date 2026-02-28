package handler

import (
	"errors"
	"sync"
)

// maxConnectionsPerWorkspace limits the number of concurrent SSE connections
// per workspace to prevent connection exhaustion.
const maxConnectionsPerWorkspace = 50

// ErrSSEConnectionLimitReached is returned when a workspace has reached the
// maximum number of concurrent SSE connections.
var ErrSSEConnectionLimitReached = errors.New("SSE connection limit reached for workspace")

// SSEHub is an in-memory pub/sub hub that tracks SSE client connections per workspace.
type SSEHub struct {
	mu      sync.RWMutex
	clients map[int64]map[chan string]struct{}
}

// NewSSEHub creates a new SSEHub.
func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[int64]map[chan string]struct{}),
	}
}

// Subscribe registers a new client channel for the given workspace.
// It returns the channel, an unsubscribe function, and an error if the
// connection limit has been reached.
func (h *SSEHub) Subscribe(workspaceID int64) (chan string, func(), error) {
	ch := make(chan string, 16)

	h.mu.Lock()
	if h.clients[workspaceID] == nil {
		h.clients[workspaceID] = make(map[chan string]struct{})
	}
	if len(h.clients[workspaceID]) >= maxConnectionsPerWorkspace {
		h.mu.Unlock()
		return nil, nil, ErrSSEConnectionLimitReached
	}
	h.clients[workspaceID][ch] = struct{}{}
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		delete(h.clients[workspaceID], ch)
		if len(h.clients[workspaceID]) == 0 {
			delete(h.clients, workspaceID)
		}
		h.mu.Unlock()
	}

	return ch, unsubscribe, nil
}

// Broadcast sends an event type string to all clients subscribed to the given workspace.
func (h *SSEHub) Broadcast(workspaceID int64, eventType string) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.clients[workspaceID] {
		select {
		case ch <- eventType:
		default:
			// Client too slow, skip to avoid blocking.
		}
	}
}

// Close closes all client channels, causing SSE handlers to exit.
// Call this during graceful shutdown to clean up SSE connections.
func (h *SSEHub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for wsID, chSet := range h.clients {
		for ch := range chSet {
			close(ch)
		}
		delete(h.clients, wsID)
	}
}
