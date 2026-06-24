<template>
  <div class="page-container">
    <div class="page-header">
      <div class="page-title">菜单管理</div>
      <div class="page-sub">维护系统菜单结构、路由与权限标识。</div>
    </div>

    <div class="toolbar">
      <el-input v-model="keyword" placeholder="搜索菜单名称" clearable style="width:200px" />
      <el-button type="success" @click="openCreateDialog(null)">新建菜单</el-button>
      <el-button type="primary" @click="loadList">刷新</el-button>
    </div>

    <el-table :data="treeData" v-loading="loading" border row-key="menuId" :tree-props="{ children: 'children' }" style="width:100%">
      <el-table-column prop="name" label="菜单名称" min-width="180" />
      <el-table-column prop="icon" label="图标" width="100">
        <template #default="{ row }">
          <el-icon v-if="row.icon"><component :is="row.icon" /></el-icon>
          <span v-else style="color:#c0c4cc">—</span>
        </template>
      </el-table-column>
      <el-table-column prop="type" label="类型" width="100" />
      <el-table-column prop="path" label="路径" min-width="220" />
      <el-table-column prop="component" label="组件" min-width="220" />
      <el-table-column prop="perm" label="权限标识" min-width="160" />
      <el-table-column prop="sortOrder" label="排序" width="80" />
      <el-table-column label="可见" width="80">
        <template #default="{ row }">
          <el-tag size="small" v-if="row.visible !== 'hide'">显示</el-tag>
          <el-tag size="small" type="info" v-else>隐藏</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag size="small" type="success" v-if="row.status === 'active'">启用</el-tag>
          <el-tag size="small" type="info" v-else>禁用</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="260" fixed="right">
        <template #default="{ row }">
          <el-button size="small" link type="primary" @click="openCreateDialog(row)">新增下级</el-button>
          <el-button size="small" link type="primary" @click="openEditDialog(row)">编辑</el-button>
          <el-button size="small" link type="danger" @click="confirmDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="formVisible" :title="formMode === 'create' ? '新建菜单' : '编辑菜单'" width="560px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="上级菜单">
          <el-tree-select
            v-model="form.parentId"
            :data="parentCandidates"
            node-key="menuId"
            :props="{ label: 'name', children: 'children' }"
            check-strictly
            :render-after-expand="false"
            style="width:100%"
            placeholder="不选择则为顶级菜单"
          />
        </el-form-item>
        <el-form-item label="菜单名称" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="form.type" style="width:100%">
            <el-option label="菜单" value="menu" />
            <el-option label="按钮" value="button" />
            <el-option label="页面" value="page" />
          </el-select>
        </el-form-item>
        <el-form-item label="图标">
          <el-input v-model="form.icon" placeholder="Element Plus icon name" />
        </el-form-item>
        <el-form-item label="路径">
          <el-input v-model="form.path" placeholder="如 /platform/config" />
        </el-form-item>
        <el-form-item label="组件">
          <el-input v-model="form.component" placeholder="组件路径" />
        </el-form-item>
        <el-form-item label="权限标识">
          <el-input v-model="form.perm" placeholder="如 platform:config:read" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sortOrder" :min="0" :max="9999" />
        </el-form-item>
        <el-form-item label="可见">
          <el-select v-model="form.visible" style="width:100%">
            <el-option label="显示" value="show" />
            <el-option label="隐藏" value="hide" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width:100%">
            <el-option label="启用" value="active" />
            <el-option label="禁用" value="inactive" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" />
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
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ferryListMenus, ferryCreateMenu, ferryUpdateMenu, ferryDeleteMenu } from '@/api/ferry'

const keyword = ref('')
const loading = ref(false)
const treeData = ref<any[]>([])
const parentCandidates = ref<any[]>([])

async function loadList() {
  loading.value = true
  try {
    const res = await ferryListMenus({ keyword: keyword.value })
    treeData.value = (res?.data as any[]) || []
    parentCandidates.value = treeData.value
  } finally {
    loading.value = false
  }
}

const formVisible = ref(false)
const submitting = ref(false)
const formMode = ref<'create' | 'edit'>('create')
const editId = ref('')
const form = reactive({
  name: '',
  parentId: '',
  type: 'menu',
  icon: '',
  path: '',
  component: '',
  perm: '',
  sortOrder: 0,
  visible: 'show',
  status: 'active',
  remark: ''
})

function resetForm(parent: any) {
  form.name = ''
  form.parentId = parent ? parent.menuId : ''
  form.type = 'menu'
  form.icon = ''
  form.path = ''
  form.component = ''
  form.perm = ''
  form.sortOrder = 0
  form.visible = 'show'
  form.status = 'active'
  form.remark = ''
}

function openCreateDialog(parent: any) {
  resetForm(parent)
  formMode.value = 'create'
  formVisible.value = true
}

function openEditDialog(row: any) {
  formMode.value = 'edit'
  editId.value = row.menuId
  form.name = row.name
  form.parentId = row.parentId || ''
  form.type = row.type || 'menu'
  form.icon = row.icon || ''
  form.path = row.path || ''
  form.component = row.component || ''
  form.perm = row.perm || ''
  form.sortOrder = row.sortOrder ?? 0
  form.visible = row.visible || 'show'
  form.status = row.status || 'active'
  form.remark = row.remark || ''
  formVisible.value = true
}

async function submitForm() {
  if (!form.name.trim()) {
    ElMessage.warning('请输入菜单名称')
    return
  }
  submitting.value = true
  try {
    const payload = {
      name: form.name.trim(),
      parentId: form.parentId || '',
      type: form.type || 'menu',
      icon: form.icon,
      path: form.path,
      component: form.component,
      perm: form.perm,
      sortOrder: form.sortOrder,
      visible: form.visible,
      status: form.status,
      remark: form.remark
    }
    if (formMode.value === 'create') {
      await ferryCreateMenu(payload)
    } else {
      await ferryUpdateMenu(editId.value, payload)
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
  ElMessageBox.confirm(`确认删除菜单 "${row.name}" 吗？（有子菜单时禁止删除）`, '提示', { type: 'warning' })
    .then(async () => {
      try {
        await ferryDeleteMenu(row.menuId)
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
.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}
</style>
