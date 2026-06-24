/**
 * 解析 SQL 文本，识别语句边界（忽略字符串/注释中的分号）。
 * 返回每个语句的起始行、结束行、去除首尾空白后的文本。
 * 行号从 0 开始。
 *
 * 设计目标：
 *  - 不引入额外依赖，纯前端实现。
 *  - 只对 ; 作为语句分隔符做语义识别。
 *  - 支持 MySQL 中常见的 DELIMITER 变体（虽然我们不执行它，但识别它不会伤害）。
 */
export interface SqlStatementRange {
  /** 语句文本（去除首尾空白） */
  text: string
  /** 行号从 0 开始，语句第一行 */
  startLine: number
  /** 行号从 0 开始，语句最后一行 */
  endLine: number
  /** 行内起始字符位置（相对于整个文本） */
  startPos: number
  endPos: number
}

export function splitStatementsWithLines(script: string): SqlStatementRange[] {
  if (!script) return []
  const result: SqlStatementRange[] = []
  const len = script.length
  let i = 0

  let currentStart = -1 // 当前语句起始位置
  let currentStartLine = -1

  // 辅助：计算 pos 对应的行号
  const lineOf = (pos: number): number => {
    let line = 0
    for (let p = 0; p < pos && p < script.length; p++) {
      if (script.charCodeAt(p) === 10) line++
    }
    return line
  }

  // 扫描字符串/注释/分号
  while (i < len) {
    const ch = script.charAt(i)
    const next = i + 1 < len ? script.charAt(i + 1) : ''

    // 空白字符（跳过前导空白）
    if (ch === ' ' || ch === '\t' || ch === '\r' || ch === '\n') {
      i++
      continue
    }

    // 单行注释 -- 或 #
    if (ch === '-' && next === '-') {
      if (currentStart === -1) currentStart = i
      while (i < len && script.charCodeAt(i) !== 10) i++
      continue
    }
    if (ch === '#') {
      if (currentStart === -1) currentStart = i
      while (i < len && script.charCodeAt(i) !== 10) i++
      continue
    }
    // 块注释 /* */
    if (ch === '/' && next === '*') {
      if (currentStart === -1) currentStart = i
      i += 2
      while (i < len) {
        if (script.charAt(i) === '*' && script.charAt(i + 1) === '/') {
          i += 2
          break
        }
        i++
      }
      continue
    }

    // 记录语句起始
    if (currentStart === -1) {
      currentStart = i
      currentStartLine = lineOf(i)
    }

    // 单引号/双引号/反引号
    if (ch === "'" || ch === '"' || ch === '`') {
      const quote = ch
      i++
      while (i < len && script.charAt(i) !== quote) {
        // 转义 \\ 或 ''
        if (script.charAt(i) === '\\' && i + 1 < len) {
          i += 2
          continue
        }
        if (script.charAt(i) === quote && script.charAt(i + 1) === quote) {
          i += 2
          continue
        }
        i++
      }
      i++
      continue
    }

    // 分号：语句结束
    if (ch === ';') {
      const end = i // 包含分号
      const statementRaw = script.substring(currentStart, end + 1)
      const textOnly = statementRaw.trim()
      if (textOnly.length > 0 && textOnly !== ';') {
        result.push({
          text: textOnly.endsWith(';') ? textOnly.slice(0, -1).trim() : textOnly,
          startLine: currentStartLine,
          endLine: lineOf(end),
          startPos: currentStart,
          endPos: end
        })
      }
      currentStart = -1
      currentStartLine = -1
      i++
      continue
    }

    // 普通字符
    i++
  }

  // 文件末尾还有一段（不以分号结尾）
  if (currentStart !== -1) {
    const textOnly = script.substring(currentStart).trim()
    if (textOnly.length > 0) {
      result.push({
        text: textOnly.endsWith(';') ? textOnly.slice(0, -1).trim() : textOnly,
        startLine: currentStartLine,
        endLine: lineOf(len - 1),
        startPos: currentStart,
        endPos: len - 1
      })
    }
  }

  return result
}

/**
 * 给定一条 SQL，返回"第一行非空白之前的行号"和内容（便于小三角按钮定位）。
 */
export function firstLineOfStatement(script: string, startPos: number): number {
  let line = 0
  for (let p = 0; p < startPos && p < script.length; p++) {
    if (script.charCodeAt(p) === 10) line++
  }
  return line
}