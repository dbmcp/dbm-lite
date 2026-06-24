﻿﻿<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DB老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <div class="page-container">
    <div class="page-header">
      <div class="page-title">健康巡检</div>
      <div class="page-sub">实时运行监控与定期健康巡检结果分析。</div>
    </div>

    <el-tabs v-model="activeTab" type="border-card">
      <el-tab-pane label="监控" name="monitor">
        <el-row :gutter="16">
          <el-col v-for="card in metrics" :key="card.title" :span="6">
            <el-card shadow="hover" :body-style="{ padding: '16px' }">
              <div class="metric-title">{{ card.title }}</div>
              <div class="metric-value">{{ card.value }}</div>
              <div class="metric-sub">{{ card.sub }}</div>
              <el-progress :percentage="card.percent" :color="card.color" :stroke-width="8" style="margin-top:8px;" />
            </el-card>
          </el-col>
        </el-row>

        <el-card shadow="hover" :body-style="{ padding: '16px' }" style="margin-top:16px;">
          <template #header>
            <div style="display:flex;align-items:center;justify-content:space-between;">
              <span>实例指标</span>
              <div style="display:flex;align-items:center;gap:10px;">
                <el-tag type="success" size="small">已连接</el-tag>
                <ColumnToggle v-model="instanceColVisible" :columns="instanceColumns" />
              </div>
            </div>
          </template>
          <el-table :data="instances" border stripe>
            <el-table-column prop="name" label="实例" width="200" v-if="instanceColVisible.name" />
            <el-table-column prop="qps" label="QPS" width="110" v-if="instanceColVisible.qps" />
            <el-table-column prop="connections" label="连接数" width="110" v-if="instanceColVisible.connections" />
            <el-table-column prop="bufferHit" label="缓冲命中率" width="140" v-if="instanceColVisible.bufferHit">
              <template #default="{ row }">
                <el-progress :percentage="Number(row.bufferHit)" :stroke-width="10" />
              </template>
            </el-table-column>
            <el-table-column prop="replicationLag" label="主从延迟(秒)" width="130" v-if="instanceColVisible.replicationLag" />
            <el-table-column prop="slowQueries" label="慢查询(1h)" width="130" v-if="instanceColVisible.slowQueries" />
            <el-table-column prop="diskUsage" label="磁盘使用率" width="160" v-if="instanceColVisible.diskUsage">
              <template #default="{ row }">
                <el-progress :percentage="Number(row.diskUsage)" :stroke-width="10" :color="row.diskUsage > 75 ? '#f56c6c' : '#67c23a'" />
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100" v-if="instanceColVisible.status">
              <template #default="{ row }">
                <el-tag :type="row.status === '正常' ? 'success' : 'warning'" size="small">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
          <el-pagination
            v-model:current-page="instancePage"
            v-model:page-size="instancePageSize"
            :page-sizes="[100, 200, 500, 1000]"
            :total="instanceTotal"
            layout="total, sizes, prev, pager, next, jumper"
            style="margin-top:16px;display:flex;justify-content:flex-end;"
          />
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="巡检" name="inspect">
        <el-card shadow="hover" :body-style="{ padding: '16px' }">
          <div style="display:flex;align-items:center;gap:10px;margin-bottom:12px;">
            <el-select v-model="inspectEnv" style="width:160px;">
              <el-option label="生产环境" value="prod" />
              <el-option label="预发布" value="uat" />
              <el-option label="测试环境" value="test" />
            </el-select>
            <el-select v-model="inspectTarget" multiple collapse-tags placeholder="选择巡检项目" style="width:280px;">
              <el-option v-for="it in inspectItems" :key="it" :label="it" :value="it" />
            </el-select>
            <el-button type="primary" @click="runInspect">一键巡检</el-button>
            <el-button @click="loadInspect">刷新</el-button>
            <div style="margin-left:auto;">
              <ColumnToggle v-model="inspectColVisible" :columns="inspectColumns" />
            </div>
          </div>

          <el-table :data="inspectResult" border stripe>
            <el-table-column prop="time" label="巡检时间" width="180" v-if="inspectColVisible.time" />
            <el-table-column prop="env" label="环境" width="100" v-if="inspectColVisible.env" />
            <el-table-column prop="instance" label="对象" width="220" v-if="inspectColVisible.instance" />
            <el-table-column prop="item" label="巡检项" width="180" v-if="inspectColVisible.item" />
            <el-table-column prop="level" label="等级" width="100" v-if="inspectColVisible.level">
              <template #default="{ row }">
                <el-tag :type="levelTag(row.level)" size="small">{{ row.level }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="detail" label="详情" show-overflow-tooltip v-if="inspectColVisible.detail" />
          </el-table>
          <el-pagination
            v-model:current-page="inspectPage"
            v-model:page-size="inspectPageSize"
            :page-sizes="[100, 200, 500, 1000]"
            :total="inspectTotal"
            layout="total, sizes, prev, pager, next, jumper"
            style="margin-top:16px;display:flex;justify-content:flex-end;"
          />
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import ColumnToggle from '@/components/ColumnToggle.vue'
import { getHealthMetrics, getHealthInstances, getHealthInspectResults, triggerHealthInspect } from '@/api/ops'

const activeTab = ref('monitor')
const inspectEnv = ref('prod')
const inspectTarget = ref<string[]>(['参数合规', '备份状态', '连接状态', '主从延迟', '磁盘空间'])
const inspectItems = ['参数合规', '备份状态', '连接状态', '主从延迟', '磁盘空间', '慢查询', '表碎片']

const metrics = ref<any[]>([])

const instanceColumns = [
  { key: 'name', label: '实例' },
  { key: 'qps', label: 'QPS' },
  { key: 'connections', label: '连接数' },
  { key: 'bufferHit', label: '缓冲命中率' },
  { key: 'replicationLag', label: '主从延迟(秒)' },
  { key: 'slowQueries', label: '慢查询(1h)' },
  { key: 'diskUsage', label: '磁盘使用率' },
  { key: 'status', label: '状态' }
]
let instanceColVisible = reactive<Record<string, boolean>>({
  name: true, qps: true, connections: true, bufferHit: true, replicationLag: true,
  slowQueries: true, diskUsage: true, status: true
})
const instances = ref<any[]>([])
const instancePage = ref(1)
const instancePageSize = ref(100)
const instanceTotal = ref(0)

const inspectColumns = [
  { key: 'time', label: '巡检时间' },
  { key: 'env', label: '环境' },
  { key: 'instance', label: '对象' },
  { key: 'item', label: '巡检项' },
  { key: 'level', label: '等级' },
  { key: 'detail', label: '详情' }
]
let inspectColVisible = reactive<Record<string, boolean>>({
  time: true, env: true, instance: true, item: true, level: true, detail: true
})
const inspectResult = ref<any[]>([])
const inspectPage = ref(1)
const inspectPageSize = ref(100)
const inspectTotal = ref(0)

function levelTag(level: string) {
  if (level === 'ERROR' || level === '错误') return 'danger'
  if (level === 'WARN' || level === '警告') return 'warning'
  return 'success'
}

async function loadMetrics() {
  try {
    const res: any = await getHealthMetrics()
    if (res && res.success) {
      metrics.value = res.data?.list || res.data || []
    } else {
      ElMessage.error(res?.message || '获取指标失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

async function loadInstances() {
  try {
    const res: any = await getHealthInstances()
    if (res && res.success) {
      instances.value = res.data?.list || res.data || []
      instanceTotal.value = res.data?.total || instances.value.length
    } else {
      ElMessage.error(res?.message || '获取实例指标失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

async function loadInspect() {
  try {
    const res: any = await getHealthInspectResults()
    if (res && res.success) {
      inspectResult.value = res.data?.list || res.data || []
      inspectTotal.value = res.data?.total || inspectResult.value.length
    } else {
      ElMessage.error(res?.message || '获取巡检结果失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

async function runInspect() {
  try {
    const res: any = await triggerHealthInspect({
      env: inspectEnv.value,
      items: inspectTarget.value,
      target: inspectTarget.value.join(',')
    })
    if (res && res.success) {
      ElMessage.success('巡检已触发')
      loadInspect()
    } else {
      ElMessage.error(res?.message || '触发巡检失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

onMounted(() => {
  loadMetrics()
  loadInstances()
  loadInspect()
})
</script>

<style scoped>
.page-container { padding: 16px 20px; }
.page-header { margin-bottom: 12px; }
.page-title { font-size: 20px; font-weight: 600; color: #303133; }
.page-sub { color: #909399; font-size: 13px; margin-top: 4px; }
.metric-title { font-size: 14px; color: #606266; }
.metric-value { font-size: 24px; font-weight: 700; color: #303133; margin-top: 4px; }
.metric-sub { font-size: 12px; color: #909399; margin-top: 2px; }
</style>
