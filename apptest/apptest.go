package apptest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
)

// Represents a request, or requests to a given medium application.
type Session struct {
	CookieJar http.CookieJar
	app       http.Handler
}

// New creates a new "session" for the given application. It handles cookies,
// can follow redirects, etc.
func New(app http.Handler) *Session {
	req := &Session{}
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	req.CookieJar = jar
	req.app = app

	return req
}

// Get makes a GET request to the given route.
func (s *Session) Get(route string, headers http.Header) *Response {
	return s.makeRequest(http.MethodGet, route, headers, nil)
}

// Options makes a OPTIONS request to the given route.
func (s *Session) Options(route string, headers http.Header) *Response {
	return s.makeRequest(http.MethodOptions, route, headers, nil)
}

// Head makes a HEAD request to the given route.
func (s *Session) Head(route string, headers http.Header) *Response {
	return s.makeRequest(http.MethodHead, route, headers, nil)
}

// Post makes a POST request to the given route with the given headers and form values.
func (s *Session) PostForm(route string, headers http.Header, formValues url.Values) *Response {
	return s.formRequest(http.MethodPost, route, headers, formValues)
}

// Put makes a PUT request to the given route with the given headers and form values.
func (s *Session) PutForm(route string, headers http.Header, formValues url.Values) *Response {
	return s.formRequest(http.MethodPut, route, headers, formValues)
}

// Patch makes a PATCH request to the given route with the given headers and form values.
func (s *Session) PatchForm(route string, headers http.Header, formValues url.Values) *Response {
	return s.formRequest(http.MethodPatch, route, headers, formValues)
}

// Delete makes a DELETE request to the given route with the given headers and form values.
func (s *Session) DeleteForm(route string, headers http.Header, formValues url.Values) *Response {
	return s.formRequest(http.MethodDelete, route, headers, formValues)
}

// PostJSON makes a POST request to the given route with the given headers and JSON body.
func (s *Session) PostJSON(route string, headers http.Header, t any) *Response {
	return s.jsonRequest(http.MethodPost, route, headers, t)
}

// PutJSON makes a PUT request to the given route with the given headers and JSON body.
func (s *Session) PutJSON(route string, headers http.Header, t any) *Response {
	return s.jsonRequest(http.MethodPut, route, headers, t)
}

// PatchJSON makes a PATCH request to the given route with the given headers and JSON body.
func (s *Session) PatchJSON(route string, headers http.Header, t any) *Response {
	return s.jsonRequest(http.MethodPatch, route, headers, t)
}

// DeleteJSON makes a DELETE request to the given route with the given headers and JSON body.
func (s *Session) DeleteJSON(route string, headers http.Header, t any) *Response {
	return s.jsonRequest(http.MethodDelete, route, headers, t)
}

// FollowRedirect follows the redirect from the given response.
func (s *Session) FollowRedirect(res *Response) *Response {
	return s.makeRequest(http.MethodGet, res.Header().Get("Location"), nil, nil)
}

func (s *Session) formRequest(method string, route string, headers http.Header, formValues url.Values) *Response {
	if headers == nil {
		headers = make(http.Header)
	}

	headers.Set("Content-Type", "application/x-www-form-urlencoded")

	data := strings.NewReader(formValues.Encode())
	return s.makeRequest(method, route, headers, data)
}

func (s *Session) jsonRequest(method string, route string, headers http.Header, t any) *Response {
	if headers == nil {
		headers = make(http.Header)
	}

	headers.Set("Content-Type", "application/json")
	data, err := json.Marshal(t)

	if err != nil {
		panic(err)
	}

	return s.makeRequest(method, route, headers, bytes.NewReader(data))
}

func (s *Session) makeRequest(method string, route string, headers http.Header, data io.Reader) *Response {
	req := httptest.NewRequest(method, route, data)
	req.URL.Scheme = "http"
	req.URL.Host = "app.test"

	// Set headers
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Set cookies
	for _, cookie := range s.CookieJar.Cookies(req.URL) {
		req.AddCookie(cookie)
	}

	recorder := httptest.NewRecorder()

	s.app.ServeHTTP(recorder, req)
	s.CookieJar.SetCookies(req.URL, recorder.Result().Cookies())

	return &Response{RawResponse: recorder}
}
