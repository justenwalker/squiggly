package proxy_test

import (
	"github.com/justenwalker/squiggly/proxy"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

type testLogger struct {
	t *testing.T
}

func (l testLogger) Log(msg string) {
	l.t.Log(msg)
}

func TestProxyDialer(t *testing.T) {
	proxyUser := os.Getenv("PROXY_USER")
	proxyPass := os.Getenv("PROXY_PASS")
	proxyHost := os.Getenv("PROXY_HOST")
	var auth *proxy.BasicAuth
	if proxyHost == "" {
		t.Skip("PROXY_HOST not defined")
	}
	if proxyUser != "" {
		auth = &proxy.BasicAuth{
			Username: proxyUser,
			Password: proxyPass,
		}
	}
	dialer := &proxy.ProxyDialer{
		Logger: testLogger{t},
		Auth: auth,
		Host: proxyHost,
	}
	client := http.Client{
		Transport: &http.Transport{
			DialContext: dialer.DialContext,
		},
	}
	resp, err := client.Get("https://example.com")
	if err != nil {
		t.Fatal("error getting example.com", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("error reading body", err)
	}
	t.Log(string(body))
}
