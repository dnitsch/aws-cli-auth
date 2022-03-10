package cmd

import (
	"fmt"
	"os"

	"github.com/dnitsch/aws-cli-auth/internal/auth"
	"github.com/dnitsch/aws-cli-auth/internal/config"
	"github.com/dnitsch/aws-cli-auth/internal/util"
	"github.com/spf13/cobra"
)

var (
	method      string
	specificCmd = &cobra.Command{
		Use:   "specific <flags>",
		Short: "Initiates a specific crednetial provider [WEB_ID]",
		Long: `Initiates a specific crednetial provider [WEB_ID] as opposed to relying on the defaultCredentialChain provider.
This is useful in CI situations where various authentication forms maybe present from AWS_ACCESS_KEY as env vars to metadata of the node.
Returns the same JSON object as the call to the AWS cli for any of the sts AssumeRole* commands`,
		Run: specific,
	}
)

// var strategy map[string]func

func init() {
	specificCmd.PersistentFlags().StringVarP(&method, "method", "m", "WEB_ID", "Runs a specific credentialProvider as opposed to rel")
	rootCmd.AddCommand(specificCmd)
}

func specific(cmd *cobra.Command, args []string) {
	var awsCreds *util.AWSCredentials
	var err error
	if method != "" {
		switch method {
		case "WEB_ID":
			awsCreds, err = auth.LoginAwsWebToken(os.Getenv("USER"))
			if err != nil {
				util.Exit(err)
			}
		default:
			util.Exit(fmt.Errorf("Unsupported Method: %s", method))
		}
	}
	config := config.SamlConfig{BaseConfig: config.BaseConfig{StoreInProfile: storeInProfile}}

	util.SetCredentials(awsCreds, config)
}
