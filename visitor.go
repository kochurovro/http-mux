package main

import (
	"context"
	"io/ioutil"
	"net/http"
)

var _ VisitorClient = &Visitor{}

type Visitor struct {
	NetClient *http.Client
}

type VisitorClient interface {
	Do(ctx context.Context, url string) (string, error)
}

func NewVisitor() *Visitor {
	var netTransport = &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
	}
	netClient := &http.Client{
		Transport: netTransport,
	}
	return &Visitor{NetClient: netClient}
}

// Do is making GET request for returning result of request
func (v *Visitor) Do(ctx context.Context, url string) (string, error) {
	_, cancel := context.WithCancel(ctx)
	req, err := http.NewRequest("GET", url, nil)

	resp, err := v.NetClient.Do(req.WithContext(ctx))
	if err != nil {
		cancel()
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		cancel()
		return "", err
	}
	return string(body), nil
}
