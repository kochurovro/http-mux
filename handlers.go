package main

import "net/http"

func InspectorHandler(w http.ResponseWriter, r *http.Request) {
	if !isPostHandler(r.Method) {
		http.Error(w, ErrPost, http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, ErrContentType, http.StatusUnsupportedMediaType)
		return
	}

}

func isPostHandler(m string) bool {
	return m == http.MethodPost
}

func midlewareWrapper(h http.Handler, n int) http.HandlerFunc {
	sema := make(chan struct{}, n)

	return func(w http.ResponseWriter, r *http.Request) {
		sema <- struct{}{}
		defer func() { <-sema }()
		h.ServeHTTP(w, r)
	}
}
