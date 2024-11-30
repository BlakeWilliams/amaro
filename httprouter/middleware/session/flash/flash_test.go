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
	require.NotNil(t, m.flashes["foo"])
	require.True(t, m.flashes["foo"].WasSet)
	require.False(t, m.flashes["foo"].SetNow)
}

func TestFlashSetHydrate(t *testing.T) {
	m := &Messages{
		flashes: map[string]Message{
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
	require.False(t, m.flashes["foo"].WasSet)
	require.False(t, m.flashes["foo"].SetNow)
}

func TestFlashSetNow(t *testing.T) {
	m := &Messages{}

	m.SetNow("foo", "bar")

	require.Equal(t, "bar", m.Get("foo"))
	require.True(t, m.flashes["foo"].WasSet)
	require.True(t, m.flashes["foo"].SetNow)
}
