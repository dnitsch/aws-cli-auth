package auth

import (
	"fmt"
	"os/user"
	"runtime"

	"github.com/dnitsch/aws-cli-auth/internal/config"
	"github.com/dnitsch/aws-cli-auth/internal/util"
	"github.com/dnitsch/aws-cli-auth/internal/web"
)

func GetSamlCreds(conf config.SamlConfig) {
	if conf.BaseConfig.CfgSectionName == "" && conf.BaseConfig.StoreInProfile {
		util.Writeln("Config-Section name must be provided if store-profile is enabled")
		util.Exit(nil)
	}

	web := web.New()
	var awsCreds *util.AWSCredentials
	var err error

	os := runtime.GOOS
	util.Writeln("Is OS: %s\nAnd conf.BaseConfig.StoreInProfile: %v", os, conf.BaseConfig.StoreInProfile)

	// Try to reuse stored credential in secret
	if !conf.BaseConfig.StoreInProfile {
		awsCreds, err = util.AWSCredential(conf.BaseConfig.Role)
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
			fmt.Printf("%v", err)
			util.Exit(err)
		}

		awsCreds.Version = 1
		util.SaveAWSCredential(conf.BaseConfig.Role, awsCreds)
	}

	util.SetCredentials(awsCreds, conf)
}
