export type ObjectType =
  | 'database'
  | 'group'
  | 'table'
  | 'view'
  | 'function'
  | 'procedure'
  | 'trigger'
  | 'column'
  | 'index'
  | 'query-save'
  | 'saved-query'
  | string

export interface TreeNode {
  type: ObjectType
  name: string
  database?: string
  table?: string
  group?: string
  children?: TreeNode[]
  rows?: number
  sizeMb?: number
  pk?: boolean
  colType?: string
  comment?: string
}

export interface TreeCallbacks {
  onNodeDblClick: (node: TreeNode) => void
  onNodeCtxMenu: (e: MouseEvent, node: TreeNode) => void
  onNodeInfoClick: (node: TreeNode) => void
}

export interface DatasourceSummary {
  datasourceId: string
  name: string
  dbType: string
  host?: string
  port?: number
}

export interface QueryTabState {
  id: string
  kind: 'query'
  title: string
  datasourceId: string
  database: string
  sql: string
  result: QueryResultState
  history: QueryHistoryItem[]
  savedQueryId?: string
}

export interface TableTabState {
  id: string
  kind: 'table'
  title: string
  datasourceId: string
  database: string
  table: string
  columns: string[]
  rows: Record<string, any>[]
  tableLoading: boolean
}

export type AnyTabState = QueryTabState | TableTabState

export interface QueryResultState {
  activeName: 'result' | 'explain' | 'message' | 'history'
  activeResultIdx: string
  statementResults: StatementResult[]
  explain: Record<string, any>[]
  explainCols: string[]
  messages: QueryMessage[]
}

export interface StatementResult {
  sql: string
  isSelect: boolean
  columns: string[]
  rows: Record<string, any>[]
  affectedRows: number
  durationMs: number
  success: boolean
  message: string
  effectiveDatabase: string
}

export interface QueryMessage {
  time: string
  level: 'INFO' | 'WARN' | 'ERROR' | string
  text: string
}

export interface QueryHistoryItem {
  time: string
  sql: string
  database: string
  success: boolean
  durationMs?: number
  status?: string
}

export interface ContextMenuItem {
  key: string
  label: string
  icon?: string
  danger?: boolean
  separator?: boolean
}
