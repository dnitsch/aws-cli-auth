package cmd

import (
	"fmt"

	"github.com/dnitsch/aws-cli-auth/internal/auth"
	"github.com/dnitsch/aws-cli-auth/internal/config"
	"github.com/spf13/cobra"
)

var (
	providerUrl      string
	principalArn     string
	acsUrl           string
	role             string
	duration         int
	reloadBeforeTime int
	samlCmd          = &cobra.Command{
		Use:   "saml <SAML ProviderUrl>",
		Short: "Get AWS credentials and out to stdout",
		Long:  `Get AWS credentials and out to stdout through your SAML provider authentication.`,
		RunE:  getSaml,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if reloadBeforeTime != 0 && reloadBeforeTime > duration {
				return fmt.Errorf("reload-before: %v, must be less than duration (-d): %v", reloadBeforeTime, duration)
			}
			return nil
		},
	}
)

func init() {
	samlCmd.PersistentFlags().StringVarP(&providerUrl, "provider", "p", "", "Saml Entity StartSSO Url")
	samlCmd.MarkPersistentFlagRequired("provider")
	samlCmd.PersistentFlags().StringVarP(&principalArn, "principal", "", "", "Principal Arn of the SAML IdP in AWS")
	samlCmd.MarkPersistentFlagRequired("principal")
	samlCmd.PersistentFlags().StringVarP(&acsUrl, "acsurl", "a", "https://signin.aws.amazon.com/saml", "Override the default ACS Url, used for checkin the post of the SAMLResponse")
	samlCmd.PersistentFlags().IntVarP(&duration, "max-duration", "d", 900, "Override default max session duration, in seconds, of the role session [900-43200]")
	samlCmd.MarkPersistentFlagRequired("max-duration")
	samlCmd.PersistentFlags().IntVarP(&reloadBeforeTime, "reload-before", "", 0, "Triggers a credentials refresh before the specified max-duration. Value provided in seconds. Should be less than the max-duration of the session")
	rootCmd.AddCommand(samlCmd)
}

func getSaml(cmd *cobra.Command, args []string) error {
	conf := config.SamlConfig{
		ProviderUrl:  providerUrl,
		PrincipalArn: principalArn,
		Duration:     duration,
		AcsUrl:       acsUrl,
		BaseConfig: config.BaseConfig{
			StoreInProfile:       storeInProfile,
			Role:                 role,
			CfgSectionName:       cfgSectionName,
			DoKillHangingProcess: killHangingProcess,
			ReloadBeforeTime:     reloadBeforeTime,
		},
	}

	if err := auth.GetSamlCreds(conf); err != nil {
		return err
	}
	return nil
}
