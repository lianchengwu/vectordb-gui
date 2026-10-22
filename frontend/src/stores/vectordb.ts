import { ref, computed } from "vue";
import type { CollectionMeta, DatabaseDetail, CollectionTab } from "../types";
import {
  ListDatabasesDetailed,
  ListCollections,
  DescribeCollection,
  DropCollection,
  CreateDatabase,
  CreateCollection,
  DropDatabase,
  UpdateDocument,
  DeleteDocument,
} from "../../bindings/vectordb-1/backend/service/vectordbservice";

const databases = ref<DatabaseDetail[]>([]);
const collectionsByDb = ref<Record<string, string[]>>({});
const expandedDbs = ref<Record<string, boolean>>({});

// Multi-Tab Collection Workspace state
const openTabs = ref<CollectionTab[]>([]);
const activeTabId = ref<string>("");

const loading = ref(false);
const loadingCollections = ref<Record<string, boolean>>({});
const error = ref<string>("");

export function useVectorDBStore() {
  const activeTab = computed(() => {
    return openTabs.value.find((t) => t.id === activeTabId.value) || null;
  });

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

  async function refreshDatabase(connId: string, database: string, dbType?: string) {
    if (!connId || !database) return;
    expandedDbs.value[database] = true;
    await loadCollections(connId, database, dbType || getDbType(database));
  }

  async function expandAllDatabases(connId: string) {
    if (!connId) return;
    for (const db of databases.value) {
      expandedDbs.value[db.name] = true;
      if (!collectionsByDb.value[db.name]) {
        await loadCollections(connId, db.name, db.dbType);
      }
    }
  }

  function collapseAllDatabases() {
    for (const db of databases.value) {
      expandedDbs.value[db.name] = false;
    }
  }

  function closeTabsByDatabase(database: string) {
    const toClose = openTabs.value.filter((t) => t.database === database);
    toClose.forEach((t) => closeTab(t.id));
  }

  // Open or switch to a collection tab
  async function openCollectionTab(connId: string, database: string, collection: string) {
    const tabId = `${connId}:${database}:${collection}`;
    const type = getDbType(database);

    const existing = openTabs.value.find((t) => t.id === tabId);
    if (existing) {
      activeTabId.value = tabId;
      return;
    }

    const newTab: CollectionTab = {
      id: tabId,
      connId,
      database,
      collection,
      dbType: type,
      meta: null,
    };

    openTabs.value.push(newTab);
    activeTabId.value = tabId;

    // Asynchronously fetch schema metadata for this tab
    try {
      const meta = await DescribeCollection(connId, database, collection, type);
      const target = openTabs.value.find((t) => t.id === tabId);
      if (target) {
        target.meta = meta;
        if (meta?.isAiCollection) {
          target.dbType = "ai";
          const dbObj = databases.value.find((d) => d.name === database);
          if (dbObj) {
            dbObj.dbType = "ai";
          }
        }
      }
    } catch (e: unknown) {
      console.warn("Failed to describe collection in tab:", e);
    }
  }

  function setActiveTab(tabId: string) {
    activeTabId.value = tabId;
  }

  function closeTab(tabId: string) {
    const idx = openTabs.value.findIndex((t) => t.id === tabId);
    if (idx === -1) return;

    openTabs.value.splice(idx, 1);

    // If closing the active tab, switch to the adjacent tab if any
    if (activeTabId.value === tabId) {
      if (openTabs.value.length === 0) {
        activeTabId.value = "";
      } else {
        const nextIdx = idx < openTabs.value.length ? idx : openTabs.value.length - 1;
        activeTabId.value = openTabs.value[nextIdx].id;
      }
    }
  }

  function closeOtherTabs(tabId: string) {
    openTabs.value = openTabs.value.filter((t) => t.id === tabId);
    activeTabId.value = tabId;
  }

  function closeAllTabs() {
    openTabs.value = [];
    activeTabId.value = "";
  }

  async function deleteCollection(connId: string, database: string, collection: string, dbType?: string) {
    if (!connId || !database || !collection) return;
    const type = dbType || getDbType(database);
    await DropCollection(connId, database, collection, type);

    // Remove from collectionsByDb
    if (collectionsByDb.value[database]) {
      collectionsByDb.value[database] = collectionsByDb.value[database].filter((c) => c !== collection);
    }

    // If any open tab corresponds to this collection, close it
    const matchingTabs = openTabs.value.filter(
      (t) => t.connId === connId && t.database === database && t.collection === collection
    );
    matchingTabs.forEach((t) => closeTab(t.id));
  }

  async function createDatabase(connId: string, database: string, dbType: string) {
    if (!connId || !database) return;
    await CreateDatabase(connId, database, dbType);
    await loadDatabases(connId);
  }

  async function createCollection(
    connId: string,
    database: string,
    collection: string,
    dbType: string,
    options: {
      description?: string;
      replicaNum?: number;
      shardNum?: number;
      indexes?: Array<{
        fieldName: string;
        fieldType: string;
        indexType: string;
        dimension?: number;
        metricType?: string;
      }>;
      language?: string;
    }
  ) {
    if (!connId || !database || !collection) return;
    const type = dbType || getDbType(database);
    await CreateCollection({
      connectionId: connId,
      database,
      dbType: type,
      collection,
      description: options.description || "",
      replicaNum: options.replicaNum ?? 2,
      shardNum: options.shardNum ?? 1,
      indexes: (options.indexes || []).map((idx) => ({
        fieldName: idx.fieldName,
        fieldType: idx.fieldType,
        indexType: idx.indexType,
        dimension: idx.dimension || 0,
        metricType: idx.metricType || "",
      })),
      language: options.language || "",
    });
    await refreshDatabase(connId, database, type);
  }

  async function dropDatabase(connId: string, database: string, dbType?: string) {
    if (!connId || !database) return;
    const type = dbType || getDbType(database);
    await DropDatabase(connId, database, type);

    // Close tabs belonging to the dropped database
    closeTabsByDatabase(database);

    // Remove local state
    delete collectionsByDb.value[database];
    delete expandedDbs.value[database];
    databases.value = databases.value.filter((d) => d.name !== database);
  }

  async function updateDocument(
    connId: string,
    database: string,
    collection: string,
    dbType: string,
    query: { documentIds?: string[]; documentSetIds?: string[]; filter?: string },
    update: Record<string, any>
  ) {
    if (!connId || !database || !collection) return;
    const type = dbType || getDbType(database);
    await UpdateDocument({
      connectionId: connId,
      database,
      collection,
      dbType: type,
      query: {
        documentIds: query.documentIds || [],
        documentSetIds: query.documentSetIds || [],
        filter: query.filter || "",
      },
      update,
    });
  }

  async function deleteDocument(
    connId: string,
    database: string,
    collection: string,
    dbType: string,
    query: { documentIds?: string[]; documentSetIds?: string[]; filter?: string }
  ) {
    if (!connId || !database || !collection) return;
    const type = dbType || getDbType(database);
    await DeleteDocument({
      connectionId: connId,
      database,
      collection,
      dbType: type,
      query: {
        documentIds: query.documentIds || [],
        documentSetIds: query.documentSetIds || [],
        filter: query.filter || "",
      },
    });
  }

  function reset() {
    databases.value = [];
    collectionsByDb.value = {};
    expandedDbs.value = {};
    openTabs.value = [];
    activeTabId.value = "";
    error.value = "";
  }

  return {
    databases,
    collectionsByDb,
    expandedDbs,
    openTabs,
    activeTabId,
    activeTab,
    loading,
    loadingCollections,
    error,
    getDbType,
    loadDatabases,
    loadCollections,
    toggleDatabase,
    refreshDatabase,
    expandAllDatabases,
    collapseAllDatabases,
    closeTabsByDatabase,
    openCollectionTab,
    setActiveTab,
    closeTab,
    closeOtherTabs,
    closeAllTabs,
    deleteCollection,
    updateDocument,
    createDatabase,
    createCollection,
    dropDatabase,
    deleteDocument,
    reset,
  };
}
