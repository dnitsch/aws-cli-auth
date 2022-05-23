package auth

import (
	"os/user"

	"github.com/dnitsch/aws-cli-auth/internal/config"
	"github.com/dnitsch/aws-cli-auth/internal/util"
	"github.com/dnitsch/aws-cli-auth/internal/web"
)

// GetSamlCreds
func GetSamlCreds(conf config.SamlConfig) error {
	if conf.BaseConfig.CfgSectionName == "" && conf.BaseConfig.StoreInProfile {
		util.Writeln("Config-Section name must be provided if store-profile is enabled")
		util.Exit(nil)
	}

	secretStore := util.NewSecretStore(conf.BaseConfig.Role)
	var awsCreds *util.AWSCredentials
	var webBrowser *web.Web
	var err error

	// Try to reuse stored credential in secret
	awsCreds, err = secretStore.AWSCredential()

	if !util.IsValid(awsCreds, conf.BaseConfig.ReloadBeforeTime) || err != nil {
		webBrowser = web.New()

		t, err := webBrowser.GetSamlLogin(conf)
		if err != nil {
			return err
		}
		user, err := user.Current()
		if err != nil {
			return err
		}

		roleObj := &util.AWSRole{RoleARN: conf.BaseConfig.Role, PrincipalARN: conf.PrincipalArn, Name: util.SessionName(user.Username, config.SELF_NAME), Duration: conf.Duration}

		awsCreds, err = LoginStsSaml(t, roleObj)
		if err != nil {
			return err
		}

		awsCreds.Version = 1
		secretStore.SaveAWSCredential(awsCreds)
	}

	util.SetCredentials(awsCreds, conf)
	return nil
}
