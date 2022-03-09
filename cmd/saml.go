package cmd

import (
	"fmt"
	"os/user"

	"github.com/dnitsch/aws-cli-auth/internal/config"
	"github.com/dnitsch/aws-cli-auth/internal/saml"
	"github.com/dnitsch/aws-cli-auth/internal/util"
	"github.com/dnitsch/aws-cli-auth/internal/web"
	"github.com/spf13/cobra"
)

var (
	providerUrl  string
	principalArn string
	acsUrl       string
	role         string
	duration     int
	samlCmd      = &cobra.Command{
		Use:   "saml <SAML ProviderUrl>",
		Short: "Get AWS credentials and out to stdout",
		Long:  `Get AWS credentials and out to stdout through your SAML provider authentication.`,
		Run:   getSaml,
	}
)

func init() {
	samlCmd.PersistentFlags().StringVarP(&providerUrl, "provider", "p", "", "Saml Entity StartSSO Url")
	samlCmd.PersistentFlags().StringVarP(&principalArn, "principal", "", "", "Principal Arn of the SAML IdP in AWS")
	samlCmd.PersistentFlags().StringVarP(&acsUrl, "acsurl", "a", "https://signin.aws.amazon.com/saml", "Override the default ACS Url, used for checkin the post of the SAMLResponse")
	samlCmd.PersistentFlags().IntVarP(&duration, "max-duration", "d", 900, "Override default max session duration, in seconds, of the role session [900-43200]")
	rootCmd.AddCommand(samlCmd)

}

func getSaml(cmd *cobra.Command, args []string) {
	if cfgSectionName == "" {
		util.Writeln("The SAML provider name is required")
		util.Exit(nil)
	}

	t, err := web.GetSamlLogin(providerUrl, acsUrl)
	if err != nil {
		fmt.Printf("Err: %v", err)
	}
	user, err := user.Current()
	if err != nil {
		fmt.Errorf(err.Error())
	}

	roleObj := &util.AWSRole{RoleARN: role, PrincipalARN: principalArn, Name: util.SessionName(user.Username, config.SELF_NAME), Duration: duration}

	creds, err := saml.LoginStsSaml(t, roleObj)
	if err != nil {
		fmt.Printf("%v", err)
	}

	util.SetCredentials(creds, cfgSectionName, storeInProfile)
}
