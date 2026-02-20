package httpclient

import (
	"crypto/tls"
	"net/http"
	"time"
)

const defaultTimeout = 30 * time.Second

func New(skipTLSVerify bool) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: skipTLSVerify, // #nosec G402 -- user-controlled for self-signed lab certs
	}

	return &http.Client{
		Timeout:   defaultTimeout,
		Transport: transport,
	}
}
