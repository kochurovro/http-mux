package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	ErrUnsupportedMediaType = "StatusUnsupportedMediaType"
	ErrMethodNotAllowed     = "Method Not Allowed"
	ErrBadRequest           = "Bad request;lu"
	ErrUnprocessableEntity  = "UnprocessableEntity"
	ErrInternalServerError  = "InternalServerError"
)

type Server struct {
	Port    int
	Limit   int
	Visitor VisitorClient
}

func NewServer(port int, limit int, visitor VisitorClient) *Server {
	return &Server{Port: port, Limit: limit, Visitor: visitor}
}

func (s *Server) Run(ctx context.Context) error {
	log.Printf("[INFO] activate server on :%d", s.Port)
	m := midlewareWrapper(http.HandlerFunc(s.InspectorHandler), s.Limit)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: http.TimeoutHandler(m, 10*time.Second, http.ErrHandlerTimeout.Error()),
	}
	go func() {
		<-ctx.Done()
		if err := srv.Close(); err != nil {
			log.Printf("[WARN] failed to close http server, %v", err)
		}
	}()

	return srv.ListenAndServe()
}
