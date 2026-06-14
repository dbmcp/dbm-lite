<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DBA老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <div class="page-container">
    <div class="page-header">DB 生命周期管理</div>

    <el-row :gutter="16">
      <el-col :span="8">
        <el-card shadow="hover" class="lifecycle-card" :body-style="{ padding: '16px' }">
          <div class="card-header">
            <el-icon :size="22" color="#2e6ba8"><Coin /></el-icon>
            <div class="card-title">数据库创建与初始化</div>
          </div>
          <div class="card-desc">描述（功能开发中）：新建数据库实例、表结构初始化、元数据注册。</div>
          <el-divider />
          <div class="feature-list">
            <div class="feature-item">
              <el-icon><Check /></el-icon>
              <span>选择数据库类型 / 环境</span>
              <el-tag size="small" type="info" round>开发中</el-tag>
            </div>
            <div class="feature-item">
              <el-icon><Check /></el-icon>
              <span>配置编码、字符集、连接数</span>
              <el-tag size="small" type="info" round>开发中</el-tag>
            </div>
            <div class="feature-item">
              <el-icon><Check /></el-icon>
              <span>初始化脚本执行</span>
              <el-tag size="small" type="info" round>开发中</el-tag>
            </div>
          </div>
          <el-button disabled type="primary" style="width:100%;margin-top:12px;">创建数据库（功能开发中）</el-button>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card shadow="hover" class="lifecycle-card" :body-style="{ padding: '16px' }">
          <div class="card-header">
            <el-icon :size="22" color="#2e6ba8"><MagicStick /></el-icon>
            <div class="card-title">版本管理与升级</div>
          </div>
          <div class="card-desc">描述（功能开发中）：记录数据库版本、Patch 历史、执行小版本升级或大版本迁移。</div>
          <el-divider />
          <div class="feature-list">
            <div class="feature-item">
              <el-icon><Check /></el-icon>
              <span>版本信息采集</span>
              <el-tag size="small" type="info" round>开发中</el-tag>
            </div>
            <div class="feature-item">
              <el-icon><Check /></el-icon>
              <span>升级脚本编排</span>
              <el-tag size="small" type="info" round>开发中</el-tag>
            </div>
            <div class="feature-item">
              <el-icon><Check /></el-icon>
              <span>升级前后校验</span>
              <el-tag size="small" type="info" round>开发中</el-tag>
            </div>
          </div>
          <el-button disabled type="primary" style="width:100%;margin-top:12px;">升级数据库（功能开发中）</el-button>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card shadow="hover" class="lifecycle-card" :body-style="{ padding: '16px' }">
          <div class="card-header">
            <el-icon :size="22" color="#2e6ba8"><Coin /></el-icon>
            <div class="card-title">备份与恢复</div>
          </div>
          <div class="card-desc">描述（功能开发中）：定时备份策略配置、备份文件管理、一键恢复到指定时间点。</div>
          <el-divider />
          <div class="feature-list">
            <div class="feature-item">
              <el-icon><Check /></el-icon>
              <span>备份策略（全量/增量）</span>
              <el-tag size="small" type="info" round>开发中</el-tag>
            </div>
            <div class="feature-item">
              <el-icon><Check /></el-icon>
              <span>备份存储与清理</span>
              <el-tag size="small" type="info" round>开发中</el-tag>
            </div>
            <div class="feature-item">
              <el-icon><Check /></el-icon>
              <span>时间点恢复（PITR）</span>
              <el-tag size="small" type="info" round>开发中</el-tag>
            </div>
          </div>
          <el-button disabled type="primary" style="width:100%;margin-top:12px;">创建备份策略（功能开发中）</el-button>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" style="margin-top:16px;">
      <el-col :span="8">
        <el-card shadow="hover" class="lifecycle-card" :body-style="{ padding: '16px' }">
          <div class="card-header">
            <el-icon :size="22" color="#2e6ba8"><Document /></el-icon>
            <div class="card-title">运行监控</div>
          </div>
          <div class="card-desc">描述（功能开发中）：监控数据库性能指标、连接数、QPS、慢查询等。</div>
          <el-divider />
          <div class="feature-list">
            <div class="feature-item">
              <el-icon><Check /></el-icon>
              <span>性能指标大盘</span>
              <el-tag size="small" type="info" round>开发中</el-tag>
            </div>
            <div class="feature-item">
              <el-icon><Check /></el-icon>
              <span>告警规则配置</span>
              <el-tag size="small" type="info" round>开发中</el-tag>
            </div>
            <div class="feature-item">
              <el-icon><Check /></el-icon>
              <span>慢查询分析</span>
              <el-tag size="small" type="info" round>开发中</el-tag>
            </div>
          </div>
          <el-button disabled type="primary" style="width:100%;margin-top:12px;">查看监控（功能开发中）</el-button>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card shadow="hover" class="lifecycle-card" :body-style="{ padding: '16px' }">
          <div class="card-header">
            <el-icon :size="22" color="#2e6ba8"><Document /></el-icon>
            <div class="card-title">容量与配额</div>
          </div>
          <div class="card-desc">描述（功能开发中）：统计各数据库/表占用空间、设置配额、自动扩容建议。</div>
          <el-divider />
          <div class="feature-list">
            <div class="feature-item">
              <el-icon><Check /></el-icon>
              <span>容量审计</span>
              <el-tag size="small" type="info" round>开发中</el-tag>
            </div>
            <div class="feature-item">
              <el-icon><Check /></el-icon>
              <span>配额管理</span>
              <el-tag size="small" type="info" round>开发中</el-tag>
            </div>
            <div class="feature-item">
              <el-icon><Check /></el-icon>
              <span>扩容建议</span>
              <el-tag size="small" type="info" round>开发中</el-tag>
            </div>
          </div>
          <el-button disabled type="primary" style="width:100%;margin-top:12px;">容量分析（功能开发中）</el-button>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card shadow="hover" class="lifecycle-card" :body-style="{ padding: '16px' }">
          <div class="card-header">
            <el-icon :size="22" color="#b25d5d"><Delete /></el-icon>
            <div class="card-title">下线与销毁</div>
          </div>
          <div class="card-desc">描述（功能开发中）：数据库下线审批、数据归档、彻底销毁。</div>
          <el-divider />
          <div class="feature-list">
            <div class="feature-item">
              <el-icon><Check /></el-icon>
              <span>下线申请与审批</span>
              <el-tag size="small" type="info" round>开发中</el-tag>
            </div>
            <div class="feature-item">
              <el-icon><Check /></el-icon>
              <span>数据归档</span>
              <el-tag size="small" type="info" round>开发中</el-tag>
            </div>
            <div class="feature-item">
              <el-icon><Check /></el-icon>
              <span>安全销毁</span>
              <el-tag size="small" type="info" round>开发中</el-tag>
            </div>
          </div>
          <el-button disabled type="danger" style="width:100%;margin-top:12px;">提交下线申请（功能开发中）</el-button>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="never" style="margin-top:16px;">
      <template #header>
        <div style="display:flex;align-items:center;justify-content:space-between;">
          <span style="font-weight:600;color:#2e6ba8;">生命周期节点概览（功能开发中）</span>
          <el-tag type="info" round>演示数据</el-tag>
        </div>
      </template>
      <el-timeline>
        <el-timeline-item timestamp="2026-06-01 10:00" color="#2e6ba8" placement="top">
          <el-card shadow="never" style="border:1px solid #d4e4f7;">
            <b>数据库创建</b> · production-db · MySQL 8.0
            <div style="color:#909399;margin-top:6px;font-size:13px;">由 admin 创建，环境: 生产</div>
          </el-card>
        </el-timeline-item>
        <el-timeline-item timestamp="2026-06-02 22:00" color="#67c23a">
          <el-card shadow="never" style="border:1px solid #e1f3d8;">
            <b>应用初始化脚本</b> · 建表 / 初始数据导入完成
            <div style="color:#909399;margin-top:6px;font-size:13px;">由 ops 执行 schema v1.0</div>
          </el-card>
        </el-timeline-item>
        <el-timeline-item timestamp="2026-06-05 02:00" color="#409eff">
          <el-card shadow="never" style="border:1px solid #b3d8ff;">
            <b>首次全量备份</b> · 2.86 GB · 耗时 12m35s
            <div style="color:#909399;margin-top:6px;font-size:13px;">存储于 /data/backups/prod</div>
          </el-card>
        </el-timeline-item>
        <el-timeline-item timestamp="2026-06-10 14:20" color="#e6a23c">
          <el-card shadow="never" style="border:1px solid #faecd8;">
            <b>小版本升级</b> · 8.0.35 → 8.0.37
            <div style="color:#909399;margin-top:6px;font-size:13px;">由 dba 审批后执行，停机 3 分钟</div>
          </el-card>
        </el-timeline-item>
        <el-timeline-item timestamp="—" color="#909399">
          <el-card shadow="never" style="border:1px dashed #c0c4cc;background:#f5f7fa;">
            <b>待开发</b> · 更多节点（监控告警 / 容量评估 / 归档 / 销毁）
            <div style="color:#909399;margin-top:6px;font-size:13px;">描述（功能开发中）</div>
          </el-card>
        </el-timeline-item>
      </el-timeline>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import {
  Coin, MagicStick, Document, Delete, Check
} from '@element-plus/icons-vue'
</script>

<style scoped>
.lifecycle-card { min-height: 320px; }
.card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}
.card-title { font-size: 16px; font-weight: 600; color: #2e6ba8; }
.card-desc { color: #606266; font-size: 13px; line-height: 1.6; }
.feature-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.feature-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  color: #606266;
  font-size: 13px;
}
.feature-item > span { flex: 1; }
</style>

