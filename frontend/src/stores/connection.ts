import { ref, computed } from "vue";
import type { ConnectionConfig, TestResult } from "../types";
import {
  ListConnections,
  SaveConnection,
  DeleteConnection,
  TestConnection,
} from "../../bindings/vectordb-1/backend/service/connectionservice";

const connections = ref<ConnectionConfig[]>([]);
const activeConnectionId = ref<string>("");
const isModalOpen = ref(false);
const editingConnection = ref<ConnectionConfig | null>(null);
const loading = ref(false);
const error = ref<string>("");

export function useConnectionStore() {
  const activeConnection = computed(() =>
    connections.value.find((c) => c.id === activeConnectionId.value) || null
  );

  async function loadConnections() {
    loading.value = true;
    error.value = "";
    try {
      const list = await ListConnections();
      connections.value = list || [];
      if (connections.value.length > 0 && !activeConnectionId.value) {
        activeConnectionId.value = connections.value[0].id;
      } else if (
        activeConnectionId.value &&
        !connections.value.some((c) => c.id === activeConnectionId.value)
      ) {
        activeConnectionId.value = connections.value[0]?.id || "";
      }
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : String(e);
    } finally {
      loading.value = false;
    }
  }

  function setActiveConnection(id: string) {
    activeConnectionId.value = id;
  }

  function openCreateModal() {
    editingConnection.value = {
      id: "",
      name: "",
      url: "http://",
      username: "root",
      apiKey: "",
      timeout: 10,
    };
    isModalOpen.value = true;
  }

  function openEditModal(conn: ConnectionConfig) {
    editingConnection.value = { ...conn };
    isModalOpen.value = true;
  }

  function closeModal() {
    isModalOpen.value = false;
    editingConnection.value = null;
  }

  async function save(config: ConnectionConfig) {
    loading.value = true;
    error.value = "";
    try {
      await SaveConnection(config);
      await loadConnections();
      if (!activeConnectionId.value || config.id === activeConnectionId.value) {
        activeConnectionId.value = config.id || connections.value[connections.value.length - 1]?.id || "";
      }
      closeModal();
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : String(e);
      throw e;
    } finally {
      loading.value = false;
    }
  }

  async function remove(id: string) {
    if (!id) return;
    loading.value = true;
    error.value = "";
    try {
      await DeleteConnection(id);
      if (activeConnectionId.value === id) {
        activeConnectionId.value = "";
      }
      await loadConnections();
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : String(e);
      throw e;
    } finally {
      loading.value = false;
    }
  }

  async function test(config: ConnectionConfig): Promise<TestResult> {
    try {
      const res = await TestConnection(config);
      return {
        success: !!res?.success,
        message: res?.message || (res?.success ? "Connection successful" : "Connection failed"),
      };
    } catch (e: unknown) {
      return {
        success: false,
        message: e instanceof Error ? e.message : String(e),
      };
    }
  }

  return {
    connections,
    activeConnectionId,
    activeConnection,
    isModalOpen,
    editingConnection,
    loading,
    error,
    loadConnections,
    setActiveConnection,
    openCreateModal,
    openEditModal,
    closeModal,
    save,
    remove,
    test,
  };
}
