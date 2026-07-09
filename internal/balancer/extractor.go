package balancer

import (
	"errors"
	"net"
	"net/http"
)

var ErrNoKey = errors.New("unable to extract routing key")

type KeyExtractor interface {
	Extract(r *http.Request) (string, error)
}

type IpExtractor struct{}

func (IpExtractor) Extract(r *http.Request) (string, error) {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip, nil
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", ErrNoKey
	}
	return host, nil
}
