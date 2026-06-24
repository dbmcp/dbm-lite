<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DB老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <div class="page-container">
    <div class="page-header">业务管理</div>

    <div class="card">
      <el-tabs v-model="activeTab">
        <!-- ==================== 项目管理 ==================== -->
        <el-tab-pane label="项目管理" name="projects">
          <div class="toolbar">
            <el-input v-model="projectKeyword" placeholder="搜索项目名称" clearable style="width:260px;" @change="loadProjectsList" />
            <el-button type="primary" :icon="PlusIcon" @click="openProjectDialog()">新建项目</el-button>
            <el-button @click="loadProjectsList"><span class="refresh-icon">⟳</span>刷新</el-button>
            <ColumnToggle v-model="projColVisible" :columns="projColumns" />
          </div>
          <el-table :data="projectsList" style="width:100%;" v-loading="projectsLoading" stripe :default-sort="{ prop: 'createdAt', order: 'descending' }">
            <el-table-column prop="name" label="项目名称" sortable :resizable="true" v-if="projColVisible.name" min-width="180" show-overflow-tooltip />
            <el-table-column prop="description" label="描述" show-overflow-tooltip :resizable="true" v-if="projColVisible.description" min-width="200" />
            <el-table-column prop="createdBy" label="创建人" sortable :resizable="true" v-if="projColVisible.createdBy" min-width="140" show-overflow-tooltip />
            <el-table-column prop="createdAt" label="创建时间" sortable :resizable="true" v-if="projColVisible.createdAt" min-width="180" show-overflow-tooltip />
            <el-table-column label="成员数" min-width="100" align="center" v-if="projColVisible.members" show-overflow-tooltip>
              <template #default="{ row }">
                <el-tag size="small" type="primary" effect="plain">{{ getMemberCount(row.projectId, 'project') }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="关联业务数" min-width="110" align="center" v-if="projColVisible.bizCount" show-overflow-tooltip>
              <template #default="{ row }">
                <el-tag size="small" type="success" effect="plain">{{ getBizCount(row.projectId) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" min-width="340" fixed="right" align="center" show-overflow-tooltip>
              <template #default="{ row }">
                <el-button size="small" type="primary" @click="openProjectMemberDialog(row)">分配成员</el-button>
                <el-button size="small" @click="manageProjectBusiness(row)">业务管理</el-button>
                <el-button size="small" @click="openProjectDialog(row)">编辑</el-button>
                <el-button size="small" type="danger" @click="handleDeleteProject(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- ==================== 业务管理 ==================== -->
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
              <el-option label="灾备" value="disaster" />
            </el-select>
            <el-button type="primary" :icon="SearchIcon" @click="loadList">搜索</el-button>
            <el-button @click="loadList"><span class="refresh-icon">⟳</span>刷新</el-button>
            <el-button type="primary" :icon="PlusIcon" @click="openDialog()">新建业务</el-button>
            <ColumnToggle v-model="colVisible" :columns="columns" />
          </div>

          <el-table :data="filteredList" style="width:100%;" v-loading="loading" stripe :default-sort="{ prop: 'createdAt', order: 'descending' }">
            <el-table-column prop="projectName" label="所属项目" sortable :resizable="true" v-if="colVisible.projectName" min-width="140" show-overflow-tooltip />
            <el-table-column prop="code" label="业务编码" sortable :resizable="true" v-if="colVisible.code" min-width="160" show-overflow-tooltip />
            <el-table-column prop="name" label="业务名称" sortable :resizable="true" v-if="colVisible.name" min-width="160" show-overflow-tooltip />
            <el-table-column prop="description" label="描述" show-overflow-tooltip :resizable="true" v-if="colVisible.description" min-width="200" />
            <el-table-column label="环境" sortable :resizable="true" v-if="colVisible.env" min-width="110" prop="env" show-overflow-tooltip>
              <template #default="{ row }">
                <el-tag v-if="row.env === 'prod'" type="danger">生产</el-tag>
                <el-tag v-else-if="row.env === 'preprod'" type="warning">预生产</el-tag>
                <el-tag v-else-if="row.env === 'dev'" type="success">开发</el-tag>
                <el-tag v-else-if="row.env === 'test'" type="info">测试</el-tag>
                <el-tag v-else-if="row.env === 'disaster'" type="info" effect="plain">灾备</el-tag>
                <el-tag v-else>{{ row.env }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="成员数" min-width="100" align="center" v-if="colVisible.members" show-overflow-tooltip>
              <template #default="{ row }">
                <el-tag size="small" type="primary" effect="plain">{{ getMemberCount(row.businessId, 'business') }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="createdAt" label="创建时间" sortable :resizable="true" v-if="colVisible.createdAt" min-width="180" show-overflow-tooltip />
            <el-table-column label="操作" min-width="320" fixed="right" align="center" show-overflow-tooltip>
              <template #default="{ row }">
                <el-button size="small" type="primary" @click="openBizMemberDialog(row)">分配成员</el-button>
                <el-button size="small" @click="openDialog(row)">编辑</el-button>
                <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <div style="margin-top:16px;display:flex;justify-content:flex-end;">
            <el-pagination
              v-model:current-page="current"
              v-model:page-size="pageSize"
              :total="filteredTotal"
              :page-sizes="[100, 200, 500, 1000]"
              layout="total, sizes, prev, pager, next, jumper"
              @current-change="loadList"
              @size-change="loadList"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- 新建/编辑 项目 -->
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

    <!-- 项目成员分配 -->
    <el-dialog v-model="projectMemberDialogVisible" :title="'分配项目成员 - ' + (currentProject?.name || '')" width="520px">
      <div v-if="projectMemberLoading" style="text-align:center;padding:20px;">
        <el-icon class="is-loading" :size="20"><Loading /></el-icon> 加载中...
      </div>
      <div v-else>
        <el-checkbox v-model="projectMemberAll" @change="toggleAllProjectMembers">全选/取消全选</el-checkbox>
        <el-divider style="margin:10px 0;" />
        <el-scrollbar style="max-height:360px;">
          <div v-for="u in allUsers" :key="u.userId" style="padding:6px 0;">
            <el-checkbox v-model="projectMemberMap[u.userId]">
              <span>{{ u.displayName || u.username }}</span>
              <span style="color:#909399;margin-left:8px;font-size:12px;">@{{ u.username }}</span>
            </el-checkbox>
          </div>
          <div v-if="allUsers.length === 0" style="text-align:center;color:#909399;padding:20px;">
            暂无账号
          </div>
        </el-scrollbar>
      </div>
      <template #footer>
        <el-button @click="projectMemberDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveProjectMembers">保存</el-button>
      </template>
    </el-dialog>

    <!-- 新建/编辑 业务 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑业务' : '新建业务'" width="560px">
      <el-form :model="form" label-width="110px" ref="formRef" :rules="rules">
        <el-form-item label="所属项目" prop="projectId">
          <el-radio-group v-if="!isEdit" v-model="form.projectMode" style="margin-bottom:8px;" size="small">
            <el-radio-button label="existing">选择已有项目</el-radio-button>
            <el-radio-button label="new">新建项目并关联</el-radio-button>
          </el-radio-group>
          <el-select v-if="form.projectMode === 'existing'" v-model="form.projectId" style="width:100%;" placeholder="请选择项目" filterable>
            <el-option v-for="p in projectsList" :key="p.projectId" :label="p.name" :value="p.projectId" />
            <template #empty>
              <div style="text-align:center;padding:12px 0;">
                <span style="color:#909399;">暂无项目</span>
                <el-button size="small" type="primary" link style="margin-left:8px;" @click="form.projectMode = 'new'">
                  点击此处新建项目
                </el-button>
              </div>
            </template>
          </el-select>
          <div v-else-if="form.projectMode === 'new'" style="width:100%;">
            <el-input v-model="form.newProjectName" placeholder="输入新项目名称（必填）" style="margin-bottom:6px;" />
            <el-input v-model="form.newProjectDesc" placeholder="输入新项目描述（可选）" type="textarea" :rows="2" />
            <div style="color:#909399;font-size:12px;margin-top:4px;">
              * 业务保存时会自动创建该项目，并把此业务关联到新项目下
            </div>
          </div>
          <el-input v-else :value="projectNameById(form.projectId)" disabled style="width:100%;" />
        </el-form-item>
        <el-form-item label="业务编码" prop="code"><el-input v-model="form.code" placeholder="如: order-service" /></el-form-item>
        <el-form-item label="业务名称" prop="name"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="环境" prop="env">
          <el-select v-model="form.env" style="width:100%;">
            <el-option label="开发" value="dev" />
            <el-option label="测试" value="test" />
            <el-option label="预生产" value="preprod" />
            <el-option label="生产" value="prod" />
            <el-option label="灾备" value="disaster" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <!-- 业务成员分配 -->
    <el-dialog v-model="bizMemberDialogVisible" :title="'分配业务成员 - ' + (currentBiz?.name || '')" width="520px">
      <div v-if="bizMemberLoading" style="text-align:center;padding:20px;">
        <el-icon class="is-loading" :size="20"><Loading /></el-icon> 加载中...
      </div>
      <div v-else>
        <el-checkbox v-model="bizMemberAll" @change="toggleAllBizMembers">全选/取消全选</el-checkbox>
        <el-divider style="margin:10px 0;" />
        <el-scrollbar style="max-height:360px;">
          <div v-for="u in allUsers" :key="u.userId" style="padding:6px 0;">
            <el-checkbox v-model="bizMemberMap[u.userId]">
              <span>{{ u.displayName || u.username }}</span>
              <span style="color:#909399;margin-left:8px;font-size:12px;">@{{ u.username }}</span>
            </el-checkbox>
          </div>
          <div v-if="allUsers.length === 0" style="text-align:center;color:#909399;padding:20px;">
            暂无账号
          </div>
        </el-scrollbar>
      </div>
      <template #footer>
        <el-button @click="bizMemberDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveBizMembers">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Search as SearchIcon, Refresh as RefreshIcon, Plus as PlusIcon, Loading } from '@element-plus/icons-vue'
import request from '@/api/request'
import ColumnToggle from '@/components/ColumnToggle.vue'

// --------- 列定义 ---------
const projColumns = [
  { key: 'name', label: '项目名称' },
  { key: 'description', label: '描述' },
  { key: 'createdBy', label: '创建人' },
  { key: 'createdAt', label: '创建时间' },
  { key: 'members', label: '成员数' },
  { key: 'bizCount', label: '关联业务数' }
]
let projColVisible = reactive<Record<string, boolean>>({
  name: true, description: true, createdBy: true, createdAt: true, members: true, bizCount: true
})

const columns = [
  { key: 'projectName', label: '所属项目' },
  { key: 'code', label: '业务编码' },
  { key: 'name', label: '业务名称' },
  { key: 'description', label: '描述' },
  { key: 'env', label: '环境' },
  { key: 'members', label: '成员数' },
  { key: 'createdAt', label: '创建时间' }
]
let colVisible = reactive<Record<string, boolean>>({
  projectName: true, code: true, name: true, description: true, env: true, members: true, createdAt: true
})

// --------- Tabs ---------
const activeTab = ref('businesses')

// --------- 项目管理 ---------
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
    const r: any = await request({ url: '/projects', method: 'GET', params: { keyword: projectKeyword.value } })
    projectsList.value = r?.list || r || []
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
        await request({ url: '/projects/' + editingProjectId.value, method: 'PUT', data: projectForm })
        ElMessage.success('更新成功')
      } else {
        await request({ url: '/projects', method: 'POST', data: projectForm })
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
        await request({ url: '/projects/' + row.projectId, method: 'DELETE' })
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

// --------- 业务管理 ---------
const loading = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const current = ref(1)
const pageSize = ref(100)
const keyword = ref('')
const envFilter = ref('')
const selectedProject = ref('')
const dialogVisible = ref(false)
const isEdit = ref(false)
const editingBizId = ref('')
const formRef = ref<FormInstance>()

const form = reactive({
  projectId: '',
  projectMode: 'existing' as 'existing' | 'new',
  newProjectName: '',
  newProjectDesc: '',
  code: '',
  name: '',
  env: 'dev',
  description: ''
})

const rules: FormRules = {
  projectId: [{ validator: validateProject, trigger: 'change' }],
  name: [{ required: true, message: '请输入业务名称' }],
  env: [{ required: true, message: '请选择环境' }]
}

function validateProject(rule: any, value: any, callback: (err?: any) => void) {
  if (form.projectMode === 'new') {
    if (!form.newProjectName || !form.newProjectName.trim()) {
      return callback(new Error('请输入新项目名称'))
    }
    return callback()
  }
  if (!form.projectId) {
    return callback(new Error('请选择所属项目'))
  }
  return callback()
}

function projectNameById(id: string): string {
  const p = projectsList.value.find((x: any) => x.projectId === id)
  return p?.name || id || ''
}

async function loadList() {
  loading.value = true
  try {
    const params: any = { keyword: keyword.value, page: current.value, pageSize: pageSize.value }
    if (selectedProject.value) params.projectId = selectedProject.value
    const r: any = await request({ url: '/businesses', method: 'GET', params })
    const items = (r?.list || r || []).map((b: any) => ({
      ...b,
      projectName: projectNameById(b.projectId)
    }))
    list.value = items
    total.value = r?.total || items.length
  } finally {
    loading.value = false
  }
}

const filteredList = computed(() => {
  if (!envFilter.value) return list.value
  return list.value.filter((b: any) => b.env === envFilter.value)
})
const filteredTotal = computed(() => {
  if (!envFilter.value) return total.value
  return filteredList.value.length
})

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
      projectMode: 'existing' as 'existing' | 'new',
      newProjectName: '',
      newProjectDesc: '',
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
      projectMode: projectsList.value.length === 0 ? 'new' as 'existing' | 'new' : 'existing' as 'existing' | 'new',
      newProjectName: '',
      newProjectDesc: '',
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
        await request({ url: '/businesses/' + editingBizId.value, method: 'PUT', data: form })
        ElMessage.success('更新成功')
      } else {
        // 若用户选择"新建项目并关联"，先创建项目，再关联业务
        if (form.projectMode === 'new') {
          const newProject: any = await request({
            url: '/projects',
            method: 'POST',
            data: { name: form.newProjectName.trim(), description: form.newProjectDesc }
          })
          const createdId = newProject?.projectId || newProject?.id || (newProject?.data?.projectId || newProject?.data?.id)
          if (!createdId) {
            throw new Error('创建项目失败，无法关联业务')
          }
          form.projectId = createdId
          ElMessage.success('已自动创建项目「' + form.newProjectName.trim() + '」')
        }
        const payload: any = {
          projectId: form.projectId,
          code: form.code,
          name: form.name,
          env: form.env,
          description: form.description
        }
        await request({ url: '/businesses', method: 'POST', data: payload })
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      loadProjectsList()
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
        await request({ url: '/businesses/' + row.businessId, method: 'DELETE' })
        ElMessage.success('删除成功')
        loadList()
      } catch (e: any) {
        ElMessage.error(e?.message || '删除失败')
      }
    }).catch(() => {})
}

// --------- 成员分配 公共 ---------
const allUsers = ref<any[]>([])
const usersLoaded = ref(false)

async function loadUsers() {
  if (usersLoaded.value) return
  try {
    const r: any = await request({ url: '/account', method: 'GET', params: { pageSize: 9999 } })
    allUsers.value = r?.list || r || []
    usersLoaded.value = true
  } catch (e: any) {
    try {
      const r2: any = await request({ url: '/users', method: 'GET', params: { pageSize: 9999 } })
      allUsers.value = r2?.list || r2 || []
      usersLoaded.value = true
    } catch (e2: any) {
      ElMessage.error('加载账号列表失败')
    }
  }
}

// --------- 项目成员 ---------
const currentProject = ref<any>(null)
const projectMemberDialogVisible = ref(false)
const projectMemberLoading = ref(false)
const projectMemberMap = reactive<Record<string, boolean>>({})
const projectMemberAll = ref(false)

async function openProjectMemberDialog(row: any) {
  currentProject.value = row
  projectMemberDialogVisible.value = true
  projectMemberLoading.value = true
  Object.keys(projectMemberMap).forEach((k) => delete projectMemberMap[k])
  try {
    await loadUsers()
    const members: any = await request({ url: '/projects/' + row.projectId + '/members', method: 'GET' })
    const memberList: any[] = Array.isArray(members) ? members : (members?.list || [])
    memberList.forEach((m: any) => {
      const uid = m.userId || m.user_id
      if (uid) projectMemberMap[uid] = true
    })
    projectMemberAll.value = allUsers.value.length > 0 && allUsers.value.every((u: any) => projectMemberMap[u.userId])
  } catch (e: any) {
    ElMessage.error('加载成员列表失败')
  } finally {
    projectMemberLoading.value = false
  }
}

function toggleAllProjectMembers(val: boolean) {
  allUsers.value.forEach((u: any) => { projectMemberMap[u.userId] = val })
}

async function saveProjectMembers() {
  if (!currentProject.value) return
  const userIds = allUsers.value.filter((u: any) => projectMemberMap[u.userId]).map((u: any) => u.userId)
  try {
    await request({
      url: '/projects/' + currentProject.value.projectId + '/members',
      method: 'POST',
      data: { userIds, role: 'developer' }
    })
    ElMessage.success('已保存')
    projectMemberDialogVisible.value = false
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  }
}

// --------- 业务成员 ---------
const currentBiz = ref<any>(null)
const bizMemberDialogVisible = ref(false)
const bizMemberLoading = ref(false)
const bizMemberMap = reactive<Record<string, boolean>>({})
const bizMemberAll = ref(false)

async function openBizMemberDialog(row: any) {
  currentBiz.value = row
  bizMemberDialogVisible.value = true
  bizMemberLoading.value = true
  Object.keys(bizMemberMap).forEach((k) => delete bizMemberMap[k])
  try {
    await loadUsers()
    const members: any = await request({ url: '/businesses/' + row.businessId + '/members', method: 'GET' })
    const memberList: any[] = Array.isArray(members) ? members : (members?.list || [])
    memberList.forEach((m: any) => {
      const uid = m.userId || m.user_id
      if (uid) bizMemberMap[uid] = true
    })
    bizMemberAll.value = allUsers.value.length > 0 && allUsers.value.every((u: any) => bizMemberMap[u.userId])
  } catch (e: any) {
    ElMessage.error('加载成员列表失败')
  } finally {
    bizMemberLoading.value = false
  }
}

function toggleAllBizMembers(val: boolean) {
  allUsers.value.forEach((u: any) => { bizMemberMap[u.userId] = val })
}

async function saveBizMembers() {
  if (!currentBiz.value) return
  const userIds = allUsers.value.filter((u: any) => bizMemberMap[u.userId]).map((u: any) => u.userId)
  try {
    await request({
      url: '/businesses/' + currentBiz.value.businessId + '/members',
      method: 'POST',
      data: { userIds, role: 'developer' }
    })
    ElMessage.success('已保存')
    bizMemberDialogVisible.value = false
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  }
}

// --------- 成员数 / 业务数 懒统计 ---------
const memberCountCache = reactive<Record<string, number>>({})
const bizCountCache = reactive<Record<string, number>>({})

function getMemberCount(id: string, kind: string): number {
  const key = kind + '_' + id
  if (memberCountCache[key] !== undefined) return memberCountCache[key]
  memberCountCache[key] = 0
  const url = kind === 'project' ? '/projects/' + id + '/members' : '/businesses/' + id + '/members'
  request({ url, method: 'GET' }).then((members: any) => {
    const arr: any[] = Array.isArray(members) ? members : (members?.list || [])
    memberCountCache[key] = arr.length
  }).catch(() => {})
  return memberCountCache[key]
}

function getBizCount(projectId: string): number {
  if (bizCountCache[projectId] !== undefined) return bizCountCache[projectId]
  bizCountCache[projectId] = 0
  request({ url: '/businesses', method: 'GET', params: { projectId, pageSize: 1 } })
    .then((r: any) => {
      const total2 = r?.total ?? (Array.isArray(r) ? r.length : 0)
      bizCountCache[projectId] = total2
    }).catch(() => {})
  return bizCountCache[projectId]
}

onMounted(async () => {
  await loadProjectsList()
  if (activeTab.value === 'businesses') {
    loadList()
  }
})
</script>
