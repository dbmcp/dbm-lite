<template>
  <div class="priv-page">
    <div class="page-header">
      <h2 class="title">敏感列配置</h2>
      <div class="actions">
        <el-button type="primary" @click="openCreate">
          <el-icon><Plus /></el-icon>
          新增配置
        </el-button>
      </div>
    </div>

    <el-table :data="columns" border stripe v-loading="loading">
      <el-table-column label="数据源" prop="datasourceId" min-width="160" />
      <el-table-column label="数据库" prop="databaseName" min-width="140" />
      <el-table-column label="数据表" prop="tableName" min-width="140" />
      <el-table-column label="列名" prop="columnName" min-width="140" />
      <el-table-column label="脱敏规则" prop="rule" width="140">
        <template #default="scope">
          <el-tag size="small">{{ ruleText(scope.row.rule) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" prop="createdAt" width="180" />
      <el-table-column label="操作" width="100">
        <template #default="scope">
          <el-button size="small" link type="danger" @click="remove(scope.row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 新增弹窗 -->
    <el-dialog v-model="dialogVisible" title="新增敏感列配置" width="520px">
      <el-form :model="form" label-width="96px">
        <el-form-item label="数据源">
          <el-select v-model="form.datasourceId" placeholder="选择数据源" style="width: 100%" @change="onDsChange">
            <el-option v-for="d in datasources" :key="d.datasourceId" :label="d.name" :value="d.datasourceId" />
          </el-select>
        </el-form-item>
        <el-form-item label="数据库">
          <el-select v-model="form.databaseName" placeholder="选择数据库" style="width: 100%" @change="onDbChange">
            <el-option v-for="d in databases" :key="d" :label="d" :value="d" />
          </el-select>
        </el-form-item>
        <el-form-item label="数据表">
          <el-select v-model="form.tableName" placeholder="选择表" style="width: 100%" @change="onTableChange">
            <el-option v-for="t in tables" :key="t" :label="t" :value="t" />
          </el-select>
        </el-form-item>
        <el-form-item label="列名">
          <el-select v-model="form.columnName" placeholder="选择列" filterable style="width: 100%">
            <el-option v-for="c in columns" :key="c" :label="c" :value="c" />
          </el-select>
        </el-form-item>
        <el-form-item label="脱敏规则">
          <el-select v-model="form.rule" style="width: 100%">
            <el-option label="打码屏蔽（mask）" value="mask" />
            <el-option label="邮箱脱敏（email）" value="email" />
            <el-option label="手机号脱敏（phone）" value="phone" />
            <el-option label="完全隐藏（hide）" value="hide" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { listSensitiveColumns, createSensitiveColumn, deleteSensitiveColumn } from '@/api/priv'
import { listAllDatasources } from '@/api/datasource'

const loading = ref(false)
const columns = ref<any[]>([])
const datasources = ref<any[]>([])
const databases = ref<string[]>([])
const tables = ref<string[]>([])
const columnOptions = ref<string[]>([])
const dialogVisible = ref(false)
const form = reactive({
  datasourceId: '',
  databaseName: '',
  tableName: '',
  columnName: '',
  rule: 'mask'
})

function ruleText(r: string) {
  const map: Record<string, string> = { mask: '打码屏蔽', email: '邮箱脱敏', phone: '手机号脱敏', hide: '完全隐藏' }
  return map[r] || r
}

async function load() {
  try {
    loading.value = true
    const res: any = await listSensitiveColumns()
    columns.value = Array.isArray(res) ? res : res?.list || []
  } catch (e) {
    columns.value = []
  } finally {
    loading.value = false
  }
}

async function loadDS() {
  try {
    const res: any = await listAllDatasources(1, 1000)
    datasources.value = res?.list || []
  } catch (e) {}
}

async function onDsChange() {
  databases.value = []
  tables.value = []
  columnOptions.value = []
  if (!form.datasourceId) return
  try {
    const { getDatabases } = await import('@/api/sql')
    databases.value = (await getDatabases(form.datasourceId)) || []
  } catch (e) {}
}

async function onDbChange() {
  tables.value = []
  columnOptions.value = []
  if (!form.datasourceId || !form.databaseName) return
  try {
    const { listTables } = await import('@/api/sql')
    const raw = await listTables(form.datasourceId, form.databaseName)
    tables.value = Array.isArray(raw) ? raw : (raw as any)?.list || []
  } catch (e) {}
}

async function onTableChange() {
  columnOptions.value = []
  if (!form.datasourceId || !form.databaseName || !form.tableName) return
  try {
    const { listColumns } = await import('@/api/sql')
    const raw = await listColumns(form.datasourceId, form.databaseName, form.tableName)
    const arr = Array.isArray(raw) ? raw : (raw as any)?.list || []
    columnOptions.value = arr.map((x: any) => x.column || x.name || x)
  } catch (e) {}
}

function openCreate() {
  form.datasourceId = ''
  form.databaseName = ''
  form.tableName = ''
  form.columnName = ''
  form.rule = 'mask'
  dialogVisible.value = true
}

async function submit() {
  if (!form.datasourceId || !form.databaseName || !form.tableName || !form.columnName) {
    ElMessage.warning('请完整填写数据源/库/表/列')
    return
  }
  try {
    await createSensitiveColumn(form)
    ElMessage.success('保存成功')
    dialogVisible.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  }
}

async function remove(row: any) {
  try {
    await ElMessageBox.confirm('确认删除该敏感列配置？', '提示', { type: 'warning' })
    await deleteSensitiveColumn(row.id)
    ElMessage.success('删除成功')
    load()
  } catch (e) {}
}

onMounted(() => {
  load()
  loadDS()
})
</script>

<style scoped>
.priv-page {
  padding: 16px 20px;
  background: #f5f7fa;
  min-height: calc(100vh - 60px);
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
  padding: 16px 20px;
  border-radius: 8px;
  margin-bottom: 12px;
}
.title {
  margin: 0;
  font-size: 18px;
  color: #303133;
}
</style>
