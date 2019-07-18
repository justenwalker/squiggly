package cmd

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/justenwalker/squiggly/config"
	"github.com/justenwalker/squiggly/logging"
	"github.com/justenwalker/squiggly/pac"
	"github.com/justenwalker/squiggly/proxy"
	"github.com/spf13/cobra"
)

const (
	defaultService = "squiggly"
	defaultAddress = "localhost:8800"
)

var (
	service  string
	username string
	password string
	pacURL   string
	address  string
	interval time.Duration
	verbose  bool
)

// proxyCmd represents the proxy command
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Start the proxy server",
	Long:  `Starts the proxy server.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runProxy(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(proxyCmd)
	proxyCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
	proxyCmd.Flags().StringVarP(&pacURL, "pac", "p", "", "url to the proxy auto config (PAC) file")
	proxyCmd.Flags().StringVarP(&address, "address", "a", "localhost:8800", "listen address for the proxy server")
	proxyCmd.Flags().DurationVarP(&interval, "interval", "i", 10*time.Second, "time between checks of the PAC file to turn off the proxy")
	proxyCmd.Flags().StringVarP(&service, "service", "s", defaultService, "service name, used to distinguish between auth configurations")
	proxyCmd.Flags().StringVarP(&username, "user", "u", "", "user name, used to log into proxy servers. Omit to use an unauthenticated proxy.")
}

func runProxy() error {
	cfg := &config.DynamicConfig{}
	proxyPAC := &pac.PAC{URL: pacURL}
	cfg.SetProxy(proxyPAC.Proxy)
	if _, err := proxyPAC.Refresh(); err != nil {
		log.Println("Unable to parse PAC:", err)
	}
	options := []proxy.Option{
		proxy.Proxy(cfg.Proxy),
	}
	if username != "" {
		auth, err := proxyAuth(service, username)
		if err != nil {
			return err
		}
		options = append(options, proxy.ProxyAuth(auth))
	}
	if verbose {
		logger := &logging.StandardLogger{}
		options = append(options, proxy.Log(logger))
		cfg.Logger = logger
	}
	prx := proxy.New(options...)
	srv := &http.Server{
		Addr:    address,
		Handler: prx,
	}
	defer prx.Close()
	// Listen for Interrupt
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

	// PAC Refresher & Proxy Enabler
	quitCh := make(chan struct{})
	go func() {
		var proxyDisabled bool
		for {
			select {
			case <-time.After(interval):
				if _, err := proxyPAC.Refresh(); err != nil { // PAC unreachable
					if !proxyDisabled {
						proxyDisabled = true
						cfg.SetProxyEnabled(false)
					}
				} else { // PAC is reachable
					if proxyDisabled {
						proxyDisabled = false
						cfg.SetProxyEnabled(true)
					}
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
	return srv.ListenAndServe()
}
