<template>
  <div class="tree-node-wrapper">
    <div
      class="tree-node"
      :style="{ paddingLeft: (level - 1) * 16 + 8 + 'px' }"
      @click="toggle"
      @dblclick.stop.prevent="onDblClick"
      @contextmenu.stop.prevent="onCtxMenu"
    >
      <span class="arrow" :class="{ invisible: !hasChildren, open: expanded }">
        <svg viewBox="0 0 10 10" width="9" height="9">
          <polygon points="2,2 8,5 2,8" fill="#888" />
        </svg>
      </span>
      <span class="icon" v-html="iconFor(node.type)"></span>
      <span class="name" :title="node.name">{{ node.name }}</span>
      <span v-if="extraInfo()" class="extra">{{ extraInfo() }}</span>
    </div>
    <div v-if="hasChildren && expanded" class="children">
      <TreeItem
        v-for="(child, idx) in node.children"
        :key="'c-' + idx + '-' + (child.name || '')"
        :node="child"
        :level="level + 1"
        :datasource-id="datasourceId"
        :search-keyword="searchKeyword"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, ref } from 'vue'
import type { TreeNode, TreeCallbacks } from '../types'

const props = defineProps<{
  node: TreeNode
  level: number
  searchKeyword?: string
  datasourceId?: string
}>()

const callbacks = inject<TreeCallbacks>('sqlTreeCallbacks')
const expanded = ref<boolean>(false)
const hasChildren = computed(() => Array.isArray(props.node.children) && (props.node.children as any[]).length > 0)

function iconFor(type: string) {
  const t = (type || '').toLowerCase()
  if (t === 'database') {
    return '<svg viewBox="0 0 16 16" width="14" height="14"><ellipse cx="8" cy="2.5" rx="5" ry="1.3" fill="#3f51b5" stroke="#1a237e" stroke-width="0.5"/><path d="M3 3 v3 a5 1.3 0 0 0 10 0 v-3" fill="#5c6bc0" stroke="#1a237e" stroke-width="0.5"/><ellipse cx="8" cy="7.5" rx="5" ry="1.3" fill="#3f51b5" stroke="#1a237e" stroke-width="0.5"/><path d="M3 8 v3 a5 1.3 0 0 0 10 0 v-3" fill="#5c6bc0" stroke="#1a237e" stroke-width="0.5"/><ellipse cx="8" cy="12.5" rx="5" ry="1.3" fill="#3f51b5" stroke="#1a237e" stroke-width="0.5"/></svg>'
  }
  if (t === 'table') {
    return '<svg viewBox="0 0 16 16" width="14" height="14"><rect x="2" y="3" width="12" height="10" rx="1" fill="#ffffff" stroke="#2e7d32" stroke-width="0.8"/><path d="M2 6.5 h12 M2 9.5 h12" stroke="#a5d6a7" stroke-width="0.6"/><path d="M5.5 3 v10 M9.5 3 v10" stroke="#a5d6a7" stroke-width="0.6"/></svg>'
  }
  if (t === 'view') {
    return '<svg viewBox="0 0 16 16" width="14" height="14"><rect x="2" y="3" width="12" height="10" rx="1" fill="#ffffff" stroke="#0277bd" stroke-width="0.8"/><circle cx="8" cy="8" r="1.5" fill="#0288d1"/><path d="M4 5.5 h8 M4 11 h8" stroke="#4fc3f7" stroke-width="0.6"/></svg>'
  }
  if (t === 'function') {
    return '<svg viewBox="0 0 16 16" width="14" height="14"><rect x="3" y="3" width="10" height="10" rx="1.5" fill="#fff3e0" stroke="#ef6c00" stroke-width="0.8"/><path d="M5 9 q1.5 -4 3 -4 q1.5 0 2 4 q0.5 2 1.5 2" fill="none" stroke="#ef6c00" stroke-width="1" stroke-linecap="round"/></svg>'
  }
  if (t === 'procedure') {
    return '<svg viewBox="0 0 16 16" width="14" height="14"><rect x="3" y="3" width="10" height="10" rx="1.5" fill="#f3e5f5" stroke="#7b1fa2" stroke-width="0.8"/><path d="M5 6 h6 M5 8.5 h6 M5 11 h4" stroke="#7b1fa2" stroke-width="0.8"/></svg>'
  }
  if (t === 'trigger') {
    return '<svg viewBox="0 0 16 16" width="14" height="14"><path d="M8 2 l4 7 h-3 l1.5 5 h-5 l1.5 -5 h-3 z" fill="#fff59d" stroke="#f57f17" stroke-width="0.6"/></svg>'
  }
  if (t === 'event') {
    return '<svg viewBox="0 0 16 16" width="14" height="14"><circle cx="8" cy="8" r="5" fill="#ffebee" stroke="#c62828" stroke-width="0.8"/><path d="M8 5 v4 l3 2" fill="none" stroke="#c62828" stroke-width="1" stroke-linecap="round"/></svg>'
  }
  if (t === 'index') {
    return '<svg viewBox="0 0 16 16" width="14" height="14"><path d="M11 3 l-5 10" stroke="#1565c0" stroke-width="1.2"/><path d="M5 3.5 h5 M7.5 12 h4" stroke="#1565c0" stroke-width="0.8"/></svg>'
  }
  if (t === 'column') {
    return '<svg viewBox="0 0 16 16" width="14" height="14"><rect x="3" y="7" width="10" height="2" rx="1" fill="#e8e8e8" stroke="#757575" stroke-width="0.5"/><circle cx="4.3" cy="8" r="0.5" fill="#757575"/></svg>'
  }
  if (t === 'group') {
    return '<svg viewBox="0 0 16 16" width="14" height="14"><path d="M2 4 h4 l1.2 1.5 h6.8 v7 h-12 z" fill="#ffe082" stroke="#ef6c00" stroke-width="0.8"/></svg>'
  }
  if (t === 'query-save') {
		return '<svg viewBox="0 0 16 16" width="14" height="14"><rect x="2" y="3" width="12" height="10" rx="1" fill="#e3f2fd" stroke="#1976d2" stroke-width="0.8"/><path d="M4 6 h8 M4 9 h6 M4 12 h4" stroke="#1976d2" stroke-width="0.8" stroke-linecap="round"/><circle cx="12" cy="12" r="1.5" fill="#1976d2"/></svg>'
	}
	if (t === 'saved-query') {
		return '<svg viewBox="0 0 16 16" width="14" height="14"><rect x="3" y="4" width="10" height="8" rx="1" fill="#ffffff" stroke="#1976d2" stroke-width="0.8"/><path d="M5 7 h6 M5 9 h4 M5 11 h3" stroke="#1976d2" stroke-width="0.6" stroke-linecap="round"/></svg>'
	}
	return '<svg viewBox="0 0 16 16" width="14" height="14"><circle cx="8" cy="8" r="4.5" fill="#e0e0e0" stroke="#757575" stroke-width="0.5"/></svg>'
}

function formatNumber(v: any): string {
  if (v === undefined || v === null || v === '') return ''
  const n = typeof v === 'number' ? v : parseFloat(String(v))
  if (isNaN(n)) return ''
  return n.toLocaleString()
}

function formatSize(v: any): string {
  if (v === undefined || v === null || v === '') return ''
  const mb = typeof v === 'number' ? v : parseFloat(String(v))
  if (isNaN(mb)) return ''
  if (mb < 0.1) {
    const kb = mb * 1024
    if (kb < 1) return Math.round(kb * 1024) + ' B'
    return kb.toFixed(1) + ' KB'
  }
  if (mb < 1024) return mb.toFixed(2) + ' MB'
  const gb = mb / 1024
  return gb.toFixed(2) + ' GB'
}

function extraInfo(): string {
  const n: any = props.node
  const parts: string[] = []
  if (typeof n.rows === 'number') {
    parts.push(formatNumber(n.rows) + ' 行')
  }
  if (typeof n.sizeMb === 'number') {
    parts.push(formatSize(n.sizeMb))
  }
  if (parts.length === 0 && n.colType) {
    return String(n.colType)
  }
  if (parts.length === 0 && n.comment) {
    return String(n.comment)
  }
  return parts.join(' · ')
}

function toggle() {
  if (props.node.type === 'saved-query') {
    callbacks?.onNodeDblClick?.(props.node)
    return
  }
  if (hasChildren.value) expanded.value = !expanded.value
}

function onDblClick() {
  callbacks?.onNodeDblClick?.(props.node)
}

function onCtxMenu(e: MouseEvent) {
  if (!props.node.datasourceId && props.datasourceId) {
    (props.node as any).datasourceId = props.datasourceId
  }
  callbacks?.onNodeCtxMenu?.(e, props.node)
}
</script>

<style scoped>
.tree-node-wrapper { user-select: none; }
.tree-node {
  display: flex;
  align-items: center;
  height: 24px;
  font-size: 12.5px;
  color: #303133;
  cursor: pointer;
  border-radius: 2px;
  padding-right: 8px;
  transition: background 0.1s;
}
.tree-node:hover { background: #e3f2fd; }
.tree-node:active { background: #bbdefb; }
.arrow {
  width: 14px; height: 14px;
  display: inline-flex; align-items: center; justify-content: center;
  margin-right: 2px; flex-shrink: 0;
  transition: transform 0.12s ease;
}
.arrow.invisible { visibility: hidden; }
.arrow.open { transform: rotate(90deg); }
.icon {
  width: 18px; height: 16px;
  display: inline-flex; align-items: center; justify-content: center;
  margin-right: 4px; flex-shrink: 0;
}
.name {
  flex: 1; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.extra {
  color: #78909c; font-size: 11px; margin-left: 10px;
  padding: 1px 6px; background: #f5f7fa; border-radius: 2px;
  white-space: nowrap;
}
.children { display: block; }
</style>
