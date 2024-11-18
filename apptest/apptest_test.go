package apptest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResponseBody_MultiReadable(t *testing.T) {
	rec := &httptest.ResponseRecorder{
		Body: bytes.NewBuffer([]byte("omg wow")),
	}

	res := Response{RawResponse: rec}

	require.Equal(t, "omg wow", res.Body().String())
	require.Equal(t, "omg wow", res.Body().String())
}

func TestNonBody(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		statusCode int
	}{
		{"GET", http.MethodGet, 200},
		{"OPTIONS", http.MethodOptions, 200},
		{"HEAD", http.MethodHead, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			done := make(chan struct{})

			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer close(done)
				require.Equal(t, tt.method, r.Method)
			})

			session := New(h)

			var res *Response
			switch tt.method {
			case http.MethodGet:
				res = session.Get("/foo", nil)
			case http.MethodOptions:
				res = session.Options("/foo", nil)
			case http.MethodHead:
				res = session.Head("/foo", nil)
			default:
				t.Fatalf("unsupported method: %s", tt.method)
			}

			<-done
			require.Equal(t, tt.statusCode, res.Code())
		})
	}
}

func TestForm(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		body       url.Values
		statusCode int
	}{
		{"post", http.MethodPost, url.Values{"foo": {"bar"}}, 200},
		{"put", http.MethodPut, url.Values{"foo": {"bar"}}, 200},
		{"patch", http.MethodPatch, url.Values{"foo": {"bar"}}, 200},
		{"delete", http.MethodDelete, url.Values{"foo": {"bar"}}, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done := make(chan struct{})

			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer close(done)
				require.Equal(t, tt.method, r.Method)

				if tt.body != nil {
					require.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

					// Hacky workaround since Go stdlib doesn't support
					// DELETE with formdata
					if tt.method == http.MethodDelete {
						r.Method = http.MethodPost
						err := r.ParseForm()
						require.NoError(t, err)
						r.Method = http.MethodDelete
					} else {
						err := r.ParseForm()
						require.NoError(t, err)
					}

					require.Equal(t, tt.body, r.PostForm)
				}
			})

			session := New(h)

			var res *Response
			switch tt.method {
			case http.MethodPost:
				res = session.PostForm("/foo", nil, tt.body)
			case http.MethodPut:
				res = session.PutForm("/foo", nil, tt.body)
			case http.MethodPatch:
				res = session.PatchForm("/foo", nil, tt.body)
			case http.MethodDelete:
				res = session.DeleteForm("/foo", nil, tt.body)
			default:
				t.Fatalf("unsupported method: %s", tt.method)
			}

			<-done
			require.Equal(t, tt.statusCode, res.Code())
		})
	}
}

func TestJSON(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		body       map[string]any
		statusCode int
	}{
		{"post", http.MethodPost, map[string]any{"foo": "bar"}, 200},
		{"put", http.MethodPut, map[string]any{"foo": "bar"}, 200},
		{"patch", http.MethodPatch, map[string]any{"foo": "bar"}, 200},
		{"delete", http.MethodDelete, map[string]any{"foo": "bar"}, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done := make(chan struct{})

			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer close(done)
				require.Equal(t, tt.method, r.Method)
				require.Equal(t, "application/json", r.Header.Get("Content-Type"))

				if tt.body != nil {
					var payload map[string]any
					body, err := io.ReadAll(r.Body)
					require.NoError(t, err)
					err = json.Unmarshal(body, &payload)
					require.NoError(t, err)

					require.Equal(t, tt.body, payload)
				}
			})

			session := New(h)

			var res *Response
			switch tt.method {
			case http.MethodPost:
				res = session.PostJSON("/foo", nil, tt.body)
			case http.MethodPut:
				res = session.PutJSON("/foo", nil, tt.body)
			case http.MethodPatch:
				res = session.PatchJSON("/foo", nil, tt.body)
			case http.MethodDelete:
				res = session.DeleteJSON("/foo", nil, tt.body)
			default:
				t.Fatalf("unsupported method: %s", tt.method)
			}

			<-done
			require.Equal(t, tt.statusCode, res.Code())
		})
	}
}
