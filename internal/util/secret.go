// taken from AWS-CLI-OIDC
package util

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/dnitsch/aws-cli-auth/internal/config"
	"github.com/werf/lockgate"
	"github.com/werf/lockgate/pkg/file_locker"
	"github.com/zalando/go-keyring"
)

var lockDir = os.TempDir() + "/aws-clie-auth-lock"
var locker lockgate.Locker
var lockResource = "aws-cli-auth"

func init() {
	var err error
	locker, err = file_locker.NewFileLocker(lockDir)
	if err != nil {
		Writeln("Can't setup lock dir: %s", lockDir)
		Exit(err)
	}

	Secret.AWSCredentials = make(map[string]string)
}

var secretService = config.SELF_NAME
var secretUser = os.Getenv("USER")

var Secret SecretStore

type SecretStore struct {
	AWSCredentials map[string]string `json:"credentials"`
}

func (s *SecretStore) Load() {
	acquired, lock, err := locker.Acquire(lockResource, lockgate.AcquireOptions{Shared: false, Timeout: 3 * time.Minute})
	if err != nil {
		Writeln("Can't load secret due to locked now")
		Exit(err)
	}
	defer func() {
		if acquired {
			if err := locker.Release(lock); err != nil {
				Writeln("Can't unlock")
				Exit(err)
			}
		}
	}()

	if !acquired {
		Writeln("Can't load secret due to locked now")
		Exit(err)
	}

	jsonStr, err := keyring.Get(secretService, secretUser)
	if err != nil {
		if err == keyring.ErrNotFound {
			return
		}
		Writeln("Can't load secret due to unexpected error: %v", err)
		Exit(err)
	}
	if err := json.Unmarshal([]byte(jsonStr), &s); err != nil {
		Writeln("Can't load secret due to broken data: %v", err)
		Exit(err)
	}
}

func (s *SecretStore) Save(roleArn, cred string) {
	acquired, lock, err := locker.Acquire(lockResource, lockgate.AcquireOptions{Shared: false, Timeout: 3 * time.Minute})
	if err != nil {
		Writeln("Can't save secret due to locked now")
		Exit(err)
	}
	defer func() {
		if acquired {
			if err := locker.Release(lock); err != nil {
				Writeln("Can't unlock")
				Exit(err)
			}
		}
	}()

	// Load the latest credentials
	jsonStr, err := keyring.Get(secretService, secretUser)
	if err != nil {
		if err != keyring.ErrNotFound {
			Writeln("Can't load secret due to unexpected error: %v", err)
			Exit(err)
		}
	}
	if jsonStr != "" {
		if err := json.Unmarshal([]byte(jsonStr), &s); err != nil {
			Writeln("Can't load secret due to broken data: %v", err)
			Exit(err)
		}
	}

	// Add/Update credential
	s.AWSCredentials[roleArn] = cred

	// Save
	newJsonStr, err := json.Marshal(s)
	if err != nil {
		Writeln("Can't unlock: %v", err)
		Exit(err)
	}
	if err := keyring.Set(secretService, secretUser, string(newJsonStr)); err != nil {
		Writeln("Can't save secret: %v", err)
		Exit(err)
	}
}

func AWSCredential(roleArn string) (*AWSCredentials, error) {
	Secret.Load()

	jsonStr, ok := Secret.AWSCredentials[roleArn]
	if !ok {
		Exit(fmt.Errorf("Not found the credential for %s", roleArn))
	}

	Writeln("Got credential from OS secret store for %s", roleArn)

	var cred AWSCredentials

	err := json.Unmarshal([]byte(jsonStr), &cred)
	if err != nil {
		Writeln("Can't load secret due to the broken data")
		Exit(err)
	}

	return &cred, nil
}

func SaveAWSCredential(roleArn string, cred *AWSCredentials) {
	jsonStr, err := json.Marshal(cred)
	if err != nil {
		Writeln("Can't save secret due to the broken data")
		Exit(err)
	}

	Secret.Save(roleArn, string(jsonStr))

	Write("The AWS credentials has been saved in OS secret store")
}

func Clear() error {
	return keyring.Delete(secretService, secretUser)
}

func IsValid(cred *AWSCredentials) bool {
	if cred == nil {
		return false
	}

	sess, err := session.NewSession()
	if err != nil {
		Writeln("Failed to create aws client session")
		Exit(err)
	}

	creds := credentials.NewStaticCredentialsFromCreds(credentials.Value{
		AccessKeyID:     cred.AWSAccessKey,
		SecretAccessKey: cred.AWSSecretKey,
		SessionToken:    cred.AWSSessionToken,
	})

	svc := sts.New(sess, aws.NewConfig().WithCredentials(creds))

	input := &sts.GetCallerIdentityInput{}

	_, err = svc.GetCallerIdentity(input)

	if err != nil {
		Writeln("The previous credential isn't valid")
	}

	return err == nil
}
