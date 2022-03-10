package auth

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/dnitsch/aws-cli-auth/internal/config"
	"github.com/dnitsch/aws-cli-auth/internal/util"
	"github.com/pkg/errors"
)

// AWSRole aws role attributes
type AWSRole struct {
	RoleARN      string
	PrincipalARN string
	Name         string
}

// LoginStsSaml exchanges saml response for STS creds
func LoginStsSaml(samlResponse string, role *util.AWSRole) (*util.AWSCredentials, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create session")
	}

	svc := sts.New(sess)

	params := &sts.AssumeRoleWithSAMLInput{
		PrincipalArn:    aws.String(role.PrincipalARN), // Required
		RoleArn:         aws.String(role.RoleARN),      // Required
		SAMLAssertion:   aws.String(samlResponse),      // Required
		DurationSeconds: aws.Int64(3600),
	}

	resp, err := svc.AssumeRoleWithSAML(params)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to retrieve STS credentials using SAML")
	}

	return &util.AWSCredentials{
		AWSAccessKey:    aws.StringValue(resp.Credentials.AccessKeyId),
		AWSSecretKey:    aws.StringValue(resp.Credentials.SecretAccessKey),
		AWSSessionToken: aws.StringValue(resp.Credentials.SessionToken),
		PrincipalARN:    aws.StringValue(resp.AssumedRoleUser.Arn),
		Expires:         resp.Credentials.Expiration.Local(),
	}, nil
}

func LoginAwsWebToken(username string) (*util.AWSCredentials, error) {
	// var role string
	sess, err := session.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create session")
	}

	svc := sts.New(sess)
	r, exists := os.LookupEnv(config.AWS_ROLE_ARN)
	if !exists {
		util.Exit(fmt.Errorf("Role Var Not Found"))
	}
	token, err := util.GetWebIdTokenFileContents()
	if err != nil {
		util.Exit(err)
	}
	sessionName := util.SessionName(username, config.SELF_NAME)
	input := &sts.AssumeRoleWithWebIdentityInput{
		RoleArn:          &r,
		RoleSessionName:  &sessionName,
		WebIdentityToken: &token,
	}

	resp, err := svc.AssumeRoleWithWebIdentity(input)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to retrieve STS credentials using SAML")
	}

	return &util.AWSCredentials{
		AWSAccessKey:    aws.StringValue(resp.Credentials.AccessKeyId),
		AWSSecretKey:    aws.StringValue(resp.Credentials.SecretAccessKey),
		AWSSessionToken: aws.StringValue(resp.Credentials.SessionToken),
		PrincipalARN:    aws.StringValue(resp.AssumedRoleUser.Arn),
		Expires:         resp.Credentials.Expiration.Local(),
	}, nil
}
