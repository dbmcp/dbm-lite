<template>
  <div class="page-container">
    <div class="page-header">
      <div class="page-title">角色管理</div>
      <div class="page-sub">配置平台角色，用于授予用户不同的访问与操作权限。</div>
    </div>

    <div class="toolbar">
      <el-input v-model="keyword" placeholder="搜索角色名 / 描述" clearable style="width:220px" />
      <el-select v-model="statusFilter" placeholder="状态" clearable style="width:120px">
        <el-option label="启用" value="active" />
        <el-option label="禁用" value="inactive" />
      </el-select>
      <el-button type="primary" plain @click="onSearch">查询</el-button>
      <el-button type="success" @click="openCreateDialog">新建角色</el-button>
    </div>

    <el-table :data="list" border stripe v-loading="loading" style="width:100%">
      <el-table-column prop="name" label="角色名" min-width="180" />
      <el-table-column prop="codes" label="角色编码" width="180" />
      <el-table-column prop="description" label="描述" min-width="240" />
      <el-table-column prop="permissions" label="权限标识" min-width="260">
        <template #default="{ row }">
          <template v-if="row.permissions && row.permissions.length">
            <el-tag v-for="p in row.permissions.slice(0, 8)" :key="p" size="small" style="margin-right:4px;margin-bottom:4px">{{ p }}</el-tag>
            <span v-if="row.permissions.length > 8" style="color:#909399">等 {{ row.permissions.length }} 项</span>
          </template>
          <span v-else style="color:#c0c4cc">—</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.status === 'active'" type="success">启用</el-tag>
          <el-tag v-else type="info">禁用</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="创建时间" width="180" />
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" link @click="openEditDialog(row)">编辑</el-button>
          <el-button size="small" type="danger" link @click="confirmDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pager">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="size"
        :total="total"
        :page-sizes="[100, 200, 500, 1000]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="onSearch"
        @current-change="onSearch"
      />
    </div>

    <el-dialog v-model="formVisible" :title="formMode === 'create' ? '新建角色' : '编辑角色'" width="560px">
      <el-form :model="form" label-width="110px">
        <el-form-item label="角色名" required>
          <el-input v-model="form.name" placeholder="如 运维工程师" />
        </el-form-item>
        <el-form-item label="角色编码">
          <el-input v-model="form.codes" placeholder="可选，英文标识" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width:100%">
            <el-option label="启用" value="active" />
            <el-option label="禁用" value="inactive" />
          </el-select>
        </el-form-item>
        <el-form-item label="权限标识">
          <el-select
            v-model="form.permissions"
            multiple
            filterable
            allow-create
            default-first-option
            placeholder="回车新增权限标识，可多选"
            style="width:100%"
          >
            <el-option v-for="p in permissionOptions" :key="p" :label="p" :value="p" />
          </el-select>
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
import { ferryListRoles, ferryCreateRole, ferryUpdateRole, ferryDeleteRole } from '@/api/ferry'

const keyword = ref('')
const statusFilter = ref('')
const loading = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const size = ref(100)

const permissionOptions = [
  'datasource:manage',
  'sql:read',
  'sql:write',
  'data:export',
  'audit:read',
  'user:read',
  'system:settings',
  'workflow:design'
]

function onSearch() {
  page.value = 1
  loadList()
}

async function loadList() {
  loading.value = true
  try {
    const res: any = await ferryListRoles({
      page: page.value,
      size: size.value,
      keyword: keyword.value,
      status: statusFilter.value
    })
    const items: any[] = Array.isArray(res?.data) ? res.data : Array.isArray(res) ? res : []
    list.value = items.map((r: any) => ({
      id: r.id ?? r.roleId ?? '',
      name: r.name ?? '',
      codes: r.codes ?? '',
      description: r.description ?? '',
      status: r.status ?? 'active',
      permissions: Array.isArray(r.permissions) ? r.permissions : [],
      createdAt: r.createdAt ?? ''
    }))
    total.value = typeof res?.total === 'number' ? res.total : list.value.length
  } catch (e: any) {
    ElMessage.error(e?.message || '加载失败')
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
  codes: '',
  description: '',
  status: 'active',
  permissions: [] as string[]
})

function resetForm() {
  form.name = ''
  form.codes = ''
  form.description = ''
  form.status = 'active'
  form.permissions = []
}

function openCreateDialog() {
  resetForm()
  formMode.value = 'create'
  formVisible.value = true
}

function openEditDialog(row: any) {
  resetForm()
  formMode.value = 'edit'
  editId.value = row.id
  form.name = row.name
  form.codes = row.codes || ''
  form.description = row.description || ''
  form.status = row.status || 'active'
  form.permissions = (row.permissions || []).slice()
  formVisible.value = true
}

async function submitForm() {
  if (!form.name.trim()) {
    ElMessage.warning('请输入角色名')
    return
  }
  submitting.value = true
  try {
    if (formMode.value === 'create') {
      await ferryCreateRole({
        name: form.name.trim(),
        codes: form.codes,
        description: form.description,
        status: form.status,
        permissions: form.permissions
      })
    } else {
      await ferryUpdateRole(editId.value, {
        name: form.name.trim(),
        codes: form.codes,
        description: form.description,
        status: form.status,
        permissions: form.permissions
      })
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
  ElMessageBox.confirm(`确认删除角色 "${row.name}" 吗？此操作不可恢复。`, '提示', {
    type: 'warning'
  })
    .then(async () => {
      try {
        await ferryDeleteRole(row.id)
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
  flex-wrap: wrap;
}
.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 12px;
}
</style>
