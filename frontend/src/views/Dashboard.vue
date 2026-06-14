<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DBA老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <div class="page-container">
    <div style="display:grid;grid-template-columns:repeat(4,1fr);gap:16px;margin-bottom:20px;">
      <div class="stat-card" v-for="(stat, idx) in stats" :key="idx">
        <div class="stat-label">{{ stat.label }}</div>
        <div class="stat-number">{{ stat.value }}</div>
      </div>
    </div>

    <div class="card" style="margin-bottom:20px;">
      <div class="page-header" style="margin-bottom:12px;">快速开始</div>
      <div style="display:grid;grid-template-columns:repeat(3,1fr);gap:16px;">
        <el-card shadow="hover" @click="router.push('/datasources')" style="cursor:pointer;">
          <div style="text-align:center;padding:20px;">
            <el-icon :size="40" color="#409eff"><Coin /></el-icon>
            <div style="margin-top:12px;font-weight:600;">管理数据源</div>
            <div style="color:#909399;font-size:13px;margin-top:4px;">添加、编辑、测试数据库连接</div>
          </div>
        </el-card>
        <el-card shadow="hover" @click="router.push('/sql/workbench')" style="cursor:pointer;">
          <div style="text-align:center;padding:20px;">
            <el-icon :size="40" color="#67c23a"><EditPen /></el-icon>
            <div style="margin-top:12px;font-weight:600;">SQL 工作台</div>
            <div style="color:#909399;font-size:13px;margin-top:4px;">编写和执行 SQL 查询</div>
          </div>
        </el-card>
        <el-card v-if="userStore.isAdmin" shadow="hover" @click="router.push('/accounts')" style="cursor:pointer;">
          <div style="text-align:center;padding:20px;">
            <el-icon :size="40" color="#e6a23c"><User /></el-icon>
            <div style="margin-top:12px;font-weight:600;">管理账号</div>
            <div style="color:#909399;font-size:13px;margin-top:4px;">创建和管理平台用户</div>
          </div>
        </el-card>
      </div>
    </div>

    <div class="card">
      <div class="page-header" style="margin-bottom:12px;">最近执行记录</div>
      <el-table :data="recentHistory" style="width:100%;" max-height="300" empty-text="暂无数据">
        <el-table-column prop="datasourceName" label="数据源" width="180" />
        <el-table-column prop="databaseName" label="数据库" width="140" />
        <el-table-column prop="sqlText" label="SQL" show-overflow-tooltip />
        <el-table-column prop="executeMs" label="耗时(ms)" width="100" />
        <el-table-column prop="username" label="用户" width="120" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.status === 'success'" type="success">成功</el-tag>
            <el-tag v-else type="danger">失败</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="时间" width="180" />
      </el-table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { getAuditStats, listSqlHistory } from '@/api/sql'

const router = useRouter()
const userStore = useUserStore()
const stats = ref([
  { label: '总用户数', value: 0 },
  { label: '总数据源数', value: 0 },
  { label: '总 SQL 执行数', value: 0 },
  { label: '今日 SQL 执行', value: 0 }
])
const recentHistory = ref<any[]>([])

async function loadData() {
  try {
    const s: any = await getAuditStats()
    stats.value = [
      { label: '总用户数', value: s.totalUsers || 0 },
      { label: '总数据源数', value: s.totalDatasources || 0 },
      { label: '总 SQL 执行数', value: s.totalSqlExec || 0 },
      { label: '今日 SQL 执行', value: s.todaySqlExec || 0 }
    ]
  } catch (e) {}
  try {
    const r: any = await listSqlHistory(1, 10)
    recentHistory.value = r.list || []
  } catch (e) {}
}

onMounted(loadData)
</script>

