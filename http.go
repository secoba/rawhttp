package rawhttp

import (
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
func Get(conn Conn, url string) (*client.Request, *http.Response, error) {
	return DefaultClient.Get(conn, url)
}

// Post makes a POST request to a given URL
func Post(conn Conn, url string, mimetype string, r io.Reader) (*client.Request, *http.Response, error) {
	return DefaultClient.Post(conn, url, mimetype, r)
}

// Do sends a http request and returns a response
func Do(conn Conn, req *http.Request) (*client.Request, *http.Response, error) {
	return DefaultClient.Do(conn, req)
}

// Dor sends a retryablehttp request and returns a response
func Dor(conn Conn, req *retryablehttp.Request) (*client.Request, *http.Response, error) {
	return DefaultClient.Dor(conn, req)
}

// DoRaw does a raw request with some configuration
func DoRaw(conn Conn, method, url, uripath string, headers map[string][]string, body io.Reader) (*client.Request, *http.Response, error) {
	return DefaultClient.DoRaw(conn, method, url, uripath, headers, body)
}

// DoRawWithOptions does a raw request with some configuration
func DoRawWithOptions(conn Conn, method, url, uripath string, headers map[string][]string, body io.Reader, options *Options) (*client.Request, *http.Response, error) {
	return DefaultClient.DoRawWithOptions(conn, method, url, uripath, headers, body, options)
}
