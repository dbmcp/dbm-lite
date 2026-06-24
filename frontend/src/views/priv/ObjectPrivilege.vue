<template>
  <div class="object-privilege">
    <div class="page-header">
      <div class="title-wrap">
        <el-icon :size="22" color="#409EFF"><Key /></el-icon>
        <span class="title-text">对象权限管理</span>
      </div>
      <el-breadcrumb separator="/" class="crumb">
        <el-breadcrumb-item>DB 操作管控域</el-breadcrumb-item>
        <el-breadcrumb-item>对象权限</el-breadcrumb-item>
      </el-breadcrumb>
    </div>

    <el-card class="card" shadow="never">
      <div class="empty-state">
        <el-icon :size="64" color="#ccc"><Lock /></el-icon>
        <p class="empty-text">对象权限管理功能已优化调整</p>
        <p class="empty-hint">权限管理已集成到数据源配置和用户管理中</p>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { Key, Lock } from '@element-plus/icons-vue'
</script>

<style scoped>
.object-privilege {
  min-height: 100%;
  padding: 16px;
  background: #f5f7fa;
}

.page-header {
  margin-bottom: 16px;
}
.page-header .title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
.title-text {
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
}

.card {
  min-height: 400px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
}

.empty-text {
  margin-top: 16px;
  font-size: 16px;
  color: #6b7280;
}

.empty-hint {
  margin-top: 8px;
  font-size: 14px;
  color: #9ca3af;
}
</style>