package ai

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// Manager holds multiple AI clients mapped by agent ID
type Manager struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]Client
}

// NewManager creates a new AI manager
func NewManager() *Manager {
	return &Manager{clients: make(map[uuid.UUID]Client)}
}

// Register associates an agent with a client
func (m *Manager) Register(agentID uuid.UUID, client Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[agentID] = client
}

// Get returns the client for an agent
func (m *Manager) Get(agentID uuid.UUID) (Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.clients[agentID]
	if !ok {
		return nil, fmt.Errorf("no AI client registered for agent %s", agentID)
	}
	return c, nil
}

// All returns a copy of all mappings
func (m *Manager) All() map[uuid.UUID]Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[uuid.UUID]Client, len(m.clients))
	for k, v := range m.clients {
		out[k] = v
	}
	return out
}

// Count returns number of registered clients
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}

// Unregister removes a mapping
func (m *Manager) Unregister(agentID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, agentID)
}
