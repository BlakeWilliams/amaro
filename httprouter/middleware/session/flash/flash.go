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
		// snapshot captures a copy of the initial entries so that
		// the current request can't overwrite an existing flash message
		snapshot map[string]message
		// next is the set of flashes that will be available to the next
		// request. They are not readable in this request.
		next map[string]message

		toRemove map[string]struct{}
		mu       sync.RWMutex
	}

	message struct {
		Value string
	}

	// FlashableRequestContext is a request context that reutrns a session with
	// flash support. This is necessary to use the middleware.
	FlashableRequestContext interface {
		Flash() *Messages
		httprouter.RequestContext
	}
)

// Set sets the given flash message.
func (m *Messages) Set(name string, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.next == nil {
		m.next = make(map[string]message)
	}

	m.next[name] = message{Value: value}
}

// SetNow sets the given flash message, and ensures that it's available _only_
// for this request. This can override existing flash messages.
func (m *Messages) SetNow(name string, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.snapshot == nil {
		m.snapshot = make(map[string]message)
	}

	m.snapshot[name] = message{Value: value}

	if m.toRemove == nil {
		m.toRemove = make(map[string]struct{})
	}
	m.toRemove[name] = struct{}{}
}

// Get returns the flash message for the given flash. If the flash message was
// set using `SetNow` it will be available during this request. If it was set
// using `Set` it will be available until its value is read or it is overwritten.
func (m *Messages) Get(name string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.snapshot == nil {
		return ""
	}

	flash, ok := m.snapshot[name]
	if !ok {
		return ""
	}

	if m.toRemove == nil {
		m.toRemove = make(map[string]struct{})
	}
	m.toRemove[name] = struct{}{}

	return flash.Value
}

func (m *Messages) MarshalJSON() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	output, err := json.Marshal(m.next)
	if err != nil {
		return nil, fmt.Errorf("Could not marshal flashes: %w", err)
	}

	return output, nil
}

func (m *Messages) UnmarshalJSON(data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	err := json.Unmarshal(data, &m.snapshot)
	if err != nil {
		return fmt.Errorf("Could not unmarshal flashes: %w", err)
	}

	return nil
}

// Rollover takes the unread flash messages and sets them to be reused in the
// next request if they haven't been overwritten.
func (m *Messages) Rollover() {
	for k, message := range m.snapshot {
		if _, ok := m.toRemove[k]; ok {
			continue
		}

		if _, ok := m.next[k]; ok {
			continue
		}

		m.Set(k, message.Value)
	}
}

// Middleware is necessary to ensure that the flash messages are cleaned up
// after being used.
func Middleware[T FlashableRequestContext](ctx context.Context, rc T, next httprouter.Handler[T]) {
	next(ctx, rc)

	rc.Flash().Rollover()
}
