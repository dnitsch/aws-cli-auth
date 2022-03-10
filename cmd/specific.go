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
		Run:   specific,
	}
)

// var strategy map[string]func

func init() {
	specificCmd.PersistentFlags().StringVarP(&method, "method", "m", "", "If aws-cli-auth exited improprely in a previous run there is a chance that there could be hanging processes left over - this will clean them up forcefully")
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
