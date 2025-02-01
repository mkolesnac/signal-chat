package apiclient

import (
	"net/http"
)

type SpyRoundTripper struct {
	Request  *http.Request
	Response *http.Response
}

func (s *SpyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	s.Request = req
	if s.Response != nil {
		return s.Response, nil
	}
	return &http.Response{StatusCode: http.StatusOK}, nil
}
