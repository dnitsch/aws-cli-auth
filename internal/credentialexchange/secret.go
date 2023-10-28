package credentialexchange

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/werf/lockgate"
	"github.com/werf/lockgate/pkg/file_locker"
	"github.com/zalando/go-keyring"
)

var (
	ErrUnableToLoadAWSCred = errors.New("unable to laod AWS credential")
)

// bit of an antipattern to store types away from their business objects
type AWSCredentials struct {
	Version         int
	AWSAccessKey    string    `json:"AccessKeyId"`
	AWSSecretKey    string    `json:"SecretAccessKey"`
	AWSSessionToken string    `json:"SessionToken"`
	PrincipalARN    string    `json:"-"`
	Expires         time.Time `json:"Expiration"`
}

// AWSRole aws role attributes
type AWSRole struct {
	RoleARN      string
	PrincipalARN string
	Name         string
	Duration     int
}

type SecretStore struct {
	AWSCredentials *AWSCredentials
	AWSCredJson    string
	roleArn        string
	lockDir        string
	locker         lockgate.Locker
	lockResource   string
	secretService  string
	secretUser     string
}

func NewSecretStore(role string) (*SecretStore, error) {
	namer := fmt.Sprintf("%s-%s", SELF_NAME, RoleKeyConverter(role))
	lockDir := os.TempDir() + "/aws-clie-auth-lock"
	locker, err := file_locker.NewFileLocker(lockDir)
	if err != nil {
		return nil, fmt.Errorf("Can't setup lock dir: %s", lockDir)
	}
	return &SecretStore{
		lockDir:       lockDir,
		locker:        locker,
		lockResource:  namer,
		secretService: namer,
		roleArn:       role,
		secretUser:    os.Getenv("USER"),
	}, nil
}

func (s *SecretStore) load() error {
	acquired, lock, err := s.locker.Acquire(s.lockResource, lockgate.AcquireOptions{Shared: false, Timeout: 3 * time.Minute})
	if err != nil {
		// Errorf("Can't load secret due to locked now")
		// Exit(err)
		return err
	}
	defer func() {
		if acquired {
			if err := s.locker.Release(lock); err != nil {
				// Errorf("Can't unlock")
				// Exit(err)
				fmt.Fprintf(os.Stderr, "")
			}
		}
	}()

	if !acquired {
		// Errorf("Can't load secret due to locked now")
		// Exit(err)
		return err
	}

	creds := &AWSCredentials{}

	jsonStr, err := keyring.Get(s.secretService, s.secretUser)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil
		}
		// Errorf("Can't load secret due to unexpected error: %v", err)
		// Exit(err)
		return err
	}

	if err := json.Unmarshal([]byte(jsonStr), &creds); err != nil {
		// Errorf("Can't load secret due to broken data: %v", err)
		// Exit(err)
		return err
	}
	if err := WriteIniSection(s.roleArn); err != nil {
		// Errorf("Can't save role to ")
		// Exit(err)
		return err
	}

	s.AWSCredentials = creds
	s.AWSCredJson = jsonStr
	return nil
}

func (s *SecretStore) save() error {
	acquired, lock, err := s.locker.Acquire(s.lockResource, lockgate.AcquireOptions{Shared: false, Timeout: 3 * time.Minute})

	if err != nil {
		// Errorf("Can't save secret due to lock")
		// Exit(err)
		return err

	}

	defer func() {
		if acquired {
			if err := s.locker.Release(lock); err != nil {
				// Errorf("Can't unlock")
				// Exit(err)
				// return err
				fmt.Fprintf(os.Stderr, "Can't unlock: %s", err)
			}
		}
	}()

	if err := keyring.Set(s.secretService, s.secretUser, s.AWSCredJson); err != nil {
		// Errorf("Can't save secret: %v", err)
		// Exit(err)
		return err
	}
	return nil
}

func (s *SecretStore) AWSCredential() (*AWSCredentials, error) {
	if err := s.load(); err != nil {
		return nil, fmt.Errorf("secret store: %s, %w", err, ErrUnableToLoadAWSCred)
	}

	if s.AWSCredentials == nil && s.AWSCredJson == "" {
		// Infof("Not found the credential for %s", s.roleArn)
		return nil, nil
	}

	fmt.Fprintf(os.Stderr, "Got credential from OS secret store for %s", s.roleArn)

	return s.AWSCredentials, nil
}

func (s *SecretStore) SaveAWSCredential(cred *AWSCredentials) error {
	s.AWSCredentials = cred
	jsonStr, err := json.Marshal(cred)
	if err != nil {
		// Errorf("Can't save secret due to the broken data")
		// Exit(err)
		return err
	}
	s.AWSCredJson = string(jsonStr)
	if err := s.save(); err != nil {
		return err
	}

	fmt.Fprint(os.Stderr, "The AWS credential has been saved in OS secret store")
	return nil
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
		keyring.Delete(fmt.Sprintf("%s-%s", SELF_NAME, v), s.secretUser)
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
