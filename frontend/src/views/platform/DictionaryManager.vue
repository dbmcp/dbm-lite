<template>
  <div class="page-container">
    <div class="page-header">
      <div class="page-title">字典管理</div>
      <div class="page-sub">维护数据字典类型及其字典项。</div>
    </div>

    <div class="toolbar">
      <el-input v-model="keyword" placeholder="搜索字典名称 / 类型" clearable style="width:220px" />
      <el-select v-model="statusFilter" placeholder="状态" clearable style="width:120px">
        <el-option label="启用" value="active" />
        <el-option label="禁用" value="inactive" />
      </el-select>
      <el-button type="primary" plain @click="onSearch">查询</el-button>
      <el-button type="success" @click="openDictDialog(null)">新建字典</el-button>
    </div>

    <el-table :data="list" border stripe v-loading="loading" style="width:100%">
      <el-table-column prop="name" label="字典名称" min-width="200" />
      <el-table-column prop="type" label="类型标识" min-width="220" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag size="small" type="success" v-if="row.status === 'active'">启用</el-tag>
          <el-tag size="small" type="info" v-else>禁用</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="remark" label="备注" min-width="200" />
      <el-table-column prop="createdAt" label="创建时间" width="180" />
      <el-table-column label="操作" width="260" fixed="right">
        <template #default="{ row }">
          <el-button size="small" link type="primary" @click="openItemsDialog(row)">字典项</el-button>
          <el-button size="small" link type="primary" @click="openDictDialog(row)">编辑</el-button>
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

    <el-dialog v-model="dictFormVisible" :title="dictFormMode === 'create' ? '新建字典' : '编辑字典'" width="480px">
      <el-form :model="dictForm" label-width="90px">
        <el-form-item label="字典名称" required>
          <el-input v-model="dictForm.name" placeholder="如：用户状态" />
        </el-form-item>
        <el-form-item label="类型标识" required>
          <el-input v-model="dictForm.type" placeholder="如：user_status" :disabled="dictFormMode === 'edit'" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="dictForm.status" style="width:100%">
            <el-option label="启用" value="active" />
            <el-option label="禁用" value="inactive" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="dictForm.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dictFormVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitDictForm">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="itemsVisible" :title="'字典项维护：' + currentDictName" width="720px">
      <div class="toolbar" style="margin-bottom:10px">
        <el-input v-model="itemKeyword" placeholder="搜索字典项 / 值" clearable style="width:220px" />
        <el-button type="primary" plain @click="loadItems">查询</el-button>
        <el-button type="success" @click="openItemDialog(null)">新建字典项</el-button>
      </div>
      <el-table :data="itemList" border stripe v-loading="itemsLoading" style="width:100%">
        <el-table-column prop="label" label="标签" min-width="140" />
        <el-table-column prop="value" label="值" min-width="160" />
        <el-table-column prop="sortOrder" label="排序" width="80" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag size="small" type="success" v-if="row.status === 'active'">启用</el-tag>
            <el-tag size="small" type="info" v-else>禁用</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="160" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" link type="primary" @click="openItemDialog(row)">编辑</el-button>
            <el-button size="small" link type="danger" @click="confirmDeleteItem(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-dialog v-model="itemFormVisible" :title="itemFormMode === 'create' ? '新建字典项' : '编辑字典项'" width="480px" append-to-body>
        <el-form :model="itemForm" label-width="90px">
          <el-form-item label="标签" required>
            <el-input v-model="itemForm.label" />
          </el-form-item>
          <el-form-item label="值" required>
            <el-input v-model="itemForm.value" />
          </el-form-item>
          <el-form-item label="排序">
            <el-input-number v-model="itemForm.sortOrder" :min="0" :max="9999" />
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="itemForm.status" style="width:100%">
              <el-option label="启用" value="active" />
              <el-option label="禁用" value="inactive" />
            </el-select>
          </el-form-item>
          <el-form-item label="备注">
            <el-input v-model="itemForm.remark" type="textarea" :rows="2" />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="itemFormVisible = false">取消</el-button>
          <el-button type="primary" :loading="submitting" @click="submitItemForm">保存</el-button>
        </template>
      </el-dialog>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ferryListDictionaries,
  ferryCreateDictionary,
  ferryUpdateDictionary,
  ferryDeleteDictionary,
  ferryListDictItems,
  ferryCreateDictItem,
  ferryUpdateDictItem,
  ferryDeleteDictItem
} from '@/api/ferry'

const keyword = ref('')
const statusFilter = ref('')
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
    const res: any = await ferryListDictionaries({
      keyword: keyword.value,
      status: statusFilter.value,
      page: page.value,
      size: size.value
    })
    const items: any[] = Array.isArray(res?.data) ? res.data : Array.isArray(res) ? res : []
    list.value = items.map((r: any) => ({
      id: r.id ?? r.dictId ?? '',
      name: r.name ?? '',
      type: r.type ?? '',
      status: r.status ?? 'active',
      remark: r.remark ?? '',
      createdAt: r.createdAt ?? ''
    }))
    total.value = typeof res?.total === 'number' ? res.total : list.value.length
  } catch (e: any) {
    ElMessage.error(e?.message || '加载失败')
  } finally {
    loading.value = false
  }
}

const dictFormVisible = ref(false)
const submitting = ref(false)
const dictFormMode = ref<'create' | 'edit'>('create')
const editDictId = ref('')
const dictForm = reactive({ name: '', type: '', status: 'active', remark: '' })

function openDictDialog(row: any) {
  if (row) {
    dictFormMode.value = 'edit'
    editDictId.value = row.id
    dictForm.name = row.name
    dictForm.type = row.type
    dictForm.status = row.status || 'active'
    dictForm.remark = row.remark || ''
  } else {
    dictFormMode.value = 'create'
    editDictId.value = ''
    dictForm.name = ''
    dictForm.type = ''
    dictForm.status = 'active'
    dictForm.remark = ''
  }
  dictFormVisible.value = true
}

async function submitDictForm() {
  if (!dictForm.name.trim() || !dictForm.type.trim()) {
    ElMessage.warning('请填写字典名称和类型标识')
    return
  }
  submitting.value = true
  try {
    const payload = { name: dictForm.name.trim(), type: dictForm.type.trim(), status: dictForm.status, remark: dictForm.remark }
    if (dictFormMode.value === 'create') {
      await ferryCreateDictionary(payload)
    } else {
      await ferryUpdateDictionary(editDictId.value, payload)
    }
    ElMessage.success('保存成功')
    dictFormVisible.value = false
    loadList()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  } finally {
    submitting.value = false
  }
}

function confirmDelete(row: any) {
  ElMessageBox.confirm(`确认删除字典 "${row.name}" 吗？相关字典项会一起被清理。`, '提示', { type: 'warning' })
    .then(async () => {
      try {
        await ferryDeleteDictionary(row.id)
        ElMessage.success('已删除')
        loadList()
      } catch (e: any) {
        ElMessage.error(e?.message || '删除失败')
      }
    })
    .catch(() => {})
}

const itemsVisible = ref(false)
const itemsLoading = ref(false)
const itemList = ref<any[]>([])
const itemKeyword = ref('')
const currentDictType = ref('')
const currentDictName = ref('')

async function openItemsDialog(row: any) {
  currentDictType.value = row.type
  currentDictName.value = row.name
  itemsVisible.value = true
  itemKeyword.value = ''
  loadItems()
}

async function loadItems() {
  itemsLoading.value = true
  try {
    const res: any = await ferryListDictItems({ dictType: currentDictType.value, keyword: itemKeyword.value })
    const items: any[] = Array.isArray(res?.data) ? res.data : Array.isArray(res) ? res : []
    itemList.value = items.map((r: any) => ({
      id: r.id ?? r.itemId ?? '',
      label: r.label ?? '',
      value: r.value ?? '',
      sortOrder: typeof r.sortOrder === 'number' ? r.sortOrder : 0,
      status: r.status ?? 'active',
      remark: r.remark ?? ''
    }))
  } catch (e: any) {
    ElMessage.error(e?.message || '加载失败')
  } finally {
    itemsLoading.value = false
  }
}

const itemFormVisible = ref(false)
const itemFormMode = ref<'create' | 'edit'>('create')
const editItemId = ref('')
const itemForm = reactive({ label: '', value: '', sortOrder: 0, status: 'active', remark: '' })

function openItemDialog(row: any) {
  if (row) {
    itemFormMode.value = 'edit'
    editItemId.value = row.id
    itemForm.label = row.label
    itemForm.value = row.value
    itemForm.sortOrder = row.sortOrder ?? 0
    itemForm.status = row.status || 'active'
    itemForm.remark = row.remark || ''
  } else {
    itemFormMode.value = 'create'
    editItemId.value = ''
    itemForm.label = ''
    itemForm.value = ''
    itemForm.sortOrder = 0
    itemForm.status = 'active'
    itemForm.remark = ''
  }
  itemFormVisible.value = true
}

async function submitItemForm() {
  if (!itemForm.label.trim() || !itemForm.value.trim()) {
    ElMessage.warning('请填写标签和值')
    return
  }
  submitting.value = true
  try {
    const payload = { dictType: currentDictType.value, label: itemForm.label.trim(), value: itemForm.value.trim(), sortOrder: itemForm.sortOrder, status: itemForm.status, remark: itemForm.remark }
    if (itemFormMode.value === 'create') {
      await ferryCreateDictItem(payload)
    } else {
      await ferryUpdateDictItem(editItemId.value, payload)
    }
    ElMessage.success('保存成功')
    itemFormVisible.value = false
    loadItems()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  } finally {
    submitting.value = false
  }
}

function confirmDeleteItem(row: any) {
  ElMessageBox.confirm(`确认删除字典项 "${row.label}" 吗？`, '提示', { type: 'warning' })
    .then(async () => {
      try {
        await ferryDeleteDictItem(row.id)
        ElMessage.success('已删除')
        loadItems()
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
