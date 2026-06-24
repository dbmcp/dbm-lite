<template>
  <div class="pg-bar">
    <div class="pg-info">
      第 {{ start }} - {{ end }} 条 (共 {{ total }} 条) · 第 {{ page }} 页 / 共 {{ totalPages }} 页
    </div>
    <div class="pg-controls">
      <select v-model.number="pageSizeLocal" class="pg-select" @change="onPageSizeChange">
        <option v-for="ps in pageSizeOptions" :key="ps" :value="ps">{{ ps }} 条/页</option>
      </select>
      <button class="pg-btn" :disabled="page <= 1" @click="goTo(1)">«</button>
      <button class="pg-btn" :disabled="page <= 1" @click="goTo(page - 1)">‹</button>
      <span class="pg-pager-info">{{ page }} / {{ totalPages }}</span>
      <button class="pg-btn" :disabled="page >= totalPages" @click="goTo(page + 1)">›</button>
      <button class="pg-btn" :disabled="page >= totalPages" @click="goTo(totalPages)">»</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const props = withDefaults(defineProps<{
  total: number
  page: number
  pageSize: number
  pageSizeOptions?: number[]
}>(), {
  total: 0,
  page: 1,
  pageSize: 50,
  pageSizeOptions: () => [20, 50, 100, 500, 1000]
})

const emit = defineEmits<{
  (e: 'update:page', v: number): void
  (e: 'update:pageSize', v: number): void
}>()

const pageSizeLocal = ref(props.pageSize)

watch(() => props.pageSize, (v) => {
  pageSizeLocal.value = v
})

const totalPages = computed(() => {
  if (!props.total) return 1
  return Math.max(1, Math.ceil(props.total / pageSizeLocal.value))
})

const start = computed(() => {
  if (!props.total) return 0
  return (props.page - 1) * pageSizeLocal.value + 1
})

const end = computed(() => {
  if (!props.total) return 0
  return Math.min(props.total, props.page * pageSizeLocal.value)
})

function goTo(p: number) {
  if (p < 1) p = 1
  if (p > totalPages.value) p = totalPages.value
  emit('update:page', p)
}

function onPageSizeChange() {
  emit('update:pageSize', pageSizeLocal.value)
  emit('update:page', 1)
}
</script>

<style scoped>
.pg-bar {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 12px;
  background: #fafafa;
  border-top: 1px solid #e4e7ed;
  font-size: 12px;
  color: #606266;
}

.pg-info {
  color: #606266;
}

.pg-controls {
  display: flex;
  align-items: center;
  gap: 4px;
}

.pg-select {
  height: 24px;
  padding: 0 4px;
  border: 1px solid #dcdfe6;
  border-radius: 3px;
  background: #ffffff;
  color: #606266;
  font-size: 12px;
  outline: none;
  cursor: pointer;
}

.pg-select:focus {
  border-color: #409eff;
}

.pg-btn {
  height: 24px;
  min-width: 26px;
  padding: 0 8px;
  background: #ffffff;
  border: 1px solid #dcdfe6;
  border-radius: 3px;
  color: #606266;
  font-size: 12px;
  cursor: pointer;
  outline: none;
}

.pg-btn:hover:not(:disabled) {
  border-color: #409eff;
  color: #409eff;
  background: #ecf5ff;
}

.pg-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.pg-pager-info {
  padding: 0 4px;
  color: #303133;
  font-weight: 500;
}
</style>
