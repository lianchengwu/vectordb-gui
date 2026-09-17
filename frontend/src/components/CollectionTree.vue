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
            @click="onToggleDb($event, db.name)"
            @contextmenu.prevent.stop="onDbContextMenu($event, db)"
            class="group px-2 py-1.5 rounded-md hover:bg-slate-100 dark:hover:bg-slate-800/70 flex items-center justify-between cursor-pointer transition select-none"
          >
            <div
              class="flex items-center gap-1.5 min-w-0 flex-1"
              @contextmenu.prevent.stop="onDbContextMenu($event, db)"
            >
              <ChevronRight
                :class="[
                  'w-3.5 h-3.5 text-slate-400 dark:text-slate-500 transition-transform duration-150',
                  expandedDbs[db.name] ? 'rotate-90 text-slate-700 dark:text-slate-300' : '',
                ]"
              />
              <Database :class="['w-3.5 h-3.5 shrink-0', db.dbType === 'ai' ? 'text-purple-500 dark:text-purple-400' : 'text-blue-500 dark:text-blue-400']" />
              <span
                class="font-medium text-slate-800 dark:text-slate-200 truncate"
                :title="db.name"
                @contextmenu.prevent.stop="onDbContextMenu($event, db)"
              >
                {{ db.name }}
              </span>

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
              <button
                type="button"
                @click.stop="onDbContextMenu($event, db)"
                class="opacity-0 group-hover:opacity-100 p-0.5 rounded text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-200/70 dark:hover:bg-slate-700/60 transition shrink-0 ml-0.5"
                title="数据库操作 (或右键点击)"
              >
                <MoreVertical class="w-3 h-3" />
              </button>
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
              @click="onSelectColl($event, db.name, coll)"
              @contextmenu.prevent.stop="onContextMenu($event, db.name, coll, db.dbType)"
              :class="[
                'px-2 py-1.5 rounded-md flex items-center justify-between gap-1 cursor-pointer transition select-none group',
                activeTab && activeTab.database === db.name && activeTab.collection === coll
                  ? 'bg-blue-50 text-blue-600 border border-blue-200 font-medium dark:bg-blue-600/20 dark:text-blue-300 dark:border-blue-500/30'
                  : 'hover:bg-slate-100 text-slate-600 hover:text-slate-900 dark:hover:bg-slate-800/60 dark:text-slate-400 dark:hover:text-slate-200',
              ]"
            >
              <div class="flex items-center gap-2 min-w-0 flex-1">
                <BookOpen v-if="db.dbType === 'ai'" class="w-3.5 h-3.5 text-purple-500 dark:text-purple-400 shrink-0 group-hover:text-purple-600" />
                <Table v-else class="w-3.5 h-3.5 text-indigo-500 dark:text-indigo-400 shrink-0 group-hover:text-blue-500" />
                <span class="truncate" :title="coll">{{ coll }}</span>
              </div>
              <button
                type="button"
                @click.stop="onContextMenu($event, db.name, coll, db.dbType)"
                class="opacity-0 group-hover:opacity-100 p-0.5 rounded text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-200/70 dark:hover:bg-slate-700/60 transition shrink-0"
                title="更多操作 (或右键点击)"
              >
                <MoreVertical class="w-3 h-3" />
              </button>
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

    <!-- Right-click Context Menu for Collection -->
    <Teleport to="body">
      <div
        v-if="contextMenu.visible"
        :style="{ top: `${contextMenu.y}px`, left: `${contextMenu.x}px` }"
        class="fixed z-[9999] min-w-[160px] bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-lg shadow-xl py-1 text-xs text-slate-700 dark:text-slate-200 select-none"
        @click.stop
        @contextmenu.prevent.stop
      >
        <div class="px-3 py-1.5 border-b border-slate-100 dark:border-slate-800/60 font-mono text-[10px] text-slate-400 truncate max-w-[200px]" :title="contextMenu.collection">
          {{ contextMenu.collection }}
        </div>
        <button
          type="button"
          @click="handleOpenCollection"
          class="w-full px-3 py-1.5 text-left flex items-center gap-2 hover:bg-slate-100 dark:hover:bg-slate-800 transition text-slate-700 dark:text-slate-300"
        >
          <ExternalLink class="w-3.5 h-3.5 text-slate-400" />
          <span>打开集合</span>
        </button>
        <button
          type="button"
          @click="handleCopyName"
          class="w-full px-3 py-1.5 text-left flex items-center gap-2 hover:bg-slate-100 dark:hover:bg-slate-800 transition text-slate-700 dark:text-slate-300"
        >
          <Copy class="w-3.5 h-3.5 text-slate-400" />
          <span>复制名称</span>
        </button>
        <button
          type="button"
          @click="handleCopyFullPath"
          class="w-full px-3 py-1.5 text-left flex items-center gap-2 hover:bg-slate-100 dark:hover:bg-slate-800 transition text-slate-700 dark:text-slate-300"
        >
          <Copy class="w-3.5 h-3.5 text-slate-400" />
          <span>复制完整路径</span>
        </button>
        <div class="my-1 border-t border-slate-100 dark:border-slate-800/60"></div>
        <button
          type="button"
          @click="handlePromptDelete"
          class="w-full px-3 py-1.5 text-left flex items-center gap-2 text-rose-600 dark:text-rose-400 hover:bg-rose-50 dark:hover:bg-rose-950/40 transition font-medium"
        >
          <Trash2 class="w-3.5 h-3.5" />
          <span>删除集合</span>
        </button>
      </div>
    </Teleport>

    <!-- Database Right-click Context Menu -->
    <Teleport to="body">
      <div
        v-if="dbContextMenu.visible"
        :style="{ top: `${dbContextMenu.y}px`, left: `${dbContextMenu.x}px` }"
        class="fixed z-[9999] min-w-[170px] bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-lg shadow-xl py-1 text-xs text-slate-700 dark:text-slate-200 select-none"
        @click.stop
        @contextmenu.prevent.stop
      >
        <!-- Header with DB name and type badge -->
        <div class="px-3 py-1.5 border-b border-slate-100 dark:border-slate-800/60 flex items-center justify-between gap-2 max-w-[220px]">
          <span class="font-mono text-[11px] font-semibold text-slate-800 dark:text-slate-200 truncate" :title="dbContextMenu.db?.name">
            {{ dbContextMenu.db?.name }}
          </span>
          <span
            :class="[
              'px-1 py-0.2 rounded text-[9px] font-mono font-bold shrink-0',
              dbContextMenu.db?.dbType === 'ai'
                ? 'bg-purple-100 text-purple-700 dark:bg-purple-950 dark:text-purple-300'
                : 'bg-blue-100 text-blue-700 dark:bg-blue-950 dark:text-blue-300'
            ]"
          >
            {{ dbContextMenu.db?.dbType === 'ai' ? 'AI' : 'BASE' }}
          </span>
        </div>

        <!-- Refresh Collections -->
        <button
          type="button"
          @click="handleDbRefresh"
          class="w-full px-3 py-1.5 text-left flex items-center gap-2 hover:bg-slate-100 dark:hover:bg-slate-800 transition text-slate-700 dark:text-slate-300"
        >
          <RefreshCw class="w-3.5 h-3.5 text-slate-400" />
          <span>刷新集合列表</span>
        </button>

        <!-- Copy Database Name -->
        <button
          type="button"
          @click="handleDbCopyName"
          class="w-full px-3 py-1.5 text-left flex items-center gap-2 hover:bg-slate-100 dark:hover:bg-slate-800 transition text-slate-700 dark:text-slate-300"
        >
          <Copy class="w-3.5 h-3.5 text-slate-400" />
          <span>复制数据库名称</span>
        </button>

        <div class="my-1 border-t border-slate-100 dark:border-slate-800/60"></div>

        <!-- Toggle Expand / Collapse -->
        <button
          type="button"
          @click="handleDbToggle"
          class="w-full px-3 py-1.5 text-left flex items-center gap-2 hover:bg-slate-100 dark:hover:bg-slate-800 transition text-slate-700 dark:text-slate-300"
        >
          <FolderOpen v-if="expandedDbs[dbContextMenu.db?.name || '']" class="w-3.5 h-3.5 text-slate-400" />
          <Folder v-else class="w-3.5 h-3.5 text-slate-400" />
          <span>{{ expandedDbs[dbContextMenu.db?.name || ''] ? '折叠数据库' : '展开数据库' }}</span>
        </button>

        <!-- Expand All / Collapse All -->
        <button
          type="button"
          @click="handleDbExpandAll"
          class="w-full px-3 py-1.5 text-left flex items-center gap-2 hover:bg-slate-100 dark:hover:bg-slate-800 transition text-slate-700 dark:text-slate-300"
        >
          <Layers class="w-3.5 h-3.5 text-slate-400" />
          <span>全部展开数据库</span>
        </button>
        <button
          type="button"
          @click="handleDbCollapseAll"
          class="w-full px-3 py-1.5 text-left flex items-center gap-2 hover:bg-slate-100 dark:hover:bg-slate-800 transition text-slate-700 dark:text-slate-300"
        >
          <Folder class="w-3.5 h-3.5 text-slate-400" />
          <span>全部折叠数据库</span>
        </button>

        <!-- Close DB Tabs (if any) -->
        <template v-if="openTabsCountForDb(dbContextMenu.db?.name) > 0">
          <div class="my-1 border-t border-slate-100 dark:border-slate-800/60"></div>
          <button
            type="button"
            @click="handleDbCloseTabs"
            class="w-full px-3 py-1.5 text-left flex items-center gap-2 hover:bg-slate-100 dark:hover:bg-slate-800 transition text-slate-700 dark:text-slate-300"
          >
            <X class="w-3.5 h-3.5 text-slate-400" />
            <span>关闭该库所有标签页 ({{ openTabsCountForDb(dbContextMenu.db?.name) }})</span>
          </button>
        </template>

        <div class="my-1 border-t border-slate-100 dark:border-slate-800/60"></div>

        <!-- Database Properties Modal -->
        <button
          type="button"
          @click="handleDbProperties"
          class="w-full px-3 py-1.5 text-left flex items-center gap-2 hover:bg-slate-100 dark:hover:bg-slate-800 transition text-slate-700 dark:text-slate-300"
        >
          <Info class="w-3.5 h-3.5 text-slate-400" />
          <span>数据库属性</span>
        </button>
      </div>
    </Teleport>

    <!-- Database Properties Modal -->
    <Teleport to="body">
      <div
        v-if="dbPropertiesModal.visible"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4 text-xs select-none"
        @click.self="dbPropertiesModal.visible = false"
      >
        <div class="w-full max-w-md rounded-xl bg-white border border-slate-200 shadow-2xl overflow-hidden flex flex-col text-slate-800 dark:bg-slate-900 dark:border-slate-700 dark:text-slate-100 transition-colors duration-200">
          <div class="px-5 py-3.5 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between shrink-0">
            <div class="flex items-center gap-2 font-semibold text-sm text-slate-900 dark:text-white">
              <Database class="w-4 h-4 text-blue-500" />
              <span>数据库属性</span>
            </div>
            <button
              type="button"
              @click="dbPropertiesModal.visible = false"
              class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-1 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition"
            >
              <X class="w-4 h-4" />
            </button>
          </div>

          <div class="p-5 space-y-3 font-mono text-[11px]">
            <div class="p-3.5 rounded-lg bg-slate-50 border border-slate-200 dark:bg-slate-800/60 dark:border-slate-700/60 space-y-2.5">
              <div class="flex justify-between items-center">
                <span class="text-slate-500 dark:text-slate-400 font-sans">数据库名称:</span>
                <span class="font-bold text-slate-900 dark:text-white">{{ dbPropertiesModal.db?.name }}</span>
              </div>
              <div class="flex justify-between items-center">
                <span class="text-slate-500 dark:text-slate-400 font-sans">数据库类型:</span>
                <span
                  :class="[
                    'px-2 py-0.5 rounded text-[10px] font-bold',
                    dbPropertiesModal.db?.dbType === 'ai'
                      ? 'bg-purple-100 text-purple-700 dark:bg-purple-950 dark:text-purple-300'
                      : 'bg-blue-100 text-blue-700 dark:bg-blue-950 dark:text-blue-300'
                  ]"
                >
                  {{ dbPropertiesModal.db?.dbType === 'ai' ? 'AI 知识库 (AI_DB)' : '基础向量库 (BASE)' }}
                </span>
              </div>
              <div class="flex justify-between items-center">
                <span class="text-slate-500 dark:text-slate-400 font-sans">集合总数:</span>
                <span class="text-slate-800 dark:text-slate-200 font-semibold">
                  {{ collectionsByDb[dbPropertiesModal.db?.name || '']?.length ?? 0 }} 个集合
                </span>
              </div>
              <div v-if="dbPropertiesModal.db?.createTime" class="flex justify-between items-center">
                <span class="text-slate-500 dark:text-slate-400 font-sans">创建时间:</span>
                <span class="text-slate-700 dark:text-slate-300">{{ dbPropertiesModal.db.createTime }}</span>
              </div>
            </div>
          </div>

          <div class="px-5 py-3 border-t border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/30 flex items-center justify-end">
            <button
              type="button"
              @click="dbPropertiesModal.visible = false"
              class="px-4 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-700 text-white font-medium transition shadow-sm"
            >
              确定
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Delete Collection Confirmation Modal -->
    <Teleport to="body">
      <div
        v-if="deleteModal.visible"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4 text-xs"
        @click.self="closeDeleteModal"
      >
        <div class="w-full max-w-md rounded-xl bg-white border border-slate-200 shadow-2xl overflow-hidden flex flex-col text-slate-800 dark:bg-slate-900 dark:border-slate-700 dark:text-slate-100 transition-colors duration-200">
          <!-- Modal Header -->
          <div class="px-5 py-3.5 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between shrink-0">
            <div class="flex items-center gap-2 text-rose-600 dark:text-rose-400 font-semibold text-sm">
              <AlertTriangle class="w-4 h-4" />
              <span>删除 Collection 确认</span>
            </div>
            <button
              type="button"
              @click="closeDeleteModal"
              :disabled="deleteModal.loading"
              class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-1 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition disabled:opacity-50"
            >
              <X class="w-4 h-4" />
            </button>
          </div>

          <!-- Modal Body -->
          <div class="p-5 space-y-3.5">
            <p class="text-slate-600 dark:text-slate-300">
              确定要删除集合 <span class="font-semibold font-mono text-rose-600 dark:text-rose-400">{{ deleteModal.collection }}</span> 吗？
            </p>

            <div class="p-3 rounded-lg bg-slate-50 border border-slate-200 dark:bg-slate-800/60 dark:border-slate-700/60 space-y-1.5 font-mono text-[11px]">
              <div class="flex justify-between text-slate-500 dark:text-slate-400">
                <span>所属数据库:</span>
                <span class="text-slate-800 dark:text-slate-200 font-medium">{{ deleteModal.database }}</span>
              </div>
              <div class="flex justify-between text-slate-500 dark:text-slate-400">
                <span>集合名称:</span>
                <span class="text-slate-800 dark:text-slate-200 font-medium">{{ deleteModal.collection }}</span>
              </div>
              <div class="flex justify-between text-slate-500 dark:text-slate-400">
                <span>类型:</span>
                <span class="text-slate-800 dark:text-slate-200 font-medium">{{ deleteModal.dbType === 'ai' ? 'AI 知识库' : 'Base 集合' }}</span>
              </div>
            </div>

            <div class="p-2.5 rounded-md bg-rose-50 border border-rose-200/80 text-rose-700 dark:bg-rose-950/40 dark:border-rose-900/60 dark:text-rose-300 text-[11px] flex items-start gap-2">
              <AlertCircle class="w-3.5 h-3.5 shrink-0 mt-0.5" />
              <span>注意：此操作不可恢复，集合及其包含的所有向量与文档数据将被永久删除！</span>
            </div>

            <!-- Error display if deletion fails -->
            <div
              v-if="deleteModal.error"
              class="p-2.5 rounded-md bg-rose-100 border border-rose-300 text-rose-800 dark:bg-rose-900/40 dark:border-rose-800 dark:text-rose-200 text-[11px] break-all"
            >
              删除失败: {{ deleteModal.error }}
            </div>
          </div>

          <!-- Modal Footer -->
          <div class="px-5 py-3 border-t border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/30 flex items-center justify-end gap-2.5 shrink-0">
            <button
              type="button"
              @click="closeDeleteModal"
              :disabled="deleteModal.loading"
              class="px-3 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-300 transition disabled:opacity-50"
            >
              取消
            </button>
            <button
              type="button"
              @click="confirmDelete"
              :disabled="deleteModal.loading"
              class="px-3.5 py-1.5 rounded-lg bg-rose-600 hover:bg-rose-700 text-white font-medium transition flex items-center gap-1.5 disabled:opacity-50 shadow-sm"
            >
              <Loader2 v-if="deleteModal.loading" class="w-3.5 h-3.5 animate-spin" />
              <Trash2 v-else class="w-3.5 h-3.5" />
              <span>确认删除</span>
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Feedback Toast -->
    <Teleport to="body">
      <div
        v-if="toastMessage"
        class="fixed bottom-6 right-6 z-50 px-4 py-2.5 rounded-lg shadow-xl bg-slate-900 text-white dark:bg-slate-100 dark:text-slate-900 text-xs flex items-center gap-2 select-none"
      >
        <CheckCircle2 class="w-4 h-4 text-emerald-400 dark:text-emerald-600 shrink-0" />
        <span>{{ toastMessage }}</span>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
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
  Trash2,
  AlertTriangle,
  CheckCircle2,
  ExternalLink,
  Copy,
  MoreVertical,
  FolderOpen,
  Folder,
  Layers,
  Info,
} from "lucide-vue-next";
import type { DatabaseDetail } from "../types";
import { useVectorDBStore } from "../stores/vectordb";
import { useConnectionStore } from "../stores/connection";

const vdbStore = useVectorDBStore();
const connStore = useConnectionStore();

const {
  databases,
  collectionsByDb,
  expandedDbs,
  openTabs,
  activeTab,
  loading,
  loadingCollections,
  error,
  loadDatabases,
  toggleDatabase,
  refreshDatabase,
  expandAllDatabases,
  collapseAllDatabases,
  closeTabsByDatabase,
  openCollectionTab,
} = vdbStore;
const { activeConnectionId } = connStore;

const searchQuery = ref("");

interface ContextMenuState {
  visible: boolean;
  x: number;
  y: number;
  database: string;
  collection: string;
  dbType: string;
}

interface DeleteModalState {
  visible: boolean;
  loading: boolean;
  error: string;
  database: string;
  collection: string;
  dbType: string;
}

interface DbContextMenuState {
  visible: boolean;
  x: number;
  y: number;
  db: DatabaseDetail | null;
}

interface DbPropertiesModalState {
  visible: boolean;
  db: DatabaseDetail | null;
}

const contextMenu = ref<ContextMenuState>({
  visible: false,
  x: 0,
  y: 0,
  database: "",
  collection: "",
  dbType: "base",
});

const deleteModal = ref<DeleteModalState>({
  visible: false,
  loading: false,
  error: "",
  database: "",
  collection: "",
  dbType: "base",
});

const dbContextMenu = ref<DbContextMenuState>({
  visible: false,
  x: 0,
  y: 0,
  db: null,
});

const dbPropertiesModal = ref<DbPropertiesModalState>({
  visible: false,
  db: null,
});

const toastMessage = ref("");
let toastTimeout: ReturnType<typeof setTimeout> | null = null;

function showToast(msg: string) {
  toastMessage.value = msg;
  if (toastTimeout) clearTimeout(toastTimeout);
  toastTimeout = setTimeout(() => {
    toastMessage.value = "";
  }, 2500);
}

let lastMenuOpenTime = 0;

function onContextMenu(e: MouseEvent, dbName: string, coll: string, dbType?: string) {
  e.preventDefault();
  e.stopPropagation();
  closeDbContextMenu();
  lastMenuOpenTime = Date.now();
  const menuWidth = 160;
  const menuHeight = 165;
  const clientX = e.clientX || (e.target as HTMLElement)?.getBoundingClientRect?.().left || 100;
  const clientY = e.clientY || (e.target as HTMLElement)?.getBoundingClientRect?.().bottom || 100;
  const x = Math.min(clientX, window.innerWidth - menuWidth - 8);
  const y = Math.min(clientY, window.innerHeight - menuHeight - 8);

  contextMenu.value = {
    visible: true,
    x: Math.max(8, x),
    y: Math.max(8, y),
    database: dbName,
    collection: coll,
    dbType: dbType || "base",
  };
}

function closeContextMenu() {
  contextMenu.value.visible = false;
}

function onDbContextMenu(e: MouseEvent, db: DatabaseDetail) {
  e.preventDefault();
  e.stopPropagation();
  closeContextMenu();
  lastMenuOpenTime = Date.now();
  const menuWidth = 180;
  const menuHeight = 240;
  const clientX = e.clientX || (e.target as HTMLElement)?.getBoundingClientRect?.().left || 100;
  const clientY = e.clientY || (e.target as HTMLElement)?.getBoundingClientRect?.().bottom || 100;
  const x = Math.min(clientX, window.innerWidth - menuWidth - 8);
  const y = Math.min(clientY, window.innerHeight - menuHeight - 8);

  dbContextMenu.value = {
    visible: true,
    x: Math.max(8, x),
    y: Math.max(8, y),
    db,
  };
}

function closeDbContextMenu() {
  dbContextMenu.value.visible = false;
}

function closeAllMenus(e?: MouseEvent) {
  // If the menu was just opened in this event cycle (<200ms), ignore the click event
  if (Date.now() - lastMenuOpenTime < 200) return;
  // If it's a non-primary mouse click (e.g. right click), don't close
  if (e && e.button !== 0) return;
  closeContextMenu();
  closeDbContextMenu();
}

async function handleDbRefresh() {
  const db = dbContextMenu.value.db;
  closeDbContextMenu();
  if (activeConnectionId.value && db) {
    await refreshDatabase(activeConnectionId.value, db.name, db.dbType);
    showToast(`已刷新数据库 "${db.name}"`);
  }
}

async function handleDbCopyName() {
  const db = dbContextMenu.value.db;
  closeDbContextMenu();
  if (db?.name) {
    try {
      await navigator.clipboard.writeText(db.name);
      showToast(`已复制数据库名称: ${db.name}`);
    } catch {
      // ignore
    }
  }
}

function handleDbToggle() {
  const db = dbContextMenu.value.db;
  closeDbContextMenu();
  if (activeConnectionId.value && db) {
    toggleDatabase(activeConnectionId.value, db.name);
  }
}

async function handleDbExpandAll() {
  closeDbContextMenu();
  if (activeConnectionId.value) {
    await expandAllDatabases(activeConnectionId.value);
    showToast("已展开全部数据库");
  }
}

function handleDbCollapseAll() {
  closeDbContextMenu();
  collapseAllDatabases();
  showToast("已折叠全部数据库");
}

function openTabsCountForDb(dbName?: string): number {
  if (!dbName) return 0;
  return openTabs.value.filter((t) => t.database === dbName).length;
}

function handleDbCloseTabs() {
  const db = dbContextMenu.value.db;
  closeDbContextMenu();
  if (db?.name) {
    closeTabsByDatabase(db.name);
    showToast(`已关闭数据库 "${db.name}" 的打开标签页`);
  }
}

function handleDbProperties() {
  const db = dbContextMenu.value.db;
  closeDbContextMenu();
  if (db) {
    dbPropertiesModal.value = {
      visible: true,
      db,
    };
  }
}

function handleOpenCollection() {
  if (activeConnectionId.value && contextMenu.value.database && contextMenu.value.collection) {
    openCollectionTab(activeConnectionId.value, contextMenu.value.database, contextMenu.value.collection);
  }
  closeContextMenu();
}

async function handleCopyName() {
  if (contextMenu.value.collection) {
    try {
      await navigator.clipboard.writeText(contextMenu.value.collection);
      showToast(`已复制: ${contextMenu.value.collection}`);
    } catch {
      // ignore
    }
  }
  closeContextMenu();
}

async function handleCopyFullPath() {
  if (contextMenu.value.database && contextMenu.value.collection) {
    const fullPath = `${contextMenu.value.database}.${contextMenu.value.collection}`;
    try {
      await navigator.clipboard.writeText(fullPath);
      showToast(`已复制完整路径: ${fullPath}`);
    } catch {
      // ignore
    }
  }
  closeContextMenu();
}

function handlePromptDelete() {
  deleteModal.value = {
    visible: true,
    loading: false,
    error: "",
    database: contextMenu.value.database,
    collection: contextMenu.value.collection,
    dbType: contextMenu.value.dbType,
  };
  closeContextMenu();
}

function closeDeleteModal() {
  if (deleteModal.value.loading) return;
  deleteModal.value.visible = false;
  deleteModal.value.error = "";
}

async function confirmDelete() {
  if (!activeConnectionId.value || !deleteModal.value.database || !deleteModal.value.collection) return;
  deleteModal.value.loading = true;
  deleteModal.value.error = "";
  try {
    await vdbStore.deleteCollection(
      activeConnectionId.value,
      deleteModal.value.database,
      deleteModal.value.collection,
      deleteModal.value.dbType
    );
    const deletedName = deleteModal.value.collection;
    deleteModal.value.visible = false;
    deleteModal.value.error = "";
    showToast(`集合 "${deletedName}" 已成功删除`);
  } catch (err: unknown) {
    deleteModal.value.error = err instanceof Error ? err.message : String(err);
  } finally {
    deleteModal.value.loading = false;
  }
}

function handleKeyDown(e: KeyboardEvent) {
  if (e.key === "Escape") {
    if (contextMenu.value.visible) {
      closeContextMenu();
    } else if (dbContextMenu.value.visible) {
      closeDbContextMenu();
    } else if (deleteModal.value.visible && !deleteModal.value.loading) {
      closeDeleteModal();
    } else if (dbPropertiesModal.value.visible) {
      dbPropertiesModal.value.visible = false;
    }
  }
}

onMounted(() => {
  window.addEventListener("click", closeAllMenus);
  window.addEventListener("keydown", handleKeyDown);
});

onUnmounted(() => {
  window.removeEventListener("click", closeAllMenus);
  window.removeEventListener("keydown", handleKeyDown);
  if (toastTimeout) clearTimeout(toastTimeout);
});

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

function onToggleDb(e: MouseEvent, dbName: string) {
  if (e && e.button !== 0) return;
  if (activeConnectionId.value) {
    toggleDatabase(activeConnectionId.value, dbName);
  }
}

function onSelectColl(e: MouseEvent, dbName: string, coll: string) {
  if (e && e.button !== 0) return;
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
