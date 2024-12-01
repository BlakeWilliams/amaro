// Package csrf provides Cross Site Request Forgery (CSRF) protection
// middleware for amaro/httprouter.
//
// To use, combine with amaro/httprouter/middleware/session:
//
//	type SessionData struct {
//	   CSRFToken     *csrf.Token
//	}
//
// Then implement the CSRFable interface on your `RequestContext`
//
//	func (r *RequestContext) SetCSRF(t *csrf.Token) {
//	   r.Session().CSRFToken = t
//	}
//
//	func (r *RequestContext) CSRF() t *csrf.Token {
//	   r.Session().CSRFToken
//	}
package csrf
