<template>
  <div class="ds-node-wrapper">
    <div
      class="ds-node-row"
      :class="{ 'is-selected': isExpanded, 'is-hover': hovered, 'is-connected': isConnected }"
      @click.stop="handleToggle"
      @dblclick.stop="handleDblClick"
      @contextmenu.stop.prevent="handleCtxMenu"
      @mouseenter="hovered = true"
      @mouseleave="hovered = false"
    >
      <span class="arrow" :class="{ open: isExpanded }">
        <svg viewBox="0 0 10 10" width="10" height="10">
          <polygon points="2,1 2,9 9,5" :fill="isConnected ? '#409EFF' : '#c0c4cc'" />
        </svg>
      </span>

      <span class="conn-icon">
        <svg v-if="isLoading" class="rotating" viewBox="0 0 24 24" width="14" height="14">
          <path d="M12 4V1L8 5l4 4V6c3.31 0 6 2.69 6 6 0 1.01-.25 1.97-.7 2.8l1.46 1.46A7.93 7.93 0 0 0 20 12c0-4.42-3.58-8-8-8zm0 14c-3.31 0-6-2.69-6-6 0-1.01.25-1.97.7-2.8L5.24 7.74A7.93 7.93 0 0 0 4 12c0 4.42 3.58 8 8 8v3l4-4-4-4v3z" fill="#409EFF" />
        </svg>
        <svg v-else viewBox="0 0 24 24" width="14" height="14">
          <ellipse cx="12" cy="5" rx="9" ry="2" :fill="iconColor" stroke="none" />
          <path d="M3 5v6c0 1.1 2.2 2 5 2s5-0.9 5-2V5" :stroke="iconColor" stroke-width="1.5" fill="none" />
          <path d="M3 11v6c0 1.1 2.2 2 5 2s5-0.9 5-2v-6" :stroke="iconColor" stroke-width="1.5" fill="none" />
          <ellipse cx="12" cy="11" rx="9" ry="2" :fill="iconLightColor" stroke="none" />
          <ellipse cx="12" cy="17" rx="9" ry="2" :fill="iconColor" stroke="none" />
        </svg>
      </span>

      <span class="conn-name" :title="datasource.name">{{ datasource.name }}</span>

      <span class="conn-type">{{ typeLabel.toUpperCase() }}</span>
    </div>

    <div v-if="isConnected && isExpanded" class="ds-children">
      <TreeItem
        v-for="(node, idx) in filteredChildren"
        :key="'c-' + idx"
        :node="node"
        :level="1"
        :datasource-id="datasource.datasourceId"
        :search-keyword="searchKeyword"
      />
      <div v-if="filteredChildren.length === 0 && treeRoot.length > 0" class="no-match">
        无匹配项
      </div>
      <div v-else-if="treeRoot.length === 0" class="no-match">
        暂无数据库对象
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import TreeItem from './TreeItem.vue'
import type { DatasourceSummary, TreeNode, ObjectType } from '../types'

const props = defineProps<{
  datasource: DatasourceSummary
  searchKeyword?: string
  isConnected: boolean
  isLoading: boolean
  isExpanded: boolean
  treeRoot: TreeNode[]
}>()

const emit = defineEmits<{
  (e: 'toggle'): void
  (e: 'refresh'): void
  (e: 'close-conn'): void
}>()

const hovered = ref(false)

const normalizedType = computed(() => {
  const t = (props.datasource.dbType || 'mysql').toLowerCase()
  const map: Record<string, string> = {
    mysql: 'mysql',
    mariadb: 'mysql',
    tidb: 'tidb',
    sqlite: 'sqlite',
    sqlite3: 'sqlite',
    postgres: 'postgresql',
    postgresql: 'postgresql',
    mongodb: 'mongodb',
    oracle: 'oracle',
    'mssql': 'sqlserver',
    sqlserver: 'sqlserver',
    'sql-server': 'sqlserver'
  }
  return map[t] || 'mysql'
})

const typeLabel = computed(() => {
  const map: Record<string, string> = {
    mysql: 'MySQL',
    tidb: 'TiDB',
    sqlite: 'SQLite',
    postgresql: 'PostgreSQL',
    mongodb: 'MongoDB',
    oracle: 'Oracle',
    sqlserver: 'SQL Server'
  }
  return map[normalizedType.value] || 'MySQL'
})

const iconColor = computed(() => {
  if (props.isLoading) return '#909399'
  if (!props.isConnected) return '#c0c4cc'
  const map: Record<string, string> = {
    mysql: '#00758f',
    tidb: '#e85c2c',
    sqlite: '#003b57',
    postgresql: '#336791',
    mongodb: '#4db33d',
    oracle: '#f80000',
    sqlserver: '#a91d22'
  }
  return map[normalizedType.value] || '#00758f'
})

const iconLightColor = computed(() => {
  const c = iconColor.value
  return c + 'aa'
})

function getSavedQueries(database: string, datasourceId: string): TreeNode[] {
  const key = `sql_saved_queries_${datasourceId || 'default'}_${database || 'default'}`
  const data = localStorage.getItem(key)
  if (!data) return []
  try {
    const queries: any[] = JSON.parse(data)
    return queries.map((q: any) => ({
      type: 'saved-query' as ObjectType,
      name: q.title || '未命名',
      database: database,
      datasourceId: datasourceId,
      sql: q.sql,
      savedQueryId: q.id,
      children: []
    }))
  } catch {
    return []
  }
}

const enhancedTreeRoot = computed<TreeNode[]>(() => {
  const result: TreeNode[] = []
  for (const node of props.treeRoot) {
    if (node.type === 'database' && node.name && props.datasource.datasourceId) {
      const savedQueries = getSavedQueries(node.name, props.datasource.datasourceId)
      const querySaveNode: TreeNode = {
        type: 'query-save',
        name: '查询保存',
        database: node.name,
        datasourceId: props.datasource.datasourceId,
        children: savedQueries
      }
      const children = Array.isArray(node.children) ? [...node.children] : []
      children.push(querySaveNode)
      result.push({ ...node, children })
    } else {
      result.push(node)
    }
  }
  return result
})

const filteredChildren = computed<TreeNode[]>(() => {
  const kw = (props.searchKeyword || '').trim().toLowerCase()
  if (!kw) return enhancedTreeRoot.value
  const filterNodes = (nodes: TreeNode[]): TreeNode[] => {
    const result: TreeNode[] = []
    for (const node of nodes) {
      const nameMatch = (node.name || '').toLowerCase().includes(kw)
      const childMatch = node.children && node.children.length > 0 ? filterNodes(node.children).length > 0 : false
      if (nameMatch || childMatch) {
        const copy: TreeNode = { ...node }
        if (childMatch && node.children) copy.children = filterNodes(node.children)
        result.push(copy)
      }
    }
    return result
  }
  return filterNodes(enhancedTreeRoot.value)
})

function handleToggle() {
  emit('toggle')
}

function handleDblClick() {
  emit('toggle')
}

function handleCtxMenu(e: MouseEvent) {
  const ctxMenu = (window as any).__sqlCtxMenu
  if (ctxMenu && ctxMenu.openDatasourceMenu) {
    ctxMenu.openDatasourceMenu(e, {
      datasourceId: props.datasource.datasourceId,
      name: props.datasource.name,
      dbType: props.datasource.dbType,
      connected: props.isConnected
    })
  }
}
</script>

<style scoped>
.ds-node-wrapper {
  user-select: none;
}

.ds-node-row {
  display: flex;
  align-items: center;
  height: 26px;
  padding: 0 8px;
  cursor: pointer;
  border-radius: 3px;
  transition: background 0.1s;
}

.ds-node-row:hover {
  background: #ecf5ff;
}

.ds-node-row.is-selected {
  background: #d9ecff;
}

.arrow {
  width: 14px;
  height: 14px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-right: 2px;
  flex-shrink: 0;
  transition: transform 0.15s ease;
}

.arrow.open {
  transform: rotate(45deg);
}

.conn-icon {
  width: 16px;
  height: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-right: 6px;
  flex-shrink: 0;
}

.conn-icon .rotating {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.conn-name {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: #303133;
  font-size: 13px;
  font-weight: 500;
}

.conn-type {
  color: #909399;
  font-size: 11px;
  padding: 1px 6px;
  background: #f4f4f5;
  border-radius: 2px;
  margin-left: 8px;
  flex-shrink: 0;
}

.ds-children {
  display: block;
  padding-left: 8px;
}

.no-match {
  padding: 8px 12px;
  color: #c0c4cc;
  font-size: 12px;
  font-style: italic;
}
</style>
