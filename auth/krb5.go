package auth

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"strings"
	"text/template"
)

const krb5template = `
# Other applications require this directory to perform krb5 configuration.
includedir /etc/krb5.conf.d/
{{ $realm := .Name | upper }}
[libdefaults]
 default_realm = {{ $realm }}

[domain_realm]
{{- range .Domains }}
  {{ . }} = {{ $realm }}
{{- end }}
  .{{ .Name | lower }} = {{ $realm }}
  {{ .Name | lower }} = {{ $realm }}


[realms]
{{ $realm }} = {
{{- range .Domains }}
	kdc = {{ . }}:88
	master_kdc = {{ . }}:88
	kpasswd = {{ . }}:464
	kpasswd_server = {{ . }}:464
{{- end }}
}
`

type realm struct {
	Name    string
	Domains []string
}

func DiscoverKrb5Config(name string) (io.ReadCloser, error) {
	if name == "" {
		return nil, fmt.Errorf("realm cannot be empty")
	}
	realm := discoverRealm(name)
	buf := &bytes.Buffer{}
	t, err := template.New("krb5.conf").Funcs(map[string]interface{}{
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
	}).Parse(krb5template)
	if err != nil {
		return nil, err
	}
	if err := t.Execute(buf, &realm); err != nil {
		return nil, err
	}
	return ioutil.NopCloser(buf), nil
}

func discoverRealm(name string) realm {
	var r realm
	r.Name = strings.ToUpper(name)
	addrs, err := net.LookupHost(name)
	if err != nil {
		return r
	}
	for _, addr := range addrs {
		if names, err := net.LookupAddr(addr); err == nil {
			for _, name := range names {
				r.Domains = append(r.Domains, name)
			}
		}
	}
	return r
}
