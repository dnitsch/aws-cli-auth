package auth

// import (
// 	"context"
// 	"encoding/base64"
// 	"fmt"
// 	"net"
// 	"net/http"
// 	"net/url"
// 	"strings"
// 	"time"

// 	"github.com/beevik/etree"
// 	"github.com/dnitsch/aws-cli-auth/lib"
// 	pkce "github.com/nirasan/go-oauth-pkce-code-verifier"
// 	"github.com/pkg/browser"
// 	"github.com/pkg/errors"
// 	"github.com/spf13/viper"
// )

// func getSAMLAssertion(client *lib.OIDCClient, TokenResponse *lib.TokenResponse) (string, error) {
// 	audience := client.config.GetString(lib.OIDC_PROVIDER_TOKEN_EXCHANGE_AUDIENCE)
// 	subjectTokenType := client.config.GetString(lib.OIDC_PROVIDER_TOKEN_EXCHANGE_SUBJECT_TOKEN_TYPE)

// 	var subjectToken string
// 	if subjectTokenType == TOKEN_TYPE_ID_TOKEN {
// 		subjectToken = lib.TokenResponse.IDToken
// 	} else if subjectTokenType == TOKEN_TYPE_ACCESS_TOKEN {
// 		subjectToken = lib.TokenResponse.AccessToken
// 	}

// 	form := client.ClientForm()
// 	form.Set("audience", audience)
// 	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
// 	form.Set("subject_token", subjectToken)
// 	form.Set("subject_token_type", subjectTokenType)
// 	form.Set("requested_token_type", "urn:ietf:params:oauth:token-type:saml2")

// 	res, err := client.Token().
// 		Request().
// 		Form(form).
// 		Post()

// 	Traceln("Exchanged SAML assertion response status: %d", res.Status())

// 	if res.Status() != 200 {
// 		if res.MediaType() != "" {
// 			var json map[string]interface{}
// 			err := res.ReadJson(&json)
// 			if err == nil {
// 				return "", errors.Errorf("Failed to exchange saml2 token, error: %s error_description: %s",
// 					json["error"], json["error_description"])
// 			}
// 		}
// 		return "", errors.Errorf("Failed to exchange saml2 token, statusCode: %d", res.Status())
// 	}

// 	var saml2TokenResponse *lib.TokenResponse
// 	err = res.ReadJson(&saml2TokenResponse)
// 	if err != nil {
// 		return "", errors.Wrap(err, "Failed to parse token exchange response")
// 	}

// 	Traceln("SAML2 Assertion: %s", saml2lib.TokenResponse.AccessToken)

// 	// TODO: Validation
// 	return saml2lib.TokenResponse.AccessToken, nil
// }

// func createSAMLResponse(client *lib.OIDCClient, samlAssertion string) (string, error) {
// 	s, err := base64.RawURLEncoding.DecodeString(samlAssertion)
// 	if err != nil {
// 		return "", errors.Wrap(err, "Failed to decode SAML2 assertion")
// 	}

// 	doc := etree.NewDocument()
// 	if err := doc.ReadFromBytes(s); err != nil {
// 		return "", errors.Wrap(err, "Parse error")
// 	}

// 	assertionElement := doc.FindElement(".//Assertion")
// 	if assertionElement == nil {
// 		return "", errors.New("No Assertion element")
// 	}

// 	issuerElement := assertionElement.FindElement("./Issuer")
// 	if issuerElement == nil {
// 		return "", errors.New("No Issuer element")
// 	}

// 	subjectConfirmationDataElement := doc.FindElement(".//SubjectConfirmationData")
// 	if subjectConfirmationDataElement == nil {
// 		return "", errors.New("No SubjectConfirmationData element")
// 	}

// 	recipient := subjectConfirmationDataElement.SelectAttr("Recipient")
// 	if recipient == nil {
// 		return "", errors.New("No Recipient attribute")
// 	}

// 	issueInstant := assertionElement.SelectAttr("IssueInstant")
// 	if issueInstant == nil {
// 		return "", errors.New("No IssueInstant attribute")
// 	}

// 	newDoc := etree.NewDocument()

// 	samlp := newDoc.CreateElement("samlp:Response")
// 	samlp.CreateAttr("xmlns:samlp", "urn:oasis:names:tc:SAML:2.0:protocol")
// 	if assertionElement.Space != "" {
// 		samlp.CreateAttr("xmlns:"+assertionElement.Space, "urn:oasis:names:tc:SAML:2.0:assertion")
// 	}
// 	samlp.CreateAttr("Destination", recipient.Value)
// 	// samlp.CreateAttr("ID", "ID_760649d5-ebe0-4d8a-a107-4a16dd3e9ecd")
// 	samlp.CreateAttr("Version", "2.0")
// 	samlp.CreateAttr("IssueInstant", issueInstant.Value)
// 	samlp.AddChild(issuerElement.Copy())

// 	status := samlp.CreateElement("samlp:Status")
// 	statusCode := status.CreateElement("samlp:StatusCode")
// 	statusCode.CreateAttr("Value", "urn:oasis:names:tc:SAML:2.0:status:Success")
// 	assertionElement.RemoveAttr("xmlns:saml")
// 	samlp.AddChild(assertionElement)

// 	// newDoc.WriteTo(os.Stderr)

// 	samlResponse, err := newDoc.WriteToString()

// 	return samlResponse, nil
// }

// func DoSamlLogin() (*lib.TokenResponse, error) {

// 	url, err := url.Parse("https://fedhub.iairgroup.com/idp/startSSO.ping?PARTNER=urn:amazon:webservices")
// 	if err != nil {
// 		return nil, err
// 	}

// 	rc, err := lib.NewRestClient(&lib.RestClientConfig{
// 		ClientCert:         "",
// 		ClientKey:          "",
// 		ClientCA:           "",
// 		InsecureSkipVerify: false,
// 	})

// 	if err != nil {
// 		return nil, err
// 	}
// 	conf := &viper.Viper{}
// 	metadata := &lib.OIDCMetadataResponse{}

// 	c := &lib.OIDCClient{
// 		restClient: rc,
// 		base:       &lib.WebTarget{url: *url, client: rc},
// 		config:     conf,
// 		metadata:   metadata,
// 	}
// 	return doLogin(c)
// }

// func doLogin(client *lib.OIDCClient) (*lib.TokenResponse, error) {
// 	listener, err := net.Listen("tcp", "127.0.0.1:")
// 	if err != nil {
// 		return nil, errors.Wrap(err, "Cannot start local http server to handle login redirect")
// 	}
// 	port := listener.Addr().(*net.TCPAddr).Port

// 	clientId := client.config.GetString(CLIENT_ID)
// 	redirect := fmt.Sprintf("http://127.0.0.1:%d", port)
// 	v, err := pkce.CreateCodeVerifierWithLength(pkce.MaxLength)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "Cannot generate OAuth2 PKCE code_challenge")
// 	}
// 	challenge := v.CodeChallengeS256()
// 	verifier := v.String()

// 	authReq := client.Authorization().
// 		QueryParam("response_type", "code").
// 		QueryParam("client_id", clientId).
// 		QueryParam("redirect_uri", redirect).
// 		QueryParam("code_challenge", challenge).
// 		QueryParam("code_challenge_method", "S256").
// 		QueryParam("scope", "openid")

// 	additionalQuery := client.config.GetString(OIDC_AUTHENTICATION_REQUEST_ADDITIONAL_QUERY)
// 	if additionalQuery != "" {
// 		queries := strings.Split(additionalQuery, "&")
// 		for _, q := range queries {
// 			kv := strings.Split(q, "=")
// 			if len(kv) == 1 {
// 				authReq = authReq.QueryParam(kv[0], "")
// 			} else if len(kv) == 2 {
// 				authReq = authReq.QueryParam(kv[0], kv[1])
// 			} else {
// 				return nil, errors.Errorf("Invalid additional query: %s", q)
// 			}
// 		}
// 	}
// 	url := authReq.Url()

// 	code := launch(client, url.String(), listener)
// 	if code != "" {
// 		return codeToToken(client, verifier, code, redirect)
// 	} else {
// 		return nil, errors.New("Login failed, can't retrieve authorization code")
// 	}
// }

// func launch(client *lib.OIDCClient, url string, listener net.Listener) string {
// 	c := make(chan string)

// 	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
// 		url := req.URL
// 		q := url.Query()
// 		code := q.Get("code")

// 		res.Header().Set("Content-Type", "text/html")

// 		// Redirect to user-defined successful/failure page
// 		successful := client.RedirectToSuccessfulPage()
// 		if successful != nil && code != "" {
// 			url := successful.Url()
// 			res.Header().Set("Location", (&url).String())
// 			res.WriteHeader(302)
// 		}
// 		failure := client.RedirectToFailurePage()
// 		if failure != nil && code == "" {
// 			url := failure.Url()
// 			res.Header().Set("Location", (&url).String())
// 			res.WriteHeader(302)
// 		}

// 		// Response result page
// 		message := "Login "
// 		if code != "" {
// 			message += "successful"
// 		} else {
// 			message += "failed"
// 		}
// 		res.Header().Set("Cache-Control", "no-store")
// 		res.Header().Set("Pragma", "no-cache")
// 		res.WriteHeader(200)
// 		res.Write([]byte(fmt.Sprintf(`<!DOCTYPE html>
// <body>
// %s
// </body>
// </html>
// `, message)))

// 		if f, ok := res.(http.Flusher); ok {
// 			f.Flush()
// 		}

// 		time.Sleep(100 * time.Millisecond)

// 		c <- code
// 	})

// 	srv := &http.Server{}
// 	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer srv.Shutdown(ctx)

// 	go func() {
// 		if err := srv.Serve(listener); err != nil {
// 			// cannot panic, because this probably is an intentional close
// 		}
// 	}()

// 	var code string
// 	if err := browser.OpenURL(url); err == nil {
// 		code = <-c
// 	}

// 	return code
// }

// func GetFreePort() (int, error) {
// 	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
// 	if err != nil {
// 		return 0, err
// 	}

// 	l, err := net.ListenTCP("tcp", addr)
// 	if err != nil {
// 		return 0, err
// 	}
// 	defer l.Close()
// 	return l.Addr().(*net.TCPAddr).Port, nil
// }

// func codeToToken(client *lib.OIDCClient, verifier string, code string, redirect string) (*lib.TokenResponse, error) {
// 	form := client.ClientForm()
// 	form.Set("grant_type", "authorization_code")
// 	form.Set("code", code)
// 	form.Set("code_verifier", verifier)
// 	form.Set("redirect_uri", redirect)

// 	Traceln("code2token params:", form)

// 	res, err := client.Token().Request().Form(form).Post()

// 	if err != nil {
// 		return nil, errors.Wrap(err, "Failed to turn code into token")
// 	}

// 	if res.Status() != 200 {
// 		if res.MediaType() != "" {
// 			var json map[string]interface{}
// 			err := res.ReadJson(&json)
// 			if err == nil {
// 				return nil, errors.Errorf("Failed to turn code into token, error: %s error_description: %s",
// 					json["error"], json["error_description"])
// 			}
// 		}
// 		return nil, errors.Errorf("Failed to turn code into token")
// 	}

// 	var tokenResponse lib.TokenResponse
// 	res.ReadJson(&tokenResponse)
// 	return &tokenResponse, nil
// }
