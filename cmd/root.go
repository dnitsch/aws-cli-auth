package cmd

import (
	"fmt"
	"os"

	"github.com/dnitsch/aws-cli-auth/internal/config"
	"github.com/dnitsch/aws-cli-auth/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgSectionName     string
	cfgFile            string
	storeInProfile     bool
	killHangingProcess bool
	verbose            bool
	rootCmd            = &cobra.Command{
		Use:   "aws-cli-auth",
		Short: "CLI tool for retrieving AWS temporary credentials",
		Long: `CLI tool for retrieving AWS temporary credentials using SAML providers, or specified method of retrieval - i.e. force AWS_WEB_IDENTITY.
Useful in situations like CI jobs or containers where multiple env vars might be present.
Stores them under the $HOME/.aws/credentials file under a specified path or returns the crednetial_process payload for use in config`,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		util.Exit(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&role, "role", "r", "", "Set the role you want to assume when SAML or OIDC process completes")
	rootCmd.PersistentFlags().StringVarP(&cfgSectionName, "cfg-section", "", "", "config section name in the yaml config file")
	rootCmd.PersistentFlags().BoolVarP(&storeInProfile, "store-profile", "s", false, "By default the credentials are returned to stdout to be used by the credential_process. Set this flag to instead store the credentials under a named profile section")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(fmt.Sprintf(".%s", config.SELF_NAME))
	}

	viper.AutomaticEnv()
	viper.WriteConfig()

	util.IsTraceEnabled = verbose

	if err := viper.ReadInConfig(); err == nil {
		util.Traceln("Using config file:", viper.ConfigFileUsed())
	}
}
