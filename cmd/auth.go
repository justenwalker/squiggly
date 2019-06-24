package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/howeyc/gopass"
	"github.com/justenwalker/squiggly/proxy"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

var (
	authService  string
	authUsername string
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Sets the proxy authentication credentials",
	Run: func(cmd *cobra.Command, args []string) {
		password, err := promptPassword(authUsername)
		if err != nil {
			log.Fatal(err)
		}
		if err := authenticate(authService, authUsername, password); err != nil {
			log.Fatal(err)
		}
	},
}

func defaultUser() string {
	return os.Getenv("USER")
}

func init() {
	RootCmd.AddCommand(authCmd)
	authCmd.Flags().StringVarP(&authService, "service", "s", defaultService, "service name, used to distinguish between auth configurations")
	authCmd.Flags().StringVarP(&authUsername, "user", "u", defaultUser(), "user name, used to log into proxy servers")
}

func proxyAuth(service, username string) (*proxy.BasicAuth, error) {
	if service == "" {
		return nil, fmt.Errorf("service name missing")
	}
	if username == "" {
		return nil, fmt.Errorf("user name missing")
	}
	password, err := keyring.Get(service, username)
	if err != nil {
		return nil, err
	}
	return &proxy.BasicAuth{Username: username, Password: password}, nil
}

func promptPassword(username string) ([]byte, error) {
	return gopass.GetPasswdPrompt(fmt.Sprintf("[%s] Password: ", username), true, os.Stdin, os.Stdout)
}

func authenticate(service, username string, password []byte) error {
	return keyring.Set(service, username, string(password))
}
