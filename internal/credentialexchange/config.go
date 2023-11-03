package credentialexchange

const (
	SELF_NAME        = "aws-cli-auth"
	WEB_ID_TOKEN_VAR = "AWS_WEB_IDENTITY_TOKEN_FILE"
	AWS_ROLE_ARN     = "AWS_ROLE_ARN"
	INI_CONF_SECTION = "role"
)

type BaseConfig struct {
	Role                 string
	RoleChain            []string
	Username             string
	CfgSectionName       string
	StoreInProfile       bool
	DoKillHangingProcess bool
	ReloadBeforeTime     int
}

type CredentialConfig struct {
	BaseConfig         BaseConfig
	ProviderUrl        string
	PrincipalArn       string
	AcsUrl             string
	Duration           int
	IsSso              bool
	SsoRegion          string
	SsoRole            string
	SsoUserEndpoint    string
	SsoCredFedEndpoint string
}
