import type {
  ConnectionConfig,
  NetworkHop,
  ProxyConfig,
  SSHTunnelConfig,
} from "../../bindings/vectordb-1/backend/storage/models";
import type {
  CollectionMeta,
  DatabaseDetail,
  FieldMeta,
  IndexColumn,
  QueryDocumentResponse,
} from "../../bindings/vectordb-1/backend/client/models";

export type {
  ConnectionConfig,
  NetworkHop,
  ProxyConfig,
  SSHTunnelConfig,
  CollectionMeta,
  DatabaseDetail,
  FieldMeta,
  IndexColumn,
  QueryDocumentResponse,
};

export interface CollectionTab {
  id: string; // `${connId}:${database}:${collection}`
  connId: string;
  database: string;
  collection: string;
  dbType: string;
  meta: CollectionMeta | null;
}

export interface TestResult {
  success: boolean;
  message: string;
}
