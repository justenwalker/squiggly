package main

import (
	"fmt"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/justenwalker/squiggly/proxy"
	"github.com/justenwalker/squiggly/pac"
	"github.com/justenwalker/squiggly/config"
	"github.com/zalando/go-keyring"
	"github.com/howeyc/gopass"
)

var (
	help     bool
	service  string
	username string
	password string
	pacURL   string
	address  string
	interval  string
	verbose   bool
)

const defaultServiceName = "sqiggly"
const defaultAddress = "localhost:8800"
const defaultControl = "localhost:8801"

type logger struct {
}

func (l logger) Log(msg string) {
	log.Println(msg)
}

func main() {
	flag.StringVar(&service, "service", defaultServiceName, "The service name for credential storage")
	flag.StringVar(&username, "username", "", "The username for proxy authorization")
	flag.StringVar(&password, "password", "", "The password for proxy authorization")
	flag.StringVar(&pacURL, "pac", "", "The url of the PAC (proxy auto config)")
	flag.StringVar(&address, "addr", defaultAddress, "The address on which this local proxy server will listen.")
	flag.StringVar(&interval, "interval", "10s", "The interval to test if the PAC is available.")
	flag.BoolVar(&help, "help", false, "Print help text")
	flag.BoolVar(&verbose, "verbose", false, "Print lots of extra information")
	flag.Parse()
	if help {
		flag.PrintDefaults()
		os.Exit(0)
	}
	cfg := &config.DynamicConfig{}
	proxyPAC := &pac.PAC{URL: pacURL}
	cfg.SetProxy(proxyPAC.Proxy)
	if _,err := proxyPAC.Refresh(); err != nil {
		log.Println("Unable to parse PAC:",err)
	}
	dur,err := time.ParseDuration(interval)
	if err != nil {
		log.Fatal(err)
	}
	var auth *proxy.BasicAuth
	if username != "" {
		if password != "" {
			err := keyring.Set(service, username, password)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			secret, err := keyring.Get(service, username)
			if err != nil {
				if err == keyring.ErrNotFound {
					pb,err := gopass.GetPasswdPrompt(fmt.Sprintf("[%s] Password: ",username), true, os.Stdin, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
					password = string(pb)
					if err := keyring.Set(service, username, password); err != nil {
						log.Fatal(err)
					}
				} else {
					log.Fatal(err)
				}
			}
			password = secret
		}
		auth = &proxy.BasicAuth{Username: username, Password: password}
	}
	options := []proxy.Option{
		proxy.Proxy(cfg.Proxy),
		proxy.ProxyAuth(auth),
	}
	if verbose {
		options = append(options, proxy.Log(logger{}))
	}
	srv := &http.Server{
		Addr: address,
		Handler: proxy.New(options...),
	}
	// Listen for Interrupt
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

	// PAC Refresher & Proxy Enabler
	quitCh := make(chan struct{})
	go func() {
		var proxyDisabled bool
		for {
			select {
			case <- time.After(dur):
				if _,err := proxyPAC.Refresh(); err != nil && !proxyDisabled {
					proxyDisabled = true
					cfg.SetProxyEnabled(false)
				} else if proxyDisabled {
					proxyDisabled = false
					cfg.SetProxyEnabled(true)
				}
			case <-quitCh:
				return
			}
		}
	}()

	// Shut Down on Signal
	go func() {
		<-sig
		close(quitCh)
		srv.Close()
	}()

	// Run Proxy
	log.Println("LISTEN", address)
	log.Fatal(srv.ListenAndServe())
}
