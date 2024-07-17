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
	duration           int
	RootCmd            = &cobra.Command{
		Use:   "aws-cli-auth",
		Short: "CLI tool for retrieving AWS temporary credentials",
		Long: `CLI tool for retrieving AWS temporary credentials using SAML providers, or specified method of retrieval - i.e. force AWS_WEB_IDENTITY.
Useful in situations like CI jobs or containers where multiple env vars might be present.
Stores them under the $HOME/.aws/credentials file under a specified path or returns the crednetial_process payload for use in config`,
		Version: fmt.Sprintf("%s-%s", Version, Revision),
	}
)

func Execute(ctx context.Context) {
	if err := RootCmd.ExecuteContext(ctx); err != nil {
		fmt.Errorf("cli error: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func init() {
	RootCmd.PersistentFlags().StringSliceVarP(&roleChain, "role-chain", "", []string{}, "If specified it will assume the roles from the base credentials, in order they are specified in")
	RootCmd.PersistentFlags().BoolVarP(&storeInProfile, "store-profile", "s", false, `By default the credentials are returned to stdout to be used by the credential_process. 
	Set this flag to instead store the credentials under a named profile section. You can then reference that profile name via the CLI or for use in an SDK`)
	RootCmd.PersistentFlags().StringVarP(&cfgSectionName, "cfg-section", "", "", "Config section name in the default AWS credentials file. To enable priofi")
	// When specifying store in profile the config section name must be provided
	RootCmd.MarkFlagsRequiredTogether("store-profile", "cfg-section")
	RootCmd.PersistentFlags().IntVarP(&duration, "max-duration", "d", 900, `Override default max session duration, in seconds, of the role session [900-43200]. 
NB: This cannot be higher than the 3600 as the API does not allow for AssumeRole for sessions longer than an hour`)
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
}
