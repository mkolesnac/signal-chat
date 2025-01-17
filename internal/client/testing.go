package client

import (
	"net/http"
)

type SpyRoundTripper struct {
	Request *http.Request
}

func (s *SpyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	s.Request = req
	return &http.Response{StatusCode: http.StatusOK}, nil
}
