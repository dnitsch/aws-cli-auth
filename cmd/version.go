package cmd

import (
	"fmt"

	"github.com/dnitsch/aws-cli-auth/internal/config"
	"github.com/dnitsch/aws-cli-auth/internal/util"
	"github.com/spf13/cobra"
)

var (
	Version  string = "0.0.1"
	Revision string = "1111aaaa"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: fmt.Sprintf("Get version number %s", config.SELF_NAME),
	Long:  `Version and Revision number of the installed CLI`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\nRevision: %s\n", Version, Revision)
		util.CleanExit()
	},
}
