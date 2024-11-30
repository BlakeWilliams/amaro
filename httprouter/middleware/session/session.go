package session

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/blakewilliams/amaro/httprouter"
)

// Verifier is an interface that allows a given string of bytes to be encoded
// and decoded, verifying that the contents have not been tampered with.
type Verifier interface {
	Encode([]byte) (string, error)
	Decode(string) ([]byte, error)
}

// Persistable is an interface that must be implemented by a RequestContext or
// other object to allow the middleware and other application components access
// and set the session data.
type Persistable[T any] interface {
	// SetSessionData sets the session data for the request.
	SetSessionData(T)
	// SessionData should return a pointer to the session data.
	SessionData() T

	httprouter.RequestContext
}

// Middleware is a middleware function that can be used with the httprouter
// middleware stack. It creates a new session and adds it to the RequestContext
// for the request.
func Middleware[RC Persistable[T], T any](store Store[T]) func(ctx context.Context, rc RC, next httprouter.Handler[RC]) {
	return func(ctx context.Context, rc RC, next httprouter.Handler[RC]) {
		cookie, err := rc.Request().Cookie(store.Name)
		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			panic(err)
		}

		data, err := store.FromCookie(cookie)
		if err != nil {
			panic(err)
		}

		// Marshal the original data so that we can check if any has changed at
		// the end of the request when determining if we should send an updated
		// cookie
		originalData, err := store.marshal(data)
		if err != nil {
			panic(err)
		}
		rc.SetSessionData(data)

		defer func() {
			newData, err := store.marshal(rc.SessionData())
			if err != nil {
				panic(err)
			}

			if !bytes.Equal(newData, originalData) {
				err = store.Write(rc, rc.SessionData())
				if err != nil {
					panic(err)
				}
			}
		}()

		next(ctx, rc)
	}
}

// Store is a wrapper around a http.Cookie that provides signed messages,
// allowing you to securely store data in a cookie.
//
// The data stored is still readable by the client, so secrets and sensitive
// data should not be stored in Store.Data.
type Store[T any] struct {
	verifier   Verifier
	Name       string
	initState  func() T
	cookieOpts *CookieOptions
}

// CookieOptions are the options that are used when creating the underlying
// http.Cookie.
type CookieOptions struct {
	// Path is the path that the cookie is valid for. Defaults to unset.
	Path string
	// Domain is the domain that the cookie is valid for. Defaults to unset.
	Domain string
	// MaxAge is the maximum age of the cookie in seconds. Defaults to unset.
	MaxAge int
	// Secure indicates whether the cookie should only be sent over HTTPS.
	// Defaults to false, ensuring the cookie is sent over only HTTPS.
	Secure bool
	// HttpOnly indicates whether the cookie should only be sent via HTTP(S)
	HTTPOnly bool
	// SameSite indicates whether the cookie should only be sent to the same
	// site. Defaults to http.SameSiteLaxMode.
	SameSite http.SameSite
}

// New creates a new Store with the given name and verifies Data using the
// passed in Verifier. If opts is nil, the default options will result in a
// cookie that is SameSiteLaxMode, Secure, and HTTPOnly.
//
// Name is the name of the cookie, verifier will be used to validate the
// cookie, and initState will be used to initialize the cookie when it is
// missing or invalid.
func New[T any](name string, verifier Verifier, opts *CookieOptions, initState func() T) Store[T] {
	if opts == nil {
		opts = &CookieOptions{
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			HTTPOnly: true,
		}
	}

	return Store[T]{
		Name:       name,
		initState:  initState,
		verifier:   verifier,
		cookieOpts: opts,
	}
}

// FromRequest reads the cookie with the provided name from the Request. The
// data is then decoded and verified using the Verifier.
func (s Store[T]) FromRequest(rc httprouter.RequestContext) (T, error) {
	cookie, err := rc.Request().Cookie(s.Name)

	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		return s.initState(), fmt.Errorf("Could not create session from request: %w", err)
	}

	return s.FromCookie(cookie)
}

// FromCookie attempts to decode the data from the passed in Cookie and verifies
// the data hasn't been tampered with.
func (s Store[T]) FromCookie(cookie *http.Cookie) (T, error) {
	data := s.initState()
	if cookie == nil {
		return data, nil
	}

	decodedMessage, err := s.verifier.Decode(cookie.Value)

	if err != nil {
		return data, err
	}

	err = json.Unmarshal(decodedMessage, &data)

	if err != nil {
		return data, fmt.Errorf("Could not decode session: %w", err)
	}

	return data, nil
}

// Write encodes the Data into a JSON object, signs it using the message
// verifier, then writes it to the passed in httprouter.RequestContext using the
// name provided by New.
func (s Store[T]) Write(rc httprouter.RequestContext, data T) error {
	cookie, err := s.ToCookie(data)
	if err != nil {
		return err
	}

	http.SetCookie(rc.Response(), cookie)

	return nil
}

func (s Store[T]) marshal(data T) ([]byte, error) {
	jsonValue, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Could not marshal session data: %w", err)
	}
	return jsonValue, nil
}

// ToCookie returns the underlying http.ToCookie that is used to store the session.
func (s Store[T]) ToCookie(data T) (*http.Cookie, error) {
	jsonValue, err := s.marshal(data)

	encodedData, err := s.verifier.Encode(jsonValue)
	if err != nil {
		return nil, fmt.Errorf("could not encode data: %w", err)
	}

	cookie := &http.Cookie{
		Name:  s.Name,
		Path:  "/",
		Value: string(encodedData),
	}

	if s.cookieOpts.Domain != "" {
		cookie.Domain = s.cookieOpts.Domain
	}

	if s.cookieOpts.MaxAge != 0 {
		cookie.MaxAge = s.cookieOpts.MaxAge
	}

	if s.cookieOpts.SameSite != 0 {
		cookie.SameSite = s.cookieOpts.SameSite
	} else {
		cookie.SameSite = http.SameSiteLaxMode
	}

	cookie.Secure = s.cookieOpts.Secure
	cookie.HttpOnly = s.cookieOpts.HTTPOnly

	return cookie, nil
}
