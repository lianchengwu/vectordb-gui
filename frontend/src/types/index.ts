import type { ConnectionConfig, ProxyConfig, SSHTunnelConfig } from "../../bindings/vectordb-1/backend/storage/models";
import type { CollectionMeta, FieldMeta, IndexColumn, QueryDocumentResponse } from "../../bindings/vectordb-1/backend/client/models";

export type {
  ConnectionConfig,
  ProxyConfig,
  SSHTunnelConfig,
  CollectionMeta,
  FieldMeta,
  IndexColumn,
  QueryDocumentResponse,
};

export interface TestResult {
  success: boolean;
  message: string;
}
