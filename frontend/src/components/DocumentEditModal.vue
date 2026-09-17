<template>
  <div
    v-if="isOpen"
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-fade-in select-none text-xs"
    @keydown.esc="handleEsc"
  >
    <div
      class="w-full max-w-2xl max-h-[90vh] rounded-xl bg-white border border-slate-200 shadow-2xl overflow-hidden flex flex-col text-slate-800 dark:bg-slate-900 dark:border-slate-700 dark:text-slate-100 transition-colors duration-200"
    >
      <!-- Header -->
      <div class="px-6 py-3.5 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between shrink-0 bg-slate-50/60 dark:bg-slate-950/40">
        <div class="flex items-center gap-2.5 min-w-0">
          <div class="w-8 h-8 rounded-lg bg-blue-50 border border-blue-200 text-blue-600 dark:bg-blue-600/20 dark:border-blue-500/30 dark:text-blue-400 flex items-center justify-center shrink-0">
            <Pencil class="w-4 h-4" />
          </div>
          <div class="truncate">
            <div class="flex items-center gap-2">
              <h3 class="text-sm font-semibold text-slate-900 dark:text-white tracking-tight">编辑文档</h3>
              <span
                :class="[
                  'px-1.5 py-0.2 rounded text-[9px] font-mono font-bold uppercase',
                  effectiveDbType === 'ai'
                    ? 'bg-purple-100 text-purple-700 dark:bg-purple-950 dark:text-purple-300'
                    : 'bg-blue-100 text-blue-700 dark:bg-blue-950 dark:text-blue-300'
                ]"
              >
                {{ effectiveDbType === 'ai' ? 'AI 知识库' : 'BASE 集合' }}
              </span>
            </div>
            <p class="text-[11px] text-slate-500 dark:text-slate-400 font-mono mt-0.5 truncate">
              {{ database }}.{{ collection }} <span v-if="primaryKeyVal">| 主键: {{ primaryKeyVal }}</span>
            </p>
          </div>
        </div>

        <!-- Mode Switch & Close -->
        <div class="flex items-center gap-2 shrink-0">
          <div class="flex items-center bg-slate-200/80 dark:bg-slate-800 p-0.5 rounded-lg text-[11px]">
            <button
              type="button"
              @click="activeMode = 'form'"
              :class="[
                'px-2.5 py-1 rounded-md transition font-medium flex items-center gap-1',
                activeMode === 'form'
                  ? 'bg-white text-blue-600 dark:bg-slate-700 dark:text-white shadow-sm'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'
              ]"
            >
              <List class="w-3 h-3" />
              <span>表单编辑</span>
            </button>
            <button
              type="button"
              @click="switchToCodeMode"
              :class="[
                'px-2.5 py-1 rounded-md transition font-medium flex items-center gap-1',
                activeMode === 'code'
                  ? 'bg-white text-blue-600 dark:bg-slate-700 dark:text-white shadow-sm'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'
              ]"
            >
              <Code class="w-3 h-3" />
              <span>JSON 源码</span>
            </button>
          </div>

          <button
            type="button"
            @click="close"
            class="text-slate-400 hover:text-slate-700 dark:hover:text-white p-1 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition"
          >
            <X class="w-4 h-4" />
          </button>
        </div>
      </div>

      <!-- Error banner if any -->
      <div
        v-if="errorMessage"
        class="px-6 py-2.5 bg-rose-50 border-b border-rose-200 text-rose-700 dark:bg-rose-950/40 dark:border-rose-900 dark:text-rose-300 flex items-start gap-2.5 shrink-0 transition-all"
      >
        <AlertCircle class="w-4 h-4 shrink-0 text-rose-500 mt-0.5" />
        <div class="flex-1 min-w-0">
          <div class="flex items-center justify-between gap-2">
            <span class="font-semibold text-rose-800 dark:text-rose-200 text-[11px]">操作失败</span>
            <div class="flex items-center gap-2 shrink-0">
              <button
                type="button"
                @click="copyError"
                class="px-1.5 py-0.5 rounded text-[10px] bg-rose-100 hover:bg-rose-200 text-rose-800 dark:bg-rose-900/50 dark:hover:bg-rose-800 dark:text-rose-200 transition font-medium flex items-center gap-1"
                title="复制完整错误信息"
              >
                <Check v-if="copiedError" class="w-3 h-3 text-emerald-600 dark:text-emerald-400" />
                <Copy v-else class="w-3 h-3" />
                <span>{{ copiedError ? '已复制' : '复制错误' }}</span>
              </button>
              <button
                v-if="errorMessage.length > 70"
                type="button"
                @click="showFullError = !showFullError"
                class="text-[10px] text-rose-600 hover:text-rose-800 dark:text-rose-400 dark:hover:text-rose-200 underline font-medium"
              >
                {{ showFullError ? '收起' : '展开完整错误' }}
              </button>
            </div>
          </div>
          <p
            :class="[
              'mt-1 font-mono text-[11px] leading-relaxed break-all select-text',
              showFullError ? 'whitespace-pre-wrap max-h-48 overflow-y-auto bg-rose-100/50 dark:bg-rose-900/20 p-2 rounded' : 'line-clamp-2'
            ]"
          >
            {{ errorMessage }}
          </p>
        </div>
      </div>

      <!-- Body -->
      <div class="flex-1 overflow-y-auto p-6 space-y-4">
        <!-- Form Mode -->
        <div v-if="activeMode === 'form'" class="space-y-4">
          <!-- Read-only Info Banner (PK & Vectors) -->
          <div v-if="readOnlyFields.length > 0" class="p-3 rounded-lg bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 space-y-2">
            <div class="text-[11px] font-medium text-slate-500 dark:text-slate-400 flex items-center gap-1">
              <Lock class="w-3 h-3 text-slate-400" />
              <span>系统只读 / 向量与主键属性（不可修改）</span>
            </div>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-2 text-[11px]">
              <div
                v-for="rf in readOnlyFields"
                :key="rf.key"
                class="flex items-center justify-between px-2 py-1 rounded bg-white dark:bg-slate-900 border border-slate-200/80 dark:border-slate-800"
              >
                <div class="flex items-center gap-1.5 truncate">
                  <span class="font-mono text-slate-700 dark:text-slate-300">{{ rf.key }}</span>
                  <span
                    :class="[
                      'px-1 py-0.2 rounded text-[9px] font-mono',
                      rf.badge === 'PK' ? 'bg-amber-100 text-amber-700 dark:bg-amber-950 dark:text-amber-300' :
                      rf.badge === 'VECTOR' ? 'bg-indigo-100 text-indigo-700 dark:bg-indigo-950 dark:text-indigo-300' :
                      'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400'
                    ]"
                  >
                    {{ rf.badge }}
                  </span>
                </div>
                <span class="text-slate-400 dark:text-slate-500 font-mono truncate max-w-[120px]" :title="String(rf.value)">
                  {{ rf.summary }}
                </span>
              </div>
            </div>
          </div>

          <!-- Editable Fields Section -->
          <div>
            <div class="flex items-center justify-between mb-2">
              <span class="text-xs font-semibold text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
                <Sliders class="w-3.5 h-3.5 text-blue-500" />
                <span>可编辑字段</span>
              </span>
              <button
                type="button"
                @click="openAddFieldModal"
                class="text-[11px] text-blue-600 hover:text-blue-500 dark:text-blue-400 flex items-center gap-1 font-medium transition hover:underline"
              >
                <Plus class="w-3 h-3" />
                <span>添加新字段</span>
              </button>
            </div>

            <div v-if="editableFields.length === 0" class="py-7 px-4 text-center text-slate-400 dark:text-slate-500 border border-dashed border-slate-200 dark:border-slate-800 rounded-lg space-y-2">
              <p v-if="effectiveDbType === 'ai'" class="text-xs text-slate-600 dark:text-slate-300">
                知识库系统内置字段（文件名、状态、大小、切片等）由系统自动维护，不可修改。
              </p>
              <p v-else class="text-xs">当前文档暂无其他可编辑标量字段。</p>
              <p class="text-[11px] text-slate-500">点击右上角“添加新字段”或点击下方快捷添加常用属性：</p>
              <div class="flex items-center justify-center gap-2 pt-1 flex-wrap">
                <button
                  type="button"
                  @click="quickAddField('author', '')"
                  class="px-2.5 py-1 rounded-md bg-slate-100 hover:bg-blue-50 hover:text-blue-600 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 text-[11px] transition font-mono border border-slate-200 dark:border-slate-700"
                >
                  + author (作者)
                </button>
                <button
                  type="button"
                  @click="quickAddField('is_enabled', true)"
                  class="px-2.5 py-1 rounded-md bg-slate-100 hover:bg-blue-50 hover:text-blue-600 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 text-[11px] transition font-mono border border-slate-200 dark:border-slate-700 cursor-pointer"
                >
                  + is_enabled ({{ effectiveDbType === 'ai' ? '状态 1/0' : '启用状态' }})
                </button>
                <button
                  type="button"
                  @click="quickAddField('is_open', true)"
                  class="px-2.5 py-1 rounded-md bg-slate-100 hover:bg-blue-50 hover:text-blue-600 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 text-[11px] transition font-mono border border-slate-200 dark:border-slate-700 cursor-pointer"
                >
                  + is_open ({{ effectiveDbType === 'ai' ? '状态 1/0' : '开放状态' }})
                </button>
                <button
                  type="button"
                  @click="quickAddField('category', '')"
                  class="px-2.5 py-1 rounded-md bg-slate-100 hover:bg-blue-50 hover:text-blue-600 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 text-[11px] transition font-mono border border-slate-200 dark:border-slate-700 cursor-pointer"
                >
                  + category (分类)
                </button>
              </div>
            </div>

            <div v-else class="space-y-3">
              <div
                v-for="item in editableFields"
                :key="item.key"
                class="p-3 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 shadow-sm space-y-1.5"
              >
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <label class="font-mono font-medium text-slate-800 dark:text-slate-200">{{ item.key }}</label>
                    <span class="px-1.5 py-0.2 rounded text-[10px] font-mono bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400">
                      {{ item.type }}
                    </span>
                  </div>

                  <button
                    type="button"
                    @click="removeField(item.key)"
                    class="text-slate-400 hover:text-rose-500 dark:hover:text-rose-400 p-0.5 rounded transition"
                    title="移除此字段更新"
                  >
                    <Trash2 class="w-3.5 h-3.5" />
                  </button>
                </div>

                <!-- Control: Boolean -->
                <div v-if="item.type === 'boolean'" class="flex items-center gap-3 pt-1">
                  <label class="relative inline-flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      v-model="editValues[item.key]"
                      class="sr-only peer"
                    />
                    <div class="w-9 h-5 bg-slate-200 peer-focus:outline-none rounded-full peer dark:bg-slate-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-slate-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all dark:border-slate-600 peer-checked:bg-blue-600"></div>
                  </label>
                  <span class="font-mono font-semibold" :class="editValues[item.key] ? 'text-emerald-600 dark:text-emerald-400' : 'text-slate-400'">
                    {{ editValues[item.key] ? (effectiveDbType === 'ai' ? '1 (启用 / true)' : 'true (启用)') : (effectiveDbType === 'ai' ? '0 (禁用 / false)' : 'false (禁用)') }}
                  </span>
                  <span v-if="effectiveDbType === 'ai'" class="text-[10px] font-mono text-purple-600 dark:text-purple-400 bg-purple-50 dark:bg-purple-950/60 px-1.5 py-0.5 rounded border border-purple-200 dark:border-purple-800/40">
                    保存时自动存为 UInt64 ({{ editValues[item.key] ? 1 : 0 }})
                  </span>
                </div>

                <!-- Control: Number -->
                <div v-else-if="item.type === 'number'">
                  <input
                    type="number"
                    step="any"
                    v-model.number="editValues[item.key]"
                    class="w-full px-3 py-1.5 rounded-lg bg-slate-50 border border-slate-300 text-slate-900 dark:bg-slate-950 dark:border-slate-700 dark:text-white font-mono text-xs focus:outline-none focus:ring-1 focus:ring-blue-500"
                  />
                </div>

                <!-- Control: Object / JSON -->
                <div v-else-if="item.type === 'object'">
                  <textarea
                    rows="3"
                    v-model="editObjectStrings[item.key]"
                    @input="handleObjectInput(item.key)"
                    placeholder='例如: {"key": "value"}'
                    class="w-full px-3 py-1.5 rounded-lg bg-slate-50 border border-slate-300 text-slate-900 dark:bg-slate-950 dark:border-slate-700 dark:text-white font-mono text-xs focus:outline-none focus:ring-1 focus:ring-blue-500 leading-relaxed"
                  ></textarea>
                  <p v-if="objectErrors[item.key]" class="text-[10px] text-rose-500 mt-0.5">
                    JSON 格式错误: {{ objectErrors[item.key] }}
                  </p>
                </div>

                <!-- Control: String / Default -->
                <div v-else>
                  <textarea
                    v-if="String(editValues[item.key] ?? '').length > 60"
                    rows="2"
                    v-model="editValues[item.key]"
                    class="w-full px-3 py-1.5 rounded-lg bg-slate-50 border border-slate-300 text-slate-900 dark:bg-slate-950 dark:border-slate-700 dark:text-white font-mono text-xs focus:outline-none focus:ring-1 focus:ring-blue-500 leading-relaxed"
                  ></textarea>
                  <input
                    v-else
                    type="text"
                    v-model="editValues[item.key]"
                    class="w-full px-3 py-1.5 rounded-lg bg-slate-50 border border-slate-300 text-slate-900 dark:bg-slate-950 dark:border-slate-700 dark:text-white font-mono text-xs focus:outline-none focus:ring-1 focus:ring-blue-500"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- JSON Code Mode -->
        <div v-else class="space-y-2">
          <div class="flex items-center justify-between text-[11px] text-slate-500 dark:text-slate-400">
            <span>直接编辑需要提交的更新字段 JSON 字典：</span>
            <span v-if="jsonSyntaxError" class="text-rose-500 font-mono">{{ jsonSyntaxError }}</span>
            <span v-else class="text-emerald-500 font-mono">JSON 语法有效</span>
          </div>
          <textarea
            v-model="codeModeJson"
            @input="validateCodeModeJson"
            rows="14"
            class="w-full p-3 rounded-lg bg-slate-50 border border-slate-300 text-slate-900 dark:bg-slate-950 dark:border-slate-700 dark:text-slate-100 font-mono text-xs leading-relaxed focus:outline-none focus:ring-2 focus:ring-blue-500 shadow-inner"
            :placeholder="'{\n  &quot;field&quot;: &quot;new_value&quot;\n}'"
          ></textarea>
        </div>
      </div>

      <!-- Footer -->
      <div class="px-6 py-3.5 border-t border-slate-200 bg-slate-50 dark:border-slate-800 dark:bg-slate-900 flex items-center justify-between shrink-0">
        <div class="text-[11px] text-slate-500 dark:text-slate-400">
          <span v-if="pendingCount > 0" class="text-blue-600 dark:text-blue-400 font-medium">
            准备更新 {{ pendingCount }} 个字段
          </span>
          <span v-else>未更改字段</span>
        </div>

        <div class="flex items-center gap-2">
          <button
            type="button"
            @click="close"
            :disabled="saving"
            class="px-4 py-1.5 rounded-lg text-xs font-medium bg-white border border-slate-200 hover:bg-slate-100 text-slate-700 dark:bg-slate-800 dark:border-transparent dark:hover:bg-slate-700 dark:text-white transition shadow-sm disabled:opacity-50"
          >
            取消
          </button>
          <button
            type="button"
            @click="handleSave"
            :disabled="saving || !!jsonSyntaxError || pendingCount === 0"
            class="px-4 py-1.5 rounded-lg text-xs font-medium bg-blue-600 hover:bg-blue-500 text-white transition shadow-sm flex items-center gap-1.5 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Loader2 v-if="saving" class="w-3.5 h-3.5 animate-spin" />
            <Check v-else class="w-3.5 h-3.5" />
            <span>保存修改</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Mini Modal for adding new custom field -->
    <div
      v-if="showAddFieldModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-xs p-4"
      @click.self="showAddFieldModal = false"
    >
      <div class="w-full max-w-sm rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 p-4 shadow-xl space-y-3">
        <h4 class="font-semibold text-slate-900 dark:text-white text-xs">添加可更新字段</h4>
        <div class="space-y-2">
          <div>
            <label class="text-[11px] text-slate-500">字段名称 (Key)</label>
            <input
              type="text"
              v-model="newFieldKey"
              placeholder="例如: author 或 category"
              class="w-full px-3 py-1.5 rounded-md bg-slate-50 border border-slate-300 dark:bg-slate-950 dark:border-slate-700 dark:text-white text-xs font-mono mt-1"
            />
            <p v-if="newFieldError" class="text-[10px] text-rose-500 mt-1 font-mono">
              {{ newFieldError }}
            </p>
          </div>
          <div>
            <label class="text-[11px] text-slate-500">字段类型</label>
            <select
              v-model="newFieldType"
              class="w-full px-3 py-1.5 rounded-md bg-slate-50 border border-slate-300 dark:bg-slate-950 dark:border-slate-700 dark:text-white text-xs mt-1"
            >
              <option value="string">文本 (String)</option>
              <option value="boolean">布尔 / 状态开关 (Boolean {{ effectiveDbType === 'ai' ? '→ 自动转为 UInt64 1/0' : '' }})</option>
              <option value="number">数字 (Number {{ effectiveDbType === 'ai' ? '→ UInt64' : '' }})</option>
            </select>
          </div>
        </div>
        <div class="flex justify-end gap-2 pt-2">
          <button
            type="button"
            @click="showAddFieldModal = false"
            class="px-3 py-1 rounded-md text-xs bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:text-slate-300"
          >
            取消
          </button>
          <button
            type="button"
            @click="confirmAddField"
            :disabled="!newFieldKey.trim() || !!newFieldError"
            class="px-3 py-1 rounded-md text-xs bg-blue-600 hover:bg-blue-500 text-white disabled:opacity-50 disabled:cursor-not-allowed"
          >
            添加
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import {
  Pencil,
  Check,
  X,
  Lock,
  Sliders,
  List,
  Code,
  AlertCircle,
  Loader2,
  Trash2,
  Plus,
  Copy,
} from "lucide-vue-next";
import type { CollectionMeta } from "../types";
import { useVectorDBStore } from "../stores/vectordb";

const props = defineProps<{
  isOpen: boolean;
  document: Record<string, unknown> | null;
  collectionMeta: CollectionMeta | null;
  database: string;
  collection: string;
  dbType: string;
  connId: string;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "saved", updatedDoc: Record<string, unknown>): void;
}>();

const vdbStore = useVectorDBStore();

const activeMode = ref<"form" | "code">("form");
const saving = ref(false);
const errorMessage = ref("");
const showFullError = ref(false);
const copiedError = ref(false);

async function copyError() {
  if (!errorMessage.value) return;
  try {
    await navigator.clipboard.writeText(errorMessage.value);
    copiedError.value = true;
    setTimeout(() => {
      copiedError.value = false;
    }, 2000);
  } catch {
    const ta = document.createElement("textarea");
    ta.value = errorMessage.value;
    document.body.appendChild(ta);
    ta.select();
    document.execCommand("copy");
    document.body.removeChild(ta);
    copiedError.value = true;
    setTimeout(() => {
      copiedError.value = false;
    }, 2000);
  }
}

// Form state
const editValues = ref<Record<string, any>>({});
const editObjectStrings = ref<Record<string, string>>({});
const objectErrors = ref<Record<string, string>>({});

// Code mode state
const codeModeJson = ref("");
const jsonSyntaxError = ref("");

// Add Field Modal
const showAddFieldModal = ref(false);
const newFieldKey = ref("");
const newFieldType = ref<"string" | "boolean" | "number">("string");

const effectiveDbType = computed(() => {
  if (props.collectionMeta?.isAiCollection) return "ai";
  if (props.dbType === "ai") return "ai";
  return "base";
});

// AI Suite read-only built-in attributes (case-insensitive)
const AI_READONLY_FIELDS = new Set([
  "documentsetid",
  "documentsetname",
  "document_set_id",
  "document_set_name",
  "status",
  "filetype",
  "file_type",
  "bytesize",
  "byte_size",
  "bytelength",
  "byte_length",
  "chunknum",
  "chunk_num",
  "createtime",
  "create_time",
  "updatetime",
  "update_time",
  "lastupdatetime",
  "last_update_time",
  "textprefix",
  "text_prefix",
  "text",
  "textlength",
  "text_length",
  "indexedprogress",
  "indexed_progress",
  "indexedstatus",
  "indexed_status",
  "indexederrormsg",
  "indexed_error_msg",
  "keywords",
  "appendtitletochunk",
  "append_title_to_chunk",
  "appendkeywordstochunk",
  "append_keywords_to_chunk",
  "chunksplitter",
  "chunk_splitter",
  "splitterprocess",
  "splitter_process",
  "splitterpreprocess",
  "splitter_preprocess",
  "parsingprocess",
  "parsing_process",
  "parsingtype",
  "parsing_type",
  "metadata",
]);

function isAiReadOnly(key: string): boolean {
  return AI_READONLY_FIELDS.has(key.toLowerCase());
}

// Determine Primary Key
const primaryKeyCol = computed(() => {
  if (effectiveDbType.value === "ai") {
    if (props.document && "documentSetId" in props.document) return "documentSetId";
    return "documentSetId";
  }
  // Check CollectionMeta
  if (props.collectionMeta?.fields) {
    const pk = props.collectionMeta.fields.find((f) => f.primaryKey);
    if (pk) return pk.fieldName;
  }
  if (props.document) {
    if ("id" in props.document) return "id";
    if ("_id" in props.document) return "_id";
    if ("doc_id" in props.document) return "doc_id";
  }
  return "id";
});

const primaryKeyVal = computed(() => {
  if (!props.document) return "";
  const pk = primaryKeyCol.value;
  return String(props.document[pk] ?? "");
});

function isVectorVal(val: unknown): boolean {
  return Array.isArray(val) && val.length > 5 && typeof val[0] === "number";
}

// Categorize fields
const readOnlyFields = computed(() => {
  if (!props.document) return [];
  const list: { key: string; value: unknown; summary: string; badge: string }[] = [];

  for (const [k, v] of Object.entries(props.document)) {
    if (k === primaryKeyCol.value) {
      list.push({
        key: k,
        value: v,
        summary: String(v),
        badge: "PK",
      });
    } else if (isVectorVal(v)) {
      list.push({
        key: k,
        value: v,
        summary: `[${(v as number[]).slice(0, 2).join(", ")}, ... (${(v as number[]).length} 维)]`,
        badge: "VECTOR",
      });
    } else if (effectiveDbType.value === "ai" && isAiReadOnly(k)) {
      list.push({
        key: k,
        value: v,
        summary: String(v),
        badge: "SYS",
      });
    }
  }
  return list;
});

interface EditableFieldItem {
  key: string;
  type: "string" | "number" | "boolean" | "object";
}

const editableFields = computed<EditableFieldItem[]>(() => {
  const fields: EditableFieldItem[] = [];
  const roKeys = new Set(readOnlyFields.value.map((f) => f.key));

  for (const key of Object.keys(editValues.value)) {
    if (roKeys.has(key)) continue;
    const val = editValues.value[key];
    let t: "string" | "number" | "boolean" | "object" = "string";
    if (typeof val === "boolean") t = "boolean";
    else if (typeof val === "number") t = "number";
    else if (val && typeof val === "object") t = "object";

    fields.push({
      key,
      type: t,
    });
  }
  return fields;
});

const pendingCount = computed(() => {
  if (activeMode.value === "code") {
    try {
      const parsed = JSON.parse(codeModeJson.value);
      return Object.keys(parsed).length;
    } catch {
      return 0;
    }
  }
  return editableFields.value.length;
});

function isBooleanLikeKey(key: string): boolean {
  const lower = key.toLowerCase();
  return (
    lower.startsWith("is_") ||
    lower.startsWith("has_") ||
    lower.startsWith("enable") ||
    lower.startsWith("open") ||
    lower.endsWith("_enabled") ||
    lower.endsWith("_open") ||
    lower.endsWith("_status")
  );
}

function initForm() {
  errorMessage.value = "";
  showFullError.value = false;
  copiedError.value = false;
  jsonSyntaxError.value = "";
  editValues.value = {};
  editObjectStrings.value = {};
  objectErrors.value = {};

  if (!props.document) return;

  const roKeys = new Set([primaryKeyCol.value]);

  for (const [k, v] of Object.entries(props.document)) {
    if (roKeys.has(k) || isVectorVal(v)) continue;
    if (effectiveDbType.value === "ai" && isAiReadOnly(k)) continue;

    // In VectorDB, documents strictly use flat scalar attributes.
    // If there are nested objects (like legacy metaData), flatten them into top-level editable fields.
    if (v !== null && typeof v === "object" && !Array.isArray(v)) {
      for (const [subK, subV] of Object.entries(v as Record<string, any>)) {
        if (effectiveDbType.value === "ai" && isAiReadOnly(subK)) continue;
        if (effectiveDbType.value === "ai" && (subV === 1 || subV === 0) && isBooleanLikeKey(subK)) {
          editValues.value[subK] = subV === 1;
        } else {
          editValues.value[subK] = subV;
        }
      }
    } else {
      if (effectiveDbType.value === "ai" && (v === 1 || v === 0) && isBooleanLikeKey(k)) {
        editValues.value[k] = v === 1;
      } else {
        editValues.value[k] = v;
      }
    }
  }
}

function quickAddField(key: string, defaultVal: unknown) {
  editValues.value[key] = defaultVal;
}

watch(
  () => props.isOpen,
  (open) => {
    if (open) {
      activeMode.value = "form";
      initForm();
    }
  },
  { immediate: true }
);

function handleObjectInput(key: string) {
  try {
    const parsed = JSON.parse(editObjectStrings.value[key]);
    editValues.value[key] = parsed;
    delete objectErrors.value[key];
  } catch (err: unknown) {
    objectErrors.value[key] = err instanceof Error ? err.message : "格式错误";
  }
}

function switchToCodeMode() {
  const payload: Record<string, any> = {};
  for (const [k, v] of Object.entries(editValues.value)) {
    payload[k] = v;
  }
  codeModeJson.value = JSON.stringify(payload, null, 2);
  jsonSyntaxError.value = "";
  activeMode.value = "code";
}

function validateCodeModeJson() {
  try {
    const parsed = JSON.parse(codeModeJson.value);
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
      jsonSyntaxError.value = "必须是合法的 JSON 对象键值对";
      return;
    }
    for (const k of Object.keys(parsed)) {
      if (k === primaryKeyCol.value) {
        jsonSyntaxError.value = `禁止包含主键字段 "${k}"（主键不可修改）`;
        return;
      }
      if (effectiveDbType.value === "ai" && isAiReadOnly(k)) {
        jsonSyntaxError.value = `禁止包含系统只读字段 "${k}"（系统属性不可修改）`;
        return;
      }
      const val = parsed[k];
      if (val !== null && typeof val === "object" && !Array.isArray(val)) {
        jsonSyntaxError.value = `字段 "${k}" 不能是嵌套对象，VectorDB 仅支持标量字段`;
        return;
      }
    }
    jsonSyntaxError.value = "";
  } catch (e: unknown) {
    jsonSyntaxError.value = e instanceof Error ? e.message : "JSON 语法无效";
  }
}

function removeField(key: string) {
  delete editValues.value[key];
  delete editObjectStrings.value[key];
  delete objectErrors.value[key];
}

const newFieldError = computed(() => {
  const k = newFieldKey.value.trim();
  if (!k) return "";
  if (k === primaryKeyCol.value) {
    return `"${k}" 是主键字段，数据库禁止修改主键`;
  }
  if (effectiveDbType.value === "ai" && isAiReadOnly(k)) {
    return `"${k}" 是 AI 知识库系统内置字段，禁止添加或修改`;
  }
  if (k in editValues.value) {
    return `字段 "${k}" 已存在于待编辑列表中`;
  }
  return "";
});

function openAddFieldModal() {
  newFieldKey.value = "";
  newFieldType.value = "string";
  showAddFieldModal.value = true;
}

function confirmAddField() {
  const k = newFieldKey.value.trim();
  if (!k || newFieldError.value) return;

  if (newFieldType.value === "boolean") {
    editValues.value[k] = true;
  } else if (newFieldType.value === "number") {
    editValues.value[k] = 0;
  } else {
    editValues.value[k] = "";
  }
  showAddFieldModal.value = false;
}

async function handleSave() {
  errorMessage.value = "";
  showFullError.value = false;
  copiedError.value = false;
  saving.value = true;

  try {
    let updatePayload: Record<string, any> = {};

    if (activeMode.value === "code") {
      updatePayload = JSON.parse(codeModeJson.value);
    } else {
      // Sync object fields first
      for (const [k, err] of Object.entries(objectErrors.value)) {
        if (err) {
          throw new Error(`字段 "${k}" 的 JSON 格式无效，请先修复`);
        }
      }
      for (const [k, v] of Object.entries(editValues.value)) {
        updatePayload[k] = v;
      }
    }

    const pkVal = primaryKeyVal.value;
    if (!pkVal) {
      throw new Error("无法定位当前文档主键 ID，无法保存");
    }

    // In VectorDB (both Base and AI collections), documents strictly require scalar fields
    // and reject nested map/object fields (error 14100).
    // Automatically unpack/flatten any nested objects so they never send invalid map types.
    const finalPayload: Record<string, any> = {};
    for (const [k, v] of Object.entries(updatePayload)) {
      if (v && typeof v === "object" && !Array.isArray(v)) {
        for (const [subK, subV] of Object.entries(v as Record<string, any>)) {
          finalPayload[subK] = subV;
        }
      } else {
        finalPayload[k] = v;
      }
    }

    // Filter out primary key or AI read-only system fields to guarantee
    // server never receives forbidden fields like documentsetName (error 14110).
    const filteredPayload: Record<string, any> = {};
    for (const [k, v] of Object.entries(finalPayload)) {
      if (k === primaryKeyCol.value) continue;
      if (effectiveDbType.value === "ai" && isAiReadOnly(k)) continue;

      // Tencent VectorDB AI collections strictly only support uint64, string, or string element array type.
      // Unsupported bool type triggers error 14100 (e.g. field "is_open" is unsupported bool type...).
      // Automatically coerce boolean to uint64 (1 / 0).
      if (effectiveDbType.value === "ai" && typeof v === "boolean") {
        filteredPayload[k] = v ? 1 : 0;
      } else {
        filteredPayload[k] = v;
      }
    }

    if (Object.keys(filteredPayload).length === 0) {
      throw new Error(
        effectiveDbType.value === "ai"
          ? "未检测到可更新的自定义属性。文件名（documentSetName）与系统内置属性由知识库自动维护，不可修改。请点击“添加新字段”添加自定义属性（如 author、is_enabled 等）后再保存。"
          : "请至少输入一个需要更新的有效字段。"
      );
    }

    await vdbStore.updateDocument(
      props.connId,
      props.database,
      props.collection,
      effectiveDbType.value,
      {
        documentIds: [pkVal],
        documentSetIds: [pkVal],
      },
      filteredPayload
    );

    emit("saved", {
      ...props.document,
      ...filteredPayload,
    });
    emit("close");
  } catch (err: unknown) {
    errorMessage.value = err instanceof Error ? err.message : String(err);
  } finally {
    saving.value = false;
  }
}

function handleEsc() {
  if (!saving.value && !showAddFieldModal.value) {
    close();
  }
}

function close() {
  if (saving.value) return;
  emit("close");
}
</script>
