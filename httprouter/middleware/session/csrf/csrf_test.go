package csrf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCSRF(t *testing.T) {
	csrf := NewCSRF()

	authToken := csrf.AuthenticityToken()
	valid, err := csrf.VerifyAuthenticityToken(authToken)
	require.NoError(t, err)

	require.True(t, valid)
}

func TestCSRFInvalidToken(t *testing.T) {
	csrf := NewCSRF()

	valid, err := csrf.VerifyAuthenticityToken("osYx4ryskqR/XW/0O0HLdMdK93lkOB6r4zC01hZ8xKFni4wPA9bNFLLB2g6SOLMIdPj+rwhX0lkDMFiZ6d5zRw==")
	require.NoError(t, err)

	require.False(t, valid)
}

func TestCSRFInvalidLength(t *testing.T) {
	csrf := NewCSRF(WithTokenLength(38))

	valid, err := csrf.VerifyAuthenticityToken("osYx4ryskqR/XW/0O0HLdMdK93lkOB6r4zC01hZ8xKFni4wPA9bNFLLB2g6SOLMIdPj+rwhX0lkDMFiZ6d5zRw==")

	require.Error(t, err)
	require.False(t, valid)
}
