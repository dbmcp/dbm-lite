﻿<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DB老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <div class="page-container">
    <div class="page-header">
      <div class="page-title">SQL 审核</div>
      <div class="page-sub">发起 SQL 变更流程、配置审核规范与自动化检查。</div>
    </div>

    <el-tabs v-model="activeTab" type="border-card">
      <el-tab-pane label="流程发起" name="flow">
        <el-form :model="flowForm" label-width="110px" style="max-width: 860px;">
          <el-form-item label="所属数据源">
            <el-select v-model="flowForm.datasourceId" placeholder="请选择数据源" style="width:260px;">
              <el-option v-for="ds in datasources" :key="ds.id" :label="ds.name" :value="ds.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="变更类型">
            <el-radio-group v-model="flowForm.changeType">
              <el-radio value="DDL">结构变更 (DDL)</el-radio>
              <el-radio value="DML">数据变更 (DML)</el-radio>
              <el-radio value="MIX">混合变更</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="执行环境">
            <el-radio-group v-model="flowForm.env">
              <el-radio value="dev">开发</el-radio>
              <el-radio value="test">测试</el-radio>
              <el-radio value="uat">预发布</el-radio>
              <el-radio value="prod">生产</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="SQL 内容">
            <el-input
              v-model="flowForm.sql"
              type="textarea"
              :rows="10"
              placeholder="请输入需要审核的 SQL 语句，例如：ALTER TABLE orders ADD COLUMN status TINYINT DEFAULT 0;"
            />
          </el-form-item>
          <el-form-item label="风险等级">
            <el-tag :type="riskType">{{ riskText }}</el-tag>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="submitFlow">提交审核</el-button>
            <el-button @click="simulateCheck">模拟语法检查</el-button>
          </el-form-item>
        </el-form>

        <el-divider content-position="left">近期审核工单</el-divider>
        <el-table :data="flowList" border stripe>
          <el-table-column prop="id" label="工单号" width="140" v-if="flowColVisible.id" />
          <el-table-column prop="changeType" label="类型" width="110" v-if="flowColVisible.changeType" />
          <el-table-column prop="datasourceId" label="数据源" width="180" v-if="flowColVisible.datasourceId" />
          <el-table-column prop="env" label="环境" width="80" v-if="flowColVisible.env" />
          <el-table-column prop="sql" label="SQL 摘要" show-overflow-tooltip v-if="flowColVisible.sql" />
          <el-table-column prop="risk" label="风险" width="90" v-if="flowColVisible.risk">
            <template #default="{ row }">
              <el-tag :type="riskTableType(row.risk)" size="small">{{ row.risk }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100" v-if="flowColVisible.status">
            <template #default="{ row }">
              <el-tag size="small">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="createTime" label="创建时间" width="180" v-if="flowColVisible.createTime" />
        </el-table>
        <el-pagination
          v-model:current-page="flowPage"
          v-model:page-size="flowPageSize"
          :page-sizes="[100, 200, 500, 1000]"
          :total="flowTotal"
          layout="total, sizes, prev, pager, next, jumper"
          style="margin-top:16px;display:flex;justify-content:flex-end;"
        />
      </el-tab-pane>

      <el-tab-pane label="审核规范配置" name="rules">
        <el-card shadow="never" :body-style="{ padding: '12px 16px' }">
          <p style="color:#606266;margin:0 0 12px;">
            以下为全局生效的 SQL 审核规则，命中规则将在审核流程中自动标记为警告或失败。
          </p>
        </el-card>
        <div style="display:flex;justify-content:flex-end;margin-bottom:8px;margin-top:12px;">
          <ColumnToggle v-model="ruleColVisible" :columns="ruleColumns" />
        </div>
        <el-table :data="ruleList" border stripe>
          <el-table-column type="index" label="#" width="50" v-if="ruleColVisible.index" />
          <el-table-column prop="name" label="规则名称" width="220" v-if="ruleColVisible.name" />
          <el-table-column prop="scope" label="作用范围" width="130" v-if="ruleColVisible.scope" />
          <el-table-column prop="desc" label="规则描述" v-if="ruleColVisible.desc" />
          <el-table-column prop="level" label="等级" width="90" v-if="ruleColVisible.level">
            <template #default="{ row }">
              <el-tag :type="row.level === 'ERROR' ? 'danger' : 'warning'" size="small">{{ row.level }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="启用状态" width="120" v-if="ruleColVisible.enabled">
            <template #default="{ row }">
              <el-switch v-model="row.enabled" />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120" v-if="ruleColVisible.action">
            <template #default>
              <el-button link type="primary" size="small">编辑</el-button>
            </template>
          </el-table-column>
        </el-table>
        <el-pagination
          v-model:current-page="rulePage"
          v-model:page-size="rulePageSize"
          :page-sizes="[100, 200, 500, 1000]"
          :total="ruleTotal"
          layout="total, sizes, prev, pager, next, jumper"
          style="margin-top:16px;display:flex;justify-content:flex-end;"
        />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, reactive, computed } from 'vue'
import { ElMessage } from 'element-plus'
import ColumnToggle from '@/components/ColumnToggle.vue'
import { listAuditFlows, createAuditFlow, listAuditRules } from '@/api/ops'

const activeTab = ref('flow')

const datasources = ref([
  { id: 'ds-001', name: '生产-订单库' },
  { id: 'ds-002', name: '测试-商品库' },
  { id: 'ds-003', name: '本地-SQLite' }
])

const flowForm = reactive({
  datasourceId: '',
  changeType: 'DDL',
  env: 'test',
  sql: ''
})

const flowColumns = [
  { key: 'id', label: '工单号' },
  { key: 'changeType', label: '类型' },
  { key: 'datasourceId', label: '数据源' },
  { key: 'env', label: '环境' },
  { key: 'sql', label: 'SQL 摘要' },
  { key: 'risk', label: '风险' },
  { key: 'status', label: '状态' },
  { key: 'createTime', label: '创建时间' }
]
const flowColVisible = reactive<Record<string, boolean>>({
  id: true, changeType: true, datasourceId: true, env: true,
  sql: true, risk: true, status: true, createTime: true
})
const flowList = ref<any[]>([])
const flowPage = ref(1)
const flowPageSize = ref(100)
const flowTotal = ref(0)

const ruleColumns = [
  { key: 'index', label: '#' },
  { key: 'name', label: '规则名称' },
  { key: 'scope', label: '作用范围' },
  { key: 'desc', label: '规则描述' },
  { key: 'level', label: '等级' },
  { key: 'enabled', label: '启用状态' },
  { key: 'action', label: '操作' }
]
const ruleColVisible = reactive<Record<string, boolean>>({
  index: true, name: true, scope: true, desc: true, level: true, enabled: true, action: true
})
const ruleList = ref<any[]>([])
const rulePage = ref(1)
const rulePageSize = ref(100)
const ruleTotal = ref(0)

const riskText = computed(() => {
  const sql = flowForm.sql.trim()
  if (!sql) return '未判定'
  if (/DROP\s+TABLE|TRUNCATE/i.test(sql)) return '高危'
  if (/ALTER\s+TABLE|UPDATE|DELETE/i.test(sql)) return '中等'
  if (/CREATE\s+TABLE|INSERT/i.test(sql)) return '低危'
  return '低危'
})
const riskType = computed(() => {
  if (riskText.value === '高危') return 'danger'
  if (riskText.value === '中等') return 'warning'
  return 'success'
})

function riskTableType(r: string) {
  if (r === '高危') return 'danger'
  if (r === '中等') return 'warning'
  return 'success'
}

function simulateCheck() {
  if (!flowForm.sql.trim()) {
    ElMessage.warning('请先输入 SQL')
    return
  }
  ElMessage.success('语法检查通过，风险等级：' + riskText.value)
}

async function loadFlows() {
  try {
    const res: any = await listAuditFlows()
    if (res && res.success) {
      flowList.value = res.data?.list || res.data || []
      flowTotal.value = res.data?.total || flowList.value.length
    } else {
      ElMessage.error(res?.message || '获取工单列表失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

async function loadRules() {
  try {
    const res: any = await listAuditRules()
    if (res && res.success) {
      ruleList.value = res.data?.list || res.data || []
      ruleTotal.value = res.data?.total || ruleList.value.length
    } else {
      ElMessage.error(res?.message || '获取规则列表失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

async function submitFlow() {
  if (!flowForm.datasourceId || !flowForm.sql.trim()) {
    ElMessage.warning('请完善表单')
    return
  }
  try {
    const res: any = await createAuditFlow({
      datasourceId: flowForm.datasourceId,
      changeType: flowForm.changeType,
      env: flowForm.env,
      sql: flowForm.sql
    })
    if (res && res.success) {
      if (res.data) {
        flowList.value.unshift(res.data)
      }
      ElMessage.success('工单已提交，等待审核，风险等级：' + (res.data?.risk || riskText.value))
    } else {
      ElMessage.error(res?.message || '提交失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

onMounted(() => {
  loadFlows()
  loadRules()
})
</script>

<style scoped>
.page-container { padding: 16px 20px; }
.page-header { margin-bottom: 12px; }
.page-title { font-size: 20px; font-weight: 600; color: #303133; }
.page-sub { color: #909399; font-size: 13px; margin-top: 4px; }
</style>
