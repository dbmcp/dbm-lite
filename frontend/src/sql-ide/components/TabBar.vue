<template>
  <div class="tab-bar">
    <div class="tab-list">
      <div
        v-for="tab in tabs"
        :key="tab.id"
        class="tab-item"
        :class="{ active: activeTabId === tab.id }"
        @click="switchTab(tab.id)"
      >
        <span class="tab-icon">
          <svg v-if="tab.kind === 'query'" viewBox="0 0 16 16" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
            <path d="M3 13 l3.5 -3.5 M9 5.5 L11.5 3 l1.5 1.5 L10.5 8.5 z" />
            <path d="M2 14 h12" />
          </svg>
          <svg v-else viewBox="0 0 16 16" width="14" height="14" fill="none" stroke="#4a90e2" stroke-width="1.3" stroke-linejoin="round">
            <rect x="2" y="3" width="12" height="10" rx="1" />
            <path d="M2 6.5 h12 M2 9.5 h12" stroke="#b8cce8" />
            <path d="M5.5 3 v10 M9.5 3 v10" stroke="#b8cce8" />
          </svg>
        </span>
        <span class="tab-name" :title="tab.title">{{ tab.title }}</span>
        <span v-if="executing && activeTabId === tab.id" class="tab-spinner">
          <svg viewBox="0 0 16 16" width="12" height="12">
            <circle cx="8" cy="8" r="6" stroke="#1a73e8" stroke-width="2" fill="none" stroke-dasharray="18 37.7" />
          </svg>
        </span>
        <span class="tab-close" @click.stop="closeTab(tab.id)">×</span>
      </div>
      <div class="tab-add" @click="onAddQuery">
        <svg viewBox="0 0 16 16" width="14" height="14" fill="none" stroke="#1a73e8" stroke-width="1.5" stroke-linecap="round">
          <path d="M8 3 v10 M3 8 h10" />
        </svg>
        <span>新建查询</span>
      </div>
    </div>
    <div class="tab-bar-status">
      <span class="status-icon">
        <svg v-if="!executing" viewBox="0 0 16 16" width="14" height="14" fill="none" stroke="#34a853" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="8" cy="8" r="5.5" />
          <path d="M5 8 l2 2 l4 -4" />
        </svg>
        <svg v-else class="spin" viewBox="0 0 16 16" width="14" height="14">
          <circle cx="8" cy="8" r="6" stroke="#1a73e8" stroke-width="2" fill="none" stroke-dasharray="18 37.7" />
        </svg>
      </span>
      <span>{{ executing ? '执行中...' : '就绪' }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { AnyTabState } from '../types'

defineProps<{
  tabs: AnyTabState[]
  activeTabId: string
  executing: boolean
}>()

const emit = defineEmits<{
  (e: 'switch', id: string): void
  (e: 'close', id: string): void
  (e: 'addQuery'): void
}>()

function switchTab(id: string) { emit('switch', id) }
function closeTab(id: string) { emit('close', id) }
function onAddQuery() { emit('addQuery') }
</script>

<style scoped>
.tab-bar {
  display: flex; align-items: center;
  padding: 0 10px 0 10px;
  background: #f8f9fa;
  border-bottom: 1px solid #dadce0;
  height: 36px;
  flex: 0 0 auto;
}

.tab-list { display: flex; align-items: stretch; flex: 1 1 auto; min-width: 0; overflow-x: auto; }

.tab-item {
  display: flex; align-items: center; gap: 6px;
  padding: 0 10px;
  height: 30px;
  background: #eaeaec;
  border: 1px solid transparent;
  border-bottom: none;
  border-radius: 3px 3px 0 0;
  margin-right: 3px;
  margin-top: 5px;
  cursor: pointer;
  font-size: 13px; color: #5a5a5a;
  white-space: nowrap; flex: 0 0 auto;
}
.tab-item.active {
  background: #ffffff; color: #202124;
  border: 1px solid #dadce0;
  border-bottom: 2px solid #1a73e8;
  font-weight: 600;
}
.tab-icon { display: inline-flex; align-items: center; }
.tab-name { max-width: 180px; overflow: hidden; text-overflow: ellipsis; }
.tab-close {
  margin-left: 4px; color: #80868b; font-size: 16px; line-height: 1;
  padding: 0 2px; border-radius: 2px; display: inline-flex; align-items: center;
}
.tab-close:hover { background: #e0e0e0; color: #202124; }

.tab-spinner {
  display: inline-flex; align-items: center;
  animation: spin 1s linear infinite;
}

.tab-add {
  display: inline-flex; align-items: center; gap: 4px;
  padding: 0 10px; height: 30px; line-height: 30px;
  color: #1a73e8; cursor: pointer; font-size: 13px;
  margin-right: 8px; margin-top: 5px; border-radius: 3px;
}
.tab-add:hover { background: #e8f0fe; }

.tab-bar-status {
  display: flex; align-items: center; gap: 6px;
  font-size: 12px; color: #5f6368;
  padding-left: 10px; border-left: 1px solid #e8eaed; flex: 0 0 auto;
}
.status-icon { display: inline-flex; align-items: center; }
.spin { animation: spin 1s linear infinite; }

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
