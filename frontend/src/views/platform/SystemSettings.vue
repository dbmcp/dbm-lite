<template>
  <div class="page-container">
    <div class="page-header">
      <div class="page-title">系统配置</div>
      <div class="page-sub">调整平台全局参数与功能开关。所有修改将即时生效。</div>
    </div>

    <div class="toolbar">
      <el-input v-model="keyword" placeholder="搜索配置项" clearable style="width:220px" />
      <el-select v-model="activeGroup" placeholder="分组" clearable style="width:180px">
        <el-option v-for="g in groups" :key="g" :label="g" :value="g" />
      </el-select>
      <el-button type="primary" plain @click="onSearch">查询</el-button>
      <el-button type="success" @click="appendNew">新增配置</el-button>
      <el-button type="primary" :loading="saving" @click="saveAll">保存全部</el-button>
    </div>

    <el-empty v-if="!loading && !Object.keys(groupedList).length" description="暂无配置项" />

    <div v-else>
      <el-card v-for="(items, group) in groupedList" :key="group" shadow="never" class="section-card" v-loading="loading">
        <template #header>
          <div class="card-header">
            <el-icon><Setting /></el-icon>
            <span class="group-name">{{ group }}</span>
            <span class="count-tag">共 {{ items.length }} 项</span>
          </div>
        </template>
        <el-table :data="items" border stripe style="width:100%">
          <el-table-column prop="settingKey" label="配置项" min-width="220" />
          <el-table-column label="配置值" min-width="360">
            <template #default="{ row }">
              <el-input v-model="row.value" :placeholder="'请输入 ' + (row.settingKey || '值')" />
            </template>
          </el-table-column>
          <el-table-column label="说明" min-width="260">
            <template #default="{ row }">
              <el-input v-model="row.remark" placeholder="配置项的用途说明" />
            </template>
          </el-table-column>
          <el-table-column prop="updatedAt" label="更新时间" width="180" />
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button v-if="row.isNew" size="small" type="danger" link @click="removeNew(row)">移除</el-button>
              <span v-else style="color:#c0c4cc">—</span>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Setting } from '@element-plus/icons-vue'
import { ferryListSystemSettings, ferrySaveSystemSettings } from '@/api/ferry'

const keyword = ref('')
const activeGroup = ref('')
const loading = ref(false)
const saving = ref(false)
const allItems = ref<any[]>([])

function deriveGroup(key: string): string {
  if (!key) return '其他设置'
  const lower = key.toLowerCase()
  if (lower.startsWith('system.') || lower.startsWith('sys.')) return '系统设置'
  if (lower.startsWith('security.') || lower.startsWith('auth.') || lower.startsWith('login.')) return '安全与认证'
  if (lower.startsWith('email.') || lower.startsWith('mail.') || lower.startsWith('notify.')) return '邮件与通知'
  if (lower.startsWith('file.') || lower.startsWith('storage.') || lower.startsWith('oss.') || lower.startsWith('upload.')) return '文件与存储'
  if (lower.startsWith('db.') || lower.startsWith('database.') || lower.startsWith('mysql.')) return '数据库配置'
  if (lower.startsWith('cache.') || lower.startsWith('redis.')) return '缓存配置'
  if (lower.startsWith('log.') || lower.startsWith('logging.') || lower.startsWith('audit.')) return '日志与审计'
  if (lower.startsWith('api.') || lower.startsWith('integration.') || lower.startsWith('thirdparty.') || lower.startsWith('third.')) return '接口与集成'
  if (lower.startsWith('ui.') || lower.startsWith('theme.') || lower.startsWith('frontend.') || lower.startsWith('app.')) return '界面与主题'
  return '其他设置'
}

function normalize(raw: any): any {
  const key = raw.settingKey ?? raw.key ?? ''
  return {
    id: raw.id ?? '',
    settingKey: key,
    value: raw.value ?? '',
    remark: raw.remark ?? '',
    updatedAt: raw.updatedAt ?? '',
    group: raw.group ?? raw.category ?? deriveGroup(key),
    isNew: !!raw.isNew || false
  }
}

function onSearch() {
  loadList()
}

async function loadList() {
  loading.value = true
  try {
    const res: any = await ferryListSystemSettings(keyword.value)
    const items: any[] = Array.isArray(res) ? res : Array.isArray(res?.data) ? res.data : []
    allItems.value = items.map(normalize)
  } catch (e: any) {
    ElMessage.error(e?.message || '加载失败')
  } finally {
    loading.value = false
  }
}

const filteredList = computed(() => {
  let items = allItems.value
  if (keyword.value) {
    const kw = keyword.value.toLowerCase()
    items = items.filter((x: any) =>
      (x.settingKey || '').toLowerCase().includes(kw) ||
      (x.value || '').toLowerCase().includes(kw) ||
      (x.remark || '').toLowerCase().includes(kw)
    )
  }
  if (activeGroup.value) {
    items = items.filter((x: any) => x.group === activeGroup.value)
  }
  return items
})

const groupedList = computed(() => {
  const map: Record<string, any[]> = {}
  for (const it of filteredList.value) {
    if (!map[it.group]) map[it.group] = []
    map[it.group].push(it)
  }
  return map
})

const groups = computed(() => {
  const set = new Set<string>()
  for (const it of allItems.value) set.add(it.group)
  return Array.from(set)
})

function appendNew() {
  const newItem = normalize({
    settingKey: '',
    value: '',
    remark: '',
    group: activeGroup.value || '其他设置',
    isNew: true
  })
  allItems.value.unshift(newItem)
}

function removeNew(row: any) {
  const idx = allItems.value.findIndex((x: any) => x.settingKey === row.settingKey && x.isNew)
  if (idx >= 0) {
    allItems.value.splice(idx, 1)
    return
  }
  const idx2 = allItems.value.indexOf(row)
  if (idx2 >= 0) allItems.value.splice(idx2, 1)
}

async function saveAll() {
  const toSave = allItems.value.filter((x: any) => (x.settingKey || '').trim())
  if (!toSave.length) {
    ElMessage.warning('没有可保存的配置项')
    return
  }
  const seen = new Set<string>()
  for (const it of toSave) {
    const k = (it.settingKey || '').trim()
    if (seen.has(k)) {
      ElMessage.warning('存在重复的配置项：' + k)
      return
    }
    seen.add(k)
  }
  saving.value = true
  try {
    await ferrySaveSystemSettings(toSave.map((x: any) => ({
      settingKey: (x.settingKey || '').trim(),
      value: x.value || '',
      remark: x.remark || ''
    })))
    ElMessage.success('保存成功')
    await loadList()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  } finally {
    saving.value = false
  }
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
.section-card {
  border-radius: 8px;
  margin-bottom: 16px;
}
.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
}
.group-name {
  font-weight: 600;
  color: #1f2937;
}
.count-tag {
  margin-left: auto;
  color: #909399;
  font-size: 12px;
}
</style>
