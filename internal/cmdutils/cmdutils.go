package cmdutils

import (
	"context"
	"errors"
	"fmt"
	"os/user"

	"github.com/dnitsch/aws-cli-auth/internal/credentialexchange"
	"github.com/dnitsch/aws-cli-auth/internal/web"
)

var (
	ErrMissingArg       = errors.New("missing arg")
	ErrUnableToValidate = errors.New("unable to validate token")
)

type SecretStorageImpl interface {
	AWSCredential() (*credentialexchange.AWSCredentials, error)
	Clear() error
	ClearAll() error
	SaveAWSCredential(cred *credentialexchange.AWSCredentials) error
}

// GetCredsWebUI
func GetCredsWebUI(ctx context.Context, svc credentialexchange.AuthSamlApi, secretStore SecretStorageImpl, conf credentialexchange.CredentialConfig, webConfig *web.WebConfig) error {
	if conf.BaseConfig.CfgSectionName == "" && conf.BaseConfig.StoreInProfile {
		// Debug("Config-Section name must be provided if store-profile is enabled")
		return fmt.Errorf("Config-Section name must be provided if store-profile is enabled %w", ErrMissingArg)
	}

	// Try to reuse stored credential in secret
	storedCreds, err := secretStore.AWSCredential()
	if err != nil {
		return err
	}

	credsValid, err := credentialexchange.IsValid(ctx, storedCreds, conf.BaseConfig.ReloadBeforeTime, svc)
	if err != nil {
		return fmt.Errorf("failed to validate: %s, %w", err, ErrUnableToValidate)
	}

	if !credsValid {
		if conf.IsSso {
			return refreshAwsSsoCreds(ctx, conf, secretStore, svc, webConfig)
		}
		return refreshSamlCreds(ctx, conf, secretStore, svc, webConfig)
	}

	return credentialexchange.SetCredentials(storedCreds, conf)
}

func refreshAwsSsoCreds(ctx context.Context, conf credentialexchange.CredentialConfig, secretStore SecretStorageImpl, svc credentialexchange.AuthSamlApi, webConfig *web.WebConfig) error {
	webBrowser := web.New(webConfig)
	capturedCreds, err := webBrowser.GetSSOCredentials(conf)
	if err != nil {
		return err
	}
	awsCreds := &credentialexchange.AWSCredentials{}
	awsCreds.FromRoleCredString(capturedCreds)
	return completeCredStorage(secretStore, awsCreds, conf)
}

func refreshSamlCreds(ctx context.Context, conf credentialexchange.CredentialConfig, secretStore SecretStorageImpl, svc credentialexchange.AuthSamlApi, webConfig *web.WebConfig) error {

	webBrowser := web.New(webConfig)

	samlResp, err := webBrowser.GetSamlLogin(conf)
	if err != nil {
		return err
	}
	user, err := user.Current()
	if err != nil {
		return err
	}

	roleObj := credentialexchange.AWSRole{
		RoleARN:      conf.BaseConfig.Role,
		PrincipalARN: conf.PrincipalArn,
		Name:         credentialexchange.SessionName(user.Username, credentialexchange.SELF_NAME),
		Duration:     conf.Duration,
	}
	awsCreds, err := credentialexchange.LoginStsSaml(ctx, samlResp, roleObj, svc)
	if err != nil {
		return err
	}
	return completeCredStorage(secretStore, awsCreds, conf)
}

func completeCredStorage(secretStore SecretStorageImpl, awsCreds *credentialexchange.AWSCredentials, conf credentialexchange.CredentialConfig) error {
	awsCreds.Version = 1
	if err := secretStore.SaveAWSCredential(awsCreds); err != nil {
		return err
	}
	return credentialexchange.SetCredentials(awsCreds, conf)
}
