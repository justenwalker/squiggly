package cmd

import (
	"fmt"
	"github.com/justenwalker/squiggly/auth"
	"log"
	"os"

	"github.com/howeyc/gopass"
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

func proxyAuth(service, username string) (auth.Credentials, error) {
	if service == "" {
		return auth.Credentials{}, fmt.Errorf("service name missing")
	}
	if username == "" {
		return auth.Credentials{}, fmt.Errorf("user name missing")
	}
	password, err := keyring.Get(service, username)
	if err != nil {
		return auth.Credentials{}, err
	}
	return auth.Credentials{Username: username, Password: password}, nil
}

func promptPassword(username string) ([]byte, error) {
	return gopass.GetPasswdPrompt(fmt.Sprintf("[%s] Password: ", username), true, os.Stdin, os.Stdout)
}

func authenticate(service, username string, password []byte) error {
	return keyring.Set(service, username, string(password))
}
