package cmdutils

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path"

	"github.com/dnitsch/aws-cli-auth/internal/credentialexchange"
	"github.com/dnitsch/aws-cli-auth/internal/web"
)

var (
	ErrMissingArg = errors.New("missing arg")
)

// GetSamlCreds
func GetSamlCreds(svc credentialexchange.AuthSamlApi, conf credentialexchange.SamlConfig) error {
	if conf.BaseConfig.CfgSectionName == "" && conf.BaseConfig.StoreInProfile {
		// Debug("Config-Section name must be provided if store-profile is enabled")
		return fmt.Errorf("Config-Section name must be provided if store-profile is enabled %w", ErrMissingArg)
	}

	secretStore, err := credentialexchange.NewSecretStore(conf.BaseConfig.Role)
	if err != nil {
		return err
	}

	// Try to reuse stored credential in secret
	storedCreds, err := secretStore.AWSCredential()
	if err != nil {
		return err
	}

	// creds := credentials.NewStaticCredentialsFromCreds(credentials.Value{
	// 	AccessKeyID:     storedCreds.AWSAccessKey,
	// 	SecretAccessKey: storedCreds.AWSSecretKey,
	// 	SessionToken:    storedCreds.AWSSessionToken,
	// })
	// svc.Config.Credentials = creds

	credsValid, err := credentialexchange.IsValid(storedCreds, conf.BaseConfig.ReloadBeforeTime)
	if err != nil {
		return err
	}
	if !credsValid || err != nil {
		if err := refreshCreds(conf, secretStore, svc); err != nil {
			return err
		}
	}

	credentialexchange.SetCredentials(storedCreds, conf)
	return nil
}

func refreshCreds(conf credentialexchange.SamlConfig, secretStore *credentialexchange.SecretStore, svc credentialexchange.AuthSamlApi) error {

	datadir := path.Join(credentialexchange.HomeDir(), fmt.Sprintf(".%s-data", credentialexchange.SELF_NAME))
	os.MkdirAll(datadir, 0755)

	webBrowser := web.New(datadir)

	samlResp, err := webBrowser.GetSamlLogin(conf)
	if err != nil {
		return err
	}
	user, err := user.Current()
	if err != nil {
		return err
	}

	roleObj := &credentialexchange.AWSRole{RoleARN: conf.BaseConfig.Role, PrincipalARN: conf.PrincipalArn, Name: credentialexchange.SessionName(user.Username, credentialexchange.SELF_NAME), Duration: conf.Duration}

	awsCreds, err := credentialexchange.LoginStsSaml(samlResp, *roleObj, svc)
	if err != nil {
		return err
	}

	awsCreds.Version = 1
	if err := secretStore.SaveAWSCredential(awsCreds); err != nil {
		return err
	}
	return nil
}
