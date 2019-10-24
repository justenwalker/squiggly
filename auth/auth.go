package auth

import (
	"fmt"
	"net/http"

	"github.com/justenwalker/squiggly/logging"
)

type Auth struct {
	Logger  logging.Logger
	cs      CredentialStore
	krbconf string
	spnego  *SPNEGO
}

func NewAuth(cs CredentialStore, spnego *SPNEGO) *Auth {
	return &Auth{
		cs:     cs,
		spnego: spnego,
	}
}

func (a Auth) log(msg string) {
	if a.Logger == nil {
		return
	}
	a.Logger.Log(msg)
}

func (a Auth) Authorize(resp *http.Response, pc ProxyConnection) error {
	ah := GetHeader(resp)
	switch {
	case ah.IsBasic():
		a.log("basic proxy-auth")
		return BasicAuth{
			CredentialStore: a.cs,
		}.Authorize(resp, pc)
	case ah.IsNTLM():
		a.log("NTLM proxy-auth")
		cred, err := a.cs.Credentials(getHost(pc.Proxy()))
		if err != nil {
			return err
		}
		return NTLM{Credentials: cred}.Authorize(resp, pc)
	case ah.IsNegotiate():
		if a.spnego != nil {
			a.log("Negotiate proxy-auth")
			return a.spnego.Authorize(resp, pc)
		}
	}
	return fmt.Errorf("unsupported proxy auth type: %v", ah)
}
