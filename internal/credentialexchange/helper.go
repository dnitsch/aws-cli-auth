package credentialexchange

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	ini "gopkg.in/ini.v1"
)

var (
	ErrSectionNotFound = errors.New("section not found")
	ErrConfigFailure   = errors.New("config error")
)

func HomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("unable to get the user home dir")
	}
	return home
}

func ConfigIniFile(basePath string) string {
	var base string
	if basePath != "" {
		base = basePath
	} else {
		base = HomeDir()
	}
	return path.Join(base, fmt.Sprintf(".%s.ini", SELF_NAME))
}

func SessionName(username, selfName string) string {
	return fmt.Sprintf("%s-%s", username, selfName)
}

func SetCredentials(creds *AWSCredentials, config SamlConfig) error {

	if config.BaseConfig.StoreInProfile {
		if err := storeCredentialsInProfile(*creds, config.BaseConfig.CfgSectionName); err != nil {
			return err
		}
		return nil
	}
	return returnStdOutAsJson(*creds)
}

func storeCredentialsInProfile(creds AWSCredentials, configSection string) error {
	var awsConfPath string

	if overriddenpath, exists := os.LookupEnv("AWS_SHARED_CREDENTIALS_FILE"); exists {
		awsConfPath = overriddenpath
	} else {
		// os.MkdirAll(datadir, 0755)
		awsCredsPath := path.Join(HomeDir(), ".aws", "credentials")
		if _, err := os.Stat(awsCredsPath); os.IsNotExist(err) {
			os.Mkdir(awsCredsPath, 0655)
		}
		awsConfPath = awsCredsPath
	}

	cfg, err := ini.Load(awsConfPath)
	if err != nil {
		return err
	}
	cfg.Section(configSection).Key("aws_access_key_id").SetValue(creds.AWSAccessKey)
	cfg.Section(configSection).Key("aws_secret_access_key").SetValue(creds.AWSSecretKey)
	cfg.Section(configSection).Key("aws_session_token").SetValue(creds.AWSSessionToken)
	cfg.SaveTo(awsConfPath)

	return nil
}

func returnStdOutAsJson(creds AWSCredentials) error {
	creds.Version = 1

	jsonBytes, err := json.Marshal(creds)
	if err != nil {
		// Errorf("Unexpected AWS credential response")
		return err
	}
	fmt.Println(string(jsonBytes))
	return nil
}

func GetWebIdTokenFileContents() (string, error) {
	// var content *string
	file, exists := os.LookupEnv(WEB_ID_TOKEN_VAR)
	if !exists {
		return "", fmt.Errorf("FileNotPresent: %s", WEB_ID_TOKEN_VAR)
	}
	content, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

type callerIdApi interface {
	GetCallerIdentity(input *sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error)
}

// IsValid checks current credentials and
// returns them if they are still valid
// if reloadTimeBefore is less than time left on the creds
// then it will re-request a login
func IsValid(currentCreds *AWSCredentials, relaodBeforeTime int) (bool, error) {
	if currentCreds == nil {
		return false, nil
	}

	sess, err := session.NewSession()
	if err != nil {
		return false, fmt.Errorf("session error: %s, %w", err, ErrUnableSessionCreate)
	}

	creds := credentials.NewStaticCredentialsFromCreds(credentials.Value{
		AccessKeyID:     currentCreds.AWSAccessKey,
		SecretAccessKey: currentCreds.AWSSecretKey,
		SessionToken:    currentCreds.AWSSessionToken,
	})

	svc := sts.New(sess, aws.NewConfig().WithCredentials(creds))
	svc.Config.Credentials = creds // aws.NewConfig().WithCredentials(creds)
	input := &sts.GetCallerIdentityInput{}

	if _, err := svc.GetCallerIdentity(input); err != nil {
		// Errorf("The previous credential isn't valid")
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == sts.ErrCodeExpiredTokenException {
				return false, nil
			}
		}
		return false, fmt.Errorf("the previous credential isn't valid: %w", ErrUnableAssume)
	}

	return !ReloadBeforeExpiry(currentCreds.Expires, relaodBeforeTime), nil
}

// ReloadBeforeExpiry returns true if the time
// to expiry is less than the specified time in seconds
// false if there is more than required time in seconds
// before needing to recycle credentials
func ReloadBeforeExpiry(expiry time.Time, reloadBeforeSeconds int) bool {
	now := time.Now()
	diff := expiry.Sub(now)
	return diff.Seconds() < float64(reloadBeforeSeconds)
}

// WriteIniSection update ini sections in own config file
func WriteIniSection(role string) error {
	section := fmt.Sprintf("%s.%s", INI_CONF_SECTION, RoleKeyConverter(role))
	cfg, err := ini.Load(ConfigIniFile(""))
	if err != nil {
		return fmt.Errorf("Fail to read Ini file: %v, %w", err, ErrConfigFailure)
	}
	if !cfg.HasSection(section) {
		sct, err := cfg.NewSection(section)
		if err != nil {
			return err
		}
		sct.Key("name").SetValue(role)
		cfg.SaveTo(ConfigIniFile(""))
	}

	return nil
}

func GetAllIniSections() ([]string, error) {
	sections := []string{}
	cfg, err := ini.Load(ConfigIniFile(""))
	if err != nil {
		return nil, err
	}
	for _, v := range cfg.Section(INI_CONF_SECTION).ChildSections() {
		sections = append(sections, strings.Replace(v.Name(), fmt.Sprintf("%s.", INI_CONF_SECTION), "", -1))
	}
	return sections, nil
}
