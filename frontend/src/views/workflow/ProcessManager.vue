<template>
  <div class="ferry-page">
    <div class="ferry-header">
      <div>
        <h2 class="ferry-title">
          <el-icon :size="22" style="vertical-align: middle; margin-right: 8px"><Setting /></el-icon>
          流程管理
        </h2>
        <div class="ferry-subtitle">定义与配置业务审批流程，包括节点结构、表单 Schema 与通知规则。</div>
      </div>
    </div>

    <div class="ferry-stats">
      <div class="stat-box stat-total">
        <div class="stat-label">流程总数</div>
        <div class="stat-value">{{ stats.total }}</div>
      </div>
      <div class="stat-box stat-enabled">
        <div class="stat-label">已启用</div>
        <div class="stat-value">{{ stats.enabled }}</div>
      </div>
      <div class="stat-box stat-disabled">
        <div class="stat-label">已禁用</div>
        <div class="stat-value">{{ stats.disabled }}</div>
      </div>
      <div class="stat-box stat-submits">
        <div class="stat-label">近 7 天提交次数</div>
        <div class="stat-value">{{ stats.recentSubmits }}</div>
      </div>
    </div>

    <div class="ferry-toolbar">
      <el-input v-model="keyword" placeholder="搜索流程名称" clearable style="width: 240px" @keyup.enter="loadList" />
      <el-select v-model="classifyFilter" placeholder="分类" clearable style="width: 160px">
        <el-option
          v-for="c in classifyList"
          :key="c.id"
          :label="c.name"
          :value="c.id"
        />
      </el-select>
      <el-select v-model="enabledFilter" placeholder="启用状态" clearable style="width: 140px">
        <el-option label="已启用" :value="true" />
        <el-option label="已禁用" :value="false" />
      </el-select>
      <el-button type="primary" @click="loadList">
        <el-icon><Search /></el-icon>搜索
      </el-button>
      <div class="ferry-spacer" ></div>
      <ColumnToggle v-model="colVisible" :columns="columns" />
      <el-button type="success" @click="openCreate">
        <el-icon><Plus /></el-icon>新建流程
      </el-button>
    </div>

    <el-table
      :data="filteredList"
      v-loading="loading"
      border
      stripe
      @selection-change="onSelectionChange"
    >
      <el-table-column type="selection" width="48" />
      <el-table-column v-if="colVisible.name" label="流程名称" min-width="220" show-overflow-tooltip>
        <template #default="{ row }">
          <el-icon :size="16" style="vertical-align: middle; color: #409eff">
            <component :is="row.icon ? row.icon : 'Document'" />
          </el-icon>
          <span style="margin-left: 6px">{{ row.name }}</span>
        </template>
      </el-table-column>
      <el-table-column v-if="colVisible.icon" label="图标" width="90" align="center">
        <template #default="{ row }">
          <span v-if="row.icon" style="font-size: 16px">{{ row.icon }}</span>
          <span v-else style="color: #c0c4cc; font-size: 12px">—</span>
        </template>
      </el-table-column>
      <el-table-column v-if="colVisible.classifyName" prop="classifyName" label="分类" width="160" show-overflow-tooltip>
        <template #default="{ row }">
          <el-tag size="small" type="info" effect="plain">{{ row.classifyName || '未分类' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column v-if="colVisible.description" prop="description" label="描述" min-width="260" show-overflow-tooltip>
        <template #default="{ row }">
          <span v-if="row.description">{{ row.description }}</span>
          <span v-else style="color: #c0c4cc">—</span>
        </template>
      </el-table-column>
      <el-table-column v-if="colVisible.status" label="状态" width="100" align="center">
        <template #default="{ row }">
          <el-tag size="small" :type="row.enabled ? 'success' : 'info'">
            {{ row.enabled ? '已启用' : '已禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column v-if="colVisible.submitCnt" prop="submitCnt" label="提交次数" width="110" align="center" />
      <el-table-column v-if="colVisible.creatorName" prop="creatorName" label="创建人" width="120" show-overflow-tooltip />
      <el-table-column v-if="colVisible.createdAt" prop="createdAt" label="创建时间" width="180" show-overflow-tooltip />
      <el-table-column v-if="colVisible.updatedAt" prop="updatedAt" label="更新时间" width="180" show-overflow-tooltip />
      <el-table-column v-if="colVisible.actions" label="操作" width="280" fixed="right" show-overflow-tooltip>
        <template #default="{ row }">
          <el-button size="small" type="primary" link @click="openEdit(row)">编辑</el-button>
          <el-button size="small" type="success" link @click="toggleEnabled(row)">
            {{ row.enabled ? '禁用' : '启用' }}
          </el-button>
          <el-button size="small" type="warning" link @click="cloneProcess(row)">克隆</el-button>
          <el-button size="small" type="danger" link @click="deleteProcess(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="ferry-pager">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[100, 200, 500, 1000]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadList"
        @current-change="loadList"
      />
    </div>

    <el-dialog
      v-model="formVisible"
      :title="formMode === 'create' ? '新建流程' : '编辑流程'"
      width="920px"
      top="6vh"
    >
      <el-tabs v-model="formActiveTab" class="ferry-form-tabs">
        <el-tab-pane label="基础信息" name="basic">
          <el-form :model="form" label-width="100px">
            <el-form-item label="流程名称" required>
              <el-input v-model="form.name" placeholder="请输入流程名称" />
            </el-form-item>
            <el-form-item label="图标">
              <el-input v-model="form.icon" placeholder="emoji 或图标文本，如：📋" maxlength="6" />
            </el-form-item>
            <el-form-item label="所属分类" required>
              <el-select v-model="form.classifyId" placeholder="选择分类" style="width: 100%" clearable>
                <el-option
                  v-for="c in classifyList"
                  :key="c.id"
                  :label="c.name"
                  :value="c.id"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="描述">
              <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入流程描述" />
            </el-form-item>
            <el-form-item label="是否启用">
              <el-switch v-model="form.enabled" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="节点结构" name="structure">
          <div class="ferry-flow-head">
            <div class="ferry-flow-tip">
              拖拽节点调整顺序，或点击 <b>＋ 新增节点</b> 添加人工审批 / 条件节点。
            </div>
            <div class="ferry-flow-actions">
              <el-button size="small" @click="addNode('userTask')">
                <el-icon><User /></el-icon>＋ 人工节点
              </el-button>
              <el-button size="small" @click="addNode('condition')">
                <el-icon><Operation /></el-icon>＋ 条件节点
              </el-button>
              <el-button size="small" type="danger" plain @click="resetStructureToDefault">重置默认</el-button>
            </div>
          </div>

          <div class="ferry-flow-canvas" ref="flowCanvasRef">
            <template v-for="(node, idx) in structure.nodes" :key="node.id">
              <div
                class="ferry-flow-node"
                :class="'node-' + node.type"
                draggable="true"
                @dragstart="onDragStart(idx, $event)"
                @dragover.prevent
                @drop="onDrop(idx, $event)"
              >
                <div class="ferry-flow-node-head">
                  <span class="ferry-flow-node-title">
                    <el-icon :size="14"><component :is="nodeIcon(node.type)" /></el-icon>
                    {{ nodeLabel(node) }}
                  </span>
                  <el-button
                    v-if="node.type !== 'start' && node.type !== 'end'"
                    link
                    type="danger"
                    size="small"
                    @click="removeNode(idx)"
                  >删除</el-button>
                </div>
                <div class="ferry-flow-node-body">
                  <el-form label-width="70px" size="small">
                    <el-form-item label="名称">
                      <el-input v-model="node.name" placeholder="节点名称" />
                    </el-form-item>
                    <el-form-item v-if="node.type === 'userTask'" label="审批人">
                      <el-input
                        v-model="node.approvers"
                        type="textarea"
                        :rows="2"
                        placeholder="多个审批人用逗号或换行分隔，支持用户 ID / 角色 / 部门"
                      />
                    </el-form-item>
                    <el-form-item v-if="node.type === 'condition'" label="条件">
                      <el-input
                        v-model="node.condition"
                        type="textarea"
                        :rows="2"
                        placeholder="如：amount > 5000 或 role === 'manager'"
                      />
                    </el-form-item>
                  </el-form>
                </div>
              </div>
              <div v-if="idx < structure.nodes.length - 1" class="ferry-flow-arrow">
                <div class="ferry-flow-arrow-line" ></div>
                <el-icon class="ferry-flow-arrow-icon" :size="16"><Right /></el-icon>
              </div>
            </template>
          </div>

          <div class="ferry-flow-json">
            <div class="ferry-flow-json-head">
              <span>节点结构 JSON（自动生成，可手动编辑）</span>
              <el-button size="small" text @click="syncStructureFromText">从文本同步</el-button>
            </div>
            <el-input
              v-model="structureText"
              type="textarea"
              :rows="6"
              placeholder="结构 JSON，包含 nodes 与 edges"
              class="ferry-flow-json-textarea"
            />
          </div>
        </el-tab-pane>

        <el-tab-pane label="表单 Schema" name="formSchema">
          <div class="ferry-schema-head">
            <div class="ferry-schema-tip">
              简单键值 Schema：字段名 = 类型 + 标签 + 是否必填。示例：
              <code>amount|number|金额|true</code>
            </div>
            <el-button size="small" type="primary" plain @click="addSchemaField">
              <el-icon><Plus /></el-icon>新增字段
            </el-button>
          </div>
          <el-table :data="formSchemaFields" border size="small" class="ferry-schema-table">
            <el-table-column label="字段名" min-width="140">
              <template #default="{ row }">
                <el-input v-model="row.key" placeholder="如：amount" size="small" />
              </template>
            </el-table-column>
            <el-table-column label="类型" width="130">
              <template #default="{ row }">
                <el-select v-model="row.type" size="small" style="width: 100%">
                  <el-option label="string" value="string" />
                  <el-option label="number" value="number" />
                  <el-option label="boolean" value="boolean" />
                  <el-option label="textarea" value="textarea" />
                  <el-option label="date" value="date" />
                  <el-option label="select" value="select" />
                </el-select>
              </template>
            </el-table-column>
            <el-table-column label="标签" min-width="160">
              <template #default="{ row }">
                <el-input v-model="row.label" placeholder="显示标签，如：金额" size="small" />
              </template>
            </el-table-column>
            <el-table-column label="必填" width="80" align="center">
              <template #default="{ row }">
                <el-checkbox v-model="row.required" />
              </template>
            </el-table-column>
            <el-table-column label="默认值" min-width="140">
              <template #default="{ row }">
                <el-input v-model="row.default" placeholder="默认值（可选）" size="small" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="80" align="center">
              <template #default="{ $index }">
                <el-button link type="danger" size="small" @click="removeSchemaField($index)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="ferry-schema-json">
            <div class="ferry-schema-json-head">
              <span>生成的 JSON Schema（可直接编辑）</span>
              <el-button size="small" text @click="syncFormSchemaFromText">从文本同步</el-button>
            </div>
            <el-input
              v-model="formSchemaText"
              type="textarea"
              :rows="6"
              class="ferry-schema-json-textarea"
            />
          </div>
        </el-tab-pane>

        <el-tab-pane label="通知配置" name="notify">
          <el-form :model="notify" label-width="120px">
            <el-form-item label="启用通知">
              <el-switch v-model="notify.enabled" />
            </el-form-item>
            <el-form-item label="通知渠道">
              <el-checkbox-group v-model="notify.channels">
                <el-checkbox label="email">邮件</el-checkbox>
                <el-checkbox label="sms">短信</el-checkbox>
                <el-checkbox label="inbox">站内信</el-checkbox>
                <el-checkbox label="webhook">Webhook</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
            <el-form-item label="通知时机">
              <el-checkbox-group v-model="notify.events">
                <el-checkbox label="submit">提交时</el-checkbox>
                <el-checkbox label="approve">通过时</el-checkbox>
                <el-checkbox label="reject">驳回时</el-checkbox>
                <el-checkbox label="transfer">转交时</el-checkbox>
                <el-checkbox label="finish">完成时</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
            <el-form-item label="额外接收人">
              <el-input
                v-model="notify.extraReceivers"
                type="textarea"
                :rows="2"
                placeholder="多个接收人用逗号或换行分隔（用户 ID / 邮箱 / 角色）"
              />
            </el-form-item>
            <el-form-item label="Webhook URL">
              <el-input v-model="notify.webhookUrl" placeholder="https://example.com/hook" />
            </el-form-item>
            <el-form-item label="通知模板">
              <el-input
                v-model="notify.template"
                type="textarea"
                :rows="4"
                placeholder="支持变量：{{ '${processName}' }}、{{ '${workOrderId}' }}、{{ '${creator}' }}、{{ '${nodeName}' }}"
              />
            </el-form-item>
          </el-form>
          <div class="ferry-notify-json">
            <div class="ferry-notify-json-head">
              <span>生成的 JSON（可直接编辑）</span>
              <el-button size="small" text @click="syncNotifyFromText">从文本同步</el-button>
            </div>
            <el-input v-model="notifyText" type="textarea" :rows="5" class="ferry-notify-json-textarea" />
          </div>
        </el-tab-pane>
      </el-tabs>

      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="doSubmit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ColumnToggle from '@/components/ColumnToggle.vue'
import {
  ferryListProcesses,
  ferryCreateProcess,
  ferryUpdateProcess,
  ferryDeleteProcess,
  ferryToggleProcessEnabled,
  ferryCloneProcess,
  ferryListClassifies
} from '@/api/ferry'

const columns = [
  { key: 'name', label: '流程名称' },
  { key: 'icon', label: '图标' },
  { key: 'classifyName', label: '分类' },
  { key: 'description', label: '描述' },
  { key: 'status', label: '状态' },
  { key: 'submitCnt', label: '提交次数' },
  { key: 'creatorName', label: '创建人' },
  { key: 'createdAt', label: '创建时间' },
  { key: 'updatedAt', label: '更新时间' },
  { key: 'actions', label: '操作' }
]
const colVisible = reactive<Record<string, boolean>>({
  name: true,
  icon: true,
  classifyName: true,
  description: true,
  status: true,
  submitCnt: true,
  creatorName: true,
  createdAt: true,
  updatedAt: true,
  actions: true
})

interface FlowNode {
  id: string
  type: 'start' | 'userTask' | 'condition' | 'end'
  name: string
  approvers?: string
  condition?: string
}

interface FlowStructure {
  nodes: FlowNode[]
  edges: { from: string; to: string }[]
}

interface SchemaField {
  key: string
  type: string
  label: string
  required: boolean
  default: string
}

interface NotifyConfig {
  enabled: boolean
  channels: string[]
  events: string[]
  extraReceivers: string
  webhookUrl: string
  template: string
}

const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(100)
const keyword = ref('')
const classifyFilter = ref('')
const enabledFilter = ref<string | boolean>('')
const loading = ref(false)
const classifyList = ref<any[]>([])
const selectedIds = ref<string[]>([])

const stats = reactive({
  total: 0,
  enabled: 0,
  disabled: 0,
  recentSubmits: 0
})

const filteredList = computed(() => {
  return list.value.filter((row) => {
    if (classifyFilter.value && row.classifyId !== classifyFilter.value) return false
    if (enabledFilter.value !== '' && enabledFilter.value !== null && enabledFilter.value !== undefined) {
      if (Boolean(row.enabled) !== Boolean(enabledFilter.value)) return false
    }
    return true
  })
})

function classifyNameById(id: string) {
  const c = classifyList.value.find((x) => x.id === id)
  return c ? c.name : ''
}

async function loadList() {
  loading.value = true
  try {
    const res = await ferryListProcesses({
      page: page.value,
      size: size.value,
      keyword: keyword.value
    })
    const raw = (res?.data as any[]) || []
    list.value = raw.map((r) => ({
      ...r,
      classifyId: r.classifyId || r.classify || '',
      classifyName: r.classifyName || classifyNameById(r.classifyId || r.classify || '')
    }))
    total.value = (res?.total as number) || list.value.length
    recomputeStats()
  } finally {
    loading.value = false
  }
}

function recomputeStats() {
  const all = list.value
  stats.total = all.length
  stats.enabled = all.filter((x) => !!x.enabled).length
  stats.disabled = all.filter((x) => !x.enabled).length
  stats.recentSubmits = all.reduce((sum, x) => sum + (Number(x.submitCnt) || 0), 0)
}

async function loadClassifies() {
  try {
    const res = await ferryListClassifies()
    classifyList.value = (res?.data as any[]) || []
  } catch {}
}

function onSelectionChange(rows: any[]) {
  selectedIds.value = rows.map((r) => r.id)
}

// ============ 表单 ============
const formVisible = ref(false)
const formMode = ref<'create' | 'edit'>('create')
const formActiveTab = ref('basic')
const submitting = ref(false)

const form = ref<any>({
  id: '',
  name: '',
  icon: '',
  classifyId: '',
  description: '',
  enabled: true
})

const structure = reactive<FlowStructure>({
  nodes: [],
  edges: []
})

const structureText = ref('')

const formSchemaFields = ref<SchemaField[]>([])
const formSchemaText = ref('')

const notify = reactive<NotifyConfig>({
  enabled: true,
  channels: ['inbox', 'email'],
  events: ['submit', 'approve', 'reject'],
  extraReceivers: '',
  webhookUrl: '',
  template: '【${processName}】工单号 ${workOrderId} 已由 ${creator} 提交，当前节点：${nodeName}。'
})
const notifyText = ref('')

let syncLock = false

function defaultStructure(): FlowStructure {
  const n1: FlowNode = { id: 'n_start', type: 'start', name: '开始' }
  const n2: FlowNode = { id: 'n_user_1', type: 'userTask', name: '部门主管审批', approvers: 'department.manager' }
  const n3: FlowNode = { id: 'n_user_2', type: 'userTask', name: '总经理审批', approvers: 'general.manager' }
  const n4: FlowNode = { id: 'n_end', type: 'end', name: '结束' }
  return {
    nodes: [n1, n2, n3, n4],
    edges: [
      { from: n1.id, to: n2.id },
      { from: n2.id, to: n3.id },
      { from: n3.id, to: n4.id }
    ]
  }
}

function resetStructureToDefault() {
  const def = defaultStructure()
  structure.nodes = def.nodes
  structure.edges = def.edges
}

function rebuildEdgesFromNodes() {
  structure.edges = []
  for (let i = 0; i < structure.nodes.length - 1; i++) {
    structure.edges.push({ from: structure.nodes[i].id, to: structure.nodes[i + 1].id })
  }
}

function addNode(type: 'userTask' | 'condition') {
  const id = 'n_' + type + '_' + Date.now() + '_' + Math.floor(Math.random() * 1000)
  const node: FlowNode =
    type === 'userTask'
      ? { id, type, name: '人工节点', approvers: '' }
      : { id, type, name: '条件节点', condition: '' }
  const endIdx = structure.nodes.findIndex((n) => n.type === 'end')
  if (endIdx >= 0) {
    structure.nodes.splice(endIdx, 0, node)
  } else {
    structure.nodes.push(node)
  }
  rebuildEdgesFromNodes()
}

function removeNode(idx: number) {
  structure.nodes.splice(idx, 1)
  rebuildEdgesFromNodes()
}

let dragIndex = -1
function onDragStart(idx: number, _ev: DragEvent) {
  dragIndex = idx
}
function onDrop(idx: number, _ev: DragEvent) {
  if (dragIndex < 0 || dragIndex === idx) return
  const node = structure.nodes[dragIndex]
  if (node.type === 'start' || node.type === 'end') {
    ElMessage.info('开始节点与结束节点不可调整顺序')
    dragIndex = -1
    return
  }
  if (idx === 0 || idx === structure.nodes.length - 1) {
    dragIndex = -1
    return
  }
  const arr = structure.nodes.slice()
  arr.splice(dragIndex, 1)
  let insertAt = idx
  if (dragIndex < idx) insertAt = idx - 1
  arr.splice(insertAt, 0, node)
  structure.nodes = arr
  rebuildEdgesFromNodes()
  dragIndex = -1
}

function nodeIcon(type: string) {
  switch (type) {
    case 'start':
      return 'VideoPlay'
    case 'end':
      return 'CircleCheck'
    case 'userTask':
      return 'User'
    case 'condition':
      return 'Operation'
    default:
      return 'Document'
  }
}

function nodeLabel(node: FlowNode) {
  const map: Record<string, string> = {
    start: '开始',
    userTask: '人工节点',
    condition: '条件节点',
    end: '结束'
  }
  return node.name || map[node.type] || '节点'
}

// ============ Schema ============
function addSchemaField() {
  formSchemaFields.value.push({
    key: '',
    type: 'string',
    label: '',
    required: false,
    default: ''
  })
}
function removeSchemaField(idx: number) {
  formSchemaFields.value.splice(idx, 1)
}

function buildSchemaFromFields(): any[] {
  return formSchemaFields.value
    .filter((f) => f.key.trim())
    .map((f) => ({
      key: f.key.trim(),
      type: f.type,
      label: f.label || f.key,
      required: !!f.required,
      default: f.default
    }))
}

function populateFieldsFromSchema(data: any[]) {
  if (!Array.isArray(data)) return
  formSchemaFields.value = data.map((f: any) => ({
    key: String(f.key ?? ''),
    type: String(f.type ?? 'string'),
    label: String(f.label ?? ''),
    required: !!f.required,
    default: String(f.default ?? '')
  }))
}

// ============ 同步文本 JSON <-> 结构化 ============
function syncStructureToText() {
  if (syncLock) return
  structureText.value = JSON.stringify(
    { nodes: structure.nodes, edges: structure.edges },
    null,
    2
  )
}

function syncStructureFromText() {
  if (!structureText.value.trim()) {
    resetStructureToDefault()
    syncStructureToText()
    return
  }
  try {
    const parsed = JSON.parse(structureText.value)
    if (Array.isArray(parsed.nodes)) {
      structure.nodes = parsed.nodes
      structure.edges = Array.isArray(parsed.edges) ? parsed.edges : []
      if (structure.edges.length === 0) rebuildEdgesFromNodes()
      ElMessage.success('结构已同步')
    } else {
      ElMessage.warning('JSON 中缺少 nodes 字段')
    }
  } catch (e: any) {
    ElMessage.warning('节点结构 JSON 解析失败：' + e.message)
  }
}

function syncFormSchemaToText() {
  if (syncLock) return
  formSchemaText.value = JSON.stringify(buildSchemaFromFields(), null, 2)
}

function syncFormSchemaFromText() {
  if (!formSchemaText.value.trim()) {
    formSchemaFields.value = []
    formSchemaText.value = '[]'
    return
  }
  try {
    const parsed = JSON.parse(formSchemaText.value)
    if (Array.isArray(parsed)) {
      populateFieldsFromSchema(parsed)
      ElMessage.success('Schema 已同步')
    } else if (parsed.fields && Array.isArray(parsed.fields)) {
      populateFieldsFromSchema(parsed.fields)
      ElMessage.success('Schema 已同步')
    } else {
      ElMessage.warning('JSON 格式不正确，应为字段数组')
    }
  } catch (e: any) {
    ElMessage.warning('表单 Schema JSON 解析失败：' + e.message)
  }
}

function syncNotifyToText() {
  if (syncLock) return
  notifyText.value = JSON.stringify(
    {
      enabled: notify.enabled,
      channels: notify.channels,
      events: notify.events,
      extraReceivers: notify.extraReceivers,
      webhookUrl: notify.webhookUrl,
      template: notify.template
    },
    null,
    2
  )
}

function syncNotifyFromText() {
  if (!notifyText.value.trim()) {
    notifyText.value = '{}'
    return
  }
  try {
    const parsed = JSON.parse(notifyText.value)
    notify.enabled = !!parsed.enabled
    notify.channels = Array.isArray(parsed.channels) ? parsed.channels : []
    notify.events = Array.isArray(parsed.events) ? parsed.events : []
    notify.extraReceivers = String(parsed.extraReceivers ?? '')
    notify.webhookUrl = String(parsed.webhookUrl ?? '')
    notify.template = String(parsed.template ?? '')
    ElMessage.success('通知配置已同步')
  } catch (e: any) {
    ElMessage.warning('通知配置 JSON 解析失败：' + e.message)
  }
}

watch(
  () => JSON.stringify(structure.nodes),
  () => {
    syncLock = true
    try {
      syncStructureToText()
    } finally {
      nextTick(() => (syncLock = false))
    }
  },
  { deep: true }
)
watch(
  () => JSON.stringify(formSchemaFields.value),
  () => {
    syncLock = true
    try {
      syncFormSchemaToText()
    } finally {
      nextTick(() => (syncLock = false))
    }
  },
  { deep: true }
)
watch(
  () =>
    JSON.stringify({
      enabled: notify.enabled,
      channels: notify.channels,
      events: notify.events,
      extraReceivers: notify.extraReceivers,
      webhookUrl: notify.webhookUrl,
      template: notify.template
    }),
  () => {
    syncLock = true
    try {
      syncNotifyToText()
    } finally {
      nextTick(() => (syncLock = false))
    }
  }
)

// ============ 打开 / 关闭 ============
function resetForm() {
  form.value = {
    id: '',
    name: '',
    icon: '',
    classifyId: '',
    description: '',
    enabled: true
  }
  resetStructureToDefault()
  formSchemaFields.value = [
    { key: 'title', type: 'string', label: '标题', required: true, default: '' },
    { key: 'amount', type: 'number', label: '金额', required: false, default: '' },
    { key: 'reason', type: 'textarea', label: '事由', required: true, default: '' }
  ]
  notify.enabled = true
  notify.channels = ['inbox', 'email']
  notify.events = ['submit', 'approve', 'reject']
  notify.extraReceivers = ''
  notify.webhookUrl = ''
  notify.template = '【${processName}】工单号 ${workOrderId} 已由 ${creator} 提交，当前节点：${nodeName}。'
  syncStructureToText()
  syncFormSchemaToText()
  syncNotifyToText()
  formActiveTab.value = 'basic'
}

function openCreate() {
  resetForm()
  formMode.value = 'create'
  formVisible.value = true
}

function parseStructure(raw: any): FlowStructure {
  if (!raw) return defaultStructure()
  let obj: any = raw
  if (typeof raw === 'string') {
    try {
      obj = JSON.parse(raw)
    } catch {
      return defaultStructure()
    }
  }
  if (Array.isArray(obj.nodes) && obj.nodes.length) {
    return {
      nodes: obj.nodes.map((n: any) => ({
        id: String(n.id ?? ''),
        type: (n.type as any) || 'userTask',
        name: String(n.name ?? ''),
        approvers: n.approvers ? String(n.approvers) : '',
        condition: n.condition ? String(n.condition) : ''
      })),
      edges: Array.isArray(obj.edges) ? obj.edges : []
    }
  }
  return defaultStructure()
}

function parseFormSchema(raw: any): SchemaField[] {
  if (!raw) return []
  let arr: any = raw
  if (typeof raw === 'string') {
    try {
      arr = JSON.parse(raw)
    } catch {
      return []
    }
  }
  if (Array.isArray(arr)) return arr
  if (Array.isArray(arr.fields)) return arr.fields
  return []
}

function parseNotify(raw: any): NotifyConfig {
  const def: NotifyConfig = {
    enabled: true,
    channels: ['inbox', 'email'],
    events: ['submit', 'approve', 'reject'],
    extraReceivers: '',
    webhookUrl: '',
    template: ''
  }
  if (!raw) return def
  let obj: any = raw
  if (typeof raw === 'string') {
    try {
      obj = JSON.parse(raw)
    } catch {
      return def
    }
  }
  return {
    enabled: obj.enabled !== undefined ? !!obj.enabled : def.enabled,
    channels: Array.isArray(obj.channels) ? obj.channels : def.channels,
    events: Array.isArray(obj.events) ? obj.events : def.events,
    extraReceivers: String(obj.extraReceivers ?? ''),
    webhookUrl: String(obj.webhookUrl ?? ''),
    template: String(obj.template ?? '')
  }
}

function openEdit(row: any) {
  form.value = {
    id: row.id,
    name: row.name,
    icon: row.icon || '',
    classifyId: row.classifyId || row.classify || '',
    description: row.description || '',
    enabled: !!row.enabled
  }
  const st = parseStructure(row.structure)
  structure.nodes = st.nodes
  structure.edges = st.edges.length ? st.edges : []
  if (structure.edges.length === 0) rebuildEdgesFromNodes()

  formSchemaFields.value = parseFormSchema(row.formSchema)
  const n = parseNotify(row.notify)
  notify.enabled = n.enabled
  notify.channels = n.channels
  notify.events = n.events
  notify.extraReceivers = n.extraReceivers
  notify.webhookUrl = n.webhookUrl
  notify.template = n.template

  syncStructureToText()
  syncFormSchemaToText()
  syncNotifyToText()
  formMode.value = 'edit'
  formActiveTab.value = 'basic'
  formVisible.value = true
}

async function doSubmit() {
  if (!form.value.name?.trim()) {
    ElMessage.warning('请输入流程名称')
    formActiveTab.value = 'basic'
    return
  }
  if (!structure.nodes.length) {
    ElMessage.warning('请至少配置一个节点结构')
    formActiveTab.value = 'structure'
    return
  }
  const start = structure.nodes.find((n) => n.type === 'start')
  const end = structure.nodes.find((n) => n.type === 'end')
  if (!start || !end) {
    ElMessage.warning('节点结构必须包含开始节点与结束节点')
    formActiveTab.value = 'structure'
    return
  }

  let structurePayload: FlowStructure = { nodes: structure.nodes, edges: structure.edges }
  try {
    if (structureText.value.trim()) {
      const parsed = JSON.parse(structureText.value)
      if (Array.isArray(parsed.nodes)) {
        structurePayload = parsed
      }
    }
  } catch {
    // fallback to structured data
  }

  let schemaPayload: any = buildSchemaFromFields()
  try {
    if (formSchemaText.value.trim()) {
      const parsed = JSON.parse(formSchemaText.value)
      schemaPayload = Array.isArray(parsed) ? parsed : parsed.fields || parsed
    }
  } catch {
    // fallback to structured data
  }

  let notifyPayload: any = {
    enabled: notify.enabled,
    channels: notify.channels,
    events: notify.events,
    extraReceivers: notify.extraReceivers,
    webhookUrl: notify.webhookUrl,
    template: notify.template
  }
  try {
    if (notifyText.value.trim()) {
      const parsed = JSON.parse(notifyText.value)
      notifyPayload = parsed
    }
  } catch {
    // fallback
  }

  submitting.value = true
  try {
    const payload = {
      name: form.value.name.trim(),
      icon: form.value.icon || '',
      classifyId: form.value.classifyId || '',
      classify: form.value.classifyId || '',
      description: form.value.description || '',
      structure: structurePayload,
      formSchema: schemaPayload,
      notify: notifyPayload
    }
    if (formMode.value === 'create') {
      const res = await ferryCreateProcess(payload)
      ElMessage.success('创建成功')
      if (res?.data?.id) {
        const id = (res.data as any).id
        try {
          await ferryToggleProcessEnabled(id, !!form.value.enabled)
        } catch {}
      }
    } else {
      await ferryUpdateProcess(form.value.id, payload)
      if (form.value.enabled !== undefined) {
        try {
          await ferryToggleProcessEnabled(form.value.id, !!form.value.enabled)
        } catch {}
      }
      ElMessage.success('更新成功')
    }
    formVisible.value = false
    loadList()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  } finally {
    submitting.value = false
  }
}

async function toggleEnabled(row: any) {
  try {
    await ferryToggleProcessEnabled(row.id, !row.enabled)
    ElMessage.success('操作成功')
    loadList()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

function cloneProcess(row: any) {
  ElMessageBox.confirm(`确认克隆流程 "${row.name}" 吗？`, '提示', { type: 'info' })
    .then(async () => {
      await ferryCloneProcess(row.id)
      ElMessage.success('克隆成功')
      loadList()
    })
    .catch(() => {})
}

function deleteProcess(row: any) {
  ElMessageBox.confirm(`确认删除流程 "${row.name}" 吗？该操作不可恢复。`, '提示', {
    type: 'warning',
    confirmButtonText: '确认删除',
    cancelButtonText: '取消'
  })
    .then(async () => {
      await ferryDeleteProcess(row.id)
      ElMessage.success('删除成功')
      loadList()
    })
    .catch(() => {})
}

onMounted(() => {
  loadClassifies()
  loadList()
})
</script>

<style scoped>
.ferry-page {
  padding: 16px;
}

.ferry-header {
  background: linear-gradient(90deg, #ecf5ff 0%, #f0f9eb 100%);
  border-radius: 8px;
  padding: 18px 24px;
  margin-bottom: 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
}

.ferry-title {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 4px 0;
}

.ferry-subtitle {
  font-size: 13px;
  color: #606266;
}

.ferry-stats {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 14px;
}

.stat-box {
  flex: 1 1 180px;
  background: #fff;
  border-radius: 8px;
  padding: 14px 18px;
  min-width: 160px;
  border-left: 4px solid #409eff;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}

.stat-box.stat-enabled {
  border-left-color: #67c23a;
}
.stat-box.stat-disabled {
  border-left-color: #909399;
}
.stat-box.stat-submits {
  border-left-color: #e6a23c;
}

.stat-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 6px;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}

.ferry-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.ferry-spacer {
  flex: 1;
}

.ferry-pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.ferry-form-tabs :deep(.el-tabs__content) {
  padding-top: 8px;
}

/* Flow chart */
.ferry-flow-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  gap: 12px;
  flex-wrap: wrap;
}
.ferry-flow-tip {
  color: #606266;
  font-size: 13px;
}
.ferry-flow-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.ferry-flow-canvas {
  background: linear-gradient(180deg, #fafbfc 0%, #f2f6fc 100%);
  border: 1px dashed #dcdfe6;
  border-radius: 8px;
  padding: 18px 12px;
  display: flex;
  align-items: stretch;
  justify-content: flex-start;
  gap: 0;
  overflow-x: auto;
  margin-bottom: 14px;
  min-height: 180px;
}

.ferry-flow-node {
  background: #fff;
  border-radius: 8px;
  border: 1px solid #dcdfe6;
  min-width: 220px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.ferry-flow-node.node-start {
  border-color: #67c23a;
  background: linear-gradient(180deg, #f0f9eb 0%, #fff 100%);
}
.ferry-flow-node.node-end {
  border-color: #909399;
  background: linear-gradient(180deg, #f4f4f5 0%, #fff 100%);
}
.ferry-flow-node.node-userTask {
  border-color: #409eff;
  background: linear-gradient(180deg, #ecf5ff 0%, #fff 100%);
}
.ferry-flow-node.node-condition {
  border-color: #e6a23c;
  background: linear-gradient(180deg, #fdf6ec 0%, #fff 100%);
}

.ferry-flow-node-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border-bottom: 1px solid #ebeef5;
  background: rgba(255, 255, 255, 0.6);
  cursor: move;
  user-select: none;
}
.ferry-flow-node-title {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
.ferry-flow-node-body {
  padding: 10px 12px;
  flex: 1;
}

.ferry-flow-arrow {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 6px;
  min-width: 40px;
  position: relative;
}
.ferry-flow-arrow-line {
  width: 100%;
  height: 2px;
  background: repeating-linear-gradient(90deg, #c0c4cc 0 6px, transparent 6px 10px);
}
.ferry-flow-arrow-icon {
  color: #909399;
  margin-left: -10px;
  background: #fff;
  border-radius: 50%;
}

.ferry-flow-json-head,
.ferry-schema-json-head,
.ferry-notify-json-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
  font-size: 13px;
  color: #606266;
}

.ferry-schema-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
  gap: 12px;
  flex-wrap: wrap;
}
.ferry-schema-tip {
  color: #606266;
  font-size: 13px;
}
.ferry-schema-table {
  margin-bottom: 14px;
}
</style>
