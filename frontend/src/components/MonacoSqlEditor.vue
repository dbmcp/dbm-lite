<template>
  <div class="monaco-sql-editor" :style="{ height }">
    <div v-if="!editorReady" class="monaco-loader">
      <div class="monaco-loader__spinner"></div>
      <div class="monaco-loader__text">{{ loadError ? '编辑器资源加载失败' : '正在加载 SQL 编辑器...' }}</div>
      <el-button v-if="loadError" type="primary" size="small" @click="retryInternal">重试</el-button>
    </div>
    <div ref="editorEl" class="monaco-sql-editor__container"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch as watchProp } from 'vue'

interface StatementRange {
  startLineNumber: number
  endLineNumber: number
  startColumn: number
  endColumn: number
  text: string
}

interface MonacoEditorApi {
  getValue: () => string
  setValue: (val: string) => void
  layout: () => void
  focus: () => void
  getSelections: () => Array<{ startLineNumber: number; endLineNumber: number; startColumn: number; endColumn: number }>
  getStatementRanges: () => StatementRange[]
  raw: () => any
}

const props = withDefaults(
  defineProps<{
    modelValue?: string
    height?: string | number
    fontSize?: number
    theme?: 'vs' | 'vs-dark'
    completionDatabases?: string[]
    completionTables?: string[]
    completionColumns?: string[]
  }>(),
  {
    modelValue: '',
    height: '100%',
    fontSize: 14,
    theme: 'vs',
    completionDatabases: () => [],
    completionTables: () => [],
    completionColumns: () => []
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', val: string): void
  (e: 'editor-ready', api: MonacoEditorApi): void
  (e: 'change', val: string): void
  (e: 'execute-statement', text: string, range: { startLineNumber: number; endLineNumber: number }): void
}>()

const editorEl = ref<HTMLElement | null>(null)
const editorReady = ref<boolean>(false)
const loadError = ref<boolean>(false)
let editor: any = null
let monaco: any = null
let widgets: any[] = []

function parseStatementRanges(script: string): StatementRange[] {
  const result: StatementRange[] = []
  if (!script) return result
  const len = script.length
  let pos = 0
  let curStart = -1
  let curStartLine = 1
  let curStartCol = 1
  let curLine = 1
  let curCol = 1
  const advanceLineCol = (ch: string) => {
    if (ch === '\n') { curLine++; curCol = 1 } else { curCol++ }
  }
  const flush = (endPos: number, endLine: number, endCol: number) => {
    if (curStart === -1) return
    const raw = script.substring(curStart, endPos)
    const text = raw.trim().replace(/;+$/, '').trim()
    if (text.length === 0) { curStart = -1; return }
    result.push({
      startLineNumber: curStartLine,
      endLineNumber: endLine,
      startColumn: curStartCol,
      endColumn: endCol,
      text
    })
    curStart = -1
  }
  while (pos < len) {
    const ch = script.charAt(pos)
    const next = pos + 1 < len ? script.charAt(pos + 1) : ''
    if (ch === '\r') { pos++; continue }
    if (ch === ' ' || ch === '\t' || ch === '\n') {
      advanceLineCol(ch); pos++; continue
    }
    if (ch === '-' && next === '-') {
      while (pos < len && script.charAt(pos) !== '\n') { advanceLineCol(script.charAt(pos)); pos++ }
      continue
    }
    if (ch === '#') {
      while (pos < len && script.charAt(pos) !== '\n') { advanceLineCol(script.charAt(pos)); pos++ }
      continue
    }
    if (ch === '/' && next === '*') {
      pos += 2
      while (pos < len) {
        if (script.charAt(pos) === '*' && script.charAt(pos + 1) === '/') { pos += 2; break }
        advanceLineCol(script.charAt(pos)); pos++
      }
      continue
    }
    if (curStart === -1) { curStart = pos; curStartLine = curLine; curStartCol = curCol }
    if (ch === "'" || ch === '"' || ch === '`') {
      const q = ch
      advanceLineCol(ch); pos++
      while (pos < len && script.charAt(pos) !== q) {
        if (script.charAt(pos) === '\\' && pos + 1 < len) { advanceLineCol(script.charAt(pos)); pos++; advanceLineCol(script.charAt(pos)); pos++; continue }
        if (script.charAt(pos) === q && script.charAt(pos + 1) === q) { advanceLineCol(script.charAt(pos)); pos++; advanceLineCol(script.charAt(pos)); pos++; continue }
        advanceLineCol(script.charAt(pos)); pos++
      }
      advanceLineCol(ch); pos++
      continue
    }
    if (ch === ';') {
      flush(pos + 1, curLine, curCol + 1)
      pos++; advanceLineCol(ch)
      continue
    }
    advanceLineCol(ch); pos++
  }
  if (curStart !== -1) {
    flush(pos, curLine, curCol)
  }
  return result
}

function clearWidgets() {
  if (!editor) return
  widgets.forEach((w) => { try { editor.removeContentWidget(w) } catch (_) { } })
  widgets = []
}

let lastDecorationIds: string[] = []

function renderStatementRunButtons() {
  if (!editor || !monaco) return
  clearWidgets()
  const text = editor.getValue() as string
  const ranges = parseStatementRanges(text)
  const seen = new Set<number>()
  const newDecorations: any[] = []
  ranges.forEach((r, idx) => {
    if (seen.has(r.startLineNumber)) return
    seen.add(r.startLineNumber)
    const dom = document.createElement('div')
    dom.className = 'dbm-run-gutter-btn'
    dom.title = '执行该 SQL 语句'
    dom.textContent = '▶'
    dom.onclick = (ev) => {
      ev.stopPropagation()
      emit('execute-statement', r.text, { startLineNumber: r.startLineNumber, endLineNumber: r.endLineNumber })
    }
    const widget = {
      getId: () => 'run-gutter-' + r.startLineNumber + '-' + idx,
      getDomNode: () => dom,
      getPosition: () => ({
        position: new monaco.Position(r.startLineNumber, 1),
        preference: [monaco.editor.ContentWidgetPositionPreference.LEFT]
      })
    }
    try { editor.addContentWidget(widget); widgets.push(widget) } catch (_) { }
    // 同时在行首的 lineDecorations 上标记一个可见的执行图标区（glyphMargin）
    try {
      newDecorations.push({
        range: new monaco.Range(r.startLineNumber, 1, r.startLineNumber, 1),
        options: {
          isWholeLine: false,
          linesDecorationsClassName: 'dbm-run-decoration-line'
        }
      })
    } catch (_) { }
  })
  try {
    lastDecorationIds = editor.deltaDecorations(lastDecorationIds, newDecorations)
  } catch (_) { }
  try { editor.layoutContentWidgets() } catch (_) { }
}

function registerCompletions(m: any) {
  const keywords = [
    'SELECT', 'FROM', 'WHERE', 'AND', 'OR', 'NOT', 'NULL', 'IN', 'LIKE', 'IS', 'AS',
    'ORDER BY', 'GROUP BY', 'HAVING', 'LIMIT', 'OFFSET', 'UNION', 'UNION ALL',
    'INSERT INTO', 'VALUES', 'UPDATE', 'SET', 'DELETE FROM', 'CREATE TABLE', 'ALTER TABLE',
    'DROP TABLE', 'TRUNCATE', 'CREATE INDEX', 'CREATE DATABASE', 'USE', 'SHOW DATABASES',
    'SHOW TABLES', 'DESCRIBE', 'EXPLAIN', 'BEGIN', 'COMMIT', 'ROLLBACK', 'START TRANSACTION',
    'PRIMARY KEY', 'FOREIGN KEY', 'REFERENCES', 'UNIQUE', 'DEFAULT', 'AUTO_INCREMENT',
    'INNER JOIN', 'LEFT JOIN', 'RIGHT JOIN', 'JOIN', 'ON', 'DISTINCT', 'COUNT', 'SUM',
    'AVG', 'MIN', 'MAX', 'BETWEEN', 'EXISTS', 'CASE', 'WHEN', 'THEN', 'ELSE', 'END',
    'IFNULL', 'COALESCE', 'CAST', 'CONVERT', 'VARCHAR', 'INT', 'BIGINT', 'INTEGER',
    'TEXT', 'DATE', 'DATETIME', 'TIMESTAMP', 'FLOAT', 'DOUBLE', 'DECIMAL', 'BOOLEAN',
    'TRUE', 'FALSE', 'CHAR', 'NVARCHAR', 'NCHAR'
  ]
  m.languages.registerCompletionItemProvider('sql', {
    triggerCharacters: [' ', '.', '('],
    provideCompletionItems: () => {
      const suggestions: any[] = []
      keywords.forEach((kw) => suggestions.push({ label: kw, kind: 17, insertText: kw, detail: 'SQL 关键字' }))
      ;(props.completionDatabases || []).forEach((db) => suggestions.push({ label: db, kind: 19, insertText: db, detail: '数据库' }))
      ;(props.completionTables || []).forEach((tbl) => suggestions.push({ label: tbl, kind: 5, insertText: tbl, detail: '表' }))
      ;(props.completionColumns || []).forEach((col) => suggestions.push({ label: col, kind: 10, insertText: col, detail: '列' }))
      return { suggestions }
    }
  })
}

function buildApi(): MonacoEditorApi {
  return {
    getValue: () => (editor ? editor.getValue() : ''),
    setValue: (v: string) => editor?.setValue(v ?? ''),
    layout: () => editor?.layout(),
    focus: () => editor?.focus(),
    getSelections: () => editor?.getSelections?.() || [],
    getStatementRanges: () => (editor ? parseStatementRanges(editor.getValue()) : []),
    raw: () => editor
  }
}

async function loadMonaco() {
  if (monaco) return
  try {
    const m = await import(/* @vite-ignore */ 'monaco-editor')
    monaco = m
    registerCompletions(m)
  } catch (err) {
    console.error('[MonacoSqlEditor] load failed:', err)
    loadError.value = true
  }
}

// MonacoEnvironment 配置：使用本地 worker，避免跨域问题
function createEditorWorker() {
  // 使用 Vite 的 ?worker 语法创建一个空的 editor worker（通过 Blob）
  const code = `
    self.MonacoEnvironment = { baseUrl: '/' };
    importScripts('/node_modules/.vite/deps/monaco-editor_esm_vs_editor_editor.worker.js');
  `
  try {
    const blob = new Blob([code], { type: 'application/javascript' })
    return new Worker(URL.createObjectURL(blob), { name: 'editor' })
  } catch (e) {
    console.warn('[MonacoSqlEditor] Worker 创建失败，fallback 到主线程:', e)
    return undefined as any
  }
}

if (typeof (window as any).MonacoEnvironment === 'undefined') {
  ;(window as any).MonacoEnvironment = {
    getWorker: function (_moduleId: string, label: string): Worker | undefined {
      try {
        // 通过动态 import 加载 worker（Vite 支持 ?worker 语法）
        // 为了简单起见，我们直接在主线程工作（monaco 支持此回退）
        // 但为了更好的性能，我们通过 Blob 创建一个简单的本地 worker
        const workerFiles: Record<string, string> = {
          json: 'monaco-editor/esm/vs/language/json/json.worker.js',
          css: 'monaco-editor/esm/vs/language/css/css.worker.js',
          scss: 'monaco-editor/esm/vs/language/css/css.worker.js',
          less: 'monaco-editor/esm/vs/language/css/css.worker.js',
          html: 'monaco-editor/esm/vs/language/html/html.worker.js',
          handlebars: 'monaco-editor/esm/vs/language/html/html.worker.js',
          razor: 'monaco-editor/esm/vs/language/html/html.worker.js',
          typescript: 'monaco-editor/esm/vs/language/typescript/ts.worker.js',
          javascript: 'monaco-editor/esm/vs/language/typescript/ts.worker.js'
        }
        const workerPath = workerFiles[label] || 'monaco-editor/esm/vs/editor/editor.worker.js'
        const blobContent = `
          self.MonacoEnvironment = { baseUrl: '/' };
          try {
            self.importScripts('/' + '${workerPath}');
          } catch (err) {
            // importScripts failed - worker will be silent
          }
        `
        const blob = new Blob([blobContent], { type: 'application/javascript' })
        return new Worker(URL.createObjectURL(blob), { name: label })
      } catch (err) {
        // 回退：不提供 worker，monaco 会在主线程工作（会有警告但不影响功能）
        console.warn('[MonacoSqlEditor] Worker 回退到主线程:', err)
        return undefined
      }
    }
  }
}

async function createEditor() {
  if (!monaco || !editorEl.value) return
  editor = monaco.editor.create(editorEl.value, {
    value: props.modelValue || '',
    language: 'sql',
    theme: props.theme || 'vs',
    automaticLayout: true,
    minimap: { enabled: false },
    fontSize: props.fontSize || 14,
    scrollBeyondLastLine: false,
    renderLineHighlight: 'line',
    bracketPairColorization: { enabled: true },
    folding: true,
    tabSize: 4,
    insertSpaces: false,
    wordWrap: 'on',
    fontFamily: 'Consolas, "Courier New", monospace',
    lineNumbers: 'on',
    glyphMargin: true
  })
  editor.onDidChangeModelContent(() => {
    const val = editor.getValue()
    emit('update:modelValue', val)
    emit('change', val)
    renderStatementRunButtons()
  })
  editor.onDidChangeModel(() => renderStatementRunButtons())
  editor.onDidScrollChange(() => renderStatementRunButtons())
  renderStatementRunButtons()
  editorReady.value = true
  // 把 editor API 暴露到 DOM 元素上，方便浏览器测试工具访问
  if (editorEl.value) {
    ;(editorEl.value as any).__monacoEditorApi = buildApi()
  }
  emit('editor-ready', buildApi())
}

async function retryInternal() {
  loadError.value = false
  editorReady.value = false
  await loadMonaco()
  await new Promise((res) => requestAnimationFrame(() => res(null)))
  createEditor()
}

onMounted(async () => {
  try {
    await loadMonaco()
    await new Promise((res) => requestAnimationFrame(() => res(null)))
    createEditor()
  } catch (err) {
    console.error('[MonacoSqlEditor] init failed:', err)
    loadError.value = true
  }
})

onBeforeUnmount(() => {
  clearWidgets()
  if (editor) { try { editor.dispose() } catch (_) { } editor = null }
})

watchProp(() => props.theme, (val) => {
  if (monaco) { try { monaco.editor.setTheme(val || 'vs') } catch (_) { } }
})

watchProp(() => props.modelValue, (val) => {
  if (editor && editor.getValue() !== val) editor.setValue(val ?? '')
})

</script>

<style scoped>
.monaco-sql-editor {
  position: relative;
  min-height: 200px;
  width: 100%;
}
.monaco-sql-editor__container {
  position: absolute;
  inset: 0;
}
.monaco-sql-editor :deep(.dbm-run-gutter-btn) {
  width: 18px;
  height: 18px;
  margin-top: 2px;
  margin-left: -24px;
  color: #409eff;
  cursor: pointer;
  font-size: 13px;
  line-height: 18px;
  text-align: center;
  border-radius: 4px;
  user-select: none;
  font-weight: 700;
}
.monaco-sql-editor :deep(.dbm-run-gutter-btn:hover) {
  background-color: #ecf5ff;
  color: #1f6feb;
}
.monaco-sql-editor :deep(.dbm-run-decoration-line) { /* reserved */ }
.monaco-loader {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  background: #fafafa;
  color: #606266;
  font-size: 13px;
  flex-direction: column;
  z-index: 10;
}
.monaco-loader__spinner {
  width: 24px;
  height: 24px;
  border: 2px solid #dcdfe6;
  border-top-color: #409eff;
  border-radius: 50%;
  animation: monaco-spin 0.9s linear infinite;
}
@keyframes monaco-spin { to { transform: rotate(360deg); } }
</style>
