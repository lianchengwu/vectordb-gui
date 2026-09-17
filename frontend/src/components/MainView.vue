<template>
  <main class="flex-1 h-full flex flex-col bg-slate-50 dark:bg-slate-950 overflow-hidden text-slate-800 dark:text-slate-100 select-none transition-colors duration-200">
    <!-- Horizontal Collection Tabs Bar -->
    <div
      v-if="openTabs.length > 0"
      class="h-9 px-2 bg-slate-200/80 dark:bg-slate-900 border-b border-slate-300 dark:border-slate-800 flex items-center justify-between shrink-0 overflow-hidden"
    >
      <!-- Scrollable Tabs List -->
      <div class="flex items-center gap-1 overflow-x-auto h-full flex-1 no-scrollbar py-0.5">
        <div
          v-for="tab in openTabs"
          :key="tab.id"
          @click="setActiveTab(tab.id)"
          :class="[
            'group h-7 px-2.5 rounded-md flex items-center gap-1.5 cursor-pointer text-xs font-medium transition shrink-0 select-none border',
            tab.id === activeTabId
              ? 'bg-white dark:bg-slate-950 text-slate-900 dark:text-white border-slate-300/80 dark:border-slate-800 shadow-sm'
              : 'border-transparent text-slate-600 dark:text-slate-400 hover:bg-slate-300/60 dark:hover:bg-slate-800/70 hover:text-slate-900 dark:hover:text-slate-200',
          ]"
          :title="`${tab.database} / ${tab.collection}`"
        >
          <!-- Tab Type Icon -->
          <BookOpen v-if="tab.dbType === 'ai'" class="w-3 h-3 text-purple-500 shrink-0" />
          <Table v-else class="w-3 h-3 text-blue-500 shrink-0" />

          <!-- Tab Label (Collection Name with DB hint) -->
          <span class="truncate max-w-[140px]">{{ tab.collection }}</span>
          <span class="text-[10px] text-slate-400 font-mono font-normal">({{ tab.database }})</span>

          <!-- Close Tab Button -->
          <button
            type="button"
            @click.stop="closeTab(tab.id)"
            class="p-0.5 rounded text-slate-400 hover:text-rose-600 dark:hover:text-rose-400 hover:bg-slate-200 dark:hover:bg-slate-800 transition ml-0.5 opacity-60 group-hover:opacity-100"
            title="关闭标签页"
          >
            <X class="w-3 h-3" />
          </button>
        </div>
      </div>

      <!-- Quick Tab Actions -->
      <div v-if="openTabs.length > 1" class="flex items-center gap-1 shrink-0 ml-2 pl-2 border-l border-slate-300 dark:border-slate-800">
        <button
          type="button"
          @click="closeAllTabs"
          class="px-1.5 py-1 rounded text-[11px] text-slate-500 hover:text-rose-600 dark:text-slate-400 dark:hover:text-rose-400 hover:bg-slate-300/50 dark:hover:bg-slate-800 transition"
          title="关闭全部标签页"
        >
          全部关闭
        </button>
      </div>
    </div>

    <!-- Empty State: No Tabs Open -->
    <div
      v-if="!activeTab"
      class="flex-1 flex flex-col items-center justify-center p-8 text-center"
    >
      <div class="w-16 h-16 rounded-2xl bg-white border border-slate-200 dark:bg-slate-900 dark:border-slate-800 flex items-center justify-center text-slate-400 dark:text-slate-600 mb-4 shadow-md dark:shadow-xl">
        <DatabaseZap class="w-8 h-8 text-blue-500/70" />
      </div>
      <h2 class="text-base font-semibold text-slate-900 dark:text-white tracking-tight">Tencent Cloud VectorDB Client</h2>
      <p class="text-xs text-slate-500 dark:text-slate-400 mt-1.5 max-w-sm leading-relaxed">
        请在左侧侧边栏点击任意 Database 下的 Collection，即可打开标签页并支持多个 Collection 之间并行切换浏览。
      </p>
    </div>

    <!-- Active Tab Workspace -->
    <template v-else>
      <!-- Top Navigation & Header for Active Collection -->
      <header class="h-12 px-6 border-b border-slate-200 bg-white/90 dark:border-slate-800 dark:bg-slate-900/60 flex items-center justify-between shrink-0 shadow-sm dark:shadow-none">
        <!-- Breadcrumb & Metadata badges -->
        <div class="flex items-center gap-3 min-w-0">
          <div class="flex items-center gap-1.5 text-xs font-mono text-slate-500 dark:text-slate-400 truncate">
            <span class="text-slate-700 dark:text-slate-300 font-medium">{{ activeTab.database }}</span>
            <span class="text-slate-400 dark:text-slate-600">/</span>
            <span class="text-slate-900 dark:text-white font-bold text-sm tracking-tight font-sans">{{ activeTab.collection }}</span>
          </div>

          <!-- Quick badges -->
          <div class="hidden sm:flex items-center gap-1.5 ml-2">
            <!-- DB Type badge -->
            <span
              :class="[
                'px-2 py-0.5 rounded-full text-[10px] font-mono font-bold',
                activeTab.dbType === 'ai'
                  ? 'bg-purple-100 text-purple-700 dark:bg-purple-950 dark:text-purple-300 border border-purple-200 dark:border-purple-800'
                  : 'bg-blue-100 text-blue-700 dark:bg-blue-950 dark:text-blue-300 border border-blue-200 dark:border-blue-800'
              ]"
            >
              {{ activeTab.dbType === 'ai' ? 'AI 知识库' : 'BASE 库' }}
            </span>

            <template v-if="activeTab.meta">
              <span
                v-if="vectorMetric"
                class="px-2 py-0.5 rounded-full text-[10px] font-mono font-medium bg-emerald-50 border border-emerald-200 text-emerald-700 dark:bg-emerald-950/70 dark:border-emerald-700/60 dark:text-emerald-300"
                title="向量索引度量方式"
              >
                {{ vectorMetric }}
              </span>
              <span
                class="px-2 py-0.5 rounded-full text-[10px] font-mono bg-slate-100 border border-slate-200 text-slate-600 dark:bg-slate-800 dark:border-slate-700 dark:text-slate-400"
                title="分片数"
              >
                {{ activeTab.meta.shardNum || 1 }} 分片
              </span>
              <span
                class="px-2 py-0.5 rounded-full text-[10px] font-mono bg-slate-100 border border-slate-200 text-slate-600 dark:bg-slate-800 dark:border-slate-700 dark:text-slate-400"
                title="副本数"
              >
                {{ activeTab.meta.replicaNum || 0 }} 副本
              </span>
            </template>
          </div>
        </div>

        <!-- Sub-Tab Switcher: Data vs Schema -->
        <div class="flex items-center bg-slate-100 dark:bg-slate-800/80 p-0.5 rounded-lg border border-slate-200 dark:border-slate-700/80 text-xs shrink-0">
          <button
            @click="activeSubTab = 'data'"
            :class="[
              'px-3.5 py-1 rounded-md font-medium transition flex items-center gap-1.5',
              activeSubTab === 'data'
                ? 'bg-blue-600 text-white shadow-sm'
                : 'text-slate-600 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white',
            ]"
          >
            <Table2 class="w-3.5 h-3.5" />
            <span>数据浏览</span>
          </button>
          <button
            @click="activeSubTab = 'schema'"
            :class="[
              'px-3.5 py-1 rounded-md font-medium transition flex items-center gap-1.5',
              activeSubTab === 'schema'
                ? 'bg-blue-600 text-white shadow-sm'
                : 'text-slate-600 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white',
            ]"
          >
            <Layers class="w-3.5 h-3.5" />
            <span>表结构 (Schema)</span>
          </button>
        </div>
      </header>

      <!-- Content Area (Keyed by activeTab.id for isolated state) -->
      <div class="flex-1 overflow-hidden">
        <DataExplorer
          v-if="activeSubTab === 'data'"
          :key="`${activeTab.id}:data`"
          :conn-id="activeTab.connId"
          :database="activeTab.database"
          :collection="activeTab.collection"
          :db-type="activeTab.dbType"
          :collection-meta="activeTab.meta"
        />

        <SchemaViewer
          v-else-if="activeSubTab === 'schema'"
          :key="`${activeTab.id}:schema`"
          :meta="activeTab.meta"
        />
      </div>
    </template>
  </main>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { DatabaseZap, Table2, Layers, BookOpen, Table, X } from "lucide-vue-next";
import { useVectorDBStore } from "../stores/vectordb";
import DataExplorer from "./DataExplorer.vue";
import SchemaViewer from "./SchemaViewer.vue";

const vdbStore = useVectorDBStore();
const {
  openTabs,
  activeTabId,
  activeTab,
  setActiveTab,
  closeTab,
  closeAllTabs,
} = vdbStore;

const activeSubTab = ref<"data" | "schema">("data");

const vectorMetric = computed(() => {
  if (!activeTab.value?.meta?.indexes) return null;
  const vectorIndex = activeTab.value.meta.indexes.find(
    (idx) => idx.indexType?.toLowerCase().includes("vector") || idx.metricType
  );
  return vectorIndex?.metricType || null;
});

// Reset sub-tab to data view when active collection tab changes
watch(
  () => activeTabId.value,
  () => {
    activeSubTab.value = "data";
  }
);
</script>

<style scoped>
.no-scrollbar::-webkit-scrollbar {
  display: none;
}
.no-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
</style>
