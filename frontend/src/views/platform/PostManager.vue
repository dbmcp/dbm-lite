<template>
  <div class="page-container">
    <div class="page-header">
      <div class="page-title">岗位管理</div>
      <div class="page-sub">维护组织内的岗位定义，可用于后续用户画像与授权范围控制。</div>
    </div>

    <div class="toolbar">
      <el-input v-model="keyword" placeholder="搜索岗位名称 / 描述" clearable style="width:220px" />
      <el-select v-model="statusFilter" placeholder="状态" clearable style="width:120px">
        <el-option label="启用" value="active" />
        <el-option label="禁用" value="inactive" />
      </el-select>
      <el-button type="primary" plain @click="onSearch">查询</el-button>
      <el-button type="success" @click="openCreateDialog">新建岗位</el-button>
    </div>

    <el-table :data="pageList" border stripe v-loading="loading" style="width:100%">
      <el-table-column prop="sortOrder" label="排序" width="80" />
      <el-table-column prop="name" label="岗位名称" min-width="200" />
      <el-table-column prop="description" label="描述" min-width="260" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.status === 'active'" type="success">启用</el-tag>
          <el-tag v-else type="info">禁用</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="创建时间" width="180" />
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" link @click="openEditDialog(row)">编辑</el-button>
          <el-button size="small" type="danger" link @click="confirmDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pager">
      <el-pagination
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="total"
        :current-page="page"
        :page-size="size"
        :page-sizes="[100, 200, 500, 1000]"
        @current-change="onPageChange"
        @size-change="onSizeChange"
      />
    </div>

    <el-dialog v-model="formVisible" :title="formMode === 'create' ? '新建岗位' : '编辑岗位'" width="520px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="岗位名称" required>
          <el-input v-model="form.name" placeholder="如 数据库管理员" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sortOrder" :min="0" :max="9999" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width:100%">
            <el-option label="启用" value="active" />
            <el-option label="禁用" value="inactive" />
          </el-select>
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
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ferryListPosts, ferryCreatePost, ferryUpdatePost, ferryDeletePost } from '@/api/ferry'

const keyword = ref('')
const statusFilter = ref('')
const loading = ref(false)
const allItems = ref<any[]>([])
const page = ref(1)
const size = ref(100)

const filteredList = computed(() => {
  let items = allItems.value
  if (keyword.value) {
    const kw = keyword.value.toLowerCase()
    items = items.filter((r: any) =>
      (r.name || '').toLowerCase().includes(kw) ||
      (r.description || '').toLowerCase().includes(kw)
    )
  }
  if (statusFilter.value) {
    items = items.filter((r: any) => r.status === statusFilter.value)
  }
  return items
})
const total = computed(() => filteredList.value.length)
const pageList = computed(() => {
  const start = (page.value - 1) * size.value
  return filteredList.value.slice(start, start + size.value)
})

function onSearch() {
  page.value = 1
  loadList()
}
function onPageChange(p: number) {
  page.value = p
}
function onSizeChange(s: number) {
  size.value = s
  page.value = 1
}

async function loadList() {
  loading.value = true
  try {
    const res: any = await ferryListPosts(keyword.value)
    const items: any[] = Array.isArray(res) ? res : Array.isArray(res?.data) ? res.data : []
    allItems.value = items.map((r: any) => ({
      id: r.id ?? r.postId ?? '',
      name: r.name ?? '',
      description: r.description ?? '',
      sortOrder: typeof r.sortOrder === 'number' ? r.sortOrder : 0,
      status: r.status ?? 'active',
      createdAt: r.createdAt ?? ''
    }))
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
  description: '',
  sortOrder: 0,
  status: 'active'
})

function resetForm() {
  form.name = ''
  form.description = ''
  form.sortOrder = 0
  form.status = 'active'
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
  form.description = row.description || ''
  form.sortOrder = row.sortOrder ?? 0
  form.status = row.status || 'active'
  formVisible.value = true
}

async function submitForm() {
  if (!form.name.trim()) {
    ElMessage.warning('请输入岗位名称')
    return
  }
  submitting.value = true
  try {
    if (formMode.value === 'create') {
      await ferryCreatePost({
        name: form.name.trim(),
        description: form.description,
        sortOrder: form.sortOrder,
        status: form.status
      })
    } else {
      await ferryUpdatePost(editId.value, {
        name: form.name.trim(),
        description: form.description,
        sortOrder: form.sortOrder,
        status: form.status
      })
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
  ElMessageBox.confirm(`确认删除岗位 "${row.name}" 吗？`, '提示', { type: 'warning' })
    .then(async () => {
      try {
        await ferryDeletePost(row.id)
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
