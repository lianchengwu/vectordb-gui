import { ref } from "vue";
import type { CollectionMeta, DatabaseDetail } from "../types";
import {
  ListDatabasesDetailed,
  ListCollections,
  DescribeCollection,
} from "../../bindings/vectordb-1/backend/service/vectordbservice";

const databases = ref<DatabaseDetail[]>([]);
const collectionsByDb = ref<Record<string, string[]>>({});
const expandedDbs = ref<Record<string, boolean>>({});
const activeDatabase = ref<string>("");
const activeCollection = ref<string>("");
const activeDbType = ref<string>("base");
const activeCollectionMeta = ref<CollectionMeta | null>(null);
const loading = ref(false);
const loadingCollections = ref<Record<string, boolean>>({});
const loadingMeta = ref(false);
const error = ref<string>("");

export function useVectorDBStore() {
  async function loadDatabases(connId: string) {
    if (!connId) {
      reset();
      return;
    }
    loading.value = true;
    error.value = "";
    try {
      const dbs = await ListDatabasesDetailed(connId);
      databases.value = dbs || [];
      // Auto expand the first database if available
      if (databases.value.length > 0) {
        const firstDb = databases.value[0];
        expandedDbs.value[firstDb.name] = true;
        await loadCollections(connId, firstDb.name, firstDb.dbType);
      }
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : String(e);
      databases.value = [];
    } finally {
      loading.value = false;
    }
  }

  function getDbType(database: string): string {
    const d = databases.value.find((item) => item.name === database);
    return d?.dbType || "base";
  }

  async function loadCollections(connId: string, database: string, dbType?: string) {
    if (!connId || !database) return;
    loadingCollections.value[database] = true;
    try {
      const type = dbType || getDbType(database);
      const colls = await ListCollections(connId, database, type);
      collectionsByDb.value[database] = colls || [];
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : String(e);
      collectionsByDb.value[database] = [];
    } finally {
      loadingCollections.value[database] = false;
    }
  }

  async function toggleDatabase(connId: string, database: string) {
    const isExpanded = !!expandedDbs.value[database];
    expandedDbs.value[database] = !isExpanded;
    if (!isExpanded && !collectionsByDb.value[database]) {
      await loadCollections(connId, database, getDbType(database));
    }
  }

  async function selectCollection(connId: string, database: string, collection: string) {
    activeDatabase.value = database;
    activeCollection.value = collection;
    const type = getDbType(database);
    activeDbType.value = type;
    loadingMeta.value = true;
    try {
      const meta = await DescribeCollection(connId, database, collection, type);
      activeCollectionMeta.value = meta;
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : String(e);
      activeCollectionMeta.value = null;
    } finally {
      loadingMeta.value = false;
    }
  }

  function reset() {
    databases.value = [];
    collectionsByDb.value = {};
    expandedDbs.value = {};
    activeDatabase.value = "";
    activeCollection.value = "";
    activeDbType.value = "base";
    activeCollectionMeta.value = null;
    error.value = "";
  }

  return {
    databases,
    collectionsByDb,
    expandedDbs,
    activeDatabase,
    activeCollection,
    activeDbType,
    activeCollectionMeta,
    loading,
    loadingCollections,
    loadingMeta,
    error,
    getDbType,
    loadDatabases,
    loadCollections,
    toggleDatabase,
    selectCollection,
    reset,
  };
}
