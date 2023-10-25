package credentialexchange

import (
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"
)

var (
	ErrUnableAssume        = errors.New("unable to assume")
	ErrUnableSessionCreate = errors.New("unable to create a sesion")
)

// AWSRole aws role attributes
type AWSRoleConfig struct {
	RoleARN      string
	PrincipalARN string
	Name         string
}

type AuthSamlApi interface {
	AssumeRoleWithSAML(input *sts.AssumeRoleWithSAMLInput) (*sts.AssumeRoleWithSAMLOutput, error)
}

// LoginStsSaml exchanges saml response for STS creds
func LoginStsSaml(samlResponse string, role AWSRole, svc AuthSamlApi) (*AWSCredentials, error) {

	params := &sts.AssumeRoleWithSAMLInput{
		PrincipalArn:    aws.String(role.PrincipalARN), // Required
		RoleArn:         aws.String(role.RoleARN),      // Required
		SAMLAssertion:   aws.String(samlResponse),      // Required
		DurationSeconds: aws.Int64(int64(role.Duration)),
	}

	resp, err := svc.AssumeRoleWithSAML(params)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve STS credentials using SAML: %s, %w", err.Error(), ErrUnableAssume)
	}

	return &AWSCredentials{
		AWSAccessKey:    aws.StringValue(resp.Credentials.AccessKeyId),
		AWSSecretKey:    aws.StringValue(resp.Credentials.SecretAccessKey),
		AWSSessionToken: aws.StringValue(resp.Credentials.SessionToken),
		PrincipalARN:    aws.StringValue(resp.AssumedRoleUser.Arn),
		Expires:         resp.Credentials.Expiration.Local(),
	}, nil
}

type authWebTokenApi interface {
	AssumeRoleWithWebIdentity(input *sts.AssumeRoleWithWebIdentityInput) (*sts.AssumeRoleWithWebIdentityOutput, error)
}

func LoginAwsWebToken(username string, svc authWebTokenApi) (*AWSCredentials, error) {
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

	resp, err := svc.AssumeRoleWithWebIdentity(input)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve STS credentials using token file: %s, %w", err.Error(), ErrUnableAssume)
	}

	return &AWSCredentials{
		AWSAccessKey:    aws.StringValue(resp.Credentials.AccessKeyId),
		AWSSecretKey:    aws.StringValue(resp.Credentials.SecretAccessKey),
		AWSSessionToken: aws.StringValue(resp.Credentials.SessionToken),
		PrincipalARN:    aws.StringValue(resp.AssumedRoleUser.Arn),
		Expires:         resp.Credentials.Expiration.Local(),
	}, nil
}

type authAssumeRoleCredsApi interface {
	AssumeRole(input *sts.AssumeRoleInput) (*sts.AssumeRoleOutput, error)
}

// AssumeRoleWithCreds
func AssumeRoleWithCreds(svc authAssumeRoleCredsApi, username, role string) (*AWSCredentials, error) {

	sessionName := SessionName(username, SELF_NAME)

	input := &sts.AssumeRoleInput{
		RoleArn:         &role,
		RoleSessionName: &sessionName,
	}
	roleCreds, err := svc.AssumeRole(input)

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve STS credentials using Role Provided, %w", ErrUnableAssume)
	}

	return &AWSCredentials{
		AWSAccessKey:    aws.StringValue(roleCreds.Credentials.AccessKeyId),
		AWSSecretKey:    aws.StringValue(roleCreds.Credentials.SecretAccessKey),
		AWSSessionToken: aws.StringValue(roleCreds.Credentials.SessionToken),
		PrincipalARN:    aws.StringValue(roleCreds.AssumedRoleUser.Arn),
		Expires:         roleCreds.Credentials.Expiration.Local(),
	}, nil
}
