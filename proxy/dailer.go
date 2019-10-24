package proxy

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/justenwalker/squiggly/auth"
	"github.com/justenwalker/squiggly/logging"
)

var errProxyAuth = errors.New("proxy auth required")

type ProxyDialer struct {
	Logger logging.Logger
	Auth   *auth.Auth
	Host   *url.URL
	Dialer func(ctx context.Context, network, addr string) (net.Conn, error)
}

func host(u *url.URL) string {
	if strings.IndexRune(u.Host, ':') == -1 {
		switch u.Scheme {
		case "", "http":
			return u.Host + ":80"
		case "https":
			return u.Host + ":443"
		}
	}
	return u.Host
}

func (d *ProxyDialer) log(msg string) {
	if d.Logger == nil {
		return
	}
	d.Logger.Log(msg)
}

func (d *ProxyDialer) logf(format string, v ...interface{}) {
	if d.Logger == nil {
		return
	}
	d.Logger.Log(fmt.Sprintf(format, v...))
}

func (d *ProxyDialer) dialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	if d.Dialer == nil {
		if dl, ok := ctx.Deadline(); ok {
			return net.DialTimeout(network, addr, dl.Sub(time.Now()))
		}
		return net.Dial(network, addr)
	}
	return d.Dialer(ctx, network, addr)
}

type proxyConnection struct {
	dialer *ProxyDialer
	proxy  *url.URL
	conn   net.Conn
	addr   string
}

func (c *proxyConnection) Proxy() *url.URL {
	return c.proxy
}

func (c *proxyConnection) Connect(auth string) (*http.Response, error) {
	connectReq := &http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Opaque: c.addr},
		Host:   c.addr,
		Header: make(http.Header),
	}
	if auth != "" {
		connectReq.Header.Set("Proxy-Authorization", auth)
	}
	if err := connectReq.Write(c.conn); err != nil {
		return nil, err
	}
	br := bufio.NewReader(c.conn)
	resp, err := http.ReadResponse(br, connectReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		return resp, nil
	case http.StatusProxyAuthRequired:
		_, _ = io.Copy(ioutil.Discard, resp.Body)
		return resp, errProxyAuth
	}
	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, err
	}
	return nil, fmt.Errorf("proxy return error '%s': %s", resp.Status, string(out))
}

func (d *ProxyDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	c, err := d.dialContext(ctx, network, host(d.Host))
	if err != nil {
		return nil, err
	}
	pc := &proxyConnection{
		dialer: d,
		proxy:  d.Host,
		conn:   c,
		addr:   addr,
	}
	resp, err := pc.Connect("")
	// connection success
	if err == nil {
		return c, nil
	}
	// no response from proxy
	if resp == nil {
		c.Close()
		return nil, err
	}
	// unexpected status from proxy
	if resp.StatusCode != http.StatusProxyAuthRequired {
		d.logf("Proxy Auth Status: %v", resp.Status)
		c.Close()
		return nil, err
	}
	// try proxy auth
	err = d.Auth.Authorize(resp, pc)
	// proxy auth failed
	if err != nil {
		d.logf("Proxy Auth Failed: %v", err)
		c.Close()
		return nil, err
	}
	// proxy auth success
	return c, nil
}
