<template>
  <div class="h-full overflow-y-auto p-6 space-y-6 text-slate-200">
    <!-- Empty State -->
    <div v-if="!meta" class="py-16 text-center text-slate-500">
      <FileSpreadsheet class="w-10 h-10 mx-auto mb-2 text-slate-600 opacity-50" />
      <p class="text-sm">暂无 Schema 元数据信息</p>
    </div>

    <template v-else>
      <!-- Overview Grid -->
      <div>
        <h4 class="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-3">集合概览 (Overview)</h4>
        <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
          <div class="p-3.5 rounded-xl bg-slate-900 border border-slate-800">
            <div class="text-[11px] text-slate-400">所属数据库</div>
            <div class="text-sm font-semibold text-white mt-1 truncate" :title="meta.database">{{ meta.database }}</div>
          </div>
          <div class="p-3.5 rounded-xl bg-slate-900 border border-slate-800">
            <div class="text-[11px] text-slate-400">集合名称</div>
            <div class="text-sm font-semibold text-white mt-1 truncate" :title="meta.collection">{{ meta.collection }}</div>
          </div>
          <div class="p-3.5 rounded-xl bg-slate-900 border border-slate-800">
            <div class="text-[11px] text-slate-400">分片数 (Shards)</div>
            <div class="text-sm font-semibold text-white mt-1 font-mono">{{ meta.shardNum || 1 }}</div>
          </div>
          <div class="p-3.5 rounded-xl bg-slate-900 border border-slate-800">
            <div class="text-[11px] text-slate-400">副本数 (Replicas)</div>
            <div class="text-sm font-semibold text-white mt-1 font-mono">{{ meta.replicaNum || 0 }}</div>
          </div>
        </div>
        <div v-if="meta.description" class="mt-2 text-xs text-slate-400 px-1">
          <span class="font-medium text-slate-300">描述:</span> {{ meta.description }}
        </div>
      </div>

      <!-- Fields Specification Table -->
      <div>
        <h4 class="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-3 flex items-center justify-between">
          <span>字段定义 (Fields)</span>
          <span class="text-[11px] font-normal text-slate-500 font-mono">共 {{ meta.fields?.length || 0 }} 个字段</span>
        </h4>

        <div class="rounded-xl border border-slate-800 bg-slate-900 overflow-hidden shadow-sm">
          <table class="w-full text-left text-xs">
            <thead class="bg-slate-950/70 border-b border-slate-800 text-slate-400 font-medium">
              <tr>
                <th class="px-4 py-3">字段名 (Field Name)</th>
                <th class="px-4 py-3">数据类型 (Type)</th>
                <th class="px-4 py-3">主键 (Primary Key)</th>
                <th class="px-4 py-3">用途 (Usage)</th>
                <th class="px-4 py-3">描述 (Description)</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800/80">
              <tr
                v-for="field in meta.fields"
                :key="field.fieldName"
                class="hover:bg-slate-800/40 transition"
              >
                <td class="px-4 py-3 font-mono font-medium text-white flex items-center gap-2">
                  <KeyRound v-if="field.primaryKey" class="w-3.5 h-3.5 text-amber-400 shrink-0" />
                  <span :class="field.primaryKey ? 'text-amber-300' : ''">{{ field.fieldName }}</span>
                </td>
                <td class="px-4 py-3 font-mono">
                  <span
                    :class="[
                      'px-2 py-0.5 rounded text-[11px] font-medium inline-block',
                      field.fieldType?.toLowerCase().includes('vector')
                        ? 'bg-indigo-950/60 text-indigo-300 border border-indigo-700/50'
                        : 'bg-slate-800 text-slate-300',
                    ]"
                  >
                    {{ field.fieldType }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <span
                    v-if="field.primaryKey"
                    class="px-2 py-0.5 rounded text-[10px] font-semibold bg-amber-950/60 border border-amber-700/50 text-amber-300"
                  >
                    PRIMARY KEY
                  </span>
                  <span v-else class="text-slate-600">-</span>
                </td>
                <td class="px-4 py-3 text-slate-400 font-mono text-[11px]">
                  {{ field.fieldUsage || '-' }}
                </td>
                <td class="px-4 py-3 text-slate-400">
                  {{ field.description || '-' }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Indexes Specification Table -->
      <div>
        <h4 class="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-3 flex items-center justify-between">
          <span>索引配置 (Indexes)</span>
          <span class="text-[11px] font-normal text-slate-500 font-mono">共 {{ meta.indexes?.length || 0 }} 个索引</span>
        </h4>

        <div v-if="!meta.indexes || meta.indexes.length === 0" class="p-4 rounded-xl border border-slate-800 bg-slate-900 text-center text-xs text-slate-500">
          未配置额外索引
        </div>

        <div v-else class="rounded-xl border border-slate-800 bg-slate-900 overflow-hidden shadow-sm">
          <table class="w-full text-left text-xs">
            <thead class="bg-slate-950/70 border-b border-slate-800 text-slate-400 font-medium">
              <tr>
                <th class="px-4 py-3">目标字段</th>
                <th class="px-4 py-3">索引类型 (Index Type)</th>
                <th class="px-4 py-3">度量方式 (Metric Type)</th>
                <th class="px-4 py-3">索引参数 (Parameters)</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800/80">
              <tr
                v-for="(idx, i) in meta.indexes"
                :key="i"
                class="hover:bg-slate-800/40 transition"
              >
                <td class="px-4 py-3 font-mono font-medium text-white">
                  {{ idx.fieldName }}
                </td>
                <td class="px-4 py-3 font-mono">
                  <span
                    :class="[
                      'px-2 py-0.5 rounded text-[11px] font-medium inline-block',
                      idx.indexType?.toLowerCase().includes('vector')
                        ? 'bg-blue-950/60 text-blue-300 border border-blue-700/50'
                        : 'bg-slate-800 text-slate-300',
                    ]"
                  >
                    {{ idx.indexType }}
                  </span>
                </td>
                <td class="px-4 py-3 font-mono">
                  <span
                    v-if="idx.metricType"
                    class="px-2 py-0.5 rounded text-[11px] bg-emerald-950/60 border border-emerald-700/50 text-emerald-300 font-semibold"
                  >
                    {{ idx.metricType }}
                  </span>
                  <span v-else class="text-slate-600">-</span>
                </td>
                <td class="px-4 py-3 font-mono text-[11px] text-slate-400">
                  <div v-if="idx.params && Object.keys(idx.params).length > 0" class="flex flex-wrap gap-1.5">
                    <span
                      v-for="(paramVal, paramKey) in idx.params"
                      :key="paramKey"
                      class="px-1.5 py-0.5 rounded bg-slate-800 border border-slate-700 text-slate-300"
                    >
                      {{ paramKey }}: {{ paramVal }}
                    </span>
                  </div>
                  <span v-else class="text-slate-600">-</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { FileSpreadsheet, KeyRound } from "lucide-vue-next";
import type { CollectionMeta } from "../types";

defineProps<{
  meta: CollectionMeta | null;
}>();
</script>
