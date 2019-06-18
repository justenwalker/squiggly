package proxy

import (
	"encoding/base64"
	"log"
	"net/http"
	"net/url"
)

// Option configures the proxy server
type Option func(srv *Server)

// BasicAuth represents the username and password
type BasicAuth struct {
	Username string
	Password string
}

// Encoded password
func (b *BasicAuth) Encoded() string {
	if b == nil {
		return ""
	}
	userpass := []byte(url.UserPassword(b.Username, b.Password).String())
	return base64.StdEncoding.EncodeToString(userpass)
}

// Proxy is an option that controls which upstream proxy is used for each request
// The proxy function may return a nil URL which indicates a direct connection should be made.
func Proxy(proxy func(req *http.Request) (*url.URL, error)) Option {
	return func(s *Server) {
		s.proxyFunc = proxy
	}
}

// ProxyAuth is an option that sets the proxy basic authorization credentials
func ProxyAuth(auth *BasicAuth) Option {
	return func(s *Server) {
		s.proxyAuth = auth
	}
}

// Log sets the logger on the server for debug purposes
func Log(logger Logger) Option {
	return func(s *Server) {
		s.logger = logger
		s.server.Verbose = (logger != nil)
		if logger != nil {
			s.server.Logger = log.New(newLogWriter(logger), "goproxy: ", 0)
		}
	}
}
