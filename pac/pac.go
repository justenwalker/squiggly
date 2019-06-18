package pac

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jackwakefield/gopac"
)

const lastModifiedFormat = "2006-01-02 15:04:05 GMT"

var noProxyTransport http.RoundTripper = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}).DialContext,
	MaxIdleConns:          0,
	IdleConnTimeout:       0 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

var noProxyClient = &http.Client{
	Transport: noProxyTransport,
}

type PAC struct {
	URL          string
	parsed       *gopac.Parser
	etag         string
	lastModified time.Time
	mu           sync.Mutex
}

// Proxy returns the proxy URL for a request, or nil if the request should not be proxied
type Proxy interface {
	Proxy(req *http.Request) (*url.URL, error)
}

type directProxy int

// Direct is an anti-proxy; it always returns a direct connection
const Direct = directProxy(0)

func (p directProxy) Proxy(req *http.Request) (*url.URL, error) {
	return nil, nil
}

type proxyURL struct {
	URL *url.URL
}

func (p proxyURL) Proxy(req *http.Request) (*url.URL, error) {
	return p.URL, nil
}

// ProxyForRequest uses the PAC to discover zero or more proxies that match the request
func (r *PAC) ProxyForRequest(url, host string) ([]Proxy, error) {
	// gopac.Parser.FindProxy is not concurrency safe
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.parsed == nil {
		return nil, nil
	}
	result, err := r.parsed.FindProxy(url, host)
	if err != nil {
		return nil, err
	}
	return parsePACResult(result)
}

// Proxy function to be used in a transport
func (r *PAC) Proxy(req *http.Request) (*url.URL, error) {
	proxies, err := r.ProxyForRequest(req.URL.String(), req.URL.Hostname())
	if err != nil {
		return nil, err
	}
	if len(proxies) == 0 {
		return nil, nil
	}
	return proxies[0].Proxy(req)
}

// Refresh fetches the PAC file
// The boolean returned indicates if an update occurred
func (r *PAC) Refresh() (bool, error) {
	u, err := url.Parse(r.URL)
	if err != nil {
		return false, err
	}
	switch u.Scheme {
	case "file":
		path := filepath.FromSlash(u.Path)
		stat, err := os.Stat(path)
		if err != nil {
			return false, err
		}
		modified := stat.ModTime()
		if modified.After(r.lastModified) {
			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				return false, err
			}
			r.mu.Lock()
			defer r.mu.Unlock()
			parser := &gopac.Parser{}
			if err := parser.ParseBytes(bytes); err != nil {
				return false, err
			}
			r.parsed = parser
			r.lastModified = modified
			return true, nil
		}
	case "http", "https":
		req, err := http.NewRequest(http.MethodGet, u.String(), nil)
		if err != nil {
			return false, err
		}
		if r.etag != "" {
			req.Header.Set("If-None-Match", r.etag)
		} else if !r.lastModified.IsZero() {
			req.Header.Set("If-Modified-Since", r.lastModified.Format(lastModifiedFormat))
		}
		resp, err := noProxyClient.Do(req)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}
		switch resp.StatusCode {
		case http.StatusNotModified:
			return false, nil
		case http.StatusOK:
			r.mu.Lock()
			defer r.mu.Unlock()
			parser := &gopac.Parser{}
			if err := parser.ParseBytes(bytes); err != nil {
				return false, err
			}
			r.parsed = parser
			r.etag = resp.Header.Get("ETag")
			if lm := resp.Header.Get("Last-Modified"); lm != "" {
				if date, err := time.Parse(lastModifiedFormat, lm); err == nil {
					r.lastModified = date
				}
			}
			return true, nil
		default:
			return false, fmt.Errorf("GET '%v': %s\n%s", u, resp.Status, string(bytes))
		}
	}
	return false, nil
}

func parsePACResult(result string) ([]Proxy, error) {
	var proxies []Proxy
	for _, p := range strings.Split(result, ";") {
		p = strings.ToLower(strings.TrimSpace(p))
		if strings.HasPrefix(p, "proxy") {
			hostport := strings.TrimSpace(strings.TrimPrefix(p, "proxy"))
			purl, err := url.Parse(fmt.Sprintf("http://%s/", hostport))
			if err != nil {
				return nil, err
			}
			proxies = append(proxies, proxyURL{purl})
		}
		if strings.EqualFold(p, "direct") {
			proxies = append(proxies, Direct)
		}
	}
	return proxies, nil
}
