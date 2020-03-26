package auth

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/Azure/go-ntlmssp"
)

type NTLM struct {
	Credentials
}

// Authorize a proxy connection using NTLM
func (n NTLM) Authorize(resp *http.Response, pc ProxyConnection) error {
	header := &strings.Builder{}
	negotiateMessage, err := ntlmssp.NewNegotiateMessage(n.Realm, "")
	if err != nil {
		return err
	}
	authHeader := GetHeader(resp)
	switch {
	case authHeader.IsNTLM():
		header.WriteString("NTLM ")
	case authHeader.IsNegotiate():
		header.WriteString("Negotiate ")
	}
	header.WriteString(base64.StdEncoding.EncodeToString(negotiateMessage))

	resp, err = pc.Connect(header.String())
	if err == nil || resp == nil {
		return nil
	}
	if resp.StatusCode != http.StatusProxyAuthRequired {
		return err
	}
	authHeader = GetHeader(resp)
	challengeMessage, err := authHeader.GetData()
	if err != nil {
		return err
	}
	if len(challengeMessage) == 0 {
		return fmt.Errorf("NTLM challenge data is blank")
	}
	header = &strings.Builder{}
	authenticateMessage, err := ntlmssp.ProcessChallenge(challengeMessage, n.Username, n.Password)
	if err != nil {
		return err
	}
	switch {
	case authHeader.IsNTLM():
		header.WriteString("NTLM ")
	case authHeader.IsNegotiate():
		header.WriteString("Negotiate ")
	}
	header.WriteString(base64.StdEncoding.EncodeToString(authenticateMessage))
	return mustConnect(header.String(), pc)
}
