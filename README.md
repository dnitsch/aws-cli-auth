# aws-cli-auth

CLI tool for retrieving AWS temporary credentials using OIDC or SAML providers.

Firstly, this package currently deals with SAML only (OIDC to come), however if you have an OIDC IdP provider set up to AWS you can use this [package](https://github.com/openstandia/aws-cli-oidc) and likewise this [package](https://github.com/Versent/saml2aws) for standard SAML AWS integrations.

If, however, you need to support a non standard user journeys enforced by your IdP i.e. a sub company selection within your organization portal, or a selection screen for different MFA providers - PingID or RSA HardToken etc.... you cannot reliably automate the flow or it would have to be too specific.

As such this approach uses [go-rod](https://github.com/go-rod/rod) library to uniformly allow the user to complete any and all auth steps and selections in a managed browser session up to the point of where the SAMLResponse were to be sent to AWS ACS service `https://signin.aws.amazon.com/saml`. Capturing this via hijack request and posting to AWS STS service to exchange this for the temporary credentials.

The advantage of using SAML is that real users can gain access to the AWS Console UI or programatically and audited as the same person in cloudtrail. 

By default the tool creates the session name - which can be audited including the persons username from the localhost.

## Known Issues

- Even though a datadir is created to store the chromium session data it is advised to still open settings and save the username/password manually the first time you are presented with the login screen.

## Install

Download from [Releases page](https://github.com/dnitsch/aws-cli-auth/releases).

MacOS

```bash
curl -L https://github.com/dnitsch/aws-cli-auth/releases/download/v0.1.0/aws-cli-auth-darwin-amd64 -o aws-cli-auth
chmod +x aws-cli-auth
sudo mv aws-cli-auth /usr/local/bin
```

## Usage

```bash
CLI tool for retrieving AWS temporary credentials using OIDC or SAML providers. 
Stores them under the $HOME/.aws/credentials file under a specified path

Usage:
  aws-cli-auth [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  saml        Get AWS credentials and out to stdout

Flags:
      --cfg-section string   config section name in the yaml config file
  -h, --help                 help for aws-cli-auth
  -r, --role string          Set the role you want to assume when SAML or OIDC process completes
  -s, --store-profile        By default the credentials are returned to stdout to be used by the credential_process

Use "aws-cli-auth [command] --help" for more information about a command.
```

### SAML 



```bash
Get AWS credentials and out to stdout through your SAML provider authentication.

Usage:
  aws-cli-auth saml <SAML ProviderUrl> [flags]

Flags:
  -a, --acsurl string      Override the default ACS Url, used for checkin the post of the SAMLResponse (default "https://signin.aws.amazon.com/saml")
  -h, --help               help for saml
  -d, --max-duration int   Override default max session duration, in seconds, of the role session [900-43200] (default 900)
      --principal string   Principal Arn of the SAML IdP in AWS
  -p, --provider string    Saml Entity StartSSO Url

Global Flags:
      --cfg-section string   config section name in the yaml config file
  -r, --role string          Set the role you want to assume when SAML or OIDC process completes
  -s, --store-profile        By default the credentials are returned to stdout to be used by the credential_process
```

Example:

```bash
aws-cli-auth saml --cfg-section nonprod_saml_admin -p "https://your-idp.com/idp/foo?PARTNER=urn:amazon:webservices" --principal "arn:aws:iam::XXXXXXXXXX:saml-provider/IDP_ENTITY_ID" -r "arn:aws:iam::XXXXXXXXXX:role/Developer" -d 3600 -s
```

The PartnerId in most IdPs is usually `urn:amazon:webservices` - but you can change this for anything you stored it as.

If successful will store the creds under the specified config section in credentials profile as per below example

```ini
[default]
aws_access_key_id     = XXXXX
aws_secret_access_key = YYYYYYYYY

[another_profile]
aws_access_key_id     = XXXXX
aws_secret_access_key = YYYYYYYYY

[nonprod_saml_admin]
aws_access_key_id     = XXXXXX
aws_secret_access_key = YYYYYYYYY
aws_session_token     = ZZZZZZZZZZZZZZZZZZZZ
```

To give it a quick test.

```bash
aws sts get-caller-identity --profile=nonprod_saml_admin
```

<!-- ### Integrate aws-cli

[Sourcing credentials with an external process](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sourcing-external.html) describes how to integrate aws-cli with external tool.
You can use `aws-cli-auth` as the external process. Add the following lines to your `.aws/config` file.

```
[profile ]
credential_process=aws-cli-auth get-cred -p myop -r arn:aws:iam::123456789012:role/developer -j -s -d 43200
```

Caution: The AWS temporary credentials will be saved into your OS secret store by using `-s` option to reduce authentication each time you use `aws-cli` tool.
-->
## Licence
 WFTPL

## Acknowldgements
  - [Hiroyuki Wada](https://github.com/wadahiro) [package](https://github.com/openstandia/aws-cli-oidc) 
  - [Mark Wolfe](https://github.com/wolfeidau) [package](https://github.com/Versent/saml2aws)
