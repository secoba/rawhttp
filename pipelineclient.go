package pkg

import (
	"context"
	"io"
	"net/http"

	retryablehttp "github.com/projectdiscovery/retryablehttp-go"
	urlutil "github.com/projectdiscovery/utils/url"
	"github.com/secoba/rawhttp/clientpipeline"
)

// PipelineClient is a client for making pipelined http requests
type PipelineClient struct {
	client  *clientpipeline.PipelineClient
	options PipelineOptions
}

// NewPipelineClient creates a new pipelined http request client
func NewPipelineClient(ctx context.Context, options PipelineOptions) *PipelineClient {
	client := &PipelineClient{
		client: &clientpipeline.PipelineClient{
			Ctx:                ctx,
			Dial:               options.Dialer,
			Addr:               options.Host,
			MaxConns:           options.MaxConnections,
			MaxPendingRequests: options.MaxPendingRequests,
			ReadTimeout:        options.Timeout,
		},
		options: options,
	}
	return client
}

// Head makes a HEAD request to a given URL
func (c *PipelineClient) Head(url string) (*clientpipeline.Request, *http.Response, error) {
	return c.DoRaw("HEAD", url, "", nil, nil, nil)
}

// Get makes a GET request to a given URL
func (c *PipelineClient) Get(url string) (*clientpipeline.Request, *http.Response, error) {
	return c.DoRaw("GET", url, "", nil, nil, nil)
}

// Post makes a POST request to a given URL
func (c *PipelineClient) Post(url string, mimetype string, body io.Reader) (*clientpipeline.Request, *http.Response, error) {
	headers := make(map[string][]string)
	headers["Content-Type"] = []string{mimetype}
	return c.DoRaw("POST", url, "", headers, body, nil)
}

// Do sends a http request and returns a response
func (c *PipelineClient) Do(req *http.Request) (*clientpipeline.Request, *http.Response, error) {
	method := req.Method
	headers := req.Header
	url := req.URL.String()
	body := req.Body
	return c.DoRaw(method, url, "", headers, body, nil)
}

// Dor sends a retryablehttp request and returns a response
func (c *PipelineClient) Dor(req *retryablehttp.Request) (*clientpipeline.Request, *http.Response, error) {
	method := req.Method
	headers := req.Header
	url := req.URL.String()
	body := req.Body

	return c.do(method, url, "", headers, body, nil, c.options)
}

// DoRaw does a raw request with some configuration
func (c *PipelineClient) DoRaw(method, url, uripath string, headers map[string][]string, body io.Reader, raw []byte) (*clientpipeline.Request, *http.Response, error) {
	return c.do(method, url, uripath, headers, body, raw, c.options)
}

// DoRawWithOptions performs a raw request with additional options
func (c *PipelineClient) DoRawWithOptions(method, url, uripath string, headers map[string][]string, body io.Reader, raw []byte, options PipelineOptions) (*clientpipeline.Request, *http.Response, error) {
	return c.do(method, url, uripath, headers, body, raw, options)
}

func (c *PipelineClient) do(method, url, uripath string, headers map[string][]string, body io.Reader, raw []byte, options PipelineOptions) (*clientpipeline.Request, *http.Response, error) {
	if headers == nil {
		headers = make(map[string][]string)
	}
	u, err := urlutil.ParseURL(url, true)
	if err != nil {
		return nil, nil, err
	}
	// standard path
	path := u.Path
	if path == "" {
		path = "/"
	}
	if !u.Params.IsEmpty() {
		path += "?" + u.Params.Encode()
	}
	// override if custom one is specified
	if uripath != "" {
		path = uripath
	}

	req := clientpipeline.ToRequest(
		method, u.Host, path, nil, headers, body,
		raw, options.AutomaticHostHeader, options.AutomaticContentLength)
	var resp clientpipeline.Response

	err = c.client.Do(req, &resp)

	// response => net/http response
	r := http.Response{
		StatusCode:    resp.Status.Code,
		ContentLength: resp.ContentLength(),
		Header:        make(http.Header),
	}

	for _, header := range resp.Headers {
		r.Header.Set(header.Key, header.Value)
	}

	r.Body = io.NopCloser(resp.Body)

	return req, &r, err
}
