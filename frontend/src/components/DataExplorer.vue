<template>
  <div class="h-full flex flex-col text-slate-800 dark:text-slate-200 select-none text-xs transition-colors duration-200">
    <!-- Toolbar: Filter bar & Pagination controls -->
    <div class="p-3 border-b border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-950/40 flex flex-wrap items-center justify-between gap-3 shrink-0 shadow-sm dark:shadow-none">
      <!-- Filter Bar -->
      <div class="flex items-center gap-2 flex-1 min-w-[300px]">
        <div class="relative flex-1">
          <Filter class="w-3.5 h-3.5 text-slate-400 dark:text-slate-500 absolute left-2.5 top-1/2 -translate-y-1/2 pointer-events-none" />
          <input
            v-model="filterInput"
            @keydown.enter="onQuery"
            type="text"
            placeholder='过滤条件，例如: id in ("doc-1") 或 age > 18'
            class="w-full pl-8 pr-8 py-1.5 rounded-lg bg-slate-50 border border-slate-300 text-slate-900 placeholder-slate-400 dark:bg-slate-900 dark:border-slate-700 dark:text-white dark:placeholder-slate-500 text-xs focus:outline-none focus:ring-1 focus:ring-blue-500 transition font-mono shadow-inner"
          />
          <button
            v-if="filterInput"
            @click="onClearFilter"
            class="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-700 dark:hover:text-white"
            title="清空过滤条件"
          >
            <X class="w-3.5 h-3.5" />
          </button>
        </div>

        <button
          @click="onQuery"
          :disabled="loading"
          class="px-3.5 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-500 text-white font-medium transition flex items-center gap-1.5 disabled:opacity-50 shadow-sm"
        >
          <Loader2 v-if="loading" class="w-3.5 h-3.5 animate-spin" />
          <Search v-else class="w-3.5 h-3.5" />
          查询
        </button>

        <button
          @click="fetchData"
          :disabled="loading"
          class="p-1.5 rounded-lg border border-slate-200 bg-white hover:bg-slate-100 text-slate-600 dark:border-slate-700 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-300 transition shadow-sm"
          title="刷新数据"
        >
          <RefreshCw :class="['w-3.5 h-3.5', loading ? 'animate-spin text-blue-500' : '']" />
        </button>
      </div>

      <!-- Pagination Bar -->
      <div class="flex items-center gap-3 shrink-0">
        <!-- Limit Selector -->
        <div class="flex items-center gap-1.5 text-slate-500 dark:text-slate-400">
          <span>每页</span>
          <select
            v-model.number="limit"
            @change="onLimitChange"
            class="px-2 py-1 rounded bg-white border border-slate-300 text-slate-800 dark:bg-slate-800 dark:border-slate-700 dark:text-white text-xs font-mono focus:outline-none shadow-sm"
          >
            <option :value="10">10</option>
            <option :value="20">20</option>
            <option :value="50">50</option>
            <option :value="100">100</option>
          </select>
          <span>条</span>
        </div>

        <!-- Page Controls -->
        <div class="flex items-center gap-1">
          <button
            @click="prevPage"
            :disabled="offset === 0 || loading"
            class="px-2.5 py-1 rounded border border-slate-200 bg-white hover:bg-slate-100 text-slate-700 dark:border-slate-700 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-300 transition disabled:opacity-30 disabled:cursor-not-allowed flex items-center gap-1 shadow-sm"
          >
            <ChevronLeft class="w-3.5 h-3.5" /> 上一页
          </button>

          <span class="px-2 py-1 text-slate-500 dark:text-slate-400 font-mono text-[11px]">
            {{ offset + 1 }} - {{ offset + documents.length }}
            <span v-if="totalCount > 0" class="text-slate-400 dark:text-slate-500">/ {{ totalCount }}</span>
          </span>

          <button
            @click="nextPage"
            :disabled="documents.length < limit || loading"
            class="px-2.5 py-1 rounded border border-slate-200 bg-white hover:bg-slate-100 text-slate-700 dark:border-slate-700 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-300 transition disabled:opacity-30 disabled:cursor-not-allowed flex items-center gap-1 shadow-sm"
          >
            下一页 <ChevronRight class="w-3.5 h-3.5" />
          </button>
        </div>
      </div>
    </div>

    <!-- Error Alert -->
    <div v-if="error" class="m-4 p-3 rounded-lg bg-rose-50 border border-rose-200 text-rose-700 dark:bg-rose-950/40 dark:border-rose-800/60 dark:text-rose-300 flex items-start gap-2 text-xs shrink-0 shadow-sm">
      <AlertCircle class="w-4 h-4 text-rose-500 dark:text-rose-400 shrink-0 mt-0.5" />
      <div class="flex-1 break-all">
        <p class="font-medium">查询返回错误</p>
        <p class="text-[11px] text-rose-600 dark:text-rose-300/80 mt-0.5">{{ error }}</p>
      </div>
    </div>

    <!-- Data Table Area -->
    <div class="flex-1 overflow-auto bg-white dark:bg-slate-900/50">
      <!-- Loading Skeleton -->
      <div v-if="loading && documents.length === 0" class="py-20 text-center text-slate-400 dark:text-slate-500 flex flex-col items-center gap-3">
        <Loader2 class="w-6 h-6 animate-spin text-blue-500" />
        <p>正在读取集合文档数据...</p>
      </div>

      <!-- Empty State -->
      <div v-else-if="!loading && documents.length === 0" class="py-20 text-center text-slate-400 dark:text-slate-500">
        <FileQuestion class="w-10 h-10 mx-auto mb-2 text-slate-300 dark:text-slate-600 opacity-60" />
        <p class="text-sm font-medium">暂无匹配文档数据</p>
        <p v-if="activeFilter" class="text-[11px] text-slate-500 dark:text-slate-600 mt-1">尝试修改或清空过滤条件</p>
      </div>

      <!-- Table -->
      <table v-else class="w-full border-collapse text-left text-xs font-mono">
        <thead class="bg-slate-50 sticky top-0 z-10 border-b border-slate-200 text-slate-600 dark:bg-slate-950/90 dark:border-slate-800 dark:text-slate-400 shadow-sm">
          <tr>
            <th class="px-3 py-2.5 w-12 text-center font-medium border-r border-slate-200/80 dark:border-slate-800/60">#</th>
            <th
              v-for="col in displayColumns"
              :key="col"
              class="px-3 py-2.5 font-medium whitespace-nowrap border-r border-slate-200/80 dark:border-slate-800/60 text-slate-700 dark:text-slate-300"
            >
              <div class="flex items-center gap-1.5">
                <span :class="isPrimaryKey(col) ? 'text-amber-700 dark:text-amber-300 font-bold' : ''">{{ col }}</span>
                <span
                  v-if="isPrimaryKey(col)"
                  class="px-1 py-0.2 rounded text-[9px] bg-amber-50 border border-amber-300 text-amber-700 dark:bg-amber-950/80 dark:border-amber-700/60 dark:text-amber-300 font-sans"
                >
                  PK
                </span>
                <span
                  v-else-if="isVectorField(col)"
                  class="px-1 py-0.2 rounded text-[9px] bg-indigo-50 border border-indigo-200 text-indigo-700 dark:bg-indigo-950/80 dark:border-indigo-700/60 dark:text-indigo-300 font-sans"
                >
                  VECTOR
                </span>
              </div>
            </th>
            <th class="px-3 py-2.5 w-20 text-center font-medium text-slate-500 dark:text-slate-400">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-200/70 dark:divide-slate-800/60">
          <tr
            v-for="(doc, rowIdx) in documents"
            :key="rowIdx"
            class="hover:bg-blue-50/40 dark:hover:bg-slate-800/50 transition group"
          >
            <!-- Row index -->
            <td class="px-3 py-2 text-center text-slate-400 dark:text-slate-500 border-r border-slate-200/60 dark:border-slate-800/40 text-[11px]">
              {{ offset + rowIdx + 1 }}
            </td>

            <!-- Columns -->
            <td
              v-for="col in displayColumns"
              :key="col"
              class="px-3 py-2 border-r border-slate-200/60 dark:border-slate-800/40 max-w-[280px] truncate"
            >
              <!-- Vector column cell with badge -->
              <div v-if="isVectorValue(doc[col])" class="flex items-center gap-1.5">
                <button
                  @click="openDetail(doc)"
                  class="inline-flex items-center gap-1 px-2 py-0.5 rounded bg-indigo-50 border border-indigo-200 text-indigo-700 hover:bg-indigo-100 dark:bg-indigo-950/70 dark:border-indigo-700/60 dark:text-indigo-300 dark:hover:bg-indigo-900/80 transition text-[11px]"
                  title="点击查看高维向量详情"
                >
                  <Sparkles class="w-3 h-3 text-indigo-500 dark:text-indigo-400" />
                  <span>{{ formatVectorSummary(doc[col]) }}</span>
                </button>
              </div>

              <!-- Plain scalar/object values -->
              <span v-else :class="isPrimaryKey(col) ? 'text-amber-800 dark:text-amber-200 font-semibold' : 'text-slate-800 dark:text-slate-300'" :title="String(doc[col] ?? '')">
                {{ formatScalarValue(doc[col]) }}
              </span>
            </td>

            <!-- Action column -->
            <td class="px-3 py-2 text-center">
              <button
                @click="openDetail(doc)"
                class="px-2 py-1 rounded bg-slate-100 hover:bg-blue-600 hover:text-white text-slate-600 dark:bg-slate-800 dark:hover:bg-blue-600 dark:hover:text-white dark:text-slate-400 transition text-[11px] inline-flex items-center gap-1 shadow-sm"
                title="查看完整文档 JSON"
              >
                <Eye class="w-3 h-3" />
                <span>JSON</span>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Detail Modal -->
    <JsonDetailModal
      :is-open="detailModalOpen"
      :document="selectedDoc"
      @close="detailModalOpen = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from "vue";
import {
  Filter,
  Search,
  RefreshCw,
  X,
  ChevronLeft,
  ChevronRight,
  AlertCircle,
  FileQuestion,
  Eye,
  Loader2,
  Sparkles,
} from "lucide-vue-next";
import { QueryDocuments } from "../../bindings/vectordb-1/backend/service/vectordbservice";
import type { CollectionMeta } from "../types";
import JsonDetailModal from "./JsonDetailModal.vue";

const props = defineProps<{
  connId: string;
  database: string;
  collection: string;
  collectionMeta: CollectionMeta | null;
}>();

const filterInput = ref("");
const activeFilter = ref("");
const limit = ref(20);
const offset = ref(0);
const totalCount = ref(0);
const documents = ref<Record<string, unknown>[]>([]);
const loading = ref(false);
const error = ref("");

const detailModalOpen = ref(false);
const selectedDoc = ref<Record<string, unknown> | null>(null);

// Dynamic column list derived from schema fields or returned documents
const displayColumns = computed(() => {
  const set = new Set<string>();

  if (props.collectionMeta?.fields) {
    for (const f of props.collectionMeta.fields) {
      set.add(f.fieldName);
    }
  }

  for (const doc of documents.value) {
    if (doc && typeof doc === "object") {
      for (const k of Object.keys(doc)) {
        set.add(k);
      }
    }
  }

  return Array.from(set);
});

function isPrimaryKey(colName: string): boolean {
  if (props.collectionMeta?.fields) {
    const f = props.collectionMeta.fields.find((field) => field.fieldName === colName);
    if (f?.primaryKey) return true;
  }
  return colName === "id" || colName === "_id";
}

function isVectorField(colName: string): boolean {
  if (props.collectionMeta?.fields) {
    const f = props.collectionMeta.fields.find((field) => field.fieldName === colName);
    if (f?.fieldType?.toLowerCase().includes("vector")) return true;
  }
  return colName.toLowerCase().includes("vector");
}

function isVectorValue(val: unknown): boolean {
  return Array.isArray(val) && val.length > 0 && typeof val[0] === "number";
}

function formatVectorSummary(val: unknown): string {
  if (!Array.isArray(val)) return "[]";
  const dim = val.length;
  if (dim <= 2) {
    return `[${val.map((v) => Number(v).toFixed(3)).join(", ")}]`;
  }
  return `[${Number(val[0]).toFixed(3)}, ${Number(val[1]).toFixed(3)}, ... (${dim} 维)]`;
}

function formatScalarValue(val: unknown): string {
  if (val === null || val === undefined) return "-";
  if (typeof val === "object") return JSON.stringify(val);
  return String(val);
}

async function fetchData() {
  if (!props.connId || !props.database || !props.collection) return;
  loading.value = true;
  error.value = "";
  try {
    const res = await QueryDocuments({
      connectionId: props.connId,
      database: props.database,
      collection: props.collection,
      limit: limit.value,
      offset: offset.value,
      filter: activeFilter.value,
    });

    if (res) {
      totalCount.value = Number(res.count) || 0;
      const docs: Record<string, unknown>[] = [];
      if (res.documents) {
        for (const item of res.documents) {
          if (item && typeof item === "object") {
            docs.push(item as Record<string, unknown>);
          }
        }
      }
      documents.value = docs;
    } else {
      documents.value = [];
      totalCount.value = 0;
    }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : String(e);
    documents.value = [];
  } finally {
    loading.value = false;
  }
}

function onQuery() {
  activeFilter.value = filterInput.value.trim();
  offset.value = 0;
  fetchData();
}

function onClearFilter() {
  filterInput.value = "";
  activeFilter.value = "";
  offset.value = 0;
  fetchData();
}

function onLimitChange() {
  offset.value = 0;
  fetchData();
}

function prevPage() {
  if (offset.value >= limit.value) {
    offset.value -= limit.value;
    fetchData();
  }
}

function nextPage() {
  offset.value += limit.value;
  fetchData();
}

function openDetail(doc: Record<string, unknown>) {
  selectedDoc.value = doc;
  detailModalOpen.value = true;
}

watch(
  () => [props.connId, props.database, props.collection],
  () => {
    offset.value = 0;
    filterInput.value = "";
    activeFilter.value = "";
    fetchData();
  }
);

onMounted(() => {
  fetchData();
});
</script>
