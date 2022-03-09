package saml

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/dnitsch/aws-cli-auth/internal/util"
	"github.com/pkg/errors"
)

// AWSRole aws role attributes
type AWSRole struct {
	RoleARN      string
	PrincipalARN string
	Name         string
}

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
