package client

type ResponseHeader struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

type ListDatabasesResponse struct {
	ResponseHeader
	Databases []struct {
		Database string `json:"database"`
	} `json:"databases"`
}

type ListCollectionsRequest struct {
	Database string `json:"database"`
}

type ListCollectionsResponse struct {
	ResponseHeader
	Collections []struct {
		Collection string `json:"collection"`
	} `json:"collections"`
}

type DescribeCollectionRequest struct {
	Database   string `json:"database"`
	Collection string `json:"collection"`
}

type FieldMeta struct {
	FieldName   string `json:"fieldName"`
	FieldType   string `json:"fieldType"`
	FieldUsage  string `json:"fieldUsage"`
	PrimaryKey  bool   `json:"primaryKey"`
	Description string `json:"description"`
}

type IndexMeta struct {
	FieldName  string                 `json:"fieldName"`
	IndexType  string                 `json:"indexType"`
	MetricType string                 `json:"metricType"`
	Params     map[string]interface{} `json:"params"`
}

type CollectionMeta struct {
	Database    string      `json:"database"`
	Collection  string      `json:"collection"`
	ReplicaNum  int         `json:"replicaNum"`
	ShardNum    int         `json:"shardNum"`
	Description string      `json:"description"`
	Fields      []FieldMeta `json:"fields"`
	Indexes     []IndexMeta `json:"indexes"`
}

type DescribeCollectionResponse struct {
	ResponseHeader
	Collection CollectionMeta `json:"collection"`
}

type QueryDocumentRequest struct {
	Database   string `json:"database"`
	Collection string `json:"collection"`
	Query      struct {
		Limit  int    `json:"limit"`
		Offset int    `json:"offset"`
		Filter string `json:"filter,omitempty"`
	} `json:"query"`
}

type QueryDocumentResponse struct {
	ResponseHeader
	Count     int                      `json:"count"`
	Documents []map[string]interface{} `json:"documents"`
}
