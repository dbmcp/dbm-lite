<template>
  <div class="gt-container" :class="{ 'gt-auto-height': autoHeight }">
    <div v-if="rows.length === 0" class="gt-empty">
      <div class="gt-empty-icon">📋</div>
      <div>{{ emptyText || '暂无数据' }}</div>
    </div>
    <div v-else class="gt-wrapper">
      <div v-if="showHeader" class="gt-toolbar">
        <span class="gt-title">{{ title }}</span>
        <div v-if="showColumnPicker && columns.length > 0" class="gt-actions">
          <button class="gt-col-btn" @click="toggleColumnPicker" title="列选择">
            <span>☰</span>
            <span>列选择</span>
          </button>
        </div>
      </div>
      
      <div class="gt-body">
        <table class="gt-table">
          <thead>
            <tr>
              <th v-if="showCheckbox" class="gt-col-check"><input type="checkbox" @change="toggleSelectAll" :checked="isAllSelected" /></th>
              <th v-if="showRowNumber" class="gt-col-num">#</th>
              <th v-for="(col, ci) in visibleColumns" :key="'h'+ci" :style="col.width ? { width: col.width, minWidth: col.width } : {}">
                {{ col.label }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(row, ri) in paginatedRows" :key="'r'+ri" :class="{ odd: ri%2===0, even: ri%2!==0 }" @click="handleRowClick(row, ri)">
              <td v-if="showCheckbox" class="gt-col-check"><input type="checkbox" v-model="selectedRows[ri]" /></td>
              <td v-if="showRowNumber" class="gt-col-num">{{ (currentPage-1)*pageSize + ri + 1 }}</td>
              <td v-for="(col, ci) in visibleColumns" :key="'c'+ci" :style="col.width ? { width: col.width, minWidth: col.width } : {}" @click.stop="handleCellClick(row, col.key, $event)">
                <template v-if="col.render">
                  <div v-html="col.render(row, col.key)"></div>
                </template>
                <template v-else>
                  <span :class="{ 'gt-null': isNull(row[col.key]) }">{{ formatVal(row[col.key]) }}</span>
                </template>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      
      <div v-if="showStatusBar" class="gt-footer">
        <div v-if="showStatusLeft" class="gt-footer-left">
          <span>{{ total }} 条记录</span>
        </div>
        <div v-if="showPagination" class="gt-footer-center">
          <button class="gt-pg-btn" @click="firstPage" :disabled="currentPage <= 1" title="首页">
            <svg viewBox="0 0 16 16" width="12" height="12"><path d="M2 4v8h2V4H2zm4 0l4 4-4 4V4z"/></svg>
          </button>
          <button class="gt-pg-btn" @click="prevPage" :disabled="currentPage <= 1" title="上一页">
            <svg viewBox="0 0 16 16" width="12" height="12"><path d="M6 4l-4 4 4 4V4z"/></svg>
          </button>
          <span>第 {{ currentPage }} 页 (共 {{ totalPages }} 页)</span>
          <button class="gt-pg-btn" @click="nextPage" :disabled="currentPage >= totalPages" title="下一页">
            <svg viewBox="0 0 16 16" width="12" height="12"><path d="M10 4l4 4-4 4V4z"/></svg>
          </button>
          <button class="gt-pg-btn" @click="lastPage" :disabled="currentPage >= totalPages" title="末页">
            <svg viewBox="0 0 16 16" width="12" height="12"><path d="M14 4v8h-2V4h2zm-4 0l-4 4 4 4V4z"/></svg>
          </button>
          <select v-model="currentPageSize" class="gt-pg-size">
            <option :value="20">20</option>
            <option :value="50">50</option>
            <option :value="100">100</option>
            <option :value="500">500</option>
          </select>
          <span>条/页</span>
        </div>
        <div v-if="showStatusRight" class="gt-footer-right">
          <span>上次更新: {{ lastUpdated }}</span>
        </div>
      </div>
      
      <div v-if="showColumnPicker" class="gt-col-menu" v-show="columnPickerOpen">
        <div v-for="(col, ci) in columns" :key="'col'+ci" @click="toggleColumn(col.key)">
          <span :class="{ checked: isColVisible(col.key) }">{{ isColVisible(col.key) ? '✓' : '○' }}</span>
          <span>{{ col.label }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'

export interface TableColumn {
  key: string
  label: string
  width?: string
  className?: string
  render?: (row: any, colKey: string) => string
}

const props = withDefaults(defineProps<{
  columns: TableColumn[]
  rows: any[]
  title?: string
  emptyText?: string
  showHeader?: boolean
  showCheckbox?: boolean
  showRowNumber?: boolean
  showPagination?: boolean
  showColumnPicker?: boolean
  showStatusBar?: boolean
  showStatusLeft?: boolean
  showStatusRight?: boolean
  autoHeight?: boolean
  pageSize?: number
}>(), {
  title: '',
  emptyText: '暂无数据',
  showHeader: true,
  showCheckbox: false,
  showRowNumber: true,
  showPagination: true,
  showColumnPicker: true,
  showStatusBar: true,
  showStatusLeft: true,
  showStatusRight: true,
  autoHeight: false,
  pageSize: 50
})

const emit = defineEmits<{
  (e: 'row-click', row: any, index: number): void
  (e: 'cell-click', row: any, colKey: string, event: Event): void
  (e: 'selection-change', selected: any[]): void
  (e: 'page-change', page: number): void
}>()

const currentPage = ref(1)
const currentPageSize = ref(props.pageSize)
const lastUpdated = ref(new Date().toLocaleString('zh-CN'))
const columnPickerOpen = ref(false)
const selectedRows = ref<Record<number, boolean>>({})
const hiddenColumns = ref<Set<string>>(new Set())

const total = computed(() => props.rows.length)
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / currentPageSize.value)))

const visibleColumns = computed(() => {
  return props.columns.filter(col => !hiddenColumns.value.has(col.key))
})

const paginatedRows = computed(() => {
  const start = (currentPage.value - 1) * currentPageSize.value
  return props.rows.slice(start, start + currentPageSize.value)
})

const isAllSelected = computed(() => {
  for (let i = 0; i < paginatedRows.value.length; i++) {
    if (!selectedRows.value[i]) return false
  }
  return paginatedRows.value.length > 0
})

watch(() => props.rows, () => {
  currentPage.value = 1
  lastUpdated.value = new Date().toLocaleString('zh-CN')
  selectedRows.value = {}
}, { deep: true })

watch(() => props.pageSize, () => {
  currentPageSize.value = props.pageSize
  currentPage.value = 1
})

function formatVal(val: any): string {
  if (val === null || val === undefined) return ''
  if (typeof val === 'object') return JSON.stringify(val)
  return String(val)
}

function isNull(val: any): boolean {
  return val === null || val === undefined || val === ''
}

function toggleSelectAll() {
  const isAll = isAllSelected.value
  for (let i = 0; i < paginatedRows.value.length; i++) {
    selectedRows.value[i] = !isAll
  }
  emitSelectionChange()
}

function emitSelectionChange() {
  const selected: any[] = []
  for (const [idx, checked] of Object.entries(selectedRows.value)) {
    if (checked) {
      const actualIdx = (currentPage.value - 1) * props.pageSize + parseInt(idx)
      if (props.rows[actualIdx]) selected.push(props.rows[actualIdx])
    }
  }
  emit('selection-change', selected)
}

function toggleColumnPicker() {
  columnPickerOpen.value = !columnPickerOpen.value
}

function isColVisible(colKey: string): boolean {
  return !hiddenColumns.value.has(colKey)
}

function toggleColumn(colKey: string) {
  if (hiddenColumns.value.has(colKey)) {
    hiddenColumns.value.delete(colKey)
  } else {
    hiddenColumns.value.add(colKey)
  }
}

function firstPage() { currentPage.value = 1; emit('page-change', currentPage.value) }
function prevPage() { currentPage.value = Math.max(1, currentPage.value - 1); emit('page-change', currentPage.value) }
function nextPage() { currentPage.value = Math.min(totalPages.value, currentPage.value + 1); emit('page-change', currentPage.value) }
function lastPage() { currentPage.value = totalPages.value; emit('page-change', currentPage.value) }

function handleRowClick(row: any, index: number) { emit('row-click', row, index) }
function handleCellClick(row: any, colKey: string, event: Event) { emit('cell-click', row, colKey, event) }
</script>

<style scoped>
.gt-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 200px;
  background: #fff;
  border: 1px solid #e8e8e8;
  border-radius: 4px;
}

.gt-container.gt-auto-height {
  flex: 1;
  min-height: 0;
}

.gt-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #999;
  padding: 40px;
}

.gt-empty-icon {
  font-size: 48px;
  margin-bottom: 12px;
}

.gt-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  position: relative;
}

.gt-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: #fafafa;
  border-bottom: 1px solid #e8e8e8;
  flex-shrink: 0;
}

.gt-title {
  font-weight: 600;
  color: #333;
  font-size: 14px;
}

.gt-actions {
  display: flex;
  gap: 8px;
}

.gt-col-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background: transparent;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  color: #666;
}

.gt-col-btn:hover {
  border-color: #409EFF;
  color: #409EFF;
}

.gt-body {
  flex: 1;
  overflow: auto;
  min-height: 0;
}

.gt-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
  table-layout: auto;
  min-width: 100%;
}

.gt-table th,
.gt-table td {
  padding: 10px 16px;
  text-align: left;
  border-bottom: 1px solid #e8e8e8;
  font-size: 14px;
  white-space: nowrap;
  vertical-align: middle;
}

.gt-table th {
  background: #fafafa;
  font-weight: 600;
  color: #606266;
  white-space: nowrap;
  position: sticky;
  top: 0;
  z-index: 1;
}

.gt-table tbody tr:hover {
  background: #f5f7fa;
}

.gt-table tbody tr.odd {
  background: #fff;
}

.gt-table tbody tr.even {
  background: #fafafa;
}

.gt-col-check {
  width: 40px;
  min-width: 40px;
  text-align: center;
  padding: 8px 4px;
}

.gt-col-num {
  width: 50px;
  min-width: 50px;
  text-align: center;
  color: #999;
}

.gt-null {
  color: #bbb;
  font-style: italic;
}

.gt-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: #fafafa;
  border-top: 1px solid #e8e8e8;
  font-size: 12px;
  gap: 12px;
  flex-shrink: 0;
}

.gt-footer-left {
  color: #409EFF;
}

.gt-footer-center {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #67c23a;
}

.gt-footer-right {
  color: #999;
  font-size: 11px;
}

.gt-pg-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 20px;
  background: #fff;
  border: 1px solid #d9d9d9;
  border-radius: 2px;
  cursor: pointer;
}

.gt-pg-btn:hover:not(:disabled) {
  border-color: #409EFF;
  background: #ecf5ff;
}

.gt-pg-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.gt-pg-size {
  padding: 2px 6px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  font-size: 12px;
  outline: none;
}

.gt-col-menu {
  position: absolute;
  right: 12px;
  top: 40px;
  background: #fff;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.1);
  z-index: 100;
  max-height: 200px;
  overflow-y: auto;
}

.gt-col-menu > div {
  padding: 6px 16px;
  font-size: 12px;
  color: #666;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
}

.gt-col-menu > div:hover {
  background: #f5f7fa;
}

.gt-col-menu > div span:first-child {
  font-size: 10px;
  color: #999;
}

.gt-col-menu > div span:first-child.checked {
  color: #67c23a;
}
</style>