// taken from AWS-CLI-OIDC - initially
package util

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dnitsch/aws-cli-auth/internal/config"
	"github.com/werf/lockgate"
	"github.com/werf/lockgate/pkg/file_locker"
	"github.com/zalando/go-keyring"
)

type SecretStore struct {
	AWSCredentials *AWSCredentials
	AWSCredJson    string
	roleArn        string
	lockDir        string
	locker         lockgate.Locker
	lockResource   string
	secretService  string
	secretUser     string
	// keyring        keyring.Keyring
}

func NewSecretStore(role string) *SecretStore {
	namer := fmt.Sprintf("%s-%s", config.SELF_NAME, RoleKeyConverter(role))
	lockDir := os.TempDir() + "/aws-clie-auth-lock"
	locker, err := file_locker.NewFileLocker(lockDir)
	if err != nil {
		Writeln("Can't setup lock dir: %s", lockDir)
		Exit(err)
	}
	return &SecretStore{
		lockDir:       lockDir,
		locker:        locker,
		lockResource:  namer,
		secretService: namer,
		roleArn:       role,
		secretUser:    os.Getenv("USER"),
	}
}

func (s *SecretStore) load() {
	acquired, lock, err := s.locker.Acquire(s.lockResource, lockgate.AcquireOptions{Shared: false, Timeout: 3 * time.Minute})
	if err != nil {
		Writeln("Can't load secret due to locked now")
		Exit(err)
	}
	defer func() {
		if acquired {
			if err := s.locker.Release(lock); err != nil {
				Writeln("Can't unlock")
				Exit(err)
			}
		}
	}()

	if !acquired {
		Writeln("Can't load secret due to locked now")
		Exit(err)
	}

	creds := &AWSCredentials{}

	jsonStr, err := keyring.Get(s.secretService, s.secretUser)
	if err != nil {
		if err == keyring.ErrNotFound {
			return
		}
		Writeln("Can't load secret due to unexpected error: %v", err)
		Exit(err)
	}

	if err := json.Unmarshal([]byte(jsonStr), &creds); err != nil {
		Writeln("Can't load secret due to broken data: %v", err)
		Exit(err)
	}
	if err := WriteIniSection(s.roleArn); err != nil {
		Writeln("Can't save role to ")
		Exit(err)
	}

	s.AWSCredentials = creds
	s.AWSCredJson = jsonStr
}

func (s *SecretStore) save() {
	acquired, lock, err := s.locker.Acquire(s.lockResource, lockgate.AcquireOptions{Shared: false, Timeout: 3 * time.Minute})

	if err != nil {
		Writeln("Can't save secret due to lock")
		Exit(err)
	}

	defer func() {
		if acquired {
			if err := s.locker.Release(lock); err != nil {
				Writeln("Can't unlock")
				Exit(err)
			}
		}
	}()

	if err := keyring.Set(s.secretService, s.secretUser, s.AWSCredJson); err != nil {
		Writeln("Can't save secret: %v", err)
		Exit(err)
	}
}

func (s *SecretStore) AWSCredential() (*AWSCredentials, error) {
	s.load()

	if s.AWSCredentials == nil && s.AWSCredJson == "" {
		Writeln("Not found the credential for %s", s.roleArn)
		return nil, nil
	}

	Writeln("Got credential from OS secret store for %s", s.roleArn)

	return s.AWSCredentials, nil
}

func (s *SecretStore) SaveAWSCredential(cred *AWSCredentials) {
	s.AWSCredentials = cred
	jsonStr, err := json.Marshal(cred)
	if err != nil {
		Writeln("Can't save secret due to the broken data")
		Exit(err)
	}
	s.AWSCredJson = string(jsonStr)
	s.save()

	Write("The AWS credentials has been saved in OS secret store")
}

func (s *SecretStore) Clear() error {
	return keyring.Delete(s.secretService, s.secretUser)
}

func (s *SecretStore) ClearAll() error {
	secretServices, err := GetAllIniSections()
	if err != nil {
		return err
	}

	for _, v := range secretServices {
		keyring.Delete(fmt.Sprintf("%s-%s", config.SELF_NAME, v), s.secretUser)
	}
	return nil
}

// RoleKeyConverter converts a role to a key used for storing in key store
func RoleKeyConverter(role string) string {
	return strings.ReplaceAll(strings.ReplaceAll(role, ":", "_"), "/", "____")
}

// KeyRoleConverter Converts a key back to a role
func KeyRoleConverter(key string) string {
	return strings.ReplaceAll(strings.ReplaceAll(key, "____", "/"), "_", ":")
}
