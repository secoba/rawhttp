package pkg

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/projectdiscovery/fastdialer/fastdialer"
	"github.com/secoba/rawhttp/client"
	"github.com/secoba/rawhttp/proxy"
)

// Dialer can dial a remote HTTP server.
type Dialer interface {
	Dial(protocol, addr string, options *Options) (Conn, error) // Dial dials a remote http server returning a Conn.
	DialWithProxy(protocol, addr, proxyURL string, timeout time.Duration, options *Options) (Conn, error)
	DialTimeout(protocol, addr string, timeout time.Duration, options *Options) (Conn, error) // Dial dials a remote http server with timeout returning a Conn.
}

type dialer struct {
	sync.Mutex // protects following fields
	//conns      map[string][]Conn // maps addr to a, possibly empty, slice of existing Conns
}

func (d *dialer) Dial(protocol, addr string, options *Options) (Conn, error) {
	return d.dialTimeout(protocol, addr, 0, options)
}

func (d *dialer) DialTimeout(protocol, addr string, timeout time.Duration, options *Options) (Conn, error) {
	return d.dialTimeout(protocol, addr, timeout, options)
}

func (d *dialer) dialTimeout(protocol, addr string, timeout time.Duration, options *Options) (Conn, error) {
	d.Lock()
	//if d.conns == nil {
	//	d.conns = make(map[string][]Conn)
	//}
	//if c, ok := d.conns[addr]; ok {
	//	if len(c) > 0 {
	//		conn2 := c[0]
	//		c[0] = c[len(c)-1]
	//		d.Unlock()
	//		return conn2, nil
	//	}
	//}
	d.Unlock()
	c, err := clientDial(protocol, addr, timeout, options)
	return &conn{
		Client: client.NewClient(c),
		Conn:   c,
		dialer: d,
	}, err
}

func (d *dialer) DialWithProxy(protocol, addr, proxyURL string, timeout time.Duration, options *Options) (Conn, error) {
	var c net.Conn
	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("unsupported proxy error: %w", err)
	}
	switch u.Scheme {
	case "http":
		c, err = proxy.HTTPFastDialer(proxyURL, timeout, options.FastDialer)(addr)
	case "socks5", "socks5h":
		c, err = proxy.Socks5Dialer(proxyURL, timeout)(addr)
	default:
		return nil, fmt.Errorf("unsupported proxy protocol: %s", proxyURL)
	}
	if err != nil {
		return nil, fmt.Errorf("proxy error: %w", err)
	}
	if protocol == "https" {
		if c, err = TlsHandshake(c, addr, timeout); err != nil {
			return nil, fmt.Errorf("tls handshake error: %w", err)
		}
	}

	return &conn{
		Client: client.NewClient(c),
		Conn:   c,
		dialer: d,
	}, err
}

func clientDial(protocol, addr string, timeout time.Duration, options *Options) (net.Conn, error) {
	//if timeout > 0 {
	//	ctx, cancel = context.WithTimeout(pCtx, timeout)
	//	defer cancel()
	//} else {
	//	ctx = pCtx
	//}

	// http
	if protocol == "http" {
		if options.FastDialer != nil {
			return options.FastDialer.Dial(context.Background(), "tcp", addr)
		} else if timeout > 0 {
			return net.DialTimeout("tcp", addr, timeout)
		}
		return net.Dial("tcp", addr)
	}

	// https
	tlsConfig := &tls.Config{InsecureSkipVerify: true, Renegotiation: tls.RenegotiateOnceAsClient}
	if options.SNI != "" {
		tlsConfig.ServerName = options.SNI
	}

	if options.FastDialer == nil {
		// always use fastdialer tls dial if available
		opts := fastdialer.DefaultOptions
		if timeout > 0 {
			opts.DialerTimeout = timeout
		}
		var err error
		options.FastDialer, err = fastdialer.NewDialer(opts)
		// use net.Dialer if fastdialer tls dial is not available
		if err != nil {
			var d *net.Dialer
			if timeout > 0 {
				d = &net.Dialer{Timeout: timeout}
			} else {
				d = &net.Dialer{Timeout: 8 * time.Second} // should be more than enough
			}
			return tls.DialWithDialer(d, "tcp", addr, tlsConfig)
		}
	}

	return options.FastDialer.DialTLS(context.Background(), "tcp", addr)
}

// TlsHandshake tls handshake on a plain connection
func TlsHandshake(conn net.Conn, addr string, timeout time.Duration) (net.Conn, error) {
	var (
		ctx    = context.Background()
		cancel context.CancelFunc
	)

	colonPos := strings.LastIndex(addr, ":")
	if colonPos == -1 {
		colonPos = len(addr)
	}
	hostname := addr[:colonPos]

	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	tlsConn := tls.Client(conn, &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         hostname,
	})
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		return nil, err
	}
	return tlsConn, nil
}

// Conn is an interface implemented by a connection
type Conn interface {
	client.Client
	io.Closer

	SetTimeout(duration time.Duration)
	Release()
	Stop() error
}

type conn struct {
	client.Client
	net.Conn
	*dialer
}

func (c *conn) Release() {
	//c.dialer.Lock()
	//defer c.dialer.Unlock()
	//addr := c.Conn.RemoteAddr().String()
	//c.dialer.conns[addr] = append(c.dialer.conns[addr], c)
}

func (c *conn) Stop() error {
	return c.Conn.Close()
}

func (c *conn) SetTimeout(timeout time.Duration) {
	_ = c.Conn.SetDeadline(time.Now().Add(timeout))
	//_ = c.SetReadDeadline(time.Now().Add(timeout))
	//_ = c.SetWriteDeadline(time.Now().Add(timeout))
}
