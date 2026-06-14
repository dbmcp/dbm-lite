<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DB老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <div class="dbuser-page">
    <el-card shadow="never" style="margin-bottom:12px;">
      <div style="display:flex;align-items:center;justify-content:space-between;">
        <div>
          <h2 style="margin:0;color:#409EFF;">数据库权限管理</h2>
          <p style="margin:6px 0 0;color:#909399;font-size:13px;">管理数据库用户和权限（功能开发中）</p>
        </div>
        <el-button type="primary" style="background:#409EFF;border-color:#409EFF;" @click="openCreate">
          <el-icon><Plus /></el-icon>&nbsp;新建数据库用户
        </el-button>
      </div>
    </el-card>

    <el-card shadow="never">
      <div style="margin-bottom:8px;display:flex;justify-content:flex-end;">
        <ColumnToggle v-model="colVisible" :columns="columns" />
      </div>
      <el-table :data="users" border :default-sort="{ prop: 'username', order: 'ascending' }">
        <el-table-column prop="username" label="用户名" sortable :resizable="true" v-if="colVisible.username" min-width="180">
          <template #default="scope">
            <el-icon color="#409EFF"><Lock /></el-icon>&nbsp;{{ scope.row.username }}
          </template>
        </el-table-column>
        <el-table-column prop="datasource" label="所属数据源" sortable :resizable="true" v-if="colVisible.datasource" min-width="180" />
        <el-table-column prop="db" label="数据库" sortable :resizable="true" v-if="colVisible.db" min-width="160" />
        <el-table-column prop="privileges" label="权限" :resizable="true" v-if="colVisible.privileges" min-width="300">
          <template #default="scope">
            <el-tag v-for="p in scope.row.privileges.split(',')" :key="p" type="primary" size="small" style="margin-right:4px;">{{ p }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" sortable :resizable="true" v-if="colVisible.status" min-width="100" align="center" prop="status">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'active' ? 'success' : 'info'" size="small">{{ scope.row.status === 'active' ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="200" align="center" fixed="right">
          <template #default="scope">
            <el-button link type="primary" @click="editPrivileges(scope.row)">权限编辑</el-button>
            <el-button link type="danger" @click="deleteUser(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" title="新建数据库用户" width="520px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="数据源">
          <el-select v-model="form.datasource" placeholder="选择">
            <el-option label="production_db" value="production_db" />
          </el-select>
        </el-form-item>
        <el-form-item label="用户名"><el-input v-model="form.username" /></el-form-item>
        <el-form-item label="密码"><el-input v-model="form.password" type="password" /></el-form-item>
        <el-form-item label="数据库"><el-input v-model="form.db" placeholder="例如：* 或 库名" /></el-form-item>
        <el-form-item label="权限">
          <el-checkbox-group v-model="form.privileges">
            <el-checkbox value="SELECT" />
            <el-checkbox value="INSERT" />
            <el-checkbox value="UPDATE" />
            <el-checkbox value="DELETE" />
            <el-checkbox value="CREATE" />
            <el-checkbox value="DROP" />
            <el-checkbox value="INDEX" />
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" style="background:#409EFF;border-color:#409EFF;" @click="saveUser">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Lock } from '@element-plus/icons-vue'
import request from '@/api/request'
import ColumnToggle from '@/components/ColumnToggle.vue'

const columns = [
  { key: 'username', label: '用户名' },
  { key: 'datasource', label: '所属数据源' },
  { key: 'db', label: '数据库' },
  { key: 'privileges', label: '权限' },
  { key: 'status', label: '状态' }
]
const colVisible = reactive<Record<string, boolean>>({
  username: true, datasource: true, db: true, privileges: true, status: true
})

const dialogVisible = ref(false)
const form = ref<any>({ username: '', password: '', db: '*', datasource: '', privileges: [] as string[] })

const users = ref([
  { id: 1, username: 'app_user', datasource: 'production_db', db: '*', privileges: 'SELECT,INSERT,UPDATE', status: 'active' },
  { id: 2, username: 'read_user', datasource: 'production_db', db: '*', privileges: 'SELECT', status: 'active' },
  { id: 3, username: 'report_user', datasource: 'analytics_db', db: '*', privileges: 'SELECT', status: 'active' },
  { id: 4, username: 'batch_user', datasource: 'production_db', db: '*', privileges: 'SELECT,INSERT,UPDATE,DELETE', status: 'active' },
  { id: 5, username: 'test_user', datasource: 'cms_db', db: 'test_db', privileges: 'ALL', status: 'disabled' }
])

function openCreate() {
  form.value = { username: '', password: '', db: '*', datasource: '', privileges: [] }
  dialogVisible.value = true
}

function editPrivileges(row: any) {
  form.value = { ...row, privileges: row.privileges.split(',') }
  dialogVisible.value = true
}

async function saveUser() {
  try {
    await request({ url: '/ops/db-users', method: 'POST', data: form.value })
  } catch (e) { /* demo */ }
  users.value.push({ id: Date.now(), username: form.value.username, datasource: form.value.datasource, db: form.value.db, privileges: (form.value.privileges || []).join(','), status: 'active' })
  ElMessage.success('已保存（功能开发中）')
  dialogVisible.value = false
}

async function deleteUser(row: any) {
  try {
    await ElMessageBox.confirm('确定删除用户「' + row.username + '」？', '确认', { type: 'warning' })
    try { await request({ url: '/ops/db-users/' + row.id, method: 'DELETE' }) } catch (e) { /* demo */ }
    users.value = users.value.filter((u: any) => u.id !== row.id)
    ElMessage.success('已删除')
  } catch (e) { /* cancel */ }
}
</script>

<style scoped>
.dbuser-page { background: #f5f7fa; }
</style>

