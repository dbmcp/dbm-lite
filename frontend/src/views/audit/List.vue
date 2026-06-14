<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DBA老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <div class="page-container">
    <div class="page-header">操作审计</div>
    <div class="card">
      <el-tabs v-model="activeCategory" @tab-change="onCategoryChange">
        <el-tab-pane label="平台操作日志" name="platform">
          <div class="toolbar" style="margin-top: 12px;">
            <el-input v-model="platformKw" placeholder="搜索操作/对象" clearable style="width:220px;" @change="loadPlatformList" />
            <el-input v-model="platformUserKw" placeholder="操作人" clearable style="width:160px;" @change="loadPlatformList" />
            <el-select v-model="platformAction" placeholder="操作类型" clearable style="width:200px;" @change="loadPlatformList">
              <el-option label="登录" value="auth.login" />
              <el-option label="创建账号" value="account.create" />
              <el-option label="更新账号" value="account.update" />
              <el-option label="删除账号" value="account.delete" />
              <el-option label="创建项目" value="project.create" />
              <el-option label="更新项目" value="project.update" />
              <el-option label="删除项目" value="project.delete" />
              <el-option label="创建业务" value="business.create" />
              <el-option label="更新业务" value="business.update" />
              <el-option label="删除业务" value="business.delete" />
              <el-option label="创建服务器" value="server.create" />
              <el-option label="更新服务器" value="server.update" />
              <el-option label="删除服务器" value="server.delete" />
              <el-option label="测试服务器连接" value="server.testConnection" />
              <el-option label="执行插件" value="plugin.execute" />
            </el-select>
            <el-button type="primary" @click="loadPlatformList" :icon="SearchIcon">搜索</el-button>
            <el-button @click="loadPlatformList" :icon="RefreshIcon">刷新</el-button>
            <ColumnToggle v-model="platColVisible" :columns="platColumns" />
          </div>

          <el-table :data="platformList" style="width:100%;margin-top:12px;" v-loading="platformLoading" stripe max-height="60vh" :default-sort="{ prop: 'createdAt', order: 'descending' }">
            <el-table-column prop="username" label="操作人" sortable :resizable="true" v-if="platColVisible.username" min-width="120" />
            <el-table-column prop="action" label="操作类型" sortable :resizable="true" v-if="platColVisible.action" min-width="180">
              <template #default="{ row }">
                <el-tag size="small">{{ row.action }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="target" label="目标对象" sortable :resizable="true" v-if="platColVisible.target" min-width="130" />
            <el-table-column prop="targetId" label="对象ID" sortable :resizable="true" v-if="platColVisible.targetId" min-width="160" show-overflow-tooltip />
            <el-table-column prop="detail" label="详情" :resizable="true" v-if="platColVisible.detail" min-width="260" show-overflow-tooltip />
            <el-table-column prop="ipAddress" label="IP" sortable :resizable="true" v-if="platColVisible.ipAddress" min-width="140" />
            <el-table-column label="状态" sortable :resizable="true" v-if="platColVisible.status" min-width="100" prop="status">
              <template #default="{ row }">
                <el-tag size="small" v-if="row.status === 'success'" type="success">成功</el-tag>
                <el-tag size="small" v-else type="danger">失败</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="createdAt" label="时间" sortable :resizable="true" v-if="platColVisible.createdAt" min-width="180" />
          </el-table>

          <div style="margin-top:16px;display:flex;justify-content:flex-end;">
            <el-pagination
              v-model:current-page="platformCurrent"
              v-model:page-size="pageSize"
              :total="platformTotal"
              :page-sizes="[10, 20, 50, 100]"
              layout="total, sizes, prev, pager, next, jumper"
              @current-change="loadPlatformList"
              @size-change="loadPlatformList"
            />
          </div>
        </el-tab-pane>

        <el-tab-pane label="数据库操作日志" name="database">
          <div class="toolbar" style="margin-top: 12px;">
            <el-input v-model="dbKw" placeholder="搜索操作/对象" clearable style="width:220px;" @change="loadDbList" />
            <el-input v-model="dbUserKw" placeholder="操作人" clearable style="width:160px;" @change="loadDbList" />
            <el-select v-model="dbAction" placeholder="操作类型" clearable style="width:200px;" @change="loadDbList">
              <el-option label="SQL 执行" value="sql.execute" />
              <el-option label="创建数据源" value="datasource.create" />
              <el-option label="更新数据源" value="datasource.update" />
              <el-option label="删除数据源" value="datasource.delete" />
              <el-option label="测试数据源连接" value="datasource.testConnection" />
              <el-option label="创建备份" value="backup.create" />
              <el-option label="删除备份" value="backup.delete" />
            </el-select>
            <el-button type="primary" @click="loadDbList" :icon="SearchIcon">搜索</el-button>
            <el-button @click="loadDbList" :icon="RefreshIcon">刷新</el-button>
            <ColumnToggle v-model="dbColVisible" :columns="dbColumns" />
          </div>

          <el-table :data="dbList" style="width:100%;margin-top:12px;" v-loading="dbLoading" stripe max-height="60vh" :default-sort="{ prop: 'createdAt', order: 'descending' }">
            <el-table-column prop="username" label="操作人" sortable :resizable="true" v-if="dbColVisible.username" min-width="120" />
            <el-table-column prop="action" label="操作类型" sortable :resizable="true" v-if="dbColVisible.action" min-width="180">
              <template #default="{ row }">
                <el-tag size="small" :type="row.action === 'sql.execute' ? 'warning' : ''">{{ row.action }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="target" label="目标对象" sortable :resizable="true" v-if="dbColVisible.target" min-width="130" />
            <el-table-column prop="targetId" label="对象ID" sortable :resizable="true" v-if="dbColVisible.targetId" min-width="160" show-overflow-tooltip />
            <el-table-column prop="detail" label="详情" :resizable="true" v-if="dbColVisible.detail" min-width="260" show-overflow-tooltip />
            <el-table-column prop="ipAddress" label="IP" sortable :resizable="true" v-if="dbColVisible.ipAddress" min-width="140" />
            <el-table-column label="状态" sortable :resizable="true" v-if="dbColVisible.status" min-width="100" prop="status">
              <template #default="{ row }">
                <el-tag size="small" v-if="row.status === 'success'" type="success">成功</el-tag>
                <el-tag size="small" v-else type="danger">失败</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="createdAt" label="时间" sortable :resizable="true" v-if="dbColVisible.createdAt" min-width="180" />
          </el-table>

          <div style="margin-top:16px;display:flex;justify-content:flex-end;">
            <el-pagination
              v-model:current-page="dbCurrent"
              v-model:page-size="pageSize"
              :total="dbTotal"
              :page-sizes="[10, 20, 50, 100]"
              layout="total, sizes, prev, pager, next, jumper"
              @current-change="loadDbList"
              @size-change="loadDbList"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Search as SearchIcon, Refresh as RefreshIcon } from '@element-plus/icons-vue'
import { listAuditLogs } from '@/api/sql'
import ColumnToggle from '@/components/ColumnToggle.vue'

const platColumns = [
  { key: 'username', label: '操作人' },
  { key: 'action', label: '操作类型' },
  { key: 'target', label: '目标对象' },
  { key: 'targetId', label: '对象ID' },
  { key: 'detail', label: '详情' },
  { key: 'ipAddress', label: 'IP' },
  { key: 'status', label: '状态' },
  { key: 'createdAt', label: '时间' }
]
const platColVisible = reactive<Record<string, boolean>>({
  username: true, action: true, target: true, targetId: true,
  detail: true, ipAddress: true, status: true, createdAt: true
})

const dbColumns = [
  { key: 'username', label: '操作人' },
  { key: 'action', label: '操作类型' },
  { key: 'target', label: '目标对象' },
  { key: 'targetId', label: '对象ID' },
  { key: 'detail', label: '详情' },
  { key: 'ipAddress', label: 'IP' },
  { key: 'status', label: '状态' },
  { key: 'createdAt', label: '时间' }
]
const dbColVisible = reactive<Record<string, boolean>>({
  username: true, action: true, target: true, targetId: true,
  detail: true, ipAddress: true, status: true, createdAt: true
})

const activeCategory = ref('platform')
const pageSize = ref(20)

const platformList = ref<any[]>([])
const platformTotal = ref(0)
const platformCurrent = ref(1)
const platformLoading = ref(false)
const platformKw = ref('')
const platformUserKw = ref('')
const platformAction = ref('')

const dbList = ref<any[]>([])
const dbTotal = ref(0)
const dbCurrent = ref(1)
const dbLoading = ref(false)
const dbKw = ref('')
const dbUserKw = ref('')
const dbAction = ref('')

async function loadPlatformList() {
  platformLoading.value = true
  try {
    const r: any = await listAuditLogs(platformCurrent.value, pageSize.value, platformAction.value, platformUserKw.value, platformKw.value, 'platform')
    platformList.value = r.list || []
    platformTotal.value = r.total || 0
  } finally {
    platformLoading.value = false
  }
}

async function loadDbList() {
  dbLoading.value = true
  try {
    const r: any = await listAuditLogs(dbCurrent.value, pageSize.value, dbAction.value, dbUserKw.value, dbKw.value, 'database')
    dbList.value = r.list || []
    dbTotal.value = r.total || 0
  } finally {
    dbLoading.value = false
  }
}

function onCategoryChange() {
  if (activeCategory.value === 'platform') {
    loadPlatformList()
  } else {
    loadDbList()
  }
}

onMounted(() => {
  loadPlatformList()
})
</script>

