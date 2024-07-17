package cmd

import (
	"fmt"
	"os"
	"os/user"

	"github.com/dnitsch/aws-cli-auth/internal/credentialexchange"
	"github.com/dnitsch/aws-cli-auth/internal/web"
	"github.com/spf13/cobra"
)

var (
	force    bool
	ClearCmd = &cobra.Command{
		Use:   "clear-cache <flags>",
		Short: "Clears any stored credentials in the OS secret store",
		RunE:  clear,
	}
)

func init() {
	cobra.OnInitialize(samlInitConfig)
	ClearCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, `If aws-cli-auth exited improprely in a previous run there is a chance that there could be hanging processes left over.

This will forcefully all chromium processes.

If you are on a windows machine and also use chrome as your current/main browser this will also kill those processes. 

Use with caution.
`)
	RootCmd.AddCommand(ClearCmd)
}

func clear(cmd *cobra.Command, args []string) error {
	user, err := user.Current()
	if err != nil {
		return err
	}
	secretStore, err := credentialexchange.NewSecretStore("",
		fmt.Sprintf("%s-%s", credentialexchange.SELF_NAME, credentialexchange.RoleKeyConverter("")),
		os.TempDir(), user.Username)

	if err != nil {
		return err
	}

	if force {
		w := &web.Web{}
		if err := w.ForceKill(datadir); err != nil {
			return err
		}
		fmt.Fprint(os.Stderr, "Chromium Cache cleared")
	}

	if err := secretStore.ClearAll(); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
	}

	if err := os.Remove(credentialexchange.ConfigIniFile("")); err != nil {
		return err
	}

	return nil
}
