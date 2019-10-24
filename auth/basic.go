package auth

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type ProxyConnection interface {
	Proxy() *url.URL
	Connect(auth string) (*http.Response, error)
}

func mustConnect(auth string, pc ProxyConnection) error {
	if _, err := pc.Connect(auth); err != nil {
		return err
	}
	return nil
}

type BasicAuth struct {
	CredentialStore
}

func (a BasicAuth) Authorize(resp *http.Response, pc ProxyConnection) error {
	authHeader := GetHeader(resp)
	if !authHeader.IsBasic() {
		return fmt.Errorf("wrong auth header")
	}
	opts := authHeader.GetOptions()
	realm := opts["realm"]
	creds, err := a.CredentialStore.Credentials(realm)
	if err != nil {
		return err
	}
	sb := &strings.Builder{}
	sb.WriteString("Basic ")
	n := len(creds.Username) + len(creds.Password) + 1
	userpass := bytes.NewBuffer(make([]byte, 0, n))
	userpass.WriteString(creds.Username)
	userpass.WriteByte(':')
	userpass.WriteString(creds.Password)
	sb.WriteString(base64.StdEncoding.EncodeToString(userpass.Bytes()))
	return mustConnect(sb.String(), pc)
}
