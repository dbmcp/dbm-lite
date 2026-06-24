﻿<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DB老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <div class="page-container">
    <div class="page-header">
      <div class="page-title">DB 生命周期管理</div>
      <div class="page-sub">
        覆盖数据库从创建到销毁的全流程环节：参数维护与高可用管理为生命周期中的核心运维节点。
      </div>
    </div>

    <el-card shadow="hover" :body-style="{ padding: '16px' }">
      <template #header>生命周期节点概览</template>
      <div class="lifecycle-timeline">
        <div v-for="(node, idx) in nodes" :key="node.key" class="timeline-item" :class="{ active: node.key === currentNode }" @click="currentNode = node.key">
          <div class="timeline-circle" :style="{ background: node.color }">{{ idx + 1 }}</div>
          <div class="timeline-title">{{ node.title }}</div>
          <div class="timeline-sub">{{ node.status }}</div>
          <div v-if="idx < nodes.length - 1" class="timeline-line"></div>
        </div>
      </div>
    </el-card>

    <el-row :gutter="16" style="margin-top:16px;">
      <el-col v-for="node in nodes" :key="node.key" :span="8">
        <el-card shadow="hover" class="lifecycle-card" :body-style="{ padding: '16px' }">
          <div class="card-header">
            <el-icon :size="22" :color="node.color"><component :is="node.icon" /></el-icon>
            <div class="card-title">{{ node.title }}</div>
          </div>
          <div class="card-desc">{{ node.desc }}</div>
          <el-divider />
          <div class="feature-list">
            <div v-for="(f, i) in node.features" :key="i" class="feature-item">
              <el-icon><Check /></el-icon>
              <span>{{ f }}</span>
              <el-tag size="small" type="info" round>Demo</el-tag>
            </div>
          </div>
          <el-button disabled type="primary" style="width:100%;margin-top:12px;">进入（功能开发中）</el-button>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="hover" :body-style="{ padding: '16px' }" style="margin-top:16px;">
      <template #header>
        <div style="display:flex;align-items:center;justify-content:space-between;">
          <span>示例数据库列表</span>
          <ColumnToggle v-model="colVisible" :columns="columns" />
        </div>
      </template>
      <el-table :data="dbList" border stripe>
        <el-table-column prop="name" label="数据库名称" width="180" v-if="colVisible.name" />
        <el-table-column prop="dbType" label="数据库类型" width="130" v-if="colVisible.dbType" />
        <el-table-column prop="version" label="版本" width="120" v-if="colVisible.version" />
        <el-table-column prop="env" label="环境" width="90" v-if="colVisible.env">
          <template #default="{ row }">
            <el-tag :type="envTag(row.env)" size="small">{{ row.env }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="currentPhase" label="当前阶段" width="140" v-if="colVisible.currentPhase" />
        <el-table-column prop="haMode" label="高可用模式" width="140" v-if="colVisible.haMode" />
        <el-table-column prop="params" label="关键参数" show-overflow-tooltip v-if="colVisible.params" />
        <el-table-column prop="status" label="运行状态" width="100" v-if="colVisible.status">
          <template #default="{ row }">
            <el-tag :type="row.status === '运行中' ? 'success' : 'warning'" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="180" v-if="colVisible.createTime" />
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
import { onMounted, ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import ColumnToggle from '@/components/ColumnToggle.vue'
import { Check } from '@element-plus/icons-vue'
import {
  Coin, MagicStick, DataLine, Histogram, Lock, Refresh, Tools, Setting, Delete, Upload
} from '@element-plus/icons-vue'
import { listLifecycleNodes, listLifecycleDBs } from '@/api/ops'

const iconMap: Record<string, any> = {
  Coin, MagicStick, DataLine, Histogram, Lock, Refresh, Tools, Setting, Delete, Upload
}

const defaultNodes = [
  { key: 'create', title: '数据库创建与初始化', status: '活跃', color: '#409eff', icon: 'Coin',
    desc: '新建数据库实例、表结构初始化、元数据注册。',
    features: ['选择数据库类型 / 环境', '配置编码、字符集、连接数', '执行初始化脚本'] },
  { key: 'upgrade', title: '版本管理与升级', status: '近期操作', color: '#67c23a', icon: 'MagicStick',
    desc: '记录数据库版本、Patch 历史、执行小版本升级或大版本迁移。',
    features: ['版本历史记录', 'Patch 发布与回滚', '大版本升级向导'] },
  { key: 'backup', title: '备份与恢复', status: '每日运行', color: '#e6a23c', icon: 'DataLine',
    desc: '全量 / 增量备份、备份集校验与一键恢复演练。',
    features: ['定时全备 / 增备', '备份集自动校验', '一键恢复演练'] },
  { key: 'monitor', title: '运行监控', status: '实时', color: '#2e6ba8', icon: 'Histogram',
    desc: '连接数、QPS、慢查询、主从延迟等核心指标监控与告警。',
    features: ['实时指标采集', '告警阈值配置', '慢 SQL 分析'] },
  { key: 'capacity', title: '容量与配额', status: '本月达标', color: '#8e44ad', icon: 'Lock',
    desc: '容量趋势预测、空间配额管理、存储回收建议。',
    features: ['容量趋势预测', '空间配额分配', '存储回收建议'] },
  { key: 'params', title: '参数维护', status: '待优化 3 项', color: '#f56c6c', icon: 'Tools',
    desc: '集中维护核心参数（innodb_buffer_pool_size、max_connections 等），支持按版本 / 环境比对。',
    features: ['参数集中管理', '参数变更自动巡检', '参数历史与回滚'] },
  { key: 'ha', title: '高可用管理', status: 'M-S 架构', color: '#13c2c2', icon: 'Refresh',
    desc: '主从 / MGR / 集群状态巡检，自动切换演练与切换历史记录。',
    features: ['主从状态巡检', '自动切换演练', '切换历史记录'] },
  { key: 'offline', title: '下线与销毁', status: '流程审核中', color: '#909399', icon: 'Delete',
    desc: '归档、只读、下线、销毁全流程审批与数据清理。',
    features: ['归档 / 只读切换', '审批流程', '安全销毁'] }
]

const currentNode = ref('create')
const nodes = ref<any[]>(defaultNodes.map((n) => ({ ...n, icon: iconMap[n.icon] })))
const dbList = ref<any[]>([])

const columns = [
  { key: 'name', label: '数据库名称' },
  { key: 'dbType', label: '数据库类型' },
  { key: 'version', label: '版本' },
  { key: 'env', label: '环境' },
  { key: 'currentPhase', label: '当前阶段' },
  { key: 'haMode', label: '高可用模式' },
  { key: 'params', label: '关键参数' },
  { key: 'status', label: '运行状态' },
  { key: 'createTime', label: '创建时间' }
]
const colVisible = reactive<Record<string, boolean>>({
  name: true, dbType: true, version: true, env: true, currentPhase: true,
  haMode: true, params: true, status: true, createTime: true
})
const currentPage = ref(1)
const pageSize = ref(100)
const total = ref(0)

function envTag(env: string) {
  if (env === '生产' || env === 'prod') return 'danger'
  if (env === '预发布' || env === 'uat') return 'warning'
  if (env === '测试' || env === 'test') return 'info'
  return 'success'
}

async function loadNodes() {
  try {
    const res: any = await listLifecycleNodes()
    if (res && res.success) {
      const apiNodes = res.data?.list || res.data || []
      if (apiNodes && apiNodes.length > 0) {
        nodes.value = apiNodes.map((n: any) => {
          const iconName = n.icon || 'Coin'
          return { ...n, icon: iconMap[iconName] || Coin, features: n.features || [] }
        })
        if (nodes.value.length > 0) {
          currentNode.value = nodes.value[0].key
        }
      }
    } else {
      ElMessage.error(res?.message || '获取生命周期节点失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

async function loadDBs() {
  try {
    const res: any = await listLifecycleDBs()
    if (res && res.success) {
      dbList.value = res.data?.list || res.data || []
      total.value = res.data?.total || dbList.value.length
    } else {
      ElMessage.error(res?.message || '获取数据库列表失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

onMounted(() => {
  loadNodes()
  loadDBs()
})
</script>

<style scoped>
.page-container { padding: 16px 20px; }
.page-header { margin-bottom: 12px; }
.page-title { font-size: 20px; font-weight: 600; color: #303133; }
.page-sub { color: #909399; font-size: 13px; margin-top: 4px; }
.lifecycle-timeline {
  display: flex;
  align-items: flex-start;
  overflow-x: auto;
  padding: 10px 0;
}
.timeline-item {
  position: relative;
  min-width: 130px;
  text-align: center;
  cursor: pointer;
  padding: 4px 10px;
}
.timeline-circle {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  color: #fff;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
}
.timeline-title { margin-top: 8px; font-weight: 500; color: #303133; font-size: 13px; }
.timeline-sub { color: #909399; font-size: 12px; margin-top: 2px; }
.timeline-line {
  position: absolute;
  top: 18px;
  right: -50%;
  width: 100%;
  height: 2px;
  background: #dcdfe6;
  z-index: -1;
}
.lifecycle-card { min-height: 280px; }
.card-header { display: flex; align-items: center; gap: 10px; }
.card-title { font-size: 16px; font-weight: 600; color: #303133; }
.card-desc { color: #606266; font-size: 13px; margin-top: 6px; line-height: 1.5; }
.feature-list { padding-left: 0; }
.feature-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 0;
  color: #606266;
  font-size: 13px;
}
</style>
