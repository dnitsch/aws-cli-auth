package credentialexchange

import (
	"context"
	"encoding/json"
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
	ErrMissingEnvVar       = errors.New("missing env var")
	ErrUnmarshalCred       = errors.New("unable to unmarshal credential from string")
)

// AWSRole aws role attributes
type AWSRoleConfig struct {
	RoleARN      string
	PrincipalARN string
	Name         string
}

// AWSCredentials is a representation of the returned credential
type AWSCredentials struct {
	Version         int
	AWSAccessKey    string    `json:"AccessKeyId"`
	AWSSecretKey    string    `json:"SecretAccessKey"`
	AWSSessionToken string    `json:"SessionToken"`
	PrincipalARN    string    `json:"-"`
	Expires         time.Time `json:"Expiration"`
}

func (a *AWSCredentials) FromRoleCredString(cred string) (*AWSCredentials, error) {
	// RoleCreds can be encapsulated in this function
	// never used outside of this scope for now
	type RoleCreds struct {
		RoleCreds struct {
			AccessKey    string `json:"accessKeyId"`
			SecretKey    string `json:"secretAccessKey"`
			SessionToken string `json:"sessionToken"`
			Expiration   int64  `json:"expiration"`
		} `json:"roleCredentials"`
	}
	rc := &RoleCreds{}
	if err := json.Unmarshal([]byte(cred), rc); err != nil {
		return nil, fmt.Errorf("%s, %w", err, ErrUnmarshalCred)
	}
	a.AWSAccessKey = rc.RoleCreds.AccessKey
	a.AWSSecretKey = rc.RoleCreds.SecretKey
	a.AWSSessionToken = rc.RoleCreds.SessionToken
	a.Expires = time.UnixMilli(rc.RoleCreds.Expiration)
	return a, nil
}

type AuthSamlApi interface {
	AssumeRoleWithSAML(ctx context.Context, params *sts.AssumeRoleWithSAMLInput, optFns ...func(*sts.Options)) (*sts.AssumeRoleWithSAMLOutput, error)
	GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
	AssumeRole(ctx context.Context, params *sts.AssumeRoleInput, optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
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
		return nil, fmt.Errorf("roleVar not found, %s is empty, %w", AWS_ROLE_ARN, ErrMissingEnvVar)
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

// AssumeRoleWithCreds uses existing creds retrieved from anywhere
// to pass to a credential provider and assume a specific role
//
// Most common use case is role chaining an WeBId role to a specific one
func assumeRoleWithCreds(ctx context.Context, currentCreds *AWSCredentials, svc AuthSamlApi, username, role string) (*AWSCredentials, error) {

	input := &sts.AssumeRoleInput{
		RoleArn:         &role,
		RoleSessionName: aws.String(SessionName(username, SELF_NAME)),
	}

	roleCreds, err := svc.AssumeRole(ctx, input, func(o *sts.Options) {
		o.Credentials = &credsProvider{currentCreds.AWSAccessKey, currentCreds.AWSSecretKey, currentCreds.AWSSessionToken, currentCreds.Expires}
	})

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

// AssumeRoleInChain loops over all the roles provided
func AssumeRoleInChain(ctx context.Context, baseCreds *AWSCredentials, svc AuthSamlApi, username string, roles []string) (*AWSCredentials, error) {
	var awsCreds *AWSCredentials
	for _, r := range roles {
		c, err := assumeRoleWithCreds(ctx, baseCreds, svc, username, r)
		if err != nil {
			return nil, err
		}
		awsCreds = c
	}
	return awsCreds, nil
}
