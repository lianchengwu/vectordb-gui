<template>
  <div class="h-full flex flex-col text-slate-600 dark:text-slate-300 select-none text-xs transition-colors duration-200">
    <!-- Search & Refresh Header -->
    <div class="p-2.5 border-b border-slate-200 dark:border-slate-800/80 flex items-center gap-1.5">
      <div class="relative flex-1">
        <Search class="w-3.5 h-3.5 text-slate-400 dark:text-slate-500 absolute left-2.5 top-1/2 -translate-y-1/2 pointer-events-none" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="搜索 Collection..."
          class="w-full pl-8 pr-2.5 py-1.5 rounded-md bg-slate-100 border border-slate-200 text-slate-900 placeholder-slate-400 dark:bg-slate-800/60 dark:border-slate-700/60 dark:text-white dark:placeholder-slate-500 text-xs focus:outline-none focus:ring-1 focus:ring-blue-500/50 focus:border-blue-500 transition"
        />
        <button
          v-if="searchQuery"
          @click="searchQuery = ''"
          class="absolute right-2 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-700 dark:hover:text-white"
        >
          <X class="w-3 h-3" />
        </button>
      </div>

      <button
        @click="refresh"
        :disabled="loading || !activeConnectionId"
        class="p-1.5 rounded-md border border-slate-200 bg-white hover:bg-slate-100 text-slate-600 dark:border-slate-700/60 dark:bg-slate-800/60 dark:hover:bg-slate-700 dark:text-slate-400 dark:hover:text-white transition disabled:opacity-40 disabled:cursor-not-allowed shrink-0 shadow-sm"
        title="刷新数据库与集合"
      >
        <RefreshCw :class="['w-3.5 h-3.5', loading ? 'animate-spin text-blue-500' : '']" />
      </button>
    </div>

    <!-- Tree Content -->
    <div class="flex-1 overflow-y-auto p-2 space-y-1">
      <!-- Loading State -->
      <div v-if="loading && databases.length === 0" class="py-8 text-center text-slate-400 dark:text-slate-500 flex flex-col items-center gap-2">
        <Loader2 class="w-5 h-5 animate-spin text-blue-500" />
        <span>加载数据库中...</span>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="p-3 my-2 rounded-lg bg-rose-50 border border-rose-200 text-rose-700 dark:bg-rose-950/30 dark:border-rose-800/50 dark:text-rose-300 text-xs">
        <div class="flex items-start gap-2">
          <AlertCircle class="w-4 h-4 text-rose-500 shrink-0 mt-0.5" />
          <div class="flex-1 break-all">
            <p class="font-medium">加载失败</p>
            <p class="text-[11px] text-rose-600 dark:text-rose-300/80 mt-0.5">{{ error }}</p>
            <button
              @click="refresh"
              class="mt-2 px-2.5 py-1 rounded bg-rose-600 hover:bg-rose-700 text-[11px] text-white transition flex items-center gap-1 shadow-sm"
            >
              <RefreshCw class="w-3 h-3" /> 重试
            </button>
          </div>
        </div>
      </div>

      <!-- Empty Databases -->
      <div v-else-if="!activeConnectionId" class="py-12 text-center text-slate-400 dark:text-slate-500 px-4">
        <FolderTree class="w-8 h-8 text-slate-300 dark:text-slate-600 mx-auto mb-2 opacity-50" />
        <p>请先在上方选择或创建连接</p>
      </div>

      <div v-else-if="databases.length === 0" class="py-12 text-center text-slate-400 dark:text-slate-500 px-4">
        <Database class="w-8 h-8 text-slate-300 dark:text-slate-600 mx-auto mb-2 opacity-50" />
        <p>该实例下暂无 Database</p>
      </div>

      <!-- Databases & Collections Tree -->
      <div v-else class="space-y-0.5">
        <div v-for="db in filteredDatabases" :key="db.name" class="rounded-md">
          <!-- Database Node -->
          <div
            @click="onToggleDb(db.name)"
            class="group px-2 py-1.5 rounded-md hover:bg-slate-100 dark:hover:bg-slate-800/70 flex items-center justify-between cursor-pointer transition select-none"
          >
            <div class="flex items-center gap-1.5 min-w-0">
              <ChevronRight
                :class="[
                  'w-3.5 h-3.5 text-slate-400 dark:text-slate-500 transition-transform duration-150',
                  expandedDbs[db.name] ? 'rotate-90 text-slate-700 dark:text-slate-300' : '',
                ]"
              />
              <Database :class="['w-3.5 h-3.5 shrink-0', db.dbType === 'ai' ? 'text-purple-500 dark:text-purple-400' : 'text-blue-500 dark:text-blue-400']" />
              <span class="font-medium text-slate-800 dark:text-slate-200 truncate" :title="db.name">{{ db.name }}</span>

              <!-- DB Type Badge (Base vs AI) -->
              <span
                :class="[
                  'px-1.5 py-0.2 rounded text-[9px] font-semibold tracking-wider shrink-0 font-mono',
                  db.dbType === 'ai'
                    ? 'bg-purple-100 text-purple-700 dark:bg-purple-950 dark:text-purple-300 border border-purple-200 dark:border-purple-800'
                    : 'bg-blue-100 text-blue-700 dark:bg-blue-950 dark:text-blue-300 border border-blue-200 dark:border-blue-800'
                ]"
              >
                {{ db.dbType === 'ai' ? 'AI' : 'BASE' }}
              </span>
            </div>

            <div class="flex items-center gap-1">
              <Loader2 v-if="loadingCollections[db.name]" class="w-3 h-3 animate-spin text-slate-400 dark:text-slate-500" />
              <span
                v-else-if="collectionsByDb[db.name]"
                class="px-1.5 py-0.5 rounded-full bg-slate-200/80 text-[10px] text-slate-600 dark:bg-slate-800 dark:text-slate-400 font-mono"
              >
                {{ collectionsByDb[db.name].length }}
              </span>
            </div>
          </div>

          <!-- Collections (Children) -->
          <div v-if="expandedDbs[db.name]" class="pl-4 pr-1 py-0.5 space-y-0.5 border-l border-slate-200 dark:border-slate-800/80 ml-3.5 my-0.5">
            <div
              v-if="loadingCollections[db.name] && (!collectionsByDb[db.name] || collectionsByDb[db.name].length === 0)"
              class="py-2 px-2 text-[11px] text-slate-400 dark:text-slate-500 flex items-center gap-1.5"
            >
              <Loader2 class="w-3 h-3 animate-spin text-blue-500 dark:text-blue-400" />
              <span>加载集合...</span>
            </div>

            <div
              v-else-if="!collectionsByDb[db.name] || collectionsByDb[db.name].length === 0"
              class="py-1.5 px-2 text-[11px] text-slate-400 dark:text-slate-500 italic"
            >
              暂无集合
            </div>

            <div
              v-else
              v-for="coll in getFilteredCollections(db.name)"
              :key="coll"
              @click="onSelectColl(db.name, coll)"
              :class="[
                'px-2 py-1.5 rounded-md flex items-center gap-2 cursor-pointer transition select-none group',
                activeTab && activeTab.database === db.name && activeTab.collection === coll
                  ? 'bg-blue-50 text-blue-600 border border-blue-200 font-medium dark:bg-blue-600/20 dark:text-blue-300 dark:border-blue-500/30'
                  : 'hover:bg-slate-100 text-slate-600 hover:text-slate-900 dark:hover:bg-slate-800/60 dark:text-slate-400 dark:hover:text-slate-200',
              ]"
            >
              <BookOpen v-if="db.dbType === 'ai'" class="w-3.5 h-3.5 text-purple-500 dark:text-purple-400 shrink-0 group-hover:text-purple-600" />
              <Table v-else class="w-3.5 h-3.5 text-indigo-500 dark:text-indigo-400 shrink-0 group-hover:text-blue-500" />
              <span class="truncate" :title="coll">{{ coll }}</span>
            </div>

            <div
              v-if="collectionsByDb[db.name] && collectionsByDb[db.name].length > 0 && getFilteredCollections(db.name).length === 0"
              class="py-1.5 px-2 text-[11px] text-slate-400 dark:text-slate-500 italic"
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
  BookOpen,
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
  activeTab,
  loading,
  loadingCollections,
  error,
  loadDatabases,
  toggleDatabase,
  openCollectionTab,
} = vdbStore;
const { activeConnectionId } = connStore;

const searchQuery = ref("");

const filteredDatabases = computed(() => {
  if (!searchQuery.value.trim()) return databases.value;
  const q = searchQuery.value.toLowerCase();
  return databases.value.filter((db) => {
    if (db.name.toLowerCase().includes(q)) return true;
    const colls = collectionsByDb.value[db.name] || [];
    return colls.some((c) => c.toLowerCase().includes(q));
  });
});

function getFilteredCollections(dbName: string): string[] {
  const colls = collectionsByDb.value[dbName] || [];
  if (!searchQuery.value.trim()) return colls;
  const q = searchQuery.value.toLowerCase();
  return colls.filter((c) => c.toLowerCase().includes(q));
}

function onToggleDb(dbName: string) {
  if (activeConnectionId.value) {
    toggleDatabase(activeConnectionId.value, dbName);
  }
}

function onSelectColl(dbName: string, coll: string) {
  if (activeConnectionId.value) {
    openCollectionTab(activeConnectionId.value, dbName, coll);
  }
}

function refresh() {
  if (activeConnectionId.value) {
    loadDatabases(activeConnectionId.value);
  }
}
</script>
