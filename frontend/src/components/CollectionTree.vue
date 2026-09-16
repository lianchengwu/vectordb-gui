<template>
  <div class="h-full flex flex-col text-slate-300 select-none text-xs">
    <!-- Search & Refresh Header -->
    <div class="p-2.5 border-b border-slate-800/80 flex items-center gap-1.5">
      <div class="relative flex-1">
        <Search class="w-3.5 h-3.5 text-slate-500 absolute left-2.5 top-1/2 -translate-y-1/2 pointer-events-none" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="搜索 Collection..."
          class="w-full pl-8 pr-2.5 py-1.5 rounded-md bg-slate-800/60 border border-slate-700/60 text-white placeholder-slate-500 text-xs focus:outline-none focus:ring-1 focus:ring-blue-500/50 focus:border-blue-500 transition"
        />
        <button
          v-if="searchQuery"
          @click="searchQuery = ''"
          class="absolute right-2 top-1/2 -translate-y-1/2 text-slate-400 hover:text-white"
        >
          <X class="w-3 h-3" />
        </button>
      </div>

      <button
        @click="refresh"
        :disabled="loading || !activeConnectionId"
        class="p-1.5 rounded-md border border-slate-700/60 bg-slate-800/60 hover:bg-slate-700 text-slate-400 hover:text-white transition disabled:opacity-40 disabled:cursor-not-allowed shrink-0"
        title="刷新数据库与集合"
      >
        <RefreshCw :class="['w-3.5 h-3.5', loading ? 'animate-spin text-blue-400' : '']" />
      </button>
    </div>

    <!-- Tree Content -->
    <div class="flex-1 overflow-y-auto p-2 space-y-1">
      <!-- Loading State -->
      <div v-if="loading && databases.length === 0" class="py-8 text-center text-slate-500 flex flex-col items-center gap-2">
        <Loader2 class="w-5 h-5 animate-spin text-blue-500" />
        <span>加载数据库中...</span>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="p-3 my-2 rounded-lg bg-rose-950/30 border border-rose-800/50 text-rose-300 text-xs">
        <div class="flex items-start gap-2">
          <AlertCircle class="w-4 h-4 text-rose-400 shrink-0 mt-0.5" />
          <div class="flex-1 break-all">
            <p class="font-medium">加载失败</p>
            <p class="text-[11px] text-rose-300/80 mt-0.5">{{ error }}</p>
            <button
              @click="refresh"
              class="mt-2 px-2.5 py-1 rounded bg-rose-900/60 hover:bg-rose-800 text-[11px] text-white transition flex items-center gap-1"
            >
              <RefreshCw class="w-3 h-3" /> 重试
            </button>
          </div>
        </div>
      </div>

      <!-- Empty Databases -->
      <div v-else-if="!activeConnectionId" class="py-12 text-center text-slate-500 px-4">
        <FolderTree class="w-8 h-8 text-slate-600 mx-auto mb-2 opacity-50" />
        <p>请先在上方选择或创建连接</p>
      </div>

      <div v-else-if="databases.length === 0" class="py-12 text-center text-slate-500 px-4">
        <Database class="w-8 h-8 text-slate-600 mx-auto mb-2 opacity-50" />
        <p>该实例下暂无 Database</p>
      </div>

      <!-- Databases & Collections Tree -->
      <div v-else class="space-y-0.5">
        <div v-for="db in filteredDatabases" :key="db" class="rounded-md">
          <!-- Database Node -->
          <div
            @click="onToggleDb(db)"
            class="group px-2 py-1.5 rounded-md hover:bg-slate-800/70 flex items-center justify-between cursor-pointer transition select-none"
          >
            <div class="flex items-center gap-1.5 min-w-0">
              <ChevronRight
                :class="[
                  'w-3.5 h-3.5 text-slate-500 transition-transform duration-150',
                  expandedDbs[db] ? 'rotate-90 text-slate-300' : '',
                ]"
              />
              <Database class="w-3.5 h-3.5 text-blue-400 shrink-0" />
              <span class="font-medium text-slate-200 truncate" :title="db">{{ db }}</span>
            </div>

            <div class="flex items-center gap-1">
              <Loader2 v-if="loadingCollections[db]" class="w-3 h-3 animate-spin text-slate-500" />
              <span
                v-else-if="collectionsByDb[db]"
                class="px-1.5 py-0.5 rounded-full bg-slate-800 text-[10px] text-slate-400 font-mono"
              >
                {{ collectionsByDb[db].length }}
              </span>
            </div>
          </div>

          <!-- Collections (Children) -->
          <div v-if="expandedDbs[db]" class="pl-4 pr-1 py-0.5 space-y-0.5 border-l border-slate-800/80 ml-3.5 my-0.5">
            <div
              v-if="loadingCollections[db] && (!collectionsByDb[db] || collectionsByDb[db].length === 0)"
              class="py-2 px-2 text-[11px] text-slate-500 flex items-center gap-1.5"
            >
              <Loader2 class="w-3 h-3 animate-spin text-blue-400" />
              <span>加载集合...</span>
            </div>

            <div
              v-else-if="!collectionsByDb[db] || collectionsByDb[db].length === 0"
              class="py-1.5 px-2 text-[11px] text-slate-500 italic"
            >
              暂无集合
            </div>

            <div
              v-else
              v-for="coll in getFilteredCollections(db)"
              :key="coll"
              @click="onSelectColl(db, coll)"
              :class="[
                'px-2 py-1.5 rounded-md flex items-center gap-2 cursor-pointer transition select-none group',
                activeDatabase === db && activeCollection === coll
                  ? 'bg-blue-600/20 text-blue-300 border border-blue-500/30 font-medium'
                  : 'hover:bg-slate-800/60 text-slate-400 hover:text-slate-200',
              ]"
            >
              <Table class="w-3.5 h-3.5 text-indigo-400 shrink-0 group-hover:text-blue-400" />
              <span class="truncate" :title="coll">{{ coll }}</span>
            </div>

            <div
              v-if="collectionsByDb[db] && collectionsByDb[db].length > 0 && getFilteredCollections(db).length === 0"
              class="py-1.5 px-2 text-[11px] text-slate-500 italic"
            >
              无匹配集合
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import {
  Search,
  RefreshCw,
  X,
  Database,
  ChevronRight,
  Table,
  FolderTree,
  AlertCircle,
  Loader2,
} from "lucide-vue-next";
import { useVectorDBStore } from "../stores/vectordb";
import { useConnectionStore } from "../stores/connection";

const vdbStore = useVectorDBStore();
const connStore = useConnectionStore();

const {
  databases,
  collectionsByDb,
  expandedDbs,
  activeDatabase,
  activeCollection,
  loading,
  loadingCollections,
  error,
  loadDatabases,
  toggleDatabase,
  selectCollection,
} = vdbStore;

const { activeConnectionId } = connStore;

const searchQuery = ref("");

const filteredDatabases = computed(() => {
  if (!searchQuery.value.trim()) return databases.value;
  const q = searchQuery.value.toLowerCase();
  return databases.value.filter((db) => {
    if (db.toLowerCase().includes(q)) return true;
    const colls = collectionsByDb.value[db] || [];
    return colls.some((c) => c.toLowerCase().includes(q));
  });
});

function getFilteredCollections(db: string): string[] {
  const colls = collectionsByDb.value[db] || [];
  if (!searchQuery.value.trim()) return colls;
  const q = searchQuery.value.toLowerCase();
  return colls.filter((c) => c.toLowerCase().includes(q));
}

function onToggleDb(db: string) {
  if (activeConnectionId.value) {
    toggleDatabase(activeConnectionId.value, db);
  }
}

function onSelectColl(db: string, coll: string) {
  if (activeConnectionId.value) {
    selectCollection(activeConnectionId.value, db, coll);
  }
}

function refresh() {
  if (activeConnectionId.value) {
    loadDatabases(activeConnectionId.value);
  }
}
</script>
