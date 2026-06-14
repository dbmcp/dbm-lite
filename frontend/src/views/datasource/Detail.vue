<template>
  <div class="page-container">
    <div class="page-header">
      <div class="title-row">
        <el-button :icon="ArrowLeft" text @click="goBack" style="margin-right:8px">返回</el-button>
        <span class="title-text">数据源详情</span>
      </div>
      <div class="header-actions">
        <el-button :icon="RefreshIcon" @click="loadDetail" :loading="loading">刷新</el-button>
        <el-button type="primary" :icon="EditIcon" v-if="userStore.isAdmin" @click="openEdit">编辑</el-button>
        <el-button type="danger" :icon="DeleteIcon" v-if="userStore.isAdmin" @click="handleDelete">删除</el-button>
        <el-button type="success" :icon="Connection" @click="testConnection" :loading="testing">测试连接</el-button>
      </div>
    </div>

    <el-empty v-if="notFound" description="数据源不存在或已被删除" :image-size="120">
      <el-button type="primary" @click="goBack">返回列表</el-button>
    </el-empty>

    <template v-else-if="detail">
      <div class="summary-card">
        <div class="summary-left">
          <div class="ds-color-bar" :style="{ background: colorOf(datasource.colorLabel) }"></div>
          <div>
            <div class="ds-name">{{ datasource.name }}</div>
            <div class="ds-meta">
              <el-tag :type="dbTypeTagType(datasource.dbType)" size="small">
                {{ dbTypeLabel(datasource.dbType) }}
              </el-tag>
              <el-tag v-if="datasource.env === 'prod'" type="danger" size="small">生产</el-tag>
              <el-tag v-else-if="datasource.env === 'stage'" type="warning" size="small">预发</el-tag>
              <el-tag v-else-if="datasource.env === 'test'" type="success" size="small">测试</el-tag>
              <el-tag v-else size="small">开发</el-tag>
              <el-tag v-if="datasource.readOnly" type="info" size="small">只读</el-tag>
              <span v-if="datasource.tags" class="tags">标签: {{ datasource.tags }}</span>
            </div>
          </div>
        </div>

        <div class="summary-right">
          <div class="status-block" :class="{ 'ok': detail.connectionOK, 'fail': !detail.connectionOK && datasource.connStatus, 'none': !datasource.connStatus }">
            <div class="status-label">连接状态</div>
            <div class="status-value">
              <span class="status-dot" :class="{ 'status-ok': detail.connectionOK, 'status-fail': !detail.connectionOK && datasource.connStatus }"></span>
              <span>{{ detail.connectionOK ? '可连接' : (datasource.connStatus ? '连接失败' : '未测试') }}</span>
            </div>
            <div v-if="detail.latencyMs" class="status-sub">延迟 {{ detail.latencyMs }} ms</div>
            <div v-if="detail.version" class="status-sub">版本 {{ detail.version }}</div>
            <div v-if="detail.lastTestAt" class="status-sub">最近测试 {{ formatTime(detail.lastTestAt) }}</div>
          </div>
        </div>
      </div>

      <el-row :gutter="16">
        <el-col :span="12">
          <div class="info-card">
            <div class="card-title">基本信息</div>
            <el-descriptions :column="1" border size="default">
              <el-descriptions-item label="数据源 ID"><span class="mono">{{ datasource.datasourceId }}</span></el-descriptions-item>
              <el-descriptions-item label="数据库类型">{{ dbTypeLabel(datasource.dbType) }}</el-descriptions-item>
              <el-descriptions-item label="地址">
                <span v-if="datasource.dbType === 'sqlite'" class="mono">{{ datasource.filePath || '(内存库)' }}</span>
                <span v-else class="mono">{{ datasource.host }}:{{ datasource.port }}</span>
              </el-descriptions-item>
              <el-descriptions-item label="默认数据库">{{ datasource.defaultDatabase || '-' }}</el-descriptions-item>
              <el-descriptions-item label="用户名">{{ datasource.username || '-' }}</el-descriptions-item>
              <el-descriptions-item label="密码">{{ detail.passwordSet ? '••••••（已设置，脱敏展示）' : '未设置' }}</el-descriptions-item>
              <el-descriptions-item v-if="datasource.charset" label="字符集">{{ datasource.charset }}</el-descriptions-item>
              <el-descriptions-item v-if="datasource.timezone" label="时区">{{ datasource.timezone }}</el-descriptions-item>
              <el-descriptions-item v-if="datasource.sslMode" label="SSL">{{ datasource.sslMode }}</el-descriptions-item>
            </el-descriptions>
          </div>
        </el-col>

        <el-col :span="12">
          <div class="info-card">
            <div class="card-title">组织与时间</div>
            <el-descriptions :column="1" border size="default">
              <el-descriptions-item label="环境">{{ envLabel(datasource.env) }}</el-descriptions-item>
              <el-descriptions-item label="关联业务">{{ datasource.businessId || '-' }}</el-descriptions-item>
              <el-descriptions-item label="关联项目">{{ datasource.projectId || '-' }}</el-descriptions-item>
              <el-descriptions-item label="创建者">{{ datasource.createdBy || '-' }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ formatTime(datasource.createdAt) }}</el-descriptions-item>
              <el-descriptions-item label="更新时间">{{ formatTime(datasource.updatedAt) }}</el-descriptions-item>
              <el-descriptions-item label="状态">{{ datasource.status === 'active' ? '启用' : '禁用' }}</el-descriptions-item>
              <el-descriptions-item v-if="datasource.remark" label="备注">{{ datasource.remark }}</el-descriptions-item>
            </el-descriptions>
          </div>
        </el-col>
      </el-row>

      <div class="info-card">
        <div class="card-title">
          <span>数据库列表</span>
          <el-button size="small" type="primary" plain @click="loadDatabases" :loading="dbsLoading" style="margin-left:12px">刷新</el-button>
        </div>
        <el-empty v-if="databases.length === 0 && !dbsLoading" description="暂无数据库信息" />
        <el-table v-else :data="databases" style="width:100%" size="default" v-loading="dbsLoading">
          <el-table-column prop="name" label="数据库名称" min-width="180" />
          <el-table-column prop="sizeMb" label="大小 (MB)" width="140">
            <template #default="{ row }">
              <span>{{ row.sizeMb != null ? row.sizeMb : '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="tableCount" label="表数量" width="120">
            <template #default="{ row }">
              <span>{{ row.tableCount != null ? row.tableCount : '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="140">
            <template #default="{ row }">
              <el-button size="small" @click="goWorkbench(row.name)">进入 SQL</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </template>

    <!-- 编辑弹窗（复用 List 中的编辑逻辑） -->
    <EditDialog
      v-if="editVisible"
      :datasource-id="datasourceId"
      @close="editVisible = false"
      @saved="onSaved"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowLeft, Refresh as RefreshIcon, Edit as EditIcon, Delete as DeleteIcon, Connection
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { useDatasourceStore } from '@/stores/datasource'
import {
  getDatasource, updateDatasource, deleteDatasource, testConnectionById, listDatabases, listTables
} from '@/api/datasource'
import EditDialog from './components/EditDialog.vue'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const dsStore = useDatasourceStore()

const datasourceId = String(route.params.id || '')
const loading = ref(false)
const testing = ref(false)
const dbsLoading = ref(false)
const notFound = ref(false)

const datasource = ref<any>({
  datasourceId: '', name: '', dbType: '', host: '', port: 0, username: '',
  defaultDatabase: '', filePath: '', charset: '', timezone: '', sslMode: '',
  readOnly: false, colorLabel: '', version: '', tags: '', env: '',
  remark: '', status: '', createdBy: '', createdAt: '', updatedAt: '',
  lastConnTestAt: '', connStatus: '', connLatencyMs: 0
})

const detail = reactive<any>({
  connectionOK: false,
  lastTestAt: '',
  latencyMs: 0,
  version: '',
  passwordSet: false
})

const databases = ref<any[]>([])
const editVisible = ref(false)

function colorOf(label: string) {
  const map: Record<string, string> = {
    blue: '#409eff', green: '#67c23a', red: '#f56c6c',
    yellow: '#e6a23c', purple: '#8e44ad', orange: '#e67e22', gray: '#909399'
  }
  return map[label] || '#409eff'
}

function dbTypeLabel(type: string) {
  const m: Record<string, string> = { mysql: 'MySQL', tidb: 'TiDB', sqlite: 'SQLite' }
  return m[type] || type
}

function dbTypeTagType(type: string) {
  const m: Record<string, string> = { mysql: 'primary', tidb: 'success', sqlite: 'info' }
  return m[type] || ''
}

function envLabel(env: string) {
  const m: Record<string, string> = { prod: '生产', stage: '预发', test: '测试', dev: '开发' }
  return m[env] || env || '-'
}

function formatTime(t: any) {
  if (!t) return '-'
  const d = new Date(t)
  if (isNaN(d.getTime())) return String(t)
  return d.toLocaleString('zh-CN', { hour12: false })
}

async function loadDetail() {
  if (!datasourceId) return
  loading.value = true
  try {
    const ds: any = await getDatasource(datasourceId)
    if (!ds || !ds.datasourceId) {
      notFound.value = true
      return
    }
    Object.assign(datasource.value, ds)
    detail.connectionOK = ds.connStatus === 'ok'
    detail.lastTestAt = ds.lastConnTestAt || ''
    detail.latencyMs = ds.connLatencyMs || 0
    detail.version = ds.version || ''
    detail.passwordSet = !!(ds.password) || true

    if (detail.connectionOK) {
      loadDatabases()
    }
  } catch (e: any) {
    if (e?.status === 404) {
      notFound.value = true
    } else {
      ElMessage.error(e?.message || '加载数据源信息失败')
    }
  } finally {
    loading.value = false
  }
}

async function loadDatabases() {
  if (!datasourceId) return
  dbsLoading.value = true
  try {
    const dbs: any = await listDatabases(datasourceId)
    const list = Array.isArray(dbs) ? dbs : (dbs?.list || [])
    const enriched: any[] = []
    for (const name of list) {
      let tableCount: number | null = null
      let sizeMb: number | null = null
      try {
        const tables = await listTables(datasourceId, name)
        const tArr = Array.isArray(tables) ? tables : []
        tableCount = tArr.length
        if (tArr.length > 0 && tArr[0].sizeMb != null) {
          sizeMb = tArr.reduce((sum: number, t: any) => sum + (Number(t.sizeMb) || 0), 0)
        }
      } catch {}
      enriched.push({ name, tableCount, sizeMb: sizeMb != null ? Number(sizeMb.toFixed(2)) : null })
    }
    databases.value = enriched
  } catch (e: any) {
    ElMessage.error(e?.message || '加载数据库列表失败')
  } finally {
    dbsLoading.value = false
  }
}

async function testConnection() {
  testing.value = true
  try {
    const res: any = await testConnectionById(datasourceId)
    const ok = !!res?.success
    detail.connectionOK = ok
    detail.latencyMs = res?.latencyMs || 0
    detail.version = res?.version || ''
    detail.lastTestAt = new Date().toISOString()
    datasource.value.connStatus = ok ? 'ok' : 'fail'
    datasource.value.connLatencyMs = res?.latencyMs || 0
    datasource.value.version = res?.version || ''
    if (ok) {
      ElMessage.success(`连接成功${res?.latencyMs ? '（' + res.latencyMs + ' ms）' : ''}`)
      loadDatabases()
    } else {
      ElMessage.error(res?.message || '连接失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '连接测试失败')
  } finally {
    testing.value = false
  }
}

function openEdit() {
  editVisible.value = true
}

function onSaved() {
  editVisible.value = false
  loadDetail()
  ElMessage.success('更新成功')
}

function handleDelete() {
  ElMessageBox.confirm(
    `确认删除数据源「${datasource.value.name}」？删除后不可恢复。`,
    '⚠️ 删除确认',
    { type: 'warning', confirmButtonText: '确认删除', cancelButtonText: '取消', distinguishCancelAndClose: true }
  ).then(async () => {
    try {
      await ElMessageBox.prompt(
        `请输入数据源名称的前 4 个字符以二次确认（${(datasource.value.name || '').slice(0, 4)}）`,
        '🔒 二次确认',
        {
          inputPattern: new RegExp('^' + (datasource.value.name || '').slice(0, 4).replace(/[.*+?^${}()|[\]\\]/g, '\\$&')),
          inputErrorMessage: '输入不匹配，请重新输入'
        }
      )
    } catch { return }
    try {
      await deleteDatasource(datasourceId)
      ElMessage.success('删除成功')
      goBack()
    } catch (e: any) {
      ElMessage.error(e?.message || '删除失败')
    }
  }).catch(() => {})
}

function goBack() {
  router.push('/datasources')
}

function goWorkbench(dbName: string) {
  dsStore.setDatasource(datasourceId, datasource.value.name)
  router.push({ path: '/sql/workbench', query: { db: dbName } })
}

onMounted(() => {
  loadDetail()
})
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
}
.title-row { display: flex; align-items: baseline; }
.title-text { font-size: 20px; font-weight: 600; color: #303133; margin-left: 4px; }
.header-actions { display: flex; gap: 8px; align-items: center; }

.summary-card {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  background: linear-gradient(135deg, #fff 0%, #f4f8ff 100%);
  border: 1px solid #e4e7ed;
  border-radius: 10px;
  padding: 18px 20px;
  margin-bottom: 14px;
}
.summary-left { display: flex; align-items: flex-start; gap: 14px; }
.ds-color-bar {
  width: 6px; height: 52px; border-radius: 3px; flex-shrink: 0;
}
.ds-name { font-size: 22px; font-weight: 600; color: #303133; }
.ds-meta { margin-top: 8px; display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.tags { color: #909399; font-size: 13px; }
.summary-right { min-width: 240px; }
.status-block {
  background: #fff; border: 1px solid #e4e7ed; border-radius: 8px;
  padding: 12px 16px; text-align: right;
}
.status-block.ok { border-color: #67c23a; background: #f0f9eb; }
.status-block.fail { border-color: #f56c6c; background: #fef0f0; }
.status-label { color: #909399; font-size: 12px; }
.status-value {
  font-size: 16px; font-weight: 600; margin-top: 4px;
  display: flex; justify-content: flex-end; align-items: center; gap: 6px;
}
.status-sub { font-size: 12px; color: #606266; margin-top: 2px; }
.status-dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; background: #c0c4cc; }
.status-ok { background: #67c23a; }
.status-fail { background: #f56c6c; }

.info-card {
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 14px 16px;
  margin-bottom: 14px;
}
.card-title {
  font-size: 15px; font-weight: 600; color: #303133;
  padding-bottom: 10px; margin-bottom: 4px;
  border-bottom: 1px solid #ebeef5;
}
.mono { font-family: Menlo, Monaco, Consolas, monospace; }
</style>
