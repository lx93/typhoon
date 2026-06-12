package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"typhoon/internal/relay"
)

const defaultRelayLimit = 5

type BrokerClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func (c BrokerClient) ListRelays(ctx context.Context, limit int) (relay.ListResponse, error) {
	endpoint, err := RelayListURL(c.BaseURL, limit)
	if err != nil {
		return relay.ListResponse{}, err
	}

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return relay.ListResponse{}, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return relay.ListResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return relay.ListResponse{}, brokerStatusError(resp)
	}

	var out relay.ListResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return relay.ListResponse{}, fmt.Errorf("decode relay list: %w", err)
	}
	return out, nil
}

func RelayListURL(baseURL string, limit int) (string, error) {
	if strings.TrimSpace(baseURL) == "" {
		return "", fmt.Errorf("broker URL is required")
	}
	if limit < 1 {
		limit = defaultRelayLimit
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse broker URL: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("broker URL must include scheme and host")
	}

	basePath := strings.Trim(parsed.Path, "/")
	pathParts := []string{"api/v1/relays"}
	if basePath != "" {
		pathParts = append([]string{basePath}, pathParts...)
	}
	parsed.Path = "/" + strings.Join(pathParts, "/")

	query := parsed.Query()
	query.Set("limit", fmt.Sprintf("%d", limit))
	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}

func brokerStatusError(resp *http.Response) error {
	var apiErr relay.ErrorResponse
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	_ = json.Unmarshal(body, &apiErr)
	if apiErr.Error == "" {
		apiErr.Error = strings.TrimSpace(string(body))
	}
	if apiErr.Error == "" {
		apiErr.Error = resp.Status
	}
	return fmt.Errorf("broker list relays: %s", apiErr.Error)
}
