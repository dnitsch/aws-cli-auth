package cmd

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/dnitsch/aws-cli-auth/internal/cmdutils"
	"github.com/dnitsch/aws-cli-auth/internal/credentialexchange"
	"github.com/spf13/cobra"
)

var (
	ErrUnableToCreateSession = errors.New("sts - cannot start a new session")
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
	conf := credentialexchange.SamlConfig{
		ProviderUrl:  providerUrl,
		PrincipalArn: principalArn,
		Duration:     duration,
		AcsUrl:       acsUrl,
		BaseConfig: credentialexchange.BaseConfig{
			StoreInProfile:       storeInProfile,
			Role:                 role,
			CfgSectionName:       cfgSectionName,
			DoKillHangingProcess: killHangingProcess,
			ReloadBeforeTime:     reloadBeforeTime,
		},
	}

	sess, err := session.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session %s, %w", err, ErrUnableToCreateSession)
	}

	svc := sts.New(sess)

	if err := cmdutils.GetSamlCreds(svc, conf); err != nil {
		return err
	}
	return nil
}
