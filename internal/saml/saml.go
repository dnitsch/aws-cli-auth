package saml

import (
	"fmt"
	"os/user"

	"github.com/dnitsch/aws-cli-auth/internal/config"
	"github.com/dnitsch/aws-cli-auth/internal/util"
	"github.com/dnitsch/aws-cli-auth/internal/web"
)

func GetSamlCreds(conf config.SamlConfig) {
	if conf.BaseConfig.CfgSectionName == "" && conf.BaseConfig.StoreInProfile {
		util.Writeln("Config-Section name must be provided if store-profile is enabled")
		util.Exit(nil)
	}

	var awsCreds *util.AWSCredentials
	var err error

	// Try to reuse stored credential in secret
	if !conf.BaseConfig.StoreInProfile {
		awsCreds, err = util.AWSCredential(conf.BaseConfig.Role)
	}

	if !util.IsValid(awsCreds) || err != nil {

		t, err := web.GetSamlLogin(conf.ProviderUrl, conf.AcsUrl)
		if err != nil {
			fmt.Printf("Err: %v", err)
		}
		user, err := user.Current()
		if err != nil {
			fmt.Errorf(err.Error())
		}

		roleObj := &util.AWSRole{RoleARN: conf.BaseConfig.Role, PrincipalARN: conf.PrincipalArn, Name: util.SessionName(user.Username, config.SELF_NAME), Duration: conf.Duration}

		awsCreds, err = LoginStsSaml(t, roleObj)
		if err != nil {
			fmt.Printf("%v", err)
			util.Exit(err)
		}

		awsCreds.Version = 1
		util.SaveAWSCredential(conf.BaseConfig.Role, awsCreds)
	}

	util.SetCredentials(awsCreds, conf)
}
