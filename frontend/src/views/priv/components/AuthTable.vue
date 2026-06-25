<template>
  <div class="rt-root">
    <div v-if="data.length === 0" class="rt-empty">
      <div class="rt-empty-icon">📋</div>
      <div>{{ emptyText || '暂无数据' }}</div>
    </div>
    <div v-else class="rt-result-list">
      <div class="rt-result-item">
        <div class="rt-result-header" @click="toggleExpand">
          <span class="rt-result-arrow" :class="{ expanded: expanded }">▶</span>
          <span class="rt-result-title">{{ title || '列表' }}</span>
          <span v-if="total !== undefined" class="rt-result-rows">{{ total }} 条记录</span>
          
          <div v-if="showActions && actions.length > 0" class="rt-result-actions">
            <button 
              v-for="action in actions" 
              :key="action.key"
              class="rt-action-btn"
              :class="action.class"
              @click.stop="handleAction(action)"
              :disabled="action.disabled"
              :title="action.title"
            >
              <span class="rt-action-icon">{{ action.icon }}</span>
            </button>
          </div>
        </div>
        
        <div v-if="expanded" class="rt-result-body">
          <div class="rt-table-container">
            <div class="rt-table-wrapper">
              <table class="rt-data-table">
                <thead>
                  <tr>
                    <th v-if="showCheckbox" class="rt-select-col">
                      <input type="checkbox" @change="toggleSelectAll" :checked="isAllSelected" />
                    </th>
                    <th v-for="col in columns" :key="col.key" class="rt-col-header" :style="{ width: col.width }">
                      <div class="rt-col-name">{{ col.label }}</div>
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr 
                    v-for="(row, idx) in data" 
                    :key="idx" 
                    :class="{ odd: idx % 2 === 0 }"
                    @click="handleRowClick(row, idx)"
                  >
                    <td v-if="showCheckbox" class="rt-select-col">
                      <input type="checkbox" :checked="selectedIds.includes(row.id)" @click.stop="toggleSelect(row)" />
                    </td>
                    <td v-for="col in columns" :key="col.key" class="rt-cell" :title="col.key === 'actions' ? '' : formatCell(row, col)" @click="handleCellClick(row, col.key, $event)">
                      <template v-if="col.render">
                        <component :is="col.render" :row="row" :col="col" />
                      </template>
                      <template v-else-if="props.cellRenderer">
                        <div v-html="props.cellRenderer(row, col.key)"></div>
                      </template>
                      <template v-else-if="col.formatter">
                        <div v-html="col.formatter(row, col)"></div>
                      </template>
                      <template v-else>
                        <span :class="{ 'rt-null': isNullValue(row[col.key]) }">{{ formatCell(row, col) }}</span>
                      </template>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
          
          <div v-if="showPagination && total > pageSize" class="rt-status-bar">
            <div class="rt-status-left">
              <span class="rt-status-sql">{{ statusText || '' }}</span>
            </div>
            <div class="rt-status-center">
              <button class="rt-page-btn" @click="firstPage" :disabled="currentPage <= 1" title="首页">
                <svg viewBox="0 0 16 16" width="12" height="12"><path d="M2 4v8h2V4H2zm4 0l4 4-4 4V4z"/></svg>
              </button>
              <button class="rt-page-btn" @click="prevPage" :disabled="currentPage <= 1" title="上一页">
                <svg viewBox="0 0 16 16" width="12" height="12"><path d="M6 4l-4 4 4 4V4z"/></svg>
              </button>
              <span class="rt-status-pagination">第 {{ currentPage }} 页 (共 {{ totalPages }} 页)</span>
              <button class="rt-page-btn" @click="nextPage" :disabled="currentPage >= totalPages" title="下一页">
                <svg viewBox="0 0 16 16" width="12" height="12"><path d="M10 4l4 4-4 4V4z"/></svg>
              </button>
              <button class="rt-page-btn" @click="lastPage" :disabled="currentPage >= totalPages" title="末页">
                <svg viewBox="0 0 16 16" width="12" height="12"><path d="M14 4v8h-2V4h2zm-4 0l-4 4 4 4V4z"/></svg>
              </button>
              <select :value="pageSize" class="rt-page-size" @change="(e: any) => handlePageSizeChange(Number((e.target as HTMLSelectElement).value))">
                <option :value="20">20</option>
                <option :value="50">50</option>
                <option :value="100">100</option>
              </select>
              <span class="rt-page-label">条/页</span>
            </div>
            <div class="rt-status-right">
              <span class="rt-status-records">{{ data.length }} 条记录</span>
              <span class="rt-status-separator"></span>
              <span class="rt-status-updated">上次更新: {{ lastUpdated }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'

interface TableColumn {
  key: string
  label: string
  width?: string
  formatter?: (row: any, col: TableColumn) => string
  render?: any
}

interface TableAction {
  key: string
  icon: string
  title: string
  class?: string
  disabled?: boolean
  handler?: () => void
}

const props = withDefaults(defineProps<{
  data?: any[]
  columns?: TableColumn[]
  title?: string
  emptyText?: string
  showCheckbox?: boolean
  showPagination?: boolean
  total?: number
  currentPage?: number
  pageSize?: number
  statusText?: string
  actions?: TableAction[]
  cellRenderer?: (row: any, colKey: string) => string
}>(), {
  data: () => [],
  columns: () => [],
  title: '列表',
  emptyText: '暂无数据',
  showCheckbox: false,
  showPagination: true,
  total: 0,
  currentPage: 1,
  pageSize: 50,
  actions: () => []
})

const emit = defineEmits<{
  (e: 'page-change', page: number): void
  (e: 'page-size-change', size: number): void
  (e: 'row-click', row: any, idx: number): void
  (e: 'cell-click', row: any, colKey: string, event: Event): void
  (e: 'selection-change', ids: string[]): void
  (e: 'action', action: TableAction): void
}>()

const expanded = ref(true)
const selectedIds = ref<string[]>([])
const lastUpdated = ref(new Date().toLocaleString('zh-CN'))

watch(() => props.data?.length, () => {
  lastUpdated.value = new Date().toLocaleString('zh-CN')
})

const totalPages = computed(() => {
  return Math.max(1, Math.ceil(props.total / props.pageSize))
})

const isAllSelected = computed(() => {
  return props.data.length > 0 && selectedIds.value.length === props.data.length
})

function toggleExpand() {
  expanded.value = !expanded.value
}

function formatCell(row: any, col: TableColumn): string {
  if (row === null || row === undefined) return ''
  if (typeof row === 'object' && col.key in row) {
    const v = row[col.key]
    if (v === null || v === undefined) return ''
    if (typeof v === 'object') return JSON.stringify(v)
    return String(v)
  }
  return ''
}

function isNullValue(val: any): boolean {
  return val === null || val === undefined || val === ''
}

function toggleSelectAll() {
  if (isAllSelected.value) {
    selectedIds.value = []
  } else {
    selectedIds.value = props.data.map((row: any) => row.id)
  }
  emit('selection-change', selectedIds.value)
}

function toggleSelect(row: any) {
  const idx = selectedIds.value.indexOf(row.id)
  if (idx > -1) {
    selectedIds.value.splice(idx, 1)
  } else {
    selectedIds.value.push(row.id)
  }
  emit('selection-change', selectedIds.value)
}

function handleRowClick(row: any, idx: number) {
  emit('row-click', row, idx)
}

function handleCellClick(row: any, colKey: string, event: Event) {
  emit('cell-click', row, colKey, event)
}

function handleAction(action: TableAction) {
  if (action.handler) {
    action.handler()
  }
  emit('action', action)
}

function firstPage() {
  emit('page-change', 1)
}

function prevPage() {
  emit('page-change', Math.max(1, props.currentPage - 1))
}

function nextPage() {
  emit('page-change', Math.min(totalPages.value, props.currentPage + 1))
}

function lastPage() {
  emit('page-change', totalPages.value)
}

function handlePageSizeChange(newSize: number) {
  emit('page-size-change', newSize)
}
</script>

<style scoped>
.rt-root {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.rt-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #6c757d;
}

.rt-empty-icon {
  font-size: 48px;
  margin-bottom: 12px;
}

.rt-result-list {
  flex: 1;
  overflow-y: auto;
}

.rt-result-item {
  border-bottom: 1px solid #e9ecef;
}

.rt-result-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #f8f9fa;
  cursor: pointer;
  user-select: none;
}

.rt-result-header:hover {
  background: #e9ecef;
}

.rt-result-arrow {
  font-size: 10px;
  color: #6c757d;
  transition: transform 0.2s;
}

.rt-result-arrow.expanded {
  transform: rotate(90deg);
}

.rt-result-title {
  font-weight: 500;
  color: #212529;
}

.rt-result-rows {
  font-size: 12px;
  color: #6c757d;
  padding: 2px 6px;
  background: #e9ecef;
  border-radius: 4px;
}

.rt-result-actions {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 4px;
}

.rt-action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s;
}

.rt-action-btn:hover:not(:disabled) {
  background: #e9ecef;
}

.rt-action-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.rt-action-icon {
  font-size: 14px;
}

.rt-action-btn-add .rt-action-icon { color: #28a745; }
.rt-action-btn-delete .rt-action-icon { color: #dc3545; }
.rt-action-btn-edit .rt-action-icon { color: #1976d2; }
.rt-action-btn-refresh .rt-action-icon { color: #1976d2; }
.rt-action-btn-view .rt-action-icon { color: #6c757d; }

.rt-action-separator {
  width: 1px;
  height: 16px;
  background: #dee2e6;
  margin: 0 4px;
}

.rt-result-body {
  background: #ffffff;
}

.rt-table-container {
  max-height: 400px;
  overflow: auto;
}

.rt-table-wrapper {
  overflow-x: auto;
}

.rt-data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.rt-data-table th,
.rt-data-table td {
  padding: 6px 8px;
  text-align: left;
  border-bottom: 1px solid #e9ecef;
}

.rt-data-table th {
  background: #f8f9fa;
  font-weight: 500;
  color: #495057;
  white-space: nowrap;
  position: sticky;
  top: 0;
  z-index: 1;
}

.rt-data-table tbody tr:hover {
  background: #f8f9fa;
}

.rt-data-table tbody tr.odd {
  background: #ffffff;
}

.rt-data-table tbody tr.even {
  background: #fafafa;
}

.rt-select-col {
  width: 30px;
  text-align: center;
}

.rt-col-header {
  display: flex;
  align-items: center;
}

.rt-col-name {
  flex: 1;
}

.rt-cell {
  max-width: 200px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.rt-cell:hover {
  background: #e9ecef;
}

.rt-null {
  color: #adb5bd;
  font-style: italic;
}

.rt-status-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: #f8f9fa;
  border-top: 1px solid #e9ecef;
  font-size: 12px;
}

.rt-status-left {
  flex: 1;
  overflow: hidden;
}

.rt-status-sql {
  font-family: Consolas, Monaco, monospace;
  font-size: 11px;
  color: #6c757d;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.rt-status-center {
  display: flex;
  align-items: center;
  gap: 6px;
}

.rt-page-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 20px;
  background: linear-gradient(180deg, #ffffff 0%, #f8f9fa 100%);
  border: 1px solid #dee2e6;
  border-radius: 2px;
  cursor: pointer;
  transition: all 0.15s;
}

.rt-page-btn:hover:not(:disabled) {
  background: #e3f2fd;
  border-color: #1976d2;
}

.rt-page-btn:active:not(:disabled) {
  background: #bbdefb;
}

.rt-page-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  background: #f8f9fa;
}

.rt-page-size {
  padding: 2px 6px;
  background: #ffffff;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  color: #495057;
  font-size: 12px;
  outline: none;
}

.rt-page-label {
  font-size: 11px;
  color: #6c757d;
}

.rt-status-pagination {
  color: #28a745;
  font-weight: 500;
  font-size: 12px;
}

.rt-status-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.rt-status-records {
  color: #007bff;
  font-size: 12px;
}

.rt-status-separator {
  width: 1px;
  height: 16px;
  background: #dee2e6;
}

.rt-status-updated {
  color: #6c757d;
  font-size: 11px;
}
</style>