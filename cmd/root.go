package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version  string = "0.0.1"
	Revision string = "1111aaaa"
)

var (
	cfgSectionName     string
	storeInProfile     bool
	killHangingProcess bool
	role               string
	roleChain          []string
	verbose            bool
	rootCmd            = &cobra.Command{
		Use:   "aws-cli-auth",
		Short: "CLI tool for retrieving AWS temporary credentials",
		Long: `CLI tool for retrieving AWS temporary credentials using SAML providers, or specified method of retrieval - i.e. force AWS_WEB_IDENTITY.
Useful in situations like CI jobs or containers where multiple env vars might be present.
Stores them under the $HOME/.aws/credentials file under a specified path or returns the crednetial_process payload for use in config`,
		Version: fmt.Sprintf("%s-%s", Version, Revision),
	}
)

func Execute(ctx context.Context) {
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Errorf("cli error: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&role, "role", "r", "", "Set the role you want to assume when SAML or OIDC process completes")
	rootCmd.PersistentFlags().StringSliceVarP(&roleChain, "role-chain", "", []string{}, "If specified it will assume the roles from the base credentials, in order they are specified in")
	rootCmd.PersistentFlags().StringVarP(&cfgSectionName, "cfg-section", "", "", "config section name in the yaml config file")
	rootCmd.PersistentFlags().BoolVarP(&storeInProfile, "store-profile", "s", false, "By default the credentials are returned to stdout to be used by the credential_process. Set this flag to instead store the credentials under a named profile section")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
}
