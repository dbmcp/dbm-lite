<template>
  <div class="page-container">
    <div class="page-header">
      <div class="page-title">我的申请</div>
      <div class="page-sub">本人提交工单的全生命周期追踪，支持撤回与催办操作。</div>
    </div>

    <div class="stat-row">
      <el-card shadow="never" class="stat-card" v-for="(k, i) in statKeys" :key="i">
        <div class="stat-label">{{ k.label }}</div>
        <div class="stat-value" :style="{ color: k.color }">{{ statMap[k.key] || 0 }}</div>
      </el-card>
    </div>

    <div class="toolbar">
      <el-input v-model="keyword" placeholder="搜索工单标题 / 工单号" clearable style="width: 260px" @keyup.enter="loadList" />
      <el-select v-model="filterStatus" placeholder="审批状态" clearable style="width: 140px">
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
      <div class="ferry-spacer"></div>
      <el-button type="success" @click="openNewApply">
        <el-icon><Plus /></el-icon>发起申请
      </el-button>
    </div>

    <el-table :data="displayList" border stripe v-loading="loading" style="width: 100%">
      <el-table-column label="优先级" width="80">
        <template #default="{ row }">
          <el-tag v-if="row._priority === 3" type="danger" effect="dark">非常紧急</el-tag>
          <el-tag v-else-if="row._priority === 2" type="warning">紧急</el-tag>
          <el-tag v-else type="info">普通</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="工单号" width="220">
        <template #default="{ row }">{{ row.id }}</template>
      </el-table-column>
      <el-table-column label="工单标题" min-width="260" show-overflow-tooltip>
        <template #default="{ row }">
          <el-button type="primary" link @click="openDetail(row.id)">{{ row.title || '-' }}</el-button>
        </template>
      </el-table-column>
      <el-table-column prop="processName" label="所属流程" width="200" show-overflow-tooltip />
      <el-table-column label="当前节点" width="180" show-overflow-tooltip>
        <template #default="{ row }">{{ row._currentNode || '-' }}</template>
      </el-table-column>
      <el-table-column label="审批状态" width="120">
        <template #default="{ row }">
          <el-tag :type="row.status === 'running' ? 'warning' : row.status === 'approved' ? 'success' : row.status === 'rejected' ? 'danger' : 'info'">
            {{ STATUS_MAP[row.status || ''] || row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="提交时间" width="180" />
      <el-table-column label="操作" width="260" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" size="small" link @click="openDetail(row.id)">详情</el-button>
          <el-button v-if="row.status === 'running'" type="warning" size="small" link @click="urgeApply(row)">催办</el-button>
          <el-button v-if="row.status === 'running'" type="danger" size="small" link @click="revokeApply(row)">撤回</el-button>
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
            <el-step v-for="(h, i) in detailHistory" :key="i"
              :title="h.action === 'submit' ? '发起申请' : h.action === 'approve' ? '审批通过' : h.action === 'reject' ? '审批驳回' : h.action === 'revoke' ? '撤回申请' : h.action"
              :description="h.operatorName + ' - ' + h.createdAt + (h.remark ? ' - ' + h.remark : '')"
            />
          </el-steps>
        </div>

        <div style="margin-top: 16px">
          <el-button v-if="detailWorkOrder?.status === 'running'" type="warning" @click="urgeApply({ id: detailWorkOrder?.id })">催办</el-button>
          <el-button v-if="detailWorkOrder?.status === 'running'" type="danger" @click="revokeApply({ id: detailWorkOrder?.id })">撤回</el-button>
        </div>
      </div>
    </el-drawer>

    <el-dialog v-model="revokeVisible" title="撤回备注（可选）" width="420px">
      <el-input v-model="revokeRemark" type="textarea" :rows="4" placeholder="请输入撤回备注..." />
      <template #footer>
        <el-button @click="revokeVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitRevoke">确认撤回</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="urgeVisible" title="催办备注（可选）" width="420px">
      <el-input v-model="urgeRemark" type="textarea" :rows="4" placeholder="请输入催办备注..." />
      <template #footer>
        <el-button @click="urgeVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitUrge">确认催办</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="newApplyVisible" title="发起新申请" width="560px">
      <el-form label-width="100px">
        <el-form-item label="选择流程">
          <el-select v-model="newApplyProcessId" placeholder="请选择要发起的审批流程" filterable style="width: 100%">
            <el-option v-for="p in enabledProcesses" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="工单标题">
          <el-input v-model="newApplyTitle" placeholder="请输入工单标题" />
        </el-form-item>
        <el-form-item label="优先级">
          <el-radio-group v-model="newApplyPriority">
            <el-radio :value="1">普通</el-radio>
            <el-radio :value="2">紧急</el-radio>
            <el-radio :value="3">非常紧急</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="申请内容">
          <el-input v-model="newApplyContent" type="textarea" :rows="4" placeholder="请输入申请内容说明..." />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="newApplyVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitNewApply">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ferryMyApply,
  ferryGetWorkOrderDetail,
  ferryRevokeWorkOrder,
  ferryUrgeWorkOrder,
  ferrySubmitWorkOrder,
  ferryListEnabledProcesses,
  STATUS_MAP,
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

const statKeys = [
  { key: 'total', label: '全部工单', color: '#409eff' },
  { key: 'running', label: '流转中', color: '#e6a23c' },
  { key: 'approved', label: '已通过', color: '#67c23a' },
  { key: 'rejected', label: '已驳回', color: '#f56c6c' },
  { key: 'canceled', label: '已撤销', color: '#909399' }
]
const statMap = ref<Record<string, number>>({ total: 0, running: 0, approved: 0, rejected: 0, canceled: 0 })

const enabledProcesses = ref<any[]>([])

function normalize(item: any): any {
  return {
    ...item,
    _priority: item.priority || 1,
    _currentNode: item._currentNode || item.currentNode || '-'
  }
}

async function loadEnabledProcesses() {
  try {
    const res: any = await ferryListEnabledProcesses()
    enabledProcesses.value = (res?.data || res || []) as any[]
  } catch {}
}

async function loadList() {
  loading.value = true
  try {
    const res: any = await ferryMyApply({
      page: page.value,
      size: size.value,
      keyword: keyword.value,
      status: filterStatus.value as string
    })
    const raw = ((res?.data ?? res) as any[]) || []
    list.value = raw.map(normalize)
    total.value = (res?.total as number) ?? list.value.length

    try {
      const s: any = await (await import('@/api/ferry')).ferryStatistics()
      statMap.value = {
        total: Number(s?.applyTotal || total.value),
        running: Number(s?.applyRunning || 0),
        approved: Number(s?.applyApproved || 0),
        rejected: Number(s?.applyRejected || 0),
        canceled: Number(s?.applyCanceled || 0)
      }
    } catch {
      statMap.value = {
        total: total.value,
        running: list.value.filter((t) => t.status === 'running').length,
        approved: list.value.filter((t) => t.status === 'approved').length,
        rejected: list.value.filter((t) => t.status === 'rejected').length,
        canceled: list.value.filter((t) => t.status === 'canceled').length
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
  return out
})

function resetFilters() {
  keyword.value = ''
  filterStatus.value = ''
  filterPriority.value = ''
  page.value = 1
  loadList()
}

const drawerVisible = ref(false)
const detailVisible = ref(false)
const detailWorkOrder = ref<any>(null)
const detailTasks = ref<any[]>([])
const detailHistory = ref<any[]>([])

async function openDetail(id: string) {
  drawerVisible.value = true
  detailVisible.value = false
  try {
    const res: any = await ferryGetWorkOrderDetail(id)
    const data = res?.data || res || {}
    detailWorkOrder.value = data.workOrder
    detailTasks.value = data.tasks || []
    detailHistory.value = (data.histories || []).map((h: any) => ({
      action: h.action,
      remark: h.remark,
      operatorName: h.operatorName,
      createdAt: h.createdAt
    }))
    detailVisible.value = true
  } catch (e: any) {
    ElMessage.error(e?.message || '加载失败')
  }
}

const currentStepIndex = computed(() => {
  const t = detailHistory.value.length
  return t ? t - 1 : 0
})

function prettyJSON(v: any) {
  try {
    if (!v) return '-'
    const obj = typeof v === 'string' ? JSON.parse(v) : v
    return JSON.stringify(obj, null, 2)
  } catch {
    return v
  }
}

const submitting = ref(false)
const revokeVisible = ref(false)
const revokeRemark = ref('')
const revokeRow = ref<any>(null)

const urgeVisible = ref(false)
const urgeRemark = ref('')
const urgeRow = ref<any>(null)

const newApplyVisible = ref(false)
const newApplyProcessId = ref('')
const newApplyTitle = ref('')
const newApplyPriority = ref(1)
const newApplyContent = ref('')

function revokeApply(row: any) {
  revokeRow.value = row
  revokeRemark.value = ''
  revokeVisible.value = true
}

async function submitRevoke() {
  if (!revokeRow.value) return
  submitting.value = true
  try {
    await ElMessageBox.confirm('撤回后将通知所有相关审批人，是否继续?', '撤回确认', { type: 'warning' })
    await ferryRevokeWorkOrder(revokeRow.value.id, revokeRemark.value)
    ElMessage.success('已撤回')
    revokeVisible.value = false
    loadList()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '撤回失败')
  } finally {
    submitting.value = false
  }
}

function urgeApply(row: any) {
  urgeRow.value = row
  urgeRemark.value = ''
  urgeVisible.value = true
}

async function submitUrge() {
  if (!urgeRow.value) return
  submitting.value = true
  try {
    await ferryUrgeWorkOrder(urgeRow.value.id, urgeRemark.value)
    ElMessage.success('已催办')
    urgeVisible.value = false
  } catch (e: any) {
    ElMessage.error(e?.message || '催办失败')
  } finally {
    submitting.value = false
  }
}

function openNewApply() {
  newApplyProcessId.value = ''
  newApplyTitle.value = ''
  newApplyPriority.value = 1
  newApplyContent.value = ''
  newApplyVisible.value = true
}

async function submitNewApply() {
  if (!newApplyProcessId.value) {
    ElMessage.warning('请选择流程')
    return
  }
  if (!newApplyTitle.value.trim()) {
    ElMessage.warning('请输入工单标题')
    return
  }
  submitting.value = true
  try {
    await ferrySubmitWorkOrder({
      processId: newApplyProcessId.value,
      title: newApplyTitle.value,
      priority: newApplyPriority.value,
      formData: { content: newApplyContent.value }
    })
    ElMessage.success('提交成功')
    newApplyVisible.value = false
    loadList()
  } catch (e: any) {
    ElMessage.error(e?.message || '提交失败')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadList()
  loadEnabledProcesses()
})
</script>
<style scoped>
.stat-row {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}
.stat-card {
  flex: 1;
  min-width: 150px;
  padding: 4px 8px;
}
.stat-label { color: #909399; font-size: 13px; }
.stat-value { font-size: 26px; font-weight: 700; margin-top: 4px; }
.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
  flex-wrap: wrap;
  align-items: center;
}
.ferry-spacer { flex: 1; }
.pager { display: flex; justify-content: flex-end; margin-top: 12px; }
.page-header { margin-bottom: 12px; }
.page-title { font-size: 20px; font-weight: 600; color: #303133; }
.page-sub { font-size: 13px; color: #909399; margin-top: 4px; }
</style>