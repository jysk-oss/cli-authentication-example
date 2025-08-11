package cliAuth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"golang.org/x/oauth2"
)

// ####################################################### //
//                    TokenStore Struct                    //
// ####################################################### //

type TokenStore struct {
	PrimaryAccessToken    AccessToken            `json:"primary"`
	DelegatedAccessTokens map[string]AccessToken `json:"delegated_access_tokens"`
}

type AccessToken struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
	ExpiresIn    int64     `json:"expires_in"`
}

// Update or insert a new primary access token into the TokenStore.
func (t *TokenStore) upsertPrimaryToken(token *oauth2.Token) {
	t.PrimaryAccessToken = AccessToken{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
		ExpiresIn:    token.ExpiresIn,
	}
}

// ####################################################### //
//                TokenStore Impementations                //
// ####################################################### //

// Add a new delegated token with an identifier to the TokenStore.
func (t *TokenStore) addDelegatedToken(identifier string, token *oauth2.Token) {
	if t.DelegatedAccessTokens == nil {
		t.DelegatedAccessTokens = make(map[string]AccessToken)
	}

	t.DelegatedAccessTokens[identifier] = AccessToken{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
		ExpiresIn:    int64(token.ExpiresIn),
	}
}

// Save the TokenStore to the tokens.json file in the config path.
func (t *TokenStore) save() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(t, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// Get a delegated token from the TokenStore the token will be refreshed if needed.
// If no token is found, it will request a new deletaged access token.
func (t *TokenStore) GetDelegatedToken(identifier string, scope string) (*oauth2.Token, error) {
	delegatedToken, tokenFound := t.DelegatedAccessTokens[identifier]

	if tokenFound {
		token, err := delegatedToken.refreshToken()
		if err == nil {
			return token, nil
		}
	}

	token, err := t.requestNewDelegatedToken(scope)
	if err != nil {
		return nil, err
	}

	t.addDelegatedToken(identifier, token)
	t.save()

	return token, nil
}

// Request a new delegated token with a specific scope.
func (t *TokenStore) requestNewDelegatedToken(scope string) (*oauth2.Token, error) {
	token, err := t.PrimaryAccessToken.refreshToken()
	if err != nil {
		login(t)
		token = t.PrimaryAccessToken.intoOAuth2Token()
	}

	values := url.Values{}
	values.Set("grant_type", "refresh_token")
	values.Set("refresh_token", token.RefreshToken)
	values.Set("scope", scope)
	values.Set("client_id", azureAppRegClientID)

	req, err := http.NewRequest("POST", azureTokenEndpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error retriving access token. %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "does-not-matter-but-is-required")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error retriving access token. %s", resp.Status)
	}

	var delegatedToken oauth2.Token
	if err := json.NewDecoder(resp.Body).Decode(&delegatedToken); err != nil {
		return nil, err
	}
	if delegatedToken.Expiry.IsZero() {
		delegatedToken.Expiry = time.Now().Add(time.Hour) // fallback
	}
	return &delegatedToken, nil
}

// ####################################################### //
//               AccessToken Impementations                //
// ####################################################### //

// Converts the AccessToken struct into an *oauth2.Token
func (a AccessToken) intoOAuth2Token() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  a.AccessToken,
		TokenType:    a.TokenType,
		RefreshToken: a.RefreshToken,
		Expiry:       a.Expiry,
		ExpiresIn:    int64(a.ExpiresIn),
	}
}

// Uses the refresh token to request a new access token only if the current token is expired or near expiry.
// If the refresh attempt fails (e.g., invalid or expired refresh token), it returns an error.
func (a AccessToken) refreshToken() (*oauth2.Token, error) {
	conf := &oauth2.Config{
		ClientID: azureAppRegClientID,
		Endpoint: oauth2.Endpoint{
			TokenURL: azureTokenEndpoint,
		},
	}
	ctx := context.Background()
	tokenSource := conf.TokenSource(ctx, a.intoOAuth2Token())

	return tokenSource.Token()
}

// ####################################################### //
//                    Just functions...                    //
// ####################################################### //

// Load the TokenStore from the tokens.json file.
func LoadTokenStore() (*TokenStore, error) {
	configPath, err := configPath()
	if err != nil {
		return nil, err
	}

	var tokenStore = TokenStore{
		DelegatedAccessTokens: make(map[string]AccessToken),
	}

	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("load token file: %w", err)
		}
		if err := json.Unmarshal(data, &tokenStore); err != nil {
			return nil, err
		}
	}

	return &tokenStore, nil
}

// Returns the path of the tokens.json file.
func configPath() (string, error) {
	return xdg.ConfigFile("cli-example/tokens.json")
}
