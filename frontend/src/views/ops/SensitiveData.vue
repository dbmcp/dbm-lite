﻿<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DB老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <div class="page-container">
    <div class="page-header">
      <div class="page-title">敏感数据维护</div>
      <div class="page-sub">注册敏感字段，配置脱敏规则，维护敏感数据的访问权限和审计记录。</div>
    </div>

    <el-row :gutter="16">
      <el-col :span="18">
        <el-card shadow="hover" :body-style="{ padding: '16px' }">
          <template #header>
            <div style="display:flex;align-items:center;justify-content:space-between;">
              <span>敏感字段列表</span>
              <div>
                <el-input v-model="keyword" placeholder="搜索字段名 / 表名" clearable style="width:260px;margin-right:8px;" />
                <el-button type="primary" @click="showAdd = true">新增敏感字段</el-button>
              </div>
            </div>
          </template>

          <el-table :data="sensitiveList" border stripe>
            <el-table-column prop="datasourceId" label="数据源" width="160" />
            <el-table-column prop="table" label="表名" width="160" />
            <el-table-column prop="column" label="字段名" width="160" />
            <el-table-column prop="dataType" label="数据类型" width="110" />
            <el-table-column prop="level" label="敏感等级" width="110">
              <template #default="{ row }">
                <el-tag :type="levelTag(row.level)" size="small">{{ row.level }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="maskRule" label="脱敏规则" width="160" />
            <el-table-column label="操作" width="160">
              <template #default="{ row }">
                <el-button link type="primary" size="small">编辑</el-button>
                <el-button link type="primary" size="small">查看审计</el-button>
                <el-button link type="danger" size="small" @click="removeRow(row)">移除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" :body-style="{ padding: '16px' }">
          <template #header>敏感等级统计</template>
          <div v-for="item in stat" :key="item.level" class="stat-row">
            <span>{{ item.level }}</span>
            <el-progress :percentage="item.percent" :color="item.color" :stroke-width="12" style="flex:1;margin:0 10px;" />
            <span style="color:#606266;">{{ item.count }}</span>
          </div>
        </el-card>
        <el-card shadow="hover" :body-style="{ padding: '16px' }" style="margin-top:16px;">
          <template #header>脱敏规则提示</template>
          <ul class="tip-list">
            <li>手机号：保留前 3 位与后 4 位</li>
            <li>身份证号：保留前 4 位与后 4 位</li>
            <li>邮箱：保留首字母与域名</li>
            <li>金额：整数位随机扰动 + 保留小数</li>
          </ul>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="showAdd" title="新增敏感字段" width="520px;">
      <el-form :model="addForm" label-width="100px">
        <el-form-item label="数据源">
          <el-select v-model="addForm.datasourceId" style="width:100%;">
            <el-option v-for="ds in datasources" :key="ds.id" :label="ds.name" :value="ds.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="表名">
          <el-input v-model="addForm.table" />
        </el-form-item>
        <el-form-item label="字段名">
          <el-input v-model="addForm.column" />
        </el-form-item>
        <el-form-item label="数据类型">
          <el-input v-model="addForm.dataType" />
        </el-form-item>
        <el-form-item label="敏感等级">
          <el-radio-group v-model="addForm.level">
            <el-radio value="L1">L1 公开</el-radio>
            <el-radio value="L2">L2 内部</el-radio>
            <el-radio value="L3">L3 机密</el-radio>
            <el-radio value="L4">L4 绝密</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="脱敏规则">
          <el-select v-model="addForm.maskRule" style="width:100%;">
            <el-option label="全部遮蔽" value="全部遮蔽" />
            <el-option label="手机号脱敏" value="手机号脱敏" />
            <el-option label="身份证脱敏" value="身份证脱敏" />
            <el-option label="邮箱脱敏" value="邮箱脱敏" />
            <el-option label="金额扰动" value="金额扰动" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAdd = false">取消</el-button>
        <el-button type="primary" @click="confirmAdd">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, reactive, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { listSensitiveData, createSensitiveData, deleteSensitiveData } from '@/api/ops'

const datasources = ref([
  { id: 'ds-001', name: '生产-订单库' },
  { id: 'ds-002', name: '测试-商品库' },
  { id: 'ds-003', name: '本地-SQLite' }
])

const keyword = ref('')
const showAdd = ref(false)
const sensitiveList = ref<any[]>([])

const stat = computed(() => {
  const total = sensitiveList.value.length || 1
  const levels = ['L1', 'L2', 'L3', 'L4']
  const colors = ['#67c23a', '#909399', '#e6a23c', '#f56c6c']
  return levels.map((lv, idx) => {
    const count = sensitiveList.value.filter((r) => r.level === lv).length
    return { level: lv, count, percent: Math.round((count / total) * 100), color: colors[idx] }
  })
})

function levelTag(lv: string) {
  if (lv === 'L4') return 'danger'
  if (lv === 'L3') return 'warning'
  if (lv === 'L2') return 'info'
  return 'success'
}

const addForm = reactive({
  datasourceId: 'ds-001',
  table: '',
  column: '',
  dataType: '',
  level: 'L2',
  maskRule: '全部遮蔽',
  tables: ''
})

async function loadSensitive() {
  try {
    const res: any = await listSensitiveData({ keyword: keyword.value.trim() })
    if (res && res.success) {
      sensitiveList.value = res.data?.list || res.data || []
    } else {
      ElMessage.error(res?.message || '获取列表失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

async function confirmAdd() {
  if (!addForm.table || !addForm.column) {
    ElMessage.warning('请完善表名与字段名')
    return
  }
  try {
    const res: any = await createSensitiveData({ ...addForm })
    if (res && res.success) {
      ElMessage.success('已添加敏感字段')
      showAdd.value = false
      loadSensitive()
    } else {
      ElMessage.error(res?.message || '新增失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

async function removeRow(row: any) {
  try {
    const id = row.id || (row.table + '.' + row.column)
    const res: any = await deleteSensitiveData(id)
    if (res && res.success) {
      ElMessage.success('已移除')
      loadSensitive()
    } else {
      ElMessage.error(res?.message || '移除失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

onMounted(loadSensitive)
</script>

<style scoped>
.page-container { padding: 16px 20px; }
.page-header { margin-bottom: 12px; }
.page-title { font-size: 20px; font-weight: 600; color: #303133; }
.page-sub { color: #909399; font-size: 13px; margin-top: 4px; }
.stat-row { display: flex; align-items: center; padding: 6px 0; }
.tip-list { padding-left: 18px; color: #606266; line-height: 1.9; font-size: 13px; }
</style>
