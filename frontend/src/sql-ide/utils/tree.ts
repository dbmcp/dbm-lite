import type { TreeNode } from '../types'

const SINGULAR_MAP: Record<string, string> = {
  tables: 'table',
  views: 'view',
  functions: 'function',
  procedures: 'procedure',
  triggers: 'trigger',
  indexes: 'index',
  columns: 'column'
}

const BUCKET_ORDER = ['tables', 'views', 'functions', 'procedures', 'triggers', 'indexes']

const BUCKET_LABELS: Record<string, string> = {
  tables: '表',
  views: '视图',
  functions: '函数',
  procedures: '存储过程',
  triggers: '触发器',
  indexes: '索引'
}

export function singularType(t: string): string {
  if (!t) return 'table'
  const low = t.toLowerCase()
  return SINGULAR_MAP[low] || low
}

export function pluralType(t: string): string {
  const sing = singularType(t)
  if (sing === 'index') return 'indexes'
  return sing + 's'
}

export function normalizeNode(raw: any, parentDb: string): TreeNode | undefined {
  if (!raw) return undefined
  const type = (raw.type || raw.objType || raw.kind || 'table').toString().toLowerCase()
  const db = raw.database || raw.db || parentDb
  const name = raw.name || raw.table || raw.objectName || String(raw)
  const children: TreeNode[] = []
  if (Array.isArray(raw.children) && raw.children.length > 0) {
    for (const c of raw.children) {
      const n = normalizeNode(c, db)
      if (n) children.push(n)
    }
  }
  const rowCount = typeof raw.rows === 'number' ? raw.rows : 
                  (raw.rows !== undefined && !isNaN(parseInt(String(raw.rows))) ? parseInt(String(raw.rows)) : undefined)
  const size = typeof raw.sizeMb === 'number' ? raw.sizeMb : 
               (raw.sizeMb !== undefined && !isNaN(parseFloat(String(raw.sizeMb))) ? parseFloat(String(raw.sizeMb)) : undefined)
  
  const node: TreeNode = {
    type: singularType(type),
    name,
    database: db,
    table: raw.table || (type === 'table' || type === 'view' ? name : undefined),
    children: children.length > 0 ? children : undefined,
    rows: rowCount,
    sizeMb: size,
    comment: raw.comment || raw.remark || undefined
  }
  if (raw.columns && Array.isArray(raw.columns) && raw.columns.length > 0) {
    node.children = raw.columns
      .map((col: any) => {
        if (typeof col === 'string') return { type: 'column' as const, name: col, database: db, table: name }
        return {
          type: 'column' as const,
          name: col.name || col.column || String(col),
          database: db,
          table: name,
          colType: col.type || col.colType || col.dataType || undefined,
          pk: !!col.primaryKey || !!col.pk
        }
      })
      .filter((c: any) => c && c.name)
  }
  return node
}

export function flattenTree(list: any[]): TreeNode[] {
  if (!list || list.length === 0) return []
  const byDb = new Map<string, Map<string, TreeNode[]>>()
  const ensure = (db: string): Map<string, TreeNode[]> => {
    if (!byDb.has(db)) byDb.set(db, new Map())
    return byDb.get(db)!
  }
  const collectBucket = (db: string, bucket: string, node: TreeNode) => {
    const buckets = ensure(db)
    if (!buckets.has(bucket)) buckets.set(bucket, [])
    buckets.get(bucket)!.push(node)
  }
  const walk = (raw: any, parentDb: string) => {
    if (!raw) return
    const type = (raw.type || raw.objType || raw.kind || 'table').toString().toLowerCase()
    const db = raw.database || raw.db || parentDb
    if (type === 'database' || type === 'db' || type === 'schema') {
      const schemaDb = raw.database || raw.name || parentDb
      if (schemaDb) ensure(schemaDb)
      if (Array.isArray(raw.children) && raw.children.length > 0) {
        for (const c of raw.children) walk(c, schemaDb)
      }
      return
    }
    const node = normalizeNode(raw, parentDb)
    if (!node) return
    const normalized = node.type
    if (normalized === 'table' || normalized === 'view') {
      collectBucket(db, normalized === 'view' ? 'views' : 'tables', node)
      return
    }
    if (normalized === 'function' || normalized === 'procedure' || normalized === 'trigger' || normalized === 'index') {
      collectBucket(db, pluralType(normalized), node)
      return
    }
    if (normalized === 'group' || normalized === 'folder') {
      if (Array.isArray(raw.children) && raw.children.length > 0) {
        for (const c of raw.children) walk(c, db)
      }
      return
    }
    if (Array.isArray(raw.children) && raw.children.length > 0) {
      for (const c of raw.children) walk(c, db)
    }
  }
  for (const raw of list) walk(raw, raw.database || raw.db || '')
  const out: TreeNode[] = []
  for (const [db, buckets] of byDb.entries()) {
    const children: TreeNode[] = []
    for (const bk of BUCKET_ORDER) {
      const items = buckets.get(bk)
      if (items && items.length > 0) {
        items.sort((a, b) => (a.name || '').localeCompare(b.name || ''))
        children.push({
          type: 'group',
          name: BUCKET_LABELS[bk],
          database: db,
          group: bk,
          children: items
        })
      }
    }
    if (!db) {
      out.push(...children)
    } else {
      out.push({ type: 'database', name: db, database: db, children })
    }
  }
  return out
}

export function escapeIdentifier(name: string): string {
  if (!name) return ''
  if (/[^a-z0-9_]/.test(name.toLowerCase())) return '`' + name + '`'
  return name
}
