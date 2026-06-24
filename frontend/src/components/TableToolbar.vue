<template>
  <div class="table-toolbar" :class="{ 'is-bordered': bordered }">
    <div class="tt-left">
      <span v-if="totalLabel" class="tt-total">{{ totalLabel }}</span>
      <slot name="left"></slot>
    </div>
    <div class="tt-right">
      <el-dropdown
        v-if="columns && columns.length > 0"
        trigger="click"
        @command="onColumnToggle"
        placement="top-end"
      >
        <span class="tt-col-toggle">
          <el-icon><Setting /></el-icon>
          <span>列显示</span>
          <el-icon class="arrow"><ArrowDown /></el-icon>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item
              v-for="col in columns"
              :key="col.prop"
              :command="col.prop"
              :disabled="col.locked"
            >
              <el-icon><Check v-if="visibleCols.has(col.prop)" /><Close v-else style="color:#c0c4cc" /></el-icon>
              <span>{{ col.label }}</span>
            </el-dropdown-item>
            <el-dropdown-item divided @click.native.prevent="() => showAllColumns()">
              <span>显示全部</span>
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
      <span class="tt-size-label">每页</span>
      <el-select
        v-model="localPageSize"
        size="small"
        class="tt-size-select"
        @change="onPageSizeChange"
      >
        <el-option
          v-for="size in pageSizes"
          :key="size"
          :label="size + ' 行'"
          :value="size"
        />
      </el-select>
      <el-pagination
        class="tt-pagination"
        background
        small
        layout="prev, pager, next"
        :total="total"
        :page-size="localPageSize"
        :current-page="localPage"
        @current-change="onPageChange"
      />
      <span class="tt-page-info">
        第 {{ localPage }} / {{ totalPages }} 页 · 共 {{ total }} 条
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Setting, ArrowDown, Check, Close } from '@element-plus/icons-vue'

interface ColumnMeta {
  prop: string
  label: string
  locked?: boolean
}

const props = defineProps<{
  total: number
  pageSize?: number
  page?: number
  pageSizes?: number[]
  columns?: ColumnMeta[]
  visibleColumns?: (string | number)[]
  bordered?: boolean
  totalLabel?: string
}>()

const emit = defineEmits<{
  (e: 'update:pageSize', v: number): void
  (e: 'update:page', v: number): void
  (e: 'change', info: { page: number; pageSize: number }): void
  (e: 'update:visibleColumns', v: (string | number)[]): void
}>()

const defaultSize = props.pageSize ?? 50
const sizes = props.pageSizes && props.pageSizes.length > 0
  ? props.pageSizes
  : ([10, 30, 50, 100] as number[])

const localPageSize = ref<number>(sizes.includes(defaultSize) ? defaultSize : 50)
const localPage = ref<number>(props.page ?? 1)

const visibleCols = ref<Set<string | number>>(new Set(props.visibleColumns ?? columnsToDefault(props.columns)))

const totalPages = computed(() => {
  if (localPageSize.value <= 0) return 1
  return Math.max(1, Math.ceil(props.total / localPageSize.value))
})

watch(
  () => props.page,
  (v) => {
    if (typeof v === 'number' && v >= 1) localPage.value = v
  }
)
watch(
  () => props.pageSize,
  (v) => {
    if (typeof v === 'number' && sizes.includes(v)) localPageSize.value = v
  }
)
watch(
  () => props.visibleColumns,
  (v) => {
    if (Array.isArray(v)) visibleCols.value = new Set(v)
  },
  { immediate: true }
)

function columnsToDefault(cols?: ColumnMeta[]): (string | number)[] {
  if (!cols) return []
  return cols.map((c) => c.prop)
}

function onPageChange(p: number) {
  localPage.value = p
  emit('update:page', p)
  emit('change', { page: p, pageSize: localPageSize.value })
  scrollTableTop()
}

function onPageSizeChange(s: number) {
  localPageSize.value = s
  localPage.value = 1
  emit('update:pageSize', s)
  emit('update:page', 1)
  emit('change', { page: 1, pageSize: s })
  scrollTableTop()
}

function onColumnToggle(prop: string) {
  const next = new Set(visibleCols.value)
  if (next.has(prop)) {
    if (next.size <= 1) {
      ElMessage.warning('至少保留一列')
      return
    }
    next.delete(prop)
  } else {
    next.add(prop)
  }
  visibleCols.value = next
  emit('update:visibleColumns', Array.from(next))
}

function showAllColumns() {
  const next = new Set<string | number>()
  if (props.columns) {
    for (const c of props.columns) next.add(c.prop)
  }
  visibleCols.value = next
  emit('update:visibleColumns', Array.from(next))
}

function scrollTableTop() {
  try {
    const scroller = document.querySelector('.el-table__body-wrapper')
    if (scroller && scroller.scrollTo) scroller.scrollTo({ top: 0, behavior: 'smooth' })
  } catch (_) {
    // 忽略滚动错误
  }
}
</script>

<style scoped>
.table-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 12px;
  background: #fafafa;
  border-bottom: 1px solid #ebeef5;
  flex-wrap: wrap;
}
.table-toolbar.is-bordered {
  border: 1px solid #ebeef5;
  border-bottom: none;
}
.tt-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.tt-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.tt-total {
  color: #606266;
  font-size: 13px;
}
.tt-col-toggle {
  cursor: pointer;
  color: #606266;
  font-size: 13px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 4px;
  border: 1px solid #dcdfe6;
  background: #fff;
  user-select: none;
}
.tt-col-toggle:hover {
  color: #409eff;
  border-color: #c6e2ff;
}
.tt-col-toggle .arrow {
  font-size: 12px;
}
.tt-size-label {
  color: #606266;
  font-size: 13px;
}
.tt-size-select {
  width: 90px;
}
.tt-page-info {
  color: #909399;
  font-size: 12px;
}
</style>
