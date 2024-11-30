package flash

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/blakewilliams/amaro/httprouter"
)

type (
	// Messages contains the flash messages for a session, encoding/decoding as
	// needed into the session store.
	Messages struct {
		entries map[string]message
		mu      sync.RWMutex
	}

	message struct {
		Value   string
		SetNow  bool `json:"-"`
		WasSet  bool `json:"-"`
		WasRead bool `json:"-"`
	}

	// FlashableRequestContext is a request context that reutrns a session with
	// flash support. This is necessary to use the middleware.
	FlashableRequestContext interface {
		Flash() *Messages
		httprouter.RequestContext
	}
)

func (m *Messages) Set(name string, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.entries == nil {
		m.entries = make(map[string]message)
	}

	m.entries[name] = message{Value: value, SetNow: false, WasSet: true}
}

func (m *Messages) SetNow(name string, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.entries == nil {
		m.entries = make(map[string]message)
	}

	m.entries[name] = message{Value: value, SetNow: true, WasSet: true}
}

func (m *Messages) Get(name string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.entries == nil {
		return ""
	}

	flash, ok := m.entries[name]
	if !ok {
		return ""
	}

	// If it was set now, we should return it
	if flash.SetNow {
		flash.WasRead = true
		return flash.Value
	}

	// If it _was_ set, and it's not SetNow, we return nothing
	if flash.WasSet {
		return ""
	}

	// If it wasn't set this session, we return it
	flash.WasRead = true
	return flash.Value
}

func (m *Messages) Delete(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.entries, name)
}

func (m *Messages) MarshalJSON() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	output, err := json.Marshal(m.entries)
	if err != nil {
		return nil, fmt.Errorf("Could not marshal flashes: %w", err)
	}

	return output, nil
}

func (m *Messages) UnmarshalJSON(data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	err := json.Unmarshal(data, &m.entries)
	if err != nil {
		return fmt.Errorf("Could not unmarshal flashes: %w", err)
	}

	return nil
}

// Middleware enables flash being clearing after use
func Middleware[T FlashableRequestContext](ctx context.Context, rc T, next httprouter.Handler[T]) {
	next(ctx, rc)

	flash := rc.Flash()
	for k, message := range flash.entries {
		if message.WasRead || message.SetNow {
			flash.Delete(k)
		}
	}
}
