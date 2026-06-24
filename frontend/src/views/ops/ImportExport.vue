﻿<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DB老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <div class="page-container">
    <div class="page-header">
      <div class="page-title">导入导出</div>
      <div class="page-sub">支持按数据源、表维度的数据导入 / 导出任务管理。</div>
    </div>

    <el-card shadow="hover" :body-style="{ padding: '16px' }">
      <el-form :model="form" label-width="100px" inline>
        <el-form-item label="数据源">
          <el-select v-model="form.datasourceId" placeholder="请选择数据源" style="width:220px;">
            <el-option label="生产-订单库 (ds-1)" value="ds-1" />
            <el-option label="测试-商品库 (ds-2)" value="ds-2" />
            <el-option label="本地-SQLite (ds-3)" value="ds-3" />
            <el-option label="归档库 (ds-4)" value="ds-4" />
          </el-select>
        </el-form-item>
        <el-form-item label="模式">
          <el-radio-group v-model="form.mode">
            <el-radio value="export">导出</el-radio>
            <el-radio value="import">导入</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="范围">
          <el-radio-group v-model="form.scope">
            <el-radio value="schema">表结构</el-radio>
            <el-radio value="data">数据</el-radio>
            <el-radio value="all">全部</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="表列表">
          <el-input v-model="tablesText" placeholder="逗号分隔，留空代表全库" style="width:260px;" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="submitting" @click="submitTask">提交任务</el-button>
          <el-button @click="refresh">刷新列表</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="hover" :body-style="{ padding: '16px' }" style="margin-top:16px;">
      <template #header>
        <div style="display:flex;align-items:center;justify-content:space-between;">
          <span>任务列表（对接 /import-export/tasks 接口）</span>
          <ColumnToggle v-model="colVisible" :columns="columns" />
        </div>
      </template>
      <el-table :data="list" border stripe v-loading="loading">
        <el-table-column prop="id" label="任务ID" width="200" show-overflow-tooltip v-if="colVisible.id" />
        <el-table-column prop="type" label="模式" width="100" show-overflow-tooltip v-if="colVisible.type">
          <template #default="{ row }">
            <el-tag :type="row.type === 'export' ? 'success' : 'warning'" size="small">{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="datasourceId" label="数据源ID" width="140" show-overflow-tooltip v-if="colVisible.datasourceId" />
        <el-table-column prop="scope" label="范围" width="120" show-overflow-tooltip v-if="colVisible.scope" />
        <el-table-column prop="tables" label="表列表" show-overflow-tooltip v-if="colVisible.tables" />
        <el-table-column prop="status" label="状态" width="120" show-overflow-tooltip v-if="colVisible.status">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="180" show-overflow-tooltip v-if="colVisible.createTime" />
      </el-table>
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[100, 200, 500, 1000]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        style="margin-top:16px;display:flex;justify-content:flex-end;"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import ColumnToggle from '@/components/ColumnToggle.vue'
import { createImportTask, listImportTasks } from '@/api/ops'

const columns = [
  { key: 'id', label: '任务ID' },
  { key: 'type', label: '模式' },
  { key: 'datasourceId', label: '数据源ID' },
  { key: 'scope', label: '范围' },
  { key: 'tables', label: '表列表' },
  { key: 'status', label: '状态' },
  { key: 'createTime', label: '创建时间' }
]
const colVisible = reactive<Record<string, boolean>>({
  id: true, type: true, datasourceId: true, scope: true, tables: true, status: true, createTime: true
})

const form = reactive({ datasourceId: 'ds-1', mode: 'export', scope: 'data' })
const tablesText = ref('orders, order_items')
const list = ref<any[]>([])
const loading = ref(false)
const submitting = ref(false)
const currentPage = ref(1)
const pageSize = ref(100)
const total = ref(0)

function statusType(s: string) {
  if (s === '已完成' || s === '成功') return 'success'
  if (s === '执行中' || s === '运行中') return 'warning'
  if (s === '失败') return 'danger'
  return 'info'
}

async function refresh() {
  loading.value = true
  try {
    const res: any = await listImportTasks({ page: 1, pageSize: 20 })
    list.value = res?.data?.list || []
    total.value = res?.data?.total || list.value.length
  } catch (e: any) {
    ElMessage.error(e?.message || '获取任务列表失败')
  } finally {
    loading.value = false
  }
}

async function submitTask() {
  if (!form.datasourceId) {
    ElMessage.warning('请先选择数据源')
    return
  }
  submitting.value = true
  try {
    const tables = tablesText.value.split(',').map((t) => t.trim()).filter((t) => !!t)
    const res: any = await createImportTask({
      datasourceId: form.datasourceId, mode: form.mode, scope: form.scope, tables
    })
    if (res?.data) list.value.unshift(res.data)
    ElMessage.success('任务已提交')
  } catch (e: any) {
    ElMessage.error(e?.message || '提交任务失败')
  } finally {
    submitting.value = false
  }
}

onMounted(refresh)
</script>

<style scoped>
.page-container { padding: 16px 20px; }
.page-header { margin-bottom: 12px; }
.page-title { font-size: 20px; font-weight: 600; color: #303133; }
.page-sub { color: #909399; font-size: 13px; margin-top: 4px; }
</style>
