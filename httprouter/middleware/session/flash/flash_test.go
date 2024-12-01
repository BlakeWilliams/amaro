package flash

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlashSet(t *testing.T) {
	m := &Messages{}

	m.Set("foo", "bar")

	require.Equal(t, "", m.Get("foo"))
	require.NotNil(t, m.next["foo"])
}

func TestFlashSetHydrate(t *testing.T) {
	m := &Messages{
		next: map[string]message{
			"foo": {
				Value: "bar",
			},
		},
	}

	encoded, err := m.MarshalJSON()
	require.NoError(t, err)

	m = &Messages{}
	err = m.UnmarshalJSON(encoded)
	require.NoError(t, err)

	require.Equal(t, "bar", m.Get("foo"))
}

func TestFlashSetNow(t *testing.T) {
	m := &Messages{}

	m.SetNow("foo", "bar")
	require.Equal(t, "bar", m.Get("foo"))
}
