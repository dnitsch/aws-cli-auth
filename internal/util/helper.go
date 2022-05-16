package util

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/dnitsch/aws-cli-auth/internal/config"
	ini "gopkg.in/ini.v1"
)

func HomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("unable to get the user home dir")
	}
	return home
}

func ConfigIniFile() string {
	return path.Join(HomeDir(), fmt.Sprintf(".%s.ini", config.SELF_NAME))
}

func WriteDataDir(datadir string) {
	os.MkdirAll(datadir, 0755)
}

func SessionName(username, self_name string) string {
	return fmt.Sprintf("%s-%s", username, self_name)
}

func SetCredentials(creds *AWSCredentials, config config.SamlConfig) {

	if config.BaseConfig.StoreInProfile {
		if err := storeCredentialsInProfile(*creds, config.BaseConfig.CfgSectionName); err != nil {
			Traceln("Error: %s", err.Error())
		}
		return
	}
	returnStdOutAsJson(*creds)
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
		Writeln("Fail to read file: %v", err)
		Exit(err)
	}
	cfg.Section(configSection).Key("aws_access_key_id").SetValue(creds.AWSAccessKey)
	cfg.Section(configSection).Key("aws_secret_access_key").SetValue(creds.AWSSecretKey)
	cfg.Section(configSection).Key("aws_session_token").SetValue(creds.AWSSessionToken)
	cfg.SaveTo(awsConfPath)

	return nil
}

func returnStdOutAsJson(creds AWSCredentials) {
	creds.Version = 1

	jsonBytes, err := json.Marshal(creds)
	if err != nil {
		Writeln("Unexpected AWS credential response")
		Exit(err)
	}
	fmt.Println(string(jsonBytes))
}

func GetWebIdTokenFileContents() (string, error) {
	// var content *string
	file, exists := os.LookupEnv(config.WEB_ID_TOKEN_VAR)
	if !exists {
		Exit(fmt.Errorf("FileNotPresent: %s", config.WEB_ID_TOKEN_VAR))
	}
	content, err := os.ReadFile(file)
	if err != nil {
		Exit(err)
	}
	return string(content), nil
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

func WriteIniSection(role string) error {
	section := fmt.Sprintf("%s.%s", config.INI_CONF_SECTION, RoleKeyConverter(role))
	cfg, err := ini.Load(ConfigIniFile())
	if err != nil {
		Writeln("Fail to read Ini file: %v", err)
		Exit(err)
	}
	if !cfg.HasSection(section) {
		sct, err := cfg.NewSection(section)
		if err != nil {
			return err
		}
		sct.Key("name").SetValue(role)
		cfg.SaveTo(ConfigIniFile())
	}

	return nil
}

func GetAllIniSections() ([]string, error) {
	sections := []string{}
	cfg, err := ini.Load(ConfigIniFile())
	if err != nil {
		return nil, err
	}
	for _, v := range cfg.Section(config.INI_CONF_SECTION).ChildSections() {
		sections = append(sections, strings.Replace(v.Name(), fmt.Sprintf("%s.", config.INI_CONF_SECTION), "", -1))
	}
	return sections, nil
}

// CleanExit signals 0 exit code and should clean up any current process
func CleanExit() {
	os.Exit(0)
}
