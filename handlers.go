package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const EmptyResponse = "{}"

type InspectorRequest struct {
	Urls []string `json:"urls"`
}

type InspectorResponse struct {
	Url  string `json:"url"`
	Data string `json:"data"`
}

func (s *Server) InspectorHandler(w http.ResponseWriter, r *http.Request) {
	if !isPostHandler(r.Method) {
		http.Error(w, ErrMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, ErrUnsupportedMediaType, http.StatusUnsupportedMediaType)
		return
	}
	if r.Body == nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(EmptyResponse))
		return
	}

	var d InspectorRequest
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, ErrUnprocessableEntity, http.StatusUnprocessableEntity)
		return
	}
	err = json.Unmarshal(body, &d)
	if err != nil {
		http.Error(w, ErrUnprocessableEntity, http.StatusUnprocessableEntity)
		return
	}
	if len(d.Urls) == 0 {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(EmptyResponse))
		return
	}
	if len(d.Urls) > 20 {
		http.Error(w, ErrUnprocessableEntity, http.StatusUnprocessableEntity)
		return
	}

	for i, _ := range d.Urls {
		_, err := url.ParseRequestURI(d.Urls[i])
		if err != nil {
			http.Error(w, ErrBadRequest, http.StatusBadRequest)
			return
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	sema := make(chan struct{}, 4)
	var wg sync.WaitGroup
	errc := make(chan error, 1)
	res := make([]InspectorResponse, 0)
	resultCh := make(chan InspectorResponse, 1)

	defer func() {
		res = nil
		close(sema)
		close(resultCh)
		close(errc)
	}()

	doVisit := func(url string) string {
		resp, err := s.Visitor.Do(ctx, url)
		if err != nil {
			errc <- err
			return ""
		}
		return resp
	}

	for i, _ := range d.Urls {
		sema <- struct{}{}
		wg.Add(1)

		go func(wg *sync.WaitGroup, url string) {
			defer func() {
				<-sema
			}()

			resp := doVisit(url)

			select {
			case err = <-errc:
				cancel()
				break
			default:
			}

			resultCh <- InspectorResponse{Url: url, Data: resp}
		}(&wg, d.Urls[i])
	}

	go func(q <-chan InspectorResponse) {
		for t := range resultCh {
			res = append(res, t)
			wg.Done()
		}
	}(resultCh)

	wg.Wait()
	if err != nil {
		http.Error(w, ErrBadRequest, http.StatusBadRequest)
		return
	}

	js, err := json.Marshal(res)
	if err != nil {
		http.Error(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
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
