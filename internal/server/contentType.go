package server

import (
	"net/http"
	"strings"
)

func enforceContentTypeJSON(next http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctHeader := req.Header.Get("Content-Type")
		if ctHeader == "" {
			http.Error(res, "missing content type header", http.StatusBadRequest)
			return
		}

		contentType := strings.ToLower(strings.TrimSpace(strings.Split(ctHeader, ";")[0]))
		if contentType != "application/json" {
			http.Error(res, "wrong content type", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(res, req)
	}
}
