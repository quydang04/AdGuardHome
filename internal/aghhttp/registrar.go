package aghhttp

import (
	"fmt"
	"net/http"
	"sync"
)

// Registrar registers an HTTP handler for a method and path.
//
// TODO(s.chzhen):  Implement [httputil.Router].
type Registrar interface {
	Register(method, path string, h http.HandlerFunc)
}

// EmptyRegistrar is an implementation of [Registrar] that does nothing.
type EmptyRegistrar struct{}

// type check
var _ Registrar = EmptyRegistrar{}

// Register implements the [Registrar] interface.
func (EmptyRegistrar) Register(_, _ string, _ http.HandlerFunc) {}

// WrapFunc is a wrapper function that builds an HTTP handler for a route.
type WrapFunc func(method string, h http.HandlerFunc) (wrapped http.Handler)

// DefaultRegistrar is an implementation of [Registrar] that registers handlers
// after applying a user-provided wrapper function.
type DefaultRegistrar struct {
	mux    *http.ServeMux
	wrapFn WrapFunc

	mu     sync.RWMutex
	routes map[string]*defaultRoute
}

// NewDefaultRegistrar returns a new properly initialized *DefaultRegistrar.
// mux and wrap must not be nil.
func NewDefaultRegistrar(mux *http.ServeMux, wrap WrapFunc) (r *DefaultRegistrar) {
	return &DefaultRegistrar{
		mux:    mux,
		wrapFn: wrap,
		routes: make(map[string]*defaultRoute),
	}
}

// type check
var _ Registrar = (*DefaultRegistrar)(nil)

// Register implements the [Registrar] interface.
func (r *DefaultRegistrar) Register(method, path string, h http.HandlerFunc) {
	if path == "" {
		panic("aghhttp: empty path")
	}

	wrapped := r.wrapFn(method, h)

	r.mu.Lock()
	defer r.mu.Unlock()

	route, exists := r.routes[path]
	if !exists {
		route = &defaultRoute{
			methods: make(map[string]http.Handler),
		}
		r.routes[path] = route

		pathCopy := path
		r.mux.Handle(pathCopy, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			r.serve(pathCopy, w, req)
		}))
	}

	if method == "" {
		if route.any != nil {
			panic(fmt.Sprintf("aghhttp: handler already registered for pattern %q", path))
		}

		route.any = wrapped
		if route.fallback == nil {
			route.fallback = wrapped
		}

		return
	}

	if _, dup := route.methods[method]; dup {
		panic(fmt.Sprintf(
			"aghhttp: handler for method %q already registered for pattern %q",
			method,
			path,
		))
	}

	route.methods[method] = wrapped
	if route.fallback == nil {
		route.fallback = wrapped
	}
}

// serve dispatches the request for the registered path to a handler based on
// the HTTP method.  It must be called with path already validated.
func (r *DefaultRegistrar) serve(path string, w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	route := r.routes[path]

	handler := route.handler(req.Method)
	if handler == nil {
		handler = route.fallback
	}
	r.mu.RUnlock()

	if handler == nil {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

	handler.ServeHTTP(w, req)
}

type defaultRoute struct {
	methods  map[string]http.Handler
	any      http.Handler
	fallback http.Handler
}

// handler returns handler for specific method or a fallback handler.
func (r *defaultRoute) handler(method string) http.Handler {
	if h, ok := r.methods[method]; ok {
		return h
	}

	return r.any
}
