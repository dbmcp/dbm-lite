<template>
  <div class="sqlide-root" @click="ctxMenu.show = false">
    <div class="tree-panel" :style="{ width: treePanelWidth + 'px' }">
      <div class="tree-panel-header">
        <el-select v-model="currentDsId" placeholder="选择数据源" filterable size="small" class="tree-ds-select" @change="onTreeDsChange">
          <el-option v-for="ds in datasourceList" :key="ds.datasourceId" :label="ds.name" :value="ds.datasourceId" />
        </el-select>
      </div>
      <div class="tree-panel-search">
        <el-input v-model="treeKeyword" placeholder="搜索对象..." size="small" clearable :prefix-icon="Search" />
      </div>
      <div class="tree-panel-toolbar">
        <el-button size="small" @click="refreshTree"><span class="refresh-icon">⟳</span>刷新</el-button>
        <el-button size="small" :icon="Folder" @click="expandAllTree">展开</el-button>
        <el-button size="small" :icon="FolderOpened" @click="collapseAllTree">折叠</el-button>
      </div>
      <div class="tree-panel-body">
        <div v-if="datasourceList.length === 0" class="tree-empty"><el-icon :size="20"><FolderOpened /></el-icon><span>暂无数据源</span></div>
        <div v-else-if="treeLoading" class="tree-empty small"><el-icon class="is-loading" :size="16"><Loading /></el-icon><span>加载中...</span></div>
        <div v-else-if="treeRoot.length === 0" class="tree-empty small"><el-icon :size="16"><Folder /></el-icon><span>暂无对象</span></div>
        <div v-else class="tree-content">
          <TreeItem v-for="(node, idx) in treeRoot" :key="'n-' + idx" :node="node" :level="0" :search-keyword="treeKeyword" />
        </div>
      </div>
      <div class="tree-panel-footer"><span>{{ datasourceList.length }} 个数据源</span></div>
    </div>

    <div class="resizer resizer-v" title="拖动调整左侧宽度" @mousedown.prevent="startTreeResize" />

    <div class="workspace">
      <div class="tab-bar">
        <div class="tab-list">
          <div v-for="tab in tabs" :key="tab.id" class="tab-item" :class="{ active: activeTabId === tab.id }" @click="switchTab(tab.id)">
            <el-icon class="tab-icon"><component :is="(tab as any).kind === 'query' ? EditPen : Grid" /></el-icon>
            <span class="tab-name" :title="(tab as any).title">{{ (tab as any).title }}</span>
            <span class="tab-close" @click.stop="closeTab(tab.id)">×</span>
          </div>
          <div class="tab-add" @click="addQueryTab()">+ 新建查询</div>
        </div>
        <div class="tab-bar-status">
          <el-icon><CircleCheck v-if="!executing" /><Loading v-else class="is-loading" /></el-icon>
          <span>{{ executing ? '执行中...' : '就绪' }}</span>
        </div>
      </div>

      <div class="tab-content">
        <template v-for="tab in tabs" :key="'content-' + tab.id">
          <div v-show="activeTabId === tab.id && (tab as QueryTab).kind === 'query'" class="tab-pane query-pane">
            <div class="query-toolbar">
              <el-select v-model="(tab as QueryTab).datasourceId" placeholder="选择数据源" filterable size="small" class="qt-ds" @change="(v: string) => onTabDsChange(tab as QueryTab, v)">
                <el-option v-for="ds in datasourceList" :key="ds.datasourceId" :label="ds.name" :value="ds.datasourceId" />
              </el-select>
              <el-select v-model="(tab as QueryTab).database" placeholder="选择数据库" filterable size="small" class="qt-db" :disabled="!(tab as QueryTab).datasourceId">
                <el-option v-for="db in databasesByDs.get((tab as QueryTab).datasourceId) || []" :key="db" :label="db" :value="db" />
              </el-select>
              <div class="qt-sep"></div>
              <el-button size="small" type="primary" :icon="VideoPlay" :disabled="!(tab as QueryTab).datasourceId || executing" @click="runEditorSql(tab as QueryTab)">执行</el-button>
              <el-button size="small" :icon="MagicStick" :disabled="!(tab as QueryTab).datasourceId" @click="explainSqlForTab(tab as QueryTab)">执行计划</el-button>
              <el-button size="small" link type="primary" :disabled="!(tab as QueryTab).datasourceId" @click="refreshTabDatabases(tab as QueryTab)"><span class="refresh-icon">⟳</span>刷新数据库</el-button>
            </div>

            <div class="editor-wrap" :style="{ height: editorHeightPct + '%' }">
              <textarea class="sql-textarea" v-model="(tab as QueryTab).sql" placeholder="在此输入 SQL 语句，多个语句以分号分隔..." spellcheck="false" @keydown.ctrl.enter.prevent="runEditorSql(tab as QueryTab)"></textarea>
            </div>

            <div class="resizer resizer-h" @mousedown.prevent="startResultResize"></div>

            <div class="result-wrap">
              <el-tabs v-model="(tab as QueryTab).result.activeName" class="result-tabs">
                <el-tab-pane label="结果" name="result">
                  <div v-if="(tab as QueryTab).result.statementResults.length === 0" class="result-empty"><el-icon :size="24"><Grid /></el-icon><span>尚未执行 SQL</span></div>
                  <div v-else>
                    <el-tabs v-model="(tab as QueryTab).result.activeResultIdx" type="card" size="small">
                      <el-tab-pane v-for="(r, rIdx) in (tab as QueryTab).result.statementResults" :key="'rp-' + rIdx" :label="'结果 ' + (rIdx + 1)" :name="String(rIdx)">
                        <div v-if="r.isSelect && r.columns && r.columns.length > 0" class="result-grid-wrap">
                          <div class="result-grid-toolbar">
                            <span class="result-info">共 {{ (r.rows || []).length }} 行 · {{ (r.columns || []).length }} 列</span>
                            <span class="ts-spacer"></span>
                            <el-button size="small" :icon="Download" @click="exportResultCsv(r)">导出 CSV</el-button>
                          </div>
                          <el-table :data="r.rows || []" border stripe size="small" style="width:100%" height="calc(100% - 60px)" show-overflow-tooltip>
                            <el-table-column v-for="col in r.columns" :key="'rc-' + col" :prop="col" :label="col" :min-width="120" show-overflow-tooltip>
                              <template #default="{ row }"><span v-if="row[col] === null || row[col] === undefined" class="null-cell">NULL</span><span v-else class="cell-value">{{ stringifyCell(row[col]) }}</span></template>
                            </el-table-column>
                          </el-table>
                        </div>
                        <div v-else class="result-success-box">
                          <el-icon><CircleCheck v-if="r.success !== false" /><ChatDotRound v-else /></el-icon>
                          <span>{{ r.message || (r.success !== false ? '执行成功' : '执行失败') }}</span>
                          <span v-if="r.affectedRows !== undefined">{{ r.affectedRows }} 行</span>
                        </div>
                      </el-tab-pane>
                    </el-tabs>
                  </div>
                </el-tab-pane>

                <el-tab-pane label="执行计划" name="explain">
                  <div v-if="!(tab as QueryTab).result.explain || (tab as QueryTab).result.explain.length === 0" class="result-empty"><el-icon :size="24"><MagicStick /></el-icon><span>点击 "执行计划" 按钮生成</span></div>
                  <div v-else class="result-grid-wrap">
                    <el-table :data="(tab as QueryTab).result.explain || []" border stripe size="small" style="width:100%" height="calc(100% - 24px)" show-overflow-tooltip>
                      <el-table-column v-for="col in (tab as QueryTab).result.explainCols || []" :key="'ec-' + col" :prop="col" :label="col" :min-width="140" show-overflow-tooltip>
                        <template #default="{ row }"><span v-if="row[col] === null || row[col] === undefined || row[col] === ''" class="null-cell">NULL</span><span v-else class="cell-value">{{ stringifyCell(row[col]) }}</span></template>
                      </el-table-column>
                    </el-table>
                  </div>
                </el-tab-pane>

                <el-tab-pane :label="'消息 (' + ((tab as QueryTab).result.messages || []).length + ')'" name="message">
                  <div v-if="!(tab as QueryTab).result.messages || (tab as QueryTab).result.messages.length === 0" class="result-empty"><el-icon :size="24"><ChatDotRound /></el-icon><span>暂无消息</span></div>
                  <div v-else class="result-message-list">
                    <div v-for="(m, mi) in (tab as QueryTab).result.messages" :key="'m-' + mi" class="result-message-item">
                      <span class="msg-time">{{ m.time }}</span>
                      <el-tag size="small">{{ m.level }}</el-tag>
                      <span class="msg-text">{{ m.text }}</span>
                    </div>
                  </div>
                </el-tab-pane>

                <el-tab-pane :label="'历史 (' + ((tab as QueryTab).history || []).length + ')'" name="history">
                  <div class="history-toolbar">
                    <el-input v-model="historyKeyword" placeholder="过滤关键字" size="small" clearable style="width:260px" :prefix-icon="Search" @change="loadHistoryForTab(tab as QueryTab)" />
                    <el-button size="small" @click="loadHistoryForTab(tab as QueryTab)"><span class="refresh-icon">⟳</span>刷新</el-button>
                    <span class="result-info">展示最近 100 条</span>
                  </div>
                  <div v-if="!(tab as QueryTab).history || (tab as QueryTab).history.length === 0" class="result-empty"><el-icon :size="24"><Clock /></el-icon><span>暂无执行历史</span></div>
                  <el-table v-else :data="(tab as QueryTab).history" size="small" border stripe height="calc(100% - 48px)" style="margin: 0">
                    <el-table-column prop="time" label="时间" width="180" />
                    <el-table-column prop="sql" label="SQL" show-overflow-tooltip min-width="240">
                      <template #default="scope"><span style="font-family: Consolas, Monaco, monospace; font-size: 12px;">{{ scope.row.sql }}</span></template>
                    </el-table-column>
                    <el-table-column prop="database" label="数据库" width="130" />
                    <el-table-column prop="durationMs" label="耗时(ms)" width="100" align="right"><template #default="scope">{{ scope.row.durationMs ?? '-' }}</template></el-table-column>
                    <el-table-column label="状态" width="100">
                      <template #default="scope"><el-tag v-if="scope.row.success" type="success" size="small">成功</el-tag><el-tag v-else type="danger" size="small">失败</el-tag></template>
                    </el-table-column>
                    <el-table-column label="操作" width="110">
                      <template #default="scope"><el-button link type="primary" size="small" @click="replaySql(tab as QueryTab, scope.row)">重新执行</el-button></template>
                    </el-table-column>
                  </el-table>
                </el-tab-pane>
              </el-tabs>
            </div>
          </div>

          <div v-show="activeTabId === tab.id && (tab as TableTab).kind === 'table'" class="tab-pane table-pane">
            <div class="table-toolbar">
              <el-input v-model="(tab as TableTab).table" size="small" style="width: 200px; margin-right: 8px" readonly />
              <el-button size="small" type="primary" :disabled="(tab as TableTab).tableLoading" @click="loadTableData(tab as TableTab)"><span class="refresh-icon">⟳</span>刷新</el-button>
              <el-button size="small" :icon="Plus" @click="addTableRow(tab as TableTab)">插入行</el-button>
              <el-button size="small" :icon="Delete" @click="deleteSelectedRows(tab as TableTab)">删除行</el-button>
              <el-button size="small" type="success" :icon="Check" :disabled="(tab as TableTab).editedRows.size === 0" @click="saveTableChanges(tab as TableTab)">保存更改</el-button>
              <span class="result-info">共 {{ (tab as TableTab).rowCount }} 行 · {{ (tab as TableTab).columns.length }} 列</span>
              <span v-if="(tab as TableTab).editedRows.size > 0" class="edited-indicator">{{ (tab as TableTab).editedRows.size }} 行已修改</span>
            </div>
            <div class="table-view-wrap">
              <div v-if="(tab as TableTab).tableLoading" class="result-empty"><el-icon class="is-loading" :size="20"><Loading /></el-icon><span>加载中...</span></div>
              <div v-else-if="!(tab as TableTab).columns || (tab as TableTab).columns.length === 0" class="result-empty"><el-icon :size="24"><Grid /></el-icon><span>暂无数据</span></div>
              <div v-else class="table-data-wrap">
                <el-table 
                  :data="(tab as TableTab).rows || []" 
                  border 
                  stripe 
                  size="small" 
                  style="width:100%" 
                  height="calc(100% - 52px)" 
                  show-overflow-tooltip
                  :row-class-name="(row, idx) => ((tab as TableTab).editedRows.has(idx) ? 'edited-row' : '')"
                  @cell-dblclick="(row, col, cell, event) => handleCellEdit(tab as TableTab, row, col.prop as string, cell)"
                >
                  <el-table-column type="selection" width="50" />
                  <el-table-column type="index" label="#" width="50" />
                  <el-table-column v-for="col in (tab as TableTab).columns" :key="'tc-' + col" :prop="col" :label="col" :min-width="100" show-overflow-tooltip>
                    <template #default="{ row, $index }">
                      <template v-if="editingCell?.tabId === tab.id && editingCell?.rowIdx === $index && editingCell?.colName === col">
                        <el-input 
                          v-model="row[col]" 
                          size="small" 
                          class="cell-editor" 
                          @blur="finishCellEdit(tab as TableTab, $index)"
                          @keydown.enter="finishCellEdit(tab as TableTab, $index)"
                          @keydown.escape="cancelCellEdit()"
                          ref="editInputRef"
                        />
                      </template>
                      <template v-else>
                        <span v-if="row[col] === null || row[col] === undefined" class="null-cell">NULL</span>
                        <span v-else class="cell-value">{{ stringifyCell(row[col]) }}</span>
                      </template>
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </div>
            <div class="table-status-bar">
              <div class="status-left">
                <span class="status-sql">SELECT * FROM {{ (tab as TableTab).database }}.{{ (tab as TableTab).table }}</span>
              </div>
              <div class="status-center">
                <el-button size="small" :icon="SkipBack" :disabled="(tab as TableTab).currentPage <= 1" @click="goToFirstPage(tab as TableTab)">首页</el-button>
                <el-button size="small" :icon="ChevronLeft" :disabled="(tab as TableTab).currentPage <= 1" @click="prevPage(tab as TableTab)">上一页</el-button>
                <span class="page-info">
                  第 {{ ((tab as TableTab).currentPage - 1) * (tab as TableTab).pageSize + 1 }}-{{ Math.min((tab as TableTab).currentPage * (tab as TableTab).pageSize, (tab as TableTab).rowCount) }} 条记录，共 {{ (tab as TableTab).rowCount }} 条 · 
                  第 {{ (tab as TableTab).currentPage }} 页 / 共 {{ Math.ceil((tab as TableTab).rowCount / (tab as TableTab).pageSize) }} 页
                </span>
                <el-button size="small" :icon="ChevronRight" :disabled="(tab as TableTab).currentPage >= Math.ceil((tab as TableTab).rowCount / (tab as TableTab).pageSize)" @click="nextPage(tab as TableTab)">下一页</el-button>
                <el-button size="small" :icon="SkipForward" :disabled="(tab as TableTab).currentPage >= Math.ceil((tab as TableTab).rowCount / (tab as TableTab).pageSize)" @click="goToLastPage(tab as TableTab)">末页</el-button>
              </div>
              <div class="status-right">
                <span class="refresh-time">上次刷新: {{ (tab as TableTab).lastRefresh }}</span>
                <el-select v-model="(tab as TableTab).pageSize" size="small" style="width: 80px; margin-left: 12px" @change="onTablePageSizeChange(tab as TableTab)">
                  <el-option :label="'10 行'" :value="10" />
                  <el-option :label="'30 行'" :value="30" />
                  <el-option :label="'50 行'" :value="50" />
                  <el-option :label="'100 行'" :value="100" />
                </el-select>
              </div>
            </div>
          </div>
        </template>
      </div>
    </div>

    <div v-if="ctxMenu.show" class="ctx-menu-overlay" @click="ctxMenu.show = false" @contextmenu.prevent="ctxMenu.show = false">
      <div class="ctx-menu" :style="{ left: ctxMenu.x + 'px', top: ctxMenu.y + 'px' }" @click.stop>
        <template v-if="ctxMenu.node && ctxMenu.node.type === 'database'">
          <div class="ctx-menu-item" @click="handleCtxAction('set-db')"><el-icon><DataLine /></el-icon><span>设为当前数据库</span></div>
          <div class="ctx-menu-item" @click="handleCtxAction('query')"><el-icon><EditPen /></el-icon><span>新建查询</span></div>
          <div class="ctx-menu-sep"></div>
          <div class="ctx-menu-item" @click="handleCtxAction('copy-name')"><el-icon><DocumentCopy /></el-icon><span>复制名称</span></div>
        </template>
        <template v-else-if="ctxMenu.node && (ctxMenu.node.type === 'table' || ctxMenu.node.type === 'view')">
          <div class="ctx-menu-item" @click="handleCtxAction('open-table')"><el-icon><Grid /></el-icon><span>打开表数据</span></div>
          <div class="ctx-menu-item" @click="handleCtxAction('select-top-100')"><el-icon><VideoPlay /></el-icon><span>查询前 200 行</span></div>
          <div class="ctx-menu-item" @click="handleCtxAction('ddl')"><el-icon><Document /></el-icon><span>查看 DDL</span></div>
          <div class="ctx-menu-sep"></div>
          <div class="ctx-menu-item" @click="handleCtxAction('copy-name')"><el-icon><DocumentCopy /></el-icon><span>复制名称</span></div>
        </template>
        <template v-else-if="ctxMenu.node && (ctxMenu.node.type === 'function' || ctxMenu.node.type === 'procedure' || ctxMenu.node.type === 'trigger')">
          <div class="ctx-menu-item" @click="handleCtxAction('ddl')"><el-icon><Document /></el-icon><span>查看 DDL</span></div>
          <div class="ctx-menu-sep"></div>
          <div class="ctx-menu-item" @click="handleCtxAction('copy-name')"><el-icon><DocumentCopy /></el-icon><span>复制名称</span></div>
        </template>
        <template v-else-if="ctxMenu.node && ctxMenu.node.type === 'column'">
          <div class="ctx-menu-item" @click="handleCtxAction('insert-col')"><el-icon><EditPen /></el-icon><span>插入列名到编辑器</span></div>
          <div class="ctx-menu-item" @click="handleCtxAction('copy-name')"><el-icon><DocumentCopy /></el-icon><span>复制名称</span></div>
        </template>
        <template v-else-if="ctxMenu.node && (ctxMenu.node.type === 'group' || ctxMenu.node.type === 'schema')">
          <div class="ctx-menu-item" @click="handleCtxAction('query')"><el-icon><EditPen /></el-icon><span>新建查询</span></div>
          <div class="ctx-menu-item" @click="handleCtxAction('copy-name')"><el-icon><DocumentCopy /></el-icon><span>复制名称</span></div>
        </template>
        <template v-else>
          <div v-if="ctxMenu.node" class="ctx-menu-item" @click="handleCtxAction('query')"><el-icon><EditPen /></el-icon><span>新建查询</span></div>
          <div v-if="ctxMenu.node" class="ctx-menu-item" @click="handleCtxAction('copy-name')"><el-icon><DocumentCopy /></el-icon><span>复制名称</span></div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onBeforeUnmount, provide, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, EditPen, Document, Grid, Folder, FolderOpened, Loading, VideoPlay, MagicStick, DocumentCopy, Download, CircleCheck, CircleClose, Plus, Delete, Clock, Edit, ChatDotRound, DataLine, Check, SkipBack, ChevronLeft, ChevronRight, SkipForward } from '@element-plus/icons-vue'
import TreeItem, { type TreeNode } from './TreeItem.vue'
import { executeSql, explainSql, listSqlHistory, getDatabases, getFullTree, queryTableData, getTableDDL } from '@/api/sql'
import { listAllDatasources } from '@/api/datasource'

interface Datasource { datasourceId: string; name: string; dbType: string }
interface StatementResult { sql: string; isSelect: boolean; columns: string[]; rows: Record<string, any>[]; affectedRows: number; durationMs: number; success: boolean; message: string }
interface QueryResultState { activeName: string; activeResultIdx: string; statementResults: StatementResult[]; explain: Record<string, any>[]; explainCols: string[]; messages: { time: string; level: string; text: string }[] }
interface QueryTab { id: string; kind: 'query'; title: string; datasourceId: string; database: string; sql: string; result: QueryResultState; history: HistoryItem[] }
interface TableTab { id: string; kind: 'table'; title: string; datasourceId: string; database: string; table: string; columns: string[]; rows: Record<string, any>[]; tableLoading: boolean; rowCount: number; currentPage: number; pageSize: number; editedRows: Set<number>; lastRefresh: string }
interface HistoryItem { time: string; sql: string; database: string; success: boolean; durationMs?: number }
type AnyTab = QueryTab | TableTab

const treePanelWidth = ref<number>(280)
const editorHeight = ref<number>(50)
const editorHeightPct = computed(() => editorHeight.value)
const treeKeyword = ref<string>('')
const historyKeyword = ref<string>('')
const treeLoading = ref<boolean>(false)
const currentDsId = ref<string>('')
const activeTabId = ref<string>('')
const executing = ref<boolean>(false)

const tabs = ref<AnyTab[]>([])
const datasourceList = ref<Datasource[]>([])
const treeRoot = ref<TreeNode[]>([])
const databasesByDs = reactive<Map<string, string[]>>(new Map())

const ctxMenu = reactive<{ show: boolean; x: number; y: number; node: TreeNode | null }>({ show: false, x: 0, y: 0, node: null })

const editingCell = ref<{ tabId: string; rowIdx: number; colName: string } | null>(null)
const editInputRef = ref<any>(null)

provide('sqlide-menu-opener', (evt: MouseEvent, node: TreeNode) => {
  ctxMenu.node = node
  ctxMenu.x = evt.clientX
  ctxMenu.y = evt.clientY
  ctxMenu.show = true
})

provide('sqlTreeCallbacks', {
  onNodeDblClick: (node: TreeNode) => {
    const db = node.database || ''
    if (node.type === 'table' || node.type === 'view') {
      const qt = addQueryTab(currentDsId.value, db, 'SELECT * FROM ' + escapeName(node.table || node.name) + ' LIMIT 100;', '查询 ' + node.name)
      setTimeout(() => runEditorSql(qt), 50)
    } else if (node.type !== 'column') {
      addQueryTab(currentDsId.value, db, '-- 新建查询\nSELECT 1;', '查询')
    }
  },
  onNodeCtxMenu: (e: MouseEvent, node: TreeNode) => {
    ctxMenu.node = node
    ctxMenu.x = e.clientX
    ctxMenu.y = e.clientY
    ctxMenu.show = true
  },
  onNodeInfoClick: (node: TreeNode) => {
    const db = node.database || ''
    if (node.type === 'table' || node.type === 'view') {
      const tab = addQueryTab(currentDsId.value, db, '', 'DDL ' + node.name)
      getTableDDL({ datasourceId: currentDsId.value, database: db, table: node.table || node.name }).then((res: any) => {
        tab.sql = (typeof res === 'string' ? res : (res?.ddl || res?.sql || JSON.stringify(res)))
      }).catch((e: any) => ElMessage.error('获取 DDL 失败：' + (e?.message || String(e))))
    }
  }
})

function pad2(n: number) { return n < 10 ? '0' + n : String(n) }
function nowString(): string {
  const d = new Date()
  return d.getFullYear() + '-' + pad2(d.getMonth() + 1) + '-' + pad2(d.getDate()) + ' ' + pad2(d.getHours()) + ':' + pad2(d.getMinutes()) + ':' + pad2(d.getSeconds())
}

function splitSql(sql: string): string[] { if (!sql) return []; return sql.split(';').map(s => s.trim()).filter(s => s.length > 0) }

function stringifyCell(v: any): string {
  if (v === null || v === undefined) return 'NULL'
  if (typeof v === 'object') { try { return JSON.stringify(v) } catch { return String(v) } }
  return String(v)
}

function addQueryTab(dsId?: string, dbName?: string, preSql?: string, title?: string): QueryTab {
  const id = 'q-' + Date.now()
  const tab: QueryTab = {
    id, kind: 'query',
    title: title || ('查询 ' + (tabs.value.filter(t => t.kind === 'query').length + 1)),
    datasourceId: dsId || currentDsId.value || (datasourceList.value[0]?.datasourceId || ''),
    database: dbName || '',
    sql: preSql || '',
    result: { activeName: 'result', activeResultIdx: '0', statementResults: [], explain: [], explainCols: [], messages: [] },
    history: []
  }
  tabs.value.push(tab)
  switchTab(tab.id)
  return tab
}

function addTableTab(dsId: string, db: string, table: string): TableTab {
  const existing = tabs.value.find(t => t.kind === 'table' && t.datasourceId === dsId && (t as TableTab).table === table && (t as TableTab).database === db)
  if (existing) { switchTab(existing.id); loadTableData(existing as TableTab); return existing as TableTab }
  const tab: TableTab = { id: 't-' + Date.now(), kind: 'table', title: db + '.' + table, datasourceId: dsId, database: db, table: table, columns: [], rows: [], tableLoading: false, rowCount: 0, currentPage: 1, pageSize: 50, editedRows: new Set(), lastRefresh: '' }
  tabs.value.push(tab)
  switchTab(tab.id)
  loadTableData(tab)
  return tab
}

function switchTab(id: string) {
  activeTabId.value = id
  const tab = tabs.value.find(t => t.id === id)
  if (tab && tab.kind === 'query' && (tab as QueryTab).datasourceId && (tab as QueryTab).history.length === 0) {
    loadHistoryForTab(tab as QueryTab)
  }
}

function closeTab(id: string) {
  const idx = tabs.value.findIndex(t => t.id === id)
  if (idx < 0) return
  tabs.value.splice(idx, 1)
  if (activeTabId.value === id) {
    activeTabId.value = tabs.value.length > 0 ? tabs.value[Math.max(0, idx - 1)].id : ''
  }
}

function findActiveTab(): AnyTab | undefined { return tabs.value.find(t => t.id === activeTabId.value) }

async function loadDatasources() {
  try {
    const res: any = await listAllDatasources()
    const list: any[] = Array.isArray(res) ? res : (res?.list || [])
    datasourceList.value = list.map((d: any) => ({ datasourceId: d.datasourceId || d.id || String(d.id), name: d.name || String(d.id), dbType: d.dbType || d.type || 'unknown' }))
    if (datasourceList.value.length > 0 && !currentDsId.value) currentDsId.value = datasourceList.value[0].datasourceId
  } catch (e: any) { ElMessage.error('数据源加载失败：' + (e?.message || String(e))); datasourceList.value = [] }
}

function onTreeDsChange(id: string) { if (!databasesByDs.has(id)) loadDatabasesFor(id); loadTreeForCurrentDs() }

async function loadDatabasesFor(dsId: string) {
  if (!dsId) return
  try {
    const res: any = await getDatabases(dsId)
    const list: any[] = Array.isArray(res) ? res : (res?.databases || res?.list || [])
    databasesByDs.set(dsId, list.map((d: any) => (typeof d === 'string' ? d : (d.name || d.database || String(d)))))
  } catch (e: any) { ElMessage.error('数据库加载失败：' + (e?.message || String(e))); databasesByDs.set(dsId, []) }
}

async function loadTreeForCurrentDs() {
  const dsId = currentDsId.value
  if (!dsId) { treeRoot.value = []; return }
  treeLoading.value = true
  try {
    const res: any = await getFullTree(dsId)
    const list: any[] = Array.isArray(res) ? res : (res?.list || (res?.data || []))
    treeRoot.value = flattenTree(list)
  } catch (e: any) { ElMessage.error('对象树加载失败：' + (e?.message || String(e))); treeRoot.value = [] }
  finally { treeLoading.value = false }
}

function flattenTree(list: any[]): TreeNode[] {
  if (!list || list.length === 0) return []
  const byDb = new Map<string, Map<string, TreeNode[]>>()
  const ensure = (db: string): Map<string, TreeNode[]> => {
    if (!byDb.has(db)) byDb.set(db, new Map())
    return byDb.get(db)!
  }
  const collect = (db: string, bucket: string, node: TreeNode) => {
    const buckets = ensure(db)
    if (!buckets.has(bucket)) buckets.set(bucket, [])
    buckets.get(bucket)!.push(node)
  }
  const queue: { raw: any; parentDb: string }[] = []
  for (const raw of list) queue.push({ raw, parentDb: raw.database || raw.db || '' })
  while (queue.length > 0) {
    const { raw, parentDb } = queue.shift()!
    if (!raw) continue
    const db = raw.database || raw.db || parentDb
    const type = (raw.type || raw.objType || raw.kind || 'table').toString().toLowerCase()
    if (type === 'database' || type === 'db' || type === 'schema') {
      const schemaDb = raw.database || raw.name || parentDb
      if (raw.children && Array.isArray(raw.children) && raw.children.length > 0) {
        for (const c of raw.children) queue.push({ raw: c, parentDb: schemaDb })
      } else {
        ensure(schemaDb)
      }
      continue
    }
    const name = raw.name || raw.table || raw.objectName || String(raw)
    if (type === 'table' || type === 'tables' || type === 'view' || type === 'views' ||
        type === 'function' || type === 'functions' || type === 'procedure' || type === 'procedures' ||
        type === 'trigger' || type === 'triggers' || type === 'index' || type === 'indexes' ||
        type === 'column' || type === 'columns' || type === 'group' || type === 'folder') {
      if (raw.children && Array.isArray(raw.children) && raw.children.length > 0) {
        for (const c of raw.children) queue.push({ raw: c, parentDb: db })
        continue
      }
      const normalized = singularType(type)
      if (normalized === 'table' || normalized === 'view') {
        const children: TreeNode[] = []
        if (raw.columns && Array.isArray(raw.columns) && raw.columns.length > 0) {
          for (const col of raw.columns) {
            children.push({ name: typeof col === 'string' ? col : (col.name || col.column || String(col)), type: 'column', children: [], database: db, table: name })
          }
        }
        collect(db, normalized === 'view' ? 'views' : 'tables', { name, type: normalized, children, database: db, table: name })
      } else {
        collect(db, pluralType(normalized), { name, type: normalized, children: [], database: db, table: '' })
      }
      continue
    }
    if (raw.children && Array.isArray(raw.children) && raw.children.length > 0) {
      for (const c of raw.children) queue.push({ raw: c, parentDb: db })
    }
  }
  const out: TreeNode[] = []
  for (const [db, buckets] of byDb.entries()) {
    const children: TreeNode[] = []
    const bucketsOrder = ['tables', 'views', 'functions', 'procedures', 'triggers', 'indexes']
    const bucketLabels: Record<string, string> = {
      tables: '表', views: '视图', functions: '函数', procedures: '存储过程', triggers: '触发器', indexes: '索引'
    }
    for (const bk of bucketsOrder) {
      const items = buckets.get(bk)
      if (items && items.length > 0) {
        items.sort((a, b) => (a.name || '').localeCompare(b.name || ''))
        children.push({ name: bucketLabels[bk], type: 'group', children: items, database: db, table: '' })
      }
    }
    const extras = buckets.get('') || []
    children.push(...extras)
    if (db) out.push({ name: db, type: 'database', children, database: db })
    else out.push(...children)
  }
  return out
}

function singularType(t: string): string {
  const s = t.toLowerCase()
  if (s === 'tables') return 'table'
  if (s === 'views') return 'view'
  if (s === 'functions') return 'function'
  if (s === 'procedures') return 'procedure'
  if (s === 'triggers') return 'trigger'
  if (s === 'indexes') return 'index'
  if (s === 'columns') return 'column'
  return s || 'table'
}
function pluralType(t: string): string {
  const sing = singularType(t)
  if (sing === 'index') return 'indexes'
  return sing + 's'
}

function refreshTree() { loadTreeForCurrentDs(); if (currentDsId.value) loadDatabasesFor(currentDsId.value) }

function expandAllTree() {
  const flag = (nodes: TreeNode[]): TreeNode[] => nodes.map(n => ({ ...n, children: n.children ? flag(n.children) : [] }))
  treeRoot.value = flag(treeRoot.value)
}
function collapseAllTree() {
  const flag = (nodes: TreeNode[]): TreeNode[] => nodes.map(n => ({ ...n, children: [] }))
  treeRoot.value = flag(treeRoot.value)
}

async function runEditorSql(tab: QueryTab) {
  if (!tab.datasourceId) { ElMessage.warning('请先选择数据源'); return }
  const sql = (tab.sql || '').trim()
  if (!sql) { ElMessage.warning('请输入 SQL 语句'); return }
  executing.value = true
  const start = Date.now()
  tab.result.activeName = 'result'
  tab.result.statementResults = []
  tab.result.messages.push({ time: nowString(), level: 'INFO', text: '开始执行 SQL' })
  try {
    const res: any = await executeSql({ datasourceId: tab.datasourceId, database: tab.database, sql, ignoreRisk: true })
    const statList: any[] = Array.isArray(res) ? res : (Array.isArray(res?.data) ? res.data : (res ? [res] : []))
    const statements = splitSql(sql)
    for (let i = 0; i < Math.max(statList.length, statements.length); i++) {
      const r = statList[i]
      if (!r) continue
      const isSelect = !!r.isSelect || (Array.isArray(r.columns) && r.columns.length > 0 && Array.isArray(r.rows))
      const sr: StatementResult = {
        sql: r.sql || statements[i] || sql,
        isSelect: isSelect,
        columns: Array.isArray(r.columns) ? r.columns : (Array.isArray(r.columnNames) ? r.columnNames : []),
        rows: Array.isArray(r.rows) ? r.rows : (Array.isArray(r.data) ? r.data : []),
        affectedRows: typeof r.affectedRows !== 'undefined' ? r.affectedRows : (typeof r.affected_rows !== 'undefined' ? r.affected_rows : 0),
        durationMs: typeof r.durationMs !== 'undefined' ? r.durationMs : (Date.now() - start),
        success: r.success !== false && !r.error,
        message: r.message || (r.success !== false ? '执行成功' : '执行失败')
      }
      tab.result.statementResults.push(sr)
      tab.result.messages.push({ time: nowString(), level: sr.success ? 'INFO' : 'ERROR', text: (sr.success ? 'OK · ' : 'FAILED · ') + sr.message })
      tab.history.unshift({ time: nowString(), sql: sr.sql, database: tab.database, success: sr.success, durationMs: sr.durationMs })
    }
    tab.result.activeResultIdx = '0'
  } catch (e: any) {
    const msg = e?.message || String(e)
    tab.result.statementResults.push({ sql: sql, isSelect: false, columns: [], rows: [], affectedRows: 0, durationMs: Date.now() - start, success: false, message: '执行异常：' + msg })
    tab.result.messages.push({ time: nowString(), level: 'ERROR', text: msg })
    ElMessage.error(msg)
  } finally { executing.value = false }
}

async function explainSqlForTab(tab: QueryTab) {
  if (!tab.datasourceId) { ElMessage.warning('请先选择数据源'); return }
  const sql = (tab.sql || '').trim()
  if (!sql) { ElMessage.warning('请输入 SQL 语句'); return }
  executing.value = true
  tab.result.activeName = 'explain'
  try {
    const res: any = await explainSql({ datasourceId: tab.datasourceId, database: tab.database, sql, ignoreRisk: true })
    const rows = Array.isArray(res) ? res : (Array.isArray(res?.data) ? res.data : (Array.isArray(res?.rows) ? res.rows : []))
    if (rows.length > 0) { tab.result.explainCols = Object.keys(rows[0]); tab.result.explain = rows }
  } catch (e: any) { ElMessage.error(e?.message || String(e)) } finally { executing.value = false }
}

function refreshTabDatabases(tab: QueryTab) {
  if (!tab.datasourceId) return
  loadDatabasesFor(tab.datasourceId)
  if (tab.datasourceId === currentDsId.value) loadTreeForCurrentDs()
}

function onTabDsChange(tab: QueryTab, dsId: string) {
  if (!dsId) return
  if (!databasesByDs.has(dsId)) loadDatabasesFor(dsId)
  tab.database = ''
}

async function loadTableData(tab: TableTab) {
  if (!tab.datasourceId || !tab.table) return
  tab.tableLoading = true
  try {
    const limit = tab.pageSize
    const offset = (tab.currentPage - 1) * tab.pageSize
    
    const countSql = `SELECT COUNT(*) as total FROM ${escapeName(tab.table)}`
    const countRes: any = await executeSql({ datasourceId: tab.datasourceId, database: tab.database, sql: countSql, ignoreRisk: true })
    const countStatList: any[] = Array.isArray(countRes) ? countRes : (Array.isArray(countRes?.data) ? countRes.data : (countRes ? [countRes] : []))
    tab.rowCount = countStatList.length > 0 && countStatList[0] && countStatList[0].rows && countStatList[0].rows.length > 0 
      ? parseInt(String(countStatList[0].rows[0].total || countStatList[0].rows[0].COUNT)) || 0 
      : 0

    const dataSql = `SELECT * FROM ${escapeName(tab.table)} LIMIT ${limit} OFFSET ${offset}`
    const dataRes: any = await executeSql({ datasourceId: tab.datasourceId, database: tab.database, sql: dataSql, ignoreRisk: true })
    const statList: any[] = Array.isArray(dataRes) ? dataRes : (Array.isArray(dataRes?.data) ? dataRes.data : (dataRes ? [dataRes] : []))
    let rows: any[] = [], columns: string[] = []
    if (statList.length > 0 && statList[0]) {
      const r = statList[0]
      columns = Array.isArray(r.columns) ? r.columns : (Array.isArray(r.columnNames) ? r.columnNames : [])
      rows = Array.isArray(r.rows) ? r.rows : (Array.isArray(r.data) ? r.data : [])
    }
    if (columns.length === 0 && rows.length > 0) columns = Object.keys(rows[0])
    tab.rows = rows
    tab.columns = columns
    tab.lastRefresh = nowString()
    tab.editedRows.clear()
  } catch (e: any) { ElMessage.error('表数据加载失败：' + (e?.message || String(e))) }
  finally { tab.tableLoading = false }
}

function handleCellEdit(tab: TableTab, row: Record<string, any>, colName: string, cell: any) {
  const idx = tab.rows.findIndex(r => r === row)
  editingCell.value = { tabId: tab.id, rowIdx: idx, colName }
  nextTick(() => {
    const inputs = document.querySelectorAll('.cell-editor')
    if (inputs.length > 0) {
      (inputs[inputs.length - 1] as HTMLInputElement).focus()
    }
  })
}

function finishCellEdit(tab: TableTab, rowIdx: number) {
  tab.editedRows.add(rowIdx)
  editingCell.value = null
}

function cancelCellEdit() {
  editingCell.value = null
}

function addTableRow(tab: TableTab) {
  const newRow: Record<string, any> = {}
  for (const col of tab.columns) {
    newRow[col] = ''
  }
  tab.rows.unshift(newRow)
  tab.editedRows.add(0)
}

function deleteSelectedRows(tab: TableTab) {
  ElMessageBox.confirm('确定要删除选中的行吗？', '提示', { type: 'warning' }).then(async () => {
    const table = tab
    const rowsToDelete = table.rows.filter((_, idx) => table.editedRows.has(idx))
    if (rowsToDelete.length === 0) {
      ElMessage.warning('请选择要删除的行')
      return
    }
    ElMessage.success(`已标记 ${rowsToDelete.length} 行准备删除，点击"保存更改"完成操作`)
  }).catch(() => {})
}

async function saveTableChanges(tab: TableTab) {
  if (tab.editedRows.size === 0) return
  tab.tableLoading = true
  try {
    const editedRows = Array.from(tab.editedRows).sort((a, b) => b - a)
    for (const idx of editedRows) {
      const row = tab.rows[idx]
      const setClauses = tab.columns.filter(c => c !== 'id').map(c => `${escapeName(c)} = ${formatValue(row[c])}`).join(', ')
      const whereClause = row.id ? `WHERE id = ${row.id}` : ''
      let sql = ''
      if (row.id) {
        sql = `UPDATE ${escapeName(tab.table)} SET ${setClauses} ${whereClause}`
      } else {
        const cols = tab.columns.filter(c => row[c] !== undefined && row[c] !== '')
        const values = cols.map(c => formatValue(row[c]))
        sql = `INSERT INTO ${escapeName(tab.table)} (${cols.map(escapeName).join(', ')}) VALUES (${values.join(', ')})`
      }
      await executeSql({ datasourceId: tab.datasourceId, database: tab.database, sql, ignoreRisk: true })
    }
    tab.editedRows.clear()
    await loadTableData(tab)
    ElMessage.success('保存成功')
  } catch (e: any) {
    ElMessage.error('保存失败：' + (e?.message || String(e)))
  } finally {
    tab.tableLoading = false
  }
}

function formatValue(v: any): string {
  if (v === null || v === undefined || v === '') return 'NULL'
  if (typeof v === 'number') return String(v)
  return `'${String(v).replace(/'/g, "''")}'`
}

function goToFirstPage(tab: TableTab) {
  tab.currentPage = 1
  loadTableData(tab)
}

function prevPage(tab: TableTab) {
  if (tab.currentPage > 1) {
    tab.currentPage--
    loadTableData(tab)
  }
}

function nextPage(tab: TableTab) {
  const maxPage = Math.ceil(tab.rowCount / tab.pageSize)
  if (tab.currentPage < maxPage) {
    tab.currentPage++
    loadTableData(tab)
  }
}

function goToLastPage(tab: TableTab) {
  tab.currentPage = Math.max(1, Math.ceil(tab.rowCount / tab.pageSize))
  loadTableData(tab)
}

function onTablePageSizeChange(tab: TableTab) {
  tab.currentPage = 1
  loadTableData(tab)
}

function handleCtxAction(action: string) {
  const node = ctxMenu.node
  ctxMenu.show = false
  if (!node) return
  const db = node.database || ''
  const table = node.table || node.name
  if (action === 'query') addQueryTab(currentDsId.value, db, '-- 新建查询\nSELECT 1;', '查询')
  else if (action === 'set-db') {
    const active = tabs.value.find(t => t.id === activeTabId.value) as QueryTab | undefined
    if (active && active.kind === 'query') active.database = db
    ElMessage.success('已切换数据库: ' + db)
  }
  else if (action === 'select-top-100') {
    const qt = addQueryTab(currentDsId.value, db, 'SELECT * FROM ' + escapeName(table) + ' LIMIT 200;', '查询 ' + table)
    setTimeout(() => runEditorSql(qt), 50)
  } else if (action === 'open-table') addTableTab(currentDsId.value, db, table)
  else if (action === 'ddl') {
    const tab = addQueryTab(currentDsId.value, db, '', 'DDL ' + table)
    getTableDDL({ datasourceId: currentDsId.value, database: db, table: table }).then((res: any) => {
      tab.sql = (typeof res === 'string' ? res : (res?.ddl || res?.sql || JSON.stringify(res)))
    }).catch((e: any) => ElMessage.error('获取 DDL 失败：' + (e?.message || String(e))))
  } else if (action === 'copy-name') {
    const text = node.name || ''
    try { (navigator as any).clipboard?.writeText(text); ElMessage.success('已复制：' + text) } catch { ElMessage.success(text) }
  } else if (action === 'insert-col') {
    const active = tabs.value.find(t => t.id === activeTabId.value) as QueryTab | undefined
    if (active && active.kind === 'query') {
      active.sql = (active.sql || '') + (active.sql ? ' ' : '') + escapeName(node.name || '')
      ElMessage.success('已插入列名')
    } else {
      const text = node.name || ''
      try { (navigator as any).clipboard?.writeText(text); ElMessage.success('已复制：' + text) } catch { ElMessage.success(text) }
    }
  }
}

function escapeName(name: string): string {
  if (!name) return ''
  if (/[^a-z0-9_]/.test(name.toLowerCase())) return '`' + name + '`'
  return name
}

async function loadHistoryForTab(tab: QueryTab) {
  const dsId = tab.datasourceId || currentDsId.value
  if (!dsId) { tab.history = []; return }
  try {
    const res: any = await listSqlHistory(dsId, 1, 100, historyKeyword.value)
    const list: any[] = Array.isArray(res) ? res : (res?.list || (Array.isArray(res?.data) ? res.data : []))
    tab.history = list.map((item: any) => ({
      time: item.time || item.createdAt || item.created_at || nowString(),
      sql: item.sql || item.sqlText || item.sql_text || '',
      database: item.database || item.databaseName || item.database_name || '',
      success: String(item.status || item.success || '').toLowerCase() !== 'failed' && item.success !== false,
      durationMs: typeof item.durationMs !== 'undefined' ? item.durationMs : item.duration
    }))
  } catch (e: any) { tab.history = []; ElMessage.warning('历史记录加载失败：' + (e?.message || String(e))) }
}

function replaySql(tab: QueryTab, row: HistoryItem) {
  tab.sql = row.sql
  if (row.database) tab.database = row.database
  tab.result.activeName = 'result'
  runEditorSql(tab)
}

function exportResultCsv(r: StatementResult) {
  if (!r.columns || r.columns.length === 0) return
  const esc = (v: any) => { const s = v === null || v === undefined ? '' : String(v); return /[",\n]/.test(s) ? '"' + s.replace(/"/g, '""') + '"' : s }
  const header = r.columns.map(esc).join(',')
  const body = r.rows.map((row: any) => r.columns!.map((c) => esc(row[c])).join(',')).join('\n')
  const csv = '\ufeff' + header + '\n' + body
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'result-' + Date.now() + '.csv'
  a.click()
  URL.revokeObjectURL(url)
}

function startTreeResize(e: MouseEvent) {
  const startX = e.clientX
  const startW = treePanelWidth.value
  const min = 180, max = Math.max(400, Math.round((document.documentElement.clientWidth || 1024) * 0.5))
  function onMove(ev: MouseEvent) { treePanelWidth.value = Math.min(max, Math.max(min, startW + (ev.clientX - startX))) }
  function onUp() { document.removeEventListener('mousemove', onMove); document.removeEventListener('mouseup', onUp) }
  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
}

function startResultResize(e: MouseEvent) {
  const startY = e.clientY
  const startH = editorHeight.value
  const parent = (e.currentTarget as HTMLElement)?.parentElement
  const workspaceHeight = parent?.clientHeight || 400
  function onMove(ev: MouseEvent) {
    const dy = ev.clientY - startY
    const pct = startH + (dy / workspaceHeight) * 100
    editorHeight.value = Math.min(85, Math.max(15, pct))
  }
  function onUp() { document.removeEventListener('mousemove', onMove); document.removeEventListener('mouseup', onUp) }
  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
}

function onKeyDown(e: KeyboardEvent) {
  if (e.key === 'Escape') ctxMenu.show = false
  if (e.key === 'F5') {
    const tab = findActiveTab()
    if (tab && tab.kind === 'query') { e.preventDefault(); runEditorSql(tab as QueryTab) }
  }
}

onMounted(async () => {
  document.addEventListener('keydown', onKeyDown)
  await loadDatasources()
  if (datasourceList.value.length > 0) {
    currentDsId.value = datasourceList.value[0].datasourceId
    loadDatabasesFor(currentDsId.value)
    loadTreeForCurrentDs()
  }
  nextTick(() => { if (tabs.value.length === 0) addQueryTab() })
})

onBeforeUnmount(() => { document.removeEventListener('keydown', onKeyDown) })
</script>

<style scoped>
.sqlide-root {
  display: flex;
  flex-direction: row;
  width: 100%;
  height: 100%;
  min-height: 100%;
  background: #ffffff;
  overflow: hidden;
  margin: 0;
  padding: 0;
}

.tree-panel {
  display: flex;
  flex-direction: column;
  border-right: 1px solid #e4e7ed;
  background: #fafafa;
  overflow: hidden;
  min-width: 0;
}

.tree-panel-header {
  padding: 6px 8px;
  border-bottom: 1px solid #ebeef5;
  background: #ffffff;
}

.tree-ds-select { width: 100% }

.tree-panel-search { padding: 4px 8px }
.tree-panel-toolbar { padding: 0 6px 4px; display: flex; gap: 4px }
.tree-panel-body { flex: 1 1 auto; overflow: auto; padding: 0 4px 4px 4px }
.tree-panel-footer { border-top: 1px solid #ebeef5; padding: 6px 12px; font-size: 12px; color: #909399; background: #ffffff }

.tree-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: #909399;
  font-size: 13px;
  padding: 20px 10px;
}

.tree-empty.small { font-size: 12px; padding: 12px 6px }

.tree-content { display: flex; flex-direction: column }

.resizer {
  background: transparent;
  transition: background-color .2s;
  flex: 0 0 auto;
}
.resizer:hover { background: #dcdfe6 }
.resizer-v { width: 4px; cursor: col-resize }
.resizer-h { height: 4px; cursor: row-resize }

.workspace {
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
}

.tab-bar {
  display: flex;
  align-items: center;
  padding: 0 8px;
  background: #ffffff;
  border-bottom: 1px solid #e4e7ed;
  height: 38px;
  flex: 0 0 auto;
}

.tab-list { display: flex; align-items: center; flex: 1 1 auto; min-width: 0; overflow-x: auto }

.tab-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 12px;
  height: 30px;
  background: #f5f7fa;
  border: 1px solid #dcdfe6;
  border-radius: 4px 4px 0 0;
  margin-right: 4px;
  cursor: pointer;
  font-size: 13px;
  color: #606266;
  white-space: nowrap;
  flex: 0 0 auto;
}

.tab-item.active {
  background: #ffffff;
  color: #409EFF;
  border-bottom-color: #ffffff;
}

.tab-icon { font-size: 14px }
.tab-name { max-width: 180px; overflow: hidden; text-overflow: ellipsis }
.tab-close {
  margin-left: 6px;
  color: #909399;
  font-size: 16px;
  line-height: 1;
  padding: 0 2px;
  border-radius: 2px;
}
.tab-close:hover { background: #dcdfe6; color: #303133 }

.tab-add {
  display: inline-block;
  padding: 0 10px;
  height: 30px;
  line-height: 30px;
  color: #409EFF;
  cursor: pointer;
  font-size: 13px;
  margin-right: 8px;
  border-radius: 4px;
}
.tab-add:hover { background: #ecf5ff }

.tab-bar-status {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #909399;
  padding-left: 10px;
  border-left: 1px solid #ebeef5;
  flex: 0 0 auto;
}

.tab-content { flex: 1 1 auto; min-height: 0; overflow: hidden }

.tab-pane {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.query-toolbar {
  display: flex;
  align-items: center;
  padding: 4px 8px;
  gap: 6px;
  border-bottom: 1px solid #ebeef5;
  background: #fafafa;
  flex: 0 0 auto;
}
.qt-ds { width: 180px }
.qt-db { width: 160px }
.qt-sep { width: 1px; height: 20px; background: #dcdfe6; margin: 0 4px }
.qt-spacer { flex: 1 1 auto }

.editor-wrap {
  flex: 0 0 auto;
  min-height: 0;
  background: #ffffff;
}

.sql-textarea {
  width: 100%;
  height: 100%;
  box-sizing: border-box;
  border: 0;
  outline: none;
  padding: 8px 12px;
  font-family: Consolas, Monaco, 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.6;
  color: #303133;
  background: #ffffff;
  resize: none;
  white-space: pre;
  tab-size: 4;
}

.result-wrap {
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background: #ffffff;
}

.result-tabs { flex: 1 1 auto; display: flex; flex-direction: column; min-height: 0 }
.result-tabs :deep(.el-tabs__content) { flex: 1 1 auto; overflow: hidden }
.result-tabs :deep(.el-tab-pane) { height: 100%; display: flex; flex-direction: column }

.result-empty {
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #909399;
  font-size: 14px;
}

.result-grid-wrap {
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding: 4px;
}
.result-grid-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-bottom: 8px;
}

.result-info { font-size: 12px; color: #909399 }
.ts-spacer { flex: 1 1 auto }

.null-cell { color: #c0c4cc; font-size: 12px; font-style: italic }
.cell-value { color: #303133 }

.result-success-box {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  color: #303133;
  font-size: 13px;
  background: #f4f7fa;
  border-radius: 4px;
  margin: 4px;
}

.result-message-list {
  flex: 1 1 auto;
  overflow: auto;
  padding: 4px;
}

.result-message-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border-bottom: 1px dashed #ebeef5;
  font-size: 13px;
}

.msg-time { color: #909399; font-size: 12px; font-family: Consolas, monospace }
.msg-text { color: #303133 }

.history-toolbar {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-bottom: 1px solid #ebeef5;
  background: #fafafa;
}

.table-toolbar {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-bottom: 1px solid #ebeef5;
  background: #fafafa;
  flex: 0 0 auto;
}

.table-view-wrap { flex: 1 1 auto; min-height: 0; display: flex; flex-direction: column }
.table-data-wrap { flex: 1 1 auto; padding: 4px; min-height: 0; display: flex; flex-direction: column }

.ctx-menu-overlay {
  position: fixed;
  inset: 0;
  z-index: 3000;
  background: transparent;
}

.ctx-menu {
  position: absolute;
  min-width: 180px;
  background: #ffffff;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.12);
  padding: 4px 0;
}

.ctx-menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  cursor: pointer;
  font-size: 13px;
  color: #303133;
}
.ctx-menu-item:hover { background: #ecf5ff; color: #409EFF }

.ctx-menu-sep { height: 1px; background: #ebeef5; margin: 4px 0 }

.table-pane {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.table-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border-bottom: 1px solid #ebeef5;
  background: #fafafa;
  flex: 0 0 auto;
}

.edited-indicator {
  margin-left: 8px;
  padding: 2px 6px;
  background: #fff7e6;
  color: #d48806;
  font-size: 12px;
  border-radius: 4px;
}

.table-view-wrap {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
  background: #ffffff;
}

.table-data-wrap {
  flex: 1 1 auto;
  padding: 4px;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.cell-editor {
  width: 100%;
  margin: 0;
  padding: 2px 4px;
}

:deep(.edited-row) {
  background-color: #fffbe6 !important;
}

.table-status-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 12px;
  background: #f5f5f5;
  border-top: 1px solid #e4e7ed;
  font-size: 12px;
  flex: 0 0 auto;
}

.status-left {
  flex: 0 0 auto;
}

.status-center {
  flex: 1 1 auto;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.status-right {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-sql {
  color: #409EFF;
  font-family: Consolas, Monaco, monospace;
  font-size: 12px;
}

.page-info {
  color: #606266;
  margin: 0 8px;
}

.refresh-time {
  color: #909399;
  font-size: 12px;
}
</style>
