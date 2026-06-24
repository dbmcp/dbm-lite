<template>
  <div class="object-tree-panel">
    <div class="object-tree-search">
      <svg class="search-icon" viewBox="0 0 24 24" width="14" height="14">
        <path d="M15.5 14h-.79l-.28-.27A6.471 6.471 0 0 0 16 9.5 6.5 6.5 0 1 0 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z" fill="#909399"/>
      </svg>
      <input
        v-model="keyword"
        type="text"
        placeholder="搜索数据库/表/视图..."
        class="search-input"
        @keyup.esc="keyword = ''"
      />
      <button v-if="keyword" class="search-clear" @click="keyword = ''">
        <svg viewBox="0 0 24 24" width="12" height="12">
          <path d="M18.3 5.71L12 12.01l-6.3-6.3-1.41 1.41L10.59 13.4 4.29 19.71l1.41 1.42 6.3-6.3 6.3 6.3 1.41-1.41-6.3-6.3 6.3-6.31-1.41-1.41z" fill="#909399"/>
        </svg>
      </button>
    </div>

    <div class="object-tree-toolbar">
      <button class="tree-btn" @click="refreshActive">
        <svg viewBox="0 0 24 24" width="13" height="13">
          <path d="M17.65 6.35A7.958 7.958 0 0 0 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08A5.99 5.99 0 0 1 12 18c-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z" fill="#606266"/>
        </svg>
        <span>刷新</span>
      </button>
      <button class="tree-btn" @click="collapseAll">
        <svg viewBox="0 0 24 24" width="13" height="13">
          <path d="M7 10h10v1H7zM7 6h10v1H7zM7 14h10v1H7zM7 18h10v1H7z" fill="#606266"/>
        </svg>
        <span>折叠</span>
      </button>
      <span class="tree-count">{{ datasourceList.length }} 个连接</span>
    </div>

    <div class="object-tree-body" ref="treeBodyRef">
      <div v-if="datasourceList.length === 0" class="tree-empty">
        <svg viewBox="0 0 24 24" width="24" height="24">
          <path d="M12 2C8 2 5 4 5 7v3c0 3 3 5 7 5s7-2 7-5V7c0-3-3-5-7-5zM7 7c0-1.7 2.3-3 5-3s5 1.3 5 3c0 1.7-2.3 3-5 3s-5-1.3-5-3zm0 6v3c0 1.7 2.3 3 5 3s5-1.3 5-3v-3c-1.5 1.3-3.3 2-5 2s-3.5-.7-5-2z" fill="#909399"/>
        </svg>
        <span>暂无数据源</span>
      </div>
      <div v-else class="tree-content">
        <DatasourceNode
          v-for="ds in filteredDatasources"
          :key="ds.datasourceId"
          :datasource="ds"
          :search-keyword="keyword"
          :is-connected="isConnected(ds.datasourceId)"
          :is-loading="isLoading(ds.datasourceId)"
          :is-expanded="isExpanded(ds.datasourceId)"
          :tree-root="treeFor(ds.datasourceId)"
          @toggle="handleToggleConnection(ds.datasourceId)"
          @refresh="handleRefreshConnection(ds.datasourceId)"
          @close-conn="handleCloseConnection(ds.datasourceId)"
        />
      </div>
    </div>

    <div class="object-tree-footer">
      <span>{{ datasourceList.length }} 个连接</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import type { DatasourceSummary } from '../types'
import { useSqlIdeDatasource } from '../hooks/useDatasource'
import DatasourceNode from './DatasourceNode.vue'

const keyword = ref<string>('')
const treeBodyRef = ref<HTMLElement | null>(null)

const {
  state,
  loadDatasources,
  openConnection,
  refreshConnection,
  closeConnection,
  toggleExpand,
  isConnected,
  isLoading,
  isExpanded,
  treeFor,
  normalizeDbType
} = useSqlIdeDatasource()

const datasourceList = computed<DatasourceSummary[]>(() => state.list || [])

const filteredDatasources = computed<DatasourceSummary[]>(() => {
  const kw = keyword.value.trim().toLowerCase()
  if (!kw) return datasourceList.value
  return datasourceList.value.filter((ds) => {
    if (ds.name.toLowerCase().includes(kw)) return true
    const tree = treeFor(ds.datasourceId)
    if (tree.length === 0) return false
    const matchNode = (nodes: any[]): boolean => {
      for (const n of nodes) {
        if ((n.name || '').toLowerCase().includes(kw)) return true
        if (n.children && n.children.length > 0) {
          if (matchNode(n.children)) return true
        }
      }
      return false
    }
    return matchNode(tree)
  })
})

async function handleToggleConnection(datasourceId: string) {
  if (isConnected(datasourceId)) {
    toggleExpand(datasourceId)
  } else {
    await openConnection(datasourceId)
  }
}

async function handleRefreshConnection(datasourceId: string) {
  if (isConnected(datasourceId)) {
    await refreshConnection(datasourceId)
  } else {
    await openConnection(datasourceId)
  }
}

function handleCloseConnection(datasourceId: string) {
  closeConnection(datasourceId)
}

function collapseAll() {
  for (const ds of datasourceList.value) {
    if (isExpanded(ds.datasourceId)) {
      toggleExpand(ds.datasourceId)
    }
  }
}

function refreshActive() {
  for (const ds of datasourceList.value) {
    if (isExpanded(ds.datasourceId) || isConnected(ds.datasourceId)) {
      refreshConnection(ds.datasourceId)
    }
  }
}

onMounted(async () => {
  if (!state.list || state.list.length === 0) {
    await loadDatasources()
  }
})
</script>

<style scoped>
.object-tree-panel {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  background: #ffffff;
  font-size: 13px;
}

.object-tree-search {
  position: relative;
  padding: 8px 12px;
  border-bottom: 1px solid #ebeef5;
  background: #fafbfc;
}

.search-icon {
  position: absolute;
  left: 22px;
  top: 50%;
  transform: translateY(-50%);
  pointer-events: none;
  opacity: 0.7;
}

.search-input {
  width: 100%;
  padding: 6px 30px 6px 32px;
  border: 1px solid #dcdfe6;
  border-radius: 3px;
  font-size: 13px;
  outline: none;
  background: #fff;
  box-sizing: border-box;
  transition: border-color 0.15s;
}

.search-input:focus {
  border-color: #409eff;
}

.search-clear {
  position: absolute;
  right: 20px;
  top: 50%;
  transform: translateY(-50%);
  width: 18px;
  height: 18px;
  border: none;
  background: transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 9px;
  transition: background 0.15s;
}

.search-clear:hover {
  background: #eef3fa;
}

.object-tree-toolbar {
  padding: 6px 10px;
  display: flex;
  align-items: center;
  gap: 4px;
  border-bottom: 1px solid #ebeef5;
  background: #ffffff;
}

.tree-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border: 1px solid transparent;
  background: transparent;
  color: #606266;
  font-size: 12px;
  border-radius: 3px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.tree-btn:hover {
  background: #eef5ff;
  color: #409eff;
  border-color: #d9ecff;
}

.tree-count {
  margin-left: auto;
  font-size: 11px;
  color: #909399;
}

.object-tree-body {
  flex: 1 1 auto;
  overflow: auto;
  padding: 2px 0 8px;
  background: #ffffff;
}

.tree-content {
  width: 100%;
}

.tree-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 24px 8px;
  color: #909399;
  font-size: 13px;
  flex-direction: column;
}

.object-tree-footer {
  border-top: 1px solid #ebeef5;
  padding: 6px 12px;
  font-size: 11px;
  color: #909399;
  background: #f8f9fb;
  text-align: right;
}
</style>
