package proxy

import (
	"net"
	"net/url"
	"time"

	p "golang.org/x/net/proxy"
)

func Socks5Dialer(proxyAddr string, timeout time.Duration) DialFunc {
	var (
		u      *url.URL
		err    error
		dialer p.Dialer
	)
	if u, err = url.Parse(proxyAddr); err == nil {
		dialer, err = p.FromURL(u, p.Direct)
		//dialer, err = p.SOCKS5("tcp", proxyAddr, nil, &net.Dialer{Timeout: timeout * time.Second})
	}
	return func(addr string) (net.Conn, error) {
		if err != nil {
			return nil, err
		}
		return dialer.Dial("tcp", addr)
	}
}
