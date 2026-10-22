<template>
  <Teleport to="body">
    <div
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4 text-xs select-none"
      @click.self="handleClose"
    >
      <div class="w-full max-w-2xl rounded-xl bg-white border border-slate-200 shadow-2xl overflow-hidden flex flex-col text-slate-800 dark:bg-slate-900 dark:border-slate-700 dark:text-slate-100 transition-colors duration-200 max-h-[90vh]">
        <!-- Header -->
        <div class="px-5 py-3.5 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between shrink-0">
          <div class="flex items-center gap-2 font-semibold text-sm text-slate-900 dark:text-white">
            <component :is="isAI ? BookOpen : Table" :class="['w-4 h-4', isAI ? 'text-purple-500' : 'text-indigo-500']" />
            <span>新建{{ isAI ? "知识库集合" : "集合" }}</span>
            <span class="font-mono text-[11px] text-slate-400 dark:text-slate-500">/ {{ database }}</span>
            <span
              :class="[
                'px-1.5 py-0.5 rounded text-[9px] font-mono font-bold',
                isAI
                  ? 'bg-purple-100 text-purple-700 dark:bg-purple-950 dark:text-purple-300'
                  : 'bg-blue-100 text-blue-700 dark:bg-blue-950 dark:text-blue-300',
              ]"
            >
              {{ isAI ? "AI" : "BASE" }}
            </span>
          </div>
          <button
            type="button"
            @click="handleClose"
            class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-1 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition"
          >
            <X class="w-4 h-4" />
          </button>
        </div>

        <!-- Form -->
        <div class="p-5 space-y-4 overflow-y-auto">
          <!-- Collection Name -->
          <div class="space-y-1.5">
            <label class="block text-[11px] font-medium text-slate-600 dark:text-slate-300">
              集合名称 <span class="text-rose-500">*</span>
            </label>
            <input
              ref="nameInputRef"
              v-model="name"
              type="text"
              placeholder="如 coll_articles"
              class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500/40 focus:border-blue-500 transition"
              :disabled="loading"
              @keydown.enter="handleCreate"
            />
            <p class="text-[10px] text-slate-400 dark:text-slate-500">
              1-128 个字符，字母或数字开头，仅含字母、数字、下划线或连字符
            </p>
          </div>

          <!-- Description -->
          <div class="space-y-1.5">
            <label class="block text-[11px] font-medium text-slate-600 dark:text-slate-300">描述</label>
            <input
              v-model="description"
              type="text"
              placeholder="集合描述（可选）"
              class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500/40 focus:border-blue-500 transition"
              :disabled="loading"
            />
          </div>

          <!-- Base mode: shards / replicas -->
          <div v-if="!isAI" class="grid grid-cols-2 gap-3">
            <div class="space-y-1.5">
              <label class="block text-[11px] font-medium text-slate-600 dark:text-slate-300">分片数 (Shard)</label>
              <input
                v-model.number="shardNum"
                type="number"
                min="1"
                max="32"
                class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500/40 focus:border-blue-500 transition"
                :disabled="loading"
              />
            </div>
            <div class="space-y-1.5">
              <label class="block text-[11px] font-medium text-slate-600 dark:text-slate-300">副本数 (Replica)</label>
              <input
                v-model.number="replicaNum"
                type="number"
                min="0"
                max="8"
                class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500/40 focus:border-blue-500 transition"
                :disabled="loading"
              />
            </div>
          </div>

          <!-- AI mode: embedding language -->
          <div v-if="isAI" class="space-y-1.5">
            <label class="block text-[11px] font-medium text-slate-600 dark:text-slate-300">Embedding 语言</label>
            <select
              v-model="language"
              class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500/40 focus:border-blue-500 transition"
              :disabled="loading"
            >
              <option value="zh">中文 (zh)</option>
              <option value="en">英文 (en)</option>
            </select>
          </div>

          <!-- Base mode: indexes editor -->
          <div v-if="!isAI" class="space-y-2">
            <div class="flex items-center justify-between">
              <label class="block text-[11px] font-medium text-slate-600 dark:text-slate-300">
                索引定义 <span class="text-rose-500">*</span>
              </label>
              <button
                type="button"
                class="px-2 py-1 rounded-lg border border-slate-300 dark:border-slate-700 text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition flex items-center gap-1"
                :disabled="loading"
                @click="addRow('primaryKey')"
              >
                <Plus class="w-3 h-3" />
                <span>主键</span>
              </button>
              <div class="flex gap-1.5">
                <button
                  type="button"
                  class="px-2 py-1 rounded-lg border border-slate-300 dark:border-slate-700 text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition flex items-center gap-1"
                  :disabled="loading"
                  @click="addRow('filter')"
                >
                  <Plus class="w-3 h-3" />
                  <span>标量字段</span>
                </button>
                <button
                  type="button"
                  class="px-2 py-1 rounded-lg border border-slate-300 dark:border-slate-700 text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition flex items-center gap-1"
                  :disabled="loading"
                  @click="addRow('vector')"
                >
                  <Plus class="w-3 h-3" />
                  <span>向量字段</span>
                </button>
              </div>
            </div>

            <div v-if="rows.length === 0" class="p-4 rounded-lg border border-dashed border-slate-300 dark:border-slate-700 text-center text-slate-400 dark:text-slate-500">
              点击上方按钮添加主键、标量或向量字段
            </div>

            <div v-else class="space-y-2">
              <div
                v-for="(row, i) in rows"
                :key="i"
                class="p-3 rounded-lg border border-slate-200 dark:border-slate-700/80 bg-slate-50/60 dark:bg-slate-800/40 space-y-2"
              >
                <div class="flex items-center gap-2">
                  <span
                    :class="[
                      'px-1.5 py-0.5 rounded text-[9px] font-mono font-bold shrink-0',
                      row.indexType === 'primaryKey'
                        ? 'bg-amber-100 text-amber-700 dark:bg-amber-950 dark:text-amber-300'
                        : row.indexType === 'vector'
                          ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300'
                          : 'bg-slate-200 text-slate-600 dark:bg-slate-700 dark:text-slate-300',
                    ]"
                  >
                    {{ row.indexType === "primaryKey" ? "PK" : row.indexType === "vector" ? "VECTOR" : "FILTER" }}
                  </span>
                  <input
                    v-model="row.fieldName"
                    type="text"
                    placeholder="字段名"
                    class="flex-1 px-2.5 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:border-blue-500 transition font-mono text-[11px]"
                    :disabled="loading"
                  />
                  <select
                    v-model="row.fieldType"
                    class="px-2 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-100 focus:outline-none focus:border-blue-500 transition"
                    :disabled="loading || row.indexType === 'primaryKey' || row.indexType === 'vector'"
                  >
                    <option v-for="t in fieldTypesFor(row)" :key="t" :value="t">{{ t }}</option>
                  </select>
                  <button
                    type="button"
                    class="p-1.5 rounded-lg text-slate-400 hover:text-rose-500 hover:bg-rose-50 dark:hover:bg-rose-950/40 transition shrink-0"
                    :disabled="loading"
                    @click="removeRow(i)"
                  >
                    <Trash2 class="w-3.5 h-3.5" />
                  </button>
                </div>
                <div v-if="row.indexType === 'vector'" class="flex items-center gap-2 pl-1">
                  <select
                    v-model="row.metricType"
                    class="px-2 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-100 focus:outline-none focus:border-blue-500 transition"
                    :disabled="loading"
                  >
                    <option value="COSINE">COSINE</option>
                    <option value="L2">L2</option>
                    <option value="IP">IP</option>
                  </select>
                  <div class="flex items-center gap-1.5 flex-1">
                    <span class="text-[10px] text-slate-500 dark:text-slate-400 shrink-0">维度</span>
                    <input
                      v-model.number="row.dimension"
                      type="number"
                      min="2"
                      max="65536"
                      placeholder="如 768"
                      class="w-28 px-2.5 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:border-blue-500 transition"
                      :disabled="loading"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Error -->
          <div
            v-if="error"
            class="p-2.5 rounded-lg bg-rose-50 border border-rose-200 text-rose-700 dark:bg-rose-950/30 dark:border-rose-800/50 dark:text-rose-300 flex items-start gap-1.5"
          >
            <AlertCircle class="w-3.5 h-3.5 shrink-0 mt-0.5" />
            <span class="break-all">{{ error }}</span>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-5 py-3 border-t border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/30 flex items-center justify-end gap-2 shrink-0">
          <button
            type="button"
            @click="handleClose"
            class="px-4 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition"
            :disabled="loading"
          >
            取消
          </button>
          <button
            type="button"
            @click="handleCreate"
            class="px-4 py-1.5 rounded-lg text-white font-medium transition shadow-sm flex items-center gap-1.5 disabled:opacity-60"
            :class="isAI ? 'bg-purple-600 hover:bg-purple-700' : 'bg-blue-600 hover:bg-blue-700'"
            :disabled="loading || !name.trim()"
          >
            <Loader2 v-if="loading" class="w-3.5 h-3.5 animate-spin" />
            <span>{{ loading ? "创建中..." : "创建" }}</span>
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onUnmounted } from "vue";
import { Table, BookOpen, X, AlertCircle, Loader2, Plus, Trash2 } from "lucide-vue-next";
import { useVectorDBStore } from "../stores/vectordb";
import { useConnectionStore } from "../stores/connection";

const props = defineProps<{
  database: string;
  dbType: string;
}>();

const emit = defineEmits<{
  (e: "closed"): void;
  (e: "created", database: string, collection: string): void;
}>();

const vdbStore = useVectorDBStore();
const connStore = useConnectionStore();

const isAI = computed(() => props.dbType === "ai");

const name = ref("");
const description = ref("");
const shardNum = ref(1);
const replicaNum = ref(2);
const language = ref("zh");
const loading = ref(false);
const error = ref("");
const nameInputRef = ref<HTMLInputElement | null>(null);

interface IndexRow {
  fieldName: string;
  fieldType: string;
  indexType: string;
  metricType: string;
  dimension: number | undefined;
}

const rows = ref<IndexRow[]>([
  { fieldName: "id", fieldType: "string", indexType: "primaryKey", metricType: "", dimension: undefined },
  { fieldName: "vector", fieldType: "vector", indexType: "vector", metricType: "COSINE", dimension: 768 },
]);

const SCALAR_TYPES = ["string", "uint64", "int64", "double", "bool"];

function fieldTypesFor(row: IndexRow): string[] {
  if (row.indexType === "vector") return ["vector"];
  if (row.indexType === "primaryKey") return ["string"];
  return SCALAR_TYPES;
}

function addRow(kind: "primaryKey" | "filter" | "vector") {
  if (kind === "primaryKey") {
    if (rows.value.some((r) => r.indexType === "primaryKey")) {
      error.value = "只能有一个主键字段";
      return;
    }
    rows.value.push({ fieldName: "", fieldType: "string", indexType: "primaryKey", metricType: "", dimension: undefined });
  } else if (kind === "vector") {
    rows.value.push({ fieldName: "", fieldType: "vector", indexType: "vector", metricType: "COSINE", dimension: 768 });
  } else {
    rows.value.push({ fieldName: "", fieldType: "string", indexType: "filter", metricType: "", dimension: undefined });
  }
}

function removeRow(i: number) {
  rows.value.splice(i, 1);
}

function handleClose() {
  if (loading.value) return;
  emit("closed");
}

const NAME_RE = /^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$/;

function validate(): string {
  if (!NAME_RE.test(name.value.trim())) {
    return "名称须为 1-128 个字符，字母或数字开头，仅含字母、数字、下划线或连字符";
  }
  if (isAI.value) return "";

  const pk = rows.value.filter((r) => r.indexType === "primaryKey");
  if (pk.length === 0) return "必须定义一个主键字段";
  const vectors = rows.value.filter((r) => r.indexType === "vector");
  if (vectors.length === 0) return "必须定义至少一个向量字段";

  const seen = new Set<string>();
  for (const r of rows.value) {
    if (!r.fieldName.trim()) return "存在未填写的字段名";
    if (seen.has(r.fieldName.trim())) return `字段名重复: ${r.fieldName}`;
    seen.add(r.fieldName.trim());
    if (r.indexType === "vector") {
      if (!r.dimension || r.dimension < 2 || r.dimension > 65536) {
        return `向量字段 "${r.fieldName}" 的维度须在 2-65536 之间`;
      }
    }
  }
  return "";
}

async function handleCreate() {
  if (loading.value || !name.value.trim()) return;

  const msg = validate();
  if (msg) {
    error.value = msg;
    return;
  }

  loading.value = true;
  error.value = "";
  try {
    const options: Record<string, unknown> = { description: description.value };
    if (isAI.value) {
      options.language = language.value;
    } else {
      options.shardNum = shardNum.value;
      options.replicaNum = replicaNum.value;
      options.indexes = rows.value.map((r) => ({
        fieldName: r.fieldName.trim(),
        fieldType: r.fieldType,
        indexType: r.indexType === "vector" ? "HNSW" : r.indexType,
        metricType: r.indexType === "vector" ? r.metricType : "",
        dimension: r.indexType === "vector" ? r.dimension : 0,
      }));
    }
    await vdbStore.createCollection(
      connStore.activeConnectionId?.value || "",
      props.database,
      name.value.trim(),
      props.dbType,
      options as {
        description?: string;
        replicaNum?: number;
        shardNum?: number;
        indexes?: Array<{ fieldName: string; fieldType: string; indexType: string; dimension?: number; metricType?: string }>;
        language?: string;
      }
    );
    emit("created", props.database, name.value.trim());
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

function handleKeyDown(e: KeyboardEvent) {
  if (e.key === "Escape" && !loading.value) {
    handleClose();
  }
}

onMounted(async () => {
  window.addEventListener("keydown", handleKeyDown);
  await nextTick();
  nameInputRef.value?.focus();
});

onUnmounted(() => {
  window.removeEventListener("keydown", handleKeyDown);
});
</script>
