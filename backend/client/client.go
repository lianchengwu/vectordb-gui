package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	username   string
	apiKey     string
	httpClient *http.Client
}

func NewClient(rawURL, username, apiKey string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	u := strings.TrimRight(rawURL, "/")
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		u = "http://" + u
	}
	return &Client{
		baseURL:  u,
		username: username,
		apiKey:   apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) doRequest(ctx context.Context, path string, reqBody, respBody interface{}) error {
	var bodyReader io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	fullURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	authHeader := fmt.Sprintf("Bearer account=%s&api_key=%s", c.username, c.apiKey)
	req.Header.Set("Authorization", authHeader)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var header ResponseHeader
		if jsonErr := json.Unmarshal(respData, &header); jsonErr == nil && header.Message != "" {
			return fmt.Errorf("vectordb error (%d): %s", header.Code, header.Message)
		}
		return fmt.Errorf("http error %d: %s", resp.StatusCode, string(respData))
	}

	if respBody != nil {
		if err := json.Unmarshal(respData, respBody); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}
	return nil
}

func (c *Client) Ping(ctx context.Context) (bool, string, error) {
	dbs, err := c.ListDatabases(ctx)
	if err != nil {
		return false, err.Error(), err
	}
	return true, fmt.Sprintf("Connected successfully. Found %d databases.", len(dbs)), nil
}

func (c *Client) ListDatabases(ctx context.Context) ([]string, error) {
	var resp ListDatabasesResponse
	if err := c.doRequest(ctx, "/database/list", map[string]interface{}{}, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("api error (%d): %s", resp.Code, resp.Message)
	}
	result := make([]string, 0, len(resp.Databases))
	for _, db := range resp.Databases {
		result = append(result, db.Database)
	}
	return result, nil
}

func (c *Client) ListCollections(ctx context.Context, database string) ([]string, error) {
	req := ListCollectionsRequest{Database: database}
	var resp ListCollectionsResponse
	if err := c.doRequest(ctx, "/collection/list", req, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("api error (%d): %s", resp.Code, resp.Message)
	}
	result := make([]string, 0, len(resp.Collections))
	for _, coll := range resp.Collections {
		result = append(result, coll.Collection)
	}
	return result, nil
}

func (c *Client) DescribeCollection(ctx context.Context, database, collection string) (*CollectionMeta, error) {
	req := DescribeCollectionRequest{Database: database, Collection: collection}
	var resp DescribeCollectionResponse
	if err := c.doRequest(ctx, "/collection/describe", req, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("api error (%d): %s", resp.Code, resp.Message)
	}
	return &resp.Collection, nil
}

func (c *Client) QueryDocuments(ctx context.Context, database, collection string, limit, offset int, filter string) (*QueryDocumentResponse, error) {
	req := QueryDocumentRequest{
		Database:   database,
		Collection: collection,
	}
	if limit <= 0 {
		limit = 20
	}
	req.Query.Limit = limit
	req.Query.Offset = offset
	req.Query.Filter = filter

	var resp QueryDocumentResponse
	if err := c.doRequest(ctx, "/document/query", req, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("api error (%d): %s", resp.Code, resp.Message)
	}
	return &resp, nil
}
