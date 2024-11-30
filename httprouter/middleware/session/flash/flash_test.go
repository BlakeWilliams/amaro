package flash

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlashSet(t *testing.T) {
	m := &Messages{}

	m.Set("foo", "bar")

	require.Equal(t, "", m.Get("foo"))
	require.NotNil(t, m.entries["foo"])
	require.True(t, m.entries["foo"].WasSet)
	require.False(t, m.entries["foo"].SetNow)
}

func TestFlashSetHydrate(t *testing.T) {
	m := &Messages{
		entries: map[string]message{
			"foo": {
				Value:  "bar",
				WasSet: true,
			},
		},
	}

	encoded, err := m.MarshalJSON()
	require.NoError(t, err)

	m = &Messages{}
	err = m.UnmarshalJSON(encoded)
	fmt.Println(string(encoded))
	require.NoError(t, err)

	require.Equal(t, "bar", m.Get("foo"))
	require.False(t, m.entries["foo"].WasSet)
	require.False(t, m.entries["foo"].SetNow)
}

func TestFlashSetNow(t *testing.T) {
	m := &Messages{}

	m.SetNow("foo", "bar")

	require.Equal(t, "bar", m.Get("foo"))
	require.True(t, m.entries["foo"].WasSet)
	require.True(t, m.entries["foo"].SetNow)
}
