package auth

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync/atomic"

	"gopkg.in/jcmturner/gokrb5.v7/client"
	"gopkg.in/jcmturner/gokrb5.v7/config"
	"gopkg.in/jcmturner/gokrb5.v7/krberror"
	"gopkg.in/jcmturner/gokrb5.v7/spnego"
)

// SPNEGO Authorizer
type SPNEGO struct {
	login int64
	creds Credentials
	cl *client.Client
}

// NewSPNEGO creates a new SPNEGO authenticator
func NewSPNEGO(c Credentials, krbconf string) (*SPNEGO, error) {
	var conf io.ReadCloser
	var err error
	if krbconf == "" {
		log.Println("krbconf discovered")
		conf, err = DiscoverKrb5Config(c.Realm)
	} else {
		log.Println("krbconf", krbconf)
		conf, err = os.Open(krbconf)
	}
	if err != nil {
		return nil, err
	}
	defer conf.Close()
	cfg, err := config.NewConfigFromReader(conf)
	if err != nil {
		return nil, err
	}
	cl := client.NewClientWithPassword(c.Username, c.Realm, c.Password, cfg)
	return &SPNEGO{cl: cl,creds: c}, nil
}

func (c *SPNEGO) Authorize(resp *http.Response, pc ProxyConnection) error {
	header, err := c.Header(pc.Proxy())
	if err != nil {
		return err
	}
	return mustConnect(header, pc)
}

func (c *SPNEGO) acquireCreds() error {
	if c.login == 1 {
		return nil
	}
	if !atomic.CompareAndSwapInt64(&c.login,0,1) {
		return nil
	}
	if err := c.cl.Login(); err != nil {
		atomic.StoreInt64(&c.login,0)
		return fmt.Errorf("could not acquire client credential (%s@%s): %v", c.creds.Username, c.creds.Realm, err)
	}
	return nil
}

// Header gets the Negotiate header authorizing the request
func (c *SPNEGO) Header(proxy *url.URL) (string, error) {
	if err := c.acquireCreds(); err != nil {
		return "",err
	}
	spn := getSPN(proxy)
	s := spnego.SPNEGOClient(c.cl, spn)
	st, err := s.InitSecContext()
	if err != nil {
		return "", fmt.Errorf("could not initialize context (SPN: %s): %v", spn, err)
	}
	nb, err := st.Marshal()
	if err != nil {
		return "", krberror.Errorf(err, krberror.EncodingError, "could not marshal SPNEGO")
	}
	return "Negotiate " + base64.StdEncoding.EncodeToString(nb), nil
}

func getHost(url *url.URL) string {
	return strings.TrimSuffix(strings.SplitN(url.Host, ":", 2)[0], ".")
}

func getSPN(url *url.URL) string {
	h := getHost(url)
	name, err := net.LookupCNAME(h)
	if err == nil {
		// Underlyng canonical name should be used for SPN
		h = strings.TrimSuffix(name, ".")
	}
	return "HTTP/" + getHost(url)
}
