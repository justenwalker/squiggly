package config

import (
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
)

// DynamicConfig allows dynamic configuration of the proxy server settings during runtime
type DynamicConfig struct {
	mu sync.RWMutex
	pc atomic.Value
}

// BasicAuth represents the username and password
type BasicAuth struct {
	Username string
	Password string
}

// New creates a new DynamicConfig
func New() *DynamicConfig {
	cfg := &DynamicConfig{}
	cfg.pc.Store(&proxyConfig{
		Blacklist: make(map[string]struct{}),
	})
	return cfg
}

// SetBlacklist sets the list of hosts which should never be proxied
func (c *DynamicConfig) SetBlacklist(blacklist []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	pc := c.cloneProxyConfig()
	bm := make(map[string]struct{}, len(blacklist))
	for _, b := range blacklist {
		bm[strings.TrimSpace(strings.ToLower(b))] = struct{}{}
	}
	pc.Blacklist = bm
	c.pc.Store(pc)
}

// SetProxyEnabled sets whether the proxy is enabled or all connections should be direct
func (c *DynamicConfig) SetProxyEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	pc := c.cloneProxyConfig()
	pc.Disabled = !enabled
	c.pc.Store(pc)
}

// SetProxy sets the proxy function
func (c *DynamicConfig) SetProxy(proxy func(req *http.Request) (*url.URL, error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	pc := c.cloneProxyConfig()
	pc.Func = proxy
	c.pc.Store(pc)
}

// Proxy uses the config to set the correct proxy url and authorization based on the request
func (c *DynamicConfig) Proxy(req *http.Request) (*url.URL, error) {
	if pc := c.proxyConfig(); pc != nil {
		return pc.Proxy(req)
	}
	return nil, nil
}

func (c *DynamicConfig) cloneProxyConfig() *proxyConfig {
	pc := c.proxyConfig()
	if pc == nil {
		return &proxyConfig{}
	}
	return &proxyConfig{
		Blacklist: pc.Blacklist,
		Func:      pc.Func,
		Disabled:  pc.Disabled,
	}
}

func (c *DynamicConfig) proxyConfig() *proxyConfig {
	v := c.pc.Load()
	if v == nil {
		return nil
	}
	return v.(*proxyConfig)
}

type proxyConfig struct {
	Func      func(req *http.Request) (*url.URL, error)
	Disabled  bool
	Blacklist map[string]struct{}
}

func (p *proxyConfig) Proxy(req *http.Request) (*url.URL, error) {
	if p == nil || p.Func == nil || p.Disabled {
		return nil, nil
	}
	host := strings.TrimSpace(strings.ToLower(req.URL.Host))
	if _, ok := p.Blacklist[host]; ok {
		return nil, nil
	}
	host = strings.TrimSpace(strings.ToLower(req.URL.Hostname()))
	if _, ok := p.Blacklist[host]; ok {
		return nil, nil
	}
	return p.Func(req)
}
