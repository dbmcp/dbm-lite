<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DB老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <div class="page-container">
    <div class="page-header">
      <div class="page-title">介质维护</div>
      <div class="page-sub">
        管理 MySQL / TiDB / SQLite 等数据库版本安装包、补丁、客户端工具。
        <router-link to="/platform-config">返回平台配置</router-link>
      </div>
    </div>

    <el-card shadow="hover" :body-style="{ padding: '16px' }">
      <div style="display:flex;align-items:center;gap:10px;margin-bottom:12px;">
        <el-select v-model="filterType" placeholder="数据库类型" style="width:160px;" clearable>
          <el-option label="MySQL" value="MySQL" />
          <el-option label="TiDB" value="TiDB" />
          <el-option label="SQLite" value="SQLite" />
        </el-select>
        <el-input v-model="keyword" placeholder="搜索版本号" clearable style="width:240px;" />
        <el-button type="primary">上传介质</el-button>
        <ColumnToggle v-model="colVisible" :columns="columns" />
        <el-button @click="loadMedia">刷新</el-button>
      </div>

      <el-table :data="filteredMedia" border stripe>
        <el-table-column prop="name" label="介质名称" width="220" v-if="colVisible.name" />
        <el-table-column prop="type" label="类型" width="110" v-if="colVisible.type" />
        <el-table-column prop="version" label="版本号" width="140" v-if="colVisible.version" />
        <el-table-column prop="os" label="适用系统" width="140" v-if="colVisible.os" />
        <el-table-column prop="size" label="文件大小" width="120" v-if="colVisible.size" />
        <el-table-column prop="uploader" label="上传人" width="110" v-if="colVisible.uploader" />
        <el-table-column prop="uploadTime" label="上传时间" width="180" v-if="colVisible.uploadTime" />
        <el-table-column prop="status" label="状态" width="100" v-if="colVisible.status">
          <template #default="{ row }">
            <el-tag :type="row.status === '已发布' ? 'success' : 'info'" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" v-if="colVisible.actions">
          <template #default>
            <el-button link type="primary" size="small">下载</el-button>
            <el-button link type="primary" size="small">详情</el-button>
            <el-button link type="danger" size="small">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pager">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[100, 200, 500, 1000]"
          layout="total, sizes, prev, pager, next, jumper"
        />
      </div>
    </el-card>

    <el-card shadow="hover" :body-style="{ padding: '16px' }" style="margin-top:16px;">
      <template #header>版本使用分布</template>
      <el-table :data="usageList" border stripe>
        <el-table-column prop="type" label="数据库类型" width="140" />
        <el-table-column prop="version" label="版本" width="140" />
        <el-table-column prop="instances" label="使用实例数" width="140" />
        <el-table-column prop="percent" label="占比" width="240">
          <template #default="{ row }">
            <el-progress :percentage="Number(row.percent)" :stroke-width="10" />
          </template>
        </el-table-column>
        <el-table-column prop="latest" label="是否最新" width="120">
          <template #default="{ row }">
            <el-tag :type="row.latest ? 'success' : 'warning'" size="small">
              {{ row.latest ? '最新版本' : '可升级' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import ColumnToggle from '@/components/ColumnToggle.vue'
import { listMediums } from '@/api/ops'

const columns = [
  { key: 'name', label: '介质名称' },
  { key: 'type', label: '类型' },
  { key: 'version', label: '版本号' },
  { key: 'os', label: '适用系统' },
  { key: 'size', label: '文件大小' },
  { key: 'uploader', label: '上传人' },
  { key: 'uploadTime', label: '上传时间' },
  { key: 'status', label: '状态' },
  { key: 'actions', label: '操作' }
]
const colVisible = reactive<Record<string, boolean>>({
  name: true,
  type: true,
  version: true,
  os: true,
  size: true,
  uploader: true,
  uploadTime: true,
  status: true,
  actions: true
})
const page = ref(1)
const pageSize = ref(100)
const total = ref(0)

const filterType = ref('')
const keyword = ref('')
const mediaList = ref<any[]>([])
const usageList = ref<any[]>([])

async function loadMedia() {
  try {
    const res: any = await listMediums()
    if (res && res.success) {
      const data = res.data || {}
      mediaList.value = data.list || []
      usageList.value = data.usage || []
    } else {
      ElMessage.error(res?.message || '获取介质列表失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

const filteredMedia = computed(() => {
  const kw = keyword.value.trim().toLowerCase()
  return mediaList.value.filter((m) => {
    const typeMatch = !filterType.value || m.type === filterType.value
    const kwMatch = !kw || (m.version || '').toLowerCase().includes(kw) || (m.name || '').toLowerCase().includes(kw)
    return typeMatch && kwMatch
  })
})

onMounted(loadMedia)
</script>

<style scoped>
.page-container { padding: 16px 20px; }
.page-header { margin-bottom: 12px; }
.page-title { font-size: 20px; font-weight: 600; color: #303133; }
.page-sub { color: #909399; font-size: 13px; margin-top: 4px; }

.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
