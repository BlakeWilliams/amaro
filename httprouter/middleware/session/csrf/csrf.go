package csrf

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

var Encoder = base64.StdEncoding.Strict()

// Token is a serializable Token token that can be used to validate the authenticity of requests.
type Token struct {
	Value       []byte `json:"token"`
	TokenLength int    `json:"tokenLength"`
}

var defaultTokenLength = 32

type opt func(*Token)

// NewCSRF creates a new CSRF token with the given options. This is used to
// create a new CSRF for a completely new session.
func NewCSRF(opts ...opt) *Token {
	csrf := &Token{}

	for _, opt := range opts {
		opt(csrf)
	}

	if csrf.TokenLength <= 0 {
		csrf.TokenLength = defaultTokenLength
	}

	csrf.Value = secureRandom(csrf.TokenLength)

	return csrf
}

// AuthenticityToken returns a masked authenticity token that can be used with
// the session middleware to validate CSRF.
func (c *Token) AuthenticityToken() string {
	pad := secureRandom(c.TokenLength)

	tokenBytes := []byte(c.Value)
	padBytes := []byte(pad)

	xorValue := make([]byte, len(tokenBytes))
	for i, v := range padBytes {
		xorValue[i] = v ^ tokenBytes[i]
	}

	maskedToken := append(padBytes, xorValue...)

	return Encoder.EncodeToString(maskedToken)
}

func (c *Token) VerifyAuthenticityToken(token string) (bool, error) {
	unmaskedToken, err := Encoder.DecodeString(token)
	if err != nil {
		return false, err
	}

	if len(unmaskedToken) != c.TokenLength*2 {
		return false, fmt.Errorf("invalid token length. expected %d, got %d", c.TokenLength*2, len(unmaskedToken))
	}

	pad := unmaskedToken[:c.TokenLength]
	xorToken := unmaskedToken[c.TokenLength:]

	realToken := make([]byte, len(xorToken))
	for i, v := range pad {
		realToken[i] = v ^ xorToken[i]
	}

	valid := true
	for i := 0; i < c.TokenLength; i++ {
		if realToken[i] != c.Value[i] {
			valid = false
		}
	}

	return valid, nil
}

func WithTokenLength(length int) opt {
	return func(c *Token) {
		c.TokenLength = length
	}
}

func secureRandom(length int) []byte {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return b
}
