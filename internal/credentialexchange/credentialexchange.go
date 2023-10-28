package credentialexchange

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
)

var (
	ErrUnableAssume        = errors.New("unable to assume")
	ErrUnableSessionCreate = errors.New("unable to create a sesion")
	ErrTokenExpired        = errors.New("token expired")
)

// AWSRole aws role attributes
type AWSRoleConfig struct {
	RoleARN      string
	PrincipalARN string
	Name         string
}

type AuthSamlApi interface {
	AssumeRoleWithSAML(ctx context.Context, params *sts.AssumeRoleWithSAMLInput, optFns ...func(*sts.Options)) (*sts.AssumeRoleWithSAMLOutput, error)
	GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

// LoginStsSaml exchanges saml response for STS creds
func LoginStsSaml(ctx context.Context, samlResponse string, role AWSRole, svc AuthSamlApi) (*AWSCredentials, error) {

	params := &sts.AssumeRoleWithSAMLInput{
		PrincipalArn:    aws.String(role.PrincipalARN), // Required
		RoleArn:         aws.String(role.RoleARN),      // Required
		SAMLAssertion:   aws.String(samlResponse),      // Required
		DurationSeconds: aws.Int32(int32(role.Duration)),
	}

	resp, err := svc.AssumeRoleWithSAML(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve STS credentials using SAML: %s, %w", err.Error(), ErrUnableAssume)
	}

	return &AWSCredentials{
		AWSAccessKey:    *resp.Credentials.AccessKeyId,
		AWSSecretKey:    *resp.Credentials.SecretAccessKey,
		AWSSessionToken: *resp.Credentials.SessionToken,
		PrincipalARN:    *resp.AssumedRoleUser.Arn,
		Expires:         resp.Credentials.Expiration.Local(),
	}, nil
}

type credsProvider struct {
	accessKey, secretKey, sessionToken string
	expiry                             time.Time
}

func (c *credsProvider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	return aws.Credentials{AccessKeyID: c.accessKey, SecretAccessKey: c.secretKey, SessionToken: c.sessionToken, CanExpire: true, Expires: c.expiry}, nil
}

// IsValid checks current credentials and
// returns them if they are still valid
// if reloadTimeBefore is less than time left on the creds
// then it will re-request a login
func IsValid(ctx context.Context, currentCreds *AWSCredentials, reloadBeforeTime int, svc AuthSamlApi) (bool, error) {
	if currentCreds == nil {
		return false, nil
	}

	if _, err := svc.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{}, func(o *sts.Options) {
		o.Credentials = &credsProvider{currentCreds.AWSAccessKey, currentCreds.AWSSecretKey, currentCreds.AWSSessionToken, currentCreds.Expires}
	}); err != nil {
		// var oe *smithy.OperationError
		var oe smithy.APIError
		if errors.As(err, &oe) {
			if oe.ErrorCode() == "ExpiredToken" {
				return false, nil
			}
		}
		return false, fmt.Errorf("the previous credential is invalid: %s, %w", err, ErrUnableAssume)
	}

	return !ReloadBeforeExpiry(currentCreds.Expires, reloadBeforeTime), nil
}

type authWebTokenApi interface {
	AssumeRoleWithWebIdentity(ctx context.Context, params *sts.AssumeRoleWithWebIdentityInput, optFns ...func(*sts.Options)) (*sts.AssumeRoleWithWebIdentityOutput, error)
}

func LoginAwsWebToken(ctx context.Context, username string, svc authWebTokenApi) (*AWSCredentials, error) {
	// var role string
	r, exists := os.LookupEnv(AWS_ROLE_ARN)
	if !exists {
		return nil, fmt.Errorf("roleVar not found, %s is empty", AWS_ROLE_ARN)
	}
	token, err := GetWebIdTokenFileContents()
	if err != nil {
		return nil, err
	}

	sessionName := SessionName(username, SELF_NAME)
	input := &sts.AssumeRoleWithWebIdentityInput{
		RoleArn:          &r,
		RoleSessionName:  &sessionName,
		WebIdentityToken: &token,
	}

	resp, err := svc.AssumeRoleWithWebIdentity(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve STS credentials using token file: %s, %w", err.Error(), ErrUnableAssume)
	}

	return &AWSCredentials{
		AWSAccessKey:    *resp.Credentials.AccessKeyId,
		AWSSecretKey:    *resp.Credentials.SecretAccessKey,
		AWSSessionToken: *resp.Credentials.SessionToken,
		PrincipalARN:    *resp.AssumedRoleUser.Arn,
		Expires:         resp.Credentials.Expiration.Local(),
	}, nil
}

type authAssumeRoleCredsApi interface {
	AssumeRole(ctx context.Context, params *sts.AssumeRoleInput, optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
}

// AssumeRoleWithCreds
func AssumeRoleWithCreds(ctx context.Context, svc authAssumeRoleCredsApi, username, role string) (*AWSCredentials, error) {

	sessionName := SessionName(username, SELF_NAME)

	input := &sts.AssumeRoleInput{
		RoleArn:         &role,
		RoleSessionName: &sessionName,
	}
	roleCreds, err := svc.AssumeRole(ctx, input)

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve STS credentials using Role Provided, %w", ErrUnableAssume)
	}

	return &AWSCredentials{
		AWSAccessKey:    *roleCreds.Credentials.AccessKeyId,
		AWSSecretKey:    *roleCreds.Credentials.SecretAccessKey,
		AWSSessionToken: *roleCreds.Credentials.SessionToken,
		PrincipalARN:    *roleCreds.AssumedRoleUser.Arn,
		Expires:         roleCreds.Credentials.Expiration.Local(),
	}, nil
}
