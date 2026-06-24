<template>
  <div class="qt-root">
    <div class="qt-left">
      <span class="qt-label">连接</span>
      <el-select v-model="selectedDs" filterable size="small" class="qt-select" @change="handleDsChange">
        <el-option v-for="ds in datasources" :key="ds.datasourceId" :label="ds.name" :value="ds.datasourceId" />
      </el-select>
      <span class="qt-label">数据库</span>
      <el-select v-model="selectedDb" filterable size="small" class="qt-select" :disabled="!selectedDs" @change="(v: string) => emit('dbChange', v)">
        <el-option v-for="db in databases" :key="db" :label="db" :value="db" />
      </el-select>
    </div>
    <div class="qt-sep-v"></div>
    <div class="qt-buttons">
      <button class="qt-btn qt-btn-run" :disabled="!selectedDs" @click="emit('run')" title="运行 (Ctrl+Enter)">
        <span class="qt-icon"><svg viewBox="0 0 16 16" width="14" height="14"><polygon points="3,2 14,8 3,14" fill="#2a9d3b"/></svg></span>
        <span class="qt-text">运行</span>
      </button>
      <button class="qt-btn qt-btn-stop" :disabled="!selectedDs || !executing" @click="emit('stop')" title="停止">
        <span class="qt-icon"><svg viewBox="0 0 16 16" width="14" height="14"><rect x="3" y="3" width="10" height="10" fill="#c62828"/></svg></span>
        <span class="qt-text">停止</span>
      </button>
      <button class="qt-btn qt-btn-explain" :disabled="!selectedDs" @click="emit('explain')" title="执行计划">
        <span class="qt-icon"><svg viewBox="0 0 16 16" width="14" height="14"><rect x="2" y="2" width="12" height="12" rx="1.5" fill="#007bff"/><path d="M5 6 h7 M5 9 h5 M5 12 h4" stroke="#ffffff" stroke-width="1.2" stroke-linecap="round"/></svg></span>
        <span class="qt-text">解释</span>
      </button>
      <button class="qt-btn qt-btn-beautify" :disabled="!selectedDs" @click="emit('beautify')" title="美化SQL">
        <span class="qt-icon"><svg viewBox="0 0 16 16" width="14" height="14"><rect x="2" y="2" width="12" height="12" rx="1.5" fill="#007bff"/><path d="M4 5 h8 M4 8 h6 M4 11 h4" stroke="#ffffff" stroke-width="1.2" stroke-linecap="round"/></svg></span>
        <span class="qt-text">美化</span>
      </button>
      <button class="qt-btn qt-btn-fav" :disabled="!selectedDs" @click="emit('favorite')" title="收藏">
        <span class="qt-icon"><svg viewBox="0 0 16 16" width="14" height="14"><polygon points="8,2 9.8,6 14,6 10.6,8.5 11.8,13 8,10.5 4.2,13 5.4,8.5 2,6 6.2,6" fill="#007bff" stroke="#0069d9" stroke-width="0.5"/></svg></span>
        <span class="qt-text">收藏</span>
      </button>
      <button class="qt-btn qt-btn-save" :disabled="!selectedDs" @click="emit('save')" title="保存">
        <span class="qt-icon"><svg viewBox="0 0 16 16" width="14" height="14"><path d="M2 2 h10 l2 2 v10 h-12 z" fill="#ffffff" stroke="#007bff" stroke-width="1"/><rect x="4" y="4" width="8" height="2.5" fill="#e3f2fd" stroke="#007bff" stroke-width="0.5"/><rect x="5" y="9" width="6" height="4" fill="#ffffff" stroke="#007bff" stroke-width="0.5"/></svg></span>
        <span class="qt-text">保存</span>
      </button>
      <button class="qt-btn qt-btn-refresh" :disabled="!selectedDs" @click="emit('refreshDb')" title="刷新">
        <span class="qt-icon">⟳</span>
        <span class="qt-text">刷新</span>
      </button>
    </div>
    <div class="qt-right">
      <span v-if="effectiveDb" class="qt-status">
        <span class="qt-status-dot"></span>已选数据库：<b>{{ effectiveDb }}</b>
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { DatasourceSummary } from '../types'

const props = defineProps<{
  datasourceId: string
  database: string
  datasources: DatasourceSummary[]
  databases: string[]
  executing: boolean
  effectiveDb?: string
}>()

const emit = defineEmits<{
  (e: 'dsChange', v: string): void
  (e: 'dbChange', v: string): void
  (e: 'run'): void
  (e: 'stop'): void
  (e: 'explain'): void
  (e: 'refreshDb'): void
  (e: 'save'): void
  (e: 'favorite'): void
  (e: 'beautify'): void
}>()

const selectedDs = computed<string>({
  get: () => props.datasourceId,
  set: (v: string) => emit('dsChange', v)
})

const selectedDb = computed<string>({
  get: () => props.database,
  set: (v: string) => emit('dbChange', v)
})

function handleDsChange(v: string) { emit('dsChange', v) }
</script>

<style scoped>
.qt-root {
  display: flex;
  align-items: center;
  padding: 6px 12px;
  background: linear-gradient(to bottom, #fafafa, #eeeeee);
  border-bottom: 1px solid #c0c4cc;
  gap: 8px;
  height: 44px;
  box-sizing: border-box;
}
.qt-left { display: flex; align-items: center; gap: 6px; }
.qt-label { font-size: 12px; color: #606266; padding: 2px 4px; }
.qt-select { width: 150px; }
.qt-select :deep(.el-input__wrapper) { box-shadow: 0 0 0 1px #c0c4cc; background: #ffffff; border-radius: 2px; padding: 0 8px; }
.qt-select :deep(.el-input__inner) { font-size: 12px; height: 24px; line-height: 24px; }
.qt-sep-v { width: 1px; height: 24px; background: linear-gradient(to bottom, transparent, #b0b0b0, transparent); margin: 0 4px; }
.qt-buttons { display: flex; align-items: center; gap: 2px; }
.qt-btn {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  padding: 4px 10px;
  min-width: 50px;
  height: 36px;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 3px;
  color: #303133;
  cursor: pointer;
  transition: all 0.12s ease;
}
.qt-btn:hover:not(:disabled) {
  background: linear-gradient(to bottom, #ffffff, #e3f2fd);
  border-color: #bbdefb;
  box-shadow: 0 1px 2px rgba(0,0,0,0.08);
}
.qt-btn:active:not(:disabled) { background: linear-gradient(to bottom, #e3f2fd, #bbdefb); }
.qt-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.qt-icon { display: inline-flex; align-items: center; justify-content: center; width: 18px; height: 16px; }
.qt-text { font-size: 11px; line-height: 1; white-space: nowrap; }
.qt-btn-run .qt-text { color: #2a9d3b; font-weight: 600; }
.qt-btn-stop .qt-text { color: #c62828; font-weight: 600; }
.qt-btn-explain .qt-text { color: #007bff; font-weight: 500; }
.qt-btn-beautify .qt-text { color: #007bff; font-weight: 500; }
.qt-btn-fav .qt-text { color: #007bff; font-weight: 500; }
.qt-btn-save .qt-text { color: #007bff; font-weight: 500; }
.qt-btn-refresh .qt-text { color: #007bff; font-weight: 500; }
.qt-right { margin-left: auto; display: flex; align-items: center; }
.qt-status {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 4px 10px; background: #ffffff; border: 1px solid #dcdfe6;
  border-radius: 3px; font-size: 11px; color: #606266;
}
.qt-status b { color: #1976d2; font-weight: 600; }
.qt-status-dot { width: 6px; height: 6px; border-radius: 50%; background: #4caf50; box-shadow: 0 0 0 2px rgba(76,175,80,0.2); }
</style>
