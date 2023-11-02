package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/dnitsch/aws-cli-auth/internal/cmdutils"
	"github.com/dnitsch/aws-cli-auth/internal/credentialexchange"
	"github.com/dnitsch/aws-cli-auth/internal/web"
	"github.com/spf13/cobra"
)

var (
	ErrUnableToCreateSession = errors.New("sts - cannot start a new session")
)

var (
	providerUrl        string
	principalArn       string
	acsUrl             string
	isSso              bool
	ssoRegion          string
	ssoUserEndpoint    string
	ssoFedCredEndpoint string
	role               string
	datadir            string
	duration           int
	reloadBeforeTime   int
	samlCmd            = &cobra.Command{
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
	cobra.OnInitialize(samlInitConfig)
	samlCmd.PersistentFlags().StringVarP(&providerUrl, "provider", "p", "", "Saml Entity StartSSO Url")
	samlCmd.MarkPersistentFlagRequired("provider")
	samlCmd.PersistentFlags().StringVarP(&principalArn, "principal", "", "", "Principal Arn of the SAML IdP in AWS")
	samlCmd.PersistentFlags().StringVarP(&acsUrl, "acsurl", "a", "https://signin.aws.amazon.com/saml", "Override the default ACS Url, used for checkin the post of the SAMLResponse")
	samlCmd.PersistentFlags().StringVarP(&ssoUserEndpoint, "sso-user-endpoint", "", "https://portal.sso.%s.amazonaws.com/user", "UserEndpoint in a go style fmt.Sprintf string with a region placeholder")
	samlCmd.PersistentFlags().StringVarP(&ssoFedCredEndpoint, "sso-fed-endpoint", "", "https://portal.sso.%s.amazonaws.com/federation/credentials/", "FederationCredEndpoint in a go style fmt.Sprintf string with a region placeholder")
	samlCmd.PersistentFlags().StringVarP(&ssoRegion, "sso-region", "", "eu-west-1", "If using SSO, you must set the region")
	samlCmd.PersistentFlags().IntVarP(&duration, "max-duration", "d", 900, "Override default max session duration, in seconds, of the role session [900-43200]")
	samlCmd.PersistentFlags().BoolVarP(&isSso, "is-sso", "", false, `Enables the new AWS User portal login. 
	If this option is specified the role specified must be in the *ACCOUNT_ID:ROLE_NAME*
	
	e.g.: 12345678910:PowerUser
	
	Do not specify the ARN of the role you want to assume.`)
	samlCmd.PersistentFlags().IntVarP(&reloadBeforeTime, "reload-before", "", 0, "Triggers a credentials refresh before the specified max-duration. Value provided in seconds. Should be less than the max-duration of the session")
	rootCmd.AddCommand(samlCmd)
}

func getSaml(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	conf := credentialexchange.CredentialConfig{
		ProviderUrl:  providerUrl,
		PrincipalArn: principalArn,
		Duration:     duration,
		AcsUrl:       acsUrl,
		IsSso:        isSso,
		SsoRegion:    ssoRegion,
		BaseConfig: credentialexchange.BaseConfig{
			StoreInProfile:       storeInProfile,
			Role:                 role,
			CfgSectionName:       cfgSectionName,
			DoKillHangingProcess: killHangingProcess,
			ReloadBeforeTime:     reloadBeforeTime,
		},
	}

	if isSso {
		ssoRole := strings.Split(role, ":")
		if len(ssoRole) > 2 {
			return fmt.Errorf("incorrectly formatted role for AWS SSO - must only be ACCOUNT:ROLE_NAME")
		}
		conf.SsoUserEndpoint = fmt.Sprintf("https://portal.sso.%s.amazonaws.com/user", conf.SsoRegion)
		conf.SsoCredFedEndpoint = fmt.Sprintf("https://portal.sso.%s.amazonaws.com/federation/credentials/", conf.SsoRegion) + fmt.Sprintf("?account_id=%s&role_name=%s&debug=true", ssoRole[0], ssoRole[1])
	}

	datadir := path.Join(credentialexchange.HomeDir(), fmt.Sprintf(".%s-data", credentialexchange.SELF_NAME))
	os.MkdirAll(datadir, 0755)

	secretStore, err := credentialexchange.NewSecretStore(conf.BaseConfig.Role, fmt.Sprintf("%s-%s", credentialexchange.SELF_NAME, credentialexchange.RoleKeyConverter(conf.BaseConfig.Role)), os.TempDir()+"/aws-clie-auth-lock")
	if err != nil {
		return err
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to create session %s, %w", err, ErrUnableToCreateSession)
	}
	svc := sts.NewFromConfig(cfg)

	return cmdutils.GetCredsWebUI(ctx, svc, secretStore, conf, web.NewWebConf(datadir))
}

func samlInitConfig() {
	if _, err := os.Stat(credentialexchange.ConfigIniFile("")); err != nil {
		// creating a file
		rolesInit := []byte(fmt.Sprintf("[%s]\n", credentialexchange.INI_CONF_SECTION))
		err := os.WriteFile(credentialexchange.ConfigIniFile(""), rolesInit, 0644)
		cobra.CheckErr(err)
	}

	datadir = path.Join(credentialexchange.HomeDir(), fmt.Sprintf(".%s-data", credentialexchange.SELF_NAME))

	if _, err := os.Stat(datadir); err != nil {
		cobra.CheckErr(os.MkdirAll(datadir, 0755))
	}
}
