<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DB老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <div class="page-container">
    <div class="page-header">插件管理</div>
    <div class="card">
      <div class="toolbar">
        <el-input v-model="keyword" placeholder="搜索插件名称" clearable style="width:260px;" @change="loadList" />
        <el-button type="primary" :icon="PlusIcon" @click="openDialog()">新建插件</el-button>
        <el-button :icon="RefreshIcon" @click="loadList">刷新</el-button>
        <ColumnToggle v-model="colVisible" :columns="columns" />
      </div>

      <el-table :data="list" style="width:100%;" v-loading="loading" stripe :default-sort="{ prop: 'createdAt', order: 'descending' }">
        <el-table-column prop="name" label="插件名称" sortable :resizable="true" v-if="colVisible.name" min-width="200" />
        <el-table-column prop="version" label="版本" sortable :resizable="true" v-if="colVisible.version" min-width="120" />
        <el-table-column prop="description" label="描述" :resizable="true" v-if="colVisible.description" min-width="220" show-overflow-tooltip />
        <el-table-column prop="params" label="参数" :resizable="true" v-if="colVisible.params" min-width="260" show-overflow-tooltip />
        <el-table-column prop="downloadUrl" label="下载地址" :resizable="true" v-if="colVisible.downloadUrl" min-width="240" show-overflow-tooltip />
        <el-table-column label="状态" sortable :resizable="true" v-if="colVisible.status" min-width="100" prop="status">
          <template #default="{ row }">
            <el-tag v-if="row.status === 'active'" type="success" size="small">启用</el-tag>
            <el-tag v-else type="info" size="small">禁用</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" sortable :resizable="true" v-if="colVisible.createdAt" min-width="180" />
        <el-table-column label="操作" min-width="220" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="downloadPlugin(row)" :loading="downloading === row.pluginId">下载</el-button>
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

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑插件' : '新建插件'" width="560px">
      <el-form :model="form" label-width="100px" ref="formRef" :rules="rules">
        <el-form-item label="插件名称" prop="name"><el-input v-model="form.name" placeholder="如: mysqldump-plugin" /></el-form-item>
        <el-form-item label="版本" prop="version"><el-input v-model="form.version" placeholder="如: 1.0.0" /></el-form-item>
        <el-form-item label="描述" prop="description"><el-input v-model="form.description" type="textarea" :rows="2" /></el-form-item>
        <el-form-item label="参数" prop="params"><el-input v-model="form.params" type="textarea" :rows="3" placeholder='JSON 对象，如: {"--database": "test"}' /></el-form-item>
        <el-form-item label="下载地址" prop="downloadUrl"><el-input v-model="form.downloadUrl" placeholder="http 或 文件路径" /></el-form-item>
        <el-form-item label="配置" prop="config"><el-input v-model="form.config" type="textarea" :rows="3" placeholder="可选，JSON 或文本配置" /></el-form-item>
        <el-form-item label="状态" prop="status">
          <el-select v-model="form.status" style="width:100%;">
            <el-option label="启用" value="active" />
            <el-option label="禁用" value="inactive" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Refresh as RefreshIcon, Plus as PlusIcon } from '@element-plus/icons-vue'
import request from '@/api/request'
import ColumnToggle from '@/components/ColumnToggle.vue'

const columns = [
  { key: 'name', label: '插件名称' },
  { key: 'version', label: '版本' },
  { key: 'description', label: '描述' },
  { key: 'params', label: '参数' },
  { key: 'downloadUrl', label: '下载地址' },
  { key: 'status', label: '状态' },
  { key: 'createdAt', label: '创建时间' }
]
const colVisible = reactive<Record<string, boolean>>({
  name: true, version: true, description: true, params: true,
  downloadUrl: true, status: true, createdAt: true
})

const loading = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const current = ref(1)
const pageSize = ref(10)
const keyword = ref('')
const downloading = ref('')

const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref('')
const formRef = ref<FormInstance>()

const form = reactive({
  name: '',
  version: '',
  description: '',
  params: '',
  downloadUrl: '',
  config: '',
  status: 'active'
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入插件名称' }]
}

async function loadList() {
  loading.value = true
  try {
    const r: any = await request({
      url: '/plugins',
      method: 'GET',
      params: { keyword: keyword.value, page: current.value, pageSize: pageSize.value }
    })
    list.value = r?.list || r || []
    total.value = r?.total || list.value.length
  } catch (e: any) {
    ElMessage.error(e?.message || '加载失败')
  } finally {
    loading.value = false
  }
}

function openDialog(row?: any) {
  if (row) {
    isEdit.value = true
    editId.value = row.pluginId
    Object.assign(form, {
      name: row.name || '',
      version: row.version || '',
      description: row.description || '',
      params: row.params || '',
      downloadUrl: row.downloadUrl || '',
      config: row.config || '',
      status: row.status || 'active'
    })
  } else {
    isEdit.value = false
    editId.value = ''
    Object.assign(form, {
      name: '',
      version: '',
      description: '',
      params: '',
      downloadUrl: '',
      config: '',
      status: 'active'
    })
  }
  dialogVisible.value = true
}

async function save() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      if (isEdit.value) {
        await request({ url: `/plugins/${editId.value}`, method: 'PUT', data: form })
        ElMessage.success('更新成功')
      } else {
        await request({ url: '/plugins', method: 'POST', data: form })
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      loadList()
    } catch (e: any) {
      ElMessage.error(e?.message || '保存失败')
    }
  })
}

async function downloadPlugin(row: any) {
  if (!row.downloadUrl) {
    ElMessage.warning('该插件未配置下载地址')
    return
  }
  downloading.value = row.pluginId
  try {
    // 触发浏览器下载：若为 URL 直接打开新标签
    if (row.downloadUrl.startsWith('http://') || row.downloadUrl.startsWith('https://')) {
      window.open(row.downloadUrl, '_blank')
    } else {
      await request({
        url: `/plugins/${row.pluginId}/download`,
        method: 'GET',
        responseType: 'blob'
      }).then((blob: Blob) => {
        const link = document.createElement('a')
        const url = URL.createObjectURL(blob)
        link.href = url
        link.download = row.name + '-' + (row.version || 'latest')
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        URL.revokeObjectURL(url)
      })
    }
    ElMessage.success('已触发下载')
  } catch (e: any) {
    ElMessage.error(e?.message || '下载失败，请检查下载地址或网络')
  } finally {
    downloading.value = ''
  }
}

function handleDelete(row: any) {
  ElMessageBox.confirm(`确认删除插件「${row.name}」？`, '确认', { type: 'warning' })
    .then(async () => {
      try {
        await request({ url: `/plugins/${row.pluginId}`, method: 'DELETE' })
        ElMessage.success('删除成功')
        loadList()
      } catch (e: any) {
        ElMessage.error(e?.message || '删除失败')
      }
    }).catch(() => {})
}

onMounted(loadList)
</script>

<style scoped>
</style>

