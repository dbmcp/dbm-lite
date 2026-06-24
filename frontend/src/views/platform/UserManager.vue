<template>
  <div class="page-container">
    <div class="page-header">
      <div class="page-title">用户管理</div>
      <div class="page-sub">管理平台用户账号、角色、状态与密码。</div>
    </div>

    <div class="toolbar">
      <el-input v-model="keyword" placeholder="搜索用户名 / 显示名 / 邮箱" clearable style="width:280px" />
      <el-select v-model="filterStatus" placeholder="状态" clearable style="width:120px">
        <el-option label="启用" value="active" />
        <el-option label="禁用" value="inactive" />
      </el-select>
      <el-select v-model="filterRole" placeholder="角色" clearable style="width:140px">
        <el-option label="管理员" value="admin" />
        <el-option label="普通成员" value="member" />
      </el-select>
      <el-button type="primary" @click="loadList">搜索</el-button>
      <el-button type="success" @click="openCreateDialog">新建用户</el-button>
    </div>

    <el-table :data="list" border stripe v-loading="loading" style="width:100%">
      <el-table-column prop="username" label="用户名" min-width="140" />
      <el-table-column prop="displayName" label="显示名" min-width="140" />
      <el-table-column prop="email" label="邮箱" min-width="200" />
      <el-table-column prop="phone" label="联系电话" width="140" />
      <el-table-column label="角色" width="120">
        <template #default="{ row }">
          <el-tag v-if="row.role === 'admin'" type="danger">管理员</el-tag>
          <el-tag v-else>普通成员</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.status === 'active'" type="success">启用</el-tag>
          <el-tag v-else type="info">禁用</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="创建时间" width="180" />
      <el-table-column label="操作" width="320" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" link @click="openEditDialog(row)">编辑</el-button>
          <el-button size="small" type="warning" link @click="openResetPassword(row)">重置密码</el-button>
          <el-button size="small" type="danger" link @click="confirmDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pager">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="size"
        :total="total"
        :page-sizes="[10, 30, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadList"
        @current-change="loadList"
      />
    </div>

    <el-dialog v-model="formVisible" :title="formMode === 'create' ? '新建用户' : '编辑用户'" width="520px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="用户名" required>
          <el-input v-model="form.username" :disabled="formMode === 'edit'" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item v-if="formMode === 'create'" label="密码" required>
          <el-input v-model="form.password" type="password" show-password placeholder="至少 6 位" />
        </el-form-item>
        <el-form-item label="显示名">
          <el-input v-model="form.displayName" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email" placeholder="可选" />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="form.phone" placeholder="可选" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.role" style="width:100%">
            <el-option label="管理员" value="admin" />
            <el-option label="普通成员" value="member" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="formMode === 'edit'" label="状态">
          <el-select v-model="form.status" style="width:100%">
            <el-option label="启用" value="active" />
            <el-option label="禁用" value="inactive" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="pwdVisible" title="重置密码" width="420px">
      <el-form label-width="100px">
        <el-form-item label="新密码" required>
          <el-input v-model="newPwd" type="password" show-password placeholder="至少 6 位" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="pwdVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitResetPassword">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ferryListUsers,
  ferryCreateUser,
  ferryUpdateUser,
  ferryDeleteUser,
  ferryResetUserPassword
} from '@/api/ferry'

const keyword = ref('')
const filterStatus = ref('')
const filterRole = ref('')
const loading = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const size = ref(30)

async function loadList() {
  loading.value = true
  try {
    const res = await ferryListUsers({
      page: page.value,
      size: size.value,
      keyword: keyword.value,
      status: filterStatus.value,
      role: filterRole.value
    })
    list.value = (res?.data as any[]) || []
    total.value = (res?.total as number) || list.value.length
  } finally {
    loading.value = false
  }
}

// 创建 / 编辑
const formVisible = ref(false)
const submitting = ref(false)
const formMode = ref<'create' | 'edit'>('create')
const editId = ref('')
const form = reactive({
  username: '',
  password: '',
  displayName: '',
  email: '',
  phone: '',
  role: 'member',
  status: 'active'
})

function resetForm() {
  form.username = ''
  form.password = ''
  form.displayName = ''
  form.email = ''
  form.phone = ''
  form.role = 'member'
  form.status = 'active'
}

function openCreateDialog() {
  resetForm()
  formMode.value = 'create'
  formVisible.value = true
}

function openEditDialog(row: any) {
  resetForm()
  formMode.value = 'edit'
  editId.value = row.userId
  form.username = row.username
  form.displayName = row.displayName
  form.email = row.email
  form.phone = row.phone
  form.role = row.role || 'member'
  form.status = row.status || 'active'
  formVisible.value = true
}

async function submitForm() {
  if (formMode.value === 'create') {
    if (!form.username.trim() || !form.password || form.password.length < 6) {
      ElMessage.warning('请输入用户名，且密码至少 6 位')
      return
    }
  }
  submitting.value = true
  try {
    if (formMode.value === 'create') {
      await ferryCreateUser({
        username: form.username.trim(),
        password: form.password,
        displayName: form.displayName,
        email: form.email,
        phone: form.phone,
        role: form.role,
        status: form.status
      })
    } else {
      await ferryUpdateUser(editId.value, {
        displayName: form.displayName,
        email: form.email,
        phone: form.phone,
        role: form.role,
        status: form.status
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
  ElMessageBox.confirm(`确认删除用户 "${row.username}" 吗？此操作不可恢复。`, '提示', {
    type: 'warning',
    confirmButtonText: '确认删除',
    cancelButtonText: '取消'
  })
    .then(async () => {
      try {
        await ferryDeleteUser(row.userId)
        ElMessage.success('已删除')
        loadList()
      } catch (e: any) {
        ElMessage.error(e?.message || '删除失败')
      }
    })
    .catch(() => {})
}

// 密码重置
const pwdVisible = ref(false)
const newPwd = ref('')
function openResetPassword(row: any) {
  editId.value = row.userId
  newPwd.value = ''
  pwdVisible.value = true
}

async function submitResetPassword() {
  if (newPwd.value.length < 6) {
    ElMessage.warning('密码至少 6 位')
    return
  }
  submitting.value = true
  try {
    await ferryResetUserPassword(editId.value, newPwd.value)
    ElMessage.success('密码已重置')
    pwdVisible.value = false
  } catch (e: any) {
    ElMessage.error(e?.message || '重置失败')
  } finally {
    submitting.value = false
  }
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
