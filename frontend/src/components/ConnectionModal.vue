<template>
  <div
    v-if="isModalOpen"
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4 animate-fade-in"
  >
    <div class="w-full max-w-2xl max-h-[92vh] rounded-xl bg-white border border-slate-200 shadow-2xl overflow-hidden flex flex-col text-slate-800 dark:bg-slate-900 dark:border-slate-700 dark:text-slate-100 transition-colors duration-200">
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
          网络链路拓扑 (代理链 / 隧道链)
          <span
            v-if="enabledHopsCount > 0"
            class="px-1.5 py-0.2 rounded-full text-[9px] bg-blue-600 text-white font-mono"
          >
            {{ enabledHopsCount }}
          </span>
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

          <!-- Visible Databases Filter -->
          <div class="pt-3 border-t border-slate-200/80 dark:border-slate-800 space-y-2">
            <div class="flex items-center justify-between">
              <div>
                <label class="block font-medium text-slate-700 dark:text-slate-300">
                  可显示数据库 (Visible Databases)
                </label>
                <p class="text-[11px] text-slate-500 dark:text-slate-400 mt-0.5">
                  设置在左侧导航树中允许显示的数据库。留空则默认展示集群中全部数据库。
                </p>
              </div>
              <button
                type="button"
                @click="onFetchDatabases"
                :disabled="isFetchingDbs || !form.url || !form.apiKey"
                class="px-2.5 py-1 rounded-md text-[11px] font-medium border border-blue-200 dark:border-blue-800/60 bg-blue-50/80 hover:bg-blue-100 text-blue-700 dark:bg-blue-950/40 dark:hover:bg-blue-900/50 dark:text-blue-300 transition flex items-center gap-1.5 disabled:opacity-50 disabled:cursor-not-allowed shrink-0 shadow-xs"
                title="连接集群并拉取现有数据库列表"
              >
                <Loader2 v-if="isFetchingDbs" class="w-3 h-3 animate-spin text-blue-600" />
                <RefreshCw v-else class="w-3 h-3 text-blue-600 dark:text-blue-400" />
                <span>{{ isFetchingDbs ? '拉取中...' : '拉取集群数据库' }}</span>
              </button>
            </div>

            <!-- Manual input field (comma-separated) -->
            <div>
              <input
                v-model="databaseInput"
                @blur="syncFromDatabaseInput"
                @keydown.enter.prevent="syncFromDatabaseInput"
                type="text"
                placeholder="支持逗号分隔手动输入，例如: db_test, db_prod"
                class="w-full px-3 py-2 rounded-lg bg-slate-50 border border-slate-300 text-slate-900 placeholder-slate-400 dark:bg-slate-800/80 dark:border-slate-700 dark:text-white dark:placeholder-slate-500 focus:outline-none focus:ring-1 focus:ring-blue-500 transition font-mono text-xs shadow-inner"
              />
            </div>

            <!-- Fetch error banner if any -->
            <div v-if="fetchDbError" class="p-2 rounded bg-rose-50 border border-rose-200 text-rose-700 dark:bg-rose-950/40 dark:border-rose-900 dark:text-rose-300 text-[11px] flex items-center gap-1.5">
              <AlertCircle class="w-3.5 h-3.5 shrink-0" />
              <span class="truncate">{{ fetchDbError }}</span>
            </div>

            <!-- Cluster databases chips / selection area -->
            <div v-if="availableDatabases.length > 0" class="p-3 rounded-lg bg-slate-50 dark:bg-slate-950/40 border border-slate-200 dark:border-slate-800 space-y-2">
              <div class="flex items-center justify-between text-[11px]">
                <span class="text-slate-500 dark:text-slate-400 font-medium flex items-center gap-1">
                  <Database class="w-3 h-3 text-slate-400" />
                  <span>集群中检测到 {{ availableDatabases.length }} 个数据库（点击勾选需要显示的库）：</span>
                </span>
                <div class="flex items-center gap-2">
                  <button
                    type="button"
                    @click="selectAllDbs"
                    class="text-blue-600 hover:text-blue-500 dark:text-blue-400 font-medium cursor-pointer"
                  >
                    全选
                  </button>
                  <span class="text-slate-300 dark:text-slate-700">|</span>
                  <button
                    type="button"
                    @click="clearAllDbs"
                    class="text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200 cursor-pointer"
                  >
                    清空 (全部展示)
                  </button>
                </div>
              </div>

              <!-- Database chips -->
              <div class="flex flex-wrap gap-1.5 max-h-36 overflow-y-auto p-1">
                <button
                  v-for="db in availableDatabases"
                  :key="db.name"
                  type="button"
                  @click="toggleDbSelection(db.name)"
                  :class="[
                    'px-2.5 py-1 rounded-md text-xs font-mono transition flex items-center gap-1.5 border cursor-pointer select-none',
                    isDbSelected(db.name)
                      ? 'bg-blue-50 border-blue-300 text-blue-700 dark:bg-blue-950/60 dark:border-blue-700 dark:text-blue-300 shadow-xs'
                      : 'bg-white border-slate-200 text-slate-600 hover:border-slate-300 dark:bg-slate-900 dark:border-slate-700/80 dark:text-slate-400 dark:hover:border-slate-600'
                  ]"
                >
                  <Check v-if="isDbSelected(db.name)" class="w-3 h-3 text-blue-600 dark:text-blue-400" />
                  <span v-else class="w-3 h-3 rounded-full border border-slate-300 dark:border-slate-600 shrink-0"></span>
                  <span>{{ db.name }}</span>
                  <span
                    :class="[
                      'px-1 py-0.2 rounded text-[9px] font-sans font-medium',
                      db.dbType === 'ai'
                        ? 'bg-purple-100 text-purple-700 dark:bg-purple-950 dark:text-purple-300'
                        : 'bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400'
                    ]"
                  >
                    {{ db.dbType === 'ai' ? 'AI' : 'BASE' }}
                  </span>
                </button>
              </div>
            </div>

            <!-- Selected summary -->
            <div v-if="selectedDatabases.length > 0" class="text-[11px] text-blue-600 dark:text-blue-400 flex items-center gap-1 font-mono">
              <span>已选 {{ selectedDatabases.length }} 个可见库：</span>
              <span class="font-medium truncate">{{ selectedDatabases.join(', ') }}</span>
            </div>
            <div v-else class="text-[11px] text-slate-400 dark:text-slate-500">
              当前未限制可见数据库（将展示该集群下所有数据库）。
            </div>
          </div>
        </div>

        <!-- NETWORK CHAIN TAB (Multi-Hop User-Defined Chain) -->
        <div v-show="activeTab === 'network'" class="space-y-4">
          <!-- Dynamic Pipeline Breadcrumb Banner -->
          <div class="p-3 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/90 dark:bg-slate-950/50 shadow-inner">
            <div class="text-[11px] font-medium text-slate-500 dark:text-slate-400 mb-2 flex items-center justify-between">
              <span class="flex items-center gap-1.5">
                <Network class="w-3.5 h-3.5 text-blue-500" />
                <span>当前网络链路 (自左向右依次穿透)</span>
              </span>
              <span class="text-[10px] text-slate-400 font-mono">
                共 {{ enabledHopsCount }} 个有效穿透节点
              </span>
            </div>

            <!-- Pipeline Chain Display -->
            <div class="flex items-center gap-1.5 text-[11px] font-mono flex-wrap">
              <span class="px-2 py-1 rounded bg-slate-200 dark:bg-slate-800 text-slate-700 dark:text-slate-300 font-medium">本机</span>
              <span class="text-slate-400">➔</span>

              <template v-for="(hop, idx) in chain" :key="hop.id">
                <template v-if="hop.enabled && hop.host">
                  <span
                    :class="[
                      'px-2 py-1 rounded border font-semibold flex items-center gap-1 shadow-sm',
                      hop.type === 'ssh'
                        ? 'bg-purple-50 border-purple-200 text-purple-700 dark:bg-purple-950/70 dark:border-purple-700/60 dark:text-purple-300'
                        : hop.type === 'socks5'
                        ? 'bg-cyan-50 border-cyan-200 text-cyan-700 dark:bg-cyan-950/70 dark:border-cyan-700/60 dark:text-cyan-300'
                        : 'bg-blue-50 border-blue-200 text-blue-700 dark:bg-blue-950/70 dark:border-blue-700/60 dark:text-blue-300'
                    ]"
                  >
                    <span>#{{ idx + 1 }} {{ hop.type.toUpperCase() }}</span>
                    <span class="text-[9px] font-normal opacity-80 truncate max-w-[100px]">({{ hop.host }})</span>
                  </span>
                  <span class="text-slate-400">➔</span>
                </template>
              </template>

              <span v-if="enabledHopsCount === 0" class="px-2 py-0.5 rounded bg-emerald-50 border border-emerald-200 text-emerald-700 dark:bg-emerald-950/40 dark:border-emerald-800 dark:text-emerald-300 text-[10px]">
                直接连通 (无代理)
              </span>
              <span v-if="enabledHopsCount === 0" class="text-slate-400">➔</span>

              <span class="px-2 py-1 rounded bg-slate-200 dark:bg-slate-800 text-slate-700 dark:text-slate-300 font-semibold truncate max-w-[140px]" :title="form.url || '目标 VectorDB'">
                {{ form.url ? form.url.replace(/^https?:\/\//, '') : '目标 VectorDB' }}
              </span>
            </div>
          </div>

          <!-- Add Hop Toolbar -->
          <div class="flex items-center justify-between">
            <span class="font-medium text-slate-700 dark:text-slate-300 text-xs">代理 & 隧道节点列表</span>
            <div class="flex items-center gap-1.5">
              <button
                type="button"
                @click="addHop('http')"
                class="px-2.5 py-1 rounded-md border border-slate-300 bg-white hover:bg-slate-50 text-slate-700 dark:border-slate-700 dark:bg-slate-800 dark:hover:bg-slate-750 dark:text-slate-200 transition text-[11px] font-medium flex items-center gap-1 shadow-sm"
              >
                <Plus class="w-3 h-3 text-blue-500" />
                + HTTP 代理
              </button>
              <button
                type="button"
                @click="addHop('socks5')"
                class="px-2.5 py-1 rounded-md border border-slate-300 bg-white hover:bg-slate-50 text-slate-700 dark:border-slate-700 dark:bg-slate-800 dark:hover:bg-slate-750 dark:text-slate-200 transition text-[11px] font-medium flex items-center gap-1 shadow-sm"
              >
                <Plus class="w-3 h-3 text-cyan-500" />
                + SOCKS5 代理
              </button>
              <button
                type="button"
                @click="addHop('ssh')"
                class="px-2.5 py-1 rounded-md border border-slate-300 bg-white hover:bg-slate-50 text-slate-700 dark:border-slate-700 dark:bg-slate-800 dark:hover:bg-slate-750 dark:text-slate-200 transition text-[11px] font-medium flex items-center gap-1 shadow-sm"
              >
                <Plus class="w-3 h-3 text-purple-500" />
                + SSH 隧道
              </button>
            </div>
          </div>

          <!-- Empty Chain State -->
          <div
            v-if="chain.length === 0"
            class="p-6 text-center rounded-xl border border-dashed border-slate-300 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/30 text-slate-400"
          >
            <p class="text-xs">暂无配置代理/隧道节点，当前为直连模式。</p>
            <p class="text-[11px] text-slate-500 mt-1">如需多层穿透，可点击右上角按钮自由添加 HTTP、SOCKS5 代理或 SSH 跳板机并任意调整顺序。</p>
          </div>

          <!-- Hop Cards List -->
          <div v-else class="space-y-3">
            <div
              v-for="(hop, index) in chain"
              :key="hop.id"
              class="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/80 shadow-sm overflow-hidden transition"
            >
              <!-- Card Header -->
              <div class="px-4 py-2.5 bg-slate-50/80 dark:bg-slate-800/50 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between">
                <div class="flex items-center gap-2">
                  <!-- Enable/Disable Checkbox -->
                  <input
                    type="checkbox"
                    v-model="hop.enabled"
                    class="w-3.5 h-3.5 rounded text-blue-600 focus:ring-blue-500 cursor-pointer"
                    title="启用/禁用该节点"
                  />

                  <!-- Sequence Number and Type Badge -->
                  <span class="font-mono font-bold text-xs text-slate-500 dark:text-slate-400">#{{ index + 1 }}</span>
                  <span
                    :class="[
                      'px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider',
                      hop.type === 'ssh'
                        ? 'bg-purple-100 text-purple-700 dark:bg-purple-950 dark:text-purple-300 border border-purple-300 dark:border-purple-800'
                        : hop.type === 'socks5'
                        ? 'bg-cyan-100 text-cyan-700 dark:bg-cyan-950 dark:text-cyan-300 border border-cyan-300 dark:border-cyan-800'
                        : 'bg-blue-100 text-blue-700 dark:bg-blue-950 dark:text-blue-300 border border-blue-300 dark:border-blue-800'
                    ]"
                  >
                    {{ hop.type === 'ssh' ? 'SSH 跳板机' : `${hop.type.toUpperCase()} 代理` }}
                  </span>

                  <span v-if="hop.host" class="text-xs text-slate-600 dark:text-slate-400 font-mono truncate max-w-[200px]">
                    {{ hop.host }}:{{ hop.port }}
                  </span>
                </div>

                <!-- Reorder and Delete Toolbar -->
                <div class="flex items-center gap-1">
                  <!-- Move Up -->
                  <button
                    type="button"
                    @click="moveHop(index, -1)"
                    :disabled="index === 0"
                    class="p-1 rounded hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-500 hover:text-slate-800 dark:hover:text-white disabled:opacity-30 disabled:cursor-not-allowed transition"
                    title="上移顺序 (往前穿透)"
                  >
                    <ArrowUp class="w-3.5 h-3.5" />
                  </button>

                  <!-- Move Down -->
                  <button
                    type="button"
                    @click="moveHop(index, 1)"
                    :disabled="index === chain.length - 1"
                    class="p-1 rounded hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-500 hover:text-slate-800 dark:hover:text-white disabled:opacity-30 disabled:cursor-not-allowed transition"
                    title="下移顺序 (往后穿透)"
                  >
                    <ArrowDown class="w-3.5 h-3.5" />
                  </button>

                  <!-- Delete -->
                  <button
                    type="button"
                    @click="removeHop(index)"
                    class="p-1 rounded hover:bg-rose-50 text-slate-400 hover:text-rose-600 dark:hover:bg-rose-950/40 dark:hover:text-rose-400 transition ml-1"
                    title="移除该节点"
                  >
                    <Trash2 class="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>

              <!-- Card Form Content -->
              <div v-show="hop.enabled" class="p-4 space-y-3">
                <!-- HTTP / SOCKS5 Inputs -->
                <template v-if="hop.type === 'http' || hop.type === 'socks5'">
                  <div class="grid grid-cols-3 gap-3">
                    <div class="col-span-2">
                      <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">代理主机 (Host) <span class="text-rose-500">*</span></label>
                      <input
                        v-model="hop.host"
                        type="text"
                        placeholder="127.0.0.1 或 proxy.internal"
                        class="w-full px-3 py-1.5 rounded-lg bg-slate-50 border border-slate-300 dark:bg-slate-800 dark:border-slate-700 text-slate-900 dark:text-white text-xs font-mono focus:outline-none focus:ring-1 focus:ring-blue-500"
                      />
                    </div>
                    <div>
                      <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">端口 (Port)</label>
                      <input
                        v-model.number="hop.port"
                        type="number"
                        :placeholder="hop.type === 'socks5' ? '1080' : '8080'"
                        class="w-full px-3 py-1.5 rounded-lg bg-slate-50 border border-slate-300 dark:bg-slate-800 dark:border-slate-700 text-slate-900 dark:text-white text-xs font-mono focus:outline-none focus:ring-1 focus:ring-blue-500"
                      />
                    </div>
                  </div>

                  <div class="grid grid-cols-2 gap-3">
                    <div>
                      <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">代理账号 (可选)</label>
                      <input
                        v-model="hop.username"
                        type="text"
                        placeholder="账号"
                        class="w-full px-3 py-1.5 rounded-lg bg-slate-50 border border-slate-300 dark:bg-slate-800 dark:border-slate-700 text-slate-900 dark:text-white text-xs focus:outline-none focus:ring-1 focus:ring-blue-500"
                      />
                    </div>
                    <div>
                      <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">代理密码 (可选)</label>
                      <input
                        v-model="hop.password"
                        type="password"
                        placeholder="密码"
                        class="w-full px-3 py-1.5 rounded-lg bg-slate-50 border border-slate-300 dark:bg-slate-800 dark:border-slate-700 text-slate-900 dark:text-white text-xs focus:outline-none focus:ring-1 focus:ring-blue-500"
                      />
                    </div>
                  </div>
                </template>

                <!-- SSH Tunnel Inputs -->
                <template v-if="hop.type === 'ssh'">
                  <div class="grid grid-cols-3 gap-3">
                    <div class="col-span-2">
                      <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">跳板机主机 (Host) <span class="text-rose-500">*</span></label>
                      <input
                        v-model="hop.host"
                        type="text"
                        placeholder="123.x.x.x 或 bastion.example.com"
                        class="w-full px-3 py-1.5 rounded-lg bg-slate-50 border border-slate-300 dark:bg-slate-800 dark:border-slate-700 text-slate-900 dark:text-white text-xs font-mono focus:outline-none focus:ring-1 focus:ring-blue-500"
                      />
                    </div>
                    <div>
                      <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">端口 (Port)</label>
                      <input
                        v-model.number="hop.port"
                        type="number"
                        placeholder="22"
                        class="w-full px-3 py-1.5 rounded-lg bg-slate-50 border border-slate-300 dark:bg-slate-800 dark:border-slate-700 text-slate-900 dark:text-white text-xs font-mono focus:outline-none focus:ring-1 focus:ring-blue-500"
                      />
                    </div>
                  </div>

                  <div class="grid grid-cols-2 gap-3">
                    <div>
                      <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">SSH 用户名 <span class="text-rose-500">*</span></label>
                      <input
                        v-model="hop.username"
                        type="text"
                        placeholder="例如: root / ubuntu"
                        class="w-full px-3 py-1.5 rounded-lg bg-slate-50 border border-slate-300 dark:bg-slate-800 dark:border-slate-700 text-slate-900 dark:text-white text-xs focus:outline-none focus:ring-1 focus:ring-blue-500"
                      />
                    </div>
                    <div>
                      <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">认证方式</label>
                      <select
                        v-model="hop.authType"
                        class="w-full px-3 py-1.5 rounded-lg bg-slate-50 border border-slate-300 dark:bg-slate-800 dark:border-slate-700 text-slate-900 dark:text-white text-xs focus:outline-none"
                      >
                        <option value="password">账号密码 (Password)</option>
                        <option value="key">私钥文件 / 内容 (Private Key)</option>
                      </select>
                    </div>
                  </div>

                  <div v-if="hop.authType === 'password'">
                    <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">SSH 密码 <span class="text-rose-500">*</span></label>
                    <input
                      v-model="hop.password"
                      type="password"
                      placeholder="跳板机登录密码"
                      class="w-full px-3 py-1.5 rounded-lg bg-slate-50 border border-slate-300 dark:bg-slate-800 dark:border-slate-700 text-slate-900 dark:text-white text-xs focus:outline-none focus:ring-1 focus:ring-blue-500 font-mono"
                    />
                  </div>

                  <div v-if="hop.authType === 'key'" class="space-y-2">
                    <div>
                      <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">私钥 (路径或 PEM 文本) <span class="text-rose-500">*</span></label>
                      <textarea
                        v-model="hop.privateKey"
                        rows="3"
                        placeholder="~/.ssh/id_rsa 或直接粘贴 -----BEGIN OPENSSH PRIVATE KEY-----"
                        class="w-full px-3 py-1.5 rounded-lg bg-slate-50 border border-slate-300 dark:bg-slate-800 dark:border-slate-700 text-slate-900 dark:text-white text-[11px] font-mono focus:outline-none focus:ring-1 focus:ring-blue-500 resize-y"
                      ></textarea>
                    </div>
                    <div>
                      <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1">私钥密码 (Passphrase，可选)</label>
                      <input
                        v-model="hop.passphrase"
                        type="password"
                        placeholder="若私钥受密码保护请输入"
                        class="w-full px-3 py-1.5 rounded-lg bg-slate-50 border border-slate-300 dark:bg-slate-800 dark:border-slate-700 text-slate-900 dark:text-white text-xs focus:outline-none focus:ring-1 focus:ring-blue-500 font-mono"
                      />
                    </div>
                  </div>
                </template>
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
  Plus,
  Trash2,
  ArrowUp,
  ArrowDown,
  RefreshCw,
  Check,
} from "lucide-vue-next";
import { useConnectionStore } from "../stores/connection";
import { useVectorDBStore } from "../stores/vectordb";
import type { ConnectionConfig, NetworkHop, TestResult, DatabaseDetail } from "../types";

const store = useConnectionStore();
const vdbStore = useVectorDBStore();
const { isModalOpen, editingConnection, closeModal, save, test } = store;

const activeTab = ref<"general" | "network">("general");

const showApiKey = ref(false);
const isTesting = ref(false);
const isSaving = ref(false);
const testResult = ref<TestResult | null>(null);

// Visible Databases state
const selectedDatabases = ref<string[]>([]);
const databaseInput = ref("");
const availableDatabases = ref<DatabaseDetail[]>([]);
const isFetchingDbs = ref(false);
const fetchDbError = ref("");

const form = ref<ConnectionConfig>({
  id: "",
  name: "",
  url: "http://",
  username: "root",
  apiKey: "",
  timeout: 10,
});

const chain = ref<NetworkHop[]>([]);

const isEditing = computed(() => !!editingConnection.value?.id);

const enabledHopsCount = computed(() => {
  return chain.value.filter((h) => h.enabled && h.host.trim()).length;
});

function addHop(type: "http" | "socks5" | "ssh") {
  const newHop: NetworkHop = {
    id: `hop_${Date.now()}_${Math.random().toString(36).slice(2, 7)}`,
    enabled: true,
    type,
    host: "",
    port: type === "ssh" ? 22 : type === "socks5" ? 1080 : 8080,
    username: type === "ssh" ? "root" : "",
    password: "",
    authType: type === "ssh" ? "password" : "",
    privateKey: "",
    passphrase: "",
  };
  chain.value.push(newHop);
}

function removeHop(index: number) {
  chain.value.splice(index, 1);
}

function moveHop(index: number, delta: number) {
  const targetIndex = index + delta;
  if (targetIndex < 0 || targetIndex >= chain.value.length) return;
  const item = chain.value.splice(index, 1)[0];
  chain.value.splice(targetIndex, 0, item);
}

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

      // Populate chain
      if (val.proxyChain && Array.isArray(val.proxyChain) && val.proxyChain.length > 0) {
        chain.value = val.proxyChain.map((h) => ({
          id: h.id || `hop_${Math.random().toString(36).slice(2, 7)}`,
          enabled: h.enabled !== false,
          type: h.type || "http",
          host: h.host || "",
          port: h.port || (h.type === "ssh" ? 22 : h.type === "socks5" ? 1080 : 8080),
          username: h.username || "",
          password: h.password || "",
          authType: h.authType || (h.type === "ssh" ? "password" : ""),
          privateKey: h.privateKey || "",
          passphrase: h.passphrase || "",
        }));
      } else {
        // Auto-migrate from legacy proxy or sshTunnel fields if present
        const migrated: NetworkHop[] = [];
        if (val.proxy?.enabled && val.proxy.host) {
          migrated.push({
            id: `hop_proxy_${Date.now()}`,
            enabled: true,
            type: val.proxy.type || "http",
            host: val.proxy.host,
            port: val.proxy.port || (val.proxy.type === "socks5" ? 1080 : 8080),
            username: val.proxy.username || "",
            password: val.proxy.password || "",
          });
        }
        if (val.sshTunnel?.enabled && val.sshTunnel.host) {
          migrated.push({
            id: `hop_ssh_${Date.now()}`,
            enabled: true,
            type: "ssh",
            host: val.sshTunnel.host,
            port: val.sshTunnel.port || 22,
            username: val.sshTunnel.user || "root",
            authType: val.sshTunnel.authType || "password",
            password: val.sshTunnel.password || "",
            privateKey: val.sshTunnel.privateKey || "",
            passphrase: val.sshTunnel.passphrase || "",
          });
        }
        chain.value = migrated;
      }

      activeTab.value = "general";
      testResult.value = null;
      showApiKey.value = false;

      // Populate visible databases
      selectedDatabases.value = Array.isArray(val.databases)
        ? val.databases.map((d) => d.trim()).filter(Boolean)
        : [];
      databaseInput.value = selectedDatabases.value.join(", ");
      availableDatabases.value = [];
      fetchDbError.value = "";
    }
  },
  { immediate: true }
);

function syncFromDatabaseInput() {
  const parsed = databaseInput.value
    .split(/[,;\s]+/)
    .map((s) => s.trim())
    .filter(Boolean);
  selectedDatabases.value = Array.from(new Set(parsed));
  databaseInput.value = selectedDatabases.value.join(", ");
}

function isDbSelected(name: string): boolean {
  return selectedDatabases.value.includes(name);
}

function toggleDbSelection(name: string) {
  if (isDbSelected(name)) {
    selectedDatabases.value = selectedDatabases.value.filter((d) => d !== name);
  } else {
    selectedDatabases.value.push(name);
  }
  databaseInput.value = selectedDatabases.value.join(", ");
}

function selectAllDbs() {
  selectedDatabases.value = availableDatabases.value.map((d) => d.name);
  databaseInput.value = selectedDatabases.value.join(", ");
}

function clearAllDbs() {
  selectedDatabases.value = [];
  databaseInput.value = "";
}

async function onFetchDatabases() {
  if (!form.value.url || !form.value.apiKey) return;
  syncFromDatabaseInput();
  isFetchingDbs.value = true;
  fetchDbError.value = "";
  try {
    const fullConfig = buildFullConfig();
    const dbs = await store.fetchDatabases(fullConfig);
    availableDatabases.value = dbs || [];
    if (availableDatabases.value.length === 0) {
      fetchDbError.value = "未获取到任何数据库";
    }
  } catch (err: unknown) {
    fetchDbError.value = err instanceof Error ? err.message : String(err);
  } finally {
    isFetchingDbs.value = false;
  }
}

function buildFullConfig(): ConnectionConfig {
  const config: ConnectionConfig = {
    ...form.value,
    databases: selectedDatabases.value.length > 0 ? [...selectedDatabases.value] : [],
    proxyChain: chain.value.map((h) => ({
      ...h,
      host: h.host.trim(),
      username: h.username?.trim() || "",
    })),
  };

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
  syncFromDatabaseInput();
  isSaving.value = true;
  try {
    const fullConfig = buildFullConfig();
    await save(fullConfig);
    if (store.activeConnectionId.value === fullConfig.id) {
      await vdbStore.loadDatabases(fullConfig.id);
    }
  } finally {
    isSaving.value = false;
  }
}
</script>
