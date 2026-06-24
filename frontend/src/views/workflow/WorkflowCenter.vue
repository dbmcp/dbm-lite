﻿﻿﻿﻿﻿﻿<template>
  <div class="ferry-center">
    <div class="ferry-header">
      <h2 class="ferry-title">
        <el-icon :size="22" style="vertical-align: middle; margin-right: 8px"><DocumentChecked /></el-icon>
        审批流程中心
      </h2>
      <div class="ferry-stats">
        <div class="stat-box">
          <div class="stat-label">待办</div>
          <div class="stat-value">{{ stats.todoPending }}</div>
        </div>
        <div class="stat-box">
          <div class="stat-label">我发起</div>
          <div class="stat-value">{{ stats.applyTotal }}</div>
        </div>
        <div class="stat-box">
          <div class="stat-label">流转中</div>
          <div class="stat-value">{{ stats.applyRunning }}</div>
        </div>
        <div class="stat-box">
          <div class="stat-label">已通过</div>
          <div class="stat-value">{{ stats.applyApproved }}</div>
        </div>
      </div>
    </div>

    <el-tabs v-model="activeTab" class="ferry-tabs" @tab-change="handleTabChange">
      <el-tab-pane label="我的待办" name="todo">
        <div class="ferry-toolbar">
          <el-input v-model="keywordTodo" placeholder="搜索工单标题" clearable style="width: 240px" @keyup.enter="loadTodo" />
          <el-select v-model="statusTodo" placeholder="状态" clearable style="width: 120px">
            <el-option label="待处理" value="pending" />
            <el-option label="处理中" value="processing" />
          </el-select>
          <el-button type="primary" @click="loadTodo">
            <el-icon><Search /></el-icon>搜索
          </el-button>
        </div>
        <el-table :data="todoList" v-loading="loadingTodo" border stripe>
          <el-table-column prop="workOrderId" label="工单号" width="220" />
          <el-table-column prop="nodeName" label="当前节点" width="180" />
          <el-table-column label="标题" min-width="260">
            <template #default="{ row }">
              <el-button type="primary" link @click="openDetail(row.workOrderId)">
                {{ row._title || '（点击查看）' }}
              </el-button>
            </template>
          </el-table-column>
          <el-table-column prop="assigneeName" label="处理人" width="120" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'pending' ? 'warning' : 'primary'">
                {{ TASK_STATUS_MAP[row.status || ''] || row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="createdAt" label="创建时间" width="180" />
          <el-table-column label="操作" width="220" fixed="right">
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
        <div class="ferry-pager">
          <el-pagination
            v-model:current-page="todoPage"
            v-model:page-size="todoSize"
            :total="todoTotal"
            :page-sizes="[10, 30, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="loadTodo"
            @current-change="loadTodo"
          />
        </div>
      </el-tab-pane>

      <el-tab-pane label="我的申请" name="apply">
        <div class="ferry-toolbar">
          <el-input v-model="keywordApply" placeholder="搜索标题" clearable style="width: 240px" />
          <el-select v-model="statusApply" placeholder="状态" clearable style="width: 140px">
            <el-option label="流转中" value="running" />
            <el-option label="已通过" value="approved" />
            <el-option label="已驳回" value="rejected" />
            <el-option label="已撤销" value="canceled" />
          </el-select>
          <el-button type="primary" @click="loadApply">
            <el-icon><Search /></el-icon>搜索
          </el-button>
          <el-button type="success" @click="openApplyDialog">
            <el-icon><Plus /></el-icon>发起申请
          </el-button>
        </div>
        <el-table :data="applyList" v-loading="loadingApply" border stripe>
          <el-table-column prop="id" label="工单号" width="220" />
          <el-table-column prop="title" label="标题" min-width="240" />
          <el-table-column prop="processName" label="流程" width="180" />
          <el-table-column label="优先级" width="100">
            <template #default="{ row }">
              <el-tag :type="(row.priority || 1) >= 2 ? 'danger' : 'info'" size="small">
                {{ PRIORITY_MAP[row.priority || 1] || '普通' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="statusTagType(row.status)">{{ STATUS_MAP[row.status || ''] || row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="creatorName" label="发起人" width="120" />
          <el-table-column prop="createdAt" label="发起时间" width="180" />
          <el-table-column label="操作" width="260" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" link @click="openDetail(row.id)">详情</el-button>
              <el-button
                v-if="row.status === 'running'"
                type="warning"
                size="small"
                link
                @click="confirmRevoke(row.id)"
              >撤销</el-button>
              <el-button
                v-if="row.status === 'running'"
                type="info"
                size="small"
                link
                @click="confirmUrge(row.id)"
              >催办</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="ferry-pager">
          <el-pagination
            v-model:current-page="applyPage"
            v-model:page-size="applySize"
            :total="applyTotal"
            :page-sizes="[10, 30, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="loadApply"
            @current-change="loadApply"
          />
        </div>
      </el-tab-pane>

      <el-tab-pane label="与我有关" name="related">
        <div class="ferry-toolbar">
          <el-input v-model="keywordRelated" placeholder="搜索标题" clearable style="width: 240px" @keyup.enter="loadRelated" />
          <el-button type="primary" @click="loadRelated">
            <el-icon><Search /></el-icon>搜索
          </el-button>
        </div>
        <el-table :data="relatedList" v-loading="loadingRelated" border stripe>
          <el-table-column prop="id" label="工单号" width="220" />
          <el-table-column prop="title" label="标题" min-width="240" />
          <el-table-column prop="processName" label="流程" width="180" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="statusTagType(row.status)">{{ STATUS_MAP[row.status || ''] || row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="creatorName" label="发起人" width="120" />
          <el-table-column prop="createdAt" label="时间" width="180" />
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" link @click="openDetail(row.id)">详情</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="ferry-pager">
          <el-pagination
            v-model:current-page="relatedPage"
            v-model:page-size="relatedSize"
            :total="relatedTotal"
            :page-sizes="[10, 30, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="loadRelated"
            @current-change="loadRelated"
          />
        </div>
      </el-tab-pane>

      <el-tab-pane label="所有工单" name="all" v-if="isAdmin">
        <div class="ferry-toolbar">
          <el-input v-model="keywordAll" placeholder="搜索标题" clearable style="width: 240px" />
          <el-select v-model="statusAll" placeholder="状态" clearable style="width: 140px">
            <el-option label="流转中" value="running" />
            <el-option label="已通过" value="approved" />
            <el-option label="已驳回" value="rejected" />
            <el-option label="已撤销" value="canceled" />
          </el-select>
          <el-button type="primary" @click="loadAll">
            <el-icon><Search /></el-icon>搜索
          </el-button>
        </div>
        <el-table :data="allList" v-loading="loadingAll" border stripe>
          <el-table-column prop="id" label="工单号" width="220" />
          <el-table-column prop="title" label="标题" min-width="240" />
          <el-table-column prop="processName" label="流程" width="180" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="statusTagType(row.status)">{{ STATUS_MAP[row.status || ''] || row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="creatorName" label="发起人" width="120" />
          <el-table-column prop="createdAt" label="时间" width="180" />
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" link @click="openDetail(row.id)">详情</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="ferry-pager">
          <el-pagination
            v-model:current-page="allPage"
            v-model:page-size="allSize"
            :total="allTotal"
            :page-sizes="[10, 30, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="loadAll"
            @current-change="loadAll"
          />
        </div>
      </el-tab-pane>
      <el-tab-pane label="流程管理" name="process" v-if="isAdmin">
        <div class="ferry-toolbar">
          <span style="color:#606266; font-size:13px">
            流程定义管理：维护系统内所有可用的业务流程（仅管理员可见）。
          </span>
          <el-button type="primary" @click="navigateTo('/process-manager')">进入流程管理</el-button>
        </div>
        <el-alert type="info" :closable="false" show-icon title="提示：流程管理与流程分类仅对管理员开放。" />
      </el-tab-pane>
      <el-tab-pane label="流程分类" name="classify" v-if="isAdmin">
        <div class="ferry-toolbar">
          <span style="color:#606266; font-size:13px">
            流程分类管理：维护流程分组与层级（仅管理员可见）。
          </span>
          <el-button type="primary" @click="navigateTo('/classify-manager')">进入流程分类</el-button>
        </div>
        <el-alert type="info" :closable="false" show-icon title="提示：流程管理与流程分类仅对管理员开放。" />
      </el-tab-pane>
    </el-tabs>

    <!-- 工单详情 Drawer -->
    <el-drawer
      v-model="detailVisible"
      :title="'工单详情：' + (detail.workOrder?.title || '')"
      direction="rtl"
      size="720px"
    >
      <template v-if="detail.workOrder">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="工单号">{{ detail.workOrder.id }}</el-descriptions-item>
          <el-descriptions-item label="流程名称">{{ detail.workOrder.processName }}</el-descriptions-item>
          <el-descriptions-item label="标题">{{ detail.workOrder.title }}</el-descriptions-item>
          <el-descriptions-item label="优先级">{{ PRIORITY_MAP[detail.workOrder.priority || 1] }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTagType(detail.workOrder.status)">{{ STATUS_MAP[detail.workOrder.status || ''] || detail.workOrder.status }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="发起人">{{ detail.workOrder.creatorName }}</el-descriptions-item>
          <el-descriptions-item label="发起时间">{{ detail.workOrder.createdAt }}</el-descriptions-item>
          <el-descriptions-item label="完成时间">{{ detail.workOrder.finishedAt || '-' }}</el-descriptions-item>
        </el-descriptions>

        <h4 style="margin: 20px 0 10px">表单数据</h4>
        <div class="ferry-formdata">
          <pre v-if="detail.workOrder.formData">{{ prettyJSON(detail.workOrder.formData) }}</pre>
          <span v-else style="color: #909399">无</span>
        </div>

        <h4 style="margin: 20px 0 10px">任务节点</h4>
        <el-table :data="detail.tasks" border size="small">
          <el-table-column prop="nodeName" label="节点" width="160" />
          <el-table-column prop="assigneeName" label="处理人" width="120" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag size="small" :type="taskStatusTagType(row.status)">
                {{ TASK_STATUS_MAP[row.status || ''] || row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="remark" label="备注" min-width="160" />
          <el-table-column prop="operatorName" label="操作人" width="120" />
          <el-table-column prop="handledAt" label="处理时间" width="180" />
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
              <el-button
                v-if="row.status === 'pending'"
                type="success"
                size="small"
                link
                @click="handleTask(row, 'approve', detail.workOrder!.id)"
              >通过</el-button>
              <el-button
                v-if="row.status === 'pending'"
                type="danger"
                size="small"
                link
                @click="handleTask(row, 'reject', detail.workOrder!.id)"
              >驳回</el-button>
            </template>
          </el-table-column>
        </el-table>

        <h4 style="margin: 20px 0 10px">流转历史</h4>
        <el-timeline>
          <el-timeline-item
            v-for="(h, idx) in (detail.histories || []).slice().reverse()"
            :key="h.id || idx"
            :timestamp="h.createdAt"
            :type="historyTimelineType(h.action)"
            placement="top"
          >
            <h4>{{ actionLabel(h.action) }} · {{ h.operatorName || '-' }}</h4>
            <p v-if="h.nodeName">节点：{{ h.nodeName }}</p>
            <p v-if="h.remark" style="color: #606266">{{ h.remark }}</p>
          </el-timeline-item>
        </el-timeline>
      </template>
    </el-drawer>

    <!-- 发起申请 Dialog -->
    <el-dialog v-model="applyDialogVisible" title="发起申请" width="720px">
      <el-form :model="applyForm" label-width="100px">
        <el-form-item label="流程" required>
          <el-select v-model="applyForm.processId" placeholder="选择流程" style="width: 100%" @change="onProcessSelected">
            <el-option
              v-for="p in enabledProcesses"
              :key="p.id"
              :label="p.name"
              :value="p.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="标题" required>
          <el-input v-model="applyForm.title" placeholder="请输入标题" />
        </el-form-item>
        <el-form-item label="优先级">
          <el-select v-model="applyForm.priority" style="width: 100%">
            <el-option label="普通" :value="1" />
            <el-option label="紧急" :value="2" />
            <el-option label="非常紧急" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="applyFormSelectedProcess?.description" label="流程说明">
          <div style="color: #606266">{{ applyFormSelectedProcess.description }}</div>
        </el-form-item>
        <el-form-item label="申请内容">
          <el-input
            v-model="applyForm.formText"
            type="textarea"
            :rows="6"
            placeholder="请输入详细申请内容"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="applyDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="applying" @click="doSubmitApply">提交</el-button>
      </template>
    </el-dialog>

    <!-- 通过/驳回 Dialog -->
    <el-dialog v-model="handleDialogVisible" :title="handleAction === 'approve' ? '通过任务' : '驳回任务'" width="480px">
      <el-form label-width="80px">
        <el-form-item label="备注">
          <el-input v-model="handleRemark" type="textarea" :rows="4" :placeholder="handleAction === 'approve' ? '请输入通过说明（可选）' : '请输入驳回说明'" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="handleDialogVisible = false">取消</el-button>
        <el-button :type="handleAction === 'approve' ? 'success' : 'danger'" :loading="handling" @click="doHandleTask">
          {{ handleAction === 'approve' ? '确认通过' : '确认驳回' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ferryStatistics,
  ferryMyTodo,
  ferryMyApply,
  ferryMyRelated,
  ferryAllWorkOrders,
  ferryListEnabledProcesses,
  ferrySubmitWorkOrder,
  ferryGetWorkOrderDetail,
  ferryHandleTask,
  ferryRevokeWorkOrder,
  ferryUrgeWorkOrder,
  STATUS_MAP,
  TASK_STATUS_MAP,
  PRIORITY_MAP,
  WorkOrderTask,
  WorkOrder
} from '@/api/ferry'
import { useUserStore } from '@/stores/user'
import { useRouter } from 'vue-router'

const userStore = useUserStore()
const isAdmin = computed(() => !!userStore.userInfo?.role && String(userStore.userInfo.role).toLowerCase().includes('admin'))
const router = useRouter()
function navigateTo(path: string) {
  router.push(path)
}

// ============ 统计 ============
const stats = reactive({
  todoPending: 0,
  applyTotal: 0,
  applyRunning: 0,
  applyApproved: 0
})

function loadStats() {
  ferryStatistics()
    .then((res: any) => {
      const d = res?.data || {}
      stats.todoPending = d?.todo?.pending ?? 0
      stats.applyTotal = d?.apply?.total ?? 0
      stats.applyRunning = d?.apply?.running ?? 0
      stats.applyApproved = d?.apply?.approved ?? 0
    })
    .catch(() => {})
}

// ============ 我的待办 ============
const activeTab = ref<string>('todo')
const keywordTodo = ref('')
const statusTodo = ref('')
const todoPage = ref(1)
const todoSize = ref(50)
const todoTotal = ref(0)
const todoList = ref<(WorkOrderTask & { _title?: string })[]>([])
const loadingTodo = ref(false)

async function loadTodo() {
  loadingTodo.value = true
  try {
    const res = await ferryMyTodo({
      page: todoPage.value,
      size: todoSize.value,
      keyword: keywordTodo.value,
      status: statusTodo.value
    })
    todoList.value = ((res?.data as any[]) || []).map((t) => ({ ...t, _title: t.workOrderId }))
    todoTotal.value = (res?.total as number) || todoList.value.length
  } finally {
    loadingTodo.value = false
  }
}

// ============ 我的申请 ============
const keywordApply = ref('')
const statusApply = ref('')
const applyPage = ref(1)
const applySize = ref(50)
const applyTotal = ref(0)
const applyList = ref<WorkOrder[]>([])
const loadingApply = ref(false)

async function loadApply() {
  loadingApply.value = true
  try {
    const res = await ferryMyApply({
      page: applyPage.value,
      size: applySize.value,
      keyword: keywordApply.value,
      status: statusApply.value
    })
    applyList.value = (res?.data as WorkOrder[]) || []
    applyTotal.value = (res?.total as number) || applyList.value.length
  } finally {
    loadingApply.value = false
  }
}

// ============ 与我有关 ============
const keywordRelated = ref('')
const relatedPage = ref(1)
const relatedSize = ref(50)
const relatedTotal = ref(0)
const relatedList = ref<WorkOrder[]>([])
const loadingRelated = ref(false)

async function loadRelated() {
  loadingRelated.value = true
  try {
    const res = await ferryMyRelated({
      page: relatedPage.value,
      size: relatedSize.value,
      keyword: keywordRelated.value
    })
    relatedList.value = (res?.data as WorkOrder[]) || []
    relatedTotal.value = (res?.total as number) || relatedList.value.length
  } finally {
    loadingRelated.value = false
  }
}

// ============ 所有工单 ============
const keywordAll = ref('')
const statusAll = ref('')
const allPage = ref(1)
const allSize = ref(50)
const allTotal = ref(0)
const allList = ref<WorkOrder[]>([])
const loadingAll = ref(false)

async function loadAll() {
  loadingAll.value = true
  try {
    const res = await ferryAllWorkOrders({
      page: allPage.value,
      size: allSize.value,
      keyword: keywordAll.value,
      status: statusAll.value
    })
    allList.value = (res?.data as WorkOrder[]) || []
    allTotal.value = (res?.total as number) || allList.value.length
  } finally {
    loadingAll.value = false
  }
}

function handleTabChange(t: string) {
  if (t === 'todo') loadTodo()
  else if (t === 'apply') loadApply()
  else if (t === 'related') loadRelated()
  else if (t === 'all') loadAll()
}

// ============ 工单详情 ============
const detailVisible = ref(false)
const detail = reactive<{
  workOrder: WorkOrder | null
  tasks: WorkOrderTask[]
  histories: any[]
}>({ workOrder: null, tasks: [], histories: [] })

async function openDetail(id: string) {
  try {
    const res = await ferryGetWorkOrderDetail(id)
    detail.workOrder = (res?.data as any)?.workOrder || null
    detail.tasks = ((res?.data as any)?.tasks as WorkOrderTask[]) || []
    detail.histories = ((res?.data as any)?.histories as any[]) || []
    detailVisible.value = true
  } catch (e: any) {
    ElMessage.error(e?.message || '加载详情失败')
  }
}

// ============ 处理任务 ============
const handleDialogVisible = ref(false)
const handleRemark = ref('')
const handleAction = ref<'approve' | 'reject'>('approve')
const handling = ref(false)
let pendingTask: WorkOrderTask | null = null
let pendingWorkOrderId: string | null = null

function handleTask(task: WorkOrderTask, action: 'approve' | 'reject', woId?: string) {
  pendingTask = task
  pendingWorkOrderId = woId || task.workOrderId
  handleAction.value = action
  handleRemark.value = ''
  handleDialogVisible.value = true
}

async function doHandleTask() {
  if (!pendingTask || !pendingWorkOrderId) return
  handling.value = true
  try {
    await ferryHandleTask(pendingWorkOrderId, pendingTask.id, handleAction.value, handleRemark.value)
    ElMessage.success(handleAction.value === 'approve' ? '已通过' : '已驳回')
    handleDialogVisible.value = false
    loadTodo()
    loadApply()
    if (detailVisible.value && pendingWorkOrderId === detail.workOrder?.id) {
      openDetail(pendingWorkOrderId)
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  } finally {
    handling.value = false
  }
}

// ============ 撤销 ============
function confirmRevoke(id: string) {
  ElMessageBox.prompt('请输入撤销说明（可选）', '撤销申请', {
    confirmButtonText: '确认撤销',
    cancelButtonText: '取消',
    inputPlaceholder: '说明',
    type: 'warning'
  })
    .then(async ({ value }: any) => {
      await ferryRevokeWorkOrder(id, value || '')
      ElMessage.success('已撤销')
      loadApply()
      if (detailVisible.value && id === detail.workOrder?.id) openDetail(id)
    })
    .catch(() => {})
}

function confirmUrge(id: string) {
  ElMessageBox.prompt('请输入催办说明（可选）', '催办', {
    confirmButtonText: '确认催办',
    cancelButtonText: '取消',
    inputPlaceholder: '说明',
    type: 'info'
  })
    .then(async ({ value }: any) => {
      await ferryUrgeWorkOrder(id, value || '')
      ElMessage.success('已催办')
      loadApply()
    })
    .catch(() => {})
}

// ============ 发起申请 ============
const applyDialogVisible = ref(false)
const enabledProcesses = ref<any[]>([])
const applyForm = reactive({ processId: '', title: '', priority: 1, formText: '' })
const applying = ref(false)

const applyFormSelectedProcess = computed(() =>
  enabledProcesses.value.find((p) => p.id === applyForm.processId) || null
)

async function openApplyDialog() {
  if (!enabledProcesses.value.length) {
    try {
      const res = await ferryListEnabledProcesses()
      enabledProcesses.value = (res?.data as any[]) || []
    } catch (e: any) {
      ElMessage.error(e?.message || '加载流程失败')
      return
    }
  }
  applyForm.processId = ''
  applyForm.title = ''
  applyForm.priority = 1
  applyForm.formText = ''
  applyDialogVisible.value = true
}

function onProcessSelected() {
  const proc = applyFormSelectedProcess.value
  if (proc && !applyForm.title) {
    applyForm.title = '【' + proc.name + '】申请'
  }
}

async function doSubmitApply() {
  if (!applyForm.processId) {
    ElMessage.warning('请选择流程')
    return
  }
  if (!applyForm.title.trim()) {
    ElMessage.warning('请输入标题')
    return
  }
  applying.value = true
  try {
    let formData: any = undefined
    if (applyForm.formText.trim()) {
      try {
        formData = JSON.parse(applyForm.formText)
      } catch {
        formData = { content: applyForm.formText.trim() }
      }
    }
    await ferrySubmitWorkOrder({
      processId: applyForm.processId,
      title: applyForm.title.trim(),
      priority: applyForm.priority,
      formData
    })
    ElMessage.success('申请已提交')
    applyDialogVisible.value = false
    loadApply()
    loadStats()
  } catch (e: any) {
    ElMessage.error(e?.message || '提交失败')
  } finally {
    applying.value = false
  }
}

// ============ 工具 ============
function statusTagType(status?: string) {
  switch (status) {
    case 'running':
      return 'primary'
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

function taskStatusTagType(status?: string) {
  switch (status) {
    case 'pending':
      return 'warning'
    case 'approved':
      return 'success'
    case 'rejected':
      return 'danger'
    case 'skipped':
      return 'info'
    case 'processing':
      return 'primary'
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
    return JSON.stringify(data, null, 2)
  } catch {
    return String(data)
  }
}

onMounted(() => {
  loadStats()
  loadTodo()
})
</script>

<style scoped>
.ferry-center {
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

.ferry-stats {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.stat-box {
  background: #fff;
  border-radius: 6px;
  padding: 10px 18px;
  min-width: 88px;
  text-align: center;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}

.stat-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.stat-value {
  font-size: 20px;
  font-weight: 600;
  color: #409eff;
}

.ferry-tabs {
  background: #fff;
  padding: 8px 16px;
  border-radius: 8px;
}

.ferry-toolbar {
  display: flex;
  gap: 8px;
  margin: 12px 0;
  flex-wrap: wrap;
}

.ferry-pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 12px;
}

.ferry-formdata pre {
  background: #f5f7fa;
  border-radius: 4px;
  padding: 10px;
  font-size: 12px;
  line-height: 1.6;
  max-height: 300px;
  overflow: auto;
}
</style>
