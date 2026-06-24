<template>
  <div class="ferry-page">
    <div class="ferry-header">
      <div>
        <h2 class="ferry-title">
          <el-icon :size="22" style="vertical-align: middle; margin-right: 8px"><Menu /></el-icon>
          流程分类
        </h2>
        <div class="ferry-subtitle">对流程进行分类管理，支持多级层级与报告视图。</div>
      </div>
    </div>

    <div class="ferry-stats">
      <div class="stat-box stat-total">
        <div class="stat-label">分类总数</div>
        <div class="stat-value">{{ stats.classifyTotal }}</div>
      </div>
      <div class="stat-box stat-enabled">
        <div class="stat-label">启用流程数</div>
        <div class="stat-value">{{ stats.enabledProcesses }}</div>
      </div>
      <div class="stat-box stat-disabled">
        <div class="stat-label">禁用流程数</div>
        <div class="stat-value">{{ stats.disabledProcesses }}</div>
      </div>
      <div class="stat-box stat-submits">
        <div class="stat-label">流程提交总数</div>
        <div class="stat-value">{{ stats.submitTotal }}</div>
      </div>
    </div>

    <el-tabs v-model="activeTab" class="ferry-tabs">
      <el-tab-pane label="列表视图" name="list">
        <div class="ferry-toolbar">
          <el-button type="success" @click="openCreate">
            <el-icon><Plus /></el-icon>新建分类
          </el-button>
          <div class="ferry-spacer" ></div>
          <el-button link @click="loadAll">
<span class="refresh-icon">⟳</span>刷新
          </el-button>
        </div>

        <el-table
          :data="treeData"
          v-loading="loading"
          row-key="id"
          border
          stripe
          default-expand-all
          :tree-props="{ children: 'children' }"
        >
          <el-table-column prop="name" label="名称" min-width="220" />
          <el-table-column label="父分类" width="180">
            <template #default="{ row }">{{ row.parentId ? parentNameMap[row.parentId] || '-' : '-' }}</template>
          </el-table-column>
          <el-table-column prop="sortOrder" label="排序" width="80" align="center" />
          <el-table-column prop="createdAt" label="创建时间" width="180" />
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
              <el-button size="small" type="primary" link @click="openEdit(row)">编辑</el-button>
              <el-button size="small" type="danger" link @click="doDelete(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="报告视图" name="report">
        <div class="report-grid">
          <div class="report-card">
            <div class="report-title">各分类流程数量</div>
            <div class="bar-chart">
              <div v-if="maxClassifyCount === 0" class="empty-tip">暂无数据</div>
              <div
                v-for="item in classifyProcessCounts"
                :key="item.id"
                class="bar-row"
              >
                <div class="bar-label" :title="item.name">{{ item.name }}</div>
                <div class="bar-track">
                  <div
                    class="bar-fill"
                    :style="{ width: barWidth(item.count, maxClassifyCount) + '%' }"
                  ></div>
                </div>
                <div class="bar-value">{{ item.count }}</div>
              </div>
            </div>
          </div>

          <div class="report-card">
            <div class="report-title">启用 / 禁用分布</div>
            <div class="pie-wrap">
              <div class="donut" :style="donutStyle"></div>
              <div class="pie-center">
                <div class="pie-total">{{ processList.length }}</div>
                <div class="pie-total-label">流程总数</div>
              </div>
            </div>
            <div class="pie-legend">
              <div class="legend-item">
                <span class="legend-dot legend-enabled" ></span>
                启用 {{ stats.enabledProcesses }}
                <span class="legend-pct">{{ pct(stats.enabledProcesses, processList.length) }}%</span>
              </div>
              <div class="legend-item">
                <span class="legend-dot legend-disabled" ></span>
                禁用 {{ stats.disabledProcesses }}
                <span class="legend-pct">{{ pct(stats.disabledProcesses, processList.length) }}%</span>
              </div>
            </div>
          </div>
        </div>

        <div class="report-card" style="margin-top: 12px">
          <div class="report-title">分类汇总</div>
          <el-table :data="classifySummary" border stripe size="small">
            <el-table-column prop="name" label="分类名称" min-width="220" />
            <el-table-column prop="processCount" label="流程数" width="120" align="center" />
            <el-table-column prop="submitCount" label="提交总数" width="120" align="center" />
          </el-table>
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="formVisible" :title="formMode === 'create' ? '新建分类' : '编辑分类'" width="560px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="请输入分类名称" />
        </el-form-item>
        <el-form-item label="父分类">
          <el-select v-model="form.parentId" placeholder="— 无（顶级）—" clearable style="width: 100%">
            <el-option
              v-for="c in parentOptions"
              :key="c.id"
              :label="c.name"
              :value="c.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sortOrder" :min="0" :max="9999" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="doSubmit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Menu } from '@element-plus/icons-vue'
import {
  ferryListClassifies,
  ferryCreateClassify,
  ferryUpdateClassify,
  ferryDeleteClassify,
  ferryListProcesses,
  type Classify,
  type Process
} from '@/api/ferry'

const list = ref<Classify[]>([])
const processList = ref<Process[]>([])
const loading = ref(false)
const activeTab = ref<'list' | 'report'>('list')
const formVisible = ref(false)
const formMode = ref<'create' | 'edit'>('create')
const form = ref<{ id: string; name: string; parentId: string; sortOrder: number }>({
  id: '',
  name: '',
  parentId: '',
  sortOrder: 0
})
const submitting = ref(false)

const parentNameMap = computed<Record<string, string>>(() => {
  const m: Record<string, string> = {}
  for (const c of list.value) {
    if (c.id && c.name) m[c.id] = c.name
  }
  return m
})

const parentOptions = computed<Classify[]>(() => {
  if (formMode.value === 'edit' && form.value.id) {
    return list.value.filter((c) => c.id !== form.value.id)
  }
  return list.value
})

const treeData = computed<(Classify & { children?: Classify[] })[]>(() => {
  const byParent: Record<string, Classify[]> = {}
  const roots: (Classify & { children?: Classify[] })[] = []
  for (const c of list.value) {
    const pid = (c as any).parentId
    if (pid && parentNameMap.value[pid]) {
      if (!byParent[pid]) byParent[pid] = []
      byParent[pid].push(c)
    } else {
      roots.push(c as any)
    }
  }
  const sorter = (a: Classify, b: Classify) => (a.sortOrder ?? 0) - (b.sortOrder ?? 0)
  roots.sort(sorter)
  for (const r of roots) {
    const children = byParent[r.id]
    if (children && children.length) {
      ;(r as any).children = children.sort(sorter) as any
    }
  }
  return roots
})

const stats = computed(() => {
  let enabledProcesses = 0
  let disabledProcesses = 0
  let submitTotal = 0
  for (const p of processList.value) {
    if (p.enabled) {
      enabledProcesses++
    } else {
      disabledProcesses++
    }
    submitTotal += Number((p as any).submitCnt || 0) || 0
  }
  return {
    classifyTotal: list.value.length,
    enabledProcesses,
    disabledProcesses,
    submitTotal
  }
})

const classifyProcessCounts = computed<{ id: string; name: string; count: number }[]>(() => {
  const counter: Record<string, { name: string; count: number }> = {}
  for (const c of list.value) {
    counter[c.id] = { name: c.name, count: 0 }
  }
  for (const p of processList.value) {
    const classifyId = (p as any).classifyId || (p as any).classify
    if (classifyId && counter[classifyId]) {
      counter[classifyId].count++
    }
  }
  const arr = Object.keys(counter).map((id) => ({
    id,
    name: counter[id].name,
    count: counter[id].count
  }))
  arr.sort((a, b) => b.count - a.count)
  return arr
})

const maxClassifyCount = computed(
  () => classifyProcessCounts.value.reduce((m, i) => Math.max(m, i.count), 0)
)

const classifySummary = computed<{ name: string; processCount: number; submitCount: number }[]>(() => {
  const summary: Record<string, { name: string; processCount: number; submitCount: number }> = {}
  for (const c of list.value) {
    summary[c.id] = { name: c.name, processCount: 0, submitCount: 0 }
  }
  for (const p of processList.value) {
    const classifyId = (p as any).classifyId || (p as any).classify
    if (classifyId && summary[classifyId]) {
      summary[classifyId].processCount++
      summary[classifyId].submitCount += Number((p as any).submitCnt || 0) || 0
    }
  }
  const arr = Object.values(summary)
  arr.sort((a, b) => b.processCount - a.processCount)
  return arr
})

const donutStyle = computed(() => {
  const total = processList.value.length
  if (total === 0) {
    return { background: '#f0f2f5' }
  }
  const enabledPct = (stats.value.enabledProcesses / total) * 100
  return {
    background: `conic-gradient(#67c23a 0% ${enabledPct}%, #909399 ${enabledPct}% 100%)`
  }
})

function barWidth(value: number, max: number) {
  if (max <= 0) return 0
  return Math.max(2, (value / max) * 100)
}

function pct(part: number, total: number) {
  if (total === 0) return 0
  return Math.round((part / total) * 100)
}

async function loadAll() {
  loading.value = true
  try {
    const [classifyRes, processRes] = await Promise.all([
      ferryListClassifies(),
      ferryListProcesses({ page: 1, size: 9999 })
    ])
    list.value = (classifyRes?.data as Classify[]) || []
    processList.value = (processRes?.data as Process[]) || []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  form.value = { id: '', name: '', parentId: '', sortOrder: 0 }
  formMode.value = 'create'
  formVisible.value = true
}

function openEdit(row: Classify) {
  form.value = {
    id: row.id,
    name: row.name || '',
    parentId: (row as any).parentId || '',
    sortOrder: (row as any).sortOrder ?? 0
  }
  formMode.value = 'edit'
  formVisible.value = true
}

async function doSubmit() {
  if (!form.value.name?.trim()) {
    ElMessage.warning('请输入分类名称')
    return
  }
  submitting.value = true
  try {
    if (formMode.value === 'create') {
      await ferryCreateClassify({
        name: form.value.name,
        parentId: form.value.parentId || undefined,
        sortOrder: form.value.sortOrder
      })
      ElMessage.success('创建成功')
    } else {
      await ferryUpdateClassify(form.value.id, {
        name: form.value.name,
        parentId: form.value.parentId || undefined,
        sortOrder: form.value.sortOrder
      })
      ElMessage.success('更新成功')
    }
    formVisible.value = false
    loadAll()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  } finally {
    submitting.value = false
  }
}

function doDelete(row: Classify) {
  ElMessageBox.confirm(`确认删除分类 "${row.name}" 吗？`, '提示', { type: 'warning' })
    .then(async () => {
      await ferryDeleteClassify(row.id)
      ElMessage.success('删除成功')
      loadAll()
    })
    .catch(() => {})
}

onMounted(loadAll)
</script>

<style scoped>
.ferry-page {
  padding: 16px;
}

.ferry-header {
  background: linear-gradient(90deg, #ecf5ff 0%, #f0f9eb 100%);
  border-radius: 8px;
  padding: 18px 24px;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
}

.ferry-title {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
  margin: 0;
}

.ferry-subtitle {
  font-size: 13px;
  color: #606266;
  margin-top: 4px;
}

.ferry-stats {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.stat-box {
  background: #fff;
  border-radius: 6px;
  padding: 12px 22px;
  min-width: 130px;
  text-align: center;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
  flex: 1;
}

.stat-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 6px;
}

.stat-value {
  font-size: 22px;
  font-weight: 600;
  color: #409eff;
}

.stat-enabled .stat-value {
  color: #67c23a;
}

.stat-disabled .stat-value {
  color: #909399;
}

.stat-submits .stat-value {
  color: #e6a23c;
}

.ferry-tabs {
  background: #fff;
  padding: 8px 16px 16px;
  border-radius: 8px;
}

.ferry-toolbar {
  display: flex;
  gap: 8px;
  margin: 12px 0;
  align-items: center;
  flex-wrap: wrap;
}

.ferry-spacer {
  flex: 1;
}

.report-grid {
  display: grid;
  grid-template-columns: 1.3fr 1fr;
  gap: 12px;
}

@media (max-width: 900px) {
  .report-grid {
    grid-template-columns: 1fr;
  }
}

.report-card {
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 6px;
  padding: 16px 18px;
}

.report-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px dashed #ebeef5;
}

.bar-chart {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 360px;
  overflow-y: auto;
  padding-right: 4px;
}

.empty-tip {
  color: #c0c4cc;
  font-size: 13px;
  text-align: center;
  padding: 30px 0;
}

.bar-row {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
}

.bar-label {
  width: 120px;
  color: #606266;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex-shrink: 0;
}

.bar-track {
  flex: 1;
  height: 18px;
  background: #f5f7fa;
  border-radius: 3px;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  background: linear-gradient(90deg, #409eff, #66b1ff);
  border-radius: 3px;
  transition: width 0.4s ease;
}

.bar-value {
  width: 48px;
  text-align: right;
  color: #303133;
  font-weight: 600;
  flex-shrink: 0;
}

.pie-wrap {
  position: relative;
  display: flex;
  justify-content: center;
  padding: 10px 0;
}

.donut {
  width: 160px;
  height: 160px;
  border-radius: 50%;
  position: relative;
}

.donut::after {
  content: '';
  position: absolute;
  top: 30px;
  left: 30px;
  right: 30px;
  bottom: 30px;
  background: #fff;
  border-radius: 50%;
}

.pie-center {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
}

.pie-total {
  font-size: 22px;
  font-weight: 700;
  color: #303133;
}

.pie-total-label {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}

.pie-legend {
  display: flex;
  justify-content: center;
  gap: 22px;
  margin-top: 8px;
  font-size: 13px;
  color: #606266;
}

.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.legend-dot {
  display: inline-block;
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.legend-enabled {
  background: #67c23a;
}

.legend-disabled {
  background: #909399;
}

.legend-pct {
  color: #909399;
  font-size: 12px;
  margin-left: 4px;
}
</style>
