package proxy

import (
	"context"
	"fmt"
	"github.com/justenwalker/squiggly/auth"
	"net"
	"net/http"
	"net/url"

	"github.com/justenwalker/squiggly/logging"
	"gopkg.in/elazarl/goproxy.v1"
)

// Server is a proxy server
type Server struct {
	logger    logging.Logger
	logWriter *logging.LogWriter
	proxyFunc func(req *http.Request) (*url.URL, error)
	proxyAuth *auth.Auth
	server    *goproxy.ProxyHttpServer
}

func (s *Server) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	s.server.ServeHTTP(resp, req)
}

// Close the server down and flush logs
func (s *Server) Close() error {
	if s.logWriter != nil {
		return s.logWriter.Flush()
	}
	return nil
}

func (s *Server) logf(msg string, v ...interface{}) {
	if s.logger != nil {
		s.logger.Log(fmt.Sprintf(msg, v...))
	}
}

func (s *Server) log(v ...interface{}) {
	if s.logger != nil {
		s.logger.Log(fmt.Sprint(v...))
	}
}

func (s *Server) onRequest(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	s.logf("onRequest: %s", req.URL)
	return req, nil
}

func (s *Server) onResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	if resp != nil {
		s.logf("onResponse: %s", resp.Status)
	}
	if ctx.Error != nil {
		s.logf("onResponse: ERROR:%#v", ctx.Error)
	}
	return resp
}

func (s *Server) proxyHost(host string) (string, error) {
	if s.proxyFunc == nil {
		return host, nil
	}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/", host), nil)
	if err != nil {
		return "", fmt.Errorf("host '%s' parse error : %v", host, err)
	}

	proxy, err := s.proxy(req)
	if err != nil {
		return "", err
	}
	if proxy != nil && proxy.Host != "" {
		return proxy.Host, nil
	}
	return host, nil
}

func (s *Server) getProxyHost(host string) (*url.URL, error) {
	if s.proxyFunc == nil {
		return nil, nil
	}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/", host), nil)
	if err != nil {
		return nil, fmt.Errorf("host '%s' parse error : %v", host, err)
	}
	return s.proxy(req)
}

func (s *Server) proxy(req *http.Request) (*url.URL, error) {
	if s.proxyFunc == nil {
		return nil, nil
	}
	u, err := s.proxyFunc(req)
	if err == nil {
		s.logf("PROXY SELECT: %v", u)
	}
	return u, err
}

// New creates a new proxy Server with the given options configured
func New(opts ...Option) *Server {
	srv := &Server{
		server: goproxy.NewProxyHttpServer(),
	}
	srv.server.Tr = &http.Transport{
		Dial: srv.dialer,
	}
	srv.server.ConnectDial = srv.dialer
	for _, opt := range opts {
		opt(srv)
	}
	// srv.server.OnRequest().HandleConnectFunc(srv.onConnect)
	srv.server.OnRequest().DoFunc(srv.onRequest)
	srv.server.OnResponse().DoFunc(srv.onResponse)
	return srv
}

func (s *Server) dialer(network, addr string) (net.Conn, error) {
	purl, err := s.getProxyHost(addr)
	if err != nil {
		s.logf("dialer: getProxyHost ERROR: '%s'", err)
		return nil, err
	}
	// Prevent upstream proxy from being re-directed
	if purl == nil || purl.Host == addr {
		s.logf("dialer: DIRECT -> ADDR '%s'", addr)
		return net.Dial(network, addr)
	}
	s.logf("dialer: PROXY '%s' -> ADDR '%s'", purl.Host, addr)
	dialer := &ProxyDialer{
		Logger: s.logger,
		Host:   purl,
		Auth:   s.proxyAuth,
	}
	return dialer.DialContext(context.Background(), network, addr)
}
