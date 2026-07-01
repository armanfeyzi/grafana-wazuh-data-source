package httpclient

import (
	"context"
	"net/http"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	sdkhttpclient "github.com/grafana/grafana-plugin-sdk-go/backend/httpclient"
)

const defaultTimeout = 30 * time.Second

// New creates an HTTP client using Grafana SDK options from datasource settings.
func New(ctx context.Context, settings backend.DataSourceInstanceSettings) (*http.Client, error) {
	opts, err := settings.HTTPClientOptions(ctx)
	if err != nil {
		return nil, err
	}
	if opts.Timeouts == nil {
		opts.Timeouts = &sdkhttpclient.TimeoutOptions{}
	}
	opts.Timeouts.Timeout = defaultTimeout
	return sdkhttpclient.New(opts)
}

// NewTest creates an HTTP client for unit tests.
func NewTest(skipTLSVerify bool) (*http.Client, error) {
	return sdkhttpclient.New(sdkhttpclient.Options{
		Timeouts: &sdkhttpclient.TimeoutOptions{Timeout: defaultTimeout},
		TLS: &sdkhttpclient.TLSOptions{
			InsecureSkipVerify: skipTLSVerify, // #nosec G402
		},
	})
}
