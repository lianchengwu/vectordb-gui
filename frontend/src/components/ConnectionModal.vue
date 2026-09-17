<template>
  <div
    v-if="isModalOpen"
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4 animate-fade-in"
  >
    <div class="w-full max-w-xl max-h-[90vh] rounded-xl bg-white border border-slate-200 shadow-2xl overflow-hidden flex flex-col text-slate-800 dark:bg-slate-900 dark:border-slate-700 dark:text-slate-100 transition-colors duration-200">
      <!-- Header -->
      <div class="px-6 py-3.5 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between shrink-0">
        <h3 class="text-base font-semibold tracking-tight text-slate-900 dark:text-white flex items-center gap-2">
          <Database class="w-4 h-4 text-blue-500" />
          {{ isEditing ? '编辑连接配置' : '新建 VectorDB 连接' }}
        </h3>
        <button
          @click="closeModal"
          class="text-slate-400 hover:text-slate-700 dark:hover:text-white p-1 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition"
        >
          <X class="w-4 h-4" />
        </button>
      </div>

      <!-- Tab Switcher -->
      <div class="px-6 pt-3 border-b border-slate-200 dark:border-slate-800 flex gap-4 text-xs font-medium shrink-0 bg-slate-50/50 dark:bg-slate-950/20">
        <button
          type="button"
          @click="activeTab = 'general'"
          :class="[
            'pb-2.5 border-b-2 transition flex items-center gap-1.5',
            activeTab === 'general'
              ? 'border-blue-600 text-blue-600 dark:text-blue-400 dark:border-blue-400 font-semibold'
              : 'border-transparent text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200',
          ]"
        >
          <Sliders class="w-3.5 h-3.5" />
          基础配置
        </button>
        <button
          type="button"
          @click="activeTab = 'network'"
          :class="[
            'pb-2.5 border-b-2 transition flex items-center gap-1.5',
            activeTab === 'network'
              ? 'border-blue-600 text-blue-600 dark:text-blue-400 dark:border-blue-400 font-semibold'
              : 'border-transparent text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200',
          ]"
        >
          <Network class="w-3.5 h-3.5" />
          网络代理 & SSH 隧道
          <span
            v-if="networkMode !== 'direct'"
            class="w-2 h-2 rounded-full bg-blue-500 inline-block"
          ></span>
        </button>
      </div>

      <!-- Form Body -->
      <div class="flex-1 overflow-y-auto px-6 py-4 space-y-4 text-xs">
        <!-- GENERAL TAB -->
        <div v-show="activeTab === 'general'" class="space-y-3.5">
          <!-- Name -->
          <div>
            <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">连接名称</label>
            <input
              v-model="form.name"
              type="text"
              placeholder="例如: 测试环境集群 / Prod Cluster"
              class="w-full px-3 py-2 rounded-lg bg-slate-50 border border-slate-300 text-slate-900 placeholder-slate-400 dark:bg-slate-800/80 dark:border-slate-700 dark:text-white dark:placeholder-slate-500 focus:outline-none focus:ring-1 focus:ring-blue-500 transition shadow-inner"
            />
          </div>

          <!-- URL -->
          <div>
            <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">
              实例地址 (URL) <span class="text-rose-500">*</span>
            </label>
            <input
              v-model="form.url"
              type="text"
              placeholder="http://10.0.x.x:80 或公网域名:端口"
              class="w-full px-3 py-2 rounded-lg bg-slate-50 border border-slate-300 text-slate-900 placeholder-slate-400 dark:bg-slate-800/80 dark:border-slate-700 dark:text-white dark:placeholder-slate-500 focus:outline-none focus:ring-1 focus:ring-blue-500 transition font-mono text-xs shadow-inner"
            />
            <p class="text-[11px] text-slate-500 dark:text-slate-400 mt-1">例如: http://lb-xxxx.clb.ap-guangzhou.tencentclb.com:50000</p>
          </div>

          <!-- Username & Timeout Row -->
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">
                账号 (Username) <span class="text-rose-500">*</span>
              </label>
              <input
                v-model="form.username"
                type="text"
                placeholder="默认 root"
                class="w-full px-3 py-2 rounded-lg bg-slate-50 border border-slate-300 text-slate-900 placeholder-slate-400 dark:bg-slate-800/80 dark:border-slate-700 dark:text-white dark:placeholder-slate-500 focus:outline-none focus:ring-1 focus:ring-blue-500 transition shadow-inner"
              />
            </div>
            <div>
              <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">超时时间 (秒)</label>
              <input
                v-model.number="form.timeout"
                type="number"
                min="1"
                max="300"
                class="w-full px-3 py-2 rounded-lg bg-slate-50 border border-slate-300 text-slate-900 placeholder-slate-400 dark:bg-slate-800/80 dark:border-slate-700 dark:text-white dark:placeholder-slate-500 focus:outline-none focus:ring-1 focus:ring-blue-500 transition shadow-inner"
              />
            </div>
          </div>

          <!-- API Key -->
          <div>
            <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">
              API Key (密码凭证) <span class="text-rose-500">*</span>
            </label>
            <div class="relative">
              <input
                v-model="form.apiKey"
                :type="showApiKey ? 'text' : 'password'"
                placeholder="腾讯云控制台获取的 API Key"
                class="w-full pl-3 pr-10 py-2 rounded-lg bg-slate-50 border border-slate-300 text-slate-900 placeholder-slate-400 dark:bg-slate-800/80 dark:border-slate-700 dark:text-white dark:placeholder-slate-500 focus:outline-none focus:ring-1 focus:ring-blue-500 transition font-mono text-xs shadow-inner"
              />
              <button
                type="button"
                @click="showApiKey = !showApiKey"
                class="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-700 dark:hover:text-white p-1"
              >
                <EyeOff v-if="showApiKey" class="w-3.5 h-3.5" />
                <Eye v-else class="w-3.5 h-3.5" />
              </button>
            </div>
          </div>
        </div>

        <!-- NETWORK & PROXY TAB -->
        <div v-show="activeTab === 'network'" class="space-y-4">
          <!-- Mode Selection -->
          <div>
            <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1.5">网络通路模式</label>
            <div class="grid grid-cols-4 gap-1.5 p-1 rounded-lg bg-slate-100 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-[11px]">
              <button
                type="button"
                @click="networkMode = 'direct'"
                :class="[
                  'py-1.5 rounded-md font-medium transition text-center',
                  networkMode === 'direct'
                    ? 'bg-white dark:bg-slate-700 text-blue-600 dark:text-white shadow-sm'
                    : 'text-slate-500 dark:text-slate-400 hover:text-slate-800 dark:hover:text-white',
                ]"
              >
                直连
              </button>
              <button
                type="button"
                @click="networkMode = 'http'"
                :class="[
                  'py-1.5 rounded-md font-medium transition text-center',
                  networkMode === 'http'
                    ? 'bg-white dark:bg-slate-700 text-blue-600 dark:text-white shadow-sm'
                    : 'text-slate-500 dark:text-slate-400 hover:text-slate-800 dark:hover:text-white',
                ]"
              >
                HTTP 代理
              </button>
              <button
                type="button"
                @click="networkMode = 'socks5'"
                :class="[
                  'py-1.5 rounded-md font-medium transition text-center',
                  networkMode === 'socks5'
                    ? 'bg-white dark:bg-slate-700 text-blue-600 dark:text-white shadow-sm'
                    : 'text-slate-500 dark:text-slate-400 hover:text-slate-800 dark:hover:text-white',
                ]"
              >
                SOCKS5 代理
              </button>
              <button
                type="button"
                @click="networkMode = 'ssh'"
                :class="[
                  'py-1.5 rounded-md font-medium transition text-center',
                  networkMode === 'ssh'
                    ? 'bg-white dark:bg-slate-700 text-blue-600 dark:text-white shadow-sm'
                    : 'text-slate-500 dark:text-slate-400 hover:text-slate-800 dark:hover:text-white',
                ]"
              >
                SSH 隧道 (跳板机)
              </button>
            </div>
          </div>

          <!-- Direct Info -->
          <div v-if="networkMode === 'direct'" class="p-3 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-800/40 text-slate-500 dark:text-slate-400 text-xs">
            <p>使用本机网络直接访问目标 VectorDB 实例地址（适用于拥有公网 IP 或直通专线/VPN 的场景）。</p>
          </div>

          <!-- HTTP / SOCKS5 Proxy Form -->
          <div v-if="networkMode === 'http' || networkMode === 'socks5'" class="space-y-3 p-3.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30">
            <div class="grid grid-cols-3 gap-3">
              <div class="col-span-2">
                <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">代理服务器主机 (Host) <span class="text-rose-500">*</span></label>
                <input
                  v-model="proxyForm.host"
                  type="text"
                  placeholder="127.0.0.1 或 proxy.corp.internal"
                  class="w-full px-3 py-1.5 rounded-lg bg-white border border-slate-300 dark:bg-slate-800 dark:border-slate-700 text-slate-900 dark:text-white text-xs font-mono focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
              <div>
                <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">端口 (Port)</label>
                <input
                  v-model.number="proxyForm.port"
                  type="number"
                  :placeholder="networkMode === 'socks5' ? '1080' : '8080'"
                  class="w-full px-3 py-1.5 rounded-lg bg-white border border-slate-300 dark:bg-slate-800 dark:border-slate-700 text-slate-900 dark:text-white text-xs font-mono focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
            </div>

            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">用户名 (可选)</label>
                <input
                  v-model="proxyForm.username"
                  type="text"
                  placeholder="代理账号"
                  class="w-full px-3 py-1.5 rounded-lg bg-white border border-slate-300 dark:bg-slate-800 dark:border-slate-700 text-slate-900 dark:text-white text-xs focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
              <div>
                <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">密码 (可选)</label>
                <input
                  v-model="proxyForm.password"
                  type="password"
                  placeholder="代理密码"
                  class="w-full px-3 py-1.5 rounded-lg bg-white border border-slate-300 dark:bg-slate-800 dark:border-slate-700 text-slate-900 dark:text-white text-xs focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
            </div>
          </div>

          <!-- SSH Tunnel Form -->
          <div v-if="networkMode === 'ssh'" class="space-y-3 p-3.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30">
            <div class="grid grid-cols-3 gap-3">
              <div class="col-span-2">
                <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">跳板机地址 (SSH Host) <span class="text-rose-500">*</span></label>
                <input
                  v-model="sshForm.host"
                  type="text"
                  placeholder="123.x.x.x 或 bastion.example.com"
                  class="w-full px-3 py-1.5 rounded-lg bg-white border border-slate-300 dark:bg-slate-800 dark:border-slate-700 text-slate-900 dark:text-white text-xs font-mono focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
              <div>
                <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">端口 (Port)</label>
                <input
                  v-model.number="sshForm.port"
                  type="number"
                  placeholder="22"
                  class="w-full px-3 py-1.5 rounded-lg bg-white border border-slate-300 dark:bg-slate-800 dark:border-slate-700 text-slate-900 dark:text-white text-xs font-mono focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
            </div>

            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">SSH 用户名 <span class="text-rose-500">*</span></label>
                <input
                  v-model="sshForm.user"
                  type="text"
                  placeholder="例如: root / ubuntu"
                  class="w-full px-3 py-1.5 rounded-lg bg-white border border-slate-300 dark:bg-slate-800 dark:border-slate-700 text-slate-900 dark:text-white text-xs focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
              <div>
                <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">认证方式</label>
                <select
                  v-model="sshForm.authType"
                  class="w-full px-3 py-1.5 rounded-lg bg-white border border-slate-300 dark:bg-slate-800 dark:border-slate-700 text-slate-900 dark:text-white text-xs focus:outline-none"
                >
                  <option value="password">账号密码 (Password)</option>
                  <option value="key">私钥文件 / 内容 (Private Key)</option>
                </select>
              </div>
            </div>

            <div v-if="sshForm.authType === 'password'">
              <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">SSH 密码 <span class="text-rose-500">*</span></label>
              <input
                v-model="sshForm.password"
                type="password"
                placeholder="跳板机登录密码"
                class="w-full px-3 py-1.5 rounded-lg bg-white border border-slate-300 dark:bg-slate-800 dark:border-slate-700 text-slate-900 dark:text-white text-xs focus:outline-none focus:ring-1 focus:ring-blue-500 font-mono"
              />
            </div>

            <div v-if="sshForm.authType === 'key'" class="space-y-2">
              <div>
                <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">私钥 (文件路径 或 PEM 文本) <span class="text-rose-500">*</span></label>
                <textarea
                  v-model="sshForm.privateKey"
                  rows="3"
                  placeholder="~/.ssh/id_rsa 或直接粘贴 -----BEGIN OPENSSH PRIVATE KEY-----"
                  class="w-full px-3 py-1.5 rounded-lg bg-white border border-slate-300 dark:bg-slate-800 dark:border-slate-700 text-slate-900 dark:text-white text-[11px] font-mono focus:outline-none focus:ring-1 focus:ring-blue-500 resize-y"
                ></textarea>
              </div>
              <div>
                <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">私钥密码 (Passphrase，可选)</label>
                <input
                  v-model="sshForm.passphrase"
                  type="password"
                  placeholder="若私钥被加密保护请输入密码"
                  class="w-full px-3 py-1.5 rounded-lg bg-white border border-slate-300 dark:bg-slate-800 dark:border-slate-700 text-slate-900 dark:text-white text-xs focus:outline-none focus:ring-1 focus:ring-blue-500 font-mono"
                />
              </div>
            </div>
          </div>
        </div>

        <!-- Test result banner -->
        <div
          v-if="testResult"
          :class="[
            'p-2.5 rounded-lg border text-xs flex items-start gap-2.5 transition shadow-sm',
            testResult.success
              ? 'bg-emerald-50 border-emerald-200 text-emerald-800 dark:bg-emerald-950/40 dark:border-emerald-800/60 dark:text-emerald-300'
              : 'bg-rose-50 border-rose-200 text-rose-800 dark:bg-rose-950/40 dark:border-rose-800/60 dark:text-rose-300',
          ]"
        >
          <CheckCircle2 v-if="testResult.success" class="w-4 h-4 mt-0.5 shrink-0 text-emerald-500 dark:text-emerald-400" />
          <AlertCircle v-else class="w-4 h-4 mt-0.5 shrink-0 text-rose-500 dark:text-rose-400" />
          <div class="flex-1 break-all">
            <span class="font-semibold">{{ testResult.success ? '测试连接成功' : '测试连接失败' }}:</span>
            {{ testResult.message }}
          </div>
        </div>
      </div>

      <!-- Actions Footer -->
      <div class="px-6 py-3.5 bg-slate-50 dark:bg-slate-900/90 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between shrink-0">
        <button
          type="button"
          @click="onTest"
          :disabled="isTesting || !form.url || !form.apiKey"
          class="px-3 py-1.5 rounded-lg text-xs font-medium border border-slate-300 bg-white hover:bg-slate-100 text-slate-700 dark:border-slate-700 dark:bg-slate-800 dark:hover:bg-slate-750 dark:text-slate-200 transition flex items-center gap-1.5 disabled:opacity-50 disabled:cursor-not-allowed shadow-sm"
        >
          <Loader2 v-if="isTesting" class="w-3.5 h-3.5 animate-spin" />
          <Radio v-else class="w-3.5 h-3.5 text-blue-500" />
          测试连通性
        </button>

        <div class="flex items-center gap-2">
          <button
            type="button"
            @click="closeModal"
            class="px-3.5 py-1.5 rounded-lg text-xs font-medium text-slate-600 hover:text-slate-900 hover:bg-slate-200/60 dark:text-slate-400 dark:hover:text-white dark:hover:bg-slate-800 transition"
          >
            取消
          </button>
          <button
            type="button"
            @click="onSave"
            :disabled="isSaving || !form.url || !form.apiKey"
            class="px-4 py-1.5 rounded-lg text-xs font-medium bg-blue-600 hover:bg-blue-500 text-white transition flex items-center gap-1.5 disabled:opacity-50 disabled:cursor-not-allowed shadow-sm"
          >
            <Loader2 v-if="isSaving" class="w-3.5 h-3.5 animate-spin" />
            <Save v-else class="w-3.5 h-3.5" />
            保存配置
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from "vue";
import {
  Database,
  X,
  Eye,
  EyeOff,
  Radio,
  Save,
  CheckCircle2,
  AlertCircle,
  Loader2,
  Sliders,
  Network,
} from "lucide-vue-next";
import { useConnectionStore } from "../stores/connection";
import type { ConnectionConfig, TestResult } from "../types";

const store = useConnectionStore();
const { isModalOpen, editingConnection, closeModal, save, test } = store;

const activeTab = ref<"general" | "network">("general");
const networkMode = ref<"direct" | "http" | "socks5" | "ssh">("direct");

const showApiKey = ref(false);
const isTesting = ref(false);
const isSaving = ref(false);
const testResult = ref<TestResult | null>(null);

const form = ref<ConnectionConfig>({
  id: "",
  name: "",
  url: "http://",
  username: "root",
  apiKey: "",
  timeout: 10,
});

const proxyForm = ref({
  host: "",
  port: 8080,
  username: "",
  password: "",
});

const sshForm = ref({
  host: "",
  port: 22,
  user: "root",
  authType: "password",
  password: "",
  privateKey: "",
  passphrase: "",
});

const isEditing = computed(() => !!editingConnection.value?.id);

watch(
  () => editingConnection.value,
  (val) => {
    if (val) {
      form.value = {
        id: val.id || "",
        name: val.name || "",
        url: val.url || "http://",
        username: val.username || "root",
        apiKey: val.apiKey || "",
        timeout: val.timeout || 10,
      };

      // Populate proxy & SSH tunnel state
      if (val.sshTunnel?.enabled) {
        networkMode.value = "ssh";
        sshForm.value = {
          host: val.sshTunnel.host || "",
          port: val.sshTunnel.port || 22,
          user: val.sshTunnel.user || "root",
          authType: val.sshTunnel.authType || "password",
          password: val.sshTunnel.password || "",
          privateKey: val.sshTunnel.privateKey || "",
          passphrase: val.sshTunnel.passphrase || "",
        };
      } else if (val.proxy?.enabled) {
        networkMode.value = val.proxy.type === "socks5" ? "socks5" : "http";
        proxyForm.value = {
          host: val.proxy.host || "",
          port: val.proxy.port || (val.proxy.type === "socks5" ? 1080 : 8080),
          username: val.proxy.username || "",
          password: val.proxy.password || "",
        };
      } else {
        networkMode.value = "direct";
      }

      activeTab.value = "general";
      testResult.value = null;
      showApiKey.value = false;
    }
  },
  { immediate: true }
);

function buildFullConfig(): ConnectionConfig {
  const config: ConnectionConfig = {
    ...form.value,
  };

  if (networkMode.value === "ssh") {
    config.sshTunnel = {
      enabled: true,
      host: sshForm.value.host.trim(),
      port: sshForm.value.port || 22,
      user: sshForm.value.user.trim() || "root",
      authType: sshForm.value.authType,
      password: sshForm.value.password,
      privateKey: sshForm.value.privateKey,
      passphrase: sshForm.value.passphrase,
    };
    config.proxy = {
      enabled: false,
      type: "",
      host: "",
      port: 0,
    };
  } else if (networkMode.value === "http" || networkMode.value === "socks5") {
    config.proxy = {
      enabled: true,
      type: networkMode.value,
      host: proxyForm.value.host.trim(),
      port: proxyForm.value.port || (networkMode.value === "socks5" ? 1080 : 8080),
      username: proxyForm.value.username,
      password: proxyForm.value.password,
    };
    config.sshTunnel = {
      enabled: false,
      host: "",
      port: 22,
      user: "root",
      authType: "password",
    };
  } else {
    config.proxy = {
      enabled: false,
      type: "",
      host: "",
      port: 0,
    };
    config.sshTunnel = {
      enabled: false,
      host: "",
      port: 22,
      user: "root",
      authType: "password",
    };
  }

  return config;
}

async function onTest() {
  isTesting.value = true;
  testResult.value = null;
  try {
    const fullConfig = buildFullConfig();
    const res = await test(fullConfig);
    testResult.value = res;
  } finally {
    isTesting.value = false;
  }
}

async function onSave() {
  if (!form.value.url || !form.value.apiKey) return;
  if (!form.value.name.trim()) {
    form.value.name = form.value.url.replace(/^https?:\/\//, "");
  }
  isSaving.value = true;
  try {
    const fullConfig = buildFullConfig();
    await save(fullConfig);
  } finally {
    isSaving.value = false;
  }
}
</script>
