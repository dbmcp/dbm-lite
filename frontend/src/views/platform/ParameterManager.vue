<template>
  <div class="page-container">
    <div class="page-header">
      <div class="page-title">参数配置</div>
      <div class="page-sub">维护系统全局参数，如默认值、开关和外部服务配置。</div>
    </div>

    <div class="toolbar">
      <el-input v-model="keyword" placeholder="搜索参数名称 / 键 / 值" clearable style="width:220px" />
      <el-select v-model="filterType" placeholder="类型" clearable style="width:120px">
        <el-option label="字符串" value="string" />
        <el-option label="整数" value="int" />
        <el-option label="布尔" value="bool" />
        <el-option label="JSON" value="json" />
      </el-select>
      <el-select v-model="filterStatus" placeholder="状态" clearable style="width:120px">
        <el-option label="启用" value="active" />
        <el-option label="禁用" value="inactive" />
      </el-select>
      <el-button type="primary" plain @click="onSearch">查询</el-button>
      <el-button type="success" @click="openCreateDialog">新增参数</el-button>
    </div>

    <el-table :data="list" border stripe v-loading="loading" style="width:100%">
      <el-table-column prop="name" label="参数名称" min-width="180" />
      <el-table-column prop="paramKey" label="参数键" min-width="200" />
      <el-table-column prop="paramType" label="类型" width="100" />
      <el-table-column prop="value" label="参数值" min-width="320" show-overflow-tooltip />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag size="small" type="success" v-if="row.status === 'active'">启用</el-tag>
          <el-tag size="small" type="info" v-else>禁用</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="remark" label="备注" min-width="200" />
      <el-table-column prop="updatedAt" label="更新时间" width="180" />
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button size="small" link type="primary" @click="openEditDialog(row)">编辑</el-button>
          <el-button size="small" link type="danger" @click="confirmDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pager">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="size"
        :total="total"
        :page-sizes="[100, 200, 500, 1000]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="onSearch"
        @current-change="onSearch"
      />
    </div>

    <el-dialog v-model="formVisible" :title="formMode === 'create' ? '新增参数' : '编辑参数'" width="560px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="参数名称" required>
          <el-input v-model="form.name" placeholder="如：登录有效期" />
        </el-form-item>
        <el-form-item label="参数键" required>
          <el-input v-model="form.paramKey" placeholder="如：system.login.expire" :disabled="formMode === 'edit'" />
        </el-form-item>
        <el-form-item label="参数类型">
          <el-select v-model="form.paramType" style="width:100%">
            <el-option label="字符串" value="string" />
            <el-option label="整数" value="int" />
            <el-option label="布尔" value="bool" />
            <el-option label="JSON" value="json" />
          </el-select>
        </el-form-item>
        <el-form-item label="参数值">
          <el-input v-model="form.value" type="textarea" :rows="3" placeholder="请输入参数值" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width:100%">
            <el-option label="启用" value="active" />
            <el-option label="禁用" value="inactive" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ferryListParameters,
  ferryCreateParameter,
  ferryUpdateParameter,
  ferryDeleteParameter
} from '@/api/ferry'

const keyword = ref('')
const filterStatus = ref('')
const filterType = ref('')
const loading = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const size = ref(100)

function onSearch() {
  page.value = 1
  loadList()
}

async function loadList() {
  loading.value = true
  try {
    const res: any = await ferryListParameters({
      keyword: keyword.value,
      status: filterStatus.value,
      type: filterType.value,
      page: page.value,
      size: size.value
    })
    const items: any[] = Array.isArray(res?.data) ? res.data : Array.isArray(res) ? res : []
    list.value = items.map((r: any) => ({
      id: r.id ?? r.paramId ?? '',
      name: r.name ?? '',
      paramKey: r.paramKey ?? r.key ?? '',
      paramType: r.paramType ?? r.type ?? 'string',
      value: r.value ?? '',
      status: r.status ?? 'active',
      remark: r.remark ?? '',
      updatedAt: r.updatedAt ?? ''
    }))
    total.value = typeof res?.total === 'number' ? res.total : list.value.length
  } catch (e: any) {
    ElMessage.error(e?.message || '加载失败')
  } finally {
    loading.value = false
  }
}

const formVisible = ref(false)
const submitting = ref(false)
const formMode = ref<'create' | 'edit'>('create')
const editId = ref('')
const form = reactive({
  name: '',
  paramKey: '',
  paramType: 'string',
  value: '',
  status: 'active',
  remark: ''
})

function resetForm() {
  form.name = ''
  form.paramKey = ''
  form.paramType = 'string'
  form.value = ''
  form.status = 'active'
  form.remark = ''
}

function openCreateDialog() {
  resetForm()
  formMode.value = 'create'
  formVisible.value = true
}

function openEditDialog(row: any) {
  resetForm()
  formMode.value = 'edit'
  editId.value = row.id
  form.name = row.name
  form.paramKey = row.paramKey
  form.paramType = row.paramType || 'string'
  form.value = row.value
  form.status = row.status || 'active'
  form.remark = row.remark || ''
  formVisible.value = true
}

async function submitForm() {
  if (!form.name.trim() || !form.paramKey.trim()) {
    ElMessage.warning('请填写参数名称和参数键')
    return
  }
  submitting.value = true
  try {
    const payload = {
      name: form.name.trim(),
      key: form.paramKey.trim(),
      paramKey: form.paramKey.trim(),
      paramType: form.paramType,
      value: form.value,
      status: form.status,
      remark: form.remark
    }
    if (formMode.value === 'create') {
      await ferryCreateParameter(payload)
    } else {
      await ferryUpdateParameter(editId.value, payload)
    }
    ElMessage.success('保存成功')
    formVisible.value = false
    loadList()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  } finally {
    submitting.value = false
  }
}

function confirmDelete(row: any) {
  ElMessageBox.confirm(`确认删除参数 "${row.name}" 吗？`, '提示', { type: 'warning' })
    .then(async () => {
      try {
        await ferryDeleteParameter(row.id)
        ElMessage.success('已删除')
        loadList()
      } catch (e: any) {
        ElMessage.error(e?.message || '删除失败')
      }
    })
    .catch(() => {})
}

onMounted(loadList)
</script>

<style scoped>
.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}
.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 12px;
}
</style>
