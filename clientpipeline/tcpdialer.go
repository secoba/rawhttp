package clientpipeline

// Original Source: https://github.com/valyala/fasthttp

import (
	"context"
	"errors"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

func Dial(ctx context.Context, addr string) (net.Conn, error) {
	return defaultDialer.Dial(ctx, addr)
}

func DialTimeout(ctx context.Context, addr string, timeout time.Duration) (net.Conn, error) {
	return defaultDialer.DialTimeout(ctx, addr, timeout)
}

func DialDualStack(ctx context.Context, addr string) (net.Conn, error) {
	return defaultDialer.DialDualStack(ctx, addr)
}

func DialDualStackTimeout(ctx context.Context, addr string, timeout time.Duration) (net.Conn, error) {
	return defaultDialer.DialDualStackTimeout(ctx, addr, timeout)
}

var (
	defaultDialer = &TCPDialer{Concurrency: 1000}
)

// Resolver represents interface of the tcp resolver.
type Resolver interface {
	LookupIPAddr(context.Context, string) (names []net.IPAddr, err error)
}

// TCPDialer contains options to control a group of Dial calls.
type TCPDialer struct {
	Concurrency   int
	LocalAddr     *net.TCPAddr
	Resolver      Resolver
	tcpAddrsLock  sync.Mutex
	tcpAddrsMap   map[string]*tcpAddrEntry
	concurrencyCh chan struct{}
	once          sync.Once
}

func (d *TCPDialer) Dial(ctx context.Context, addr string) (net.Conn, error) {
	return d.dial(ctx, addr, false, DefaultDialTimeout)
}

func (d *TCPDialer) DialTimeout(ctx context.Context, addr string, timeout time.Duration) (net.Conn, error) {
	return d.dial(ctx, addr, false, timeout)
}

func (d *TCPDialer) DialDualStack(ctx context.Context, addr string) (net.Conn, error) {
	return d.dial(ctx, addr, true, DefaultDialTimeout)
}

func (d *TCPDialer) DialDualStackTimeout(ctx context.Context, addr string, timeout time.Duration) (net.Conn, error) {
	return d.dial(ctx, addr, true, timeout)
}

func (d *TCPDialer) dial(ctx context.Context, addr string, dualStack bool, timeout time.Duration) (net.Conn, error) {
	d.once.Do(func() {
		if d.Concurrency > 0 {
			d.concurrencyCh = make(chan struct{}, d.Concurrency)
		}
		d.tcpAddrsMap = make(map[string]*tcpAddrEntry)
		go d.tcpAddrsClean()
	})

	addrs, idx, err := d.getTCPAddrs(ctx, addr, dualStack)
	if err != nil {
		return nil, err
	}
	network := "tcp4"
	if dualStack {
		network = "tcp"
	}

	var conn net.Conn
	n := uint32(len(addrs))
	deadline := time.Now().Add(timeout)
	for n > 0 {
		conn, err = d.tryDial(network, &addrs[idx%n], deadline, d.concurrencyCh)
		if err == nil {
			return conn, nil
		}
		if err == ErrDialTimeout {
			return nil, err
		}
		idx++
		n--
	}
	return nil, err
}

func (d *TCPDialer) tryDial(network string, addr *net.TCPAddr, deadline time.Time, concurrencyCh chan struct{}) (net.Conn, error) {
	timeout := -time.Since(deadline)
	if timeout <= 0 {
		return nil, ErrDialTimeout
	}

	if concurrencyCh != nil {
		select {
		case concurrencyCh <- struct{}{}:
		default:
			tc := time.NewTimer(timeout)
			isTimeout := false
			select {
			case concurrencyCh <- struct{}{}:
			case <-tc.C:
				isTimeout = true
			}
			tc.Stop()
			if isTimeout {
				return nil, ErrDialTimeout
			}
		}
	}

	chv := dialResultChanPool.Get()
	if chv == nil {
		chv = make(chan dialResult, 1)
	}
	ch := chv.(chan dialResult)
	go func() {
		var dr dialResult
		dr.conn, dr.err = net.DialTCP(network, d.LocalAddr, addr)
		ch <- dr
		if concurrencyCh != nil {
			<-concurrencyCh
		}
	}()

	var (
		conn net.Conn
		err  error
	)

	tc := time.NewTimer(timeout)
	select {
	case dr := <-ch:
		conn = dr.conn
		err = dr.err
		dialResultChanPool.Put(ch)
	case <-tc.C:
		err = ErrDialTimeout
	}
	tc.Stop()

	return conn, err
}

var dialResultChanPool sync.Pool

type dialResult struct {
	conn net.Conn
	err  error
}

// ErrDialTimeout is returned when TCP dialing is timed out.
var ErrDialTimeout = errors.New("dialing to the given TCP address timed out")

// DefaultDialTimeout is timeout used by Dial and DialDualStack
// for establishing TCP connections.
const DefaultDialTimeout = 3 * time.Second

type tcpAddrEntry struct {
	addrs    []net.TCPAddr
	addrsIdx uint32

	resolveTime time.Time
	pending     bool
}

// DefaultDNSCacheDuration is the duration for caching resolved TCP addresses
// by Dial* functions.
const DefaultDNSCacheDuration = time.Minute

func (d *TCPDialer) tcpAddrsClean() {
	expireDuration := 2 * DefaultDNSCacheDuration
	for {
		time.Sleep(time.Second)
		t := time.Now()

		d.tcpAddrsLock.Lock()
		for k, e := range d.tcpAddrsMap {
			if t.Sub(e.resolveTime) > expireDuration {
				delete(d.tcpAddrsMap, k)
			}
		}
		d.tcpAddrsLock.Unlock()
	}
}

func (d *TCPDialer) getTCPAddrs(ctx context.Context, addr string, dualStack bool) ([]net.TCPAddr, uint32, error) {
	d.tcpAddrsLock.Lock()
	e := d.tcpAddrsMap[addr]
	if e != nil && !e.pending && time.Since(e.resolveTime) > DefaultDNSCacheDuration {
		e.pending = true
		e = nil
	}
	d.tcpAddrsLock.Unlock()

	if e == nil {
		addrs, err := resolveTCPAddrs(ctx, addr, dualStack, d.Resolver)
		if err != nil {
			d.tcpAddrsLock.Lock()
			e = d.tcpAddrsMap[addr]
			if e != nil && e.pending {
				e.pending = false
			}
			d.tcpAddrsLock.Unlock()
			return nil, 0, err
		}

		e = &tcpAddrEntry{
			addrs:       addrs,
			resolveTime: time.Now(),
		}

		d.tcpAddrsLock.Lock()
		d.tcpAddrsMap[addr] = e
		d.tcpAddrsLock.Unlock()
	}

	idx := atomic.AddUint32(&e.addrsIdx, 1)
	return e.addrs, idx, nil
}

func resolveTCPAddrs(pCtx context.Context, addr string, dualStack bool, resolver Resolver) ([]net.TCPAddr, error) {
	host, portS, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(portS)
	if err != nil {
		return nil, err
	}

	if resolver == nil {
		resolver = net.DefaultResolver
	}

	ctx, _ := context.WithCancel(pCtx)
	ipaddrs, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}

	n := len(ipaddrs)
	addrs := make([]net.TCPAddr, 0, n)
	for i := 0; i < n; i++ {
		ip := ipaddrs[i]
		if !dualStack && ip.IP.To4() == nil {
			continue
		}
		addrs = append(addrs, net.TCPAddr{
			IP:   ip.IP,
			Port: port,
			Zone: ip.Zone,
		})
	}
	if len(addrs) == 0 {
		return nil, errNoDNSEntries
	}
	return addrs, nil
}

var errNoDNSEntries = errors.New("couldn't find DNS entries for the given domain. Try using DialDualStack")
