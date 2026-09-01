package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const DefaultBaseURL = "https://api.brightlocal.com"

// Client makes authenticated requests to the BrightLocal Management API.
type Client struct {
	apiKey     string
	baseURL    *url.URL
	httpClient *http.Client
}

func New(apiKey, baseURL string, httpClient *http.Client) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("BrightLocal API key is required: set BRIGHTLOCAL_API_KEY or pass --api-key")
	}

	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil || parsedBaseURL.Scheme == "" || parsedBaseURL.Host == "" {
		return nil, fmt.Errorf("invalid base URL %q", baseURL)
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{apiKey: apiKey, baseURL: parsedBaseURL, httpClient: httpClient}, nil
}

func (c *Client) Get(ctx context.Context, path string, query url.Values) ([]byte, error) {
	path, err := managementPath(path)
	if err != nil {
		return nil, err
	}

	escapedPath := strings.TrimRight(c.baseURL.EscapedPath(), "/") + path
	decodedPath, err := url.PathUnescape(escapedPath)
	if err != nil {
		return nil, fmt.Errorf("invalid API path: %w", err)
	}
	u := *c.baseURL
	u.Path = decodedPath
	u.RawPath = escapedPath
	u.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request BrightLocal API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read BrightLocal API response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("BrightLocal API returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	return body, nil
}

// managementPath rejects paths that could escape the Management API namespace
// when converted into an HTTP request. Every command is intentionally GET-only,
// but limiting the namespace also prevents access to legacy API endpoints.
func managementPath(path string) (string, error) {
	parsed, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("invalid API path: %w", err)
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("API path must not include a query string or fragment")
	}
	if !strings.HasPrefix(parsed.Path, "/manage/v1/") {
		return "", fmt.Errorf("API path must start with /manage/v1/")
	}
	for _, segment := range strings.Split(parsed.Path, "/") {
		if unsafePathSegment(segment) {
			return "", fmt.Errorf("API path must not contain traversal segments")
		}
	}

	return parsed.EscapedPath(), nil
}

func unsafePathSegment(segment string) bool {
	for {
		if segment == "." || segment == ".." || strings.Contains(segment, `\`) {
			return true
		}
		decoded, err := url.PathUnescape(segment)
		if err != nil || decoded == segment {
			return false
		}
		segment = decoded
	}
}
