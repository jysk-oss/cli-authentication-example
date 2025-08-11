package cliAuth

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/int128/oauth2cli"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
)

var azureTenantID = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxx"
var azureAppRegClientID = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxx"

var azureTokenEndpoint = fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", azureTenantID)
var azureAuthUrl = fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize", azureTenantID)

// The following `transport` struct and its implemented `RoundTrip` function provide a workaround
// necessary when working with Azure Entra ID. To enable the authorization flow with PKCE,
// you must configure the Azure App Registration as a "Single Page Application" (SPA) variant.
// During the token retrieval step of the authentication flow, Azure requires the presence of an
// "Origin" header. The value of this header is irrelevant; it just needs to be included.
type transport struct{}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Origin", "does-not-matter-but-is-required")
	return http.DefaultTransport.RoundTrip(req)
}

// Start the login flow after loading the TokenStore from tokens.json.
func Login() {
	tokenStore, err := LoadTokenStore()
	if err != nil {
		log.Fatal(err)
	}

	login(tokenStore)
}

// Start login flow using a specific TokenStore
func login(tokenStore *TokenStore) {
	pkceVerifier := oauth2.GenerateVerifier()
	ready := make(chan string, 1)
	defer close(ready)

	cfg := oauth2cli.Config{
		OAuth2Config: oauth2.Config{
			ClientID: azureAppRegClientID,
			Endpoint: oauth2.Endpoint{
				AuthURL:  azureAuthUrl,
				TokenURL: azureTokenEndpoint,
			},
			Scopes: []string{"openid"},
		},
		AuthCodeOptions:      []oauth2.AuthCodeOption{oauth2.S256ChallengeOption(pkceVerifier)},
		TokenRequestOptions:  []oauth2.AuthCodeOption{oauth2.VerifierOption(pkceVerifier)},
		LocalServerReadyChan: ready,
	}

	// These two lines are a part of the Azure Entra ID workaround.
	// Creating a httpClient with the transport struct and adds it to the ctx variable.
	httpClient := &http.Client{Transport: &transport{}}
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, httpClient)

	errorGroup, ctx := errgroup.WithContext(ctx)
	errorGroup.Go(func() error {
		select {
		case url := <-ready:
			browser.Stdout = browser.Stderr
			if err := browser.OpenURL(url); err != nil {
				log.Printf("could not open the browser: %s", err)
			}
			return nil
		case <-ctx.Done():
			return fmt.Errorf("context done while waiting for authorization: %w", ctx.Err())
		}
	})

	errorGroup.Go(func() error {
		token, err := oauth2cli.GetToken(ctx, cfg)
		if err != nil {
			return fmt.Errorf("could not get a token: %w", err)
		}

		tokenStore.upsertPrimaryToken(token)
		tokenStore.save()

		return nil
	})

	if err := errorGroup.Wait(); err != nil {
		log.Fatalf("authorization error: %s", err)
	}
}
