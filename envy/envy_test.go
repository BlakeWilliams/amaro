package envy

import (
	"embed"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed .env
var efs embed.FS

func TestValidEnv(t *testing.T) {
	// TODO implement better solution for clearing env
	defer func() {
		os.Unsetenv("BASIC_CASE")
	}()
	err := LoadFS(efs, "test")
	require.NoError(t, err)

	require.Equal(t, "bar", os.Getenv("BASIC_CASE"))
	require.Equal(t, "", os.Getenv("EMPTY_QUOTES"), "empty quotes")
	require.Equal(t, "hello, world", os.Getenv("QUOTED_VAL"), "quoted value")
	require.Equal(t, "hello\nworld", os.Getenv("MULTILINE"), "multiline")
	require.Equal(t, "hello, world!", os.Getenv("INLINE_COMMENT"), "inline comment")
	require.Equal(t, "'hello, world'", os.Getenv("NESTED_QUOTES"), "nested quotes")
	require.Equal(t, `hello, "world"!`, os.Getenv("SINGLE_QUOTES"), "single quotes")
	require.Equal(t, "hello", os.Getenv("EXTRA_KEY_SPACE"), "extra key space")
	require.Equal(t, "hello\nworld", os.Getenv("EXPAND_NEWLINES"), "expand newlines")
	require.Equal(t, "hello\\nworld", os.Getenv("NO_EXPAND_NEWLINES"), "no expand newlines")
	require.Equal(t, `"this is the way"`, os.Getenv("ESCAPED_QUOTES"), "escaped quotes")
}

// this will test that .env.test takes precedence
//
//go:embed .env .env.test
var testfs embed.FS

func TestPrecedence(t *testing.T) {
	// TODO implement better solution for clearing env
	defer func() {
		os.Unsetenv("BASIC_CASE")
	}()
	err := LoadFS(testfs, "test")
	require.NoError(t, err)

	require.Equal(t, "overridden", os.Getenv("BASIC_CASE"))
}

func TestInvalidEnv(t *testing.T) {
	t.Run("unexpected character", func(t *testing.T) {
		err := LoadString(`INVALID="foo bar" fo`)
		require.ErrorContains(t, err, "unexpected character")
	})
	t.Run("invalid key", func(t *testing.T) {
		err := LoadString(`INVALID-KEY="foo bar"`)
		require.ErrorContains(t, err, "invalid key")
	})
	t.Run("missing =", func(t *testing.T) {
		err := LoadString(`MISSING_EQUALS wow`)
		require.ErrorContains(t, err, "missing =")
	})
}
