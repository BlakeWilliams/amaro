package flash_test

import (
	"testing"

	"github.com/blakewilliams/amaro/httprouter/middleware/session"
	"github.com/blakewilliams/amaro/httprouter/middleware/session/flash"
	"github.com/stretchr/testify/require"
)

type sessionData struct {
	Flash *flash.Messages
}

func TestFlash(t *testing.T) {
	store := session.New("testing", session.NewVerifier("iiiiiiiiiiiiiiii"), nil, func() *sessionData {
		return &sessionData{}
	})

	data := &sessionData{
		Flash: &flash.Messages{},
	}

	data.Flash.Set("success", "You did it!")
	require.Equal(t, "", data.Flash.Get("success"))

	// Convert to cookie, then extract
	cookie, err := store.ToCookie(data)
	require.NoError(t, err)
	newData, err := store.FromCookie(cookie)
	require.NoError(t, err)

	require.Equal(t, "You did it!", newData.Flash.Get("success"))
}
