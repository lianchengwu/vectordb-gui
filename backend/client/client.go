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

	"golang.org/x/crypto/ssh"
	"vectordb-1/backend/storage"
)

type Client struct {
	baseURL    string
	username   string
	apiKey     string
	httpClient *http.Client
	sshClients []*ssh.Client
}

func (c *Client) Close() error {
	for _, cli := range c.sshClients {
		if cli != nil {
			_ = cli.Close()
		}
	}
	return nil
}

func NewClientWithConfig(cfg storage.ConnectionConfig, timeout time.Duration) (*Client, error) {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	u := strings.TrimRight(strings.TrimSpace(cfg.URL), "/")
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		u = "http://" + u
	}

	chain := cfg.ProxyChain
	if len(chain) == 0 {
		if cfg.Proxy.Enabled && strings.TrimSpace(cfg.Proxy.Host) != "" {
			chain = append(chain, storage.NetworkHop{
				Enabled:  true,
				Type:     cfg.Proxy.Type,
				Host:     cfg.Proxy.Host,
				Port:     cfg.Proxy.Port,
				Username: cfg.Proxy.Username,
				Password: cfg.Proxy.Password,
			})
		}
		if cfg.SSHTunnel.Enabled && strings.TrimSpace(cfg.SSHTunnel.Host) != "" {
			chain = append(chain, storage.NetworkHop{
				Enabled:    true,
				Type:       "ssh",
				Host:       cfg.SSHTunnel.Host,
				Port:       cfg.SSHTunnel.Port,
				Username:   cfg.SSHTunnel.User,
				AuthType:   cfg.SSHTunnel.AuthType,
				Password:   cfg.SSHTunnel.Password,
				PrivateKey: cfg.SSHTunnel.PrivateKey,
				Passphrase: cfg.SSHTunnel.Passphrase,
			})
		}
	}

	httpCli, sshClis, err := BuildChainedHTTPClient(timeout, chain)
	if err != nil {
		return nil, err
	}

	username := cfg.Username
	if username == "" {
		username = "root"
	}

	return &Client{
		baseURL:    u,
		username:   username,
		apiKey:     cfg.APIKey,
		httpClient: httpCli,
		sshClients: sshClis,
	}, nil
}
func NewClient(rawURL, username, apiKey string, timeout time.Duration) *Client {
	cfg := storage.ConnectionConfig{
		URL:      rawURL,
		Username: username,
		APIKey:   apiKey,
		Timeout:  int(timeout.Seconds()),
	}
	cli, _ := NewClientWithConfig(cfg, timeout)
	return cli
}

func (c *Client) doRequest(ctx context.Context, method, path string, reqBody, respBody interface{}) error {
	var bodyReader io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	fullURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	authHeader := fmt.Sprintf("Bearer account=%s&api_key=%s", c.username, c.apiKey)
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Sdk-Version", "tcvectordb-go-v1.9.1")

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
	return true, fmt.Sprintf("连接成功，发现 %d 个数据库", len(dbs)), nil
}
func isAIDatabase(dbType string) bool {
	t := strings.ToUpper(strings.TrimSpace(dbType))
	return t == "AI" || t == "AI_DB" || t == "AI_DOC" || strings.Contains(t, "AI")
}

func (c *Client) buildAIMeta(database, collection string, cv *AICollectionViewData) *CollectionMeta {
	meta := CollectionMeta{
		Database:       database,
		Collection:     collection,
		Description:    cv.Description,
		Indexes:        cv.Indexes,
		IsAICollection: true,
	}
	if cv.ReplicaNum != nil {
		meta.ReplicaNum = *cv.ReplicaNum
	}
	if cv.ShardNum != nil {
		meta.ShardNum = *cv.ShardNum
	}
	if len(meta.Indexes) > 0 {
		for _, idx := range meta.Indexes {
			if idx.FieldName != "" {
				meta.Fields = append(meta.Fields, FieldMeta{
					FieldName:  idx.FieldName,
					FieldType:  idx.FieldType,
					FieldUsage: idx.IndexType,
				})
			}
		}
	}
	return &meta
}

func (c *Client) ListDatabasesDetailed(ctx context.Context) ([]DatabaseDetail, error) {
	var resp ListDatabasesResponse
	err := c.doRequest(ctx, http.MethodGet, "/database/list", nil, &resp)
	if err != nil {
		errPost := c.doRequest(ctx, http.MethodPost, "/database/list", map[string]interface{}{}, &resp)
		if errPost != nil {
			return nil, err
		}
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("api error (%d): %s", resp.Code, resp.Message)
	}

	result := make([]DatabaseDetail, 0, len(resp.Databases))
	for _, db := range resp.Databases {
		if db.Name == "" {
			continue
		}
		rawType := "base"
		createTime := ""
		if resp.Info != nil {
			if info, ok := resp.Info[db.Name]; ok && info != nil {
				if info.DbType != "" {
					rawType = info.DbType
				}
				createTime = info.CreateTime
			}
		}

		dbType := "base"
		if isAIDatabase(rawType) {
			dbType = "ai"
		}

		result = append(result, DatabaseDetail{
			Name:       db.Name,
			DbType:     dbType,
			CreateTime: createTime,
		})
	}
	return result, nil
}

func (c *Client) ListDatabases(ctx context.Context) ([]string, error) {
	details, err := c.ListDatabasesDetailed(ctx)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(details))
	for _, d := range details {
		names = append(names, d.Name)
	}
	return names, nil
}

func (c *Client) ListCollections(ctx context.Context, database, dbType string) ([]string, error) {
	// 1. If identified as AI database, query AI collection view list first
	if isAIDatabase(dbType) {
		req := AICollectionViewListReq{Database: database}
		var resp AICollectionViewListRes
		if err := c.doRequest(ctx, http.MethodPost, "/ai/collectionView/list", req, &resp); err == nil && resp.Code == 0 {
			result := make([]string, 0, len(resp.CollectionViews))
			for _, cv := range resp.CollectionViews {
				if cv != nil && cv.CollectionView != "" {
					result = append(result, cv.CollectionView)
				}
			}
			return result, nil
		}
	}

	// 2. Base collection listing
	req := ListCollectionsRequest{Database: database}
	var resp ListCollectionsResponse
	err := c.doRequest(ctx, http.MethodPost, "/collection/list", req, &resp)

	// 3. If Base collection listing returns 11100, this database is an AI database!
	// Auto-retry with AI CollectionView API!
	if (err != nil && (strings.Contains(err.Error(), "11100") || strings.Contains(err.Error(), "user can only handle base data"))) ||
		(resp.Code == 11100) {
		aiReq := AICollectionViewListReq{Database: database}
		var aiResp AICollectionViewListRes
		if aiErr := c.doRequest(ctx, http.MethodPost, "/ai/collectionView/list", aiReq, &aiResp); aiErr == nil && aiResp.Code == 0 {
			result := make([]string, 0, len(aiResp.CollectionViews))
			for _, cv := range aiResp.CollectionViews {
				if cv != nil && cv.CollectionView != "" {
					result = append(result, cv.CollectionView)
				}
			}
			return result, nil
		} else if aiErr != nil {
			return nil, fmt.Errorf("AI 知识库集合请求失败: %w", aiErr)
		}
	}

	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("api error (%d): %s", resp.Code, resp.Message)
	}
	result := make([]string, 0, len(resp.Collections))
	for _, coll := range resp.Collections {
		if coll.Collection != "" {
			result = append(result, coll.Collection)
		}
	}
	return result, nil
}

func (c *Client) DescribeCollection(ctx context.Context, database, collection, dbType string) (*CollectionMeta, error) {
	// 1. If AI database, try describing via AI CollectionView API
	if isAIDatabase(dbType) {
		req := AICollectionViewDescribeReq{Database: database, CollectionView: collection}
		var resp AICollectionViewDescribeRes
		if err := c.doRequest(ctx, http.MethodPost, "/ai/collectionView/describe", req, &resp); err == nil && resp.Code == 0 && resp.CollectionView != nil {
			return c.buildAIMeta(database, collection, resp.CollectionView), nil
		}
	}

	// 2. Base collection describe
	req := DescribeCollectionRequest{Database: database, Collection: collection}
	var resp DescribeCollectionResponse
	err := c.doRequest(ctx, http.MethodPost, "/collection/describe", req, &resp)

	// 3. If Base returns 11100, auto-retry AI CollectionView describe!
	if (err != nil && (strings.Contains(err.Error(), "11100") || strings.Contains(err.Error(), "user can only handle base data"))) ||
		(resp.Code == 11100) {
		aiReq := AICollectionViewDescribeReq{Database: database, CollectionView: collection}
		var aiResp AICollectionViewDescribeRes
		if aiErr := c.doRequest(ctx, http.MethodPost, "/ai/collectionView/describe", aiReq, &aiResp); aiErr == nil && aiResp.Code == 0 && aiResp.CollectionView != nil {
			return c.buildAIMeta(database, collection, aiResp.CollectionView), nil
		} else if aiErr != nil {
			return nil, fmt.Errorf("AI 知识库集合结构请求失败: %w", aiErr)
		}
	}

	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("api error (%d): %s", resp.Code, resp.Message)
	}
	meta := resp.Collection
	if meta.Collection == "" {
		meta.Collection = collection
	}
	if meta.Database == "" {
		meta.Database = database
	}

	if len(meta.Fields) == 0 && len(meta.Indexes) > 0 {
		fieldMap := make(map[string]bool)
		for _, idx := range meta.Indexes {
			if idx.FieldName != "" && !fieldMap[idx.FieldName] {
				fieldMap[idx.FieldName] = true
				isPK := idx.IndexType == "primaryKey"
				meta.Fields = append(meta.Fields, FieldMeta{
					FieldName:  idx.FieldName,
					FieldType:  idx.FieldType,
					PrimaryKey: isPK,
				})
			}
		}
	}

	return &meta, nil
}

func (c *Client) DropCollection(ctx context.Context, database, collection, dbType string) error {
	// 1. If AI database, try dropping via AI CollectionView API
	if isAIDatabase(dbType) {
		req := AICollectionViewDropReq{Database: database, CollectionView: collection}
		var resp ResponseHeader
		err := c.doRequest(ctx, http.MethodPost, "/ai/collectionView/drop", req, &resp)
		if err == nil && resp.Code == 0 {
			return nil
		}
		if err != nil && !strings.Contains(err.Error(), "11100") && !strings.Contains(err.Error(), "base data") {
			return fmt.Errorf("AI 知识库集合删除失败: %w", err)
		}
		if resp.Code != 0 && resp.Code != 11100 {
			return fmt.Errorf("api error (%d): %s", resp.Code, resp.Message)
		}
	}

	// 2. Base collection drop
	req := DropCollectionRequest{Database: database, Collection: collection}
	var resp DropCollectionResponse
	err := c.doRequest(ctx, http.MethodPost, "/collection/drop", req, &resp)

	// 3. If Base collection returns 11100, auto-retry AI CollectionView drop!
	if (err != nil && (strings.Contains(err.Error(), "11100") || strings.Contains(err.Error(), "user can only handle base data"))) ||
		(resp.Code == 11100) {
		aiReq := AICollectionViewDropReq{Database: database, CollectionView: collection}
		var aiResp ResponseHeader
		if aiErr := c.doRequest(ctx, http.MethodPost, "/ai/collectionView/drop", aiReq, &aiResp); aiErr == nil && aiResp.Code == 0 {
			return nil
		} else if aiErr != nil {
			return fmt.Errorf("AI 知识库集合删除失败: %w", aiErr)
		} else if aiResp.Code != 0 {
			return fmt.Errorf("api error (%d): %s", aiResp.Code, aiResp.Message)
		}
		return nil
	}

	if err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("api error (%d): %s", resp.Code, resp.Message)
	}
	return nil
}

func (c *Client) QueryDocuments(ctx context.Context, database, collection, dbType string, limit, offset int, filter string) (*QueryDocumentResponse, error) {
	if limit <= 0 {
		limit = 20
	}

	// 1. If AI database, try querying via AI DocumentSet API
	if isAIDatabase(dbType) {
		req := AIDocumentSetQueryReq{
			Database:       database,
			CollectionView: collection,
		}
		req.Query.Limit = int64(limit)
		req.Query.Offset = int64(offset)
		req.Query.Filter = filter
		var resp AIDocumentSetQueryRes
		if err := c.doRequest(ctx, http.MethodPost, "/ai/documentSet/query", req, &resp); err == nil && resp.Code == 0 {
			count := resp.Count
			if count == 0 && len(resp.DocumentSets) > 0 {
				count = uint64(len(resp.DocumentSets))
			}
			return &QueryDocumentResponse{
				ResponseHeader: resp.ResponseHeader,
				Count:          count,
				Documents:      resp.DocumentSets,
			}, nil
		}
	}

	// 2. Base document query
	req := QueryDocumentRequest{
		Database:   database,
		Collection: collection,
		Query: &QueryCond{
			RetrieveVector: true,
			Filter:         filter,
			Limit:          int64(limit),
			Offset:         int64(offset),
		},
	}

	var resp QueryDocumentResponse
	err := c.doRequest(ctx, http.MethodPost, "/document/query", req, &resp)

	// 3. If Base returns 11100, auto-retry AI DocumentSet query!
	if (err != nil && (strings.Contains(err.Error(), "11100") || strings.Contains(err.Error(), "user can only handle base data"))) ||
		(resp.Code == 11100) {
		aiReq := AIDocumentSetQueryReq{
			Database:       database,
			CollectionView: collection,
		}
		aiReq.Query.Limit = int64(limit)
		aiReq.Query.Offset = int64(offset)
		aiReq.Query.Filter = filter
		var aiResp AIDocumentSetQueryRes
		if aiErr := c.doRequest(ctx, http.MethodPost, "/ai/documentSet/query", aiReq, &aiResp); aiErr == nil && aiResp.Code == 0 {
			count := aiResp.Count
			if count == 0 && len(aiResp.DocumentSets) > 0 {
				count = uint64(len(aiResp.DocumentSets))
			}
			return &QueryDocumentResponse{
				ResponseHeader: aiResp.ResponseHeader,
				Count:          count,
				Documents:      aiResp.DocumentSets,
			}, nil
		} else if aiErr != nil {
			return nil, fmt.Errorf("AI 知识库文件切片请求失败: %w", aiErr)
		}
	}

	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("api error (%d): %s", resp.Code, resp.Message)
	}
	if resp.Count == 0 && len(resp.Documents) > 0 {
		resp.Count = uint64(len(resp.Documents))
	}
	return &resp, nil
}

func flattenMap(prefix string, m map[string]interface{}, out map[string]interface{}) {
	for k, v := range m {
		key := k
		if prefix != "" && prefix != "metaData" && prefix != "metadata" {
			key = prefix + "_" + k
		}
		if subM, ok := v.(map[string]interface{}); ok {
			flattenMap(key, subM, out)
		} else {
			out[key] = v
		}
	}
}

// sanitizeUpdate ensures the update payload only contains supported scalar types.
// Tencent VectorDB strictly forbids nested map[string]interface{} types (error 14100).
// Any nested maps or objects (such as metaData: { ... }) are automatically unpacked into top-level scalar fields.
func sanitizeUpdate(update map[string]interface{}) map[string]interface{} {
	clean := make(map[string]interface{})
	for k, v := range update {
		if m, ok := v.(map[string]interface{}); ok {
			prefix := ""
			if k != "metaData" && k != "metadata" {
				prefix = k
			}
			flattenMap(prefix, m, clean)
		} else {
			clean[k] = v
		}
	}
	return clean
}

var aiReadOnlyFields = map[string]bool{
	"documentsetid":            true,
	"documentsetname":          true,
	"document_set_id":          true,
	"document_set_name":        true,
	"status":                   true,
	"filetype":                 true,
	"file_type":                true,
	"bytesize":                 true,
	"byte_size":                true,
	"bytelength":               true,
	"byte_length":              true,
	"chunknum":                 true,
	"chunk_num":                true,
	"createtime":               true,
	"create_time":              true,
	"updatetime":               true,
	"update_time":              true,
	"lastupdatetime":           true,
	"last_update_time":         true,
	"textprefix":               true,
	"text_prefix":              true,
	"text":                     true,
	"textlength":               true,
	"text_length":              true,
	"indexedprogress":          true,
	"indexed_progress":         true,
	"indexedstatus":            true,
	"indexed_status":           true,
	"indexederrormsg":          true,
	"indexed_error_msg":        true,
	"keywords":                 true,
	"appendtitletochunk":       true,
	"append_title_to_chunk":    true,
	"appendkeywordstochunk":    true,
	"append_keywords_to_chunk": true,
	"chunksplitter":            true,
	"chunk_splitter":           true,
	"splitterprocess":          true,
	"splitter_process":         true,
	"splitterpreprocess":       true,
	"splitter_preprocess":      true,
	"parsingprocess":           true,
	"parsing_process":          true,
	"parsingtype":              true,
	"parsing_type":             true,
	"metadata":                 true,
}

func filterAIRuntimeFields(update map[string]interface{}) map[string]interface{} {
	filtered := make(map[string]interface{})
	for k, v := range update {
		if aiReadOnlyFields[strings.ToLower(k)] {
			continue
		}
		// VectorDB AI Knowledge Base scalar fields strictly only support uint64, string, or string element array type.
		// Passing a bool triggers error 14100: field "xxx" is unsupported bool type, only support uint64, string or string element array type.
		// Automatically coerce bool to uint64: true -> 1, false -> 0.
		if b, ok := v.(bool); ok {
			if b {
				filtered[k] = uint64(1)
			} else {
				filtered[k] = uint64(0)
			}
			continue
		}
		// If float64 has no fractional part and is non-negative, cast to uint64
		if f, ok := v.(float64); ok && f >= 0 && f == float64(uint64(f)) {
			filtered[k] = uint64(f)
			continue
		}
		filtered[k] = v
	}
	return filtered
}

func (c *Client) UpdateDocument(ctx context.Context, database, collection, dbType string, query UpdateDocumentQuery, update map[string]interface{}) error {
	cleanUpdate := sanitizeUpdate(update)
	if len(cleanUpdate) == 0 {
		return fmt.Errorf("更新字段内容不能为空")
	}

	// 1. Explicit AI CollectionView / DocumentSet update
	if dbType == "ai" {
		aiUpdate := filterAIRuntimeFields(cleanUpdate)
		if len(aiUpdate) == 0 {
			return fmt.Errorf("未检测到可更新的自定义字段（文件名 documentSetName、状态 status 等系统内置属性由知识库自动维护，不可修改）")
		}

		aiReq := AIDocumentSetUpdateReq{
			Database:       database,
			CollectionView: collection,
			Update:         aiUpdate,
		}
		if len(query.DocumentSetIds) > 0 {
			aiReq.Query.DocumentSetId = query.DocumentSetIds
			aiReq.Query.DocumentSetIds = query.DocumentSetIds
		} else if len(query.DocumentIds) > 0 {
			aiReq.Query.DocumentSetId = query.DocumentIds
			aiReq.Query.DocumentSetIds = query.DocumentIds
		}
		if len(query.DocumentSetNames) > 0 {
			aiReq.Query.DocumentSetNames = query.DocumentSetNames
		}
		aiReq.Query.Filter = query.Filter

		var resp UpdateDocumentResponse
		if err := c.doRequest(ctx, http.MethodPost, "/ai/documentSet/update", aiReq, &resp); err != nil {
			return fmt.Errorf("AI 知识库文档更新失败: %w", err)
		}
		if resp.Code != 0 {
			return fmt.Errorf("AI 知识库更新失败 (%d): %s", resp.Code, resp.Message)
		}
		return nil
	}

	// 2. Base Collection document update
	req := UpdateDocumentRequest{
		Database:   database,
		Collection: collection,
		Query:      query,
		Update:     cleanUpdate,
	}
	var resp UpdateDocumentResponse
	err := c.doRequest(ctx, http.MethodPost, "/document/update", req, &resp)

	// 3. If Base returns 11100, auto-retry AI DocumentSet update!
	if (err != nil && (strings.Contains(err.Error(), "11100") || strings.Contains(err.Error(), "user can only handle base data"))) ||
		(resp.Code == 11100) {
		aiUpdate := filterAIRuntimeFields(cleanUpdate)
		if len(aiUpdate) == 0 {
			return fmt.Errorf("未检测到可更新的自定义字段（文件名 documentSetName、状态 status 等系统内置属性由知识库自动维护，不可修改）")
		}

		aiReq := AIDocumentSetUpdateReq{
			Database:       database,
			CollectionView: collection,
			Update:         aiUpdate,
		}
		if len(query.DocumentSetIds) > 0 {
			aiReq.Query.DocumentSetId = query.DocumentSetIds
			aiReq.Query.DocumentSetIds = query.DocumentSetIds
		} else if len(query.DocumentIds) > 0 {
			aiReq.Query.DocumentSetId = query.DocumentIds
			aiReq.Query.DocumentSetIds = query.DocumentIds
		}
		if len(query.DocumentSetNames) > 0 {
			aiReq.Query.DocumentSetNames = query.DocumentSetNames
		}
		aiReq.Query.Filter = query.Filter

		var aiResp UpdateDocumentResponse
		if aiErr := c.doRequest(ctx, http.MethodPost, "/ai/documentSet/update", aiReq, &aiResp); aiErr == nil && aiResp.Code == 0 {
			return nil
		} else if aiErr != nil {
			return fmt.Errorf("AI 知识库更新失败: %w", aiErr)
		} else if aiResp.Code != 0 {
			return fmt.Errorf("AI 知识库更新失败 (%d): %s", aiResp.Code, aiResp.Message)
		}
		return nil
	}

	if err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("api error (%d): %s", resp.Code, resp.Message)
	}
	return nil
}
