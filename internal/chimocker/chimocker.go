package chimocker

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type URLParams map[string]string

// WithUrlParam returns a pointer to a request object with the given URL params
// added to a new chi.Context object.
func WithURLParam(r *http.Request, key, value string) *http.Request {
	chiCtx := chi.NewRouteContext()
	req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))
	chiCtx.URLParams.Add(key, value)
	return req
}

// WithUrlParams returns a pointer to a request object with the given URL params
// added to a new chi.Context object. for single param assignment see WithUrlParam
func WithURLParams(r *http.Request, params URLParams) *http.Request {
	chiCtx := chi.NewRouteContext()
	for key, value := range params {
		chiCtx.URLParams.Add(key, value)
	}
	req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))
	return req
}
