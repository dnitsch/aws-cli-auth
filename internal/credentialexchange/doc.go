// credentialexchange
//
// Handles all the main flows for exchanging credentials for AWS temporary creds.
//
// Currently supports SAML as posted by an IdP to an ACS endpoint in AWS
// AWS_WEB_IDENTITY_TOKEN_FILE and optionally can specify the exact role to choose,
//
// if the TOKEN corresponds to the `chained role`.
package credentialexchange
