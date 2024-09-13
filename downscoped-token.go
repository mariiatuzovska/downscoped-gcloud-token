package downscoped

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
)

// Policy ...
type Policy struct {
	AccessBoundary AccessBoundary `json:"accessBoundary"`
}

// AccessBoundary ...
type AccessBoundary struct {
	AccessBoundaryRules []AccessBoundaryRule `json:"accessBoundaryRules"`
}

// AccessBoundaryRule ...
type AccessBoundaryRule struct {
	AvailableResource     string    `json:"availableResource"`
	AvailablePermissions  []string  `json:"availablePermissions"`
	AvailabilityCondition Condition `json:"availabilityCondition"`
}

// Condition ...
type Condition struct {
	Title      string `json:"title"`
	Expression string `json:"expression"`
}

// NewDownscopedToken creates a new downscoped token using the provided access boundary.
func NewDownscopedToken(ctx context.Context, policyRules []AccessBoundaryRule) (*oauth2.Token, error) {
	rootToken, err := getAccessTokenFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("getAccessTokenFromContext: %v", err)
	}

	// Exchange the root token for a downscoped token
	downscopedToken, err := ExchangeAccessTokenForDownscopedToken(ctx, rootToken, &Policy{
		AccessBoundary: AccessBoundary{
			AccessBoundaryRules: policyRules,
		},
	})
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken: downscopedToken,
		TokenType:   "Bearer",
	}, nil
}

// getAccessTokenFromContext retrieves an access token from the context using the Google credentials.
func getAccessTokenFromContext(ctx context.Context) (string, error) {
	creds, err := transport.Creds(ctx, option.WithTokenSource(nil))
	if err != nil {
		return "", fmt.Errorf("transport.Creds: %v", err)
	}
	token, err := creds.TokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("creds.TokenSource.Token: %v", err)
	}
	return token.AccessToken, nil
}

// ExchangeAccessTokenForDownscopedToken exchanges an access token for a downscoped token using the STS API.
func ExchangeAccessTokenForDownscopedToken(ctx context.Context, accessToken string, policy *Policy) (string, error) {
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal access boundary: %v", err)
	}

	requestBody := fmt.Sprintf(
		"grant_type=urn:ietf:params:oauth:grant-type:token-exchange&subject_token_type=urn:ietf:params:oauth:token-type:access_token&requested_token_type=urn:ietf:params:oauth:token-type:access_token&subject_token=%s&options=%s",
		accessToken, policyJSON,
	)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://sts.googleapis.com/v1/token", bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return "", fmt.Errorf("http.NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("client.Do: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read STS response: %v", err)
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return "", fmt.Errorf("failed to parse STS response: %v", err)
	}

	return tokenResponse.AccessToken, nil
}
