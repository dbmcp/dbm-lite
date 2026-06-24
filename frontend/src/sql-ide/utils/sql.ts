import type { TreeNode } from '../types'
import { escapeIdentifier } from './tree'
export { escapeIdentifier }

interface DatabaseAndReferences {
  explicitDatabase?: string
  references: string[]
}

/**
 * 从一条 SQL 中提取显式的数据库名（如 `FROM db.table`）。
 * 当 SQL 中出现多数据库引用时返回第一个作为显式数据库。
 * 同时返回所有引用的 `db.table` / `table` 片段列表。
 */
export function detectExplicitDatabase(sql: string, candidateDatabases: string[]): DatabaseAndReferences {
  const references: string[] = []
  if (!sql) return { references }
  const normalized = sql.replace(/(--|#)[^\n]*/g, '').replace(/\/\*[\s\S]*?\*\//g, '')
  const dbCandidates = (candidateDatabases || []).map((d) => d.toLowerCase())
  const twoPartPattern = /`?([a-zA-Z_][\w$]*)`?\s*\.\s*`?([a-zA-Z_][\w$]*)`?/g
  let m: RegExpExecArray | null
  let firstExplicit: string | undefined
  while ((m = twoPartPattern.exec(normalized)) !== null) {
    const rawDb = m[1]
    const rawTable = m[2]
    references.push(rawDb + '.' + rawTable)
    if (!firstExplicit && dbCandidates.indexOf(rawDb.toLowerCase()) !== -1) {
      firstExplicit = rawDb
    } else if (!firstExplicit) {
      firstExplicit = rawDb
    }
  }
  const onePartPattern = /\b(?:FROM|JOIN|UPDATE|INSERT\s+INTO|TABLE)\s+`?([a-zA-Z_][\w$]*)`?/gi
  while ((m = onePartPattern.exec(normalized)) !== null) {
    references.push(m[1])
  }
  return { explicitDatabase: firstExplicit, references }
}

/**
 * 选择最终生效数据库的优先级：
 *  1. SQL 中显式指定的数据库（当存在 db.table 形式时）
 *  2. 当前 Tab 页下拉选择的数据库
 *  3. 回退到数据源默认数据库（空字符串交由后端处理）
 */
export function resolveEffectiveDatabase(
  sql: string,
  tabDatabase: string,
  candidateDatabases: string[]
): string {
  const { explicitDatabase } = detectExplicitDatabase(sql, candidateDatabases)
  if (explicitDatabase) return explicitDatabase
  if (tabDatabase) return tabDatabase
  return ''
}

export function nodeTargetStatement(node: TreeNode): string {
  const db = node.database || ''
  const table = node.table || node.name
  const prefix = db ? escapeIdentifier(db) + '.' : ''
  if (node.type === 'table' || node.type === 'view') return prefix + escapeIdentifier(table)
  if (node.type === 'function' || node.type === 'procedure') return prefix + escapeIdentifier(table) + '()'
  return prefix + escapeIdentifier(table)
}

export function formatNumber(n?: number): string {
  if (typeof n !== 'number') return ''
  return n.toLocaleString()
}

export function pad2(n: number): string {
  return n < 10 ? '0' + n : String(n)
}

export function nowString(): string {
  const d = new Date()
  return (
    d.getFullYear() +
    '-' +
    pad2(d.getMonth() + 1) +
    '-' +
    pad2(d.getDate()) +
    ' ' +
    pad2(d.getHours()) +
    ':' +
    pad2(d.getMinutes()) +
    ':' +
    pad2(d.getSeconds())
  )
}

export function stringifyCell(v: any): string {
  if (v === null || v === undefined) return 'NULL'
  if (typeof v === 'object') {
    try {
      return JSON.stringify(v)
    } catch {
      return String(v)
    }
  }
  return String(v)
}

export function copyToClipboard(text: string): Promise<void> {
  if (typeof navigator !== 'undefined' && navigator.clipboard && (navigator.clipboard as any).writeText) {
    return (navigator.clipboard as any).writeText(text)
  }
  try {
    const ta = document.createElement('textarea')
    ta.value = text
    ta.style.position = 'fixed'
    ta.style.left = '-9999px'
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    document.body.removeChild(ta)
    return Promise.resolve()
  } catch (e: any) {
    return Promise.reject(e)
  }
}

const SQL_KW = new Set([
  'SELECT', 'FROM', 'WHERE', 'AND', 'OR', 'NOT', 'IN', 'IS', 'NULL', 'AS',
  'GROUP', 'BY', 'ORDER', 'HAVING', 'LIMIT', 'OFFSET', 'ASC', 'DESC',
  'INSERT', 'INTO', 'VALUES', 'UPDATE', 'SET', 'DELETE', 'CREATE', 'TABLE',
  'VIEW', 'INDEX', 'SCHEMA', 'DATABASE', 'DROP', 'ALTER', 'ADD', 'COLUMN',
  'PRIMARY', 'KEY', 'FOREIGN', 'REFERENCES', 'UNIQUE', 'DEFAULT', 'CHECK',
  'AUTO_INCREMENT', 'DISTINCT', 'ALL', 'UNION', 'EXISTS', 'BETWEEN', 'LIKE',
  'CASE', 'WHEN', 'THEN', 'ELSE', 'END', 'JOIN', 'LEFT', 'RIGHT', 'INNER',
  'OUTER', 'FULL', 'CROSS', 'NATURAL', 'ON', 'USING', 'WITH', 'RECURSIVE',
  'EXPLAIN', 'DESCRIBE', 'SHOW', 'BEGIN', 'COMMIT', 'ROLLBACK', 'TRANSACTION',
  'GRANT', 'REVOKE', 'TRUNCATE', 'REPLACE', 'IF', 'COUNT', 'SUM', 'AVG',
  'MIN', 'MAX', 'ROUND', 'CAST', 'CONVERT', 'CONCAT', 'SUBSTRING', 'LENGTH',
  'LOWER', 'UPPER', 'NOW', 'CURDATE', 'CURRENT_DATE', 'CURRENT_TIMESTAMP'
])

const NEWLINE_KW = new Set([
  'SELECT', 'FROM', 'WHERE', 'GROUP', 'ORDER', 'HAVING', 'LIMIT', 'OFFSET',
  'INSERT', 'VALUES', 'UPDATE', 'DELETE', 'JOIN', 'LEFT', 'RIGHT', 'INNER',
  'OUTER', 'FULL', 'UNION', 'WITH', 'EXPLAIN', 'SET', 'ON'
])

/**
 * 轻量级 SQL 美化：关键字大写、子句换行、缩进分层。
 * 不依赖任何三方库，严格对标 Navicat 风格。
 */
export function beautifySql(sql: string): string {
  if (!sql) return ''
  // 先分词：保留字符串/注释/数字/关键字/符号
  const tokens: string[] = []
  let i = 0
  const n = sql.length
  while (i < n) {
    const ch = sql.charCodeAt(i)
    // 单行注释
    if (ch === 45 && sql.charCodeAt(i + 1) === 45) { // --
      let j = i
      while (j < n && sql.charCodeAt(j) !== 10) j++
      tokens.push(sql.substring(i, j))
      i = j
      continue
    }
    // 块注释
    if (ch === 47 && sql.charCodeAt(i + 1) === 42) { // /*
      let j = i + 2
      while (j < n - 1 && !(sql.charCodeAt(j) === 42 && sql.charCodeAt(j + 1) === 47)) j++
      j = Math.min(n, j + 2)
      tokens.push(sql.substring(i, j))
      i = j
      continue
    }
    // 字符串
    if (ch === 39 || ch === 34) { // ' or "
      const q = ch
      let j = i + 1
      while (j < n) {
        if (sql.charCodeAt(j) === q) { j++; break }
        if (sql.charCodeAt(j) === 92 && j + 1 < n) { j += 2; continue } // \ escape
        j++
      }
      tokens.push(sql.substring(i, j))
      i = j
      continue
    }
    // 反引号标识符
    if (ch === 96) {
      let j = i + 1
      while (j < n && sql.charCodeAt(j) !== 96) j++
      j = Math.min(n, j + 1)
      tokens.push(sql.substring(i, j))
      i = j
      continue
    }
    // 空白
    if (ch === 32 || ch === 9 || ch === 10 || ch === 13) {
      let j = i + 1
      while (j < n) {
        const c2 = sql.charCodeAt(j)
        if (c2 === 32 || c2 === 9 || c2 === 10 || c2 === 13) j++; else break
      }
      tokens.push(' ')
      i = j
      continue
    }
    // 数字
    if (ch >= 48 && ch <= 57) {
      let j = i + 1
      while (j < n) {
        const c2 = sql.charCodeAt(j)
        if ((c2 >= 48 && c2 <= 57) || c2 === 46) j++; else break
      }
      tokens.push(sql.substring(i, j))
      i = j
      continue
    }
    // 字母/下划线 -> 关键字或标识符
    if ((ch >= 65 && ch <= 90) || (ch >= 97 && ch <= 122) || ch === 95) {
      let j = i + 1
      while (j < n) {
        const c2 = sql.charCodeAt(j)
        if ((c2 >= 65 && c2 <= 90) || (c2 >= 97 && c2 <= 122) || (c2 >= 48 && c2 <= 57) || c2 === 95) j++; else break
      }
      tokens.push(sql.substring(i, j))
      i = j
      continue
    }
    // 单独符号
    tokens.push(sql.charAt(i))
    i++
  }

  // 组装：根据关键字换行 + 缩进
  const out: string[] = []
  let depth = 0
  const newlineAndIndent = () => {
    out.push('\n' + '  '.repeat(Math.max(0, depth)))
  }
  for (let k = 0; k < tokens.length; k++) {
    const tk = tokens[k]
    if (tk === ' ' || tk === '\n' || tk === '\t') {
      // 折叠多余空格
      if (out.length > 0 && out[out.length - 1] !== ' ' && out[out.length - 1] !== '\n' && !out[out.length - 1].endsWith('\n')) {
        out.push(' ')
      }
      continue
    }
    const upper = tk.toUpperCase()
    if (tk === '(') {
      // 左括号：深度 +1；若紧跟 SELECT 则换行
      const nextNonSpace = (() => {
        for (let m = k + 1; m < tokens.length; m++) {
          if (tokens[m] !== ' ' && tokens[m] !== '\n' && tokens[m] !== '\t') return tokens[m].toUpperCase()
        }
        return ''
      })()
      if (out.length > 0 && !out[out.length - 1].endsWith('\n') && ![' ', '\n'].includes(out[out.length - 1])) {
        // 去除末端空格
        while (out.length > 0 && out[out.length - 1] === ' ') out.pop()
      }
      out.push('(')
      if (nextNonSpace === 'SELECT') {
        depth += 1
        newlineAndIndent()
      }
      continue
    }
    if (tk === ')') {
      depth = Math.max(0, depth - 1)
      // 移除末端多余空格
      while (out.length > 0 && out[out.length - 1] === ' ') out.pop()
      out.push(')')
      continue
    }
    if (tk === ',') {
      out.push(',')
      newlineAndIndent()
      continue
    }
    if (tk === ';') {
      while (out.length > 0 && out[out.length - 1] === ' ') out.pop()
      out.push(';')
      depth = 0
      out.push('\n')
      continue
    }
    if (SQL_KW.has(upper)) {
      if (NEWLINE_KW.has(upper)) {
        while (out.length > 0 && (out[out.length - 1] === ' ' || out[out.length - 1] === '\n')) out.pop()
        if (out.length > 0 && !out[out.length - 1].endsWith('\n')) newlineAndIndent()
        out.push(upper + ' ')
        continue
      }
      out.push(upper + ' ')
      continue
    }
    // 普通 token
    out.push(tk)
  }

  // 合并结果并清理空行
  const text = out.join('').replace(/[ \t]+\n/g, '\n').replace(/\n{2,}/g, '\n').trim()
  return text
}

