package cmd

import (
	"os"

	"github.com/dnitsch/aws-cli-auth/internal/util"
	"github.com/dnitsch/aws-cli-auth/internal/web"
	"github.com/spf13/cobra"
)

var (
	force    bool
	clearCmd = &cobra.Command{
		Use:   "clear-cache <flags>",
		Short: "Clears any stored credentials in the OS secret store",
		Run:   clear,
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	clearCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "If aws-cli-auth exited improprely in a previous run there is a chance that there could be hanging processes left over - this will clean them up forcefully")
	rootCmd.AddCommand(clearCmd)
}

func clear(cmd *cobra.Command, args []string) {
	web := web.New()
	secretStore := util.NewSecretStore("")

	if force {

		if err := web.ClearCache(); err != nil {
			util.Exit(err)
		}
		util.Writeln("Chromium Cache cleared")
	}
	secretStore.ClearAll()

	if err := os.Remove(util.ConfigIniFile()); err != nil {
		util.Exit(err)
	}
}
