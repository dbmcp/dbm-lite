<template>
  <div class="tv-root">
    <!-- 顶部标题栏 -->
    <div class="tv-header">
      <div class="tv-title">
        <span class="tv-ds">{{ datasourceName }}</span>
        <span class="tv-separator">/</span>
        <span class="tv-db">{{ databaseName }}</span>
        <span class="tv-separator">/</span>
        <span class="tv-table">{{ tableName }}</span>
      </div>
      <div class="tv-actions">
        <button class="tv-btn tv-btn-icon" @click="refreshTable" title="刷新数据">
          <svg viewBox="0 0 16 16" width="16" height="16"><path d="M12 6h-3V3c0-.6-.4-1-1-1H8V1c0-.6-.4-1-1-1H5c-.6 0-1 .4-1 1v1H2c-.6 0-1 .4-1 1v10c0 .6.4 1 1 1h8c.6 0 1-.4 1-1v-3h3c.6 0 1-.4 1-1V7c0-.6-.4-1-1-1zM5 3h1v1H5V3zm4 0h1v1H9V3zm3 8H6v-1h6v1zm0-3H6V7h6v1zm0-3H5V4h7v1z"/></svg>
        </button>
        <button class="tv-btn tv-btn-icon" @click="rollbackChanges" :disabled="!hasChanges" title="回滚修改">
          <svg viewBox="0 0 16 16" width="16" height="16"><path d="M8 3L6 5l4 4-4 4 2 2 6-6z" fill="#ff9800"/></svg>
        </button>
        <button class="tv-btn tv-btn-icon tv-btn-primary" @click="submitChanges" :disabled="!hasChanges" title="提交修改">
          <svg viewBox="0 0 16 16" width="16" height="16"><path d="M2 13h4v1H2v-1zm5-6h4v1H7V7zm3 6h4v1h-4v-1zM2 7h4v1H2V7zm10-4H8v1h4V3zM2 3h4v1H2V3z" fill="#4caf50"/></svg>
        </button>
        <button class="tv-btn tv-btn-icon tv-btn-danger" @click="closeViewer" title="关闭">
          <svg viewBox="0 0 16 16" width="16" height="16"><path d="M12 4L9 7l3 3-1 1-3-3-3 3-1-1 3-3-3-3 1-1 3 3 3-3z"/></svg>
        </button>
      </div>
    </div>

    <!-- 工具栏 -->
    <div class="tv-toolbar">
      <button class="tv-tool-btn" @click="addNewRow" title="添加行">
        <svg viewBox="0 0 16 16" width="14" height="14"><rect x="2" y="2" width="12" height="12" rx="1" fill="#ffffff" stroke="#2e7d32" stroke-width="0.8"/><path d="M7 4 v8 M3 8 h8" stroke="#2e7d32" stroke-width="1.2"/></svg>
        <span>添加行</span>
      </button>
      <button class="tv-tool-btn" @click="deleteSelectedRows" :disabled="!hasSelectedRows" title="删除选中行">
        <svg viewBox="0 0 16 16" width="14" height="14"><rect x="2" y="2" width="12" height="12" rx="1" fill="#ffffff" stroke="#c62828" stroke-width="0.8"/><path d="M4 4 L12 12 M12 4 L4 12" stroke="#c62828" stroke-width="1.2"/></svg>
        <span>删除选中</span>
      </button>
      <div class="tv-tool-separator"></div>
      <button class="tv-tool-btn" @click="editTableStructure" title="设计表">
        <svg viewBox="0 0 16 16" width="14" height="14"><rect x="2" y="2" width="12" height="12" rx="1" fill="#ffffff" stroke="#1976d2" stroke-width="0.8"/><path d="M3 5 h3 v3 h-3 z M6 11 h6 v1 h-6 z M9 8 h3 v3 h-3 z" stroke="#1976d2" stroke-width="0.8" fill="#e3f2fd"/></svg>
        <span>设计表</span>
      </button>
      <div class="tv-tool-separator"></div>
      <span class="tv-info">共 {{ totalRows }} 行 | 已修改 {{ modifiedCount }} 行</span>
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
              <th v-for="(col, ci) in columns" :key="'h-' + ci" class="tv-col-header" :style="{ width: getColumnWidth(col) }">
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
                v-for="(col, ci) in columns" 
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
        <span class="tv-status-sql">SELECT * FROM {{ databaseName }}.{{ tableName }} LIMIT {{ pageSize }} OFFSET {{ (currentPage - 1) * pageSize }}</span>
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
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { listDatasources, executeSql } from '@/sql-ide/service/api'

const route = useRoute()
const router = useRouter()

const datasourceId = ref('')
const databaseName = ref('')
const tableName = ref('')
const datasourceName = ref('')
const lastUpdated = ref('')

const rows = ref<any[]>([])
const originalRows = ref<any[]>([])
const columns = ref<string[]>([])
const currentPage = ref(1)
const pageSize = ref(50)

// 编辑相关状态
const editingCell = reactive<Record<number, string>>({})
const selectedRows = reactive<Record<number, boolean>>({})
const editValue = ref('')
const modifiedRows = reactive<Record<number, any>>({})

const totalRows = computed(() => rows.value.length)
const totalPages = computed(() => Math.max(1, Math.ceil(totalRows.value / pageSize.value)))

const paginatedRows = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return rows.value.slice(start, start + pageSize.value)
})

const hasChanges = computed(() => Object.keys(modifiedRows).length > 0)
const modifiedCount = computed(() => Object.keys(modifiedRows).length)

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

function getColumnWidth(col: string): string {
  const maxWidth = Math.max(
    col.length * 12,
    ...rows.value.map(row => String(row[col] || '').length * 10)
  )
  return `${Math.min(Math.max(maxWidth, 80), 400)}px`
}

function isRowModified(rowIdx: number): boolean {
  const actualIdx = (currentPage.value - 1) * pageSize.value + rowIdx
  return actualIdx in modifiedRows
}

function startEdit(rowIdx: number, col: string) {
  const actualIdx = (currentPage.value - 1) * pageSize.value + rowIdx
  if (actualIdx < rows.value.length) {
    const row = rows.value[actualIdx]
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
  const actualIdx = (currentPage.value - 1) * pageSize.value + rowIdx
  if (actualIdx < rows.value.length) {
    const row = rows.value[actualIdx]
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
  const newRow: Record<string, any> = {}
  const newOriginal: Record<string, any> = {}
  columns.value.forEach(col => {
    newRow[col] = ''
    newOriginal[col] = ''
  })
  rows.value.push(newRow)
  originalRows.value.push(newOriginal)
  
  const newIdx = rows.value.length - 1
  modifiedRows[newIdx] = newRow
  
  setTimeout(() => {
    currentPage.value = Math.ceil(rows.value.length / pageSize.value)
  }, 0)
}

function deleteSelectedRows() {
  const selectedIndices: number[] = []
  paginatedRows.value.forEach((_, idx) => {
    if (selectedRows[idx]) {
      const actualIdx = (currentPage.value - 1) * pageSize.value + idx
      selectedIndices.push(actualIdx)
    }
  })
  
  selectedIndices.sort((a, b) => b - a).forEach(idx => {
    rows.value.splice(idx, 1)
    originalRows.value.splice(idx, 1)
    delete modifiedRows[idx]
  })
  
  Object.keys(selectedRows).forEach(key => {
    delete selectedRows[key]
  })
  
  if (currentPage.value > totalPages.value) {
    currentPage.value = totalPages.value
  }
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
    const params = route.params as any
    datasourceId.value = params.datasourceId
    databaseName.value = params.database
    tableName.value = params.table
    
    const dsList = await listDatasources()
    const ds = dsList.find((d: any) => d.datasourceId === datasourceId.value || d.id === datasourceId.value)
    if (ds) {
      datasourceName.value = ds.name || ds.datasourceId
    }
    
    const res = await executeSql({
      datasourceId: datasourceId.value,
      database: databaseName.value,
      sql: `SELECT * FROM \`${databaseName.value}\`.\`${tableName.value}\``
    })
    
    const results = Array.isArray(res) ? res : (res && res.results ? res.results : [res])
    if (results.length > 0) {
      const r = results[0]
      rows.value = r.rows || r.data || []
      columns.value = r.columns || r.columnNames || (rows.value.length > 0 ? Object.keys(rows.value[0]) : [])
      originalRows.value = rows.value.map((row: any) => ({ ...row }))
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
}

function rollbackChanges() {
  Object.keys(modifiedRows).forEach(idx => {
    const numIdx = parseInt(idx)
    if (originalRows.value[numIdx]) {
      rows.value[numIdx] = { ...originalRows.value[numIdx] }
    }
    delete modifiedRows[idx]
  })
  ElMessage.success('已回滚所有修改')
}

async function submitChanges() {
  try {
    for (const idxStr of Object.keys(modifiedRows)) {
      const idx = parseInt(idxStr)
      const original = originalRows.value[idx]
      const current = rows.value[idx]
      
      if (!original || !current) continue
      
      const pkCols = columns.value.filter(c => c.toLowerCase().includes('id'))
      const pkCol = pkCols[0] || columns.value[0]
      const pkValue = original[pkCol]
      
      if (pkValue === undefined || pkValue === null || pkValue === '') {
        await insertRow(current)
      } else {
        await updateRow(current, pkCol, pkValue)
      }
    }
    
    await refreshTable()
    ElMessage.success('提交成功')
  } catch (e: any) {
    console.error('提交失败:', e)
    ElMessage.error('提交失败: ' + (e?.message || String(e)))
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
  
  const sql = `INSERT INTO \`${databaseName.value}\`.\`${tableName.value}\` (${cols.map(c => `\`${c}\``).join(', ')}) VALUES (${values.join(', ')})`
  await executeSql({
    datasourceId: datasourceId.value,
    database: databaseName.value,
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
  const sql = `UPDATE \`${databaseName.value}\`.\`${tableName.value}\` SET ${sets.join(', ')} WHERE \`${pkCol}\` = ${pkStr}`
  await executeSql({
    datasourceId: datasourceId.value,
    database: databaseName.value,
    sql
  })
}

function closeViewer() {
  router.push('/sql/sqlide')
}

onMounted(() => {
  loadTableData()
})

watch(() => route.params, () => {
  loadTableData()
}, { deep: true })

watch(pageSize, () => {
  currentPage.value = 1
})
</script>

<style scoped>
.tv-root {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #1e1e1e;
  color: #d4d4d4;
}

.tv-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background: linear-gradient(180deg, #3c3c3c 0%, #2d2d2d 100%);
  border-bottom: 1px solid #1a1a1a;
}

.tv-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;
}

.tv-ds { color: #6a9955; }
.tv-db { color: #569cd6; }
.tv-table { color: #dcdcaa; }
.tv-separator { color: #808080; }

.tv-actions {
  display: flex;
  gap: 4px;
}

.tv-btn {
  padding: 4px 8px;
  border: none;
  border-radius: 2px;
  cursor: pointer;
  transition: all 0.15s;
  background: transparent;
}

.tv-btn-icon:hover {
  background: #3c3c3c;
}

.tv-btn-primary:hover {
  background: #265f9c;
}

.tv-btn-danger:hover {
  background: #8b0000;
}

.tv-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.tv-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  background: #252526;
  border-bottom: 1px solid #1a1a1a;
}

.tv-tool-btn {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 4px 10px;
  background: #3c3c3c;
  border: none;
  border-radius: 2px;
  font-size: 12px;
  cursor: pointer;
  color: #d4d4d4;
  transition: all 0.15s;
}

.tv-tool-btn:hover:not(:disabled) {
  background: #007acc;
}

.tv-tool-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.tv-tool-separator {
  width: 1px;
  height: 18px;
  background: #3c3c3c;
  margin: 0 6px;
}

.tv-info {
  font-size: 12px;
  color: #858585;
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
  background: #2d2d2d;
  color: #c6c6c6;
  border: 1px solid #3c3c3c;
  padding: 6px 10px;
  text-align: left;
  font-weight: 600;
  white-space: nowrap;
  z-index: 1;
}

.tv-data-table tbody td {
  border: 1px solid #3c3c3c;
  padding: 4px 10px;
  color: #d4d4d4;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tv-data-table tbody tr.odd { background: #1e1e1e; }
.tv-data-table tbody tr.even { background: #252526; }
.tv-data-table tbody tr:hover { background: #2a2d2e; }
.tv-data-table tbody tr.modified { background: #2d2a00; }

.tv-select-col {
  width: 28px;
  text-align: center;
  padding: 4px 4px !important;
}

.tv-row-num-col {
  width: 40px;
  text-align: right;
  color: #858585;
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
  background: #2a2d2e;
}

.tv-null {
  color: #858585;
  font-style: italic;
}

.tv-edit-input {
  width: 100%;
  background: #1e1e1e;
  border: 1px solid #007acc;
  border-radius: 2px;
  padding: 3px 6px;
  font-size: 12px;
  font-family: Consolas, Monaco, monospace;
  color: #d4d4d4;
  outline: none;
  box-sizing: border-box;
}

.tv-status-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 12px;
  background: #252526;
  border-top: 1px solid #1a1a1a;
  font-size: 11px;
  color: #858585;
}

.tv-status-sql {
  font-family: Consolas, Monaco, monospace;
  color: #d4d4d4;
}

.tv-status-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.tv-status-separator {
  color: #3c3c3c;
}

.tv-status-pagination, .tv-status-records, .tv-status-updated {
  color: #858585;
}

.tv-pagination {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: #2d2d2d;
  border-top: 1px solid #1a1a1a;
}

.tv-page-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 22px;
  background: #3c3c3c;
  border: none;
  border-radius: 2px;
  cursor: pointer;
  transition: all 0.15s;
}

.tv-page-btn:hover:not(:disabled) {
  background: #007acc;
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
  color: #858585;
}

.tv-page-input {
  width: 45px;
  padding: 2px 4px;
  background: #1e1e1e;
  border: 1px solid #3c3c3c;
  border-radius: 2px;
  color: #d4d4d4;
  font-size: 12px;
  text-align: center;
  outline: none;
}

.tv-page-input:focus {
  border-color: #007acc;
}

.tv-page-separator {
  width: 1px;
  height: 16px;
  background: #3c3c3c;
  margin: 0 4px;
}

.tv-page-size {
  padding: 2px 6px;
  background: #1e1e1e;
  border: 1px solid #3c3c3c;
  border-radius: 2px;
  color: #d4d4d4;
  font-size: 12px;
  outline: none;
}

.tv-page-label {
  font-size: 11px;
  color: #858585;
}
</style>