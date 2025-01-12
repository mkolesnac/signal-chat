package main

import (
	"net/http"
)

type MockRoundTripper struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
	WasCalled     bool
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.WasCalled = true
	return m.roundTripFunc(req)
}

func NewTestClient(roundTripFunc func(req *http.Request) (*http.Response, error)) *http.Client {
	return &http.Client{
		Transport: &MockRoundTripper{
			roundTripFunc: roundTripFunc,
		},
	}
}
