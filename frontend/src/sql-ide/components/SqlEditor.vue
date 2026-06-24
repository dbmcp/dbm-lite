<template>
  <div class="sql-editor-container" @click="onContainerClick">
    <div class="sql-editor-wrapper">
      <textarea
        :id="'sql-editor-' + tabId"
        class="sql-editor-textarea"
        ref="textareaRef"
        :value="value"
        :readonly="readonly"
        :placeholder="placeholder"
        @input="onInput"
        @keydown="onKeydown"
        @click="onTextareaClick"
        @select="onTextareaSelect"
        @blur="onTextareaBlur"
        @focus="onTextareaFocus"
        spellcheck="false"
      ></textarea>
    </div>
    <!-- 智能提示下拉 -->
    <div v-if="suggestions.length > 0" class="suggestions-popup" :style="popupStyle">
      <div
        v-for="(item, idx) in suggestions"
        :key="item.key"
        class="suggestion-item"
        :class="{ active: idx === selectedIndex }"
        @click="applySuggestion(item)"
        @mousedown.prevent
      >
        <span class="s-icon" :class="'s-icon-' + item.type">{{ iconFor(item.type) }}</span>
        <span class="s-label">{{ item.label }}</span>
        <span class="s-comment">{{ item.comment }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'

interface Suggestion {
  key: string
  label: string
  type: 'keyword' | 'table' | 'column' | 'function'
  comment?: string
  insertText: string
}

const SQL_KEYWORDS = [
  'SELECT', 'FROM', 'WHERE', 'AND', 'OR', 'NOT', 'IN', 'LIKE', 'ORDER', 'BY', 'GROUP', 'HAVING',
  'INSERT', 'INTO', 'VALUES', 'UPDATE', 'SET', 'DELETE', 'CREATE', 'TABLE', 'DROP', 'ALTER',
  'INDEX', 'VIEW', 'FUNCTION', 'PROCEDURE', 'TRIGGER', 'EVENT', 'DATABASE', 'SCHEMA',
  'LIMIT', 'OFFSET', 'ON', 'JOIN', 'LEFT', 'RIGHT', 'INNER', 'OUTER', 'FULL', 'UNION', 'ALL',
  'DISTINCT', 'AS', 'CASE', 'WHEN', 'THEN', 'ELSE', 'END', 'NULL', 'IS', 'EXISTS',
  'BETWEEN', 'ISNULL', 'COALESCE', 'CAST', 'CONVERT', 'COUNT', 'SUM', 'AVG', 'MIN', 'MAX',
  'GROUP_CONCAT', 'CONCAT', 'SUBSTRING', 'TRIM', 'UPPER', 'LOWER', 'LEFT JOIN',
  'PRIMARY', 'KEY', 'FOREIGN', 'REFERENCES', 'UNIQUE', 'DEFAULT', 'CHECK', 'CONSTRAINT',
  'AUTO_INCREMENT', 'COMMENT', 'ENGINE', 'CHARSET', 'COLLATE', 'ROW_FORMAT', 'STORED',
  'EXPLAIN', 'ANALYZE', 'DESCRIBE', 'DESC', 'ASC', 'SHOW', 'USE', 'BEGIN', 'COMMIT',
  'ROLLBACK', 'START', 'TRANSACTION', 'WITH', 'RECURSIVE', 'OVER', 'PARTITION', 'ROW_NUMBER',
  'LATERAL', 'UNNEST', 'INNER JOIN', 'OUTER JOIN', 'CROSS JOIN', 'LEFT JOIN', 'RIGHT JOIN',
  'FULL OUTER JOIN', 'IF', 'IFNULL', 'NULLIF', 'NOW', 'CURRENT_TIMESTAMP', 'CURRENT_DATE',
  'CURRENT_TIME', 'DATE', 'TIME', 'DATETIME', 'TIMESTAMP', 'BOOLEAN', 'BOOL', 'TEXT',
  'VARCHAR', 'CHAR', 'INT', 'INTEGER', 'BIGINT', 'SMALLINT', 'TINYINT', 'FLOAT', 'DOUBLE',
  'DECIMAL', 'NUMERIC', 'BLOB', 'JSON', 'ENUM', 'RETURN', 'DECLARE', 'CONTINUE',
  'HANDLER', 'FOR', 'EACH', 'ROW', 'TRIGGER', 'AFTER', 'BEFORE', 'INSTEAD', 'REPLACE',
  'TRUNCATE', 'OPTIMIZE', 'REPAIR', 'CHECK TABLE', 'BACKUP', 'RESTORE', 'LIMIT 100', 'LIMIT 200',
  'LIMIT 500', 'LIMIT 1000'
]

const props = withDefaults(defineProps<{
  value: string
  tabId?: string
  suggestionNames?: string[]
  suggestionColumns?: string[]
  readonly?: boolean
  placeholder?: string
}>(), {
  value: '',
  tabId: '',
  suggestionNames: () => [],
  suggestionColumns: () => [],
  readonly: false,
  placeholder: '在此编写 SQL 语句，按 Ctrl+Enter 运行，输入空格/字母触发提示...'
})

const emit = defineEmits<{
  (e: 'input', v: string): void
  (e: 'run-command'): void
  (e: 'beautify'): void
}>()

const textareaRef = ref<HTMLTextAreaElement | null>(null)
const suggestions = ref<Suggestion[]>([])
const selectedIndex = ref(0)
const popupPos = ref<{ top: number; left: number }>({ top: 0, left: 0 })

// 保存选中状态
const savedSelection = ref<{ start: number; end: number } | null>(null)
// 保存编辑器内容（用于检测是否被外部修改）
const savedContent = ref('')
// 标记是否需要恢复选中状态
const shouldRestoreSelection = ref(false)

const popupStyle = computed(() => ({
  top: popupPos.value.top + 'px',
  left: popupPos.value.left + 'px'
}))

function iconFor(type: string): string {
  switch (type) {
    case 'keyword': return 'K'
    case 'table': return 'T'
    case 'column': return 'C'
    case 'function': return 'F'
    default: return '?'
  }
}

function getCaretCoordinates(ta: HTMLTextAreaElement): { top: number; left: number } {
  const rect = ta.getBoundingClientRect()
  const div = document.createElement('div')
  const style = window.getComputedStyle(ta)
  const props = [
    'boxSizing', 'width', 'height', 'overflowX', 'overflowY',
    'borderTopWidth', 'borderRightWidth', 'borderBottomWidth', 'borderLeftWidth',
    'paddingTop', 'paddingRight', 'paddingBottom', 'paddingLeft',
    'fontStyle', 'fontVariant', 'fontWeight', 'fontStretch', 'fontSize',
    'fontSizeAdjust', 'lineHeight', 'fontFamily',
    'textAlign', 'textTransform', 'textIndent', 'textDecoration',
    'letterSpacing', 'wordSpacing', 'tabSize', 'MozTabSize'
  ]
  div.style.position = 'absolute'
  div.style.visibility = 'hidden'
  div.style.whiteSpace = 'pre-wrap'
  div.style.wordWrap = 'break-word'
  div.style.top = '0'
  div.style.left = '0'
  div.style.width = ta.clientWidth + 'px'
  div.style.padding = style.padding
  div.style.font = style.font
  div.style.lineHeight = style.lineHeight
  div.style.fontFamily = style.fontFamily
  div.style.fontSize = style.fontSize

  const pos = ta.selectionStart
  const text = ta.value.substring(0, pos)
  div.textContent = text
  const span = document.createElement('span')
  span.textContent = ta.value.substring(pos) || '.'
  div.appendChild(span)
  document.body.appendChild(div)

  const spanRect = span.getBoundingClientRect()
  const divRect = div.getBoundingClientRect()
  document.body.removeChild(div)

  return {
    top: spanRect.top - divRect.top + parseInt(style.lineHeight || '20') + 8,
    left: spanRect.left - divRect.left + 2
  }
}

function computeSuggestions(text: string, pos: number): Suggestion[] {
  const before = text.substring(0, pos)
  const match = before.match(/[A-Za-z_][A-Za-z0-9_]*$/)
  if (!match) return []
  const word = match[0]
  const lower = word.toLowerCase()
  const results: Suggestion[] = []
  const added = new Set<string>()

  // 关键字匹配
  for (const kw of SQL_KEYWORDS) {
    if (kw.toLowerCase().startsWith(lower) && !added.has(kw)) {
      added.add(kw)
      results.push({
        key: 'kw_' + kw,
        label: kw,
        type: 'keyword',
        comment: 'SQL 关键字',
        insertText: kw
      })
    }
    if (results.length > 30) break
  }

  // 表/视图名匹配
  const names = props.suggestionNames || []
  for (const name of names) {
    if (name.toLowerCase().startsWith(lower) && !added.has('t_' + name)) {
      added.add('t_' + name)
      results.push({
        key: 't_' + name,
        label: name,
        type: 'table',
        comment: '表/视图',
        insertText: name
      })
    }
    if (results.length > 60) break
  }

  // 列名匹配（如果 suggestionColumns 提供）
  const cols = props.suggestionColumns || []
  for (const c of cols) {
    if (c.toLowerCase().startsWith(lower) && !added.has('c_' + c)) {
      added.add('c_' + c)
      results.push({
        key: 'c_' + c,
        label: c,
        type: 'column',
        comment: '列',
        insertText: c
      })
    }
  }

  // 识别 "表名." 前缀 -> 提供列名提示
  const dotMatch = before.match(/([A-Za-z_][A-Za-z0-9_]*)\.\s*([A-Za-z0-9_]*)$/)
  if (dotMatch) {
    for (const c of cols) {
      if (!added.has('cd_' + c)) {
        added.add('cd_' + c)
        results.unshift({
          key: 'cd_' + c,
          label: c,
          type: 'column',
          comment: '列（' + dotMatch[1] + '）',
          insertText: c
        })
      }
    }
  }

  return results.slice(0, 30)
}

function updateSuggestions() {
  const ta = textareaRef.value
  if (!ta || props.readonly) {
    suggestions.value = []
    return
  }
  const pos = ta.selectionStart
  const text = ta.value
  const items = computeSuggestions(text, pos)
  if (items.length === 0) {
    suggestions.value = []
    return
  }
  // 计算 popup 位置
  const coords = getCaretCoordinates(ta)
  popupPos.value = { top: coords.top, left: coords.left }
  suggestions.value = items
  selectedIndex.value = 0
}

function applySuggestion(item: Suggestion) {
  const ta = textareaRef.value
  if (!ta) return
  const pos = ta.selectionStart
  const text = ta.value
  const before = text.substring(0, pos)
  const after = text.substring(pos)
  const match = before.match(/[A-Za-z_][A-Za-z0-9_]*$/)
  const startPos = match ? before.length - match[0].length : pos
  // 识别 "表名." 前缀情况
  const dotMatch = before.substring(0, startPos).match(/([A-Za-z_][A-Za-z0-9_]*)\.\s*$/)
  let actualStart = startPos
  if (dotMatch && props.suggestionColumns && props.suggestionColumns.length > 0) {
    actualStart = startPos
  }
  const newText = text.substring(0, actualStart) + item.insertText + after
  const newPos = actualStart + item.insertText.length
  emit('input', newText)
  suggestions.value = []
  nextTick(() => {
    const ta2 = document.getElementById('sql-editor-' + props.tabId) as HTMLTextAreaElement
    if (ta2) {
      ta2.value = newText
      ta2.focus()
      ta2.setSelectionRange(newPos, newPos)
    }
  })
}

function onInput(e: Event) {
  const t = e.target as HTMLTextAreaElement
  emit('input', t.value)
  updateSuggestions()
}

function onTextareaClick() {
  updateSuggestions()
  // 用户主动点击编辑器，取消选中状态恢复标记
  shouldRestoreSelection.value = false
}

function onTextareaBlur() {
  // 总是保存选中状态（包括光标位置）和内容
  const ta = textareaRef.value
  if (ta) {
    const start = ta.selectionStart
    const end = ta.selectionEnd
    savedSelection.value = { start, end }
    savedContent.value = ta.value
    // 只有当有选中内容时才标记需要恢复
    shouldRestoreSelection.value = start !== end
  }
  suggestions.value = []
}

function onTextareaFocus() {
  // 恢复选中状态（不清除，以便可以多次恢复）
  const ta = textareaRef.value
  if (ta && savedSelection.value && shouldRestoreSelection.value) {
    // 只有当内容没有被外部修改时才恢复选中状态
    if (ta.value === savedContent.value) {
      ta.setSelectionRange(savedSelection.value.start, savedSelection.value.end)
    }
  }
}

function onTextareaSelect() {
  // 用户手动选择文本，更新保存的选中状态
  const ta = textareaRef.value
  if (ta) {
    const start = ta.selectionStart
    const end = ta.selectionEnd
    savedSelection.value = { start, end }
    savedContent.value = ta.value
    // 用户主动选择，不需要恢复标记
    shouldRestoreSelection.value = false
  }
}

function restoreSelection() {
  // 强制恢复选中状态
  const ta = textareaRef.value
  if (ta && savedSelection.value) {
    ta.focus()
    ta.setSelectionRange(savedSelection.value.start, savedSelection.value.end)
  }
}

defineExpose({
  restoreSelection
})

function onContainerClick() {
  // 点击容器外部区域，保持 textarea 聚焦
}

function onKeydown(e: KeyboardEvent) {
  if (e.ctrlKey && e.key === 'Enter') {
    e.preventDefault()
    emit('run-command')
    return
  }
  if (e.ctrlKey && e.shiftKey && e.key.toUpperCase() === 'F') {
    e.preventDefault()
    emit('beautify')
    return
  }

  if (suggestions.value.length > 0) {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      selectedIndex.value = (selectedIndex.value + 1) % suggestions.value.length
      return
    }
    if (e.key === 'ArrowUp') {
      e.preventDefault()
      selectedIndex.value = (selectedIndex.value - 1 + suggestions.value.length) % suggestions.value.length
      return
    }
    if (e.key === 'Tab' || e.key === 'Enter') {
      if (selectedIndex.value >= 0 && selectedIndex.value < suggestions.value.length) {
        e.preventDefault()
        applySuggestion(suggestions.value[selectedIndex.value])
        return
      }
    }
    if (e.key === 'Escape') {
      e.preventDefault()
      suggestions.value = []
      return
    }
  }
}

function hideSuggestions() {
  suggestions.value = []
}

// 外部注入 SQL 更新事件监听
let externalListener: ((e: Event) => void) | null = null
let documentClickListener: ((e: Event) => void) | null = null

onMounted(() => {
  const el = document.getElementById('sql-editor-' + props.tabId) as HTMLTextAreaElement
  if (el) textareaRef.value = el
  if (props.tabId) {
    externalListener = (e: Event) => {
      const ce = e as CustomEvent
      if (ce.detail && textareaRef.value) {
        textareaRef.value.value = ce.detail
        emit('input', ce.detail)
      }
    }
    window.addEventListener('sqlide:update-editor-' + props.tabId, externalListener)
  }
  documentClickListener = () => {
    // 点击外部关闭提示
    setTimeout(hideSuggestions, 0)
  }
  document.addEventListener('click', documentClickListener)
})

onBeforeUnmount(() => {
  if (props.tabId && externalListener) {
    window.removeEventListener('sqlide:update-editor-' + props.tabId, externalListener)
  }
  if (documentClickListener) {
    document.removeEventListener('click', documentClickListener)
  }
})
</script>

<style scoped>
.sql-editor-container {
  position: relative;
  flex: 1 1 40%;
  min-height: 150px;
  max-height: none;
  height: 100%;
  overflow: auto;
  background: #ffffff;
  border-bottom: 1px solid #e4e7ed;
}

.sql-editor-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
  min-height: 120px;
}

.sql-editor-textarea {
  width: 100%;
  min-height: 120px;
  height: 100%;
  max-height: 100%;
  padding: 10px 12px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.8;
  color: #303133;
  background: #fafafa;
  border: none;
  resize: none;
  outline: none;
  white-space: pre;
  overflow: auto;
  tab-size: 4;
}

.sql-editor-textarea:focus {
  background: #ffffff;
}

.sql-editor-textarea[readonly] {
  background: #f5f5f5;
  cursor: not-allowed;
  color: #606266;
}

.suggestions-popup {
  position: absolute;
  z-index: 1000;
  background: #ffffff;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  max-height: 240px;
  overflow-y: auto;
  min-width: 240px;
}

.suggestion-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  cursor: pointer;
  font-size: 13px;
  user-select: none;
}

.suggestion-item.active,
.suggestion-item:hover {
  background: #ecf5ff;
}

.s-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: 700;
  color: #ffffff;
  flex-shrink: 0;
}

.s-icon-keyword { background: #409EFF; }
.s-icon-table { background: #67C23A; }
.s-icon-column { background: #E6A23C; }
.s-icon-function { background: #9B59B6; }

.s-label {
  font-family: 'Consolas', 'Monaco', monospace;
  color: #303133;
  font-weight: 500;
  min-width: 80px;
}

.s-comment {
  margin-left: auto;
  color: #909399;
  font-size: 11px;
}
</style>
