package config

const SELF_NAME = "aws-cli-auth"
const WEB_ID_TOKEN_VAR = "AWS_WEB_IDENTITY_TOKEN_FILE"
const AWS_ROLE_ARN = "AWS_ROLE_ARN"

type BaseConfig struct {
	Role                 string
	CfgSectionName       string
	StoreInProfile       bool
	DoKillHangingProcess bool
}

type SamlConfig struct {
	BaseConfig   BaseConfig
	ProviderUrl  string
	PrincipalArn string
	AcsUrl       string
	Duration     int
}
