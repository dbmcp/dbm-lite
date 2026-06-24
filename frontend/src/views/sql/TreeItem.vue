<template>
  <div class="tree-item">
    <div
      class="tree-item-row"
      :class="{ 'tree-item-row--active': isActive }"
      @click.stop="handleRowClick"
      @dblclick.stop="handleDblClick"
      @contextmenu.prevent="handleCtxMenu"
    >
      <span class="tree-item-toggle" @click.stop="toggleExpand">
        <template v-if="hasChildren">
          <svg v-if="expanded" viewBox="0 0 20 20" width="10" height="10">
            <path d="M5 7 L10 12 L15 7" fill="none" stroke="#606266" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          <svg v-else viewBox="0 0 20 20" width="10" height="10">
            <path d="M7 5 L12 10 L7 15" fill="none" stroke="#606266" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </template>
      </span>
      <span class="tree-item-icon">
        <component :is="typeIcon" />
      </span>
      <span class="tree-item-name" :title="node.name">{{ node.name }}</span>
      <span v-if="extraInfo" class="tree-item-extra">{{ extraInfo }}</span>
      <span
        v-if="showInfoIcon"
        class="tree-item-info"
        @click.stop="handleInfoClick"
        title="查看表信息"
      >ℹ</span>
    </div>
    <div v-if="expanded && filteredChildren && filteredChildren.length > 0" class="tree-item-children">
      <TreeItem
        v-for="child in filteredChildren"
        :key="(child as any).type + '-' + (child as any).name"
        :node="child"
        :level="level + 1"
        :search-keyword="searchKeyword"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, ref, watch, h, PropType } from 'vue'

export interface TreeNode {
  type: 'database' | 'group' | 'table' | 'view' | 'procedure' | 'function' | 'trigger' | 'column' | 'index' | string
  name: string
  database?: string
  table?: string
  group?: string
  children?: TreeNode[]
  rows?: number
  sizeMb?: number
  pk?: boolean
  colType?: string
}

interface TreeCallbacks {
  onNodeDblClick: (node: TreeNode) => void
  onNodeCtxMenu: (e: MouseEvent, node: TreeNode) => void
  onNodeInfoClick: (node: TreeNode) => void
}

const props = defineProps({
  node: {
    type: Object as PropType<TreeNode>,
    required: true
  },
  level: {
    type: Number,
    default: 0
  },
  searchKeyword: {
    type: String,
    default: ''
  }
})

const expanded = ref<boolean>(props.level < 2)

const callbacks = inject<TreeCallbacks>('sqlTreeCallbacks', {
  onNodeDblClick: () => {},
  onNodeCtxMenu: () => {},
  onNodeInfoClick: () => {}
})

const hasChildren = computed<boolean>(() => {
  return !!(props.node.children && props.node.children.length > 0)
})

const svgIcon = (pathD: string, fill = '#409EFF') =>
  h('svg', { viewBox: '0 0 24 24', width: 16, height: 16, style: { display: 'block' } }, [
    h('path', { d: pathD, fill })
  ])

const DbIcon = () => svgIcon('M12 2C8 2 5 4 5 7v10c0 3 3 5 7 5s7-2 7-5V7c0-3-3-5-7-5zm0 2c3.3 0 5 1.3 5 3 0 .6-.4 1.3-1 2-.7.8-1.9 1.3-4 1.3s-3.3-.5-4-1.3c-.6-.7-1-1.4-1-2 0-1.7 1.7-3 5-3zM7 10c0 .6.4 1.3 1 2 .7.8 1.9 1.3 4 1.3s3.3-.5 4-1.3c.6-.7 1-1.4 1-2v1.5c0 1.7-1.7 3-5 3s-5-1.3-5-3V10zm0 4.5c0 1.7 1.7 3 5 3s5-1.3 5-3v1.5c0 1.7-1.7 3-5 3s-5-1.3-5-3v-1.5z', '#409EFF')
const TableIcon = () => svgIcon('M4 4h16v4H4zM4 10h16v4H4zM4 16h16v4H4z', '#67C23A')
const ViewIcon = () => svgIcon('M12 4C7 4 3 9 3 12s4 8 9 8 9-4 9-8-4-8-9-8zm0 4a4 4 0 110 8 4 4 0 010-8z', '#E6A23C')
const ProcIcon = () => svgIcon('M12 2a3 3 0 11-.001 6.001A3 3 0 0112 2zm0 14c-4 0-8 2-8 4v2h16v-2c0-2-4-4-8-4zM6 10h12v2H6z', '#9B59B6')
const FuncIcon = () => svgIcon('M4 6h16v3H4zM4 15h10v3H4z', '#F56C6C')
const TriggerIcon = () => svgIcon('M13 2L4 14h7l-1 8 9-12h-7l1-8z', '#E6A23C')
const ColumnIcon = () => svgIcon('M5 6h14v4H5zM5 12h14v4H5zM5 18h10v2H5z', '#909399')
const IndexIcon = () => svgIcon('M3 4h18v3H3zM3 9h14v3H3zM3 14h18v3H3zM3 19h14v2H3z', '#409EFF')
const FolderIcon = () => svgIcon('M10 4H4a2 2 0 00-2 2v12a2 2 0 002 2h16a2 2 0 002-2V8a2 2 0 00-2-2h-8l-2-2z', '#67C23A')
const GroupIcon = () => svgIcon('M3 5h18v3H3zM3 11h18v3H3zM3 17h18v3H3z', '#909399')

const typeIcon = computed(() => {
  const t = props.node.type || ''
  switch (t) {
    case 'database':
      return DbIcon
    case 'table':
      return TableIcon
    case 'view':
      return ViewIcon
    case 'procedure':
      return ProcIcon
    case 'function':
      return FuncIcon
    case 'trigger':
      return TriggerIcon
    case 'column':
      return ColumnIcon
    case 'index':
      return IndexIcon
    case 'group':
      return FolderIcon
    default:
      return GroupIcon
  }
})

const extraInfo = computed<string>(() => {
  if (props.node.type === 'table' || props.node.type === 'view') {
    const parts: string[] = []
    if (typeof props.node.rows === 'number') parts.push(props.node.rows.toLocaleString() + ' 行')
    if (typeof props.node.sizeMb === 'number') parts.push(props.node.sizeMb.toFixed(2) + ' MB')
    return parts.length > 0 ? '(' + parts.join(', ') + ')' : ''
  }
  if (props.node.type === 'column' && (props.node as any).colType) {
    return (props.node as any).colType
  }
  return ''
})

const showInfoIcon = computed<boolean>(() => {
  return props.node.type === 'table' || props.node.type === 'view'
})

const isActive = computed<boolean>(() => false)

function nodeMatches(node: TreeNode, keyword: string): boolean {
  if (!keyword) return true
  const kw = keyword.toLowerCase()
  if (node.name.toLowerCase().includes(kw)) return true
  if (node.children && node.children.length > 0) {
    return node.children.some((c) => nodeMatches(c, keyword))
  }
  return false
}

const filteredChildren = computed<TreeNode[] | undefined>(() => {
  if (!props.node.children || !props.node.children.length) return undefined
  if (!props.searchKeyword) return props.node.children
  return props.node.children.filter((c) => nodeMatches(c, props.searchKeyword))
})

watch(
  () => props.searchKeyword,
  (kw) => {
    if (kw && filteredChildren.value && filteredChildren.value.length > 0) {
      expanded.value = true
    }
  }
)

function toggleExpand() {
  if (hasChildren.value) {
    expanded.value = !expanded.value
  }
}

function handleRowClick() {
  if (hasChildren.value) {
    expanded.value = !expanded.value
  }
}

function handleDblClick() {
  if (callbacks.onNodeDblClick) callbacks.onNodeDblClick(props.node)
}

function handleCtxMenu(e: MouseEvent) {
  if (callbacks.onNodeCtxMenu) callbacks.onNodeCtxMenu(e, props.node)
}

function handleInfoClick() {
  if (callbacks.onNodeInfoClick) callbacks.onNodeInfoClick(props.node)
}
</script>

<style scoped>
.tree-item {
  font-size: 13px;
  color: #303133;
}

.tree-item-row {
  display: flex;
  align-items: center;
  padding: 3px 6px;
  cursor: pointer;
  border-radius: 3px;
  line-height: 20px;
  white-space: nowrap;
  user-select: none;
  transition: background-color 0.12s ease;
}

.tree-item-row:hover {
  background-color: #ecf5ff;
}

.tree-item-row--active {
  background-color: #d9ecff;
  color: #409eff;
}

.tree-item-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  flex-shrink: 0;
}

.tree-item-icon {
  margin-right: 4px;
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.tree-item-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tree-item-extra {
  margin-left: 6px;
  color: #909399;
  font-size: 12px;
  flex-shrink: 0;
}

.tree-item-info {
  margin-left: 6px;
  color: #409eff;
  font-size: 12px;
  cursor: pointer;
  flex-shrink: 0;
}

.tree-item-info:hover {
  color: #66b1ff;
}

.tree-item-children {
  padding-left: 14px;
}
</style>
