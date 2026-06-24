<template>
  <div class="priv-page">
    <div class="page-header">
      <h2 class="title">权限清单</h2>
      <div class="actions">
        <el-input v-model="keyword" placeholder="搜索数据源 / 库 / 表 / 用户" clearable style="width: 260px; margin-right: 12px" />
        <el-select v-model="privTypeFilter" placeholder="类型" style="width: 140px; margin-right: 12px" clearable>
          <el-option label="只读" value="query" />
          <el-option label="DML" value="dml" />
          <el-option label="DDL" value="ddl" />
        </el-select>
        <el-select v-model="expireFilter" placeholder="有效期状态" style="width: 140px" clearable>
          <el-option label="正常" value="active" />
          <el-option label="已过期" value="expired" />
        </el-select>
      </div>
    </div>

    <el-table :data="privs" border stripe v-loading="loading">
      <el-table-column label="用户" prop="userName" min-width="120" show-overflow-tooltip />
      <el-table-column label="数据源" prop="datasourceId" min-width="140" show-overflow-tooltip />
      <el-table-column label="数据库" prop="databaseName" min-width="120" show-overflow-tooltip />
      <el-table-column label="数据表" prop="tableName" min-width="120" show-overflow-tooltip />
      <el-table-column label="列" prop="columns" min-width="140" show-overflow-tooltip />
      <el-table-column label="类型" prop="privType" width="100" show-overflow-tooltip>
        <template #default="scope">
          <el-tag size="small" :type="scope.row.privType === 'query' ? 'success' : scope.row.privType === 'dml' ? 'warning' : 'danger'">{{ scope.row.privType }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="行数上限" prop="rowLimit" width="100" show-overflow-tooltip />
      <el-table-column label="有效期" width="180" show-overflow-tooltip>
        <template #default="scope">
          <el-tag size="small" :type="scope.row.isExpired ? 'info' : 'success'">
            {{ scope.row.isExpired ? '已过期' : (scope.row.expireAt || '永久') }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" prop="createdAt" width="180" show-overflow-tooltip />
      <el-table-column label="操作" width="100" show-overflow-tooltip>
        <template #default="scope">
          <el-button size="small" link type="danger" @click="revokeOne(scope.row)">回收</el-button>
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listPrivileges, revokePrivilege } from '@/api/priv'

const loading = ref(false)
const privs = ref<any[]>([])
const total = ref(0)
const current = ref(1)
const pageSize = ref(100)
const keyword = ref('')
const privTypeFilter = ref('')
const expireFilter = ref('')

async function load() {
  try {
    loading.value = true
    const res: any = await listPrivileges(current.value, pageSize.value, {
      keyword: keyword.value,
      privType: privTypeFilter.value,
      status: expireFilter.value
    })
    privs.value = res?.list || []
    total.value = res?.total || 0
  } catch (e) {
    privs.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

async function revokeOne(row: any) {
  try {
    await ElMessageBox.confirm('确认回收该权限？', '提示', { type: 'warning' })
    await revokePrivilege(row.privId)
    ElMessage.success('回收成功')
    load()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('回收失败')
  }
}

watch([keyword, privTypeFilter, expireFilter], () => {
  current.value = 1
  load()
})
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
