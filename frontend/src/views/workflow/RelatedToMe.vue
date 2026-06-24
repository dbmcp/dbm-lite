<template>
  <div class="page-container">
    <div class="page-header">
      <div class="page-title">与我相关</div>
      <div class="page-sub">聚合展示与我相关的工单动态、系统通知、流程变更更新。</div>
    </div>

    <div class="stat-row">
      <el-card shadow="never" class="stat-card">
        <div class="stat-label">全部动态</div>
        <div class="stat-value" style="color: #409eff">{{ total }}</div>
      </el-card>
      <el-card shadow="never" class="stat-card">
        <div class="stat-label">未读</div>
        <div class="stat-value" style="color: #f56c6c">{{ unreadCount }}</div>
      </el-card>
      <el-card shadow="never" class="stat-card">
        <div class="stat-label">抄送我的</div>
        <div class="stat-value" style="color: #e6a23c">{{ ccCount }}</div>
      </el-card>
      <el-card shadow="never" class="stat-card">
        <div class="stat-label">审批意见</div>
        <div class="stat-value" style="color: #67c23a">{{ opinionCount }}</div>
      </el-card>
    </div>

    <div class="toolbar">
      <el-radio-group v-model="filterType" @change="loadList">
        <el-radio-button value="all">全部</el-radio-button>
        <el-radio-button value="cc">抄送我的</el-radio-button>
        <el-radio-button value="opinion">审批意见</el-radio-button>
        <el-radio-button value="transfer">转交任务</el-radio-button>
      </el-radio-group>
      <el-input v-model="keyword" placeholder="搜索..." clearable style="width: 260px" @keyup.enter="loadList" />
      <el-button type="primary" @click="loadList">
        <span class="refresh-icon">⟳</span>刷新
      </el-button>
      <div class="ferry-spacer"></div>
      <el-button :disabled="unreadCount === 0" @click="markAllRead">
        <el-icon><Check /></el-icon>一键标记全部已读
      </el-button>
    </div>

    <el-table :data="list" border stripe v-loading="loading" style="width: 100%">
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-badge v-if="!row._read" :value="'new'" class="badge-new" />
          <span v-else style="color: #909399">已读</span>
        </template>
      </el-table-column>
      <el-table-column label="类型" width="120">
        <template #default="{ row }">
          <el-tag v-if="row._type === 'cc'" type="info">抄送</el-tag>
          <el-tag v-else-if="row._type === 'opinion'" type="success">审批意见</el-tag>
          <el-tag v-else-if="row._type === 'transfer'" type="warning">转交</el-tag>
          <el-tag v-else type="primary">通知</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="标题" min-width="320" show-overflow-tooltip>
        <template #default="{ row }">
          <el-button type="primary" link @click="openDetail(row)">{{ row._title || '-' }}</el-button>
        </template>
      </el-table-column>
      <el-table-column label="关联工单" width="220" show-overflow-tooltip>
        <template #default="{ row }">{{ row._orderId || '-' }}</template>
      </el-table-column>
      <el-table-column label="发起人" width="120" show-overflow-tooltip>
        <template #default="{ row }">{{ row._from || '-' }}</template>
      </el-table-column>
      <el-table-column prop="_time" label="时间" width="180" />
    </el-table>

    <div class="pager">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="size"
        :total="total"
        :page-sizes="[100, 200, 500, 1000]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadList"
        @current-change="loadList"
      />
    </div>

    <el-drawer v-model="drawerVisible" title="详情" size="560px">
      <div v-if="detailItem" style="line-height: 1.8">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="标题">{{ detailItem._title }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ detailItem._type }}</el-descriptions-item>
          <el-descriptions-item label="关联工单">{{ detailItem._orderId }}</el-descriptions-item>
          <el-descriptions-item label="发起人">{{ detailItem._from }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ detailItem._read ? '已读' : '未读' }}</el-descriptions-item>
          <el-descriptions-item label="时间">{{ detailItem._time }}</el-descriptions-item>
        </el-descriptions>
        <div style="margin-top: 16px">
          <div style="font-weight: 600; margin-bottom: 8px">详细内容</div>
          <el-card shadow="never" style="font-size: 13px"><pre style="white-space: pre-wrap; word-break: break-all">{{ detailItem._content }}</pre></el-card>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { ferryMyRelated, ferryStatistics } from '@/api/ferry'

const loading = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const size = ref(100)
const keyword = ref('')
const filterType = ref('all')

const unreadCount = computed(() => list.value.filter((t) => !t._read).length)
const ccCount = computed(() => list.value.filter((t) => t._type === 'cc').length)
const opinionCount = computed(() => list.value.filter((t) => t._type === 'opinion').length)

function normalize(item: any): any {
  return {
    ...item,
    _type: item._type || item.type || 'notify',
    _title: item._title || item.title || '(未命名)',
    _orderId: item._orderId || item.orderId || item.workOrderId || '-',
    _from: item._from || item.from || item.creatorName || '系统',
    _time: item._time || item.createdAt || '-',
    _content: item._content || item.content || item.remark || '无详细内容',
    _read: item._read === undefined ? true : item._read
  }
}

async function loadList() {
  loading.value = true
  try {
    const res: any = await ferryMyRelated({ page: page.value, size: size.value, keyword: keyword.value })
    const raw = ((res?.data ?? res) as any[]) || []
    list.value = raw.map(normalize)
    total.value = (res?.total as number) ?? list.value.length
  } catch {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

async function markAllRead() {
  try {
    list.value.forEach((t) => (t._read = true))
    ElMessage.success('已标记全部为已读')
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

const drawerVisible = ref(false)
const detailItem = ref<any>(null)

function openDetail(row: any) {
  detailItem.value = row
  row._read = true
  drawerVisible.value = true
}

onMounted(loadList)
</script>

<style scoped>
.stat-row {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}
.stat-card {
  flex: 1;
  min-width: 150px;
  padding: 4px 8px;
}
.stat-label { color: #909399; font-size: 13px; }
.stat-value { font-size: 26px; font-weight: 700; margin-top: 4px; }
.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
  flex-wrap: wrap;
  align-items: center;
}
.ferry-spacer { flex: 1; }
.pager { display: flex; justify-content: flex-end; margin-top: 12px; }
.page-header { margin-bottom: 12px; }
.page-title { font-size: 20px; font-weight: 600; color: #303133; }
.page-sub { font-size: 13px; color: #909399; margin-top: 4px; }
.badge-new { position: relative; }
</style>