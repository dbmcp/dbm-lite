<template>
  <div class="page-container">
    <div class="page-header">
      <div class="page-title">全部工单</div>
      <div class="page-sub">管理平台内所有审批工单（管理员视图）</div>
    </div>

    <div class="stat-row">
      <el-card shadow="never" class="stat-card" v-for="(k, i) in statKeys" :key="i">
        <div class="stat-label">{{ k.label }}</div>
        <div class="stat-value" :style="{ color: k.color }">{{ statMap[k.key] || 0 }}</div>
      </el-card>
    </div>

    <div class="toolbar">
      <el-input v-model="keyword" placeholder="搜索工单标题 / 流程名称" clearable style="width: 260px" @keyup.enter="loadList">
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
      <el-input v-model="creatorKeyword" placeholder="搜索创建人" clearable style="width: 180px" @keyup.enter="loadList" />
      <el-select v-model="filterStatus" placeholder="状态" clearable style="width: 130px">
        <el-option label="流转中" value="running" />
        <el-option label="已通过" value="approved" />
        <el-option label="已驳回" value="rejected" />
        <el-option label="已撤销" value="canceled" />
      </el-select>
      <el-select v-model="filterPriority" placeholder="优先级" clearable style="width: 120px">
        <el-option label="普通" :value="1" />
        <el-option label="紧急" :value="2" />
        <el-option label="非常紧急" :value="3" />
      </el-select>
      <el-button type="primary" @click="loadList">
        <el-icon><Search /></el-icon>搜索
      </el-button>
      <el-button @click="resetFilters">重置筛选</el-button>
    </div>

    <el-table
      :data="displayList"
      border
      stripe
      v-loading="loading"
      style="width: 100%"
      @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="48" />
      <el-table-column prop="id" label="工单号" width="220" show-overflow-tooltip />
      <el-table-column label="优先级" width="96">
        <template #default="{ row }">
          <el-tag v-if="(row.priority || 1) === 3" type="danger" effect="dark" size="small">非常紧急</el-tag>
          <el-tag v-else-if="(row.priority || 1) === 2" type="warning" size="small">紧急</el-tag>
          <el-tag v-else type="info" size="small">普通</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="title" label="标题" min-width="260" show-overflow-tooltip />
      <el-table-column prop="processName" label="流程名称" width="180" show-overflow-tooltip />
      <el-table-column prop="creatorName" label="创建人" width="120" show-overflow-tooltip />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)" size="small">{{ STATUS_MAP[row.status || ''] || row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="创建时间" width="180" show-overflow-tooltip />
      <el-table-column prop="finishedAt" label="完成时间" width="180" show-overflow-tooltip>
        <template #default="{ row }">
          <span v-if="row.finishedAt">{{ row.finishedAt }}</span>
          <span v-else style="color: #909399">-</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="260" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" size="small" link @click="openDetail(row.id)">详情</el-button>
          <el-button
            v-if="row.status === 'running'"
            type="info"
            size="small"
            link
            @click="confirmUrge(row.id)"
          >催办</el-button>
          <el-button
            v-if="row.status === 'running'"
            type="danger"
            size="small"
            link
            @click="confirmRevoke(row.id)"
          >撤销</el-button>
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

    <!-- 工单详情 Drawer -->
    <el-drawer v-model="drawerVisible" title="工单详情" size="720px">
      <div v-if="detailVisible" style="line-height: 1.8">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="工单号">{{ detailWorkOrder?.id }}</el-descriptions-item>
          <el-descriptions-item label="流程名称">{{ detailWorkOrder?.processName }}</el-descriptions-item>
          <el-descriptions-item label="标题">{{ detailWorkOrder?.title }}</el-descriptions-item>
          <el-descriptions-item label="优先级">{{ PRIORITY_MAP[(detailWorkOrder?.priority as number) || 1] }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTagType(detailWorkOrder?.status)" size="small">
              {{ STATUS_MAP[detailWorkOrder?.status || ''] || detailWorkOrder?.status }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建人">{{ detailWorkOrder?.creatorName }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ detailWorkOrder?.createdAt }}</el-descriptions-item>
          <el-descriptions-item label="完成时间">{{ detailWorkOrder?.finishedAt || '-' }}</el-descriptions-item>
        </el-descriptions>

        <div class="drawer-section">
          <div class="drawer-section-title">表单数据</div>
          <el-card v-if="detailWorkOrder?.formData" shadow="never" class="form-card"><pre>{{ prettyJSON(detailWorkOrder.formData) }}</pre></el-card>
          <el-card v-else shadow="never" class="form-card" style="color: #909399">无</el-card>
        </div>

        <div class="drawer-section">
          <div class="drawer-section-title">审批历史</div>
          <el-timeline>
            <el-timeline-item
              v-for="(h, idx) in (detailHistory || []).slice().reverse()"
              :key="h.id || idx"
              :timestamp="h.createdAt"
              :type="historyTimelineType(h.action)"
              placement="top"
            >
              <div style="font-weight: 600">{{ actionLabel(h.action) }} · {{ h.operatorName || '-' }}</div>
              <div v-if="h.nodeName" style="color: #606266; font-size: 13px">节点：{{ h.nodeName }}</div>
              <div v-if="h.remark" style="color: #606266; font-size: 13px">{{ h.remark }}</div>
            </el-timeline-item>
          </el-timeline>
        </div>

        <div v-if="detailWorkOrder?.status === 'running'" class="drawer-section">
          <el-button type="info" @click="confirmUrge(detailWorkOrder!.id)">催办</el-button>
          <el-button type="danger" @click="confirmRevoke(detailWorkOrder!.id)">撤销</el-button>
        </div>
      </div>
    </el-drawer>

    <!-- 催办 Dialog -->
    <el-dialog v-model="urgeVisible" title="催办" width="480px">
      <el-form label-width="80px">
        <el-form-item label="工单号">
          <span>{{ urgeWorkOrderId }}</span>
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="urgeRemark" type="textarea" :rows="4" placeholder="请输入催办说明（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="urgeVisible = false">取消</el-button>
        <el-button type="primary" :loading="urgeSubmitting" @click="submitUrge">确认催办</el-button>
      </template>
    </el-dialog>

    <!-- 撤销 Dialog -->
    <el-dialog v-model="revokeVisible" title="撤销工单" width="480px">
      <el-form label-width="80px">
        <el-form-item label="工单号">
          <span>{{ revokeWorkOrderId }}</span>
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="revokeRemark" type="textarea" :rows="4" placeholder="请输入撤销说明（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="revokeVisible = false">取消</el-button>
        <el-button type="danger" :loading="revokeSubmitting" @click="submitRevoke">确认撤销</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  ferryAllWorkOrders,
  ferryGetWorkOrderDetail,
  ferryUrgeWorkOrder,
  ferryRevokeWorkOrder,
  STATUS_MAP,
  PRIORITY_MAP,
  WorkOrder,
  WorkOrderHistory
} from '@/api/ferry'

const loading = ref(false)
const list = ref<WorkOrder[]>([])
const total = ref(0)
const page = ref(1)
const size = ref(100)
const keyword = ref('')
const creatorKeyword = ref('')
const filterStatus = ref('')
const filterPriority = ref<number | ''>('')
const selectedRows = ref<WorkOrder[]>([])

const statKeys = [
  { key: 'total', label: '总数', color: '#409eff' },
  { key: 'running', label: '流转中', color: '#e6a23c' },
  { key: 'approved', label: '已通过', color: '#67c23a' },
  { key: 'rejected', label: '已驳回', color: '#f56c6c' },
  { key: 'canceled', label: '撤销', color: '#909399' }
]
const statMap = ref<Record<string, number>>({ total: 0, running: 0, approved: 0, rejected: 0, canceled: 0 })

function matchesLocal(item: WorkOrder): boolean {
  if (filterPriority.value !== '' && (item.priority || 1) !== filterPriority.value) return false
  if (creatorKeyword.value.trim()) {
    const kw = creatorKeyword.value.trim().toLowerCase()
    const name = (item.creatorName || '').toLowerCase()
    if (!name.includes(kw)) return false
  }
  return true
}

const displayList = computed(() => list.value.filter(matchesLocal))

async function loadList() {
  loading.value = true
  try {
    const res = await ferryAllWorkOrders({
      page: page.value,
      size: size.value,
      keyword: keyword.value,
      status: filterStatus.value
    })
    list.value = ((res?.data as WorkOrder[]) || [])
    total.value = (res?.total as number) || list.value.length

    statMap.value = {
      total: total.value,
      running: list.value.filter((t) => t.status === 'running').length,
      approved: list.value.filter((t) => t.status === 'approved').length,
      rejected: list.value.filter((t) => t.status === 'rejected').length,
      canceled: list.value.filter((t) => t.status === 'canceled').length
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '加载失败')
  } finally {
    loading.value = false
  }
}

function resetFilters() {
  keyword.value = ''
  creatorKeyword.value = ''
  filterStatus.value = ''
  filterPriority.value = ''
  page.value = 1
  loadList()
}

function handleSelectionChange(rows: WorkOrder[]) {
  selectedRows.value = rows
}

function statusTagType(status?: string) {
  switch (status) {
    case 'running':
      return 'warning'
    case 'approved':
      return 'success'
    case 'rejected':
      return 'danger'
    case 'canceled':
      return 'info'
    default:
      return ''
  }
}

function historyTimelineType(action?: string) {
  switch (action) {
    case 'submit':
    case 'approve':
      return 'success'
    case 'reject':
    case 'revoke':
      return 'danger'
    case 'urge':
      return 'warning'
    default:
      return 'primary'
  }
}

function actionLabel(action?: string) {
  const map: Record<string, string> = {
    submit: '发起申请',
    approve: '通过',
    reject: '驳回',
    revoke: '撤销',
    urge: '催办',
    transfer: '转交',
    addsign: '加签'
  }
  return map[action || ''] || action || '操作'
}

function prettyJSON(data: any) {
  try {
    if (!data) return '—'
    const obj = typeof data === 'string' ? JSON.parse(data) : data
    return JSON.stringify(obj, null, 2)
  } catch {
    return String(data)
  }
}

// ============ 详情 ============
const drawerVisible = ref(false)
const detailVisible = ref(false)
const detailWorkOrder = ref<WorkOrder | null>(null)
const detailHistory = ref<WorkOrderHistory[]>([])

async function openDetail(id: string) {
  drawerVisible.value = true
  detailVisible.value = false
  try {
    const res = await ferryGetWorkOrderDetail(id)
    const data: any = res?.data || res || {}
    detailWorkOrder.value = data.workOrder || null
    detailHistory.value = (data.histories || []) as WorkOrderHistory[]
    detailVisible.value = true
  } catch (e: any) {
    ElMessage.error(e?.message || '加载详情失败')
  }
}

// ============ 催办 ============
const urgeVisible = ref(false)
const urgeWorkOrderId = ref('')
const urgeRemark = ref('')
const urgeSubmitting = ref(false)

function confirmUrge(id: string) {
  urgeWorkOrderId.value = id
  urgeRemark.value = ''
  urgeVisible.value = true
}

async function submitUrge() {
  if (!urgeWorkOrderId.value) return
  urgeSubmitting.value = true
  try {
    await ferryUrgeWorkOrder(urgeWorkOrderId.value, urgeRemark.value)
    ElMessage.success('已催办')
    urgeVisible.value = false
    loadList()
    if (drawerVisible.value && detailWorkOrder.value?.id === urgeWorkOrderId.value) {
      openDetail(urgeWorkOrderId.value)
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '催办失败')
  } finally {
    urgeSubmitting.value = false
  }
}

// ============ 撤销 ============
const revokeVisible = ref(false)
const revokeWorkOrderId = ref('')
const revokeRemark = ref('')
const revokeSubmitting = ref(false)

function confirmRevoke(id: string) {
  revokeWorkOrderId.value = id
  revokeRemark.value = ''
  revokeVisible.value = true
}

async function submitRevoke() {
  if (!revokeWorkOrderId.value) return
  revokeSubmitting.value = true
  try {
    await ferryRevokeWorkOrder(revokeWorkOrderId.value, revokeRemark.value)
    ElMessage.success('已撤销')
    revokeVisible.value = false
    loadList()
    if (drawerVisible.value && detailWorkOrder.value?.id === revokeWorkOrderId.value) {
      openDetail(revokeWorkOrderId.value)
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '撤销失败')
  } finally {
    revokeSubmitting.value = false
  }
}

onMounted(loadList)
</script>

<style scoped>
.page-container {
  padding: 16px;
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
.drawer-section {
  margin-top: 20px;
}
.drawer-section-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 8px;
}
.form-card {
  font-size: 12px;
  line-height: 1.6;
}
.form-card :deep(pre) {
  background: #f5f7fa;
  border-radius: 4px;
  padding: 10px;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 300px;
  overflow: auto;
  margin: 0;
}
</style>
