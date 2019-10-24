package cmd

import (
	"io"
	"log"
	"os"

	"github.com/justenwalker/squiggly/auth"
	"github.com/spf13/cobra"
)

var (
	krb5confRealm string
)

// authCmd represents the auth command
var krb5confCmd = &cobra.Command{
	Use:   "krb5conf",
	Short: "Generate a krb5.conf",
	Run: func(cmd *cobra.Command, args []string) {
		r, err := auth.DiscoverKrb5Config(krb5confRealm)
		if err != nil {
			log.Fatalf("error discovering kerberos realm '%s': %v", krb5confRealm, err)
		}
		defer r.Close()
		io.Copy(os.Stdout, r)
	},
}

func init() {
	RootCmd.AddCommand(krb5confCmd)
	krb5confCmd.Flags().StringVarP(&krb5confRealm, "realm", "r", realm, "kerberos realm")
}
