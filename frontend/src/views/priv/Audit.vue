<template>
  <div class="priv-page">
    <div class="page-header">
      <h2 class="title">操作审计</h2>
      <div class="actions">
        <el-input v-model="keyword" placeholder="搜索操作人 / 对象 / 详情" clearable style="width: 260px" />
      </div>
    </div>

    <el-table :data="logs" border stripe v-loading="loading">
      <el-table-column label="操作时间" prop="createdAt" width="180" show-overflow-tooltip />
      <el-table-column label="操作人" prop="operator" min-width="120" show-overflow-tooltip />
      <el-table-column label="操作类型" prop="operType" width="160" show-overflow-tooltip>
        <template #default="scope">
          <el-tag size="small" :type="tagType(scope.row.operType)">{{ scope.row.operType }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="对象" prop="targetId" min-width="220" show-overflow-tooltip />
      <el-table-column label="详情" prop="detail" min-width="260" show-overflow-tooltip />
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
import { listPrivAuditLogs } from '@/api/priv'

const loading = ref(false)
const logs = ref<any[]>([])
const total = ref(0)
const current = ref(1)
const pageSize = ref(100)
const keyword = ref('')

function tagType(t: string) {
  if (!t) return 'info'
  if (t.includes('grant')) return 'success'
  if (t.includes('revoke')) return 'danger'
  if (t.includes('approve')) return 'primary'
  return 'warning'
}

async function load() {
  try {
    loading.value = true
    const res: any = await listPrivAuditLogs(current.value, pageSize.value, { keyword: keyword.value })
    logs.value = res?.list || []
    total.value = res?.total || 0
  } catch (e) {
    logs.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

watch(keyword, () => {
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
.pagination-wrap {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
  background: #fff;
  padding: 12px 20px;
  border-radius: 8px;
}
</style>
