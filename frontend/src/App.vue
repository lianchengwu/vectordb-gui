<template>
  <div class="flex flex-col h-screen w-screen bg-slate-100 text-slate-900 dark:bg-slate-950 dark:text-slate-100 overflow-hidden font-sans antialiased transition-colors duration-200">
    <!-- Custom Frameless Window TitleBar with Minimize, Maximize, Close -->
    <TitleBar />

    <!-- Main Workspace Area -->
    <div class="flex flex-1 overflow-hidden">
      <!-- Sidebar with Collection Tree embedded in slot -->
      <Sidebar @connection-changed="onConnectionChanged">
        <template #tree>
          <CollectionTree />
        </template>
      </Sidebar>

      <!-- Main Content Workspace -->
      <MainView
        :conn-id="activeConnectionId"
        :database="activeDatabase"
        :collection="activeCollection"
        :collection-meta="activeCollectionMeta"
      />
    </div>

    <!-- Global Connection Edit/Create Modal -->
    <ConnectionModal />

    <!-- Global Update Check Modal -->
    <UpdateModal />
  </div>
</template>

<script setup lang="ts">
import { onMounted, watch } from "vue";
import TitleBar from "./components/TitleBar.vue";
import Sidebar from "./components/Sidebar.vue";
import CollectionTree from "./components/CollectionTree.vue";
import MainView from "./components/MainView.vue";
import ConnectionModal from "./components/ConnectionModal.vue";
import UpdateModal from "./components/UpdateModal.vue";
import { useConnectionStore } from "./stores/connection";
import { useVectorDBStore } from "./stores/vectordb";
import { useThemeStore } from "./stores/theme";

const connStore = useConnectionStore();
const vdbStore = useVectorDBStore();
const themeStore = useThemeStore();

const { activeConnectionId, loadConnections } = connStore;
const {
  activeDatabase,
  activeCollection,
  activeCollectionMeta,
  loadDatabases,
  reset,
} = vdbStore;

function onConnectionChanged(connId: string) {
  if (connId) {
    loadDatabases(connId);
  } else {
    reset();
  }
}

watch(
  () => activeConnectionId.value,
  (newId) => {
    if (newId) {
      loadDatabases(newId);
    } else {
      reset();
    }
  }
);

onMounted(async () => {
  themeStore.initTheme();
  await loadConnections();
  if (activeConnectionId.value) {
    await loadDatabases(activeConnectionId.value);
  }
});
</script>
