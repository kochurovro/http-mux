package main

import (
	"fmt"
	"net/http"
	"time"
)

const (
	ErrContentType = "Content-Type header is not application/json"
	ErrPost        = "Method is not POST"
)

const TimeoutMessage = "Your request has timed out\n"

func main() {
	m := midlewareWrapper(http.HandlerFunc(InspectorHandler), 100)
	srv := http.Server{
		Addr:    ":8080",
		Handler: http.TimeoutHandler(m, 10*time.Second, TimeoutMessage),
	}

	if err := srv.ListenAndServe(); err != nil {
		fmt.Printf("Server failed: %s\n", err)
	}

}
