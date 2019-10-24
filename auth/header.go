package auth

import (
	"encoding/base64"
	"net/http"
	"net/textproto"
	"strings"
)

type Header string

func GetHeader(resp *http.Response) Header {
	vs, ok := resp.Header[textproto.CanonicalMIMEHeaderKey("Proxy-Authenticate")]
	if !ok || len(vs) == 0 {
		return Header("")
	}
	return Header(vs[0])
}

func (h Header) IsBasic() bool {
	return strings.HasPrefix(string(h), "Basic ")
}

func (h Header) IsNegotiate() bool {
	return strings.HasPrefix(string(h), "Negotiate")
}

func (h Header) IsNTLM() bool {
	return strings.HasPrefix(string(h), "NTLM")
}

func (h Header) GetData() ([]byte, error) {
	p := strings.Split(string(h), " ")
	if len(p) < 2 {
		return nil, nil
	}
	return base64.StdEncoding.DecodeString(string(p[1]))
}

func (h Header) GetOptions() map[string]string {
	p := strings.Split(string(h), " ")
	if len(p) < 2 {
		return make(map[string]string)
	}
	opts := make(map[string]string)
	for _, part := range strings.Split(p[1], ", ") {
		vals := strings.SplitN(part, "=", 2)
		key := strings.ToLower(strings.TrimSpace(vals[0]))
		val := strings.Trim(vals[1], "\",")
		opts[key] = val
	}
	return opts
}
