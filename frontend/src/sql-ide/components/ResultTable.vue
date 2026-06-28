<template>
  <div class="rt-root">
    <div v-if="data.length === 0" class="rt-empty">
      <div class="rt-empty-icon">📋</div>
      <div>{{ emptyText || '暂无数据' }}</div>
    </div>
    <div v-else class="rt-result-list">
      <div v-for="(r, idx) in data" :key="'r-' + idx" class="rt-result-item">
        <div class="rt-result-header" @click="showExpandArrow && toggleResult(idx)">
          <span v-if="showExpandArrow" class="rt-result-arrow" :class="{ expanded: expandedResults[idx] !== false }">▶</span>
          <span class="rt-result-title">{{ r.title || (title || '结果') + ' ' + (idx + 1) }}</span>
          <span v-if="r.affectedRows !== undefined && r.affectedRows !== null && showAffectedRows" class="rt-result-affected">影响 {{ r.affectedRows }} 行</span>
          <span v-if="safeRows(r).length > 0 && showRowCount" class="rt-result-rows">{{ safeRows(r).length }} 行</span>
          <span v-if="r.durationMs && showDuration" class="rt-result-dura">{{ r.durationMs }} ms</span>
          <span v-if="showStatus" class="rt-result-status" :class="r.success !== false ? 'ok' : 'err'">{{ r.success !== false ? '成功' : '失败' }}</span>
          
          <div v-if="showActions" class="rt-result-actions">
            <button v-if="showAddButton" class="rt-action-btn rt-action-btn-add" @click.stop="addNewRow(r, idx)" :disabled="!canEdit(r)" title="添加记录">
              <span class="rt-action-icon">+</span>
            </button>
            <button v-if="showDeleteButton" class="rt-action-btn rt-action-btn-delete" @click.stop="deleteSelectedRows(r, idx)" :disabled="!hasSelectedRows(r, idx) || isSubmitting[idx]" title="删除记录">
              <span class="rt-action-icon">-</span>
            </button>
            <span v-if="showAddButton || showDeleteButton" class="rt-action-separator"></span>
            <button v-if="showSubmitButton" class="rt-action-btn rt-action-btn-submit" @click.stop="submitChanges(r, idx)" :disabled="!hasChanges(r, idx) || isSubmitting[idx]" title="提交">
              <span class="rt-action-icon">✓</span>
            </button>
            <button v-if="showRollbackButton" class="rt-action-btn rt-action-btn-rollback" @click.stop="rollbackChanges(r, idx)" :disabled="!hasChanges(r, idx) || isSubmitting[idx]" title="放弃更改">
              <span class="rt-action-icon">✕</span>
            </button>
            <span v-if="(showSubmitButton || showRollbackButton) && (showRefreshButton || showExportButton || showStopButton || showColumnPicker)" class="rt-action-separator"></span>
            <button v-if="showRefreshButton" class="rt-action-btn rt-action-btn-refresh" @click.stop="refreshResult(r, idx)" :disabled="isSubmitting[idx]" title="刷新">
              <span class="rt-action-icon">⟳</span>
            </button>
            <button v-if="showExportButton" class="rt-action-btn rt-action-btn-export" @click.stop="toggleExportMenu(r.id)" title="导出">
              <span class="rt-action-icon">⤤</span>
            </button>
            <button v-if="showStopButton" class="rt-action-btn rt-action-btn-stop" @click.stop="handleStop(r, idx)" :disabled="!isSubmitting[idx]" title="停止">
              <span class="rt-action-icon">■</span>
            </button>
            <button v-if="showColumnPicker && safeColumns(r).length > 0" class="rt-action-btn rt-action-btn-columns" @click.stop="toggleColumnPicker(r.id)" title="列选择">
              <span class="rt-action-icon">☰</span>
            </button>
          </div>
          
          <div v-if="showExportButton" class="rt-export-menu" v-show="exportMenuOpen === r.id">
            <div class="rt-export-item" @click="exportData(r, idx, 'csv', false)">导出全表 (CSV)</div>
            <div class="rt-export-item" @click="exportData(r, idx, 'csv', true)">导出选中行 (CSV)</div>
            <div class="rt-export-item" @click="exportData(r, idx, 'sql', false)">导出全表 (SQL)</div>
            <div class="rt-export-item" @click="exportData(r, idx, 'sql', true)">导出选中行 (SQL)</div>
          </div>
          
          <div v-if="showColumnPicker" class="rt-column-menu" v-show="columnPickerOpen === r.id">
            <div 
              v-for="(col, ci) in safeColumns(r)" 
              :key="'col-' + ci" 
              class="rt-column-item"
              @click="toggleColumn(r.id, col)"
            >
              <span :class="{ 'checked': isColumnVisible(r.id, col) }">{{ isColumnVisible(r.id, col) ? '✓' : '○' }}</span>
              <span>{{ props.columnLabels?.[col] || col }}</span>
            </div>
          </div>
          
          <span v-if="showCloseButton" class="rt-result-close" @click.stop="emit('close-result', idx)">✕</span>
        </div>
        
        <div v-if="!showExpandArrow || expandedResults[idx] !== false" class="rt-result-body">
          <div v-if="r.success === false" class="rt-error-box">
            <span class="rt-error-title">错误信息</span>
            <div class="rt-error-text">{{ r.message || r.error || '未知错误' }}</div>
          </div>
          <template v-else-if="safeRows(r).length > 0">
            <div v-show="viewMode === 'table'" class="rt-table-container">
              <div class="rt-table-wrapper">
                <table class="rt-data-table">
                  <thead>
                    <tr>
                      <th v-if="showCheckbox" class="rt-select-col">
                        <input type="checkbox" @change="toggleSelectAll(r, idx)" :checked="isAllSelected(r, idx)" />
                      </th>
                      <th class="rt-row-num-col">#</th>
                      <th v-for="(col, ci) in visibleColumns(r)" :key="'h-' + ci" class="rt-col-header">
                        <div class="rt-col-name">{{ props.columnLabels?.[col] || col }}</div>
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(row, ri) in paginatedRows(r, idx)" :key="'row-' + ri" :class="{ odd: ri % 2 === 0, editing: editingCell[r.id + '-' + ri], modified: isRowModified(r.id, idx, ri) }">
                      <td v-if="showCheckbox" class="rt-select-col">
                        <input type="checkbox" v-model="selectedRows[r.id + '-' + ri]" />
                      </td>
                      <td class="rt-row-num-col">{{ ((currentPages[idx] || 1) - 1) * (pageSizes[idx] || 50) + ri + 1 }}</td>
                      <td 
                        v-for="(col, ci) in visibleColumns(r)" 
                        :key="'c-' + ci" 
                        :title="formatCell(row, col)"
                        @dblclick="showEdit && startEdit(r, idx, ri, col)"
                        @click="emit('cell-click', row, col, $event)"
                        class="rt-cell"
                      >
                        <template v-if="editingCell[r.id + '-' + ri] === col">
                          <input 
                            type="text" 
                            v-model="editValue" 
                            @blur="endEdit(r, idx, ri, col)"
                            @keyup.enter="endEdit(r, idx, ri, col)"
                            @keyup.escape="cancelEdit(r, idx, ri)"
                            class="rt-edit-input"
                          />
                        </template>
                        <template v-else-if="props.cellRenderer">
                          <div v-html="props.cellRenderer(row, col)"></div>
                        </template>
                        <template v-else>
                          <span :class="{ 'rt-null': isNullValue(row[col]) }">{{ formatCell(row, col) }}</span>
                        </template>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
            
            <div v-if="showShellView" v-show="viewMode === 'shell'" class="rt-shell-wrap">
            <pre class="rt-shell-output">{{ formatShellOutput(r) }}</pre>
          </div>
            
            <div v-if="showPagination" class="rt-status-bar">
              <div class="rt-status-left" v-if="showSql">
                <span class="rt-status-sql">{{ r.sql || 'No SQL' }}</span>
              </div>
              <div class="rt-status-center">
                <button class="rt-page-btn" @click="firstPage(idx)" :disabled="(currentPages[idx] || 1) <= 1" title="首页">
                  <svg viewBox="0 0 16 16" width="12" height="12"><path d="M2 4v8h2V4H2zm4 0l4 4-4 4V4z"/></svg>
                </button>
                <button class="rt-page-btn" @click="prevPage(idx)" :disabled="(currentPages[idx] || 1) <= 1" title="上一页">
                  <svg viewBox="0 0 16 16" width="12" height="12"><path d="M6 4l-4 4 4 4V4z"/></svg>
                </button>
                <span class="rt-status-pagination">第 {{ currentPages[idx] || 1 }} 页 (共 {{ totalPages(r, idx) }} 页)</span>
                <button class="rt-page-btn" @click="nextPage(r, idx)" :disabled="(currentPages[idx] || 1) >= totalPages(r, idx)" title="下一页">
                  <svg viewBox="0 0 16 16" width="12" height="12"><path d="M10 4l4 4-4 4V4z"/></svg>
                </button>
                <button class="rt-page-btn" @click="lastPage(r, idx)" :disabled="(currentPages[idx] || 1) >= totalPages(r, idx)" title="末页">
                  <svg viewBox="0 0 16 16" width="12" height="12"><path d="M14 4v8h-2V4h2zm-4 0l-4 4 4 4V4z"/></svg>
                </button>
                <select v-model="pageSizes[idx]" class="rt-page-size">
                  <option :value="20">20</option>
                  <option :value="50">50</option>
                  <option :value="100">100</option>
                  <option :value="500">500</option>
                  <option :value="1000">1000</option>
                </select>
                <span class="rt-page-label">条/页</span>
              </div>
              <div class="rt-status-right">
                <span class="rt-status-records">{{ safeRows(r).length }} 条记录</span>
                <span class="rt-status-separator"></span>
                <span class="rt-status-updated">上次更新: {{ lastUpdated[idx] || '-' }}</span>
              </div>
            </div>
          </template>
          <div v-else-if="r.message" class="rt-message-box">
            <pre>{{ r.message }}</pre>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue'

interface ResultData {
  id?: string
  title?: string
  rows?: any[]
  data?: any[]
  success?: boolean
  message?: string
  error?: string
  affectedRows?: number
  durationMs?: number
  sql?: string
}

const props = withDefaults(defineProps<{
  data?: ResultData[]
  title?: string
  emptyText?: string
  showActions?: boolean
  showAddButton?: boolean
  showDeleteButton?: boolean
  showSubmitButton?: boolean
  showRollbackButton?: boolean
  showRefreshButton?: boolean
  showExportButton?: boolean
  showStopButton?: boolean
  showCloseButton?: boolean
  showColumnPicker?: boolean
  showCheckbox?: boolean
  showEdit?: boolean
  showPagination?: boolean
  showAffectedRows?: boolean
  showRowCount?: boolean
  showDuration?: boolean
  showStatus?: boolean
  showExpandArrow?: boolean
  showShellView?: boolean
  showSql?: boolean
  cellRenderer?: (row: any, colKey: string) => string
  columnLabels?: Record<string, string>
  expandedByDefault?: boolean
}>(), {
  data: () => [],
  title: '结果',
  emptyText: '暂无数据',
  showActions: true,
  showAddButton: true,
  showDeleteButton: true,
  showSubmitButton: true,
  showRollbackButton: true,
  showRefreshButton: true,
  showExportButton: true,
  showStopButton: true,
  showCloseButton: true,
  showColumnPicker: true,
  showCheckbox: true,
  showEdit: true,
  showPagination: true,
  showAffectedRows: true,
  showRowCount: true,
  showDuration: true,
  showStatus: true,
  showExpandArrow: true,
  showShellView: true,
  showSql: true
})

const emit = defineEmits<{
  (e: 'close-result', idx: number): void
  (e: 'refresh-single', rId: string, idx: number, sql: string): void
  (e: 'cell-click', row: any, colKey: string, event: Event): void
}>()

const viewMode = ref<'table' | 'shell'>('table')
const currentPages = reactive<Record<number, number>>({})
const pageSizes = reactive<Record<number, number>>({})
const expandedResults = reactive<Record<number, boolean>>({})
const lastUpdated = reactive<Record<number, string>>({})
const isSubmitting = reactive<Record<number, boolean>>({})
const exportMenuOpen = ref<string | null>(null)
const columnPickerOpen = ref<string | null>(null)
const editingCell = reactive<Record<string, string>>({})
const selectedRows = reactive<Record<string, boolean>>({})
const editValue = ref('')
const modifiedRows = reactive<Record<string, any>>({})
const deletedRows = reactive<Record<string, any[]>>({})
const originalRows = reactive<Record<string, any[]>>({})
const hiddenColumns = reactive<Record<string, Set<string>>>({})

watch(() => props.data, () => {
  for (let idx = 0; idx < (props.data?.length || 0); idx++) {
    expandedResults[idx] = props.expandedByDefault !== false
    if (currentPages[idx] === undefined) currentPages[idx] = 1
    if (pageSizes[idx] === undefined) pageSizes[idx] = 50
    lastUpdated[idx] = new Date().toLocaleString('zh-CN')
  }
}, { deep: true })

onMounted(() => {
  for (let idx = 0; idx < (props.data?.length || 0); idx++) {
    expandedResults[idx] = props.expandedByDefault !== false
  }
})

function safeRows(r: any): any[] {
  if (r === null || r === undefined) return []
  return Array.isArray(r.rows) ? r.rows : (Array.isArray(r.data) ? r.data : [])
}

function safeColumns(r: any): string[] {
  const rows = safeRows(r)
  if (rows.length === 0) return []
  
  const dataColumns = Object.keys(rows[0])
  
  if (props.columnLabels) {
    const orderedColumns: string[] = []
    const labelKeys = Object.keys(props.columnLabels)
    
    for (const key of labelKeys) {
      if (dataColumns.includes(key)) {
        orderedColumns.push(key)
      }
    }
    
    for (const col of dataColumns) {
      if (!orderedColumns.includes(col)) {
        orderedColumns.push(col)
      }
    }
    
    return orderedColumns
  }
  
  return dataColumns
}

function visibleColumns(r: any): string[] {
  const columns = safeColumns(r)
  const resultId = r.id || 'default'
  if (!hiddenColumns[resultId]) {
    hiddenColumns[resultId] = new Set()
  }
  return columns.filter(col => !hiddenColumns[resultId].has(col))
}

function toggleColumnPicker(id: string | undefined) {
  if (!id) return
  columnPickerOpen.value = columnPickerOpen.value === id ? null : id
}

function isColumnVisible(resultId: string | undefined, col: string): boolean {
  if (!resultId) return true
  if (!hiddenColumns[resultId]) {
    hiddenColumns[resultId] = new Set()
  }
  return !hiddenColumns[resultId].has(col)
}

function toggleColumn(resultId: string | undefined, col: string) {
  if (!resultId) return
  if (!hiddenColumns[resultId]) {
    hiddenColumns[resultId] = new Set()
  }
  if (hiddenColumns[resultId].has(col)) {
    hiddenColumns[resultId].delete(col)
  } else {
    hiddenColumns[resultId].add(col)
  }
}

function formatCell(row: any, col: string): string {
  if (row === null || row === undefined) return ''
  if (typeof row === 'object' && col in row) {
    const v = row[col]
    if (v === null || v === undefined) return ''
    if (typeof v === 'object') return JSON.stringify(v)
    return String(v)
  }
  return ''
}

function isNullValue(val: any): boolean {
  return val === null || val === undefined || val === ''
}

function toggleResult(idx: number) {
  expandedResults[idx] = !expandedResults[idx]
}

function totalPages(r: any, idx: number): number {
  const rows = safeRows(r)
  const ps = pageSizes[idx] || 50
  return Math.max(1, Math.ceil(rows.length / ps))
}

function paginatedRows(r: any, idx: number): any[] {
  const rows = safeRows(r)
  const page = currentPages[idx] || 1
  const ps = pageSizes[idx] || 50
  const start = (page - 1) * ps
  return rows.slice(start, start + ps)
}

function firstPage(idx: number) {
  currentPages[idx] = 1
}

function prevPage(idx: number) {
  currentPages[idx] = Math.max(1, (currentPages[idx] || 1) - 1)
}

function nextPage(r: any, idx: number) {
  currentPages[idx] = Math.min(totalPages(r, idx), (currentPages[idx] || 1) + 1)
}

function lastPage(r: any, idx: number) {
  currentPages[idx] = totalPages(r, idx)
}

function isRowModified(resultId: string, resultIdx: number, rowIdx: number): boolean {
  const page = currentPages[resultIdx] || 1
  const ps = pageSizes[resultIdx] || 50
  const actualIdx = (page - 1) * ps + rowIdx
  const rId = resultId + '-' + resultIdx
  const rows = modifiedRows[rId]
  return rows !== undefined && actualIdx in rows
}

function hasChanges(r: any, idx: number): boolean {
  const rId = (r.id || 'result') + '-' + idx
  return Object.keys(modifiedRows[rId] || {}).length > 0 || 
         (deletedRows[rId] && deletedRows[rId].length > 0)
}

function hasSelectedRows(r: any, idx: number): boolean {
  const rows = safeRows(r)
  const page = currentPages[idx] || 1
  const ps = pageSizes[idx] || 50
  const start = (page - 1) * ps
  for (let i = 0; i < Math.min(ps, rows.length - start); i++) {
    if (selectedRows[(r.id || 'result') + '-' + (start + i)]) return true
  }
  return false
}

function isAllSelected(r: any, idx: number): boolean {
  const rows = safeRows(r)
  const page = currentPages[idx] || 1
  const ps = pageSizes[idx] || 50
  const start = (page - 1) * ps
  const end = Math.min(start + ps, rows.length)
  for (let i = start; i < end; i++) {
    if (!selectedRows[(r.id || 'result') + '-' + i]) return false
  }
  return end > start
}

function toggleSelectAll(r: any, idx: number) {
  const rows = safeRows(r)
  const page = currentPages[idx] || 1
  const ps = pageSizes[idx] || 50
  const start = (page - 1) * ps
  const end = Math.min(start + ps, rows.length)
  const isAll = isAllSelected(r, idx)
  for (let i = start; i < end; i++) {
    selectedRows[(r.id || 'result') + '-' + i] = !isAll
  }
}

function canEdit(r: any): boolean {
  return r.success !== false && safeRows(r).length > 0
}

function addNewRow(r: any, idx: number) {
  const rows = safeRows(r)
  if (rows.length === 0) return
  const newRow: any = {}
  for (const key of Object.keys(rows[0])) {
    newRow[key] = ''
  }
  rows.push(newRow)
  const rId = (r.id || 'result') + '-' + idx
  if (!originalRows[rId]) {
    originalRows[rId] = JSON.parse(JSON.stringify(rows))
  }
  modifiedRows[rId] = modifiedRows[rId] || {}
  modifiedRows[rId][rows.length - 1] = newRow
}

function deleteSelectedRows(r: any, idx: number) {
  const rows = safeRows(r)
  const rId = (r.id || 'result') + '-' + idx
  const page = currentPages[idx] || 1
  const ps = pageSizes[idx] || 50
  const start = (page - 1) * ps
  
  if (!originalRows[rId]) {
    originalRows[rId] = JSON.parse(JSON.stringify(rows))
  }
  deletedRows[rId] = deletedRows[rId] || []
  
  const toDelete: number[] = []
  for (let i = 0; i < ps && start + i < rows.length; i++) {
    if (selectedRows[rId + '-' + (start + i)]) {
      toDelete.push(start + i)
    }
  }
  
  toDelete.sort((a, b) => b - a)
  for (const i of toDelete) {
    deletedRows[rId].push(rows.splice(i, 1)[0])
    delete selectedRows[rId + '-' + i]
  }
  
  modifiedRows[rId] = modifiedRows[rId] || {}
  for (const key of Object.keys(modifiedRows[rId])) {
    const numKey = parseInt(key)
    if (numKey >= start + toDelete.length) {
      modifiedRows[rId][numKey - toDelete.length] = modifiedRows[rId][numKey]
      delete modifiedRows[rId][numKey]
    }
  }
}

function startEdit(r: any, idx: number, rowIdx: number, col: string) {
  const page = currentPages[idx] || 1
  const ps = pageSizes[idx] || 50
  const actualIdx = (page - 1) * ps + rowIdx
  editingCell[r.id + '-' + actualIdx] = col
  editValue.value = formatCell(safeRows(r)[rowIdx], col)
  
  const rId = r.id + '-' + idx
  if (!originalRows[rId]) {
    originalRows[rId] = JSON.parse(JSON.stringify(safeRows(r)))
  }
}

function endEdit(r: any, idx: number, rowIdx: number, col: string) {
  const page = currentPages[idx] || 1
  const ps = pageSizes[idx] || 50
  const actualIdx = (page - 1) * ps + rowIdx
  
  const rows = safeRows(r)
  if (rows[rowIdx] && editValue.value !== formatCell(rows[rowIdx], col)) {
    rows[rowIdx][col] = editValue.value
    const rId = r.id + '-' + idx
    modifiedRows[rId] = modifiedRows[rId] || {}
    modifiedRows[rId][actualIdx] = { ...rows[rowIdx] }
  }
  
  delete editingCell[r.id + '-' + actualIdx]
  editValue.value = ''
}

function cancelEdit(r: any, idx: number, rowIdx: number) {
  const page = currentPages[idx] || 1
  const ps = pageSizes[idx] || 50
  const actualIdx = (page - 1) * ps + rowIdx
  delete editingCell[r.id + '-' + actualIdx]
  editValue.value = ''
}

function submitChanges(r: any, idx: number) {
  isSubmitting[idx] = true
  setTimeout(() => {
    const rId = r.id + '-' + idx
    delete modifiedRows[rId]
    delete deletedRows[rId]
    delete originalRows[rId]
    lastUpdated[idx] = new Date().toLocaleString('zh-CN')
    isSubmitting[idx] = false
  }, 500)
}

function rollbackChanges(r: any, idx: number) {
  const rId = r.id + '-' + idx
  if (originalRows[rId]) {
    const rows = safeRows(r)
    rows.splice(0, rows.length, ...JSON.parse(JSON.stringify(originalRows[rId])))
  }
  delete modifiedRows[rId]
  delete deletedRows[rId]
  delete originalRows[rId]
}

function refreshResult(r: any, idx: number) {
  emit('refresh-single', r.id || '', idx, r.sql || '')
}

function toggleExportMenu(id: string | undefined) {
  exportMenuOpen.value = exportMenuOpen.value === id ? null : id
}

function exportData(r: any, idx: number, format: string, selectedOnly: boolean) {
  toggleExportMenu(r.id)
}

function handleStop(r: any, idx: number) {
  isSubmitting[idx] = false
}

function formatShellOutput(r: any): string {
  const rows = safeRows(r)
  const columns = safeColumns(r)
  if (columns.length === 0) return ''
  
  const colWidths: number[] = []
  for (const col of columns) {
    const maxLen = Math.max(col.length, ...rows.map((row: any) => String(formatCell(row, col)).length))
    colWidths.push(maxLen + 2)
  }
  
  let output = columns.map((col, i) => col.padEnd(colWidths[i])).join('') + '\n'
  output += columns.map((_, i) => '-'.repeat(colWidths[i])).join('') + '\n'
  for (const row of rows) {
    output += columns.map((col, i) => formatCell(row, col).padEnd(colWidths[i])).join('') + '\n'
  }
  return output
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
  display: flex;
  flex-direction: column;
  height: 100%;
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
  font-weight: 600;
  color: #303133;
  font-size: 14px;
}

.rt-result-affected, .rt-result-rows, .rt-result-dura {
  font-size: 12px;
  color: #6c757d;
  padding: 2px 6px;
  background: #e9ecef;
  border-radius: 4px;
}

.rt-result-status {
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 500;
}

.rt-result-status.ok {
  background: #d4edda;
  color: #155724;
}

.rt-result-status.err {
  background: #f8d7da;
  color: #721c24;
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
.rt-action-btn-submit .rt-action-icon { color: #28a745; }
.rt-action-btn-rollback .rt-action-icon { color: #dc3545; }
.rt-action-btn-refresh .rt-action-icon { color: #1976d2; }
.rt-action-btn-export .rt-action-icon { color: #6c757d; }
.rt-action-btn-stop .rt-action-icon { color: #dc3545; }
.rt-action-btn-columns .rt-action-icon { color: #6c757d; }

.rt-action-separator {
  width: 1px;
  height: 16px;
  background: #dee2e6;
  margin: 0 4px;
}

.rt-export-menu {
  position: absolute;
  right: 32px;
  top: 40px;
  background: #ffffff;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  z-index: 100;
}

.rt-export-item {
  padding: 8px 16px;
  font-size: 12px;
  color: #495057;
  cursor: pointer;
  white-space: nowrap;
}

.rt-export-item:hover {
  background: #f8f9fa;
}

.rt-column-menu {
  position: absolute;
  right: 32px;
  top: 40px;
  background: #ffffff;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  z-index: 100;
  max-height: 200px;
  overflow-y: auto;
}

.rt-column-item {
  padding: 6px 16px;
  font-size: 12px;
  color: #495057;
  cursor: pointer;
  white-space: nowrap;
  display: flex;
  align-items: center;
  gap: 8px;
}

.rt-column-item:hover {
  background: #f8f9fa;
}

.rt-column-item span:first-child {
  font-size: 10px;
  color: #6c757d;
}

.rt-column-item span:first-child.checked {
  color: #28a745;
}

.rt-result-close {
  margin-left: 8px;
  font-size: 14px;
  color: #6c757d;
  cursor: pointer;
  padding: 2px 4px;
}

.rt-result-close:hover {
  background: #e9ecef;
  border-radius: 4px;
}

.rt-result-body {
  background: #ffffff;
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.rt-error-box {
  padding: 12px;
  background: #fff3f3;
  border-left: 4px solid #dc3545;
}

.rt-error-title {
  font-weight: 500;
  color: #dc3545;
  margin-bottom: 4px;
  display: block;
}

.rt-error-text {
  font-family: Consolas, Monaco, monospace;
  font-size: 12px;
  color: #721c24;
  white-space: pre-wrap;
}

.rt-table-container {
  flex: 1;
  overflow: auto;
  min-height: 0;
}

.rt-table-wrapper {
  overflow-x: auto;
}

.rt-data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
  line-height: 1.5;
}

.rt-data-table th,
.rt-data-table td {
  padding: 10px 16px;
  text-align: left;
  border-bottom: 1px solid #e8e8e8;
  font-size: 14px;
}

.rt-data-table th {
  background: #fafafa;
  font-weight: 600;
  color: #606266;
  white-space: nowrap;
  position: sticky;
  top: 0;
  z-index: 1;
  font-size: 14px;
}

.rt-data-table tbody tr {
  height: 40px;
}

.rt-data-table tbody tr:hover {
  background: #f5f7fa;
}

.rt-data-table tbody tr.odd {
  background: #ffffff;
}

.rt-data-table tbody tr.even {
  background: #fafafa;
}

.rt-data-table tbody tr.modified {
  background: #fff3cd;
}

.rt-select-col {
  width: 30px;
  text-align: center;
}

.rt-row-num-col {
  width: 40px;
  text-align: center;
  color: #6c757d;
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

.rt-edit-input {
  width: 100%;
  padding: 2px 4px;
  border: 1px solid #1976d2;
  border-radius: 2px;
  outline: none;
  font-size: 12px;
  background: #e3f2fd;
}

.rt-shell-wrap {
  max-height: 400px;
  overflow: auto;
}

.rt-shell-output {
  padding: 12px;
  margin: 0;
  background: #1e1e1e;
  color: #d4d4d4;
  font-family: Consolas, Monaco, monospace;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
}

.rt-status-bar {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding: 8px 12px;
  background: #f8f9fa;
  border-top: 1px solid #e9ecef;
  font-size: 12px;
  gap: 12px;
}

.rt-status-left {
  flex: 1;
  overflow: hidden;
  text-align: left;
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

.rt-message-box {
  padding: 12px;
  background: #f8f9fa;
}

.rt-message-box pre {
  margin: 0;
  font-family: Consolas, Monaco, monospace;
  font-size: 12px;
  color: #495057;
  white-space: pre-wrap;
}
</style>