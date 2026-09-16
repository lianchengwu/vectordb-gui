<template>
  <main class="flex-1 h-full flex flex-col bg-slate-950 overflow-hidden text-slate-100 select-none">
    <!-- Empty / No Collection Selected State -->
    <div
      v-if="!collection"
      class="flex-1 flex flex-col items-center justify-center p-8 text-center"
    >
      <div class="w-16 h-16 rounded-2xl bg-slate-900 border border-slate-800 flex items-center justify-center text-slate-600 mb-4 shadow-xl">
        <DatabaseZap class="w-8 h-8 text-blue-500/60" />
      </div>
      <h2 class="text-base font-semibold text-white tracking-tight">Tencent Cloud VectorDB Client</h2>
      <p class="text-xs text-slate-400 mt-1.5 max-w-sm leading-relaxed">
        请在左侧侧边栏选择目标 Database 与 Collection，即可浏览集合 Schema 结构与文档高维向量数据。
      </p>
    </div>

    <!-- Active Collection Workspace -->
    <template v-else>
      <!-- Top Navigation & Header -->
      <header class="h-14 px-6 border-b border-slate-800 bg-slate-900/60 flex items-center justify-between shrink-0">
        <!-- Breadcrumb & Metadata badges -->
        <div class="flex items-center gap-3 min-w-0">
          <div class="flex items-center gap-1.5 text-xs font-mono text-slate-400 truncate">
            <span class="text-slate-300 font-medium">{{ database }}</span>
            <span class="text-slate-600">/</span>
            <span class="text-white font-bold text-sm tracking-tight font-sans">{{ collection }}</span>
          </div>

          <!-- Quick badges -->
          <div v-if="collectionMeta" class="hidden sm:flex items-center gap-1.5 ml-2">
            <span
              v-if="vectorMetric"
              class="px-2 py-0.5 rounded-full text-[10px] font-mono font-medium bg-emerald-950/70 border border-emerald-700/60 text-emerald-300"
              title="向量索引度量方式"
            >
              {{ vectorMetric }}
            </span>
            <span
              class="px-2 py-0.5 rounded-full text-[10px] font-mono bg-slate-800 border border-slate-700 text-slate-400"
              title="分片数"
            >
              {{ collectionMeta.shardNum || 1 }} 分片
            </span>
            <span
              class="px-2 py-0.5 rounded-full text-[10px] font-mono bg-slate-800 border border-slate-700 text-slate-400"
              title="副本数"
            >
              {{ collectionMeta.replicaNum || 0 }} 副本
            </span>
          </div>
        </div>

        <!-- Tab Switcher -->
        <div class="flex items-center bg-slate-800/80 p-0.5 rounded-lg border border-slate-700/80 text-xs shrink-0">
          <button
            @click="activeTab = 'data'"
            :class="[
              'px-3.5 py-1 rounded-md font-medium transition flex items-center gap-1.5',
              activeTab === 'data'
                ? 'bg-blue-600 text-white shadow-sm'
                : 'text-slate-400 hover:text-white',
            ]"
          >
            <Table2 class="w-3.5 h-3.5" />
            <span>数据浏览</span>
          </button>
          <button
            @click="activeTab = 'schema'"
            :class="[
              'px-3.5 py-1 rounded-md font-medium transition flex items-center gap-1.5',
              activeTab === 'schema'
                ? 'bg-blue-600 text-white shadow-sm'
                : 'text-slate-400 hover:text-white',
            ]"
          >
            <Layers class="w-3.5 h-3.5" />
            <span>表结构 (Schema)</span>
          </button>
        </div>
      </header>

      <!-- Content Area -->
      <div class="flex-1 overflow-hidden">
        <DataExplorer
          v-if="activeTab === 'data'"
          :conn-id="connId"
          :database="database"
          :collection="collection"
          :collection-meta="collectionMeta"
        />

        <SchemaViewer
          v-else-if="activeTab === 'schema'"
          :meta="collectionMeta"
        />
      </div>
    </template>
  </main>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { DatabaseZap, Table2, Layers } from "lucide-vue-next";
import type { CollectionMeta } from "../types";
import DataExplorer from "./DataExplorer.vue";
import SchemaViewer from "./SchemaViewer.vue";

const props = defineProps<{
  connId: string;
  database: string;
  collection: string;
  collectionMeta: CollectionMeta | null;
}>();

const activeTab = ref<"data" | "schema">("data");

const vectorMetric = computed(() => {
  if (!props.collectionMeta?.indexes) return null;
  const vectorIndex = props.collectionMeta.indexes.find(
    (idx) => idx.indexType?.toLowerCase().includes("vector") || idx.metricType
  );
  return vectorIndex?.metricType || null;
});

// Reset tab to data view when switching collections
watch(
  () => [props.database, props.collection],
  () => {
    activeTab.value = "data";
  }
);
</script>
