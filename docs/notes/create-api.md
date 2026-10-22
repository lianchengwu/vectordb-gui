# 腾讯云 VectorDB 创建类 REST API 调研笔记

调研日期:2026-10-22

数据来源:
- 官方 Go SDK 源码(Tencent/vectordatabase-sdk-go,main 分支,与项目使用的 tcvectordb-go-v1.9.1 同源):
  - https://github.com/Tencent/vectordatabase-sdk-go
  - `tcvectordb/api/database/api.go`、`tcvectordb/api/ai_database/api.go`、`tcvectordb/api/collection/api.go`、`tcvectordb/api/collection_view/api.go`、`tcvectordb/api/common.go`
  - `tcvectordb/base_database.go`、`tcvectordb/base_collection.go`、`tcvectordb/ai_collection_view.go`、`tcvectordb/rpc_base_collection.go`
  - `tcvectordb/index_field.go`、`example/ai_demo/main.go`
- 腾讯云官方 HTTP API 文档入口(SDK 契约与其一致;具体接口页有反爬,字段以 SDK json tag 为准):
  - HTTP API 总览/使用前阅读: https://cloud.tencent.com/document/product/1709/98666
  - 产品动态(确认 /collection/create 等接口演进): https://cloud.tencent.com/document/product/1709/95477

所有 JSON 字段名均取自 SDK 结构体的 `json:"..."` tag,即 HTTP 请求体序列化后的真实字段名,与本项目 `backend/client/types.go` 现有命名习惯一致(`database`、`collection`、`shardNum`、`replicaNum`、`indexes`、`fieldName`、`fieldType`、`indexType`、`metricType`、`dimension`)。

认证与调用方式沿用项目现有 `doRequest(ctx, http.MethodPost, path, reqBody, respBody)`:`Authorization: Bearer account=xxx api_key=xxx`,`Sdk-Version: tcvectordb-go-v1.9.1`。

---

## 1. 创建数据库

### 1.1 普通数据库(集合的宿主)

- 路径:`POST /database/create`
- 来源:SDK `tcvectordb/api/database/api.go` `CreateReq`(Meta tag `path:"/database/create" method:"Post"`)

请求体(仅一个必填字段,**无 description 字段**):

```json
{
  "database": "db-test"
}
```

响应体:

```json
{
  "code": 0,
  "msg": "operation ok",
  "databases": ["db-test"],
  "affectedCount": 1
}
```

字段说明:
- `database`:必填。数据库名。SDK 注释约定名称为字母/数字/下划线/连字符,以字母数字开头(与 collectionView 命名规则一致:1-128 字符、字母或数字开头,仅含字母数字、`_`、`-`)。
- 若 database 已存在,接口幂等返回成功(code=0)。
- **没有 description 字段** —— 数据库层级不支持描述;GUI 如需备注只能本地存储。

### 1.2 AI 数据库(知识库 collectionView 的宿主)

- 路径:`POST /ai/database/create`
- 来源:SDK `tcvectordb/api/ai_database/api.go`(`path:"/ai/database/create" method:"Post"`)
- 请求体与响应结构与 1.1 完全相同(`{ "database": "ai-db-test" }`)。`/database/list` 返回的 `info.<name>.dbType` 为 `"base"` 或 `"ai"`,可用于创建后区分。

本项目现有 `ListDatabases` 已解析 `dbType`,创建时应让用户选择库类型:base 库 → `/database/create`,AI 库 → `/ai/database/create`。

---

## 2. 创建普通集合

- 路径:`POST /collection/create`
- 来源:SDK `tcvectordb/api/collection/api.go` `CreateReq`;填充逻辑见 `tcvectordb/base_collection.go:164 CreateCollection`。

### 请求体示例

```json
{
  "database": "db-test",
  "collection": "book_vec",
  "replicaNum": 2,
  "shardNum": 1,
  "description": "书籍向量集合",
  "indexes": [
    {
      "fieldName": "id",
      "fieldType": "string",
      "indexType": "primaryKey"
    },
    {
      "fieldName": "vector",
      "fieldType": "vector",
      "indexType": "HNSW",
      "dimension": 768,
      "metricType": "COSINE",
      "params": { "M": 16, "efConstruction": 200 }
    },
    {
      "fieldName": "bookName",
      "fieldType": "string",
      "indexType": "filter"
    }
  ]
}
```

### 字段说明

顶层字段:

| 字段 | 类型 | 说明 |
|---|---|---|
| `database` | string | 必填,目标数据库 |
| `collection` | string | 必填,集合名。1-128 字符,字母或数字开头,仅含字母数字、`_`、`-`(SDK `base_collection.go:120` 注释) |
| `replicaNum` | uint32 | 副本数,默认/常见值 2;不能超过节点数 |
| `shardNum` | uint32 | 分片数,默认/常见值 1 |
| `description` | string | 可选,集合描述 |
| `indexes` | array | 必填,索引数组(见下)。必须含一个 `primaryKey` 索引(字段名通常为 `id`)和至少一个向量索引 |
| `embedding` | object | 可选,写入文本自动向量化:`{ "field": "...", "vectorField": "...", "model": "bge-m3" }`(字段名 `field`/`vectorField`/`model`) |
| `ttlConfig` | object | 可选:`{ "enable": true, "timeField": "expireAt" }` |
| `filterIndexConfig` | object | 可选,标量索引高级配置:`{ "filterAll": false, "fieldsWithoutIndex": [...], "maxStrLen": 1024 }` |

`indexes[]` 每项(SDK `tcvectordb/api/common.go:61 IndexColumn`):

| 字段 | 说明 |
|---|---|
| `fieldName` | 索引字段名,同时即数据字段名 |
| `fieldType` | `string` / `uint64` / `array` / `vector` / `sparseVector` 等 |
| `fieldElementType` | 仅 `fieldType=array` 时使用,如 `"string"` |
| `indexType` | `primaryKey` / `filter` / 向量索引 `HNSW` / `IVF_FLAT` / `IVF_SQ8` / `IVF_SQ16` / `IVF_PQ` 等 |
| `dimension` | **向量索引必填**,顶层字段(非 params 内),向量维度,1-65536(需与所用模型一致) |
| `metricType` | 向量索引必填:`COSINE` / `L2` / `IP`(稀疏向量仅 `IP`) |
| `params` | 向量索引可选参数(HNSW:`M`(4-64,默认16)、`efConstruction`(100-1000,默认200);IVF 系列:`nlist`;PQ:`bits`、`nprobe`) |
| `autoId` | filter 索引可选:`""`/`true`/`false`,主键是否自动生成 |
| `diskSwapEnabled` | *bool,稀疏向量索引可选,索引落盘 |

注意:`dimension` 是索引项的**顶层字段**,与本项目 `types.go` 的 `IndexColumn.Dimension`(`json:"dimension,omitempty"`)一致,不要放进 `params`。

### 响应体

```json
{
  "code": 0,
  "msg": "operation ok",
  "affectedCount": 1
}
```

(`CreateRes` = `CommonRes{Code, Msg}` + `AffectedCount`)

### 高级特性说明(createCollectionFilters / createCollectionIndexes)

Go SDK 不存在这两个选项。等价能力是:
- 不指定 `indexes` → 1.x SDK 曾支持先建集合后加索引,但当前 SDK `CreateCollection` 的 `indexes` 为必填;新增索引走 `POST /index/add`(`AddIndex`),删除走 `/index/drop`。
- 标量字段免过滤索引:`filterIndexConfig`(`filterAll: true` 表示所有标量字段可过滤,无需逐一建 filter 索引)。

---

## 3. 创建 AI 知识库集合(collectionView)

- 路径:`POST /ai/collectionView/create`
- 来源:SDK `tcvectordb/api/collection_view/api.go` `CreateReq`(Meta tag `path:"/ai/collectionView/create" method:"Post"`);填充逻辑见 `tcvectordb/ai_collection_view.go:150 CreateCollectionView`。
- 前提:`database` 必须是 **AI 库**(由 `/ai/database/create` 创建,或 list 中 `dbType=="ai"`),否则服务端/SDK 报错。

### 请求体示例(来自 SDK `example/ai_demo/main.go` 的等价 JSON)

```json
{
  "database": "ai-db-test",
  "collectionView": "coll-view-test",
  "description": "desc",
  "indexes": [
    { "fieldName": "test_str", "fieldType": "string", "indexType": "filter" }
  ],
  "embedding": {
    "language": "zh",
    "enableWordsEmbedding": false
  },
  "splitterPreprocess": {
    "appendTitleToChunk": false,
    "appendKeywordsToChunk": true
  },
  "parsingProcess": {
    "parsingType": "AlgorithmParsing"
  },
  "shardNum": 3,
  "replicaNum": 1,
  "expectedFileNum": 10240,
  "averageFileSize": 1024
}
```

### 字段说明

| 字段 | 类型 | 说明 |
|---|---|---|
| `database` | string | 必填,必须是 AI 库 |
| `collectionView` | string | 必填,1-128 字符,字母或数字开头,仅含字母数字、`_`、`-`(SDK `ai_collection_view.go:143` 注释) |
| `description` | string | 可选 |
| `indexes` | array | 可选,标量过滤索引(`fieldName`/`fieldType`/`indexType:"filter"`)。**不需要向量索引/主键** —— 向量字段由服务端内置 |
| `embedding` | object | 可选:`{ "language": "zh"|"en" }`、`enableWordsEmbedding`(*bool, 关键词召回) |
| `splitterPreprocess` | object | 可选,预处理:`appendTitleToChunk`(*bool, 默认 false)、`appendKeywordsToChunk`(*bool, 默认 true) |
| `parsingProcess` | object | 可选,解析策略:`parsingType`: `"AlgorithmParsing"`(算法解析,默认)或 `"VisionModelParsing"`(视觉大模型解析,仅 PDF) |
| `shardNum` / `replicaNum` | *uint32 | 可选,分片/副本数 |
| `expectedFileNum` / `averageFileSize` | uint64 | 可选,预估文件数与平均文件大小(用于服务端资源规划) |

### 重要澄清

- **没有 `modelName` / `embeddingModel` 字段**:collectionView 的 embedding 对象只有 `language` 与 `enableWordsEmbedding`。Embedding 模型是在**上传文件时**(`/ai/documentSet/upload`,传给 embedding 接口的 model 参数)或查询时指定的,collectionView 层不绑定模型。普通集合才有顶层 `embedding.model`(文本自动向量化字段)。
- **没有 `splitType` / `chunkSize` / `overlap` / `parseType` 字段**:分块策略不按这些字段配置。SDK 中的实际字段是 `splitterPreprocess`(仅 append 标题/关键词开关)与 `parsingProcess.parsingType`(`AlgorithmParsing`/`VisionModelParsing`)。chunk 粒度拆分参数属于上传文档时的 `splitter` 配置,不在创建接口中。

### 响应体

```json
{
  "code": 0,
  "msg": "operation ok",
  "affectedCount": 1
}
```

(`CreateRes` = `CommonRes` + `AffectedCount`)

---

## 4. 对本项目实现的建议摘要

1. 创建数据库:POST `/database/create` 或 `/ai/database/create`(由用户选择 dbType),请求体 `{"database": name}`,响应判 `code==0` + `affectedCount`。
2. 创建集合:POST `/collection/create`,复用现有 `IndexColumn` 结构(`dimension` 在顶层);indexes 数组必须含 1 个 `primaryKey` + ≥1 个向量索引;`shardNum`/`replicaNum`/`description` 沿用现有 json tag。
3. 创建 AI 知识库:POST `/ai/collectionView/create`,字段为 `collectionView`/`description`/`indexes`(仅 filter)/`embedding.language`/`splitterPreprocess`/`parsingProcess.parsingType`;不要暴露 modelName/chunkSize/overlap 等不存在的字段。
4. 三个接口响应统一为 `{code, msg, affectedCount}`,与项目 `ResponseHeader` + `affectedCount` 兼容。

字段名与结构均以 SDK json tag 为准(来源见文首),官方文档页 1709/98666 为接口索引入口。
