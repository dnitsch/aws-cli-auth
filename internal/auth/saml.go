package auth

import (
	"os/user"

	"github.com/dnitsch/aws-cli-auth/internal/config"
	"github.com/dnitsch/aws-cli-auth/internal/util"
	"github.com/dnitsch/aws-cli-auth/internal/web"
)

// GetSamlCreds
func GetSamlCreds(conf config.SamlConfig) {
	if conf.BaseConfig.CfgSectionName == "" && conf.BaseConfig.StoreInProfile {
		util.Writeln("Config-Section name must be provided if store-profile is enabled")
		util.Exit(nil)
	}

	web := web.New()
	secretStore := util.NewSecretStore(conf.BaseConfig.Role)
	var awsCreds *util.AWSCredentials

	var err error

	// Try to reuse stored credential in secret
	if !conf.BaseConfig.StoreInProfile {
		awsCreds, err = secretStore.AWSCredential()
	}

	if !util.IsValid(awsCreds) || err != nil {

		t, err := web.GetSamlLogin(conf)
		if err != nil {
			util.Writeln("Err: %v", err)
		}
		user, err := user.Current()
		if err != nil {
			util.Writeln(err.Error())
		}

		roleObj := &util.AWSRole{RoleARN: conf.BaseConfig.Role, PrincipalARN: conf.PrincipalArn, Name: util.SessionName(user.Username, config.SELF_NAME), Duration: conf.Duration}

		awsCreds, err = LoginStsSaml(t, roleObj)
		if err != nil {
			util.Writeln("%v", err)
			util.Exit(err)
		}

		awsCreds.Version = 1
		secretStore.SaveAWSCredential(awsCreds)
	}

	util.SetCredentials(awsCreds, conf)
}
