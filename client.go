package rawhttp

import (
	"fmt"
	"github.com/projectdiscovery/fastdialer/fastdialer"
	"github.com/projectdiscovery/gologger"
	retryablehttp "github.com/projectdiscovery/retryablehttp-go"
	urlutil "github.com/projectdiscovery/utils/url"
	"github.com/secoba/rawhttp/client"
	"io"
	"net/http"
	"strings"
)

// Client is a client for making raw http requests with go
type Client struct {
	dialer     Dialer
	Options    *Options
	connection Conn
}

// AutomaticHostHeader sets Host header for requests automatically
func AutomaticHostHeader(enable bool) {
	DefaultClient.Options.AutomaticHostHeader = enable
}

// AutomaticContentLength performs automatic calculation of request content length.
func AutomaticContentLength(enable bool) {
	DefaultClient.Options.AutomaticContentLength = enable
}

// NewClient creates a new rawhttp client with provided options
func NewClient(options *Options) *Client {
	c := &Client{
		dialer:  new(dialer),
		Options: options,
	}
	if options.FastDialer == nil {
		var err error
		opts := fastdialer.DefaultOptions
		opts.DialerTimeout = options.Timeout / 2 // fastdialer会增加2倍超时
		options.FastDialer, err = fastdialer.NewDialer(opts)
		if err != nil {
			gologger.Error().Msgf("Could not create fast dialer: %s\n", err)
		}
	}
	return c
}

// Head makes a HEAD request to a given URL
func (c *Client) Head(conn Conn, url string) (*client.Request, *http.Response, error) {
	defer func() {
		if conn != nil {
			conn.Stop()
		}
	}()
	return c.DoRaw(conn, "HEAD", url, "", nil, nil)
}

// Get makes a GET request to a given URL
func (c *Client) Get(conn Conn, url string) (*client.Request, *http.Response, error) {
	defer func() {
		if conn != nil {
			conn.Stop()
		}
	}()
	return c.DoRaw(conn, "GET", url, "", nil, nil)
}

// Post makes a POST request to a given URL
func (c *Client) Post(conn Conn, url string, mimetype string, body io.Reader) (*client.Request, *http.Response, error) {
	defer func() {
		if conn != nil {
			conn.Stop()
		}
	}()
	headers := make(map[string][]string)
	headers["Content-Type"] = []string{mimetype}
	return c.DoRaw(conn, "POST", url, "", headers, body)
}

// Do sends a http request and returns a response
func (c *Client) Do(conn Conn, req *http.Request) (*client.Request, *http.Response, error) {
	defer func() {
		if conn != nil {
			conn.Stop()
		}
	}()

	method := req.Method
	headers := req.Header
	url := req.URL.String()
	body := req.Body

	return c.DoRaw(conn, method, url, "", headers, body)
}

// Dor sends a retryablehttp request and returns the response
func (c *Client) Dor(conn Conn, req *retryablehttp.Request) (*client.Request, *http.Response, error) {
	defer func() {
		if conn != nil {
			conn.Stop()
		}
	}()

	method := req.Method
	headers := req.Header
	url := req.URL.String()
	body := req.Body

	return c.DoRaw(conn, method, url, "", headers, body)
}

// DoRaw does a raw request with some configuration
func (c *Client) DoRaw(conn Conn, method, url, uripath string, headers map[string][]string, body io.Reader) (*client.Request, *http.Response, error) {
	defer func() {
		if conn != nil {
			conn.Stop()
		}
	}()

	redirectStatus := &RedirectStatus{
		FollowRedirects: c.Options.FollowRedirects,
		MaxRedirects:    c.Options.MaxRedirects,
	}
	return c.do(conn, method, url, uripath, headers, body, redirectStatus, c.Options)
}

// DoRawWithOptions performs a raw request with additional options
func (c *Client) DoRawWithOptions(conn Conn, method, url, uripath string, headers map[string][]string, body io.Reader, options *Options) (*client.Request, *http.Response, error) {
	defer func() {
		if conn != nil {
			conn.Stop()
		}
	}()

	redirectStatus := &RedirectStatus{
		FollowRedirects: options.FollowRedirects,
		MaxRedirects:    c.Options.MaxRedirects,
	}
	return c.do(conn, method, url, uripath, headers, body, redirectStatus, options)
}

// Close closes client and any resources it holds
func (c *Client) Close() {
	if c.Options.FastDialer != nil {
		c.Options.FastDialer.Close()
	}
}

func (c *Client) getConn(protocol, host string, options *Options) (Conn, error) {
	var (
		err   error
		conn2 Conn
	)

	if options.Proxy != "" {
		conn2, err = c.dialer.DialWithProxy(protocol, host, c.Options.Proxy, c.Options.ProxyDialTimeout, options)
	} else {
		//conn2, err = c.dialer.Dial(protocol, host, options)
		if options.Timeout > 0 {
			conn2, err = c.dialer.DialTimeout(protocol, host, options.Timeout/2, options)
		} else {
			conn2, err = c.dialer.Dial(protocol, host, options)
		}
	}

	return conn2, err
}

func (c *Client) CreateConnection(url string, options *Options) (Conn, error) {
	protocol := "http"
	if strings.HasPrefix(strings.ToLower(url), "https://") {
		protocol = "https"
	}

	u, err := urlutil.ParseURL(url, true)
	if err != nil {
		return nil, err
	}

	host := u.Host

	if !strings.Contains(host, ":") {
		if protocol == "https" {
			host += ":443"
		} else {
			host += ":80"
		}
	}

	// standard path
	//path := u.Path
	//if path == "" {
	//	path = "/"
	//}
	//if !u.Params.IsEmpty() {
	//	path += "?" + u.Params.Encode()
	//}
	// override if custom one is specified
	//if uripath != "" {
	//	path = uripath
	//}

	if strings.HasPrefix(url, "https://") {
		protocol = "https"
	}

	getConn, err := c.getConn(protocol, host, options)
	if err != nil {
		return nil, err
	}

	// set timeout if any
	if options.Timeout > 0 {
		getConn.SetTimeout(options.Timeout)
	}

	return getConn, nil
}

func (c *Client) do(getConn Conn, method, url, uripath string, headers map[string][]string,
	body io.Reader, redirectStatus *RedirectStatus, options *Options) (*client.Request, *http.Response, error) {

	protocol := "http"
	if strings.HasPrefix(strings.ToLower(url), "https://") {
		protocol = "https"
	}

	if headers == nil {
		headers = make(map[string][]string)
	}
	u, err := urlutil.ParseURL(url, true)
	if err != nil {
		return nil, nil, err
	}

	host := u.Host
	if options.AutomaticHostHeader {
		// add automatic space
		headers["Host"] = []string{fmt.Sprintf(" %s", host)}
	}

	if !strings.Contains(host, ":") {
		if protocol == "https" {
			host += ":443"
		} else {
			host += ":80"
		}
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

	if strings.HasPrefix(url, "https://") {
		protocol = "https"
	}

	//getConn, err := c.getConn(protocol, host, options)
	//if err != nil {
	//	return nil, nil, nil, err
	//}
	//
	//// set timeout if any
	//if options.Timeout > 0 {
	//	getConn.SetTimeout(options.Timeout)
	//}

	req := toRequest(method, u.Host, path, nil, headers, body, options)
	req.AutomaticContentLength = options.AutomaticContentLength
	req.AutomaticHost = options.AutomaticHostHeader

	if err2 := getConn.WriteRequest(req); err2 != nil {
		return req, nil, err2
	}
	resp, err2 := getConn.ReadResponse(options.ForceReadAllBody)
	if err2 != nil {
		return req, nil, err2
	}

	r, err := toHTTPResponse(getConn, resp)
	if err != nil {
		return req, nil, err
	}

	if resp.Status.IsRedirect() && redirectStatus.FollowRedirects && redirectStatus.Current <= redirectStatus.MaxRedirects {
		//fmt.Println(redirectStatus.FollowRedirects)
		// consume the response body
		_, err3 := io.Copy(io.Discard, r.Body)
		if err4 := firstErr(err3, r.Body.Close()); err4 != nil {
			return req, nil, err4
		}
		loc := headerValue(r.Header, "Location")
		if strings.HasPrefix(loc, "/") {
			loc = fmt.Sprintf("%s://%s%s", protocol, host, loc)
		}
		redirectStatus.Current++
		return c.do(getConn, method, loc, uripath, headers, body, redirectStatus, options)
	}

	return req, r, err
}

// RedirectStatus is the current redirect status for the request
type RedirectStatus struct {
	FollowRedirects bool
	MaxRedirects    int
	Current         int
}
