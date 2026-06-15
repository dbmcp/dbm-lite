<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DB老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <div class="page-container">
    <div class="page-header">
      <div class="title-row">
        <span class="title-text">数据源管理</span>
        <span class="subtitle-text">共 {{ total }} 个数据源</span>
      </div>
      <div class="header-actions">
        <el-tooltip content="切换视图">
          <el-radio-group v-model="viewMode" size="default">
            <el-radio-button value="table">表格</el-radio-button>
            <el-radio-button value="card">卡片</el-radio-button>
          </el-radio-group>
        </el-tooltip>
        <el-button :icon="RefreshIcon" @click="loadList">刷新</el-button>
        <el-button type="primary" :icon="PlusIcon" v-if="userStore.isAdmin" @click="openDialog()">新建数据源</el-button>
        <el-button :icon="QuestionFilled" @click="showHelp = true">帮助</el-button>
      </div>
    </div>

    <div class="card">
      <div class="toolbar">
        <el-input v-model="keyword" placeholder="搜索名称 / 主机 / 备注" clearable style="width:260px;" :prefix-icon="SearchIcon" @input="debounceSearch" />
        <el-select v-model="filterDbType" placeholder="全部类型" clearable style="width:140px;" @change="loadList">
          <el-option label="MySQL" value="mysql" />
          <el-option label="TiDB" value="tidb" />
          <el-option label="SQLite" value="sqlite" />
        </el-select>
        <el-select v-model="filterConnStatus" placeholder="连接状态" clearable style="width:140px;" @change="loadList">
          <el-option label="已连通" value="ok" />
          <el-option label="未连通" value="fail" />
          <el-option label="未检测" value="untested" />
        </el-select>
        <el-select v-model="filterStatus" placeholder="使用状态" clearable style="width:140px;" @change="loadList">
          <el-option label="启用" value="active" />
          <el-option label="禁用" value="inactive" />
        </el-select>
        <el-select v-model="sortBy" placeholder="按..." style="width:160px;" @change="loadList">
          <el-option label="最新创建" value="" />
          <el-option label="名称" value="name" />
          <el-option label="最近使用" value="recent" />
          <el-option label="最近测试" value="lasttest" />
        </el-select>
        <div class="toolbar-right">
          <ColumnToggle v-model="colVisible" :columns="columns" v-if="viewMode === 'table'" />
        </div>
      </div>

      <div v-if="!loading && list.length === 0" class="empty-state">
        <div class="empty-icon" style="font-size:64px;line-height:1;">📭</div>
        <div class="empty-title">暂无数据源</div>
        <div class="empty-desc">点击右上角"新建数据源"按钮添加第一个连接</div>
        <el-button type="primary" :icon="PlusIcon" v-if="userStore.isAdmin" style="margin-top:16px;" @click="openDialog()">新建数据源</el-button>
      </div>

      <el-table v-else-if="viewMode === 'table'" :data="list" style="width:100%;" v-loading="loading" stripe :default-sort="{ prop: 'createdAt', order: 'descending' }">
        <el-table-column label="状态" width="90" align="center" fixed="left">
          <template #default="{ row }">
            <el-tooltip :content="connStatusTooltip(row)" placement="top" :show-after="400">
              <span class="status-dot" :class="statusClass(row)"></span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="数据源名称" sortable v-if="colVisible.name" min-width="180">
          <template #default="{ row }">
            <span class="color-bar" :style="{ background: colorOf(row.colorLabel) }"></span>
            <el-link type="primary" :is-underline="false" @click="goDetail(row)">{{ row.name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="数据库类型" sortable v-if="colVisible.dbType" width="110">
          <template #default="{ row }">
            <el-tag v-if="row.dbType === 'mysql'" type="primary" size="small">
              <span style="margin-right:4px;">🗄️</span>MySQL
            </el-tag>
            <el-tag v-else-if="row.dbType === 'tidb'" type="success" size="small">
              <span style="margin-right:4px;">⚡</span>TiDB
            </el-tag>
            <el-tag v-else-if="row.dbType === 'sqlite'" type="info" size="small">
              <span style="margin-right:4px;">📁</span>SQLite
            </el-tag>
            <el-tag v-else size="small">{{ row.dbType }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="地址" sortable v-if="colVisible.host" min-width="200">
          <template #default="{ row }">
            <el-tooltip v-if="row.dbType === 'sqlite'" :content="row.filePath || '(内存库)'" placement="top">
              <span class="mono">{{ row.filePath || '(内存库)' }}</span>
            </el-tooltip>
            <span v-else class="mono">{{ row.host }}:{{ row.port }}</span>
          </template>
        </el-table-column>
        <el-table-column label="默认数据库" v-if="colVisible.defaultDatabase" min-width="120">
          <template #default="{ row }">
            <el-tag size="small" v-if="row.defaultDatabase">{{ row.defaultDatabase }}</el-tag>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="username" label="用户名" v-if="colVisible.username" min-width="110">
          <template #default="{ row }">
            <span v-if="row.username">{{ row.username }}</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="连接延迟" v-if="colVisible.latency" width="120">
          <template #default="{ row }">
            <span v-if="row.connStatus === 'ok' && row.connLatencyMs" :style="{ color: row.connLatencyMs > 200 ? '#e6a23c' : '#67c23a' }">
              {{ row.connLatencyMs }} ms
            </span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="环境" v-if="colVisible.env" width="110">
          <template #default="{ row }">
            <el-tag v-if="row.env === 'prod'" type="danger" size="small">生产</el-tag>
            <el-tag v-else-if="row.env === 'stage'" type="warning" size="small">预发</el-tag>
            <el-tag v-else-if="row.env === 'test'" type="success" size="small">测试</el-tag>
            <el-tag v-else size="small">开发</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="最近连接" v-if="colVisible.lastTestAt" min-width="170" sortable="custom">
          <template #default="{ row }">
            <el-tooltip v-if="row.lastConnTestAt" :content="row.lastConnTestAt" placement="top">
              <span style="color:#606266;font-size:13px;">{{ row.lastConnTestAt }}</span>
            </el-tooltip>
            <span v-else class="text-muted">尚未测试</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="330" fixed="right">
          <template #default="{ row }">
            <el-button size="small" :type="row.connStatus === 'ok' ? 'success' : ''" @click="refreshConn(row)" :loading="refreshingConnId === row.datasourceId">
              {{ refreshingConnId === row.datasourceId ? '测试中' : '测试连接' }}
            </el-button>
            <el-button size="small" @click="goWorkbench(row)">SQL</el-button>
            <el-dropdown @command="(c) => onMoreCommand(c, row)" trigger="click">
              <el-button size="small">更多<el-icon class="el-icon--right"><arrow-down /></el-icon></el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="detail" :icon="ViewIcon">查看详情</el-dropdown-item>
                  <el-dropdown-item command="copy" :icon="CopyIcon" v-if="userStore.isAdmin">复制</el-dropdown-item>
                  <el-dropdown-item command="edit" :icon="EditIcon" v-if="userStore.isAdmin">编辑</el-dropdown-item>
                  <el-dropdown-item command="setDefault" :icon="StarIcon">设为默认</el-dropdown-item>
                  <el-dropdown-item command="resetPwd" :icon="LockIcon" v-if="userStore.isAdmin">重置密码</el-dropdown-item>
                  <el-dropdown-item divided command="delete" :icon="DeleteIcon" v-if="userStore.isAdmin" style="color:#f56c6c;">删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>

      <div v-else class="card-view" v-loading="loading">
        <div class="card-item" v-for="row in list" :key="row.datasourceId">
          <div class="card-header" :style="{ borderLeft: '4px solid ' + colorOf(row.colorLabel) }">
            <div class="card-title">
              <el-tooltip :content="connStatusTooltip(row)" placement="top">
                <span class="status-dot" :class="statusClass(row)" style="margin-right:6px;"></span>
              </el-tooltip>
              <el-link type="primary" :is-underline="false" class="ds-name" @click="goDetail(row)">{{ row.name }}</el-link>
            </div>
            <div class="card-type">
              <el-tag v-if="row.dbType === 'mysql'" type="primary" size="small">MySQL</el-tag>
              <el-tag v-else-if="row.dbType === 'tidb'" type="success" size="small">TiDB</el-tag>
              <el-tag v-else-if="row.dbType === 'sqlite'" type="info" size="small">SQLite</el-tag>
            </div>
          </div>
          <div class="card-body">
            <div class="card-row"><span class="label">地址:</span><span class="mono value">{{ row.dbType === 'sqlite' ? (row.filePath || '(内存库)') : (row.host + ':' + row.port) }}</span></div>
            <div class="card-row"><span class="label">数据库:</span><span class="value">{{ row.defaultDatabase || '-' }}</span></div>
            <div class="card-row"><span class="label">用户名:</span><span class="value">{{ row.username || '-' }}</span></div>
            <div class="card-row"><span class="label">环境:</span>
              <span class="value">
                <el-tag v-if="row.env === 'prod'" type="danger" size="small">生产</el-tag>
                <el-tag v-else-if="row.env === 'stage'" type="warning" size="small">预发</el-tag>
                <el-tag v-else-if="row.env === 'test'" type="success" size="small">测试</el-tag>
                <el-tag v-else size="small">开发</el-tag>
              </span>
            </div>
            <div class="card-row"><span class="label">最近测试:</span>
              <span class="value">
                <el-tag v-if="row.connStatus === 'ok'" type="success" size="small" style="margin-right:4px;">连通</el-tag>
                <el-tag v-else-if="row.connStatus === 'fail'" type="danger" size="small" style="margin-right:4px;">失败</el-tag>
                <span v-if="row.connLatencyMs" :style="{ color: row.connLatencyMs > 200 ? '#e6a23c' : '#67c23a', marginRight: '6px' }">{{ row.connLatencyMs }} ms</span>
                <span v-if="row.lastConnTestAt" style="color:#909399;font-size:12px;">{{ row.lastConnTestAt }}</span>
              </span>
            </div>
          </div>
          <div class="card-footer">
            <el-button size="small" type="success" @click="refreshConn(row)" :loading="refreshingConnId === row.datasourceId">测试连接</el-button>
            <el-button size="small" @click="goWorkbench(row)">SQL工作台</el-button>
            <el-button size="small" type="primary" @click="openDialog(row)" v-if="userStore.isAdmin">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)" v-if="userStore.isAdmin">删除</el-button>
          </div>
        </div>
      </div>

      <div style="margin-top:16px;display:flex;justify-content:space-between;align-items:center;">
        <span v-if="list.length > 0" class="text-muted" style="font-size:12px;">第 {{ (current - 1) * pageSize + 1 }}-{{ Math.min(current * pageSize, total) }} 条 / 共 {{ total }} 条</span>
        <span v-else></span>
        <el-pagination
          v-model:current-page="current"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="sizes, prev, pager, next, jumper"
          @current-change="loadList"
          @size-change="loadList"
        />
      </div>
    </div>

    <!-- 新建 / 编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑数据源' : '新建数据源'" width="720px" :close-on-click-modal="false">
      <el-form :model="form" label-width="110px" ref="formRef" :rules="rules">
        <div class="form-section-title">基本信息</div>
        <el-form-item label="数据源名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入数据源名称（2-64 字符）" />
        </el-form-item>

        <el-form-item label="数据库类型" prop="dbType">
          <el-radio-group v-model="form.dbType" @change="onDbTypeChange">
            <el-radio-button value="mysql" style="padding:0 20px;">🗄️ MySQL</el-radio-button>
            <el-radio-button value="tidb" style="padding:0 20px;">⚡ TiDB</el-radio-button>
            <el-radio-button value="sqlite" style="padding:0 20px;">📁 SQLite</el-radio-button>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="颜色标签" prop="colorLabel">
          <div class="color-picker">
            <div v-for="c in colorOptions" :key="c.value" class="color-option"
              :class="{ active: form.colorLabel === c.value }"
              :style="{ background: c.color }"
              @click="form.colorLabel = c.value">
              <span v-if="form.colorLabel === c.value" class="check">✓</span>
            </div>
          </div>
        </el-form-item>

        <el-form-item label="环境">
          <el-select v-model="form.env" style="width:100%;">
            <el-option label="生产" value="prod" />
            <el-option label="预发" value="stage" />
            <el-option label="测试" value="test" />
            <el-option label="开发" value="dev" />
          </el-select>
        </el-form-item>

        <div class="form-section-title">连接配置</div>

        <template v-if="form.dbType !== 'sqlite'">
          <el-row :gutter="12">
            <el-col :span="12">
              <el-form-item label="主机/IP" prop="host">
                <el-input v-model="form.host" placeholder="如 127.0.0.1" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="端口" prop="port">
                <el-input-number v-model="form.port" :min="1" :max="65535" style="width:100%;" />
              </el-form-item>
            </el-col>
          </el-row>

          <el-row :gutter="12">
            <el-col :span="12">
              <el-form-item label="用户名" prop="username">
                <el-input v-model="form.username" placeholder="数据库登录账号" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="密码" prop="password">
                <el-input v-model="form.password" type="password" show-password :placeholder="isEdit ? '不修改请留空' : '请输入数据库密码'" />
              </el-form-item>
            </el-col>
          </el-row>

          <el-row :gutter="12">
            <el-col :span="12">
              <el-form-item label="初始数据库">
                <el-input v-model="form.defaultDatabase" placeholder="如 mysql / test（可留空）" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="字符集">
                <el-select v-model="form.charset" style="width:100%;" placeholder="utf8mb4">
                  <el-option label="utf8mb4 (推荐)" value="utf8mb4" />
                  <el-option label="utf8" value="utf8" />
                  <el-option label="latin1" value="latin1" />
                  <el-option label="gbk" value="gbk" />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>

          <el-row :gutter="12">
            <el-col :span="12">
              <el-form-item label="时区">
                <el-select v-model="form.timezone" style="width:100%;" placeholder="Local">
                  <el-option label="自动检测 (Local)" value="Local" />
                  <el-option label="UTC" value="UTC" />
                  <el-option label="Asia/Shanghai (UTC+8)" value="Asia/Shanghai" />
                  <el-option label="Asia/Tokyo (UTC+9)" value="Asia/Tokyo" />
                  <el-option label="America/New_York (UTC-5)" value="America/New_York" />
                  <el-option label="Europe/London" value="Europe/London" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="连接超时秒">
                <el-input-number v-model="form.timeout" :min="1" :max="300" style="width:100%;" />
              </el-form-item>
            </el-col>
          </el-row>

          <el-form-item>
            <el-checkbox v-model="form.sslMode" :true-value="'true'" :false-value="'false'">启用 SSL 连接</el-checkbox>
            <el-checkbox v-model="form.readOnly" style="margin-left:16px;">只读模式（SQL IDE 禁用 DDL/DML）</el-checkbox>
          </el-form-item>

          <el-form-item v-if="form.sslMode === 'true'" label="SSL 证书路径">
            <el-input v-model="form.sslCaFile" placeholder="如 /etc/mysql/ca.pem（可选）" />
          </el-form-item>
        </template>

        <template v-if="form.dbType === 'sqlite'">
          <el-form-item label="数据库文件路径">
            <el-input v-model="form.filePath" placeholder="如 test.db（相对路径）或 /data/test.db；留空使用内存库" />
            <div style="font-size:12px;color:#909399;margin-top:4px;">支持 ~ 指代用户目录；留空或填写 :memory: 为内存数据库（重启后数据清空）</div>
          </el-form-item>
          <el-form-item>
            <el-checkbox v-model="form.memoryDB">使用内存数据库（临时存储，刷新页面清空）</el-checkbox>
            <el-checkbox v-model="form.readOnly" style="margin-left:16px;">只读模式</el-checkbox>
          </el-form-item>
          <el-form-item label="读写模式">
            <el-select v-model="form.openMode" style="width:100%;">
              <el-option label="读写模式（rw）" value="rw" />
              <el-option label="只读模式（ro）" value="ro" />
            </el-select>
          </el-form-item>
          <el-form-item label="默认数据库">
            <el-input v-model="form.defaultDatabase" placeholder="默认 main（通常不需要修改）" />
          </el-form-item>
        </template>

        <div class="form-section-title">组织与描述</div>
        <el-form-item label="关联业务">
          <el-select v-model="form.businessId" filterable clearable placeholder="选择关联业务（可选）" style="width:100%;">
            <el-option v-for="b in businesses" :key="b.businessId" :label="b.name + ' (' + (b.businessId || '') + ')'" :value="b.businessId" />
          </el-select>
        </el-form-item>
        <el-form-item label="关联服务器">
          <el-select v-model="form.serverId" filterable clearable placeholder="选择已有服务器（会自动填充主机）" style="width:100%;margin-bottom:8px;" @change="onServerSelect">
            <el-option v-for="s in servers" :key="s.serverId" :label="s.name + ' (' + s.host + ')'" :value="s.serverId" />
          </el-select>
          <el-checkbox v-model="form.autoCreateServer">手动填写主机时自动创建服务器记录</el-checkbox>
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="form.tags" placeholder="逗号分隔，如: prod,核心,分析" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" placeholder="选填，如 连接用途 / 维护者" />
        </el-form-item>
      </el-form>

      <div class="dialog-footer">
        <span v-if="testResult !== null" class="test-result" :class="{ ok: testResult.success, fail: !testResult.success }">
          <span v-if="testResult.success">✅ 连接成功（{{ testResult.latencyMs }} ms{{ testResult.version ? '，版本: ' + testResult.version : '' }}）</span>
          <span v-else>❌ {{ testResult.message || '连接失败' }}</span>
        </span>
        <span></span>
        <div>
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="warning" @click="testConnection" :loading="testing">测试连接</el-button>
          <el-button type="primary" @click="save">保存</el-button>
        </div>
      </div>
    </el-dialog>

    <!-- 数据库 / 表列表弹窗 -->
    <el-dialog v-model="dbDialogVisible" title="数据库 / 表列表" width="600px">
      <el-tabs v-model="activeDb" v-if="databases.length > 0" @tab-click="loadTables">
        <el-tab-pane v-for="d in databases" :key="d" :label="d" :name="d">
          <el-table :data="(tablesMap[d] || []).map((t: any) => ({ name: t }))" size="small" v-loading="tablesLoading">
            <el-table-column label="表名" prop="name" />
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-dialog>

    <!-- 重置密码弹窗 -->
    <el-dialog v-model="resetPwdVisible" title="重置密码" width="480px">
      <el-form :model="pwdForm" label-width="100px">
        <el-form-item label="新密码">
          <el-input v-model="pwdForm.password" type="password" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetPwdVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmResetPwd">确认</el-button>
      </template>
    </el-dialog>

    <!-- 帮助弹窗 -->
    <el-dialog v-model="showHelp" title="数据源管理使用说明" width="640px">
      <div style="line-height:1.8;color:#606266;">
        <p><strong>📌 快速开始</strong></p>
        <p>1. 点击右上角"新建数据源"按钮添加数据库连接</p>
        <p>2. 选择数据库类型（MySQL / TiDB / SQLite），填写主机、端口、账号、密码</p>
        <p>3. 点击"测试连接"验证配置，成功后点击"保存"</p>
        <p>4. 保存后点击"SQL"按钮进入 SQL IDE 执行查询</p>
        <p style="margin-top:12px;"><strong>🛡️ 安全与权限</strong></p>
        <p>· 密码采用 AES-256 加密存储，前后端传输均走 HTTPS（若配置）</p>
        <p>· 只读模式下，SQL IDE 禁止执行 INSERT / UPDATE / DELETE / DROP 等变更操作</p>
        <p>· 建议为生产环境分配独立只读账号</p>
        <p style="margin-top:12px;"><strong>🔍 状态指示灯</strong></p>
        <p>· 🟢 绿色：最近连接测试成功</p>
        <p>· 🔴 红色：最近测试失败</p>
        <p>· ⚪ 灰色：尚未测试或检测中</p>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  Search as SearchIcon, Refresh as RefreshIcon, Plus as PlusIcon,
  QuestionFilled, ArrowDown, DocumentCopy as CopyIcon, Edit as EditIcon,
  Delete as DeleteIcon, View as ViewIcon, StarFilled as StarIcon, Lock as LockIcon
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { useDatasourceStore } from '@/stores/datasource'
import {
  listDatasources, createDatasource, updateDatasource, deleteDatasource, copyDatasource,
  testConnection as testConn, testConnectionById, listDatabases, listTables,
  listDatasource, createDatasourceV2, updateDatasourceV2, deleteDatasourceV2
} from '@/api/datasource'
import request from '@/api/request'
import ColumnToggle from '@/components/ColumnToggle.vue'
import type { TestResult, DatasourceForm } from '@/api/datasource'

const router = useRouter()
const userStore = useUserStore()
const dsStore = useDatasourceStore()
const loading = ref(false)

const viewMode = ref<'table' | 'card'>('table')
const keyword = ref('')
const filterDbType = ref('')
const filterConnStatus = ref('')
const filterStatus = ref('')
const sortBy = ref('')
const showHelp = ref(false)

const columns = [
  { key: 'name', label: '数据源名称' },
  { key: 'dbType', label: '类型' },
  { key: 'host', label: '地址' },
  { key: 'username', label: '用户名' },
  { key: 'defaultDatabase', label: '默认库' },
  { key: 'latency', label: '延迟' },
  { key: 'env', label: '环境' },
  { key: 'lastTestAt', label: '最近连接' }
]
const colVisible = ref<Record<string, boolean>>({
  name: true, dbType: true, host: true, username: true, defaultDatabase: true,
  latency: true, env: true, lastTestAt: true
})
const list = ref<any[]>([])
const total = ref(0)
const current = ref(1)
const pageSize = ref(10)

const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref('')
const formRef = ref<FormInstance>()
const testResult = ref<TestResult | null>(null)
const testing = ref(false)

const businesses = ref<any[]>([])
const servers = ref<any[]>([])

const colorOptions = [
  { value: 'blue', color: '#409eff' },
  { value: 'green', color: '#67c23a' },
  { value: 'red', color: '#f56c6c' },
  { value: 'yellow', color: '#e6a23c' },
  { value: 'purple', color: '#8e44ad' },
  { value: 'orange', color: '#e67e22' },
  { value: 'gray', color: '#909399' }
]

const form = reactive<any>({
  name: '', dbType: 'mysql', host: '', port: 3306, username: '', password: '',
  defaultDatabase: '', filePath: '', openMode: 'rw',
  charset: 'utf8mb4', timezone: 'Local', sslMode: 'false', sslCaFile: '',
  readOnly: false, colorLabel: 'blue', timeout: 10,
  tags: '', businessId: '', serverId: '', autoCreateServer: true,
  projectId: '', env: 'dev', remark: '', status: 'active', memoryDB: false
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入数据源名称', trigger: 'blur' },
    { min: 2, max: 64, message: '长度在 2 到 64 个字符', trigger: 'blur' }
  ],
  dbType: [{ required: true, message: '请选择数据库类型' }],
  host: [
    {
      required: true,
      validator: (_rule, value, cb) => {
        if (form.dbType === 'sqlite') return cb()
        if (!value) return cb(new Error('请填写主机'))
        cb()
      },
      trigger: 'blur'
    }
  ]
}

const dbDialogVisible = ref(false)
const databases = ref<string[]>([])
const activeDb = ref('')
const tablesMap = ref<Record<string, string[]>>({})
const tablesLoading = ref(false)
const currentDs = ref<any>(null)
const refreshingConnId = ref<string>('')

const resetPwdVisible = ref(false)
const pwdForm = reactive({ password: '', datasourceId: '' })

let searchTimer: number | null = null
function debounceSearch() {
  if (searchTimer) window.clearTimeout(searchTimer)
  searchTimer = window.setTimeout(loadList, 300)
}

function colorOf(label: string) {
  const map: Record<string, string> = {
    blue: '#409eff', green: '#67c23a', red: '#f56c6c',
    yellow: '#e6a23c', purple: '#8e44ad', orange: '#e67e22', gray: '#909399'
  }
  return map[label] || '#409eff'
}

function statusClass(row: any) {
  if (row.connStatus === 'ok') return 'status-ok'
  if (row.connStatus === 'fail') return 'status-fail'
  return 'status-none'
}

function connStatusTooltip(row: any) {
  if (row.connStatus === 'ok') {
    return `连接成功${row.connLatencyMs ? '（' + row.connLatencyMs + ' ms）' : ''}\n${row.lastConnTestAt || ''}`
  }
  if (row.connStatus === 'fail') return `连接失败\n${row.lastConnTestAt || ''}`
  return '尚未测试'
}

function onDbTypeChange(type: string) {
  testResult.value = null
  if (type === 'mysql') {
    form.port = 3306
    form.defaultDatabase = 'mysql'
    form.filePath = ''
  } else if (type === 'tidb') {
    form.port = 4000
    form.defaultDatabase = 'test'
    form.filePath = ''
  } else if (type === 'sqlite') {
    form.defaultDatabase = 'main'
    form.host = ''
    form.username = ''
    form.port = 0
  }
}

function onServerSelect(serverId: string) {
  const server = servers.value.find(s => s.serverId === serverId)
  if (server) {
    form.host = server.host
    if (!form.port) form.port = 3306
  }
}

async function loadBusinessesAndServers() {
  try {
    const [bizRes, srvRes] = await Promise.all([
      request({ url: '/businesses', method: 'GET' }),
      request({ url: '/servers', method: 'GET' })
    ]) as any
    businesses.value = bizRes?.list || bizRes || []
    servers.value = srvRes?.list || srvRes || []
  } catch { /* ignore */ }
}

async function loadList() {
  loading.value = true
  try {
    const status = filterConnStatus.value && !filterStatus.value
      ? filterConnStatus.value
      : (filterStatus.value || '')
    const r: any = await listDatasources(
      current.value, pageSize.value, keyword.value, filterDbType.value, status, sortBy.value
    )
    list.value = r.list || []
    total.value = r.total || 0
  } finally {
    loading.value = false
  }
}

function openDialog(row?: any) {
  loadBusinessesAndServers()
  testResult.value = null
  if (row) {
    isEdit.value = true
    editId.value = row.datasourceId
    Object.assign(form, {
      name: row.name,
      dbType: row.dbType || 'mysql',
      host: row.host || '',
      port: row.port || 3306,
      username: row.username || '',
      password: '',
      defaultDatabase: row.defaultDatabase || '',
      filePath: row.filePath || (row.dbType === 'sqlite' ? row.host : '') || '',
      openMode: row.openMode || 'rw',
      charset: row.charset || 'utf8mb4',
      timezone: row.timezone || 'Local',
      sslMode: row.sslMode || 'false',
      sslCaFile: row.sslCaFile || '',
      readOnly: !!row.readOnly,
      colorLabel: row.colorLabel || 'blue',
      timeout: row.timeout || 10,
      tags: row.tags || '',
      businessId: row.businessId || '',
      serverId: row.serverId || '',
      autoCreateServer: true,
      projectId: row.projectId || '',
      env: row.env || 'dev',
      remark: row.remark || '',
      status: row.status || 'active',
      memoryDB: !!(row.filePath && row.filePath.indexOf(':memory') >= 0)
    })
  } else {
    isEdit.value = false
    editId.value = ''
    Object.assign(form, {
      name: '', dbType: 'mysql', host: '', port: 3306, username: '', password: '',
      defaultDatabase: 'mysql', filePath: '', openMode: 'rw',
      charset: 'utf8mb4', timezone: 'Local', sslMode: 'false', sslCaFile: '',
      readOnly: false, colorLabel: 'blue', timeout: 10,
      tags: '', businessId: '', serverId: '', autoCreateServer: true,
      projectId: '', env: 'dev', remark: '', status: 'active', memoryDB: false
    })
  }
  dialogVisible.value = true
}

async function save() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    if (!testResult.value) {
      try {
        await ElMessageBox.confirm(
          '尚未成功测试连接，是否仍然保存？（保存后可在列表中点击"测试连接"）',
          '提示',
          { type: 'warning', confirmButtonText: '仍然保存', cancelButtonText: '取消' }
        )
      } catch { return }
    } else if (!testResult.value.success) {
      try {
        await ElMessageBox.confirm(
          '最近一次测试连接失败，是否仍然保存？',
          '提示',
          { type: 'warning', confirmButtonText: '仍然保存', cancelButtonText: '取消' }
        )
      } catch { return }
    }
    try {
      const payload: any = { ...form }
      if (payload.dbType === 'sqlite' && payload.memoryDB) {
        payload.filePath = ':memory:'
      }
      if (isEdit.value) {
        await updateDatasource(editId.value, payload as DatasourceForm)
        ElMessage.success('更新成功')
      } else {
        await createDatasource(payload as DatasourceForm)
        ElMessage.success('创建成功，已自动为您加密存储密码')
      }
      dialogVisible.value = false
      loadList()
    } catch (e: any) {
      ElMessage.error(e?.message || '保存失败')
    }
  })
}

async function testConnection() {
  testResult.value = null
  testing.value = true
  try {
    const payload: any = { ...form }
    if (payload.dbType === 'sqlite' && payload.memoryDB) {
      payload.filePath = ':memory:'
    }
    const res: any = await testConn(payload)
    const tr: TestResult = {
      success: !!res?.success,
      message: res?.message || '连接成功',
      latencyMs: res?.latencyMs,
      version: res?.version
    }
    testResult.value = tr
    ElMessage.success(tr.success ? `连接成功${tr.latencyMs ? '（' + tr.latencyMs + ' ms）' : ''}` : tr.message)
  } catch (e: any) {
    testResult.value = { success: false, message: e?.message || '连接失败' }
    ElMessage.error(e?.message || '连接测试失败')
  } finally {
    testing.value = false
  }
}

async function refreshConn(row: any) {
  refreshingConnId.value = row.datasourceId
  try {
    const res: any = await testConnectionById(row.datasourceId)
    const idx = list.value.findIndex((d: any) => d.datasourceId === row.datasourceId)
    if (idx >= 0) {
      list.value[idx].connStatus = res?.success ? 'ok' : 'fail'
      list.value[idx].lastConnTestAt = new Date().toLocaleString()
      if (res?.latencyMs != null) list.value[idx].connLatencyMs = res.latencyMs
      if (res?.version) list.value[idx].version = res.version
    }
    ElMessage.success(res?.success ? '连接成功' : (res?.message || '连接失败'))
  } catch (e: any) {
    const idx = list.value.findIndex((d: any) => d.datasourceId === row.datasourceId)
    if (idx >= 0) {
      list.value[idx].connStatus = 'fail'
      list.value[idx].lastConnTestAt = new Date().toLocaleString()
    }
    ElMessage.error(e?.message || '连接失败')
  } finally {
    refreshingConnId.value = ''
  }
}

function handleDelete(row: any) {
  ElMessageBox.confirm(
    `确认删除数据源「${row.name}」？删除后不可恢复，关联 SQL 历史保留但无法直接使用。`,
    '⚠️ 删除确认',
    {
      type: 'warning',
      confirmButtonText: '确认删除',
      cancelButtonText: '取消',
      distinguishCancelAndClose: true
    }
  ).then(async () => {
    try {
      await ElMessageBox.prompt(
        `请输入数据源名称的前 4 个字符以二次确认（${(row.name || '').slice(0, 4)}）`,
        '🔒 二次确认',
        {
          inputPattern: new RegExp('^' + (row.name || '').slice(0, 4).replace(/[.*+?^${}()|[\]\\]/g, '\\$&')),
          inputErrorMessage: '输入不匹配，请重新输入'
        }
      )
    } catch {
      return
    }
    try {
      await deleteDatasource(row.datasourceId)
      ElMessage.success('删除成功')
      loadList()
    } catch (e: any) {
      ElMessage.error(e?.message || '删除失败')
    }
  }).catch(() => {})
}

async function onMoreCommand(cmd: string, row: any) {
  switch (cmd) {
    case 'detail':
      goDetail(row)
      break
    case 'edit':
      openDialog(row)
      break
    case 'copy':
      try {
        await copyDatasource(row.datasourceId)
        ElMessage.success('已复制数据源，请在列表中修改名称后使用')
        loadList()
      } catch (e: any) {
        ElMessage.error(e?.message || '复制失败')
      }
      break
    case 'setDefault':
      dsStore.setDatasource(row.datasourceId, row.name)
      ElMessage.success('已设为默认数据源')
      break
    case 'resetPwd':
      pwdForm.datasourceId = row.datasourceId
      pwdForm.password = ''
      resetPwdVisible.value = true
      break
    case 'delete':
      handleDelete(row)
      break
  }
}

async function confirmResetPwd() {
  if (!pwdForm.password) {
    ElMessage.warning('请输入新密码')
    return
  }
  try {
    await updateDatasource(pwdForm.datasourceId, { password: pwdForm.password } as DatasourceForm)
    ElMessage.success('密码已更新')
    resetPwdVisible.value = false
  } catch (e: any) {
    ElMessage.error(e?.message || '更新失败')
  }
}

async function loadDatabases(row: any) {
  currentDs.value = row
  try {
    const dbs: any = await listDatabases(row.datasourceId)
    databases.value = dbs || []
    tablesMap.value = {}
    if (databases.value.length > 0) {
      activeDb.value = databases.value[0]
      loadTables()
    }
    dbDialogVisible.value = true
  } catch {}
}

async function loadTables() {
  if (!currentDs.value || !activeDb.value) return
  tablesLoading.value = true
  try {
    const t: any = await listTables(currentDs.value.datasourceId, activeDb.value)
    const arr = Array.isArray(t) ? t : []
    const tableNames = arr.map((item: any) => {
      if (typeof item === 'string') return item
      if (item && item.name) return item.name
      if (item && item.tableName) return item.tableName
      if (item && item.table_name) return item.table_name
      const keys = Object.keys(item)
      if (keys.length > 0) return item[keys[0]]
      return String(item)
    })
    tablesMap.value = { ...tablesMap.value, [activeDb.value]: tableNames }
  } finally {
    tablesLoading.value = false
  }
}

function goWorkbench(row: any) {
  dsStore.setDatasource(row.datasourceId, row.name)
  router.push('/sql/workbench')
}

function goDetail(row: any) {
  router.push('/datasources/' + row.datasourceId)
}

onMounted(() => {
  loadList()
  loadBusinessesAndServers()
})

watch(current, () => loadList())
</script>

<style scoped>
.text-info { color: #409eff; }
.text-muted { color: #c0c4cc; }
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
}
.title-row { display: flex; align-items: baseline; gap: 14px; }
.title-text { font-size: 20px; font-weight: 600; color: #303133; }
.subtitle-text { font-size: 13px; color: #909399; }
.header-actions { display: flex; gap: 8px; align-items: center; }
.toolbar {
  display: flex; gap: 12px; align-items: center; flex-wrap: wrap; padding: 0 4px 14px 0;
}
.toolbar-right { margin-left: auto; display: flex; gap: 8px; align-items: center; }
.form-section-title {
  width: 100%;
  padding: 6px 0 8px;
  font-size: 13px;
  font-weight: 600;
  color: #303133;
  border-left: 3px solid #409eff;
  padding-left: 10px;
  background: #f4f8ff;
  margin: 6px 0 8px 0;
}
.dialog-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 8px;
}
.test-result.ok { color: #67c23a; font-weight: 500; }
.test-result.fail { color: #f56c6c; font-weight: 500; }
.color-picker { display: flex; gap: 10px; align-items: center; padding: 4px 0; }
.color-option {
  width: 28px; height: 28px; border-radius: 50%; cursor: pointer;
  border: 2px solid transparent; transition: all .2s;
  display: flex; align-items: center; justify-content: center;
  color: #fff; font-size: 14px; font-weight: bold;
}
.color-option:hover { transform: scale(1.1); }
.color-option.active { border-color: #303133; box-shadow: 0 0 0 2px rgba(64,158,255,.18); }
.status-dot { display: inline-block; width: 10px; height: 10px; border-radius: 50%; }
.status-ok { background: #67c23a; box-shadow: 0 0 0 3px rgba(103,194,58,.18); }
.status-fail { background: #f56c6c; box-shadow: 0 0 0 3px rgba(245,108,108,.18); animation: blink 1.5s infinite; }
.status-none { background: #c0c4cc; }
@keyframes blink { 0%, 100% { opacity: 1; } 50% { opacity: .4; } }
.mono { font-family: Menlo, Monaco, Consolas, monospace; }
.color-bar {
  display: inline-block;
  width: 3px; height: 14px;
  margin-right: 8px;
  border-radius: 2px;
  vertical-align: middle;
}
.ds-name { font-weight: 500; }

.empty-state {
  padding: 60px 20px;
  text-align: center;
  color: #909399;
}
.empty-title { font-size: 18px; color: #606266; margin-top: 16px; font-weight: 500; }
.empty-desc { margin-top: 8px; font-size: 13px; color: #909399; }

.card-view {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 14px;
}
.card-item {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  overflow: hidden;
  transition: box-shadow .2s, transform .2s;
  background: #fff;
}
.card-item:hover {
  box-shadow: 0 4px 16px rgba(0,0,0,.08);
  transform: translateY(-1px);
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 14px;
  background: #fafbfc;
  border-bottom: 1px solid #ebeef5;
}
.card-title { display: flex; align-items: center; font-weight: 500; color: #303133; }
.card-body { padding: 12px 14px; }
.card-row {
  display: flex;
  align-items: flex-start;
  font-size: 13px;
  padding: 3px 0;
}
.card-row .label {
  width: 70px;
  color: #909399;
  flex-shrink: 0;
}
.card-row .value { flex: 1; color: #303133; word-break: break-all; }
.card-footer {
  padding: 10px 14px;
  border-top: 1px solid #ebeef5;
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  background: #fafbfc;
}
</style>
