<template>
  <div class="grant-page">
    <div class="page-header">
      <h2 class="title">权限分配</h2>
      <div class="actions">
        <el-button type="primary" @click="submitAll">
          <el-icon><Check /></el-icon>
          保存授权
        </el-button>
        <el-button type="warning" @click="resetAll">
          <el-icon><RefreshLeft /></el-icon>
          重置
        </el-button>
        <el-button type="danger" @click="openBatchRevoke">
          <el-icon><Delete /></el-icon>
          批量回收
        </el-button>
      </div>
    </div>

    <div class="card">
      <el-form :model="filter" label-width="96px">
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item label="目标用户">
              <el-select
                v-model="filter.userId"
                placeholder="选择用户"
                filterable
                multiple
                collapse-tags
                collapse-tags-tooltip
                style="width: 100%"
              >
                <el-option v-for="u in users" :key="u.userId" :label="u.displayName || u.username" :value="u.userId" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="数据源">
              <el-select v-model="filter.datasourceId" placeholder="选择数据源" style="width: 100%" @change="onDsChange">
                <el-option v-for="d in datasources" :key="d.datasourceId" :label="d.name" :value="d.datasourceId" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="数据库">
              <el-select v-model="filter.databaseName" placeholder="选择数据库" style="width: 100%" @change="onDbChange">
                <el-option v-for="d in databases" :key="d" :label="d" :value="d" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item label="数据表">
              <el-select
                v-model="filter.tableNames"
                placeholder="选择表"
                multiple
                filterable
                collapse-tags
                collapse-tags-tooltip
                style="width: 100%"
                @change="onTableChange"
              >
                <el-option v-for="t in tables" :key="t" :label="t" :value="t" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="授权列">
              <el-select
                v-model="filter.columns"
                placeholder="选择列（留空表示全部）"
                multiple
                filterable
                collapse-tags
                collapse-tags-tooltip
                style="width: 100%"
              >
                <el-option v-for="c in columns" :key="c" :label="c" :value="c" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="权限类型">
              <el-select v-model="filter.privType" style="width: 100%">
                <el-option label="只读" value="query" />
                <el-option label="DML（写操作）" value="dml" />
                <el-option label="DDL（建表/删表）" value="ddl" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item label="有效期">
              <el-select v-model="filter.validDays" style="width: 100%">
                <el-option label="永久" :value="-1" />
                <el-option label="7 天" :value="7" />
                <el-option label="30 天" :value="30" />
                <el-option label="90 天" :value="90" />
                <el-option label="180 天" :value="180" />
                <el-option label="365 天" :value="365" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="查询行数上限">
              <el-input-number v-model="filter.rowLimit" :min="0" :max="100000" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="说明">
              <el-input v-model="filter.remark" placeholder="选填，便于追溯" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
    </div>

    <div class="card">
      <h3 class="sub-title">已授权列表</h3>
      <el-table :data="privs" v-loading="privLoading" border stripe @selection-change="onSel">
        <el-table-column type="selection" width="48" show-overflow-tooltip />
        <el-table-column label="用户" prop="userName" min-width="120" show-overflow-tooltip />
        <el-table-column label="数据源" prop="datasourceId" min-width="120" show-overflow-tooltip />
        <el-table-column label="数据库" prop="databaseName" min-width="120" show-overflow-tooltip />
        <el-table-column label="数据表" prop="tableName" min-width="120" show-overflow-tooltip />
        <el-table-column label="类型" prop="privType" width="100" show-overflow-tooltip>
          <template #default="scope">
            <el-tag size="small" :type="scope.row.privType === 'query' ? 'success' : scope.row.privType === 'dml' ? 'warning' : 'danger'">{{ scope.row.privType }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="有效期" prop="expireAt" width="180" show-overflow-tooltip>
          <template #default="scope">
            <span :class="['expire', { expired: scope.row.isExpired }]">{{ scope.row.isExpired ? '已过期' : (scope.row.expireAt || '永久') }}</span>
          </template>
        </el-table-column>
        <el-table-column label="行数上限" prop="rowLimit" width="100" show-overflow-tooltip />
        <el-table-column label="操作" width="100" show-overflow-tooltip>
          <template #default="scope">
            <el-button size="small" link type="danger" @click="revokeOne(scope.row)">回收</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="privPage"
          v-model:page-size="privPageSize"
          :total="privTotal"
          layout="total, sizes, prev, pager, next"
          :page-sizes="[100, 200, 500, 1000]"
          background
          @current-change="loadPrivs"
          @size-change="loadPrivs"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Check, RefreshLeft, Delete } from '@element-plus/icons-vue'
import { grantPrivilegeBatch, revokePrivilege, revokePrivilegeBatch, listPrivileges } from '@/api/priv'
import { listAllDatasources } from '@/api/datasource'
import { listAccounts } from '@/api/auth'

const users = ref<any[]>([])
const datasources = ref<any[]>([])
const databases = ref<string[]>([])
const tables = ref<string[]>([])
const columns = ref<string[]>([])

const filter = reactive<{
  userId: string[]
  datasourceId: string
  databaseName: string
  tableNames: string[]
  columns: string[]
  privType: string
  validDays: number
  rowLimit: number
  remark: string
}>({
  userId: [],
  datasourceId: '',
  databaseName: '',
  tableNames: [],
  columns: [],
  privType: 'query',
  validDays: 30,
  rowLimit: 1000,
  remark: ''
})

const privs = ref<any[]>([])
const privLoading = ref(false)
const privTotal = ref(0)
const privPage = ref(1)
const privPageSize = ref(100)
const selectedIds = ref<string[]>([])

async function loadUsers() {
  try {
    const res: any = await listAccounts(1, 500)
    if (res?.list) users.value = res.list
  } catch (e) {}
}

async function loadDS() {
  try {
    const res: any = await listAllDatasources(1, 1000)
    datasources.value = res?.list || []
  } catch (e) {}
}

async function onDsChange() {
  databases.value = []
  tables.value = []
  columns.value = []
  if (!filter.datasourceId) return
  try {
    const { getDatabases } = await import('@/api/sql')
    databases.value = (await getDatabases(filter.datasourceId)) || []
  } catch (e) {}
}

async function onDbChange() {
  tables.value = []
  columns.value = []
  if (!filter.datasourceId || !filter.databaseName) return
  try {
    const { listTables } = await import('@/api/sql')
    const raw = await listTables(filter.datasourceId, filter.databaseName)
    tables.value = Array.isArray(raw) ? raw : (raw as any)?.list || []
  } catch (e) {}
}

async function onTableChange() {
  columns.value = []
  if (!filter.datasourceId || !filter.databaseName || !filter.tableNames?.length) return
  try {
    const { listColumns } = await import('@/api/sql')
    const raw = await listColumns(filter.datasourceId, filter.databaseName, filter.tableNames[0])
    const arr = Array.isArray(raw) ? raw : (raw as any)?.list || []
    columns.value = arr.map((x: any) => x.column || x.name || x)
  } catch (e) {}
}

function resetAll() {
  filter.userId = []
  filter.datasourceId = ''
  filter.databaseName = ''
  filter.tableNames = []
  filter.columns = []
  filter.privType = 'query'
  filter.validDays = 30
  filter.rowLimit = 1000
  filter.remark = ''
}

async function submitAll() {
  if (!filter.userId.length) {
    ElMessage.warning('请至少选择一个目标用户')
    return
  }
  if (!filter.datasourceId) {
    ElMessage.warning('请选择数据源')
    return
  }
  try {
    const userIds = filter.userId
    const targetTables = filter.tableNames || []
    const common = {
      datasourceId: filter.datasourceId,
      databaseName: filter.databaseName || undefined,
      privType: filter.privType,
      columns: filter.columns && filter.columns.length ? filter.columns : undefined,
      rowLimit: filter.rowLimit,
      validDays: filter.validDays
    }
    const requests: Promise<any>[] = []
    if (!targetTables.length) {
      requests.push(grantPrivilegeBatch({ ...common, userIds, tableName: undefined }))
    } else {
      for (const t of targetTables) {
        requests.push(grantPrivilegeBatch({ ...common, tableName: t, userIds }))
      }
    }
    await Promise.all(requests)
    ElMessage.success('授权成功')
    loadPrivs()
  } catch (e: any) {
    ElMessage.error(e?.message || '授权失败')
  }
}

async function loadPrivs() {
  try {
    privLoading.value = true
    const res: any = await listPrivileges(privPage.value, privPageSize.value, {})
    privs.value = res?.list || []
    privTotal.value = res?.total || 0
  } catch (e) {
    privs.value = []
    privTotal.value = 0
  } finally {
    privLoading.value = false
  }
}

function onSel(rows: any[]) {
  selectedIds.value = rows.map((r) => r.privId)
}

async function revokeOne(row: any) {
  try {
    await ElMessageBox.confirm('确认回收该权限？', '提示', { type: 'warning' })
    await revokePrivilege(row.privId)
    ElMessage.success('回收成功')
    loadPrivs()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('回收失败')
  }
}

async function openBatchRevoke() {
  if (!selectedIds.value.length) {
    ElMessage.warning('请先在列表中勾选要回收的权限')
    return
  }
  try {
    await ElMessageBox.confirm('确认批量回收选中的 ' + selectedIds.value.length + ' 条权限？', '提示', { type: 'warning' })
    await revokePrivilegeBatch(selectedIds.value)
    ElMessage.success('批量回收成功')
    loadPrivs()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('批量回收失败')
  }
}

onMounted(() => {
  loadUsers()
  loadDS()
  loadPrivs()
})
</script>

<style scoped>
.grant-page {
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
  gap: 8px;
}
.card {
  background: #fff;
  padding: 16px 20px;
  border-radius: 8px;
  margin-bottom: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}
.sub-title {
  margin: 0 0 12px;
  font-size: 14px;
  color: #606266;
}
.pagination-wrap {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}
.expire.expired {
  color: #f56c6c;
}
</style>
