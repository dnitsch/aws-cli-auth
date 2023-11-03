package credentialexchange

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"time"

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

func InsertRoleIntoChain(role string, roleChain []string) []string {
	// IF role is provided it can be assumed from the WEB_ID credentials
	// this is to maintain the old implementation
	if role != "" {
		// could use this experimental slice Insert
		// https://pkg.go.dev/golang.org/x/exp/slices#Insert
		roleChain = append(roleChain[:0], append([]string{role}, roleChain[0:]...)...)
	}
	return roleChain
}

func SetCredentials(creds *AWSCredentials, config CredentialConfig) error {
	if config.BaseConfig.StoreInProfile {
		if err := storeCredentialsInProfile(*creds, config.BaseConfig.CfgSectionName); err != nil {
			return err
		}
		return nil
	}
	return returnStdOutAsJson(*creds)
}

func storeCredentialsInProfile(creds AWSCredentials, configSection string) error {
	basePath := path.Join(HomeDir(), ".aws")
	awsConfPath := path.Join(basePath, "credentials")

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		os.Mkdir(basePath, 0755)
		os.WriteFile(awsConfPath, []byte(``), 0755)
	}

	if overriddenpath, exists := os.LookupEnv("AWS_SHARED_CREDENTIALS_FILE"); exists {
		awsConfPath = overriddenpath
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
	fmt.Fprint(os.Stdout, string(jsonBytes))
	return nil
}

func GetWebIdTokenFileContents() (string, error) {
	// var content *string
	file, exists := os.LookupEnv(WEB_ID_TOKEN_VAR)
	if !exists {
		return "", fmt.Errorf("fileNotPresent: %s, %w", WEB_ID_TOKEN_VAR, ErrMissingEnvVar)
	}
	content, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// ReloadBeforeExpiry returns true if the time
// to expiry is less than the specified time in seconds
// false if there is more than required time in seconds
// before needing to recycle credentials
func ReloadBeforeExpiry(expiry time.Time, reloadBeforeSeconds int) bool {
	now := time.Now().Local()
	diff := expiry.Local().Sub(now)
	return diff.Seconds() < float64(reloadBeforeSeconds)
}

// WriteIniSection update ini sections in own config file
func WriteIniSection(role string) error {
	section := fmt.Sprintf("%s.%s", INI_CONF_SECTION, RoleKeyConverter(role))
	cfg, err := ini.Load(ConfigIniFile(""))
	if err != nil {
		return fmt.Errorf("fail to read Ini file: %v, %w", err, ErrConfigFailure)
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
