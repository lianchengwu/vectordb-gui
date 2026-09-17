package client

import (
	"encoding/json"
)

type ResponseHeader struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

type DatabaseEntry struct {
	Name string
}

func (d *DatabaseEntry) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		d.Name = s
		return nil
	}
	var obj struct {
		Database string `json:"database"`
		Name     string `json:"name"`
	}
	if err := json.Unmarshal(data, &obj); err == nil {
		if obj.Database != "" {
			d.Name = obj.Database
		} else {
			d.Name = obj.Name
		}
		return nil
	}
	return nil
}

type DatabaseInfo struct {
	CreateTime string `json:"createTime,omitempty"`
	DbType     string `json:"dbType,omitempty"` // "base" or "ai"
	Count      int64  `json:"count,omitempty"`
}

type DatabaseDetail struct {
	Name       string `json:"name"`
	DbType     string `json:"dbType"` // "base" or "ai"
	CreateTime string `json:"createTime,omitempty"`
}

type ListDatabasesResponse struct {
	ResponseHeader
	Databases     []DatabaseEntry          `json:"databases"`
	AffectedCount int                      `json:"affectedCount,omitempty"`
	Info          map[string]*DatabaseInfo `json:"info,omitempty"`
}

type ListCollectionsRequest struct {
	Database string `json:"database"`
}

type CollectionEntry struct {
	Database    string        `json:"database,omitempty"`
	Collection  string        `json:"collection,omitempty"`
	ReplicaNum  uint32        `json:"replicaNum,omitempty"`
	ShardNum    uint32        `json:"shardNum,omitempty"`
	Description string        `json:"description,omitempty"`
	Indexes     []IndexColumn `json:"indexes,omitempty"`
}

func (c *CollectionEntry) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		c.Collection = s
		return nil
	}
	var obj struct {
		Database    string        `json:"database"`
		Collection  string        `json:"collection"`
		Name        string        `json:"name"`
		ReplicaNum  uint32        `json:"replicaNum"`
		ShardNum    uint32        `json:"shardNum"`
		Description string        `json:"description"`
		Indexes     []IndexColumn `json:"indexes"`
	}
	if err := json.Unmarshal(data, &obj); err == nil {
		c.Database = obj.Database
		if obj.Collection != "" {
			c.Collection = obj.Collection
		} else {
			c.Collection = obj.Name
		}
		c.ReplicaNum = obj.ReplicaNum
		c.ShardNum = obj.ShardNum
		c.Description = obj.Description
		c.Indexes = obj.Indexes
		return nil
	}
	return nil
}

type ListCollectionsResponse struct {
	ResponseHeader
	Collections []CollectionEntry `json:"collections"`
}

// AI CollectionView (Knowledge Base) types
type AICollectionViewListReq struct {
	Database string `json:"database"`
}

type AICollectionViewItem struct {
	Database       string `json:"database"`
	CollectionView string `json:"collectionView"`
	Description    string `json:"description,omitempty"`
}

type AICollectionViewListRes struct {
	ResponseHeader
	CollectionViews []*AICollectionViewItem `json:"collectionViews"`
}

type AICollectionViewDescribeReq struct {
	Database       string `json:"database"`
	CollectionView string `json:"collectionView"`
}

type AICollectionViewData struct {
	Database       string        `json:"database"`
	CollectionView string        `json:"collectionView"`
	Description    string        `json:"description,omitempty"`
	Indexes        []IndexColumn `json:"indexes,omitempty"`
	ReplicaNum     *uint32       `json:"replicaNum,omitempty"`
	ShardNum       *uint32       `json:"shardNum,omitempty"`
}

type AICollectionViewDescribeRes struct {
	ResponseHeader
	CollectionView *AICollectionViewData `json:"collectionView"`
}

type AIDocumentSetQueryReq struct {
	Database       string `json:"database"`
	CollectionView string `json:"collectionView"`
	Query          struct {
		Limit  int64  `json:"limit,omitempty"`
		Offset int64  `json:"offset,omitempty"`
		Filter string `json:"filter,omitempty"`
	} `json:"query"`
}

type AIDocumentSetQueryRes struct {
	ResponseHeader
	Count        uint64                   `json:"count"`
	DocumentSets []map[string]interface{} `json:"documentSets"`
}

type DescribeCollectionRequest struct {
	Database   string `json:"database"`
	Collection string `json:"collection"`
}

type DropCollectionRequest struct {
	Database   string `json:"database"`
	Collection string `json:"collection"`
}

type DropCollectionResponse struct {
	ResponseHeader
	AffectedCount int `json:"affectedCount,omitempty"`
}

type AICollectionViewDropReq struct {
	Database       string `json:"database"`
	CollectionView string `json:"collectionView"`
}

type FieldMeta struct {
	FieldName   string `json:"fieldName"`
	FieldType   string `json:"fieldType"`
	FieldUsage  string `json:"fieldUsage,omitempty"`
	PrimaryKey  bool   `json:"primaryKey,omitempty"`
	Description string `json:"description,omitempty"`
}

type IndexParams struct {
	M              uint32 `json:"M,omitempty"`
	EfConstruction uint32 `json:"efConstruction,omitempty"`
	Nprobe         uint32 `json:"nprobe,omitempty"`
	Nlist          uint32 `json:"nlist,omitempty"`
}

type IndexColumn struct {
	FieldName    string                 `json:"fieldName,omitempty"`
	FieldType    string                 `json:"fieldType,omitempty"`
	IndexType    string                 `json:"indexType,omitempty"`
	Dimension    uint32                 `json:"dimension,omitempty"`
	MetricType   string                 `json:"metricType,omitempty"`
	IndexedCount uint64                 `json:"indexedCount,omitempty"`
	Params       map[string]interface{} `json:"params,omitempty"`
}

type CollectionMeta struct {
	Database       string        `json:"database"`
	Collection     string        `json:"collection"`
	ReplicaNum     uint32        `json:"replicaNum"`
	ShardNum       uint32        `json:"shardNum"`
	Description    string        `json:"description,omitempty"`
	Fields         []FieldMeta   `json:"fields,omitempty"`
	Indexes        []IndexColumn `json:"indexes,omitempty"`
	IsAICollection bool          `json:"isAiCollection,omitempty"`
}

type DescribeCollectionResponse struct {
	ResponseHeader
	Collection CollectionMeta `json:"collection"`
}

type QueryCond struct {
	DocumentIds    []string `json:"documentIds,omitempty"`
	RetrieveVector bool     `json:"retrieveVector"`
	Filter         string   `json:"filter,omitempty"`
	Limit          int64    `json:"limit,omitempty"`
	Offset         int64    `json:"offset,omitempty"`
}

type QueryDocumentRequest struct {
	Database   string     `json:"database"`
	Collection string     `json:"collection"`
	Query      *QueryCond `json:"query,omitempty"`
}

type QueryDocumentResponse struct {
	ResponseHeader
	Count     uint64                   `json:"count"`
	Documents []map[string]interface{} `json:"documents"`
}

type UpdateDocumentQuery struct {
	DocumentIds      []string `json:"documentIds,omitempty"`
	DocumentSetIds   []string `json:"documentSetIds,omitempty"`
	DocumentSetNames []string `json:"documentSetNames,omitempty"`
	Filter           string   `json:"filter,omitempty"`
}

type UpdateDocumentRequest struct {
	Database   string                 `json:"database"`
	Collection string                 `json:"collection"`
	Query      UpdateDocumentQuery    `json:"query"`
	Update     map[string]interface{} `json:"update"`
}

type AIDocumentSetUpdateReq struct {
	Database       string `json:"database"`
	CollectionView string `json:"collectionView"`
	Query          struct {
		DocumentSetId    []string `json:"documentSetId,omitempty"`
		DocumentSetIds   []string `json:"documentSetIds,omitempty"`
		DocumentSetNames []string `json:"documentSetNames,omitempty"`
		Filter           string   `json:"filter,omitempty"`
	} `json:"query"`
	Update map[string]interface{} `json:"update"`
}

type UpdateDocumentResponse struct {
	ResponseHeader
	AffectedCount int `json:"affectedCount,omitempty"`
}
