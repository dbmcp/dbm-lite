<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DB老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <div class="page-container">
    <div class="page-header">
      <span>服务器管理</span>
      <span class="page-sub">· SSH 连接与信息维护</span>
    </div>

    <!-- 顶部统计卡片 -->
    <div class="stat-row">
      <div class="stat-card stat-total">
        <div class="stat-num">{{ stats.total || 0 }}</div>
        <div class="stat-label">总服务器</div>
      </div>
      <div class="stat-card stat-active">
        <div class="stat-num">{{ stats.active || 0 }}</div>
        <div class="stat-label">已启用</div>
      </div>
      <div class="stat-card stat-ok">
        <div class="stat-num">{{ stats.connected || 0 }}</div>
        <div class="stat-label">连接成功</div>
      </div>
      <div class="stat-card stat-fail">
        <div class="stat-num">{{ stats.failed || 0 }}</div>
        <div class="stat-label">连接失败</div>
      </div>
    </div>

    <div class="card">
      <div class="toolbar">
        <el-input v-model="keyword" placeholder="搜索名称 / 主机 / 备注" clearable style="width:240px;" @change="loadList" />
        <el-select v-model="filterEnv" placeholder="环境" clearable style="width:120px;" @change="loadList">
          <el-option label="开发" value="dev" />
          <el-option label="测试" value="test" />
          <el-option label="生产" value="prod" />
        </el-select>
        <el-select v-model="filterStatus" placeholder="状态" clearable style="width:120px;" @change="loadList">
          <el-option label="启用" value="active" />
          <el-option label="禁用" value="inactive" />
        </el-select>
        <el-button type="primary" :icon="SearchIcon" @click="loadList">搜索</el-button>
        <el-button @click="loadList"><span class="refresh-icon">⟳</span>刷新</el-button>
        <el-button type="primary" :icon="PlusIcon" @click="openDialog()">添加服务器</el-button>
        <ColumnToggle v-model="colVisible" :columns="columns" />
      </div>

      <el-table :data="list" style="width:100%;" v-loading="loading" stripe :default-sort="{ prop: 'createdAt', order: 'descending' }">
        <el-table-column type="index" label="#" width="60" show-overflow-tooltip />
        <el-table-column prop="name" label="名称" min-width="160" v-if="colVisible.name" show-overflow-tooltip>
          <template #default="{ row }">
            <div class="name-cell">
              <el-icon class="host-icon"><Monitor /></el-icon>
              <span class="name-text">{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="连接信息" min-width="220" v-if="colVisible.host" show-overflow-tooltip>
          <template #default="{ row }">
            <div class="conn-cell">
              <div class="conn-addr">{{ row.host }}:{{ row.port }}</div>
              <div class="conn-sub">{{ row.os || '-' }} · {{ row.username || '-' }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="所属业务" min-width="140" v-if="colVisible.businessId" show-overflow-tooltip>
          <template #default="{ row }">
            {{ businessMap[row.businessId]?.name || row.projectId || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="env" label="环境" width="110" v-if="colVisible.env" show-overflow-tooltip>
          <template #default="{ row }">
            <el-tag v-if="row.env === 'prod'" type="danger" size="small">生产</el-tag>
            <el-tag v-else-if="row.env === 'test'" type="warning" size="small">测试</el-tag>
            <el-tag v-else size="small">开发</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="authType" label="认证" width="100" v-if="colVisible.authType" show-overflow-tooltip>
          <template #default="{ row }">
            <el-tag v-if="row.authType === 'key'" type="info" size="small">密钥</el-tag>
            <el-tag v-else type="success" size="small">密码</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="110" v-if="colVisible.status" show-overflow-tooltip>
          <template #default="{ row }">
            <el-switch
              :model-value="row.status === 'active' || !row.status ? true : false"
              active-color="#13ce66"
              inactive-color="#c0c4cc"
              size="small"
              inline-prompt
              active-text="启用"
              inactive-text="禁用"
              @change="onToggle(row)"
            />
          </template>
        </el-table-column>
        <el-table-column label="连接状态" min-width="160" v-if="colVisible.connStatus" show-overflow-tooltip>
          <template #default="{ row }">
            <el-tag v-if="row.connStatus === 'ok'" type="success" size="small" effect="light">
              <span class="conn-dot conn-dot-ok"></span>
              成功 {{ row.connLatencyMs ? '(' + row.connLatencyMs + 'ms)' : '' }}
            </el-tag>
            <el-tag v-else-if="row.connStatus === 'fail'" type="danger" size="small" effect="light">
              <span class="conn-dot conn-dot-fail"></span>
              失败
            </el-tag>
            <el-tag v-else type="info" size="small" effect="plain">未测试</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="资源" min-width="200" v-if="colVisible.resource" show-overflow-tooltip>
          <template #default="{ row }">
            <div class="resource-sub">
              <span v-if="row.cpuCores">CPU {{ row.cpuCores }}核</span>
              <span v-if="row.memoryGB">内存 {{ row.memoryGB }}GB</span>
              <span v-if="row.diskGB">磁盘 {{ row.diskGB }}GB</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" min-width="180" v-if="colVisible.createdAt" show-overflow-tooltip />
        <el-table-column label="操作" min-width="320" fixed="right" show-overflow-tooltip>
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="openDetail(row)">详情</el-button>
            <el-button size="small" type="success" @click="testConnection(row)">测试</el-button>
            <el-button size="small" type="warning" @click="openTerminal(row)">终端</el-button>
            <el-button size="small" @click="openDialog(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-row">
        <el-pagination
          v-model:current-page="current"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[100, 200, 500, 1000]"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="loadList"
          @size-change="loadList"
        />
      </div>
    </div>

    <!-- 添加 / 编辑 对话框 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑服务器' : '添加服务器'" width="640px">
      <el-form :model="form" label-width="100px" ref="formRef" :rules="rules">
        <el-form-item label="所属业务" prop="businessId">
          <el-select v-model="form.businessId" style="width:100%;" @change="onBizChange" placeholder="请选择业务">
            <el-option v-for="b in businesses" :key="b.businessId" :label="(b.projectName ? b.projectName + ' / ' : '') + (b.code ? b.code + ' - ' : '') + b.name" :value="b.businessId" />
          </el-select>
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="名称" prop="name"><el-input v-model="form.name" /></el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="环境" prop="env">
              <el-select v-model="form.env" style="width:100%;">
                <el-option label="开发" value="dev" />
                <el-option label="测试" value="test" />
                <el-option label="生产" value="prod" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="14">
            <el-form-item label="主机" prop="host"><el-input v-model="form.host" placeholder="如 192.168.1.100" /></el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="端口" prop="port"><el-input-number v-model="form.port" :min="1" :max="65535" style="width:100%;" /></el-form-item>
          </el-col>
          <el-col :span="4">
            <el-form-item label="超时(s)"><el-input-number v-model="form.timeout" :min="1" :max="600" style="width:100%;" /></el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="用户名" prop="username"><el-input v-model="form.username" /></el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="认证方式" prop="authType">
              <el-select v-model="form.authType" style="width:100%;">
                <el-option label="密码" value="password" />
                <el-option label="密钥" value="key" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item v-if="form.authType === 'password'" label="密码" :prop="isEdit ? undefined : 'password'">
          <el-input v-model="form.password" type="password" show-password :placeholder="isEdit ? '不修改请留空' : '请输入密码'" />
        </el-form-item>
        <el-form-item v-if="form.authType === 'key'" label="私钥内容" :prop="isEdit ? undefined : 'privateKey'">
          <el-input v-model="form.privateKey" type="textarea" :rows="5" :placeholder="'-----BEGIN RSA PRIVATE KEY-----'" />
        </el-form-item>
        <el-form-item v-if="form.authType === 'key'" label="密钥口令">
          <el-input v-model="form.keyPassphrase" type="password" show-password placeholder="如私钥无需口令可留空" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="warning" @click="testCurrent">测试连接</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <!-- 服务器详情 / 执行终端对话框 -->
    <el-dialog v-model="terminalVisible" :title="'服务器详情 - ' + (currentServer?.name || '')" width="780px">
      <template v-if="currentServer">
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="名称">{{ currentServer.name }}</el-descriptions-item>
          <el-descriptions-item label="主机">{{ currentServer.host }}:{{ currentServer.port }}</el-descriptions-item>
          <el-descriptions-item label="用户">{{ currentServer.username }}</el-descriptions-item>
          <el-descriptions-item label="认证">{{ currentServer.authType === 'key' ? '密钥' : '密码' }}</el-descriptions-item>
          <el-descriptions-item label="系统">{{ currentServer.os || '-' }} / {{ currentServer.arch || '-' }}</el-descriptions-item>
          <el-descriptions-item label="版本">{{ currentServer.version || '-' }}</el-descriptions-item>
          <el-descriptions-item label="CPU">{{ currentServer.cpuCores ? currentServer.cpuCores + ' 核' : '-' }}</el-descriptions-item>
          <el-descriptions-item label="内存">{{ currentServer.memoryGB ? currentServer.memoryGB + ' GB' : '-' }}</el-descriptions-item>
          <el-descriptions-item label="磁盘">{{ currentServer.diskGB ? currentServer.diskGB + ' GB' : '-' }}</el-descriptions-item>
          <el-descriptions-item label="连接">{{ currentServer.connStatus || '-' }}{{ currentServer.connLatencyMs ? ' (' + currentServer.connLatencyMs + 'ms)' : '' }}</el-descriptions-item>
          <el-descriptions-item label="备注" :span="2">{{ currentServer.remark || '-' }}</el-descriptions-item>
        </el-descriptions>
        <el-divider>执行命令</el-divider>
        <div class="terminal-box">
          <div class="terminal-history">
            <div v-for="(h, idx) in history" :key="idx" class="term-row">
              <div class="term-cmd">$ {{ h.cmd }}</div>
              <pre class="term-out">{{ h.stdout }}</pre>
              <pre v-if="h.stderr" class="term-err">{{ h.stderr }}</pre>
            </div>
            <div v-if="termLoading" class="term-wait">执行中…</div>
          </div>
          <div class="terminal-input-row">
            <el-select v-model="quickCmd" placeholder="快速选择" size="default" style="width:160px;" @change="val => { if (val) execCmd(val); }">
              <el-option label="查看CPU" value="cat /proc/cpuinfo | head -n 10" />
              <el-option label="查看内存" value="free -h" />
              <el-option label="查看磁盘" value="df -h" />
              <el-option label="系统负载" value="uptime" />
              <el-option label="当前时间" value="date" />
              <el-option label="网络连接" value="netstat -tnl" />
              <el-option label="进程信息" value="ps aux" />
            </el-select>
            <el-input v-model="cmdInput" placeholder="请输入命令..." style="flex:1;margin-left:10px;" @keyup.enter="execCmd(cmdInput)" />
            <el-button type="primary" :loading="termLoading" @click="execCmd(cmdInput)">执行</el-button>
          </div>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Search as SearchIcon, Refresh as RefreshIcon, Plus as PlusIcon, Monitor } from '@element-plus/icons-vue'
import {
  createServer, updateServer, deleteServer, testServer, testServerById,
  listServers, statsServers, execCommand, toggleServer, getServer
} from '@/api/server'
import { listAllBusinesses } from '@/api/business'
import ColumnToggle from '@/components/ColumnToggle.vue'

const columns = [
  { key: 'name', label: '名称' },
  { key: 'host', label: '连接信息' },
  { key: 'businessId', label: '所属业务' },
  { key: 'env', label: '环境' },
  { key: 'authType', label: '认证' },
  { key: 'status', label: '状态' },
  { key: 'connStatus', label: '连接状态' },
  { key: 'resource', label: '资源' },
  { key: 'createdAt', label: '创建时间' }
]
const colVisible = reactive<Record<string, boolean>>({
  name: true, host: true, businessId: true, env: true, authType: true,
  status: true, connStatus: true, resource: true, createdAt: true
})

const loading = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const current = ref(1)
const pageSize = ref(100)
const keyword = ref('')
const filterEnv = ref('')
const filterStatus = ref('')
const dialogVisible = ref(false)
const terminalVisible = ref(false)
const isEdit = ref(false)
const editId = ref('')
const formRef = ref<FormInstance>()
const businesses = ref<any[]>([])
const businessMap = ref<Record<string, any>>({})
const stats = ref<Record<string, any>>({ total: 0, active: 0, connected: 0, failed: 0 })

const currentServer = ref<any>(null)
const cmdInput = ref('')
const quickCmd = ref('')
const history = ref<{ cmd: string; stdout: string; stderr: string }[]>([])
const termLoading = ref(false)

const form = reactive({
  projectId: '', businessId: '', name: '', env: 'dev', host: '', port: 22, username: '', authType: 'password', password: '', privateKey: '', keyPassphrase: '', remark: '', timeout: 30
})

const rules: FormRules = {
  businessId: [{ required: true, message: '请选择所属业务' }],
  name: [{ required: true, message: '请输入名称' }],
  host: [{ required: true, message: '请输入主机' }],
  port: [{ required: true, message: '请输入端口' }],
  username: [{ required: true, message: '请输入用户名' }],
  authType: [{ required: true, message: '请选择认证方式' }]
}

async function loadStats() {
  try {
    const s: any = await statsServers()
    stats.value = s || {}
  } catch (e) { /* noop */ }
}

async function loadBusinesses() {
  try {
    const r: any = await listAllBusinesses()
    businesses.value = r?.list || r || []
    const m: Record<string, any> = {}
    for (const b of businesses.value) {
      m[b.businessId] = b
    }
    businessMap.value = m
  } catch (e) { /* noop */ }
}

async function loadList() {
  loading.value = true
  try {
    const r: any = await listServers(current.value, pageSize.value, keyword.value, filterEnv.value, filterStatus.value)
    list.value = r?.list || r || []
    total.value = r?.total || list.value.length
  } finally {
    loading.value = false
  }
}

function onBizChange(bizId: string) {
  const b = businesses.value.find((x: any) => x.businessId === bizId)
  if (b) form.projectId = b.projectId
}

function openDialog(row?: any) {
  if (row) {
    isEdit.value = true
    editId.value = row.serverId
    Object.assign(form, {
      projectId: row.projectId || '',
      businessId: row.businessId || '',
      name: row.name, env: row.env || 'dev', host: row.host, port: row.port || 22,
      username: row.username, authType: row.authType || 'password', password: '', privateKey: '', keyPassphrase: '', remark: row.remark || '', timeout: row.timeout || 30
    })
  } else {
    isEdit.value = false
    editId.value = ''
    Object.assign(form, {
      projectId: '', businessId: '', name: '', env: 'dev', host: '', port: 22, username: '', authType: 'password', password: '', privateKey: '', keyPassphrase: '', remark: '', timeout: 30
    })
  }
  dialogVisible.value = true
}

async function save() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      const payload: any = { ...form }
      if (isEdit.value) {
        if (form.authType === 'password' && !form.password) delete payload.password
        if (form.authType === 'key' && !form.privateKey) delete payload.privateKey
        await updateServer(editId.value, payload)
        ElMessage.success('更新成功')
      } else {
        await createServer(payload)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      loadList()
      loadStats()
    } catch (e: any) {
      ElMessage.error(e?.message || '保存失败')
    }
  })
}

async function testCurrent() {
  try {
    await testServer({
      host: form.host, port: form.port, username: form.username, authType: form.authType,
      password: form.password, privateKey: form.privateKey, keyPassphrase: form.keyPassphrase
    })
    ElMessage.success('连接测试成功')
  } catch (e: any) {
      ElMessage.error(e?.message || '连接测试失败')
  }
}

async function testConnection(row: any) {
  try {
    const r: any = await testServerById(row.serverId)
    ElMessage.success('连接测试成功，耗时 ' + (r?.latencyMs || 0) + 'ms')
    loadList()
  } catch (e: any) {
    ElMessage.error(e?.message || '连接测试失败')
  }
}

async function onToggle(row: any) {
  try {
    await toggleServer(row.serverId)
    ElMessage.success('状态已更新')
  } catch (e: any) {
    ElMessage.error(e?.message || '切换失败')
    loadList()
  }
}

async function openDetail(row: any) {
  currentServer.value = row
  history.value = []
  terminalVisible.value = true
}

async function openTerminal(row: any) {
  currentServer.value = row
  history.value = []
  terminalVisible.value = true
}

async function execCmd(cmd: string) {
  if (!cmd || !currentServer.value) return
  const toSend = cmd
  cmdInput.value = ''
  quickCmd.value = ''
  termLoading.value = true
  try {
    const r: any = await execCommand(currentServer.value.serverId, toSend)
    history.value.push({ cmd: toSend, stdout: r?.stdout || '', stderr: r?.stderr || '' })
    // 刷新最新服务器信息
    try {
      const latest: any = await getServer(currentServer.value.serverId)
      if (latest) currentServer.value = latest
    } catch (e) { /* noop */ }
  } catch (e: any) {
    ElMessage.error(e?.message || '执行失败')
  } finally {
    termLoading.value = false
  }
}

function handleDelete(row: any) {
  ElMessageBox.confirm('确认删除该服务器？此操作不可撤销', '确认', { type: 'warning' }).then(async () => {
    try {
      await deleteServer(row.serverId)
      ElMessage.success('已删除')
      loadList()
      loadStats()
    } catch (e: any) {
      ElMessage.error(e?.message || '删除失败')
    }
  }).catch(() => {})
}

onMounted(async () => {
  await loadBusinesses()
  loadList()
  loadStats()
})
</script>

<style scoped>
.page-container {
  padding: 18px 20px;
  color: #303133;
}
.page-header {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 14px;
}
.page-sub {
  font-size: 13px;
  color: #909399;
  font-weight: 400;
  margin-left: 8px;
}
.stat-row {
  display: flex;
  gap: 16px;
  margin-bottom: 14px;
}
.stat-card {
  flex: 1;
  padding: 16px 20px;
  border-radius: 6px;
  color: #fff;
  min-height: 84px;
}
.stat-card.stat-total { background: linear-gradient(135deg, #409EFF, #66b1ff); }
.stat-card.stat-active { background: linear-gradient(135deg, #67C23A, #95d475); }
.stat-card.stat-ok { background: linear-gradient(135deg, #13ce66, #67e2a0); }
.stat-card.stat-fail { background: linear-gradient(135deg, #F56C6C, #f89898); }
.stat-num { font-size: 28px; font-weight: 700; }
.stat-label { font-size: 13px; opacity: .9; margin-top: 4px; }
.card {
  background: #fff;
  border-radius: 6px;
  box-shadow: 0 1px 4px rgba(0,0,0,.06);
  padding: 16px 18px;
}
.toolbar {
  display: flex;
  gap: 10px;
  align-items: center;
  margin-bottom: 14px;
  flex-wrap: wrap;
}
.name-cell {
  display: flex;
  align-items: center;
}
.host-icon {
  color: #409EFF;
  margin-right: 6px;
}
.name-text { font-weight: 500; }
.conn-cell .conn-addr { font-family: Menlo, Consolas, monospace; font-size: 13px; }
.conn-cell .conn-sub { font-size: 12px; color: #909399; margin-top: 2px; }
.resource-sub {
  font-size: 12px;
  color: #606266;
  margin-top: 4px;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.pagination-row {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
.conn-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 6px;
  vertical-align: middle;
}
.conn-dot.conn-dot-ok { background: #67C23A; }
.conn-dot.conn-dot-fail { background: #F56C6C; }
.terminal-box {
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  background: #fafafa;
}
.terminal-history {
  max-height: 300px;
  overflow-y: auto;
  padding: 12px;
  background: #1e1e1e;
  color: #dcdcdc;
  font-family: Menlo, Consolas, monospace;
  font-size: 12px;
}
.terminal-history .term-row { margin-bottom: 10px; }
.terminal-history .term-cmd { color: #4fc3f7; }
.terminal-history .term-out { color: #e0e0e0; white-space: pre-wrap; margin: 4px 0; }
.terminal-history .term-err { color: #ff6b6b; white-space: pre-wrap; }
.terminal-history .term-wait { color: #ffb74d; }
.terminal-input-row {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  border-top: 1px solid #e4e7ed;
  background: #fff;
}
</style>
