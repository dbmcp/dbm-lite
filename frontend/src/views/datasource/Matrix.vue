<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DB老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <div class="page-container">
    <div class="page-header">
      <div class="title-row">
        <span class="title-text">数据源矩阵视图</span>
        <span class="subtitle-text">按环境分组展示</span>
      </div>
      <div class="header-actions">
        <el-button :icon="RefreshIcon" @click="loadList">刷新</el-button>
        <el-button type="primary" :icon="PlusIcon" v-if="userStore.isAdmin" @click="$emit('create')">新建数据源</el-button>
      </div>
    </div>

    <div class="matrix-container" v-loading="loading">
      <div class="matrix-column" v-for="env in environments" :key="env.key">
        <div class="column-header" :class="env.key">
          <div class="column-title">
            <span class="env-icon">{{ env.icon }}</span>
            <span>{{ env.label }}</span>
          </div>
          <span class="column-count">{{ getEnvDatasources(env.key).length }}</span>
        </div>
        
        <div class="column-content">
          <div 
            v-for="ds in getEnvDatasources(env.key)" 
            :key="ds.datasourceId" 
            class="ds-card"
            @mouseenter="showActions(ds.datasourceId)"
            @mouseleave="hideActions"
          >
            <div class="card-header" :style="{ borderLeftColor: colorOf(ds.colorLabel) }">
              <div class="card-title-row">
                <el-tooltip :content="connStatusTooltip(ds)" placement="top">
                  <span class="status-dot" :class="statusClass(ds)"></span>
                </el-tooltip>
                <span class="ds-name">{{ ds.name }}</span>
              </div>
              <el-tag :type="dbTypeTagType(ds.dbType)" size="small">{{ dbTypeLabel(ds.dbType) }}</el-tag>
            </div>
            
            <div class="card-body">
              <div class="card-info">
                <span class="mono">{{ ds.dbType === 'sqlite' ? (ds.filePath || '(内存库)') : (ds.host + ':' + ds.port) }}</span>
              </div>
              <div class="card-info">
                <span v-if="ds.connStatus === 'ok' && ds.connLatencyMs" :style="{ color: ds.connLatencyMs > 200 ? '#e6a23c' : '#67c23a' }">
                  {{ ds.connLatencyMs }} ms
                </span>
                <span v-else-if="ds.connStatus === 'fail'" class="fail-text">连接失败</span>
                <span v-else class="text-muted">未测试</span>
              </div>
            </div>
            
            <div class="card-actions" :class="{ visible: activeActions === ds.datasourceId }">
              <el-button 
                size="small" 
                :type="ds.connStatus === 'ok' ? 'success' : ''" 
                @click.stop="refreshConn(ds)" 
                :loading="refreshingConnId === ds.datasourceId"
                :icon="CheckCircle"
              >测试</el-button>
              <el-button size="small" @click.stop="goWorkbench(ds)" :icon="Terminal">SQL</el-button>
              <el-button size="small" type="primary" @click.stop="$emit('edit', ds)" :icon="EditIcon">编辑</el-button>
              <el-button size="small" type="danger" @click.stop="handleDelete(ds)" :icon="DeleteIcon">删除</el-button>
            </div>
          </div>

          <div v-if="getEnvDatasources(env.key).length === 0" class="empty-column">
            <div class="empty-icon">📭</div>
            <div class="empty-text">暂无数据源</div>
          </div>
        </div>
      </div>
    </div>

    <div class="legend-bar">
      <div class="legend-item">
        <span class="status-dot status-ok"></span>
        <span>连接成功</span>
      </div>
      <div class="legend-item">
        <span class="status-dot status-fail"></span>
        <span>连接失败</span>
      </div>
      <div class="legend-item">
        <span class="status-dot status-none"></span>
        <span>未测试</span>
      </div>
      <div class="legend-divider"></div>
      <div class="legend-item">
        <el-tag type="primary" size="small">MySQL</el-tag>
        <span>MySQL</span>
      </div>
      <div class="legend-item">
        <el-tag type="success" size="small">TiDB</el-tag>
        <span>TiDB</span>
      </div>
      <div class="legend-item">
        <el-tag type="info" size="small">SQLite</el-tag>
        <span>SQLite</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Refresh as RefreshIcon, Plus as PlusIcon, Edit as EditIcon,
  Delete as DeleteIcon, Terminal, CheckCircle
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { useDatasourceStore } from '@/stores/datasource'
import { listDatasources, deleteDatasource, testConnectionById } from '@/api/datasource'

const emit = defineEmits<{
  (e: 'create'): void
  (e: 'edit', datasource: any): void
}>()

const router = useRouter()
const userStore = useUserStore()
const dsStore = useDatasourceStore()

const loading = ref(false)
const list = ref<any[]>([])
const activeActions = ref<string>('')
const refreshingConnId = ref<string>('')

const environments = [
  { key: 'dev', label: '开发', icon: '🌱' },
  { key: 'test', label: '测试', icon: '🧪' },
  { key: 'stage', label: '预发', icon: '🚀' },
  { key: 'prod', label: '生产', icon: '🏭' }
]

function getEnvDatasources(env: string) {
  return list.value.filter(ds => ds.env === env || (!ds.env && env === 'dev'))
}

function colorOf(label: string) {
  const map: Record<string, string> = {
    blue: '#409eff', green: '#67c23a', red: '#f56c6c',
    yellow: '#e6a23c', purple: '#8e44ad', orange: '#e67e22', gray: '#909399'
  }
  return map[label] || '#409eff'
}

function statusClass(row: any) {
  if (row.connStatus === 'ok') return 'status-ok'
  if (row.connStatus === 'fail') return 'status-fail'
  return 'status-none'
}

function connStatusTooltip(row: any) {
  if (row.connStatus === 'ok') {
    return `连接成功${row.connLatencyMs ? '（' + row.connLatencyMs + ' ms）' : ''}\n${row.lastConnTestAt || ''}`
  }
  if (row.connStatus === 'fail') return `连接失败\n${row.lastConnTestAt || ''}`
  return '尚未测试'
}

function dbTypeLabel(type: string) {
  const m: Record<string, string> = { mysql: 'MySQL', tidb: 'TiDB', sqlite: 'SQLite' }
  return m[type] || type
}

function dbTypeTagType(type: string) {
  const m: Record<string, string> = { mysql: 'primary', tidb: 'success', sqlite: 'info' }
  return m[type] || ''
}

function showActions(id: string) {
  activeActions.value = id
}

function hideActions() {
  activeActions.value = ''
}

async function loadList() {
  loading.value = true
  try {
    const r: any = await listDatasources(1, 1000, '', '', '', '')
    list.value = r.list || []
  } finally {
    loading.value = false
  }
}

async function refreshConn(row: any) {
  refreshingConnId.value = row.datasourceId
  try {
    const res: any = await testConnectionById(row.datasourceId)
    const idx = list.value.findIndex((d: any) => d.datasourceId === row.datasourceId)
    if (idx >= 0) {
      list.value[idx].connStatus = res?.success ? 'ok' : 'fail'
      list.value[idx].lastConnTestAt = new Date().toLocaleString()
      if (res?.latencyMs != null) list.value[idx].connLatencyMs = res.latencyMs
    }
    ElMessage.success(res?.success ? '连接成功' : (res?.message || '连接失败'))
  } catch (e: any) {
    const idx = list.value.findIndex((d: any) => d.datasourceId === row.datasourceId)
    if (idx >= 0) {
      list.value[idx].connStatus = 'fail'
      list.value[idx].lastConnTestAt = new Date().toLocaleString()
    }
    ElMessage.error(e?.message || '连接失败')
  } finally {
    refreshingConnId.value = ''
  }
}

function goWorkbench(row: any) {
  dsStore.setDatasource(row.datasourceId, row.name)
  router.push('/sql/workbench')
}

function handleDelete(row: any) {
  ElMessageBox.confirm(
    `确认删除数据源「${row.name}」？删除后不可恢复。`,
    '⚠️ 删除确认',
    {
      type: 'warning',
      confirmButtonText: '确认删除',
      cancelButtonText: '取消',
      distinguishCancelAndClose: true
    }
  ).then(async () => {
    try {
      await ElMessageBox.prompt(
        `请输入数据源名称的前 4 个字符以二次确认（${(row.name || '').slice(0, 4)}）`,
        '🔒 二次确认',
        {
          inputPattern: new RegExp('^' + (row.name || '').slice(0, 4).replace(/[.*+?^${}()|[\]\\]/g, '\\$&')),
          inputErrorMessage: '输入不匹配，请重新输入'
        }
      )
    } catch { return }
    try {
      await deleteDatasource(row.datasourceId)
      ElMessage.success('删除成功')
      loadList()
    } catch (e: any) {
      ElMessage.error(e?.message || '删除失败')
    }
  }).catch(() => {})
}

onMounted(() => {
  loadList()
})
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 18px;
}

.title-row {
  display: flex;
  align-items: baseline;
  gap: 14px;
}

.title-text {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}

.subtitle-text {
  font-size: 13px;
  color: #909399;
}

.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.matrix-container {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.matrix-column {
  background: #fafbfc;
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid #ebeef5;
}

.column-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  background: #fff;
  border-bottom: 2px solid;
}

.column-header.dev {
  border-color: #67c23a;
}

.column-header.test {
  border-color: #409eff;
}

.column-header.stage {
  border-color: #e6a23c;
}

.column-header.prod {
  border-color: #f56c6c;
}

.column-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
  color: #303133;
}

.env-icon {
  font-size: 18px;
}

.column-count {
  background: #f5f7fa;
  color: #606266;
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 10px;
  font-weight: 500;
}

.column-content {
  padding: 12px;
  max-height: calc(100vh - 280px);
  overflow-y: auto;
}

.ds-card {
  background: #fff;
  border-radius: 8px;
  margin-bottom: 12px;
  overflow: hidden;
  transition: all 0.2s ease;
  border: 1px solid #ebeef5;
  position: relative;
}

.ds-card:hover {
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  transform: translateY(-2px);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  border-left: 4px solid #409eff;
  background: #fafbfc;
}

.card-title-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.ds-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.card-body {
  padding: 10px 12px;
}

.card-info {
  font-size: 12px;
  margin-bottom: 4px;
}

.card-info:last-child {
  margin-bottom: 0;
}

.mono {
  font-family: Menlo, Monaco, Consolas, monospace;
  color: #606266;
}

.fail-text {
  color: #f56c6c;
}

.text-muted {
  color: #c0c4cc;
}

.status-dot {
  display: inline-block;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-ok {
  background: #67c23a;
  box-shadow: 0 0 0 3px rgba(103, 194, 58, 0.18);
}

.status-fail {
  background: #f56c6c;
  box-shadow: 0 0 0 3px rgba(245, 108, 108, 0.18);
  animation: blink 1.5s infinite;
}

.status-none {
  background: #c0c4cc;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.card-actions {
  display: flex;
  gap: 4px;
  padding: 8px 12px;
  background: #f8f9fa;
  border-top: 1px solid #ebeef5;
  opacity: 0;
  transform: translateY(10px);
  transition: all 0.2s ease;
}

.card-actions.visible {
  opacity: 1;
  transform: translateY(0);
}

.empty-column {
  padding: 40px 20px;
  text-align: center;
  color: #c0c4cc;
}

.empty-icon {
  font-size: 40px;
  margin-bottom: 8px;
}

.empty-text {
  font-size: 13px;
}

.legend-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 16px;
  background: #fff;
  border-radius: 8px;
  border: 1px solid #ebeef5;
  margin-top: 16px;
  flex-wrap: wrap;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #606266;
}

.legend-divider {
  width: 1px;
  height: 20px;
  background: #ebeef5;
}

@media (max-width: 1200px) {
  .matrix-container {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .matrix-container {
    grid-template-columns: 1fr;
  }
}
</style>