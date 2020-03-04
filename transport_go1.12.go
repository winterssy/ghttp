// +build go1.12
// +build !go1.13

package ghttp

import (
	"net"
	"net/http"
	"time"
)

// DefaultTransport returns a preset HTTP transport.
// It's a clone of http.DefaultTransport indeed.
func DefaultTransport() *http.Transport {
	return &http.Transport{
		Proxy: ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
