<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DBA老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <div class="page-container">
    <div class="page-header">业务管理</div>

    <div class="card">
      <el-tabs v-model="activeTab" @tab-change="onTabChange">
        <el-tab-pane label="项目管理" name="projects">
          <div class="toolbar">
            <el-input v-model="projectKeyword" placeholder="搜索项目名称" clearable style="width:260px;" @change="loadProjectsList" />
            <el-button type="primary" :icon="PlusIcon" @click="openProjectDialog()">新建项目</el-button>
            <el-button :icon="RefreshIcon" @click="loadProjectsList">刷新</el-button>
            <ColumnToggle v-model="projColVisible" :columns="projColumns" />
          </div>
          <el-table :data="projectsList" style="width:100%;" v-loading="projectsLoading" stripe :default-sort="{ prop: 'createdAt', order: 'descending' }">
            <el-table-column prop="projectId" label="项目ID" sortable :resizable="true" v-if="projColVisible.projectId" min-width="180" />
            <el-table-column prop="name" label="项目名称" sortable :resizable="true" v-if="projColVisible.name" min-width="160" />
            <el-table-column prop="description" label="描述" show-overflow-tooltip :resizable="true" v-if="projColVisible.description" min-width="200" />
            <el-table-column prop="createdBy" label="创建人" sortable :resizable="true" v-if="projColVisible.createdBy" min-width="140" />
            <el-table-column prop="createdAt" label="创建时间" sortable :resizable="true" v-if="projColVisible.createdAt" min-width="180" />
            <el-table-column label="操作" min-width="280" fixed="right">
              <template #default="{ row }">
                <el-button size="small" type="primary" @click="manageProjectBusiness(row)">业务管理</el-button>
                <el-button size="small" @click="openProjectDialog(row)">编辑</el-button>
                <el-button size="small" type="danger" @click="handleDeleteProject(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="业务管理" name="businesses">
          <div class="toolbar">
            <el-select v-model="selectedProject" placeholder="全部项目" clearable style="width:200px;" @change="onProjectChange">
              <el-option v-for="p in projectsList" :key="p.projectId" :label="p.name" :value="p.projectId" />
            </el-select>
            <el-input v-model="keyword" placeholder="搜索业务名称/编码" clearable style="width:220px;" @change="loadList" />
            <el-select v-model="envFilter" placeholder="全部环境" clearable style="width:140px;" @change="loadList">
              <el-option label="开发" value="dev" />
              <el-option label="测试" value="test" />
              <el-option label="预生产" value="preprod" />
              <el-option label="生产" value="prod" />
            </el-select>
            <el-button type="primary" :icon="SearchIcon" @click="loadList">搜索</el-button>
            <el-button :icon="RefreshIcon" @click="loadList">刷新</el-button>
            <el-button type="primary" :icon="PlusIcon" @click="openDialog()">新建业务</el-button>
            <ColumnToggle v-model="colVisible" :columns="columns" />
          </div>

          <el-table :data="list" style="width:100%;" v-loading="loading" stripe :default-sort="{ prop: 'createdAt', order: 'descending' }">
            <el-table-column prop="projectName" label="所属项目" sortable :resizable="true" v-if="colVisible.projectName" min-width="140" />
            <el-table-column prop="code" label="业务编码" sortable :resizable="true" v-if="colVisible.code" min-width="160" />
            <el-table-column prop="name" label="业务名称" sortable :resizable="true" v-if="colVisible.name" min-width="160" />
            <el-table-column prop="description" label="描述" show-overflow-tooltip :resizable="true" v-if="colVisible.description" min-width="200" />
            <el-table-column prop="env" label="环境" sortable :resizable="true" v-if="colVisible.env" min-width="110">
              <template #default="{ row }">
                <el-tag v-if="row.env === 'prod'" type="danger">生产</el-tag>
                <el-tag v-else-if="row.env === 'preprod'" type="warning">预生产</el-tag>
                <el-tag v-else-if="row.env === 'dev'" type="success">开发</el-tag>
                <el-tag v-else-if="row.env === 'test'" type="info">测试</el-tag>
                <el-tag v-else>{{ row.env }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="createdAt" label="创建时间" sortable :resizable="true" v-if="colVisible.createdAt" min-width="180" />
            <el-table-column label="操作" min-width="220" fixed="right">
              <template #default="{ row }">
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
        </el-tab-pane>
      </el-tabs>
    </div>

    <el-dialog v-model="projectDialogVisible" :title="isEditProject ? '编辑项目' : '新建项目'" width="480px">
      <el-form :model="projectForm" label-width="100px" ref="projectFormRef" :rules="projectRules">
        <el-form-item label="项目名称" prop="name"><el-input v-model="projectForm.name" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="projectForm.description" type="textarea" :rows="4" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="projectDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveProject">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑业务' : '新建业务'" width="520px">
      <el-form :model="form" label-width="100px" ref="formRef" :rules="rules">
        <el-form-item label="所属项目" prop="projectId">
          <el-select v-model="form.projectId" style="width:100%;" placeholder="请选择项目">
            <el-option v-for="p in projectsList" :key="p.projectId" :label="p.name" :value="p.projectId" />
          </el-select>
        </el-form-item>
        <el-form-item label="业务编码" prop="code"><el-input v-model="form.code" placeholder="如: order-service" /></el-form-item>
        <el-form-item label="业务名称" prop="name"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="环境" prop="env">
          <el-select v-model="form.env" style="width:100%;">
            <el-option label="开发" value="dev" />
            <el-option label="测试" value="test" />
            <el-option label="预生产" value="preprod" />
            <el-option label="生产" value="prod" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" :rows="3" /></el-form-item>
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
import { Search as SearchIcon, Refresh as RefreshIcon, Plus as PlusIcon } from '@element-plus/icons-vue'
import {
  listProjects, createProject as createProjectApi, updateProject as updateProjectApi, deleteProject as deleteProjectApi,
  createBusiness as createBizApi, updateBusiness as updateBizApi, deleteBusiness as deleteBizApi
} from '@/api/business'
import request from '@/api/request'
import ColumnToggle from '@/components/ColumnToggle.vue'

const projColumns = [
  { key: 'projectId', label: '项目ID' },
  { key: 'name', label: '项目名称' },
  { key: 'description', label: '描述' },
  { key: 'createdBy', label: '创建人' },
  { key: 'createdAt', label: '创建时间' }
]
const projColVisible = reactive<Record<string, boolean>>({
  projectId: true, name: true, description: true, createdBy: true, createdAt: true
})

const columns = [
  { key: 'projectName', label: '所属项目' },
  { key: 'code', label: '业务编码' },
  { key: 'name', label: '业务名称' },
  { key: 'description', label: '描述' },
  { key: 'env', label: '环境' },
  { key: 'createdAt', label: '创建时间' }
]
const colVisible = reactive<Record<string, boolean>>({
  projectName: true, code: true, name: true, description: true, env: true, createdAt: true
})

const activeTab = ref('projects')

// 项目管理
const projectsList = ref<any[]>([])
const projectsLoading = ref(false)
const projectKeyword = ref('')
const projectDialogVisible = ref(false)
const isEditProject = ref(false)
const editingProjectId = ref('')
const projectFormRef = ref<FormInstance>()
const projectForm = reactive({ name: '', description: '' })
const projectRules: FormRules = {
  name: [{ required: true, message: '请输入项目名称' }]
}

async function loadProjectsList() {
  projectsLoading.value = true
  try {
    const r: any = await listProjects(projectKeyword.value)
    projectsList.value = r.list || r || []
  } finally {
    projectsLoading.value = false
  }
}

function openProjectDialog(row?: any) {
  if (row) {
    isEditProject.value = true
    editingProjectId.value = row.projectId
    projectForm.name = row.name
    projectForm.description = row.description || ''
  } else {
    isEditProject.value = false
    editingProjectId.value = ''
    projectForm.name = ''
    projectForm.description = ''
  }
  projectDialogVisible.value = true
}

async function saveProject() {
  if (!projectFormRef.value) return
  await projectFormRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      if (isEditProject.value) {
        await updateProjectApi(editingProjectId.value, projectForm)
        ElMessage.success('更新成功')
      } else {
        await createProjectApi(projectForm)
        ElMessage.success('创建成功')
      }
      projectDialogVisible.value = false
      loadProjectsList()
    } catch (e: any) {
      ElMessage.error(e?.message || '保存失败')
    }
  })
}

function handleDeleteProject(row: any) {
  ElMessageBox.confirm(`确认删除项目「${row.name}」？（若该项目下存在业务请先删除业务）`, '提示', { type: 'warning' })
    .then(async () => {
      try {
        await deleteProjectApi(row.projectId)
        ElMessage.success('删除成功')
        loadProjectsList()
      } catch (e: any) {
        ElMessage.error(e?.message || '删除失败')
      }
    }).catch(() => {})
}

function manageProjectBusiness(row: any) {
  selectedProject.value = row.projectId
  activeTab.value = 'businesses'
  current.value = 1
  loadList()
}

// 业务管理
const loading = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const current = ref(1)
const pageSize = ref(10)
const keyword = ref('')
const envFilter = ref('')
const selectedProject = ref('')
const dialogVisible = ref(false)
const isEdit = ref(false)
const editingBizId = ref('')
const formRef = ref<FormInstance>()

const form = reactive({
  projectId: '',
  code: '',
  name: '',
  env: 'dev',
  description: ''
})

const rules: FormRules = {
  projectId: [{ required: true, message: '请选择所属项目' }],
  name: [{ required: true, message: '请输入业务名称' }],
  env: [{ required: true, message: '请选择环境' }]
}

function onTabChange() {
  if (activeTab.value === 'projects') {
    loadProjectsList()
  } else if (activeTab.value === 'businesses') {
    if (projectsList.value.length === 0) {
      loadProjectsList().then(() => loadList())
    } else {
      loadList()
    }
  }
}

async function loadList() {
  loading.value = true
  try {
    const params: any = { keyword: keyword.value, page: current.value, pageSize: pageSize.value }
    if (selectedProject.value) params.projectId = selectedProject.value
    const r: any = await request({ url: '/businesses', method: 'GET', params })
    const items = (r?.list || r || []).map((b: any) => ({
      ...b,
      projectName: projectsList.value.find((p: any) => p.projectId === b.projectId)?.name || b.projectId || ''
    }))
    const filtered = envFilter.value ? items.filter((b: any) => (b.env || b.environment) === envFilter.value) : items
    list.value = filtered
    total.value = r?.total || filtered.length
  } finally {
    loading.value = false
  }
}

function onProjectChange() {
  current.value = 1
  loadList()
}

function openDialog(row?: any) {
  if (row) {
    isEdit.value = true
    editingBizId.value = row.businessId
    Object.assign(form, {
      projectId: row.projectId,
      code: row.code || '',
      name: row.name,
      env: row.env || row.environment || 'dev',
      description: row.description || ''
    })
  } else {
    isEdit.value = false
    editingBizId.value = ''
    Object.assign(form, {
      projectId: selectedProject.value || (projectsList.value[0]?.projectId || ''),
      code: '',
      name: '',
      env: 'dev',
      description: ''
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
        await updateBizApi(editingBizId.value, form)
        ElMessage.success('更新成功')
      } else {
        await createBizApi(form)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      loadList()
    } catch (e: any) {
      ElMessage.error(e?.message || '保存失败')
    }
  })
}

function handleDelete(row: any) {
  ElMessageBox.confirm(`确认删除业务「${row.name}」？`, '提示', { type: 'warning' })
    .then(async () => {
      try {
        await deleteBizApi(row.businessId)
        ElMessage.success('删除成功')
        loadList()
      } catch (e: any) {
        ElMessage.error(e?.message || '删除失败')
      }
    }).catch(() => {})
}

onMounted(async () => {
  await loadProjectsList()
  if (activeTab.value === 'businesses') {
    loadList()
  }
})
</script>

