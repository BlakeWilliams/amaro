package metal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/blakewilliams/amaro/httprouter"
	"github.com/stretchr/testify/require"
)

func TestRewrite(t *testing.T) {
	router := httprouter.New(func(r httprouter.RequestContext) httprouter.RequestContext { return r })
	router.UseMetal(MethodRewrite)
	router.Delete("/", func(ctx context.Context, rc httprouter.RequestContext) {
		rc.Response().WriteHeader(http.StatusOK)
	})

	formData := url.Values{}
	formData.Set("_method", "DELETE")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Result().StatusCode)
}
