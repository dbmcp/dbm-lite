<template>
  <div class="page-container">
    <div class="page-header">
      <div class="page-title">我的待办</div>
      <div class="page-sub">等待我处理的审批任务，支持优先级、状态、截止期筛选。</div>
    </div>

    <div class="stat-row">
      <el-card shadow="never" class="stat-card" v-for="(k, i) in statKeys" :key="i">
        <div class="stat-label">{{ k.label }}</div>
        <div class="stat-value" :style="{ color: k.color }">{{ statMap[k.key] || 0 }}</div>
      </el-card>
    </div>

    <div class="toolbar">
      <el-input v-model="keyword" placeholder="搜索工单标题 / 工单号" clearable style="width: 260px" @keyup.enter="loadList" />
      <el-select v-model="filterStatus" placeholder="处理状态" clearable style="width: 140px">
        <el-option label="待处理" value="pending" />
        <el-option label="处理中" value="processing" />
        <el-option label="已通过" value="approved" />
        <el-option label="已驳回" value="rejected" />
        <el-option label="已跳过" value="skipped" />
      </el-select>
      <el-select v-model="filterPriority" placeholder="优先级" clearable style="width: 120px">
        <el-option label="普通" :value="1" />
        <el-option label="紧急" :value="2" />
        <el-option label="非常紧急" :value="3" />
      </el-select>
      <el-select v-model="filterOverdue" placeholder="截止期" clearable style="width: 140px">
        <el-option label="已逾期" value="overdue" />
        <el-option label="今日到期" value="today" />
        <el-option label="本周内" value="week" />
      </el-select>
      <el-button type="primary" @click="loadList">
        <el-icon><Search /></el-icon>搜索
      </el-button>
      <el-button @click="resetFilters">重置筛选</el-button>
    </div>

    <el-table :data="displayList" border stripe v-loading="loading" style="width: 100%">
      <el-table-column label="优先级" width="80">
        <template #default="{ row }">
          <el-tag v-if="row._priority === 3" type="danger" effect="dark">非常紧急</el-tag>
          <el-tag v-else-if="row._priority === 2" type="warning">紧急</el-tag>
          <el-tag v-else type="info">普通</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="工单标题" min-width="260" show-overflow-tooltip>
        <template #default="{ row }">
          <el-button type="primary" link @click="openDetail(row.workOrderId)">{{ row._title || '-' }}</el-button>
        </template>
      </el-table-column>
      <el-table-column prop="nodeName" label="当前节点" width="180" show-overflow-tooltip />
      <el-table-column prop="assigneeName" label="处理人" width="120" show-overflow-tooltip />
      <el-table-column label="状态" width="120">
        <template #default="{ row }">
          <el-tag :type="row.status === 'pending' ? 'warning' : row.status === 'approved' ? 'success' : 'info'">
            {{ TASK_STATUS_MAP[row.status || ''] || row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="截止期" width="120">
        <template #default="{ row }">
          <span v-if="row._overdue" style="color: #f56c6c; font-weight: 600">已逾期</span>
          <span v-else-if="row._dueDate">{{ row._dueDate }}</span>
          <span v-else style="color: #909399">-</span>
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="创建时间" width="180" />
      <el-table-column label="操作" width="260" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" size="small" link @click="openDetail(row.workOrderId)">查看</el-button>
          <el-button
            v-if="row.status === 'pending'"
            type="success"
            size="small"
            link
            @click="handleTask(row, 'approve')"
          >通过</el-button>
          <el-button
            v-if="row.status === 'pending'"
            type="danger"
            size="small"
            link
            @click="handleTask(row, 'reject')"
          >驳回</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pager">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="size"
        :total="total"
        :page-sizes="[100, 200, 500, 1000]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadList"
        @current-change="loadList"
      />
    </div>

    <!-- 工单详情 -->
    <el-drawer v-model="drawerVisible" title="工单详情" size="600px">
      <div v-if="detailVisible" style="line-height: 1.8">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="工单号">{{ detailWorkOrder?.id }}</el-descriptions-item>
          <el-descriptions-item label="标题">{{ detailWorkOrder?.title }}</el-descriptions-item>
          <el-descriptions-item label="流程">{{ detailWorkOrder?.processName }}</el-descriptions-item>
          <el-descriptions-item label="发起人">{{ detailWorkOrder?.creatorName }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag v-if="detailWorkOrder?.status === 'running'" type="warning">流转中</el-tag>
            <el-tag v-else-if="detailWorkOrder?.status === 'approved'" type="success">已通过</el-tag>
            <el-tag v-else-if="detailWorkOrder?.status === 'rejected'" type="danger">已驳回</el-tag>
            <el-tag v-else>已撤销</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="优先级">{{ PRIORITY_MAP[detailWorkOrder?.priority as number] }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ detailWorkOrder?.createdAt }}</el-descriptions-item>
          <el-descriptions-item label="结束时间">{{ detailWorkOrder?.finishedAt || '-' }}</el-descriptions-item>
        </el-descriptions>

        <div style="margin-top: 16px">
          <div style="font-weight: 600; margin-bottom: 8px">表单数据</div>
          <el-card shadow="never" style="font-size: 13px"><pre style="white-space: pre-wrap; word-break: break-all">{{ prettyJSON(detailWorkOrder?.formData) }}</pre></el-card>
        </div>

        <div style="margin-top: 16px">
          <el-steps :active="currentStepIndex" direction="vertical" finish-status="success" simple>
            <el-step
              v-for="(h, i) in detailHistory"
              :key="i"
              :title="h.action === 'submit' ? '发起申请' : h.action === 'approve' ? '审批通过' : h.action === 'reject' ? '审批驳回' : h.action"
              :description="`${h.operatorName} · ${h.createdAt}${h.remark ? ' · ' + h.remark : ''}`"
            />
          </el-steps>
        </div>

        <div v-if="detailWorkOrder?.status === 'running'" style="margin-top: 16px">
          <el-button type="success" @click="handleTask({ id: activeTask?.id, workOrderId: detailWorkOrder?.id }, 'approve')">通过</el-button>
          <el-button type="danger" @click="handleTask({ id: activeTask?.id, workOrderId: detailWorkOrder?.id }, 'reject')">驳回</el-button>
        </div>
      </div>
    </el-drawer>

    <!-- 处理备注 -->
    <el-dialog v-model="remarkVisible" :title="handleAction === 'approve' ? '通过备注（可选）' : '驳回备注（可选）'" width="420px">
      <el-input v-model="remark" type="textarea" :rows="4" placeholder="请输入处理备注..." />
      <template #footer>
        <el-button @click="remarkVisible = false">取消</el-button>
        <el-button type="primary" :loading="handleSubmitting" @click="submitHandle">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  ferryMyTodo,
  ferryGetWorkOrderDetail,
  ferryHandleTask,
  TASK_STATUS_MAP,
  PRIORITY_MAP
} from '@/api/ferry'

const loading = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const size = ref(100)
const keyword = ref('')
const filterStatus = ref('')
const filterPriority = ref<number | ''>('')
const filterOverdue = ref('')

const statKeys = [
  { key: 'pending', label: '待处理', color: '#e6a23c' },
  { key: 'all', label: '总数', color: '#409eff' },
  { key: 'approved', label: '已通过', color: '#67c23a' },
  { key: 'overdue', label: '已逾期', color: '#f56c6c' }
]
const statMap = ref<Record<string, number>>({ pending: 0, all: 0, approved: 0, overdue: 0 })

function normalize(item: any): any {
  return {
    id: item.id,
    workOrderId: item.workOrderId,
    nodeName: item.nodeName,
    assigneeName: item.assigneeName,
    status: item.status,
    createdAt: item.createdAt,
    _priority: item.priority || 1,
    _title: item.title || (item as any)._title || (item as any).processName || '（未命名工单）',
    _createdAt: item.createdAt
  }
}

async function loadList() {
  loading.value = true
  try {
    const res: any = await ferryMyTodo({
      page: page.value,
      size: size.value,
      keyword: keyword.value,
      status: filterStatus.value as string
    })
    const raw = ((res?.data ?? res) as any[]) || []
    list.value = raw.map(normalize)

    // 额外添加优先级/截止期逻辑（截止期为创建后 7 天）
    list.value.forEach((t) => {
      if (t._createdAt) {
        const d = new Date(t._createdAt.replace(/-/g, '/'))
        if (!isNaN(d.getTime())) {
          d.setDate(d.getDate() + 7)
          const now = new Date()
          t._dueDate = d.toISOString().slice(0, 10)
          t._overdue = d.getTime() < now.getTime()
        } else {
          t._dueDate = ''
          t._overdue = false
        }
      }
    })

    total.value = (res?.total as number) ?? list.value.length

    // 统计：从后台 statistics 中取，若失败则用当前列表估算
    try {
      const s: any = await (await import('@/api/ferry')).ferryStatistics()
      statMap.value = {
        pending: Number(s?.todoPending || 0),
        all: Number(s?.todoTotal || total.value),
        approved: Number(s?.todoApproved || 0),
        overdue: list.value.filter((t) => t._overdue).length
      }
    } catch {
      statMap.value = {
        pending: list.value.filter((t) => t.status === 'pending').length,
        all: total.value,
        approved: list.value.filter((t) => t.status === 'approved').length,
        overdue: list.value.filter((t) => t._overdue).length
      }
    }
  } finally {
    loading.value = false
  }
}

const displayList = computed(() => {
  let out = list.value
  if (filterPriority.value !== '') {
    out = out.filter((t) => t._priority === filterPriority.value)
  }
  if (filterOverdue.value === 'overdue') {
    out = out.filter((t) => t._overdue)
  } else if (filterOverdue.value === 'today') {
    const today = new Date().toISOString().slice(0, 10)
    out = out.filter((t) => t._dueDate === today)
  } else if (filterOverdue.value === 'week') {
    const now = Date.now()
    out = out.filter((t) => {
      if (!t._dueDate) return false
      const d = new Date(t._dueDate).getTime()
      return d >= now && d - now <= 7 * 24 * 3600 * 1000
    })
  }
  return out
})

function resetFilters() {
  keyword.value = ''
  filterStatus.value = ''
  filterPriority.value = ''
  filterOverdue.value = ''
  page.value = 1
  loadList()
}

// ---- 详情 ----
const drawerVisible = ref(false)
const detailVisible = ref(false)
const detailWorkOrder = ref<any>(null)
const detailTasks = ref<any[]>([])
const detailHistory = ref<any[]>([])
const activeTask = ref<any>(null)

async function openDetail(workOrderId: string) {
  drawerVisible.value = true
  detailVisible.value = false
  try {
    const res: any = await ferryGetWorkOrderDetail(workOrderId)
    const data = res?.data || res || {}
    detailWorkOrder.value = data.workOrder
    detailTasks.value = data.tasks || []
    detailHistory.value = (data.histories || []).map((h: any) => ({
      action: h.action,
      remark: h.remark,
      operatorName: h.operatorName,
      createdAt: h.createdAt
    }))
    activeTask.value = detailTasks.value.find((t) => t.status === 'pending') || detailTasks.value[0]
    detailVisible.value = true
  } catch (e: any) {
    ElMessage.error(e?.message || '加载失败')
  }
}

const currentStepIndex = computed(() => {
  const total = detailHistory.value.length
  return total ? total - 1 : 0
})

function prettyJSON(v: any) {
  try {
    if (!v) return '—'
    const obj = typeof v === 'string' ? JSON.parse(v) : v
    return JSON.stringify(obj, null, 2)
  } catch {
    return v
  }
}

// ---- 审批 ----
const remarkVisible = ref(false)
const remark = ref('')
const handleAction = ref<'approve' | 'reject'>('approve')
const pendingTask = ref<any>(null)
const handleSubmitting = ref(false)

function handleTask(row: any, action: 'approve' | 'reject') {
  pendingTask.value = row
  handleAction.value = action
  remark.value = ''
  remarkVisible.value = true
}

async function submitHandle() {
  if (!pendingTask.value) return
  handleSubmitting.value = true
  try {
    const taskId = pendingTask.value.id
    const workOrderId = pendingTask.value.workOrderId || pendingTask.value._workOrderId
    await ferryHandleTask(workOrderId, taskId, handleAction.value, remark.value)
    ElMessage.success(handleAction.value === 'approve' ? '已通过' : '已驳回')
    remarkVisible.value = false
    loadList()
    if (drawerVisible.value) openDetail(workOrderId)
  } catch (e: any) {
    ElMessage.error(e?.message || '处理失败')
  } finally {
    handleSubmitting.value = false
  }
}

onMounted(loadList)
</script>

<style scoped>
.stat-row {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;
}
.stat-card {
  flex: 1;
  padding: 4px 8px;
}
.stat-label {
  color: #909399;
  font-size: 13px;
}
.stat-value {
  font-size: 26px;
  font-weight: 700;
  margin-top: 4px;
}
.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}
.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 12px;
}
.page-header {
  margin-bottom: 12px;
}
.page-title {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}
.page-sub {
  font-size: 13px;
  color: #909399;
  margin-top: 4px;
}
</style>
