package cmd

import (
	"fmt"
	"os"

	"github.com/dnitsch/aws-cli-auth/internal/credentialexchange"
	"github.com/dnitsch/aws-cli-auth/internal/web"
	"github.com/spf13/cobra"
)

var (
	force    bool
	clearCmd = &cobra.Command{
		Use:   "clear-cache <flags>",
		Short: "Clears any stored credentials in the OS secret store",
		RunE:  clear,
	}
)

func init() {
	cobra.OnInitialize(samlInitConfig)
	clearCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "If aws-cli-auth exited improprely in a previous run there is a chance that there could be hanging processes left over - this will clean them up forcefully")
	rootCmd.AddCommand(clearCmd)
}

func clear(cmd *cobra.Command, args []string) error {

	secretStore, err := credentialexchange.NewSecretStore("",
		fmt.Sprintf("%s-%s", credentialexchange.SELF_NAME, credentialexchange.RoleKeyConverter("")),
		os.TempDir(), "")
	if err != nil {
		return err
	}

	if force {
		w := &web.Web{}
		w.WithConfig(web.NewWebConf(datadir))
		if err := w.ClearCache(); err != nil {
			return err
		}
		fmt.Fprint(os.Stderr, "Chromium Cache cleared")
	}

	secretStore.ClearAll()

	if err := os.Remove(credentialexchange.ConfigIniFile("")); err != nil {
		return err
	}

	return nil
}
