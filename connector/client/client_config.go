package client

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/infracloudio/msbotbuilder-go/connector/auth"
)

// Config represents the credentials for a user program and the URL for validating the credentials.
type Config struct {
	Credentials auth.CredentialProvider
	AuthURL     url.URL
	AuthClient  *http.Client
	ReplyClient *http.Client
}

// NewClientConfig creates configuration for ConnectorClient.
func NewClientConfig(credentials auth.CredentialProvider, tokenURL string) (*Config, error) {

	if len(credentials.GetAppID()) < 0 || len(credentials.GetAppPassword()) < 0 {
		return &Config{}, errors.New("Invalid client credentials")
	}

	parsedURL, err := url.Parse(tokenURL)
	if err != nil {
		return &Config{}, errors.New("Invalid token URL")
	}

	return &Config{
		Credentials: credentials,
		AuthURL:     *parsedURL,
	}, nil
}
