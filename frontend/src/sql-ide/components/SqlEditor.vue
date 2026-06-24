<template>
  <div class="sql-editor-container" ref="containerRef">
    <div ref="editorRef" class="sql-editor"></div>
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
import { ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import CodeMirror from 'codemirror'
import 'codemirror/lib/codemirror.css'
import 'codemirror/theme/eclipse.css'
import 'codemirror/mode/sql/sql'

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

const containerRef = ref<HTMLElement | null>(null)
const editorRef = ref<HTMLElement | null>(null)
let cm: CodeMirror.Editor | null = null

const suggestions = ref<Suggestion[]>([])
const selectedIndex = ref(0)
const popupPos = ref<{ top: number; left: number }>({ top: 0, left: 0 })

function iconFor(type: string): string {
  switch (type) {
    case 'keyword': return 'K'
    case 'table': return 'T'
    case 'column': return 'C'
    case 'function': return 'F'
    default: return '?'
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
  if (!cm || props.readonly) {
    suggestions.value = []
    return
  }
  const pos = cm.getCursor()
  const text = cm.getValue()
  const doc = cm.getDoc()
  const cursorPos = doc.indexFromPos(pos)
  const items = computeSuggestions(text, cursorPos)
  if (items.length === 0) {
    suggestions.value = []
    return
  }
  const coords = cm.cursorCoords(pos)
  popupPos.value = { top: coords.bottom + 5, left: coords.left }
  suggestions.value = items
  selectedIndex.value = 0
}

function applySuggestion(item: Suggestion) {
  if (!cm) return
  const pos = cm.getCursor()
  const text = cm.getValue()
  const doc = cm.getDoc()
  const cursorPos = doc.indexFromPos(pos)
  const before = text.substring(0, cursorPos)
  const after = text.substring(cursorPos)
  const match = before.match(/[A-Za-z_][A-Za-z0-9_]*$/)
  const startPos = match ? before.length - match[0].length : cursorPos
  const newText = text.substring(0, startPos) + item.insertText + after
  emit('input', newText)
  suggestions.value = []
  nextTick(() => {
    const newPosObj = doc.posFromIndex(startPos + item.insertText.length)
    cm!.setCursor(newPosObj)
    cm!.focus()
  })
}

function handleKeydown(cm: CodeMirror.Editor, e: KeyboardEvent) {
  if (e.ctrlKey && e.key === 'Enter') {
    e.preventDefault()
    e.stopPropagation()
    emit('run-command')
    return
  }
  if (e.ctrlKey && e.shiftKey && e.key.toUpperCase() === 'F') {
    e.preventDefault()
    e.stopPropagation()
    emit('beautify')
    return
  }

  if (suggestions.value.length > 0) {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      e.stopPropagation()
      selectedIndex.value = (selectedIndex.value + 1) % suggestions.value.length
      return
    }
    if (e.key === 'ArrowUp') {
      e.preventDefault()
      e.stopPropagation()
      selectedIndex.value = (selectedIndex.value - 1 + suggestions.value.length) % suggestions.value.length
      return
    }
    if (e.key === 'Tab') {
      e.preventDefault()
      e.stopPropagation()
      if (selectedIndex.value >= 0 && selectedIndex.value < suggestions.value.length) {
        applySuggestion(suggestions.value[selectedIndex.value])
      }
      return
    }
    if (e.key === 'Enter') {
      if (selectedIndex.value >= 0 && selectedIndex.value < suggestions.value.length) {
        e.preventDefault()
        e.stopPropagation()
        applySuggestion(suggestions.value[selectedIndex.value])
        return
      }
    }
    if (e.key === 'Escape') {
      e.preventDefault()
      e.stopPropagation()
      suggestions.value = []
      return
    }
  }
}

watch(() => props.value, (newValue) => {
  if (cm && newValue !== cm.getValue()) {
    const cursor = cm.getCursor()
    cm.setValue(newValue)
    cm.setCursor(cursor)
  }
})

onMounted(() => {
  nextTick(() => {
    if (editorRef.value) {
      cm = CodeMirror(editorRef.value, {
        mode: { name: 'sql', dialect: 'mysql' },
        theme: 'eclipse',
        lineNumbers: false,
        indentUnit: 2,
        tabSize: 4,
        indentWithTabs: false,
        readOnly: props.readonly,
        lineWrapping: true,
        cursorBlinkRate: 530,
        styleSelectedText: true,
        matchBrackets: true,
        autoCloseBrackets: true,
        highlightSelectionMatches: { showToken: /\w/, annotateScrollbar: true }
      })

      cm.setValue(props.value)

      cm.on('change', () => {
        emit('input', cm!.getValue())
        updateSuggestions()
      })

      cm.on('blur', () => {
        suggestions.value = []
      })

      cm.on('focus', () => {
        updateSuggestions()
      })

      cm.on('keydown', (instance: CodeMirror.Editor, e: KeyboardEvent) => {
        handleKeydown(instance, e)
      })

      if (props.tabId) {
        const externalListener = (e: Event) => {
          const ce = e as CustomEvent
          if (ce.detail && cm) {
            cm.setValue(ce.detail)
            emit('input', ce.detail)
          }
        }
        window.addEventListener('sqlide:update-editor-' + props.tabId, externalListener)
        onBeforeUnmount(() => {
          window.removeEventListener('sqlide:update-editor-' + props.tabId, externalListener)
        })
      }
    }
  })
})

onBeforeUnmount(() => {
  if (cm) {
    cm.toTextArea()
    cm = null
  }
})

const popupStyle = {
  position: 'absolute' as const,
  top: `${popupPos.value.top}px`,
  left: `${popupPos.value.left}px`
}

function getSelectedSQL(): string {
  if (!cm) return props.value
  const selection = cm.getSelection()
  if (selection && selection.trim()) {
    return selection.trim()
  }
  return props.value
}

defineExpose({
  restoreSelection: () => {},
  getSelectedSQL
})
</script>

<style scoped>
.sql-editor-container {
  width: 100%;
  height: 100%;
  min-height: 120px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
  position: relative;
  background-color: #ffffff;
}

.sql-editor {
  width: 100%;
  height: 100%;
  min-height: 120px;
}

:deep(.CodeMirror) {
  width: 100%;
  height: 100%;
  min-height: 120px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.8;
  background-color: #ffffff !important;
}

:deep(.CodeMirror-focused) {
  outline: none;
  border: 1px solid #409EFF !important;
}

:deep(.CodeMirror-cursor) {
  border-left: 2px solid #303133 !important;
}

:deep(.CodeMirror-selected) {
  background-color: rgba(64, 158, 255, 0.3) !important;
  color: #000000 !important;
}

:deep(.CodeMirror-line) {
  color: #000000 !important;
}

:deep(.cm-keyword) {
  color: #409EFF !important;
  font-weight: bold !important;
}

:deep(.cm-string) {
  color: #67C23A !important;
}

:deep(.cm-comment) {
  color: #909399 !important;
  font-style: italic !important;
}

:deep(.cm-number) {
  color: #E6A23C !important;
}

:deep(.cm-builtin) {
  color: #9B59B6 !important;
  font-weight: 500 !important;
}

:deep(.cm-variable-2) {
  color: #000000 !important;
}

:deep(.cm-property) {
  color: #000000 !important;
}

:deep(.cm-variable) {
  color: #000000 !important;
}

:deep(.cm-operator) {
  color: #000000 !important;
}

:deep(.cm-punctuation) {
  color: #000000 !important;
}

:deep(.cm-bracket) {
  color: #000000 !important;
}

:deep(.cm-tag) {
  color: #000000 !important;
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
