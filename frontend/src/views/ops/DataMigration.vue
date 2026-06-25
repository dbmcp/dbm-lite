<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DB老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <div class="page-container">
    <div class="page-header">
      <div class="page-title">数据迁移</div>
      <div class="page-sub">支持结构对比、数据对比、全量 / 增量迁移任务。</div>
    </div>

    <el-tabs v-model="activeTab" type="border-card">
      <el-tab-pane label="迁移" name="migration">
        <el-card shadow="hover" :body-style="{ padding: '16px' }">
          <el-form :model="migrateForm" label-width="120px" inline>
            <el-form-item label="源端数据源">
              <el-select v-model="migrateForm.source" style="width:220px;">
                <el-option v-for="d in datasources" :key="d" :label="d" :value="d" />
              </el-select>
            </el-form-item>
            <el-form-item label="目标端数据源">
              <el-select v-model="migrateForm.target" style="width:220px;">
                <el-option v-for="d in datasources" :key="d" :label="d" :value="d" />
              </el-select>
            </el-form-item>
            <el-form-item label="迁移模式">
              <el-radio-group v-model="migrateForm.mode">
                <el-radio value="full">全量</el-radio>
                <el-radio value="inc">增量</el-radio>
                <el-radio value="full+inc">全量+增量</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="startMigrate">启动迁移</el-button>
              <el-button @click="loadMigrateList">刷新</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card shadow="hover" :body-style="{ padding: '16px' }" style="margin-top:16px;">
          <template #header>
            <div style="display:flex;align-items:center;justify-content:space-between;">
              <span>迁移任务列表</span>
              <ColumnToggle v-model="migrateColVisible" :columns="migrateColumns" />
            </div>
          </template>
          <el-table :data="migrateList" border stripe>
            <el-table-column prop="id" label="任务ID" width="140" v-if="migrateColVisible.id" />
            <el-table-column prop="source" label="源端" width="180" v-if="migrateColVisible.source" />
            <el-table-column prop="target" label="目标端" width="180" v-if="migrateColVisible.target" />
            <el-table-column prop="mode" label="模式" width="100" v-if="migrateColVisible.mode" />
            <el-table-column prop="progress" label="进度" width="200" v-if="migrateColVisible.progress">
              <template #default="{ row }">
                <el-progress :percentage="Number(row.progress)" :stroke-width="10" />
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100" v-if="migrateColVisible.status">
              <template #default="{ row }">
                <el-tag :type="row.status === '完成' ? 'success' : row.status === '运行中' ? 'warning' : 'info'" size="small">
                  {{ row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="createTime" label="创建时间" width="180" v-if="migrateColVisible.createTime" />
            <el-table-column label="操作" width="120" v-if="migrateColVisible.action">
              <template #default>
                <el-button link type="primary" size="small">查看详情</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-pagination
            v-model:current-page="migratePage"
            v-model:page-size="migratePageSize"
            :page-sizes="[10, 30, 50, 100]"
            :total="migrateTotal"
            layout="total, sizes, prev, pager, next, jumper"
            style="margin-top:16px;display:flex;justify-content:flex-end;"
          />
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="结构对比" name="schema">
        <el-row :gutter="16">
          <el-col :span="8">
            <el-card shadow="hover" :body-style="{ padding: '16px' }">
              <el-form label-width="90px">
                <el-form-item label="源端">
                  <el-select v-model="schemaForm.source" style="width:100%;">
                    <el-option v-for="d in datasources" :key="d" :label="d" :value="d" />
                  </el-select>
                </el-form-item>
                <el-form-item label="目标端">
                  <el-select v-model="schemaForm.target" style="width:100%;">
                    <el-option v-for="d in datasources" :key="d" :label="d" :value="d" />
                  </el-select>
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" style="width:100%;" @click="compareSchema">开始结构对比</el-button>
                </el-form-item>
              </el-form>
            </el-card>
          </el-col>
          <el-col :span="16">
            <el-card shadow="hover" :body-style="{ padding: '16px' }">
              <template #header>
                <div style="display:flex;align-items:center;justify-content:space-between;">
                  <span>结构对比结果</span>
                  <ColumnToggle v-model="schemaColVisible" :columns="schemaColumns" />
                </div>
              </template>
              <el-table :data="schemaDiff" border stripe>
                <el-table-column prop="table" label="表名" width="180" v-if="schemaColVisible.table" />
                <el-table-column prop="type" label="差异类型" width="120" v-if="schemaColVisible.type">
                  <template #default="{ row }">
                    <el-tag :type="row.type === '新增' ? 'success' : row.type === '缺失' ? 'danger' : 'warning'" size="small">
                      {{ row.type }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="detail" label="差异详情" show-overflow-tooltip v-if="schemaColVisible.detail" />
              </el-table>
              <el-pagination
                v-model:current-page="schemaPage"
                v-model:page-size="schemaPageSize"
                :page-sizes="[10, 30, 50, 100]"
                :total="schemaTotal"
                layout="total, sizes, prev, pager, next, jumper"
                style="margin-top:16px;display:flex;justify-content:flex-end;"
              />
            </el-card>
          </el-col>
        </el-row>
      </el-tab-pane>

      <el-tab-pane label="数据对比" name="data">
        <el-card shadow="hover" :body-style="{ padding: '16px' }">
          <el-form :model="dataForm" label-width="120px" inline>
            <el-form-item label="源端">
              <el-select v-model="dataForm.source" style="width:220px;">
                <el-option v-for="d in datasources" :key="d" :label="d" :value="d" />
              </el-select>
            </el-form-item>
            <el-form-item label="目标端">
              <el-select v-model="dataForm.target" style="width:220px;">
                <el-option v-for="d in datasources" :key="d" :label="d" :value="d" />
              </el-select>
            </el-form-item>
            <el-form-item label="表名">
              <el-input v-model="dataForm.tables" placeholder="例：orders, order_items" style="width:260px;" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="compareData">开始数据比对</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card shadow="hover" :body-style="{ padding: '16px' }" style="margin-top:16px;">
          <template #header>
            <div style="display:flex;align-items:center;justify-content:space-between;">
              <span>数据对比结果</span>
              <ColumnToggle v-model="dataColVisible" :columns="dataColumns" />
            </div>
          </template>
          <el-table :data="dataDiff" border stripe>
            <el-table-column prop="table" label="表名" width="180" v-if="dataColVisible.table" />
            <el-table-column prop="sourceCount" label="源端行数" width="130" v-if="dataColVisible.sourceCount" />
            <el-table-column prop="targetCount" label="目标端行数" width="130" v-if="dataColVisible.targetCount" />
            <el-table-column prop="diff" label="差异行数" width="110" v-if="dataColVisible.diff">
              <template #default="{ row }">
                <el-tag :type="Number(row.diff) > 0 ? 'warning' : 'success'" size="small">{{ row.diff }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="percent" label="一致性" width="200" v-if="dataColVisible.percent">
              <template #default="{ row }">
                <el-progress :percentage="Number(row.percent)" :stroke-width="10" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120" v-if="dataColVisible.action">
              <template #default>
                <el-button link type="primary" size="small">导出差异</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-pagination
            v-model:current-page="dataPage"
            v-model:page-size="dataPageSize"
            :page-sizes="[10, 30, 50, 100]"
            :total="dataTotal"
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
import { listMigrationTasks, createMigrationTask, getSchemaDiff, getDataDiff } from '@/api/ops'

const activeTab = ref('migration')
const datasources = ref(['生产-订单库 (MySQL)', '测试-商品库 (TiDB)', '本地-SQLite', '归档库 (MySQL)'])

const migrateForm = reactive({
  source: '生产-订单库 (MySQL)',
  target: '测试-商品库 (TiDB)',
  mode: 'full'
})

const migrateColumns = [
  { key: 'id', label: '任务ID' },
  { key: 'source', label: '源端' },
  { key: 'target', label: '目标端' },
  { key: 'mode', label: '模式' },
  { key: 'progress', label: '进度' },
  { key: 'status', label: '状态' },
  { key: 'createTime', label: '创建时间' },
  { key: 'action', label: '操作' }
]
const migrateColVisible = reactive<Record<string, boolean>>({
  id: true, source: true, target: true, mode: true, progress: true, status: true, createTime: true, action: true
})
const migrateList = ref<any[]>([])
const migratePage = ref(1)
const migratePageSize = ref(50)
const migrateTotal = ref(0)

const schemaForm = reactive({ source: '生产-订单库 (MySQL)', target: '测试-商品库 (TiDB)' })
const schemaColumns = [
  { key: 'table', label: '表名' },
  { key: 'type', label: '差异类型' },
  { key: 'detail', label: '差异详情' }
]
const schemaColVisible = reactive<Record<string, boolean>>({ table: true, type: true, detail: true })
const schemaDiff = ref<any[]>([])
const schemaPage = ref(1)
const schemaPageSize = ref(50)
const schemaTotal = ref(0)

const dataForm = reactive({
  source: '生产-订单库 (MySQL)',
  target: '测试-商品库 (TiDB)',
  tables: 'orders, order_items, users'
})
const dataColumns = [
  { key: 'table', label: '表名' },
  { key: 'sourceCount', label: '源端行数' },
  { key: 'targetCount', label: '目标端行数' },
  { key: 'diff', label: '差异行数' },
  { key: 'percent', label: '一致性' },
  { key: 'action', label: '操作' }
]
const dataColVisible = reactive<Record<string, boolean>>({
  table: true, sourceCount: true, targetCount: true, diff: true, percent: true, action: true
})
const dataDiff = ref<any[]>([])
const dataPage = ref(1)
const dataPageSize = ref(50)
const dataTotal = ref(0)

async function loadMigrateList() {
  try {
    const res: any = await listMigrationTasks()
    if (res && res.success) {
      migrateList.value = res.data?.list || res.data || []
      migrateTotal.value = res.data?.total || migrateList.value.length
    } else {
      ElMessage.error(res?.message || '获取迁移任务失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

async function startMigrate() {
  if (migrateForm.source === migrateForm.target) {
    ElMessage.warning('源端和目标端不能相同')
    return
  }
  try {
    const res: any = await createMigrationTask({
      source: migrateForm.source,
      target: migrateForm.target,
      mode: migrateForm.mode,
      tables: ''
    })
    if (res && res.success) {
      ElMessage.success('迁移任务已启动')
      loadMigrateList()
    } else {
      ElMessage.error(res?.message || '启动失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

async function compareSchema() {
  try {
    const res: any = await getSchemaDiff()
    if (res && res.success) {
      schemaDiff.value = res.data?.list || res.data || []
      schemaTotal.value = res.data?.total || schemaDiff.value.length
      ElMessage.success('结构对比完成')
    } else {
      ElMessage.error(res?.message || '结构对比失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

async function compareData() {
  if (!dataForm.tables.trim()) {
    ElMessage.warning('请填写表名')
    return
  }
  try {
    const res: any = await getDataDiff()
    if (res && res.success) {
      dataDiff.value = res.data?.list || res.data || []
      dataTotal.value = res.data?.total || dataDiff.value.length
      ElMessage.success('数据对比完成')
    } else {
      ElMessage.error(res?.message || '数据对比失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

onMounted(loadMigrateList)
</script>

<style scoped>
.page-container { padding: 16px 20px; }
.page-header { margin-bottom: 12px; }
.page-title { font-size: 20px; font-weight: 600; color: #303133; }
.page-sub { color: #909399; font-size: 13px; margin-top: 4px; }
</style>
