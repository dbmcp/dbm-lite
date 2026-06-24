<template>
  <div class="page-container">
    <div class="page-header">
      <div>
        <div class="page-title">
          <el-icon><OfficeBuilding /></el-icon>
          <span>部门管理</span>
        </div>
        <div class="page-sub">维护组织架构、部门层级、联系信息与负责人。</div>
      </div>
      <div class="page-actions">
        <el-button @click="loadList"><span class="refresh-icon">⟳</span>刷新</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreateDialog(null)">新建顶级部门</el-button>
      </div>
    </div>

    <el-card class="section-card" shadow="never" v-loading="loading">
      <div class="toolbar">
        <el-input v-model="keyword" placeholder="搜索部门名称 / 负责人" clearable style="width:240px" />
        <el-select v-model="statusFilter" placeholder="状态" clearable style="width:120px">
          <el-option label="启用" value="active" />
          <el-option label="禁用" value="inactive" />
        </el-select>
        <el-button type="primary" plain @click="loadList">查询</el-button>
      </div>

      <el-empty v-if="!treeData.length" description="暂无部门数据" />

      <el-tree
        v-else
        :data="treeData"
        node-key="id"
        :default-expand-all="true"
        :expand-on-click-node="false"
        :filter-node-method="filterNode"
        ref="treeRef"
        class="dept-tree"
      >
        <template #default="{ node, data }">
          <div class="tree-row">
            <div class="tree-left">
              <el-icon class="dept-icon"><OfficeBuilding /></el-icon>
              <span class="dept-name">{{ data.name }}</span>
              <el-tag size="small" :type="data.status === 'active' ? 'success' : 'info'" effect="plain" style="margin-left:8px">
                {{ data.status === 'active' ? '启用' : '禁用' }}
              </el-tag>
              <span class="dept-sub" v-if="data.leader">· 负责人：{{ data.leader }}</span>
              <span class="dept-sub" v-if="data.phone">· {{ data.phone }}</span>
              <span class="dept-sub" v-if="data.email">· {{ data.email }}</span>
            </div>
            <div class="tree-right">
              <el-button size="small" link type="primary" @click.stop="openCreateDialog(data)">新增下级</el-button>
              <el-button size="small" link type="primary" @click.stop="openEditDialog(data)">编辑</el-button>
              <el-button size="small" link type="danger" @click.stop="confirmDelete(data)">删除</el-button>
            </div>
          </div>
        </template>
      </el-tree>
    </el-card>

    <el-dialog v-model="formVisible" :title="formMode === 'create' ? (form.parentId ? '新增下级部门' : '新建顶级部门') : '编辑部门'" width="560px" destroy-on-close>
      <el-form :model="form" label-width="100px">
        <el-form-item label="上级部门">
          <el-tree-select
            v-model="form.parentId"
            :data="parentCandidates"
            node-key="id"
            :props="{ label: 'name', children: 'children' }"
            check-strictly
            :render-after-expand="false"
            style="width:100%"
            placeholder="不选择则为顶级部门"
            clearable
          />
        </el-form-item>
        <el-form-item label="部门名称" required>
          <el-input v-model="form.name" placeholder="如：平台运维部" />
        </el-form-item>
        <el-form-item label="负责人">
          <el-input v-model="form.leader" placeholder="负责人姓名" />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="form.phone" placeholder="联系电话" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email" placeholder="部门邮箱" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sortOrder" :min="0" :max="9999" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio value="active">启用</el-radio>
            <el-radio value="inactive">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" placeholder="备注信息（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { OfficeBuilding, Plus, Refresh } from '@element-plus/icons-vue'
import {
  ferryListDepartments,
  ferryCreateDepartment,
  ferryUpdateDepartment,
  ferryDeleteDepartment
} from '@/api/ferry'

const keyword = ref('')
const statusFilter = ref('')
const loading = ref(false)
const treeData = ref<any[]>([])
const parentCandidates = ref<any[]>([])
const treeRef = ref<any>(null)

function normalizeItem(raw: any): any {
  const r = raw || {}
  return {
    id: r.id ?? r.departmentId ?? r.deptId ?? '',
    parentId: r.parentId ?? '',
    name: r.name ?? '',
    leader: r.leader ?? '',
    phone: r.phone ?? '',
    email: r.email ?? '',
    sortOrder: typeof r.sortOrder === 'number' ? r.sortOrder : 0,
    status: r.status || 'active',
    remark: r.remark ?? '',
    userCount: r.userCount ?? 0,
    children: [] as any[]
  }
}

function buildTree(items: any[]): any[] {
  const map = new Map<string, any>()
  for (const raw of items) {
    const it = normalizeItem(raw)
    if (!it.id) it.id = 'gen_' + Math.random().toString(36).slice(2, 10)
    map.set(it.id, it)
  }
  const roots: any[] = []
  for (const it of map.values()) {
    if (it.parentId && map.has(it.parentId)) {
      map.get(it.parentId).children.push(it)
    } else {
      it.parentId = ''
      roots.push(it)
    }
  }
  const sortTree = (nodes: any[]) => {
    nodes.sort((a, b) => (a.sortOrder || 0) - (b.sortOrder || 0))
    for (const n of nodes) if (n.children && n.children.length) sortTree(n.children)
  }
  sortTree(roots)
  return roots
}

async function loadList() {
  loading.value = true
  try {
    const params: Record<string, any> = {}
    if (keyword.value) params.keyword = keyword.value
    if (statusFilter.value) params.status = statusFilter.value
    const res: any = await ferryListDepartments(params)
    const items: any[] = Array.isArray(res) ? res : Array.isArray(res?.data) ? res.data : []
    treeData.value = buildTree(items)
    parentCandidates.value = buildTree(items)
  } catch (e: any) {
    ElMessage.error(e?.message || '加载失败')
    treeData.value = []
    parentCandidates.value = []
  } finally {
    loading.value = false
  }
}

watch(keyword, (val) => {
  if (treeRef.value && typeof treeRef.value.filter === 'function') {
    treeRef.value.filter(val)
  }
})

function filterNode(value: string, data: any): boolean {
  if (!value) return true
  const text = ((data.name || '') + ' ' + (data.leader || '') + ' ' + (data.phone || '') + ' ' + (data.email || '')).toLowerCase()
  return text.includes(String(value).toLowerCase())
}

const formVisible = ref(false)
const submitting = ref(false)
const formMode = ref<'create' | 'edit'>('create')
const editId = ref('')
const form = reactive({
  name: '',
  parentId: '',
  leader: '',
  phone: '',
  email: '',
  sortOrder: 0,
  status: 'active',
  remark: ''
})

function resetForm(parent: any | null) {
  form.name = ''
  form.parentId = parent ? parent.id : ''
  form.leader = ''
  form.phone = ''
  form.email = ''
  form.sortOrder = 0
  form.status = 'active'
  form.remark = ''
}

function openCreateDialog(parent: any | null) {
  resetForm(parent)
  formMode.value = 'create'
  formVisible.value = true
}

function openEditDialog(data: any) {
  formMode.value = 'edit'
  editId.value = data.id
  form.name = data.name
  form.parentId = data.parentId || ''
  form.leader = data.leader || ''
  form.phone = data.phone || ''
  form.email = data.email || ''
  form.sortOrder = data.sortOrder ?? 0
  form.status = data.status || 'active'
  form.remark = data.remark || ''
  formVisible.value = true
}

async function submitForm() {
  if (!form.name.trim()) {
    ElMessage.warning('请输入部门名称')
    return
  }
  submitting.value = true
  try {
    const payload = {
      name: form.name.trim(),
      parentId: form.parentId || undefined,
      leader: form.leader,
      phone: form.phone,
      email: form.email,
      sortOrder: form.sortOrder,
      status: form.status,
      remark: form.remark
    }
    if (formMode.value === 'create') {
      await ferryCreateDepartment(payload)
    } else {
      await ferryUpdateDepartment(editId.value, payload)
    }
    ElMessage.success('保存成功')
    formVisible.value = false
    loadList()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  } finally {
    submitting.value = false
  }
}

function confirmDelete(row: any) {
  ElMessageBox.confirm(`确认删除部门 "${row.name}" 吗？有子部门时禁止删除。`, '提示', {
    type: 'warning',
    confirmButtonText: '确认删除',
    cancelButtonText: '取消'
  })
    .then(async () => {
      try {
        await ferryDeleteDepartment(row.id)
        ElMessage.success('已删除')
        loadList()
      } catch (e: any) {
        ElMessage.error(e?.message || '删除失败')
      }
    })
    .catch(() => {})
}

onMounted(loadList)
</script>

<style scoped>
.page-container {
  max-width: 1440px;
  margin: 0 auto;
  padding: 16px 20px 24px;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  padding: 4px 2px 16px;
}
.page-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 20px;
  font-weight: 600;
  color: #1f2937;
}
.page-title .el-icon {
  color: #409eff;
  font-size: 22px;
}
.page-sub {
  color: #6b7280;
  font-size: 13px;
  margin-top: 6px;
}
.page-actions { display: flex; gap: 8px; }
.section-card { border-radius: 8px; margin-bottom: 16px; }
.toolbar { display: flex; gap: 8px; margin-bottom: 12px; flex-wrap: wrap; }
.dept-tree { padding: 4px 8px; }
.tree-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  padding: 6px 8px;
  border-bottom: 1px dashed #f0f0f0;
}
.tree-row:hover { background: #f9fafb; }
.tree-left { display: flex; align-items: center; flex-wrap: wrap; gap: 4px; }
.dept-icon { color: #409eff; margin-right: 4px; }
.dept-name { font-weight: 500; color: #1f2937; }
.dept-sub { color: #909399; font-size: 12px; margin-left: 4px; }
.tree-right { display: flex; gap: 4px; flex-shrink: 0; }
</style>
