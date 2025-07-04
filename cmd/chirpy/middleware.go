package main

import (
	"net/http"
)

func (a *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		r.Header.Add("Cache-Control", "no-cache")
		a.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
