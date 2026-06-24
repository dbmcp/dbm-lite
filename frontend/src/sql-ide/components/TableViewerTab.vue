<template>
  <div class="tv-root">
    <!-- 顶部标题栏 -->
    <div class="tv-header">
      <div class="tv-title">
        <span class="tv-ds">{{ datasourceName }}</span>
        <span class="tv-separator">/</span>
        <span class="tv-db">{{ tab.database }}</span>
        <span class="tv-separator">/</span>
        <span class="tv-table">{{ tab.table }}</span>
      </div>
      <div class="tv-actions">
        <button class="tv-btn tv-btn-add" @click="addNewRow" title="添加记录">
          <span class="tv-btn-icon">+</span>
        </button>
        <button class="tv-btn tv-btn-delete" @click="deleteSelectedRows" :disabled="!hasSelectedRows" title="删除记录">
          <span class="tv-btn-icon">-</span>
        </button>
        <div class="tv-actions-separator"></div>
        <button class="tv-btn tv-btn-submit" @click="submitChanges" :disabled="!hasChanges" title="提交">
          <span class="tv-btn-icon">✓</span>
        </button>
        <button class="tv-btn tv-btn-rollback" @click="rollbackChanges" :disabled="!hasChanges" title="放弃更改">
          <span class="tv-btn-icon">✕</span>
        </button>
        <div class="tv-actions-separator"></div>
        <button class="tv-btn tv-btn-refresh" @click="refreshTable" title="刷新">
          <span class="tv-btn-icon">⟳</span>
        </button>
        <button class="tv-btn tv-btn-export" @click="showExportMenu = !showExportMenu" title="导出">
          <span class="tv-btn-icon">⤤</span>
        </button>
        <button class="tv-btn tv-btn-stop" @click="handleStop" :disabled="!isSubmitting" title="停止">
          <span class="tv-btn-icon">■</span>
        </button>
        
        <div v-if="showExportMenu" class="tv-export-menu">
          <div class="tv-export-item" @click="exportData('csv', false)">导出全表 (CSV)</div>
          <div class="tv-export-item" @click="exportData('csv', true)">导出选中行 (CSV)</div>
          <div class="tv-export-item" @click="exportData('sql', false)">导出全表 (SQL)</div>
          <div class="tv-export-item" @click="exportData('sql', true)">导出选中行 (SQL)</div>
        </div>
      </div>
    </div>

    <!-- 主表格区域 -->
    <div class="tv-table-container">
      <div class="tv-table-wrapper">
        <table class="tv-data-table">
          <thead>
            <tr>
              <th class="tv-select-col">
                <input type="checkbox" @change="toggleSelectAll" :checked="isAllSelected" />
              </th>
              <th class="tv-row-num-col">#</th>
              <th v-for="(col, ci) in tab.columns" :key="'h-' + ci" class="tv-col-header">
                <div class="tv-col-name">{{ col }}</div>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(row, ri) in paginatedRows" :key="'row-' + ri" :class="{ odd: ri % 2 === 0, editing: editingCell[ri], modified: isRowModified(ri) }">
              <td class="tv-select-col">
                <input type="checkbox" v-model="selectedRows[ri]" />
              </td>
              <td class="tv-row-num-col">{{ (currentPage - 1) * pageSize + ri + 1 }}</td>
              <td 
                v-for="(col, ci) in tab.columns" 
                :key="'c-' + ci" 
                :title="formatCell(row, col)"
                @dblclick="startEdit(ri, col)"
                class="tv-cell"
              >
                <template v-if="editingCell[ri] === col">
                  <input 
                    type="text" 
                    v-model="editValue" 
                    @blur="endEdit(ri, col)"
                    @keyup.enter="endEdit(ri, col)"
                    @keyup.escape="cancelEdit(ri)"
                    class="tv-edit-input"
                    ref="editInputRef"
                  />
                </template>
                <template v-else>
                  <span :class="{ 'tv-null': isNullValue(row[col]) }">{{ formatCell(row, col) }}</span>
                </template>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 底部状态栏 -->
    <div class="tv-status-bar">
      <div class="tv-status-left">
        <span class="tv-status-sql">SELECT * FROM {{ tab.database }}.{{ tab.table }}</span>
      </div>
      <div class="tv-status-right">
        <span class="tv-status-pagination">第 {{ currentPage }} 页 (共 {{ totalPages }} 页)</span>
        <span class="tv-status-separator">|</span>
        <span class="tv-status-records">{{ totalRows }} 条记录</span>
        <span class="tv-status-separator">|</span>
        <span class="tv-status-updated">上次更新: {{ lastUpdated }}</span>
      </div>
    </div>

    <!-- 分页控制 -->
    <div class="tv-pagination">
      <button class="tv-page-btn" @click="prevPage" :disabled="currentPage <= 1">
        <svg viewBox="0 0 16 16" width="14" height="14"><path d="M6 4l-4 4 4 4V4z"/></svg>
      </button>
      <button class="tv-page-btn" @click="firstPage" :disabled="currentPage <= 1">
        <svg viewBox="0 0 16 16" width="14" height="14"><path d="M2 4v8h2V4H2zm4 0l4 4-4 4V4z"/></svg>
      </button>
      <div class="tv-page-info">
        <input type="number" v-model.number="currentPage" min="1" :max="totalPages" class="tv-page-input" />
        <span> / {{ totalPages }}</span>
      </div>
      <button class="tv-page-btn" @click="lastPage" :disabled="currentPage >= totalPages">
        <svg viewBox="0 0 16 16" width="14" height="14"><path d="M14 4v8h-2V4h2zm-4 0l-4 4 4 4V4z"/></svg>
      </button>
      <button class="tv-page-btn" @click="nextPage" :disabled="currentPage >= totalPages">
        <svg viewBox="0 0 16 16" width="14" height="14"><path d="M10 4l4 4-4 4V4z"/></svg>
      </button>
      <div class="tv-page-separator"></div>
      <select v-model="pageSize" class="tv-page-size">
        <option :value="20">20</option>
        <option :value="50">50</option>
        <option :value="100">100</option>
        <option :value="500">500</option>
        <option :value="1000">1000</option>
      </select>
      <span class="tv-page-label">条/页</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { executeSql } from '@/sql-ide/service/api'
import type { TableTabState, DatasourceSummary } from '@/sql-ide/types'

const props = defineProps<{
  tab: TableTabState
  datasources: DatasourceSummary[]
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'refresh'): void
}>()

const lastUpdated = ref('')

const currentPage = ref(1)
const pageSize = ref(50)

const editingCell = reactive<Record<number, string>>({})
const selectedRows = reactive<Record<number, boolean>>({})
const editValue = ref('')
const modifiedRows = reactive<Record<number, any>>({})
const deletedRows = reactive<Record<number, any>>({})
const originalRows = ref<any[]>([])
const isSubmitting = ref(false)
const showExportMenu = ref(false)
let abortSubmit = false

const datasourceName = computed(() => {
  const ds = props.datasources.find(d => d.datasourceId === props.tab.datasourceId)
  return ds?.name || props.tab.datasourceId
})

const totalRows = computed(() => Array.isArray(props.tab.rows) ? props.tab.rows.length : 0)
const totalPages = computed(() => Math.max(1, Math.ceil(totalRows.value / pageSize.value)))

const paginatedRows = computed(() => {
  const rows = Array.isArray(props.tab.rows) ? props.tab.rows : []
  const start = (currentPage.value - 1) * pageSize.value
  return rows.slice(start, start + pageSize.value)
})

const hasChanges = computed(() => Object.keys(modifiedRows).length > 0 || Object.keys(deletedRows).length > 0)
const modifiedCount = computed(() => Object.keys(modifiedRows).length + Object.keys(deletedRows).length)

const isAllSelected = computed(() => {
  return paginatedRows.value.length > 0 && 
         paginatedRows.value.every((_, idx) => selectedRows[idx])
})

const hasSelectedRows = computed(() => {
  return paginatedRows.value.some((_, idx) => selectedRows[idx])
})

function formatCell(row: any, col: string): string {
  const val = row[col]
  if (val === null || val === undefined) return ''
  if (typeof val === 'object') return JSON.stringify(val)
  return String(val)
}

function isNullValue(val: any): boolean {
  return val === null || val === undefined || val === ''
}

function isRowModified(rowIdx: number): boolean {
  const actualIdx = (currentPage.value - 1) * pageSize.value + rowIdx
  return actualIdx in modifiedRows
}

function startEdit(rowIdx: number, col: string) {
  if (!Array.isArray(props.tab.rows)) return
  const actualIdx = (currentPage.value - 1) * pageSize.value + rowIdx
  if (actualIdx < props.tab.rows.length) {
    const row = props.tab.rows[actualIdx]
    editValue.value = formatCell(row, col)
    editingCell[rowIdx] = col
    setTimeout(() => {
      const inputs = document.querySelectorAll('.tv-edit-input')
      if (inputs.length > 0) {
        (inputs[inputs.length - 1] as HTMLInputElement).focus()
      }
    }, 0)
  }
}

function endEdit(rowIdx: number, col: string) {
  if (!Array.isArray(props.tab.rows)) {
    editingCell[rowIdx] = ''
    return
  }
  const actualIdx = (currentPage.value - 1) * pageSize.value + rowIdx
  if (actualIdx < props.tab.rows.length) {
    const row = props.tab.rows[actualIdx]
    const originalVal = originalRows.value[actualIdx]?.[col]
    
    if (editValue.value !== formatCell(originalVal, col)) {
      if (!(actualIdx in modifiedRows)) {
        modifiedRows[actualIdx] = { ...originalRows.value[actualIdx] }
      }
      row[col] = editValue.value
    } else {
      delete modifiedRows[actualIdx]
    }
  }
  editingCell[rowIdx] = ''
}

function cancelEdit(rowIdx: number) {
  editingCell[rowIdx] = ''
}

function addNewRow() {
  if (!Array.isArray(props.tab.columns) || !Array.isArray(props.tab.rows)) {
    ElMessage.warning('无法添加新行：表格数据未加载')
    return
  }
  const newRow: Record<string, any> = {}
  const newOriginal: Record<string, any> = {}
  props.tab.columns.forEach(col => {
    newRow[col] = ''
    newOriginal[col] = ''
  })
  props.tab.rows.push(newRow)
  originalRows.value.push(newOriginal)
  
  const newIdx = props.tab.rows.length - 1
  modifiedRows[newIdx] = newRow
  
  setTimeout(() => {
    currentPage.value = Math.ceil(props.tab.rows.length / pageSize.value)
  }, 0)
}

function deleteSelectedRows() {
  if (!Array.isArray(props.tab.rows)) {
    Object.keys(selectedRows).forEach(key => {
      delete selectedRows[key]
    })
    return
  }
  
  const selectedIndices: number[] = []
  paginatedRows.value.forEach((_, idx) => {
    if (selectedRows[idx]) {
      const actualIdx = (currentPage.value - 1) * pageSize.value + idx
      selectedIndices.push(actualIdx)
    }
  })
  
  selectedIndices.sort((a, b) => b - a).forEach(idx => {
    const deletedRow = props.tab.rows[idx]
    deletedRows[idx] = deletedRow
    props.tab.rows.splice(idx, 1)
    originalRows.value.splice(idx, 1)
    delete modifiedRows[idx]
    
    Object.keys(modifiedRows).forEach(key => {
      const numKey = parseInt(key)
      if (numKey > idx) {
        modifiedRows[numKey - 1] = modifiedRows[numKey]
        delete modifiedRows[numKey]
      }
    })
    
    Object.keys(deletedRows).forEach(key => {
      const numKey = parseInt(key)
      if (numKey > idx) {
        deletedRows[numKey - 1] = deletedRows[numKey]
        delete deletedRows[numKey]
      }
    })
  })
  
  Object.keys(selectedRows).forEach(key => {
    delete selectedRows[key]
  })
  
  const newTotalPages = Math.max(1, Math.ceil(props.tab.rows.length / pageSize.value))
  if (currentPage.value > newTotalPages) {
    currentPage.value = newTotalPages
  }
  
  ElMessage.success(`已标记 ${selectedIndices.length} 行待删除，点击提交确认删除`)
}

function toggleSelectAll() {
  const isChecked = !isAllSelected.value
  paginatedRows.value.forEach((_, idx) => {
    selectedRows[idx] = isChecked
  })
}

function firstPage() {
  currentPage.value = 1
}

function prevPage() {
  if (currentPage.value > 1) {
    currentPage.value--
  }
}

function nextPage() {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
  }
}

function lastPage() {
  currentPage.value = totalPages.value
}

function editTableStructure() {
  ElMessage.info('设计表功能即将实现')
}

async function loadTableData() {
  try {
    const res = await executeSql({
      datasourceId: props.tab.datasourceId,
      database: props.tab.database,
      sql: `SELECT * FROM \`${props.tab.database}\`.\`${props.tab.table}\` LIMIT 1000`
    })
    
    const results = Array.isArray(res) ? res : (res && res.results ? res.results : [res])
    if (results.length > 0) {
      const r = results[0]
      props.tab.rows = r.rows || r.data || []
      props.tab.columns = r.columns || r.columnNames || (props.tab.rows.length > 0 ? Object.keys(props.tab.rows[0]) : [])
      originalRows.value = props.tab.rows.map((row: any) => ({ ...row }))
    }
    
    lastUpdated.value = new Date().toLocaleString('zh-CN')
  } catch (e: any) {
    console.error('加载表数据失败:', e)
    ElMessage.error('加载表数据失败: ' + (e?.message || String(e)))
  }
}

function refreshTable() {
  currentPage.value = 1
  loadTableData()
  emit('refresh')
}

function handleStop() {
  if (isSubmitting.value) {
    abortSubmit = true
    ElMessage.info('正在停止提交...')
  }
}

function rollbackChanges() {
  if (!Array.isArray(props.tab.rows)) {
    Object.keys(modifiedRows).forEach(key => delete modifiedRows[key])
    Object.keys(deletedRows).forEach(key => delete deletedRows[key])
    ElMessage.success('已回滚所有修改')
    return
  }
  
  Object.keys(modifiedRows).forEach(idx => {
    const numIdx = parseInt(idx)
    if (originalRows.value[numIdx]) {
      props.tab.rows[numIdx] = { ...originalRows.value[numIdx] }
    }
    delete modifiedRows[idx]
  })
  
  const deletedKeys = Object.keys(deletedRows).map(k => parseInt(k)).sort((a, b) => b - a)
  deletedKeys.forEach(idx => {
    props.tab.rows.splice(idx, 0, deletedRows[idx])
    originalRows.value.splice(idx, 0, deletedRows[idx])
    delete deletedRows[idx]
  })
  
  Object.keys(deletedRows).forEach(key => {
    delete deletedRows[key]
  })
  
  ElMessage.success('已回滚所有修改')
}

async function submitChanges() {
  if (isSubmitting.value) return
  
  if (!Array.isArray(props.tab.columns) || !Array.isArray(props.tab.rows)) {
    isSubmitting.value = false
    return
  }
  
  isSubmitting.value = true
  abortSubmit = false
  
  try {
    const pkCols = props.tab.columns.filter(c => c.toLowerCase().includes('id'))
    const pkCol = pkCols[0] || props.tab.columns[0]
    
    const modifiedKeys = Object.keys(modifiedRows)
    for (let i = 0; i < modifiedKeys.length; i++) {
      if (abortSubmit) {
        ElMessage.warning('提交已取消')
        return
      }
      
      const idxStr = modifiedKeys[i]
      const idx = parseInt(idxStr)
      const original = originalRows.value[idx]
      const current = props.tab.rows[idx]
      
      if (!original || !current) continue
      
      const pkValue = original[pkCol]
      
      if (pkValue === undefined || pkValue === null || pkValue === '') {
        const emptyRequired = props.tab.columns.filter(col => {
          const val = current[col]
          return val === undefined || val === null || val === ''
        })
        
        if (emptyRequired.length > 0) {
          isSubmitting.value = false
          ElMessage.error(`新增行失败：以下字段不能为空：${emptyRequired.join(', ')}`)
          return
        }
        
        await insertRow(current)
      } else {
        await updateRow(current, pkCol, pkValue)
      }
      
      await new Promise(resolve => setTimeout(resolve, 50))
    }
    
    if (!abortSubmit) {
      const deletedKeys = Object.keys(deletedRows)
      for (let i = 0; i < deletedKeys.length; i++) {
        if (abortSubmit) {
          ElMessage.warning('提交已取消')
          return
        }
        
        const idxStr = deletedKeys[i]
        const deletedRow = deletedRows[idxStr]
        
        const pkCols = props.tab.columns.filter(c => c.toLowerCase().includes('id'))
        const pkCol = pkCols[0] || props.tab.columns[0]
        const pkValue = deletedRow[pkCol]
        
        if (pkValue !== undefined && pkValue !== null && pkValue !== '') {
          await deleteRow(deletedRow, pkCol, pkValue)
        }
        
        await new Promise(resolve => setTimeout(resolve, 50))
      }
    }
    
    if (!abortSubmit) {
      await refreshTable()
      ElMessage.success('提交成功')
    }
  } catch (e: any) {
    if (!abortSubmit) {
      console.error('提交失败:', e)
      ElMessage.error('提交失败: ' + (e?.message || String(e)))
    }
  } finally {
    isSubmitting.value = false
    abortSubmit = false
  }
}

async function insertRow(row: any) {
  const cols = Object.keys(row).filter(k => row[k] !== '')
  const values = cols.map(c => {
    const v = row[c]
    if (v === null || v === undefined) return 'NULL'
    if (typeof v === 'string') return `'${v.replace(/'/g, "''")}'`
    return String(v)
  })
  
  const sql = `INSERT INTO \`${props.tab.database}\`.\`${props.tab.table}\` (${cols.map(c => `\`${c}\``).join(', ')}) VALUES (${values.join(', ')})`
  await executeSql({
    datasourceId: props.tab.datasourceId,
    database: props.tab.database,
    sql
  })
}

async function updateRow(row: any, pkCol: string, pkValue: any) {
  const sets = Object.keys(row).filter(k => k !== pkCol).map(c => {
    const v = row[c]
    if (v === null || v === undefined) return `\`${c}\` = NULL`
    if (typeof v === 'string') return `\`${c}\` = '${v.replace(/'/g, "''")}'`
    return `\`${c}\` = ${String(v)}`
  })
  
  const pkStr = typeof pkValue === 'string' ? `'${pkValue.replace(/'/g, "''")}'` : String(pkValue)
  const sql = `UPDATE \`${props.tab.database}\`.\`${props.tab.table}\` SET ${sets.join(', ')} WHERE \`${pkCol}\` = ${pkStr}`
  await executeSql({
    datasourceId: props.tab.datasourceId,
    database: props.tab.database,
    sql
  })
}

async function deleteRow(row: any, pkCol: string, pkValue: any) {
  const pkStr = typeof pkValue === 'string' ? `'${pkValue.replace(/'/g, "''")}'` : String(pkValue)
  const sql = `DELETE FROM \`${props.tab.database}\`.\`${props.tab.table}\` WHERE \`${pkCol}\` = ${pkStr}`
  await executeSql({
    datasourceId: props.tab.datasourceId,
    database: props.tab.database,
    sql
  })
}

function exportData(format: 'csv' | 'sql', selectedOnly: boolean) {
  showExportMenu.value = false
  
  if (!Array.isArray(props.tab.rows) || !Array.isArray(props.tab.columns)) {
    ElMessage.warning('无法导出：表格数据未加载')
    return
  }
  
  let exportRows: any[]
  if (selectedOnly) {
    exportRows = []
    paginatedRows.value.forEach((row, idx) => {
      if (selectedRows[idx]) {
        exportRows.push(row)
      }
    })
    if (exportRows.length === 0) {
      ElMessage.warning('请先选中要导出的行')
      return
    }
  } else {
    exportRows = props.tab.rows
  }
  
  let content = ''
  let filename = ''
  
  if (format === 'csv') {
    content = props.tab.columns.join(',') + '\n'
    exportRows.forEach(row => {
      const values = props.tab.columns.map(col => {
        const v = row[col]
        if (v === null || v === undefined) return ''
        if (typeof v === 'string' && v.includes(',')) {
          return `"${v.replace(/"/g, '""')}"`
        }
        return String(v)
      })
      content += values.join(',') + '\n'
    })
    filename = `${props.tab.table}_export_${Date.now()}.csv`
  } else {
    content = `-- Export from ${props.tab.database}.${props.tab.table}\n`
    content = `-- Exported: ${new Date().toISOString()}\n\n`
    exportRows.forEach(row => {
      const cols = props.tab.columns.filter(col => row[col] !== undefined && row[col] !== null)
      const values = cols.map(col => {
        const v = row[col]
        if (v === null) return 'NULL'
        if (typeof v === 'string') return `'${v.replace(/'/g, "''")}'`
        return String(v)
      })
      content += `INSERT INTO \`${props.tab.database}\`.\`${props.tab.table}\` (`
      content += cols.map(c => `\`${c}\``).join(', ')
      content += ') VALUES (' + values.join(', ') + ');\n'
    })
    filename = `${props.tab.table}_export_${Date.now()}.sql`
  }
  
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
  
  ElMessage.success(`已导出 ${exportRows.length} 行数据`)
}

watch(() => props.tab.table, () => {
  loadTableData()
})

watch(pageSize, () => {
  currentPage.value = 1
})

onMounted(() => {
  if (!Array.isArray(props.tab.rows) || props.tab.rows.length === 0) {
    loadTableData()
  } else {
    originalRows.value = props.tab.rows.map((row: any) => ({ ...row }))
    lastUpdated.value = new Date().toLocaleString('zh-CN')
  }
})
</script>

<style scoped>
.tv-root {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #ffffff;
}

.tv-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background: linear-gradient(180deg, #f8f9fa 0%, #e9ecef 100%);
  border-bottom: 1px solid #dee2e6;
}

.tv-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;
}

.tv-ds { color: #28a745; }
.tv-db { color: #007bff; }
.tv-table { color: #6c757d; font-weight: 600; }
.tv-separator { color: #adb5bd; }

.tv-actions {
  display: flex;
  align-items: center;
  gap: 0;
  background: #ffffff;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  padding: 2px;
}

.tv-actions-separator {
  width: 1px;
  height: 20px;
  background: #dee2e6;
  margin: 0 2px;
}

.tv-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 22px;
  border: none;
  border-radius: 3px;
  background: transparent;
  cursor: pointer;
  transition: all 0.15s ease;
  font-size: 14px;
  font-weight: 600;
  color: #495057;
}

.tv-btn-icon {
  line-height: 1;
}

.tv-btn:hover {
  background: #e3f2fd;
  color: #007bff;
}

.tv-btn:active {
  background: #bbdefb;
}

.tv-btn-add:hover {
  background: #e8f5e9;
  color: #28a745;
}

.tv-btn-delete:hover {
  background: #ffebee;
  color: #dc3545;
}

.tv-btn-submit {
  color: #28a745;
}

.tv-btn-submit:hover:not(:disabled) {
  background: #e8f5e9;
  color: #28a745;
}

.tv-btn-rollback {
  color: #dc3545;
}

.tv-btn-rollback:hover:not(:disabled) {
  background: #ffebee;
  color: #dc3545;
}

.tv-btn-refresh:hover {
  background: #e3f2fd;
  color: #007bff;
}

.tv-btn-stop:hover {
  background: #ffebee;
  color: #dc3545;
}

.tv-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  pointer-events: none;
}

.tv-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  background: #f1f3f4;
  border-bottom: 1px solid #dee2e6;
}

.tv-tool-btn {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 4px 10px;
  background: #ffffff;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
  color: #495057;
  transition: all 0.15s;
}

.tv-tool-btn:hover:not(:disabled) {
  background: #e3f2fd;
  border-color: #1976d2;
  color: #1976d2;
}

.tv-tool-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.tv-tool-separator {
  width: 1px;
  height: 18px;
  background: #dee2e6;
  margin: 0 6px;
}

.tv-info {
  font-size: 12px;
  color: #6c757d;
}

.tv-table-container {
  flex: 1;
  overflow: auto;
  padding: 4px;
}

.tv-table-wrapper {
  overflow: auto;
}

.tv-data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
  font-family: Consolas, Monaco, 'Courier New', monospace;
  min-width: 100%;
}

.tv-data-table thead th {
  position: sticky;
  top: 0;
  background: #e9ecef;
  color: #495057;
  border: 1px solid #dee2e6;
  padding: 6px 10px;
  text-align: left;
  font-weight: 600;
  white-space: nowrap;
  z-index: 1;
}

.tv-data-table tbody td {
  border: 1px solid #dee2e6;
  padding: 4px 10px;
  color: #212529;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tv-data-table tbody tr.odd { background: #ffffff; }
.tv-data-table tbody tr.even { background: #f8f9fa; }
.tv-data-table tbody tr:hover { background: #e3f2fd; }
.tv-data-table tbody tr.modified { background: #fff3cd; }

.tv-select-col {
  width: 28px;
  text-align: center;
  padding: 4px 4px !important;
}

.tv-row-num-col {
  width: 40px;
  text-align: right;
  color: #6c757d;
  font-style: italic;
}

.tv-col-header {
  user-select: none;
}

.tv-col-name {
  display: flex;
  align-items: center;
  gap: 4px;
}

.tv-cell {
  cursor: cell;
  transition: background 0.1s;
}

.tv-cell:hover {
  background: #e9ecef;
}

.tv-null {
  color: #adb5bd;
  font-style: italic;
}

.tv-edit-input {
  width: 100%;
  background: #ffffff;
  border: 1px solid #007bff;
  border-radius: 2px;
  padding: 3px 6px;
  font-size: 12px;
  font-family: Consolas, Monaco, monospace;
  color: #212529;
  outline: none;
  box-sizing: border-box;
}

.tv-status-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 12px;
  background: #f1f3f4;
  border-top: 1px solid #dee2e6;
  font-size: 11px;
  color: #6c757d;
}

.tv-status-sql {
  font-family: Consolas, Monaco, monospace;
  color: #495057;
}

.tv-status-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.tv-status-separator {
  color: #dee2e6;
}

.tv-status-pagination, .tv-status-records, .tv-status-updated {
  color: #6c757d;
}

.tv-pagination {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: #e9ecef;
  border-top: 1px solid #dee2e6;
}

.tv-page-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 22px;
  background: #ffffff;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s;
}

.tv-page-btn:hover:not(:disabled) {
  background: #e3f2fd;
  border-color: #1976d2;
}

.tv-page-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.tv-page-info {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #6c757d;
}

.tv-page-input {
  width: 45px;
  padding: 2px 4px;
  background: #ffffff;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  color: #212529;
  font-size: 12px;
  text-align: center;
  outline: none;
}

.tv-page-input:focus {
  border-color: #007bff;
}

.tv-page-separator {
  width: 1px;
  height: 16px;
  background: #dee2e6;
  margin: 0 4px;
}

.tv-page-size {
  padding: 2px 6px;
  background: #ffffff;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  color: #495057;
  font-size: 12px;
  outline: none;
}

.tv-page-label {
  font-size: 11px;
  color: #6c757d;
}

.tv-export-menu {
  position: absolute;
  top: 30px;
  right: 0;
  background: #ffffff;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  z-index: 100;
  min-width: 140px;
}

.tv-export-item {
  padding: 6px 12px;
  font-size: 12px;
  color: #495057;
  cursor: pointer;
  white-space: nowrap;
}

.tv-export-item:hover {
  background: #e3f2fd;
  color: #007bff;
}
</style>