package httprouter

import (
	"net/http"
)

// RequestContext is an interface that exposes the http.Request,
// http.ResponseWriter, and route params to a handler. Custom types can
// implement this interface to be passed to handlers of the router.
type RequestContext interface {
	// Request returns the original *http.Request
	Request() *http.Request
	// Writer returns a router.Response
	Response() Response
	// Params returns the parameters extracted from the URL path based on the
	// matched route.
	Params() map[string]string
	// MatchedPath returns the path that was matched by the router.
	MatchedPath() string
}

// BasicRequestContext is a basic implementation of RequestContext. It can be embedded in
// other types to provide a default implementation of the RequestContext interface.
type rootRequestContext struct {
	req         *http.Request
	res         Response
	params      map[string]string
	matchedPath string
}

var _ RequestContext = (*rootRequestContext)(nil)

func NewRequestContext(req *http.Request, res http.ResponseWriter, matchedPath string, routeParams map[string]string) *rootRequestContext {
	return &rootRequestContext{
		req:         req,
		res:         newResponseWriter(res),
		matchedPath: matchedPath,
		params:      routeParams,
	}
}

func (r *rootRequestContext) Request() *http.Request {
	return r.req
}

func (r *rootRequestContext) Response() Response {
	return r.res
}

func (r *rootRequestContext) Params() map[string]string {
	return r.params
}

func (r *rootRequestContext) MatchedPath() string {
	return r.matchedPath
}
