<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DB老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <div class="page-container">
    <div class="page-header">账号管理</div>
    <div class="card">
      <div class="toolbar">
        <el-input v-model="keyword" placeholder="搜索用户名/邮箱/显示名" clearable style="width:260px;" />
        <el-button type="primary" @click="loadList" :icon="SearchIcon">搜索</el-button>
        <el-button type="primary" @click="openDialog()" :icon="PlusIcon">新建账号</el-button>
        <ColumnToggle v-model="colVisible" :columns="columns" />
      </div>

      <el-table :data="list" style="width:100%;" v-loading="loading" stripe :default-sort="{ prop: 'createdAt', order: 'descending' }">
        <el-table-column prop="username" label="用户名" sortable :resizable="true" v-if="colVisible.username" min-width="140" />
        <el-table-column prop="displayName" label="显示名" sortable :resizable="true" v-if="colVisible.displayName" min-width="140" />
        <el-table-column prop="email" label="邮箱" sortable :resizable="true" v-if="colVisible.email" min-width="200" />
        <el-table-column label="角色" sortable :resizable="true" v-if="colVisible.role" min-width="110" prop="role">
          <template #default="{ row }">
            <el-tag v-if="row.role === 'admin'" type="danger">管理员</el-tag>
            <el-tag v-else type="info">成员</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" sortable :resizable="true" v-if="colVisible.status" min-width="100" prop="status">
          <template #default="{ row }">
            <el-tag v-if="row.status === 'active'" type="success">启用</el-tag>
            <el-tag v-else type="info">禁用</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" sortable :resizable="true" v-if="colVisible.createdAt" min-width="180" />
        <el-table-column label="操作" min-width="280" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openDialog(row)">编辑</el-button>
            <el-button size="small" type="warning" @click="resetPwd(row)">重置密码</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div style="margin-top:16px;display:flex;justify-content:flex-end;">
        <el-pagination
          v-model:current-page="current"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 30, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="loadList"
          @size-change="loadList"
        />
      </div>
    </div>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑账号' : '新建账号'" width="480px">
      <el-form :model="form" label-width="100px" :rules="rules" ref="formRef">
        <el-form-item label="用户名" prop="username"><el-input v-model="form.username" :disabled="isEdit" /></el-form-item>
        <el-form-item v-if="!isEdit" label="密码" prop="password"><el-input v-model="form.password" type="password" show-password /></el-form-item>
        <el-form-item label="显示名"><el-input v-model="form.displayName" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="form.email" /></el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.role" style="width:100%;">
            <el-option label="管理员" value="admin" />
            <el-option label="成员" value="member" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="isEdit" label="状态">
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

    <el-dialog v-model="pwdDialogVisible" title="重置密码" width="420px">
      <el-form label-width="100px">
        <el-form-item label="新密码">
          <el-input v-model="newPwd" type="password" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="pwdDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="doResetPwd">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Search as SearchIcon, Plus as PlusIcon } from '@element-plus/icons-vue'
import { listAccounts, createAccount, updateAccount, deleteAccount, resetAccountPassword } from '@/api/auth'
import ColumnToggle from '@/components/ColumnToggle.vue'

const columns = [
  { key: 'username', label: '用户名' },
  { key: 'displayName', label: '显示名' },
  { key: 'email', label: '邮箱' },
  { key: 'role', label: '角色' },
  { key: 'status', label: '状态' },
  { key: 'createdAt', label: '创建时间' }
]
let colVisible = reactive<Record<string, boolean>>({
  username: true, displayName: true, email: true,
  role: true, status: true, createdAt: true
})

const loading = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const current = ref(1)
const pageSize = ref(50)
const keyword = ref('')
const dialogVisible = ref(false)
const pwdDialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref('')
const formRef = ref<FormInstance>()
const newPwd = ref('')
const form = reactive({ username: '', password: '', email: '', displayName: '', role: 'member', status: 'active' })
const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名' }],
  password: [{ required: true, message: '请输入密码', min: 6 }]
}

async function loadList() {
  loading.value = true
  try {
    const r: any = await listAccounts(current.value, pageSize.value, keyword.value)
    list.value = r.list || []
    total.value = r.total || 0
  } finally {
    loading.value = false
  }
}

function openDialog(row?: any) {
  if (row) {
    isEdit.value = true
    editId.value = row.userId
    form.username = row.username
    form.password = ''
    form.email = row.email || ''
    form.displayName = row.displayName || ''
    form.role = row.role || 'member'
    form.status = row.status || 'active'
  } else {
    isEdit.value = false
    editId.value = ''
    Object.assign(form, { username: '', password: '', email: '', displayName: '', role: 'member', status: 'active' })
  }
  dialogVisible.value = true
}

async function save() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      if (isEdit.value) {
        await updateAccount(editId.value, { email: form.email, displayName: form.displayName, role: form.role, status: form.status })
        ElMessage.success('更新成功')
      } else {
        await createAccount(form)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      loadList()
    } catch (e: any) {
      ElMessage.error(e?.message || '保存失败')
    }
  })
}

function resetPwd(row: any) {
  editId.value = row.userId
  newPwd.value = ''
  pwdDialogVisible.value = true
}

async function doResetPwd() {
  if (newPwd.value.length < 6) {
    ElMessage.warning('密码至少 6 位')
    return
  }
  try {
    await resetAccountPassword(editId.value, newPwd.value)
    ElMessage.success('重置成功')
    pwdDialogVisible.value = false
  } catch (e: any) {
    ElMessage.error(e?.message || '重置失败')
  }
}

function handleDelete(row: any) {
  ElMessageBox.confirm(`确认删除账号「${row.username}」？`, '提示', { type: 'warning' })
    .then(async () => {
      try {
        await deleteAccount(row.userId)
        ElMessage.success('删除成功')
        loadList()
      } catch (e: any) {
        ElMessage.error(e?.message || '删除失败')
      }
    }).catch(() => {})
}

onMounted(loadList)
</script>

