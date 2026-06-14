<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DB老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <div class="page-container">
    <div class="page-header">服务器管理</div>
    <div class="card">
      <div class="toolbar">
        <el-input v-model="keyword" placeholder="搜索名称/主机" clearable style="width:240px;" @change="loadList" />
        <el-select v-model="filterEnv" placeholder="全部环境" clearable style="width:140px;" @change="loadList">
          <el-option label="开发" value="dev" />
          <el-option label="测试" value="test" />
          <el-option label="生产" value="prod" />
        </el-select>
        <el-button type="primary" :icon="SearchIcon" @click="loadList">搜索</el-button>
        <el-button :icon="RefreshIcon" @click="loadList">刷新</el-button>
        <el-button type="primary" :icon="PlusIcon" @click="openDialog()">新建服务器</el-button>
        <ColumnToggle v-model="colVisible" :columns="columns" />
      </div>

      <el-table :data="list" style="width:100%;" v-loading="loading" stripe :default-sort="{ prop: 'createdAt', order: 'descending' }">
        <el-table-column prop="name" label="名称" sortable :resizable="true" v-if="colVisible.name" min-width="140" />
        <el-table-column label="所属业务" sortable :resizable="true" v-if="colVisible.businessId" min-width="160" prop="businessId">
          <template #default="{ row }">{{ businessMap[row.businessId]?.name || row.projectId || '-' }}</template>
        </el-table-column>
        <el-table-column prop="env" label="环境" sortable :resizable="true" v-if="colVisible.env" min-width="100">
          <template #default="{ row }">
            <el-tag v-if="row.env === 'prod'" type="danger">生产</el-tag>
            <el-tag v-else-if="row.env === 'dev'" type="success">开发</el-tag>
            <el-tag v-else>{{ row.env }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="连接" sortable :resizable="true" v-if="colVisible.host" min-width="180" prop="host">
          <template #default="{ row }">{{ row.host }}:{{ row.port }}</template>
        </el-table-column>
        <el-table-column prop="username" label="用户名" sortable :resizable="true" v-if="colVisible.username" min-width="120" />
        <el-table-column prop="authType" label="认证方式" sortable :resizable="true" v-if="colVisible.authType" min-width="120">
          <template #default="{ row }">
            <el-tag v-if="row.authType === 'key'">密钥</el-tag>
            <el-tag v-else type="success">密码</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" sortable :resizable="true" v-if="colVisible.status" min-width="100" prop="status">
          <template #default="{ row }">
            <el-tag v-if="row.status === 'active'" type="success">启用</el-tag>
            <el-tag v-else type="info">禁用</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" sortable :resizable="true" v-if="colVisible.createdAt" min-width="180" />
        <el-table-column label="操作" min-width="260" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="testConnection(row)">测试连接</el-button>
            <el-button size="small" @click="openDialog(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div style="margin-top:16px;display:flex;justify-content:flex-end;">
        <el-pagination
          v-model:current-page="current"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="loadList"
          @size-change="loadList"
        />
      </div>
    </div>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑服务器' : '新建服务器'" width="600px">
      <el-form :model="form" label-width="100px" ref="formRef" :rules="rules">
        <el-form-item label="所属业务" prop="businessId">
          <el-select v-model="form.businessId" style="width:100%;" @change="onBizChange" placeholder="请选择业务（如无请先创建）">
            <el-option v-for="b in businesses" :key="b.businessId" :label="b.projectName + ' / ' + (b.code ? b.code + ' - ' : '') + b.name" :value="b.businessId" />
          </el-select>
        </el-form-item>
        <el-form-item label="名称" prop="name"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="环境" prop="env">
          <el-select v-model="form.env" style="width:100%;">
            <el-option label="开发" value="dev" />
            <el-option label="测试" value="test" />
            <el-option label="生产" value="prod" />
          </el-select>
        </el-form-item>
        <el-form-item label="主机" prop="host"><el-input v-model="form.host" placeholder="如 192.168.1.100" /></el-form-item>
        <el-form-item label="端口" prop="port"><el-input-number v-model="form.port" :min="1" :max="65535" style="width:100%;" /></el-form-item>
        <el-form-item label="用户名" prop="username"><el-input v-model="form.username" /></el-form-item>
        <el-form-item label="认证方式" prop="authType">
          <el-select v-model="form.authType" style="width:100%;">
            <el-option label="密码" value="password" />
            <el-option label="密钥" value="key" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="form.authType === 'password'" label="密码">
          <el-input v-model="form.password" type="password" show-password :placeholder="isEdit ? '不修改请留空' : '请输入密码'" />
        </el-form-item>
        <el-form-item v-if="form.authType === 'key'" label="私钥">
          <el-input v-model="form.privateKey" type="textarea" :rows="4" placeholder="-----BEGIN RSA PRIVATE KEY-----" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="warning" @click="testCurrent">测试连接</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Search as SearchIcon, Refresh as RefreshIcon, Plus as PlusIcon } from '@element-plus/icons-vue'
import {
  createServer as createServerApi, updateServer as updateServerApi,
  deleteServer as deleteServerApi, testServer, testServerById, listServers
} from '@/api/server'
import { listAllBusinesses } from '@/api/business'
import ColumnToggle from '@/components/ColumnToggle.vue'

const columns = [
  { key: 'name', label: '名称' },
  { key: 'businessId', label: '所属业务' },
  { key: 'env', label: '环境' },
  { key: 'host', label: '连接' },
  { key: 'username', label: '用户名' },
  { key: 'authType', label: '认证方式' },
  { key: 'status', label: '状态' },
  { key: 'createdAt', label: '创建时间' }
]
const colVisible = reactive<Record<string, boolean>>({
  name: true, businessId: true, env: true, host: true,
  username: true, authType: true, status: true, createdAt: true
})

const loading = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const current = ref(1)
const pageSize = ref(10)
const keyword = ref('')
const filterEnv = ref('')
const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref('')
const formRef = ref<FormInstance>()
const businesses = ref<any[]>([])
const businessMap = ref<Record<string, any>>({})

const form = reactive({
  projectId: '', businessId: '', name: '', env: 'dev', host: '', port: 22, username: '', authType: 'password', password: '', privateKey: ''
})

const rules: FormRules = {
  businessId: [{ required: true, message: '请选择所属业务' }],
  name: [{ required: true, message: '请输入名称' }],
  host: [{ required: true, message: '请输入主机' }],
  port: [{ required: true, message: '请输入端口' }],
  username: [{ required: true, message: '请输入用户名' }],
  authType: [{ required: true, message: '请选择认证方式' }]
}

async function loadBusinesses() {
  try {
    const r: any = await listAllBusinesses()
    businesses.value = (r.list || r || []).map((b: any) => ({ ...b }))
    const m: Record<string, any> = {}
    for (const b of businesses.value) {
      m[b.businessId] = b
    }
    businessMap.value = m
  } catch (e) {
    businesses.value = []
  }
}

async function loadList() {
  loading.value = true
  try {
    const r: any = await listServers(current.value, pageSize.value, keyword.value, filterEnv.value)
    const all: any[] = (r?.list || r || []).map((s: any) => ({ ...s }))
    list.value = all
    total.value = r?.total || all.length
  } finally {
    loading.value = false
  }
}

function onBizChange(bizId: string) {
  const b = businesses.value.find((x: any) => x.businessId === bizId)
  if (b) {
    form.projectId = b.projectId
  }
}

function openDialog(row?: any) {
  if (row) {
    isEdit.value = true
    editId.value = row.serverId
    Object.assign(form, {
      projectId: row.projectId || '',
      businessId: row.businessId || '',
      name: row.name, env: row.env || 'dev', host: row.host, port: row.port,
      username: row.username, authType: row.authType || 'password', password: '', privateKey: ''
    })
  } else {
    isEdit.value = false
    editId.value = ''
    Object.assign(form, {
      projectId: '', businessId: '', name: '', env: 'dev', host: '', port: 22, username: '', authType: 'password', password: '', privateKey: ''
    })
  }
  dialogVisible.value = true
}

async function save() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      const payload = { ...form }
      delete (payload as any).privateKey
      if (isEdit.value) {
        await updateServerApi(editId.value, payload)
        ElMessage.success('更新成功')
      } else {
        await createServerApi(payload)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      loadList()
    } catch (e: any) {
      ElMessage.error(e?.message || '保存失败')
    }
  })
}

async function testCurrent() {
  try {
    const r: any = await testServer({ host: form.host, port: form.port, username: form.username, authType: form.authType, password: form.password, privateKey: form.privateKey })
    ElMessage.success('连接测试成功')
  } catch (e: any) {
    ElMessage.error(e?.message || '连接测试失败')
  }
}

async function testConnection(row: any) {
  try {
    await testServerById(row.serverId)
    ElMessage.success('连接测试成功')
  } catch (e: any) {
    ElMessage.error(e?.message || '连接测试失败')
  }
}

function handleDelete(row: any) {
  ElMessageBox.confirm('确认删除该服务器？', '确认', { type: 'warning' }).then(async () => {
    try {
      await deleteServerApi(row.serverId)
      ElMessage.success('已删除')
      loadList()
    } catch (e: any) {
      ElMessage.error(e?.message || '删除失败')
    }
  }).catch(() => {})
}

onMounted(async () => {
  await loadBusinesses()
  loadList()
})
</script>

