import { ref, computed } from "vue";
import type { CollectionMeta, DatabaseDetail, CollectionTab } from "../types";
import {
  ListDatabasesDetailed,
  ListCollections,
  DescribeCollection,
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
    openCollectionTab,
    setActiveTab,
    closeTab,
    closeOtherTabs,
    closeAllTabs,
    reset,
  };
}
