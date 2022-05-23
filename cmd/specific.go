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
		RunE: specific,
	}
)

func init() {
	specificCmd.PersistentFlags().StringVarP(&method, "method", "m", "WEB_ID", "Runs a specific credentialProvider as opposed to relying on the default chain provider fallback")
	rootCmd.AddCommand(specificCmd)
}

func specific(cmd *cobra.Command, args []string) error {
	var awsCreds *util.AWSCredentials
	var err error
	if method != "" {
		switch method {
		case "WEB_ID":
			awsCreds, err = auth.LoginAwsWebToken(os.Getenv("USER")) // TODO: redo this getUser implementation
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported Method: %s", method)
		}
	}
	config := config.SamlConfig{BaseConfig: config.BaseConfig{StoreInProfile: storeInProfile}}

	if role != "" {
		awsCreds, err = auth.AssumeRoleWithCreds(awsCreds, os.Getenv("USER"), role)
		if err != nil {
			return err
		}
	}

	if err := util.SetCredentials(awsCreds, config); err != nil {
		return err
	}
	return nil
}
