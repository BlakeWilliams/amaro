package flash

import (
	"encoding/json"
	"fmt"
	"sync"
)

type (
	// Messages contains the flash messages for a session, encoding/decoding as
	// needed into the session store.
	Messages struct {
		flashes map[string]Message
		mu      sync.RWMutex
	}

	Message struct {
		Value  string
		SetNow bool `json:"-"`
		WasSet bool `json:"-"`
	}
)

func (m *Messages) Set(name string, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.flashes == nil {
		m.flashes = make(map[string]Message)
	}

	m.flashes[name] = Message{Value: value, SetNow: false, WasSet: true}
}

func (m *Messages) SetNow(name string, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.flashes == nil {
		m.flashes = make(map[string]Message)
	}

	m.flashes[name] = Message{Value: value, SetNow: true, WasSet: true}
}

func (m *Messages) Get(name string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.flashes == nil {
		return ""
	}

	flash, ok := m.flashes[name]
	if !ok {
		return ""
	}

	// If it was set now, we should return it
	if flash.SetNow {
		return flash.Value
	}

	// If it _was_ set, and it's not SetNow, we return nothing
	if flash.WasSet {
		return ""
	}

	// If it wasn't set this session, we return it
	return flash.Value
}

func (m *Messages) MarshalJSON() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	output, err := json.Marshal(m.flashes)
	if err != nil {
		return nil, fmt.Errorf("Could not marshal flashes: %w", err)
	}

	return output, nil
}

func (m *Messages) UnmarshalJSON(data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	err := json.Unmarshal(data, &m.flashes)
	if err != nil {
		return fmt.Errorf("Could not unmarshal flashes: %w", err)
	}

	return nil
}
