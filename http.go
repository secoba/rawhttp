package rawhttp

import (
	"context"
	"github.com/secoba/rawhttp/client"
	"io"
	"net/http"

	retryablehttp "github.com/projectdiscovery/retryablehttp-go"
)

// DefaultClient is the default HTTP client for doing raw requests
var DefaultClient = Client{
	dialer:  new(dialer),
	Options: DefaultOptions,
}

// Get makes a GET request to a given URL
func Get(ctx context.Context, url string) (*client.Request, *http.Response, error) {
	return DefaultClient.Get(ctx, url)
}

// Post makes a POST request to a given URL
func Post(ctx context.Context, url string, mimetype string, r io.Reader) (*client.Request, *http.Response, error) {
	return DefaultClient.Post(ctx, url, mimetype, r)
}

// Do sends a http request and returns a response
func Do(ctx context.Context, req *http.Request) (*client.Request, *http.Response, error) {
	return DefaultClient.Do(ctx, req)
}

// Dor sends a retryablehttp request and returns a response
func Dor(ctx context.Context, req *retryablehttp.Request) (*client.Request, *http.Response, error) {
	return DefaultClient.Dor(ctx, req)
}

// DoRaw does a raw request with some configuration
func DoRaw(ctx context.Context, method, url, uripath string, headers map[string][]string, body io.Reader) (*client.Request, *http.Response, error) {
	return DefaultClient.DoRaw(ctx, method, url, uripath, headers, body)
}

// DoRawWithOptions does a raw request with some configuration
func DoRawWithOptions(ctx context.Context, method, url, uripath string, headers map[string][]string, body io.Reader, options *Options) (*client.Request, *http.Response, error) {
	return DefaultClient.DoRawWithOptions(ctx, method, url, uripath, headers, body, options)
}
