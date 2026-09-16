import type { ConnectionConfig } from "../../bindings/vectordb-1/backend/storage/models";
import type { CollectionMeta, FieldMeta, IndexMeta, QueryDocumentResponse } from "../../bindings/vectordb-1/backend/client/models";

export type { ConnectionConfig, CollectionMeta, FieldMeta, IndexMeta, QueryDocumentResponse };

export interface TestResult {
  success: boolean;
  message: string;
}
