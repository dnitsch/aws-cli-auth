package util

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/dnitsch/aws-cli-auth/internal/config"
	ini "gopkg.in/ini.v1"
)

func GetHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("unable to get the user home dir")
	}
	return home
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
			fmt.Printf("Error: %s", err.Error())
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
		awsCredsPath := path.Join(GetHomeDir(), ".aws", "credentials")
		if _, err := os.Stat(awsCredsPath); os.IsNotExist(err) {
			os.Mkdir(awsCredsPath, 0655)
		}
		awsConfPath = awsCredsPath
	}

	cfg, err := ini.Load(awsConfPath)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
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
