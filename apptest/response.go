package apptest

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
)

var RedirectCodes = []int{
	http.StatusMovedPermanently,
	http.StatusFound,
	http.StatusSeeOther,
	http.StatusTemporaryRedirect,
	http.StatusPermanentRedirect,
}

type Response struct {
	RawResponse *httptest.ResponseRecorder
	bodyCache   *bytes.Buffer
	mu          sync.Mutex
}

func (r *Response) Code() int {
	return r.RawResponse.Code
}

func (r *Response) IsRedirect() bool {
	for _, redirectCode := range RedirectCodes {
		if r.Code() == redirectCode {
			return true
		}
	}

	return false
}

func (r *Response) Body() *bytes.Buffer {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.bodyCache == nil {
		body, _ := io.ReadAll(r.RawResponse.Result().Body)
		r.bodyCache = bytes.NewBuffer(body)
	}

	return r.bodyCache
}

func (r *Response) Header() http.Header {
	return r.RawResponse.Result().Header
}
