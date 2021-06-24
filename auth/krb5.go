package auth

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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
{{- range .AdminServers }}
	admin_server = {{ . }}
{{- end }}
{{- range .KDCMasters }}
	master_kdc = {{ . }}
{{- end }}
{{- range .KDCS }}
	kdc = {{ . }}
{{- end }}
{{- range .KDCS }}
	kdc = {{ . }}
{{- end }}
{{- range .KPasswdServer }}
	kpasswd_server = {{ . }}
{{- end }}
}
`

type realm struct {
	Name          string
	Domains       []string
	AdminServers  []string
	KDCMasters    []string
	KDCS          []string
	KPasswdServer []string
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
	var (
		r     realm
		addrs []*net.SRV
		err   error
	)
	domains := make(map[string]struct{})
	r.Name = strings.ToUpper(name)
	if addrs, err = lookupSRV("kerberos-master", "udp", name); err != nil {
		log.Println("error looking up kerberos-masters:", err)
	}
	for _, addr := range addrs {
		r.KDCMasters = append(r.KDCMasters, fmt.Sprintf("%s:%d", addr.Target, addr.Port))
		domains[strings.ToLower(addr.Target)] = struct{}{}
	}
	if addrs, err = lookupSRV("kerberos-adm", "tcp", name); err != nil {
		log.Println("error looking up kerberos-adm:", err)
	}
	for _, addr := range addrs {
		r.AdminServers = append(r.AdminServers, fmt.Sprintf("%s:%d", addr.Target, addr.Port))
		domains[strings.ToLower(addr.Target)] = struct{}{}
	}
	if addrs, err = lookupSRV("kerberos", "udp", name); err != nil {
		log.Println("error looking up KDCs:", err)
	}
	for _, addr := range addrs {
		r.KDCS = append(r.KDCS, fmt.Sprintf("%s:%d", addr.Target, addr.Port))
		domains[strings.ToLower(addr.Target)] = struct{}{}
	}
	if addrs, err = lookupSRV("kpasswd", "udp", name); err != nil {
		log.Println("error looking up kpasswd servers:", err)
	}
	for _, addr := range addrs {
		r.KPasswdServer = append(r.KPasswdServer, fmt.Sprintf("%s:%d", addr.Target, addr.Port))
		domains[strings.ToLower(addr.Target)] = struct{}{}
	}
	for domain := range domains {
		r.Domains = append(r.Domains, domain)
	}
	return r
}

func lookupSRV(service, proto, name string) ([]*net.SRV, error) {
	_, addrs, err := net.LookupSRV(service, proto, name)
	var dnserr *net.DNSError
	if errors.As(err, &dnserr) {
		if dnserr.IsNotFound {
			return nil, nil
		}
		return nil, err
	}
	return addrs, nil
}
