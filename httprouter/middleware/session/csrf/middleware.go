package csrf

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/blakewilliams/amaro/httprouter"
)

var ErrTokenInvalid = errors.New("invalid authenticity token")

type (
	// CSRFable is an interface that a RequestContext can implement to receive
	// CSRF protections.
	CSRFable interface {
		// CSRF returns the CSRF token for the request.
		CSRF() *Token
		// SetCSRF sets the CSRF token for the request.
		SetCSRF(*Token)

		httprouter.RequestContext
	}

	// MiddlewareConfig is a configuration object for the CSRF middleware
	MiddlewareConfig[T CSRFable] struct {
		// TokenLength is the size of the generated authenticity token in
		// bytes, before masking.
		TokenLength int
		// HandleInvalidToken is called when an invalid token is received. If
		// nil, the default behavior is to panic and 500. The preferred
		// behavior is to implement this function, set a flash message letting
		// the user know the error, and redirect.
		HandleInvalidToken func(context.Context, T)

		// Logger returns the logger for the request.
		Logger interface {
			Error(string, ...any)
		}
	}
)

// Middleware handles validating and setting CSRF tokens for the request. If
// used with github.com/blakewilliams/amaro/session (recommended) the session
// middleware must be run first so the CSRF value will be hydrated and accessible.
func Middleware[T CSRFable](config MiddlewareConfig[T]) httprouter.Middleware[T] {
	logger := config.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}

	return func(ctx context.Context, rctx T, next httprouter.Handler[T]) {
		if rctx.CSRF() == nil {
			rctx.SetCSRF(NewCSRF(WithTokenLength(config.TokenLength)))
		}

		switch rctx.Request().Method {
		case http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete:
			var token string

			// Check header first to avoid calling ParseForm if it's unnecessary
			if headerToken := rctx.Request().Header.Get("x-csrf-token"); headerToken != "" {
				token = headerToken
			} else {
				err := rctx.Request().ParseForm()
				if err != nil {
					logger.Error("unable to parse form in csrf middleware", "error", err.Error(), "valid", "false")
					panic("invalid authenticity token")
				}

				if formTokens := rctx.Request().PostForm["authenticity_token"]; len(formTokens) > 0 {
					token = formTokens[0]
				}
			}

			if valid, err := rctx.CSRF().VerifyAuthenticityToken(token); !valid || err != nil {
				logKeys := []any{"valid", fmt.Sprint(valid)}
				if err != nil {
					logKeys = append(logKeys, "error", err.Error())
				}
				logger.Error("invalid authenticity token", logKeys...)
				if config.HandleInvalidToken != nil {
					config.HandleInvalidToken(ctx, rctx)
					return
				}

				if err != nil {
					panic(err)
				} else {
					panic(ErrTokenInvalid)
				}
			}

			next(ctx, rctx)
		default:
			next(ctx, rctx)
		}
	}
}
