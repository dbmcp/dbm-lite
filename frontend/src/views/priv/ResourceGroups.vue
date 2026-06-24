<template>
  <div class="priv-page">
    <div class="page-header">
      <h2 class="title">资源组管理</h2>
      <div class="actions">
        <el-input v-model="keyword" placeholder="搜索名称" style="width: 220px; margin-right: 12px" clearable />
        <el-button type="primary" @click="openCreateDialog">
          <el-icon><Plus /></el-icon>
          新建资源组
        </el-button>
      </div>
    </div>

    <el-table :data="groups" border stripe style="width: 100%" v-loading="loading">
      <el-table-column label="名称" prop="name" min-width="180" show-overflow-tooltip />
      <el-table-column label="备注" prop="remark" min-width="260" show-overflow-tooltip />
      <el-table-column label="已绑定用户数" prop="userCount" width="140" align="center" show-overflow-tooltip />
      <el-table-column label="已绑定数据源数" prop="datasourceCount" width="160" align="center" show-overflow-tooltip />
      <el-table-column label="创建时间" prop="createdAt" width="180" show-overflow-tooltip />
      <el-table-column label="操作" width="320" fixed="right" show-overflow-tooltip>
        <template #default="scope">
          <el-button size="small" link type="primary" @click="openBindUsers(scope.row)">绑定用户</el-button>
          <el-button size="small" link type="primary" @click="openBindDS(scope.row)">绑定数据源</el-button>
          <el-button size="small" link type="warning" @click="openEdit(scope.row)">编辑</el-button>
          <el-button size="small" link type="danger" @click="remove(scope.row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination-wrap">
      <el-pagination
        v-model:current-page="current"
        v-model:page-size="pageSize"
        :total="total"
        layout="total, sizes, prev, pager, next"
        :page-sizes="[100, 200, 500, 1000]"
        background
        @current-change="load"
        @size-change="load"
      />
    </div>

    <!-- 新建 / 编辑 -->
    <el-dialog v-model="dialogVisible" :title="editing ? '编辑资源组' : '新建资源组'" width="520px">
      <el-form :model="form" label-width="96px">
        <el-form-item label="组名称">
          <el-input v-model="form.name" placeholder="请输入名称" maxlength="64" show-word-limit />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="3" maxlength="256" show-word-limit placeholder="选填" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <!-- 绑定用户 -->
    <el-dialog v-model="bindUsersVisible" :title="'绑定用户 - ' + (bindingRow?.name || '')" width="600px">
      <el-select
        v-model="selectedUserIds"
        multiple
        filterable
        collapse-tags
        collapse-tags-tooltip
        placeholder="选择用户"
        style="width: 100%"
      >
        <el-option v-for="u in userOptions" :key="u.userId" :label="u.displayName || u.username" :value="u.userId" />
      </el-select>
      <template #footer>
        <el-button @click="bindUsersVisible = false">取消</el-button>
        <el-button type="primary" @click="doBindUsers">保存绑定</el-button>
      </template>
    </el-dialog>

    <!-- 绑定数据源 -->
    <el-dialog v-model="bindDSVisible" :title="'绑定数据源 - ' + (bindingRow?.name || '')" width="600px">
      <el-select
        v-model="selectedDsIds"
        multiple
        filterable
        collapse-tags
        collapse-tags-tooltip
        placeholder="选择数据源"
        style="width: 100%"
      >
        <el-option v-for="d in dsOptions" :key="d.id || d.datasourceId" :label="d.name" :value="d.id || d.datasourceId" />
      </el-select>
      <template #footer>
        <el-button @click="bindDSVisible = false">取消</el-button>
        <el-button type="primary" @click="doBindDS">保存绑定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  listGroups,
  createGroup,
  updateGroup,
  deleteGroup,
  bindGroupUsers,
  bindGroupDatasources
} from '@/api/priv'
import { listDatasources, listAllDatasources } from '@/api/datasource'
import { listAccounts } from '@/api/auth'

const loading = ref(false)
const groups = ref<any[]>([])
const total = ref(0)
const current = ref(1)
const pageSize = ref(100)
const keyword = ref('')

const dialogVisible = ref(false)
const editing = ref<any>(null)
const form = reactive({ name: '', remark: '' })

const bindUsersVisible = ref(false)
const bindDSVisible = ref(false)
const bindingRow = ref<any>(null)
const selectedUserIds = ref<string[]>([])
const selectedDsIds = ref<string[]>([])
const userOptions = ref<any[]>([])
const dsOptions = ref<any[]>([])

async function load() {
  try {
    loading.value = true
    const res: any = await listGroups(current.value, pageSize.value, keyword.value)
    groups.value = res?.list || []
    total.value = res?.total || 0
  } catch (e: any) {
    ElMessage.error(e?.message || '加载失败')
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  editing.value = null
  form.name = ''
  form.remark = ''
  dialogVisible.value = true
}

function openEdit(row: any) {
  editing.value = row
  form.name = row.name
  form.remark = row.remark || ''
  dialogVisible.value = true
}

async function save() {
  if (!form.name.trim()) {
    ElMessage.warning('名称不能为空')
    return
  }
  try {
    if (editing.value) {
      await updateGroup(editing.value.groupId, { name: form.name, remark: form.remark })
      ElMessage.success('编辑成功')
    } else {
      await createGroup({ name: form.name, remark: form.remark })
      ElMessage.success('新建成功')
    }
    dialogVisible.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  }
}

async function remove(row: any) {
  try {
    await ElMessageBox.confirm('确定删除资源组「' + row.name + '」？删除将同时解除关联。', '提示', {
      type: 'warning'
    })
    await deleteGroup(row.groupId)
    ElMessage.success('删除成功')
    load()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '删除失败')
  }
}

async function openBindUsers(row: any) {
  bindingRow.value = row
  selectedUserIds.value = []
  try {
    const [users] = await Promise.all([safeListUsers()])
    userOptions.value = users
    bindUsersVisible.value = true
  } catch (e) {}
}

async function safeListUsers() {
  try {
    const res: any = await (listAccounts as any)()
    if (res && Array.isArray(res.list)) return res.list
    return []
  } catch (e) {
    return []
  }
}

async function doBindUsers() {
  if (!bindingRow.value) return
  try {
    await bindGroupUsers(bindingRow.value.groupId, selectedUserIds.value)
    ElMessage.success('绑定成功')
    bindUsersVisible.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e?.message || '绑定失败')
  }
}

async function openBindDS(row: any) {
  bindingRow.value = row
  selectedDsIds.value = []
  try {
    const res: any = await listAllDatasources(1, 10000)
    dsOptions.value = (res?.list || []).map((x: any) => ({ id: x.datasourceId, name: x.name }))
    if (!dsOptions.value.length) {
      // 兼容其他字段
      dsOptions.value = (res?.list || []).map((x: any) => ({ id: x.id, name: x.name }))
    }
    bindDSVisible.value = true
  } catch (e) {}
}

async function doBindDS() {
  if (!bindingRow.value) return
  try {
    await bindGroupDatasources(bindingRow.value.groupId, selectedDsIds.value)
    ElMessage.success('绑定成功')
    bindDSVisible.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e?.message || '绑定失败')
  }
}

onMounted(load)
</script>

<style scoped>
.priv-page {
  padding: 16px 20px;
  background: #f5f7fa;
  min-height: calc(100vh - 60px);
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
  padding: 16px 20px;
  border-radius: 8px;
  margin-bottom: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}
.title {
  margin: 0;
  font-size: 18px;
  color: #303133;
}
.actions {
  display: flex;
  align-items: center;
}
.pagination-wrap {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
  background: #fff;
  padding: 12px 20px;
  border-radius: 8px;
}
</style>
