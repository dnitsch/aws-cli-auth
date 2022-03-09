package config

const SELF_NAME = "aws-cli-auth"

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
