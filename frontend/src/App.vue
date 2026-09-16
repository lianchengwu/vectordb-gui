<template>
  <div class="flex h-screen w-screen bg-slate-950 text-slate-100 overflow-hidden font-sans antialiased">
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

    <!-- Global Connection Edit/Create Modal -->
    <ConnectionModal />
  </div>
</template>

<script setup lang="ts">
import { onMounted, watch } from "vue";
import Sidebar from "./components/Sidebar.vue";
import CollectionTree from "./components/CollectionTree.vue";
import MainView from "./components/MainView.vue";
import ConnectionModal from "./components/ConnectionModal.vue";
import { useConnectionStore } from "./stores/connection";
import { useVectorDBStore } from "./stores/vectordb";

const connStore = useConnectionStore();
const vdbStore = useVectorDBStore();

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
  await loadConnections();
  if (activeConnectionId.value) {
    await loadDatabases(activeConnectionId.value);
  }
});
</script>
