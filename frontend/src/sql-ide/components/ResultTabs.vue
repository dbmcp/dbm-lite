<template>
  <div class="rt-root">
    <div class="rt-tabs-bar">
      <div class="rt-tab" :class="{ active: activeTab === 'result' }" @click="activeTab = 'result'">
        <span class="rt-tab-icon">📋</span>
        <span>结果</span>
        <span v-if="resultCount > 0" class="rt-tab-count">{{ resultCount }}</span>
      </div>
      <div class="rt-tab" :class="{ active: activeTab === 'explain' }" @click="activeTab = 'explain'">
        <span class="rt-tab-icon">📊</span>
        <span>执行计划</span>
      </div>
      <div class="rt-tab" :class="{ active: activeTab === 'message' }" @click="activeTab = 'message'">
        <span class="rt-tab-icon">💬</span>
        <span>消息</span>
      </div>
      <div class="rt-tab" :class="{ active: activeTab === 'history' }" @click="activeTab = 'history'">
        <span class="rt-tab-icon">🕐</span>
        <span>历史</span>
        <span v-if="props.historyTotal > 0" class="rt-tab-count">{{ props.historyTotal }}</span>
      </div>
      <div class="rt-tab" :class="{ active: activeTab === 'ddl' }" @click="activeTab = 'ddl'">
        <span class="rt-tab-icon">📝</span>
        <span>DDL</span>
      </div>
      <div class="rt-tab" :class="{ active: activeTab === 'favorite' }" @click="activeTab = 'favorite'">
        <span class="rt-tab-icon">⭐</span>
        <span>收藏</span>
        <span v-if="favoriteList.length > 0" class="rt-tab-count">{{ favoriteList.length }}</span>
      </div>
      <div class="rt-tabs-spacer"></div>
      <div v-if="activeTab === 'result' && hasDataResults" class="rt-view-toggle">
        <button 
          class="rt-view-btn" 
          :class="{ active: viewMode === 'table' }" 
          @click="viewMode = 'table'"
          title="表格视图"
        >
          📊 表格
        </button>
        <button 
          class="rt-view-btn" 
          :class="{ active: viewMode === 'shell' }" 
          @click="viewMode = 'shell'"
          title="Shell视图"
        >
          🖥️ Shell
        </button>
      </div>
      <div class="rt-tabs-action" @click="emit('close-all')" title="清空结果">
        ✕
      </div>
    </div>

    <div class="rt-content">
      <!-- 结果 Tab -->
      <div v-show="activeTab === 'result'" class="rt-pane">
        <div v-if="results.length === 0" class="rt-empty">
          <div class="rt-empty-icon">📋</div>
          <div>执行 SQL 后，结果将在此显示</div>
        </div>
        <div v-else class="rt-result-list">
          <div v-for="(r, idx) in results" :key="'r-' + idx" class="rt-result-item">
            <!-- 结果头部 -->
            <div class="rt-result-header" @click="toggleResult(idx)">
              <span class="rt-result-arrow" :class="{ expanded: expandedResults[idx] }">▶</span>
              <span class="rt-result-title">{{ r.title || '结果 ' + (idx + 1) }}</span>
              <span v-if="r.affectedRows !== undefined && r.affectedRows !== null" class="rt-result-affected">影响 {{ r.affectedRows }} 行</span>
              <span v-if="safeRows(r).length > 0" class="rt-result-rows">{{ safeRows(r).length }} 行</span>
              <span v-if="r.durationMs" class="rt-result-dura">{{ r.durationMs }} ms</span>
              <span class="rt-result-status" :class="r.success !== false ? 'ok' : 'err'">{{ r.success !== false ? '成功' : '失败' }}</span>
              
              <!-- 工具栏按钮 -->
              <div class="rt-result-actions">
                <button class="rt-action-btn rt-action-btn-add" @click.stop="addNewRow(r, idx)" :disabled="!canEdit(r)" title="添加记录">
                  <span class="rt-action-icon">+</span>
                </button>
                <button class="rt-action-btn rt-action-btn-delete" @click.stop="deleteSelectedRows(r, idx)" :disabled="!hasSelectedRows(r, idx) || isSubmitting[idx]" title="删除记录">
                  <span class="rt-action-icon">-</span>
                </button>
                <span class="rt-action-separator"></span>
                <button class="rt-action-btn rt-action-btn-submit" @click.stop="submitChanges(r, idx)" :disabled="!hasChanges(r, idx) || isSubmitting[idx]" title="提交">
                  <span class="rt-action-icon">✓</span>
                </button>
                <button class="rt-action-btn rt-action-btn-rollback" @click.stop="rollbackChanges(r, idx)" :disabled="!hasChanges(r, idx) || isSubmitting[idx]" title="放弃更改">
                  <span class="rt-action-icon">✕</span>
                </button>
                <span class="rt-action-separator"></span>
                <button class="rt-action-btn rt-action-btn-refresh" @click.stop="refreshResult(r, idx)" :disabled="isSubmitting[idx]" title="刷新">
                  <span class="rt-action-icon">⟳</span>
                </button>
                <button class="rt-action-btn rt-action-btn-export" @click.stop="toggleExportMenu(r.id)" title="导出">
                  <span class="rt-action-icon">⤤</span>
                </button>
                <button class="rt-action-btn rt-action-btn-stop" @click.stop="handleStop(r, idx)" :disabled="!isSubmitting[idx]" title="停止">
                  <span class="rt-action-icon">■</span>
                </button>
              </div>
              
              <!-- 导出菜单 -->
              <div v-if="exportMenuOpen === r.id" class="rt-export-menu">
                <div class="rt-export-item" @click="exportData(r, idx, 'csv', false)">导出全表 (CSV)</div>
                <div class="rt-export-item" @click="exportData(r, idx, 'csv', true)">导出选中行 (CSV)</div>
                <div class="rt-export-item" @click="exportData(r, idx, 'sql', false)">导出全表 (SQL)</div>
                <div class="rt-export-item" @click="exportData(r, idx, 'sql', true)">导出选中行 (SQL)</div>
              </div>
              
              <span class="rt-result-close" @click.stop="emit('close-result', idx)">✕</span>
            </div>
            
            <!-- 结果内容 -->
            <div v-if="expandedResults[idx]" class="rt-result-body">
              <div v-if="r.success === false" class="rt-error-box">
                <span class="rt-error-title">错误信息</span>
                <div class="rt-error-text">{{ r.message || r.error || '未知错误' }}</div>
              </div>
              <template v-else-if="safeRows(r).length > 0">
                <!-- 表格视图 -->
                <div v-show="viewMode === 'table'" class="rt-table-container">
                  <div class="rt-table-wrapper">
                    <table class="rt-data-table">
                      <thead>
                        <tr>
                          <th class="rt-select-col">
                            <input type="checkbox" @change="toggleSelectAll(r, idx)" :checked="isAllSelected(r, idx)" />
                          </th>
                          <th class="rt-row-num-col">#</th>
                          <th v-for="(col, ci) in safeColumns(r)" :key="'h-' + ci" class="rt-col-header">
                            <div class="rt-col-name">{{ col }}</div>
                          </th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr v-for="(row, ri) in paginatedRows(r, idx)" :key="'row-' + ri" :class="{ odd: ri % 2 === 0, editing: editingCell[r.id + '-' + ri], modified: isRowModified(r.id, idx, ri) }">
                          <td class="rt-select-col">
                            <input type="checkbox" v-model="selectedRows[r.id + '-' + ri]" />
                          </td>
                          <td class="rt-row-num-col">{{ ((currentPages[idx] || 1) - 1) * (pageSizes[idx] || 50) + ri + 1 }}</td>
                          <td 
                            v-for="(col, ci) in safeColumns(r)" 
                            :key="'c-' + ci" 
                            :title="formatCell(row, col)"
                            @dblclick="startEdit(r, idx, ri, col)"
                            class="rt-cell"
                          >
                            <template v-if="editingCell[r.id + '-' + ri] === col">
                              <input 
                                type="text" 
                                v-model="editValue" 
                                @blur="endEdit(r, idx, ri, col)"
                                @keyup.enter="endEdit(r, idx, ri, col)"
                                @keyup.escape="cancelEdit(r, idx, ri)"
                                class="rt-edit-input"
                                ref="editInputRef"
                              />
                            </template>
                            <template v-else>
                              <span :class="{ 'rt-null': isNullValue(row[col]) }">{{ formatCell(row, col) }}</span>
                            </template>
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                </div>
                
                <!-- Shell视图 -->
                <div v-show="viewMode === 'shell'" class="rt-shell-wrap">
                  <pre class="rt-shell-output">{{ formatShellOutput(r) }}</pre>
                </div>
                
                <!-- 状态栏 -->
                <div class="rt-status-bar">
                  <div class="rt-status-left">
                    <span class="rt-status-sql">{{ r.sql || 'No SQL' }}</span>
                  </div>
                  <div class="rt-status-center">
                    <button class="rt-page-btn" @click="firstPage(idx)" :disabled="(currentPages[idx] || 1) <= 1" title="首页">
                      <svg viewBox="0 0 16 16" width="12" height="12"><path d="M2 4v8h2V4H2zm4 0l4 4-4 4V4z"/></svg>
                    </button>
                    <button class="rt-page-btn" @click="prevPage(idx)" :disabled="(currentPages[idx] || 1) <= 1" title="上一页">
                      <svg viewBox="0 0 16 16" width="12" height="12"><path d="M6 4l-4 4 4 4V4z"/></svg>
                    </button>
                    <span class="rt-status-pagination">第 {{ currentPages[idx] || 1 }} 页 (共 {{ totalPages(r, idx) }} 页)</span>
                    <button class="rt-page-btn" @click="nextPage(r, idx)" :disabled="(currentPages[idx] || 1) >= totalPages(r, idx)" title="下一页">
                      <svg viewBox="0 0 16 16" width="12" height="12"><path d="M10 4l4 4-4 4V4z"/></svg>
                    </button>
                    <button class="rt-page-btn" @click="lastPage(r, idx)" :disabled="(currentPages[idx] || 1) >= totalPages(r, idx)" title="末页">
                      <svg viewBox="0 0 16 16" width="12" height="12"><path d="M14 4v8h-2V4h2zm-4 0l-4 4 4 4V4z"/></svg>
                    </button>
                    <select v-model="pageSizes[idx]" class="rt-page-size">
                      <option :value="20">20</option>
                      <option :value="50">50</option>
                      <option :value="100">100</option>
                      <option :value="500">500</option>
                      <option :value="1000">1000</option>
                    </select>
                    <span class="rt-page-label">条/页</span>
                  </div>
                  <div class="rt-status-right">
                    <span class="rt-status-records">{{ safeRows(r).length }} 条记录</span>
                    <span class="rt-status-separator">|</span>
                    <span class="rt-status-updated">上次更新: {{ lastUpdated[idx] || '-' }}</span>
                  </div>
                </div>
              </template>
              <div v-else class="rt-message-box">
                <div class="rt-message-text">{{ r.message || '执行成功' }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 执行计划 Tab -->
      <div v-show="activeTab === 'explain'" class="rt-pane">
        <div v-if="explainData.length === 0" class="rt-empty">
          <div class="rt-empty-icon">📊</div>
          <div>暂无执行计划数据。点击工具栏「解释」按钮可生成执行计划</div>
        </div>
        <div v-else class="rt-result-list">
          <div v-for="(exp, idx) in explainData" :key="'exp-' + idx" class="rt-result-item">
            <!-- 执行计划头部 -->
            <div class="rt-result-header" @click="toggleResult(idx)">
              <span class="rt-result-arrow" :class="{ expanded: expandedResults[idx] }">▶</span>
              <span class="rt-result-title">{{ exp.title || '执行计划 ' + (idx + 1) }}</span>
              <span v-if="getExplainRows(exp).length > 0" class="rt-result-rows">{{ getExplainRows(exp).length }} 行</span>
              <span v-if="exp.durationMs" class="rt-result-dura">{{ exp.durationMs }} ms</span>
              <span class="rt-result-status" :class="exp.success !== false ? 'ok' : 'err'">{{ exp.success !== false ? '成功' : '失败' }}</span>
              <span class="rt-result-close" @click.stop="closeExplain(idx)">✕</span>
            </div>
            
            <!-- 执行计划内容 -->
            <div v-if="expandedResults[idx]" class="rt-result-body">
              <div v-if="exp.success === false" class="rt-error-box">
                <span class="rt-error-title">错误信息</span>
                <div class="rt-error-text">{{ exp.message || exp.error || '未知错误' }}</div>
              </div>
              <template v-else-if="getExplainRows(exp).length > 0">
                <div class="rt-table-container">
                  <div class="rt-table-wrapper">
                    <table class="rt-data-table">
                      <thead>
                        <tr>
                          <th class="rt-row-num-col">#</th>
                          <th v-for="(col, ci) in getExplainColumns(exp)" :key="'h-' + ci" class="rt-col-header">
                            <div class="rt-col-name">{{ col }}</div>
                          </th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr v-for="(row, ri) in getExplainRows(exp)" :key="'row-' + ri" :class="{ odd: ri % 2 === 0 }">
                          <td class="rt-row-num-col">{{ ri + 1 }}</td>
                          <td v-for="(col, ci) in getExplainColumns(exp)" :key="'c-' + ci" :title="formatCell(row, col)">
                            <span :class="{ 'rt-null': isNullValue(row[col]) }">{{ formatCell(row, col) }}</span>
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                </div>
                
                <!-- 状态栏 -->
                <div class="rt-status-bar">
                  <div class="rt-status-left">
                    <span class="rt-status-sql">{{ exp.sql || 'EXPLAIN ' + props.sql }}</span>
                  </div>
                  <div class="rt-status-center">
                    <span class="rt-status-pagination">第 {{ currentPages[idx] || 1 }} 页 (共 {{ totalPages(exp, idx) }} 页)</span>
                    <select v-model="pageSizes[idx]" class="rt-page-size">
                      <option :value="20">20</option>
                      <option :value="50">50</option>
                      <option :value="100">100</option>
                      <option :value="500">500</option>
                      <option :value="1000">1000</option>
                    </select>
                    <span class="rt-page-label">条/页</span>
                  </div>
                  <div class="rt-status-right">
                    <span class="rt-status-records">{{ getExplainRows(exp).length }} 条记录</span>
                    <span class="rt-status-separator">|</span>
                    <span class="rt-status-updated">上次更新: {{ lastUpdated[idx] || '-' }}</span>
                  </div>
                </div>
              </template>
              <div v-else class="rt-message-box">
                <div class="rt-message-text">{{ exp.message || '执行成功' }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 消息 Tab -->
      <div v-show="activeTab === 'message'" class="rt-pane rt-message-pane">
        <div v-if="messageList.length === 0" class="rt-empty">
          <div class="rt-empty-icon">💬</div>
          <div>暂无消息</div>
        </div>
        <div v-for="(msg, idx) in messageList" :key="'m-' + idx" class="rt-message-item" :class="msg.level">
          <span class="rt-message-time">[{{ msg.time }}]</span>
          <span class="rt-message-level">[{{ msg.level }}]</span>
          <span class="rt-message-content">{{ msg.text }}</span>
        </div>
      </div>

      <!-- 历史 Tab -->
      <div v-show="activeTab === 'history'" class="rt-pane">
        <div v-if="historyList.length === 0" class="rt-empty">
          <div class="rt-empty-icon">🕐</div>
          <div>暂无执行历史</div>
        </div>
        <template v-else>
          <div class="rt-table-wrap">
            <table class="rt-data-table">
              <thead>
                <tr>
                  <th style="width: 150px;">时间</th>
                  <th style="width: 120px;">数据源</th>
                  <th style="width: 100px;">数据库</th>
                  <th>SQL</th>
                  <th style="width: 60px;">状态</th>
                  <th style="width: 80px;">影响行数</th>
                  <th style="width: 80px;">耗时</th>
                  <th style="width: 80px;">操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(h, idx) in paginatedHistory" :key="'h-' + idx" :class="{ odd: idx % 2 === 0 }">
                  <td>{{ h.createdAt || h.time || '-' }}</td>
                  <td :title="h.datasourceName">{{ (h.datasourceName || '').substring(0, 15) }}</td>
                  <td>{{ h.database || '-' }}</td>
                  <td class="rt-history-sql" :title="h.sql">{{ h.sql || '' }}</td>
                  <td><span :class="h.success !== false ? 'rt-tag-ok' : 'rt-tag-err'">{{ h.success !== false ? '成功' : '失败' }}</span></td>
                  <td>{{ h.affectedRows || 0 }}</td>
                  <td>{{ h.durationMs ? h.durationMs + ' ms' : '-' }}</td>
                  <td>
                    <button class="rt-history-action" @click="replayHistory(h)">重运行</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <div class="rt-history-pagination">
            <div class="rt-history-pagination-left">
              <button class="rt-refresh-btn" @click="emit('refresh')" title="刷新历史">⟳</button>
            </div>
            <div class="rt-history-pagination-right">
              <button class="rt-page-btn" @click="goHistoryPage(1)" :disabled="historyPage <= 1" title="首页">
                <svg viewBox="0 0 16 16" width="12" height="12"><path d="M2 4v8h2V4H2zm4 0l4 4-4 4V4z"/></svg>
              </button>
              <button class="rt-page-btn" @click="goHistoryPage(historyPage - 1)" :disabled="historyPage <= 1" title="上一页">
                <svg viewBox="0 0 16 16" width="12" height="12"><path d="M6 4l-4 4 4 4V4z"/></svg>
              </button>
              <span class="rt-page-info"><strong style="color: #28a745;">第 {{ historyPage }} 页</strong> (共 {{ historyTotalPages }} 页)</span>
              <button class="rt-page-btn" @click="goHistoryPage(historyPage + 1)" :disabled="historyPage >= historyTotalPages" title="下一页">
                <svg viewBox="0 0 16 16" width="12" height="12"><path d="M10 4l4 4-4 4V4z"/></svg>
              </button>
              <button class="rt-page-btn" @click="goHistoryPage(historyTotalPages)" :disabled="historyPage >= historyTotalPages" title="末页">
                <svg viewBox="0 0 16 16" width="12" height="12"><path d="M14 4v8h-2V4h2zm-4 0l-4 4 4 4V4z"/></svg>
              </button>
              <span class="rt-page-separator"></span>
              <select class="rt-page-size" v-model="historyPageSize" @change="onHistoryPageSizeChange">
                <option :value="10">10</option>
                <option :value="20">20</option>
                <option :value="50">50</option>
                <option :value="100">100</option>
              </select>
              <span class="rt-page-label">条/页</span>
              <span class="rt-page-records"><strong style="color: #1976d2;">{{ props.historyTotal || historyList.length }}</strong> 条记录</span>
              <span class="rt-page-separator">|</span>
              <span class="rt-page-updated">上次更新: {{ lastUpdateTime }}</span>
            </div>
          </div>
        </template>
      </div>

      <!-- DDL Tab -->
      <div v-show="activeTab === 'ddl'" class="rt-pane">
        <div v-if="!ddlText" class="rt-empty">
          <div class="rt-empty-icon">📝</div>
          <div>在左侧树节点右键「查看 DDL」，DDL 将在此显示</div>
        </div>
        <pre v-else class="rt-ddl-text">{{ ddlText }}</pre>
      </div>

      <!-- 收藏 Tab -->
      <div v-show="activeTab === 'favorite'" class="rt-pane">
        <div v-if="favoriteList.length === 0" class="rt-empty">
          <div class="rt-empty-icon">⭐</div>
          <div>暂无收藏的 SQL</div>
          <div style="font-size: 11px; color: #909399; margin-top: 8px;">点击工具栏「收藏」按钮可收藏当前 SQL</div>
        </div>
        <template v-else>
          <div class="rt-table-wrap">
            <table class="rt-data-table">
              <thead>
                <tr>
                  <th style="width: 150px;">创建时间</th>
                  <th style="width: 150px;">标题</th>
                  <th style="width: 200px;">描述</th>
                  <th>SQL 内容</th>
                  <th style="width: 160px;">操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(f, idx) in favoriteList" :key="'fav-' + f.id" :class="{ odd: idx % 2 === 0 }">
                  <td>{{ f.createdAt || '-' }}</td>
                  <td :title="f.title">{{ f.title || '-' }}</td>
                  <td :title="f.description">{{ (f.description || '-').substring(0, 30) }}{{ (f.description || '').length > 30 ? '...' : '' }}</td>
                  <td class="rt-fav-sql" :title="f.sql">{{ (f.sql || '').substring(0, 50) }}{{ (f.sql || '').length > 50 ? '...' : '' }}</td>
                  <td>
                    <button class="rt-fav-action" @click="applyFavorite(f)">插入编辑器</button>
                    <button class="rt-fav-action rt-fav-action-danger" @click="removeFavoriteItem(f.id)">取消收藏</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { executeSql } from '@/api/sql'
import { detectExplicitDatabase } from '../utils/sql'

interface FavoriteItem {
  id: string
  title: string
  description: string
  sql: string
  createdAt: string
}

const props = withDefaults(defineProps<{
  results?: any[]
  history?: any[]
  historyTotal?: number
  messages?: any[]
  explain?: any[]
  ddlText?: string
  favorites?: FavoriteItem[]
  sql?: string
  datasourceId?: string
  database?: string
}>(), {
  results: () => [],
  history: () => [],
  historyTotal: 0,
  messages: () => [],
  explain: () => [],
  ddlText: '',
  favorites: () => [],
  sql: '',
  datasourceId: '',
  database: ''
})

const emit = defineEmits<{
  (e: 'close-result', i: number): void
  (e: 'close-all'): void
  (e: 'replay', sql: string): void
  (e: 'refresh'): void
  (e: 'apply-favorite', sql: string): void
  (e: 'remove-favorite', id: string): void
  (e: 'refresh-single', rId: string, idx: number, sql: string): void
  (e: 'load-history-page', page: number, pageSize: number): void
}>()

const activeTab = ref('result')
const viewMode = ref<'table' | 'shell'>('table')
const currentPages = reactive<Record<number, number>>({})
const pageSizes = reactive<Record<number, number>>({})
const expandedResults = reactive<Record<number, boolean>>({})
const historyPage = ref(1)
const historyPageSize = ref(50)
const lastUpdated = reactive<Record<number, string>>({})
const lastUpdateTime = ref('')

// 编辑相关状态
const editingCell = reactive<Record<string, string>>({})
const selectedRows = reactive<Record<string, boolean>>({})
const editValue = ref('')
const modifiedRows = reactive<Record<string, any>>({})
const deletedRows = reactive<Record<string, any>>({})
const originalRows = reactive<Record<string, any[]>>({})

// 执行状态
const isSubmitting = reactive<Record<number, boolean>>({})
const abortSubmit = reactive<Record<number, boolean>>({})

// 导出菜单状态
const exportMenuOpen = ref<string | null>(null)

const resultCount = computed(() => props.results.filter((r: any) => r.success !== false).length)
const historyList = computed(() => {
  const list = props.history || []
  // 按时间倒序排序
  return [...list].sort((a: any, b: any) => {
    const timeA = a.time || a.createdAt || ''
    const timeB = b.time || b.createdAt || ''
    return timeB.localeCompare(timeA)
  })
})
const messageList = computed(() => props.messages || [])
const explainData = computed(() => props.explain || [])

const ddlText = computed(() => props.ddlText || '')
const favoriteList = computed(() => {
  const list = props.favorites || []
  // 按创建时间倒序排序
  return [...list].sort((a: any, b: any) => {
    const timeA = a.createdAt || ''
    const timeB = b.createdAt || ''
    return timeB.localeCompare(timeA)
  })
})

const hasDataResults = computed(() => {
  return props.results.some((r: any) => {
    return safeRows(r).length > 0
  })
})

const historyTotalPages = computed(() => {
  const total = props.historyTotal || historyList.value.length
  return Math.max(1, Math.ceil(total / historyPageSize.value))
})

function goHistoryPage(page: number) {
  const total = historyTotalPages.value
  if (page < 1) page = 1
  if (page > total) page = total
  historyPage.value = page
  emit('load-history-page', page, historyPageSize.value)
}

function onHistoryPageSizeChange() {
  historyPage.value = 1
  emit('load-history-page', 1, historyPageSize.value)
}

const paginatedHistory = computed(() => {
  return historyList.value
})

const explainColumns = computed(() => {
  if (explainData.value.length > 0) {
    return Object.keys(explainData.value[0])
  }
  return ['id', 'select_type', 'table', 'type', 'possible_keys', 'key', 'key_len', 'ref', 'rows', 'Extra']
})

function getExplainColumns(exp: any): string[] {
  const rows = getExplainRows(exp)
  if (rows.length > 0 && typeof rows[0] === 'object' && rows[0] !== null) {
    return Object.keys(rows[0])
  }
  return ['id', 'select_type', 'table', 'type', 'possible_keys', 'key', 'key_len', 'ref', 'rows', 'Extra']
}

function getExplainRows(exp: any): any[] {
  if (!exp) return []
  if (Array.isArray(exp.rows)) return exp.rows
  if (Array.isArray(exp.data)) return exp.data
  return []
}

watch(() => props.results.length, () => {
  for (let idx = 0; idx < props.results.length; idx++) {
    if (expandedResults[idx] === undefined) expandedResults[idx] = true
    if (currentPages[idx] === undefined) currentPages[idx] = 1
    if (pageSizes[idx] === undefined) pageSizes[idx] = 50
    lastUpdated[idx] = new Date().toLocaleString('zh-CN')
    const r = props.results[idx]
    if (r && r.rows && Array.isArray(r.rows)) {
      const rId = r.id + '-' + idx
      originalRows[rId] = r.rows.map((row: any) => ({ ...row }))
    }
  }
  if (props.results.length > 0) {
    activeTab.value = 'result'
  }
})

watch(() => props.explain?.length, (n: any) => {
  if (n && n > 0) activeTab.value = 'explain'
})

watch(() => props.ddlText, (n: string) => {
  if (n && n.trim().length > 0) activeTab.value = 'ddl'
})

watch(() => props.history?.length, () => {
  lastUpdateTime.value = new Date().toLocaleString('zh-CN')
})

function toggleResult(idx: number) {
  expandedResults[idx] = !expandedResults[idx]
}

function safeRows(r: any): any[] {
  if (!r) return []
  if (Array.isArray(r.rows)) return r.rows
  if (Array.isArray(r.data)) return r.data
  return []
}

function safeColumns(r: any): string[] {
  if (!r) return []
  if (Array.isArray(r.columns) && r.columns.length > 0) {
    return r.columns.map((c: any) => typeof c === 'string' ? c : (c.name || String(c)))
  }
  const rows = safeRows(r)
  if (rows.length > 0 && typeof rows[0] === 'object' && rows[0] !== null) {
    return Object.keys(rows[0])
  }
  return []
}

function formatCell(row: any, col: string): string {
  if (row === null || row === undefined) return ''
  if (typeof row === 'object' && col in row) {
    const v = row[col]
    if (v === null || v === undefined) return ''
    if (typeof v === 'object') return JSON.stringify(v)
    return String(v)
  }
  return ''
}

function isNullValue(val: any): boolean {
  return val === null || val === undefined || val === ''
}

function closeExplain(idx: number) {
  const explain = explainData.value
  if (explain && explain[idx] !== undefined) {
    explain.splice(idx, 1)
  }
}

function isRowModified(resultId: string, resultIdx: number, rowIdx: number): boolean {
  const page = currentPages[resultIdx] || 1
  const ps = pageSizes[resultIdx] || 50
  const actualIdx = (page - 1) * ps + rowIdx
  const rId = resultId + '-' + resultIdx
  const rows = modifiedRows[rId]
  return rows !== undefined && actualIdx in rows
}

function formatShellOutput(r: any): string {
  const rows = safeRows(r)
  const cols = safeColumns(r)
  if (rows.length === 0 || cols.length === 0) return ''
  
  const output: string[] = []
  const colWidths: number[] = []
  
  cols.forEach((col, ci) => {
    let maxWidth = String(col).length
    rows.forEach(row => {
      const cell = formatCell(row, col)
      maxWidth = Math.max(maxWidth, cell.length)
    })
    colWidths[ci] = maxWidth + 2
  })
  
  const header = cols.map((col, ci) => String(col).padEnd(colWidths[ci])).join('')
  output.push(header)
  
  const separator = colWidths.map(w => '-'.repeat(w - 1) + ' ').join('')
  output.push(separator)
  
  rows.forEach(row => {
    const line = cols.map((col, ci) => formatCell(row, col).padEnd(colWidths[ci])).join('')
    output.push(line)
  })
  
  return output.join('\n')
}

function paginatedRows(r: any, idx: number): any[] {
  const rows = safeRows(r)
  const page = currentPages[idx] || 1
  const ps = pageSizes[idx] || 50
  const start = (page - 1) * ps
  return rows.slice(start, start + ps)
}

function totalPages(r: any, idx: number): number {
  const total = safeRows(r).length
  const ps = pageSizes[idx] || 50
  return Math.max(1, Math.ceil(total / ps))
}

function firstPage(idx: number) {
  currentPages[idx] = 1
}

function prevPage(idx: number) {
  currentPages[idx] = Math.max(1, (currentPages[idx] || 1) - 1)
}

function nextPage(r: any, idx: number) {
  currentPages[idx] = Math.min(totalPages(r, idx), (currentPages[idx] || 1) + 1)
}

function lastPage(r: any, idx: number) {
  currentPages[idx] = totalPages(r, idx)
}

function replayHistory(h: any) {
  if (h && h.sql) {
    emit('replay', h.sql)
  }
}

// === 编辑相关方法 ===
function startEdit(r: any, resultIdx: number, rowIdx: number, col: string) {
  const rows = safeRows(r)
  const page = currentPages[resultIdx] || 1
  const ps = pageSizes[resultIdx] || 50
  const actualRowIdx = (page - 1) * ps + rowIdx
  
  if (actualRowIdx < rows.length) {
    const row = rows[actualRowIdx]
    editValue.value = formatCell(row, col)
    editingCell[r.id + '-' + rowIdx] = col
    nextTick(() => {
      const inputs = document.querySelectorAll('.rt-edit-input')
      if (inputs.length > 0) {
        (inputs[inputs.length - 1] as HTMLInputElement).focus()
      }
    })
  }
}

function endEdit(r: any, resultIdx: number, rowIdx: number, col: string) {
  const rows = safeRows(r)
  const page = currentPages[resultIdx] || 1
  const ps = pageSizes[resultIdx] || 50
  const actualRowIdx = (page - 1) * ps + rowIdx
  const rId = r.id + '-' + resultIdx
  
  if (actualRowIdx < rows.length) {
    const row = rows[actualRowIdx]
    const originalVal = originalRows[rId]?.[actualRowIdx]?.[col]
    
    if (editValue.value !== formatCell(originalVal, col)) {
      const currentModified = modifiedRows[rId]
      if (!currentModified || !(actualRowIdx in currentModified)) {
        modifiedRows[rId] = currentModified || {}
        modifiedRows[rId][actualRowIdx] = { ...(originalRows[rId]?.[actualRowIdx] || row) }
      }
      row[col] = editValue.value
    } else {
      if (modifiedRows[rId] && actualRowIdx in modifiedRows[rId]) {
        delete modifiedRows[rId][actualRowIdx]
      }
    }
  }
  editingCell[r.id + '-' + rowIdx] = ''
}

function cancelEdit(r: any, resultIdx: number, rowIdx: number) {
  editingCell[r.id + '-' + rowIdx] = ''
}

function toggleSelectAll(r: any, resultIdx: number) {
  const isChecked = !isAllSelected(r, resultIdx)
  const paginated = paginatedRows(r, resultIdx)
  
  paginated.forEach((row, idx) => {
    selectedRows[r.id + '-' + idx] = isChecked
  })
}

function isAllSelected(r: any, resultIdx: number): boolean {
  const paginated = paginatedRows(r, resultIdx)
  return paginated.length > 0 && paginated.every((row, idx) => selectedRows[r.id + '-' + idx])
}

function hasSelectedRows(r: any, resultIdx: number): boolean {
  const paginated = paginatedRows(r, resultIdx)
  return paginated.some((row, idx) => selectedRows[r.id + '-' + idx])
}

// === 工具栏操作方法 ===
function canEdit(r: any): boolean {
  if (!props.datasourceId) return false
  const rows = safeRows(r)
  const cols = safeColumns(r)
  return rows.length > 0 && cols.length > 0
}

function hasChanges(r: any, resultIdx: number): boolean {
  const rId = r.id + '-' + resultIdx
  const modified = modifiedRows[rId] ? Object.keys(modifiedRows[rId]).length > 0 : false
  const deleted = deletedRows[rId] ? Object.keys(deletedRows[rId]).length > 0 : false
  return modified || deleted
}

function toggleExportMenu(resultId: string) {
  exportMenuOpen.value = exportMenuOpen.value === resultId ? null : resultId
}

function refreshResult(r: any, resultIdx: number) {
  const rId = r.id + '-' + resultIdx
  
  if (r.sql) {
    emit('refresh-single', rId, resultIdx, r.sql)
  } else if (props.sql) {
    emit('refresh-single', rId, resultIdx, props.sql)
  } else {
    emit('refresh')
  }
  lastUpdated[resultIdx] = new Date().toLocaleString('zh-CN')
}

function handleStop(r: any, resultIdx: number) {
  if (isSubmitting[resultIdx]) {
    abortSubmit[resultIdx] = true
    ElMessage.info('正在停止提交...')
  }
}

function addNewRow(r: any, resultIdx: number) {
  const rows = safeRows(r)
  const cols = safeColumns(r)
  if (rows.length === 0 || cols.length === 0) {
    ElMessage.warning('无法添加新行：表格数据未加载')
    return
  }
  
  const newRow: Record<string, any> = {}
  cols.forEach(col => {
    newRow[col] = ''
  })
  rows.push(newRow)
  
  const rId = r.id + '-' + resultIdx
  const newIdx = rows.length - 1
  if (!originalRows[rId]) {
    originalRows[rId] = rows.map((row: any) => ({ ...row }))
  } else {
    originalRows[rId].push({ ...newRow })
  }
  
  if (!modifiedRows[rId]) modifiedRows[rId] = {}
  modifiedRows[rId][newIdx] = newRow
  
  ElMessage.success('已添加新行')
  
  setTimeout(() => {
    currentPages[resultIdx] = Math.ceil(rows.length / (pageSizes[resultIdx] || 50))
  }, 0)
}

function deleteSelectedRows(r: any, resultIdx: number) {
  const page = currentPages[resultIdx] || 1
  const ps = pageSizes[resultIdx] || 50
  const rId = r.id + '-' + resultIdx
  
  if (!r || !r.rows || !Array.isArray(r.rows)) {
    Object.keys(selectedRows).forEach(key => {
      if (key.startsWith(r.id + '-')) {
        delete selectedRows[key]
      }
    })
    return
  }
  
  const rows = r.rows
  
  if (rows.length === 0) {
    Object.keys(selectedRows).forEach(key => {
      if (key.startsWith(r.id + '-')) {
        delete selectedRows[key]
      }
    })
    return
  }
  
  const selectedIndices: number[] = []
  const paginated = paginatedRows(r, resultIdx)
  
  paginated.forEach((row, idx) => {
    const key = r.id + '-' + idx
    if (selectedRows[key]) {
      const actualIdx = (page - 1) * ps + idx
      selectedIndices.push(actualIdx)
    }
  })
  
  if (selectedIndices.length === 0) {
    ElMessage.warning('请先选中要删除的行')
    return
  }
  
  if (!deletedRows[rId]) deletedRows[rId] = {}
  
  selectedIndices.sort((a, b) => b - a).forEach(idx => {
    const deletedRow = rows[idx]
    deletedRows[rId][idx] = deletedRow
    rows.splice(idx, 1)
    
    if (originalRows[rId]) {
      originalRows[rId].splice(idx, 1)
    }
    
    if (modifiedRows[rId]) {
      delete modifiedRows[rId][idx]
      
      Object.keys(modifiedRows[rId]).forEach(key => {
        const numKey = parseInt(key)
        if (numKey > idx) {
          modifiedRows[rId][numKey - 1] = modifiedRows[rId][numKey]
          delete modifiedRows[rId][numKey]
        }
      })
    }
    
    if (deletedRows[rId]) {
      Object.keys(deletedRows[rId]).forEach(key => {
        const numKey = parseInt(key)
        if (numKey > idx) {
          deletedRows[rId][numKey - 1] = deletedRows[rId][numKey]
          delete deletedRows[rId][numKey]
        }
      })
    }
  })
  
  Object.keys(selectedRows).forEach(key => {
    if (key.startsWith(r.id + '-')) {
      delete selectedRows[key]
    }
  })
  
  const newTotalPages = Math.max(1, Math.ceil(rows.length / ps))
  if ((currentPages[resultIdx] || 1) > newTotalPages) {
    currentPages[resultIdx] = newTotalPages
  }
  
  ElMessage.success(`已标记 ${selectedIndices.length} 行待删除，点击提交确认删除`)
}

async function submitChanges(r: any, resultIdx: number) {
  const rId = r.id + '-' + resultIdx
  
  console.log('=== 开始提交更改 ===')
  console.log('结果ID:', rId)
  console.log('props.datasourceId:', props.datasourceId)
  console.log('props.database:', props.database)
  console.log('r.sql:', r.sql)
  console.log('props.sql:', props.sql)
  
  const modified = modifiedRows[rId] || {}
  const deleted = deletedRows[rId] || {}
  const modifiedCount = Object.keys(modified).length
  const deletedCount = Object.keys(deleted).length
  
  console.log('修改的行数:', modifiedCount)
  console.log('删除的行数:', deletedCount)
  
  if (modifiedCount === 0 && deletedCount === 0) {
    ElMessage.info('没有需要提交的更改')
    return
  }
  
  if (!props.datasourceId) {
    ElMessage.warning('无法提交：缺少数据源信息')
    return
  }
  
  const tableName = extractTableName(r)
  console.log('提取的表名:', tableName)
  
  if (!tableName) {
    ElMessage.warning('无法从SQL中识别表名，无法提交更改')
    return
  }
  
  const effectiveDb = props.database || detectExplicitDatabase(r.sql || props.sql || '', []).explicitDatabase || ''
  console.log('effectiveDb:', effectiveDb)
  
  if (!effectiveDb) {
    ElMessage.warning('无法确定数据库，无法提交更改')
    return
  }
  
  isSubmitting[resultIdx] = true
  abortSubmit[resultIdx] = false
  
  try {
    const cols = safeColumns(r)
    const pkCols = cols.filter(c => c.toLowerCase().includes('id'))
    const pkCol = pkCols[0] || cols[0]
    
    const modifiedKeys = Object.keys(modified)
    for (let i = 0; i < modifiedKeys.length; i++) {
      if (abortSubmit[resultIdx]) {
        ElMessage.warning('提交已取消')
        return
      }
      
      const idxStr = modifiedKeys[i]
      const idx = parseInt(idxStr)
      const original = modifiedRows[rId][idx]
      const rows = r && r.rows && Array.isArray(r.rows) ? r.rows : safeRows(r)
      const current = rows[idx]
      
      if (!original || !current) continue
      
      const pkValue = original[pkCol]
      
      if (pkValue === undefined || pkValue === null || pkValue === '') {
        const emptyRequired = cols.filter(col => {
          const val = current[col]
          return val === undefined || val === null || val === ''
        })
        
        if (emptyRequired.length > 0) {
          isSubmitting[resultIdx] = false
          ElMessage.error(`新增行失败：以下字段不能为空：${emptyRequired.join(', ')}`)
          return
        }
        
        await insertRow(current, effectiveDb, tableName)
      } else {
        await updateRow(current, pkCol, pkValue, effectiveDb, tableName)
      }
      
      await new Promise(resolve => setTimeout(resolve, 50))
    }
    
    if (!abortSubmit[resultIdx]) {
      const deletedKeys = Object.keys(deleted)
      for (let i = 0; i < deletedKeys.length; i++) {
        if (abortSubmit[resultIdx]) {
          ElMessage.warning('提交已取消')
          return
        }
        
        const idxStr = deletedKeys[i]
        const deletedRow = deletedRows[rId][idxStr]
        
        const pkValue = deletedRow[pkCol]
        
        if (pkValue !== undefined && pkValue !== null && pkValue !== '') {
          await deleteRow(deletedRow, pkCol, pkValue, effectiveDb, tableName)
        }
        
        await new Promise(resolve => setTimeout(resolve, 50))
      }
    }
    
    if (!abortSubmit[resultIdx]) {
      await refreshResult(r, resultIdx)
      ElMessage.success(`已成功提交 ${modifiedCount} 条修改和 ${deletedCount} 条删除`)
      
      Object.keys(modified).forEach(idx => {
        delete modifiedRows[rId][idx]
      })
      Object.keys(deleted).forEach(idx => {
        delete deletedRows[rId][idx]
      })
    }
  } catch (e: any) {
    if (!abortSubmit[resultIdx]) {
      console.error('提交失败:', e)
      ElMessage.error('提交失败: ' + (e?.message || String(e)))
    }
  } finally {
    isSubmitting[resultIdx] = false
    abortSubmit[resultIdx] = false
  }
}

function extractTableName(r: any): string | null {
  const sql = (r.sql || props.sql || '').trim()
  if (!sql) return null
  
  const fromMatch = sql.match(/FROM\s+(?:`([^`]+)`|([a-zA-Z_][\w$]*))(?:\.(?:`([^`]+)`|([a-zA-Z_][\w$]*)))?/i)
  if (fromMatch) {
    const table = fromMatch[3] || fromMatch[4] || fromMatch[1] || fromMatch[2]
    return table
  }
  
  const updateMatch = sql.match(/UPDATE\s+(?:`([^`]+)`|([a-zA-Z_][\w$]*))(?:\.(?:`([^`]+)`|([a-zA-Z_][\w$]*)))?/i)
  if (updateMatch) {
    const table = updateMatch[3] || updateMatch[4] || updateMatch[1] || updateMatch[2]
    return table
  }
  
  const insertMatch = sql.match(/INSERT\s+INTO\s+(?:`([^`]+)`|([a-zA-Z_][\w$]*))(?:\.(?:`([^`]+)`|([a-zA-Z_][\w$]*)))?/i)
  if (insertMatch) {
    const table = insertMatch[3] || insertMatch[4] || insertMatch[1] || insertMatch[2]
    return table
  }
  
  return null
}

async function insertRow(row: any, database: string, table: string) {
  if (!props.datasourceId) {
    throw new Error('缺少数据源ID')
  }
  
  const cols = Object.keys(row).filter(k => row[k] !== '')
  const values = cols.map(c => {
    const v = row[c]
    if (v === null || v === undefined) return 'NULL'
    if (typeof v === 'string') return `'${v.replace(/'/g, "''")}'`
    return String(v)
  })
  
  const sql = `INSERT INTO \`${database}\`.\`${table}\` (${cols.map(c => `\`${c}\``).join(', ')}) VALUES (${values.join(', ')})`
  console.log('执行INSERT:', sql)
  
  const result = await executeSql({ datasourceId: props.datasourceId, database, sql, ignoreRisk: true })
  console.log('INSERT结果:', result)
  
  return result
}

async function updateRow(row: any, pkCol: string, pkValue: any, database: string, table: string) {
  if (!props.datasourceId) {
    throw new Error('缺少数据源ID')
  }
  
  const sets = Object.keys(row).filter(k => k !== pkCol).map(c => {
    const v = row[c]
    if (v === null || v === undefined) return `\`${c}\` = NULL`
    if (typeof v === 'string') return `\`${c}\` = '${v.replace(/'/g, "''")}'`
    return `\`${c}\` = ${String(v)}`
  })
  
  const pkStr = typeof pkValue === 'string' ? `'${pkValue.replace(/'/g, "''")}'` : String(pkValue)
  const sql = `UPDATE \`${database}\`.\`${table}\` SET ${sets.join(', ')} WHERE \`${pkCol}\` = ${pkStr}`
  console.log('执行UPDATE:', sql)
  
  const result = await executeSql({ datasourceId: props.datasourceId, database, sql, ignoreRisk: true })
  console.log('UPDATE结果:', result)
  
  return result
}

async function deleteRow(row: any, pkCol: string, pkValue: any, database: string, table: string) {
  if (!props.datasourceId) {
    throw new Error('缺少数据源ID')
  }
  
  const pkStr = typeof pkValue === 'string' ? `'${pkValue.replace(/'/g, "''")}'` : String(pkValue)
  const sql = `DELETE FROM \`${database}\`.\`${table}\` WHERE \`${pkCol}\` = ${pkStr}`
  console.log('执行DELETE:', sql)
  
  const result = await executeSql({ datasourceId: props.datasourceId, database, sql, ignoreRisk: true })
  console.log('DELETE结果:', result)
  
  return result
}

function rollbackChanges(r: any, resultIdx: number) {
  const rId = r.id + '-' + resultIdx
  
  const modified = modifiedRows[rId] || {}
  const deleted = deletedRows[rId] || {}
  
  if (r && r.rows && Array.isArray(r.rows)) {
    const rows = r.rows
    
    Object.keys(modified).forEach(idxStr => {
      const idx = parseInt(idxStr)
      if (originalRows[rId] && originalRows[rId][idx]) {
        rows[idx] = { ...originalRows[rId][idx] }
      }
      delete modifiedRows[rId][idx]
    })
    
    const deletedKeys = Object.keys(deleted).map(k => parseInt(k)).sort((a, b) => b - a)
    deletedKeys.forEach(idx => {
      if (deletedRows[rId][idx]) {
        rows.splice(idx, 0, deletedRows[rId][idx])
        if (originalRows[rId]) {
          originalRows[rId].splice(idx, 0, deletedRows[rId][idx])
        }
        delete deletedRows[rId][idx]
      }
    })
  }
  
  ElMessage.success('已回滚所有更改')
}

function exportData(r: any, resultIdx: number, format: 'csv' | 'sql', selectedOnly: boolean) {
  exportMenuOpen.value = null
  
  const rows = safeRows(r)
  const cols = safeColumns(r)
  
  if (rows.length === 0 || cols.length === 0) {
    ElMessage.warning('无法导出：表格数据未加载')
    return
  }
  
  let exportRows: any[]
  if (selectedOnly) {
    exportRows = []
    const page = currentPages[resultIdx] || 1
    const ps = pageSizes[resultIdx] || 50
    const paginated = paginatedRows(r, resultIdx)
    
    paginated.forEach((row, idx) => {
      const key = r.id + '-' + idx
      if (selectedRows[key]) {
        exportRows.push(row)
      }
    })
    if (exportRows.length === 0) {
      ElMessage.warning('请先选中要导出的行')
      return
    }
  } else {
    exportRows = rows
  }
  
  let content = ''
  let filename = ''
  
  if (format === 'csv') {
    content = cols.join(',') + '\n'
    exportRows.forEach(row => {
      const values = cols.map(col => {
        const v = formatCell(row, col)
        if (v.includes(',') || v.includes('"') || v.includes('\n')) {
          return `"${v.replace(/"/g, '""')}"`
        }
        return v
      })
      content += values.join(',') + '\n'
    })
    filename = `result_export_${Date.now()}.csv`
  } else {
    content = `-- Export from query result\n`
    content += `-- Exported: ${new Date().toISOString()}\n\n`
    const tableName = 'result_table'
    exportRows.forEach(row => {
      const rowCols = cols.filter(col => row[col] !== undefined && row[col] !== null)
      const values = rowCols.map(col => {
        const v = row[col]
        if (v === null) return 'NULL'
        if (typeof v === 'string') return `'${v.replace(/'/g, "''")}'`
        return String(v)
      })
      content += `INSERT INTO \`${tableName}\` (`
      content += rowCols.map(c => `\`${c}\``).join(', ')
      content += ') VALUES (' + values.join(', ') + ');\n'
    })
    filename = `result_export_${Date.now()}.sql`
  }
  
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
  
  ElMessage.success(`已导出 ${exportRows.length} 行数据`)
}

// === 收藏相关方法 ===
function applyFavorite(f: FavoriteItem) {
  if (f && f.sql) {
    emit('apply-favorite', f.sql)
  }
}

function removeFavoriteItem(id: string) {
  emit('remove-favorite', id)
}
</script>

<style scoped>
.rt-root {
  flex: 1 1 60%;
  min-height: 0;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #ffffff;
  border-top: 1px solid #d0d0d0;
}

.rt-tabs-bar {
  display: flex;
  align-items: center;
  background: #f5f5f5;
  border-bottom: 1px solid #d0d0d0;
  height: 34px;
  flex-shrink: 0;
}

.rt-tab {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 0 14px;
  height: 34px;
  font-size: 12.5px;
  color: #606266;
  cursor: pointer;
  border-right: 1px solid #e0e0e0;
  transition: all 0.15s;
}

.rt-tab:hover {
  background: #eef3f8;
  color: #1976d2;
}

.rt-tab.active {
  background: #ffffff;
  color: #1976d2;
  border-bottom: 2px solid #1976d2;
  height: 33px;
  font-weight: 500;
}

.rt-tab-icon { font-size: 13px; }
.rt-tab-count {
  background: #1976d2;
  color: #ffffff;
  border-radius: 8px;
  padding: 1px 6px;
  font-size: 10.5px;
  font-weight: 500;
}

.rt-tab-count-blue {
  background: transparent;
  color: #1976d2;
  border: 1px solid #1976d2;
  border-radius: 8px;
  padding: 1px 6px;
  font-size: 10.5px;
  font-weight: 500;
}

.rt-tabs-spacer { flex: 1 1 auto; }

.rt-view-toggle {
  display: flex;
  gap: 4px;
  padding: 0 8px;
  border-left: 1px solid #e0e0e0;
}

.rt-view-btn {
  padding: 4px 8px;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 3px;
  font-size: 11px;
  cursor: pointer;
  color: #606266;
  transition: all 0.15s;
}

.rt-view-btn:hover {
  background: #eef3f8;
  color: #1976d2;
  border-color: #bbdefb;
}

.rt-view-btn.active {
  background: #1976d2;
  color: #ffffff;
  border-color: #1976d2;
}

.rt-tabs-action {
  padding: 0 14px;
  color: #909399;
  cursor: pointer;
  font-size: 14px;
  transition: color 0.15s;
}
.rt-tabs-action:hover { color: #c62828; }

.rt-content {
  flex: 1 1 auto;
  min-height: 0;
  overflow: auto;
}

.rt-pane {
  height: 100%;
  min-height: 0;
  overflow: auto;
}

.rt-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  min-height: 150px;
  color: #909399;
  font-size: 13px;
  gap: 10px;
}

.rt-empty-icon { font-size: 32px; opacity: 0.6; }

.rt-result-list { padding: 8px; }

.rt-result-item {
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  margin-bottom: 8px;
  overflow: hidden;
}

.rt-result-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 12px;
  background: #f5f7fa;
  cursor: pointer;
  font-size: 12.5px;
  transition: background 0.1s;
}

.rt-result-header:hover { background: #eef3f8; }

.rt-result-arrow {
  font-size: 10px;
  color: #1976d2;
  transition: transform 0.15s;
  display: inline-block;
}

.rt-result-arrow.expanded {
  transform: rotate(90deg);
}

.rt-result-title { font-weight: 500; color: #303133; flex-shrink: 0; }
.rt-result-affected { color: #ef6c00; font-size: 11.5px; }
.rt-result-rows { color: #2e7d32; font-size: 11.5px; }
.rt-result-dura { color: #7e57c2; font-size: 11.5px; }

.rt-result-status {
  padding: 2px 8px;
  border-radius: 2px;
  font-size: 11px;
  font-weight: 500;
  flex-shrink: 0;
}
.rt-result-status.ok { background: #e8f5e9; color: #2e7d32; }
.rt-result-status.err { background: #ffebee; color: #c62828; }

.rt-result-actions {
  display: flex;
  align-items: center;
  gap: 1px;
  margin-left: auto;
  position: relative;
  background: linear-gradient(180deg, #f8f9fa 0%, #e9ecef 100%);
  border: 1px solid #dee2e6;
  border-radius: 4px;
  padding: 2px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.rt-result-close {
  color: #909399;
  cursor: pointer;
  padding: 0 8px;
  font-size: 12px;
}
.rt-result-close:hover { color: #c62828; }

.rt-result-body {
  background: #ffffff;
  padding: 0;
}

.rt-error-box {
  padding: 12px 16px;
  background: #fff5f5;
  border-left: 3px solid #c62828;
}
.rt-error-title { color: #c62828; font-weight: 600; font-size: 12.5px; margin-bottom: 4px; display: block; }
.rt-error-text { color: #424242; font-size: 12px; white-space: pre-wrap; font-family: Consolas, Monaco, monospace; }

.rt-message-box {
  padding: 16px;
  background: #f5f5f5;
  border-left: 3px solid #909399;
}
.rt-message-text { font-size: 13px; color: #424242; }

.rt-table-container {
  overflow: auto;
  max-height: 400px;
}

.rt-table-wrapper {
  overflow: auto;
}

.rt-table-wrap {
  max-height: 400px;
  overflow: auto;
  background: #ffffff;
}

.rt-data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
  font-family: Consolas, Monaco, 'Courier New', monospace;
  min-width: 100%;
}

.rt-data-table thead th {
  position: sticky;
  top: 0;
  background: linear-gradient(180deg, #e9ecef 0%, #dee2e6 100%);
  color: #495057;
  border: 1px solid #ced4da;
  border-bottom: 2px solid #adb5bd;
  padding: 6px 12px;
  text-align: left;
  font-weight: 600;
  white-space: nowrap;
  z-index: 1;
}

.rt-data-table thead th:hover {
  background: #e3f2fd;
}

.rt-data-table tbody td {
  border: 1px solid #dee2e6;
  padding: 5px 12px;
  color: #212529;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.rt-data-table tbody tr.odd { background: #ffffff; }
.rt-data-table tbody tr.even { background: #f6f8fa; }
.rt-data-table tbody tr:hover { background: #e7f3ff; }
.rt-data-table tbody tr.editing { background: #fff3cd; }
.rt-data-table tbody tr.modified { background: #fff9e6; }
.rt-data-table tbody tr.selected { background: #e3f2fd; }

.rt-select-col {
  width: 28px;
  text-align: center;
  padding: 4px 4px !important;
}

.rt-row-num-col {
  width: 40px;
  text-align: right;
  color: #6c757d;
  font-style: italic;
}

.rt-col-header {
  user-select: none;
}

.rt-col-name {
  display: flex;
  align-items: center;
  gap: 4px;
}

.rt-cell {
  cursor: cell;
  transition: background 0.1s;
}

.rt-cell:hover {
  background: #e9ecef;
}

.rt-null {
  color: #adb5bd;
  font-style: italic;
}

.rt-edit-input {
  width: 100%;
  background: #ffffff;
  border: 1px solid #007bff;
  border-radius: 2px;
  padding: 3px 6px;
  font-size: 12px;
  font-family: Consolas, Monaco, monospace;
  color: #212529;
  outline: none;
  box-sizing: border-box;
}

.rt-action-separator {
  width: 1px;
  height: 20px;
  background: #dee2e6;
  margin: 0 2px;
}

.rt-action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 22px;
  border: none;
  border-radius: 3px;
  background: transparent;
  cursor: pointer;
  transition: all 0.15s ease;
  font-size: 14px;
  font-weight: 600;
  color: #495057;
}

.rt-action-icon {
  line-height: 1;
}

.rt-action-btn:hover:not(:disabled) {
  background: #e3f2fd;
  color: #007bff;
}

.rt-action-btn:active {
  background: #bbdefb;
}

.rt-action-btn-add:hover:not(:disabled) {
  background: #e8f5e9;
  color: #28a745;
}

.rt-action-btn-delete:hover:not(:disabled) {
  background: #ffebee;
  color: #dc3545;
}

.rt-action-btn-submit {
  color: #28a745;
}

.rt-action-btn-submit:hover:not(:disabled) {
  background: #e8f5e9;
  color: #28a745;
}

.rt-action-btn-rollback {
  color: #dc3545;
}

.rt-action-btn-rollback:hover:not(:disabled) {
  background: #ffebee;
  color: #dc3545;
}

.rt-action-btn-stop:hover:not(:disabled) {
  background: #ffebee;
  color: #dc3545;
}

.rt-action-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  pointer-events: none;
}

.rt-export-menu {
  position: absolute;
  top: 30px;
  right: 0;
  background: #ffffff;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  z-index: 100;
  min-width: 140px;
}

.rt-export-item {
  padding: 6px 12px;
  font-size: 12px;
  color: #495057;
  cursor: pointer;
  white-space: nowrap;
}

.rt-export-item:hover {
  background: #e3f2fd;
  color: #007bff;
}

.rt-status-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 16px;
  background: linear-gradient(180deg, #f8f9fa 0%, #e9ecef 100%);
  border-top: 1px solid #dee2e6;
  font-size: 11.5px;
  color: #495057;
  min-height: 26px;
}

.rt-status-left {
  flex: 1;
  overflow: hidden;
}

.rt-status-sql {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  color: #212529;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 12px;
}

.rt-status-center {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

.rt-status-right {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.rt-status-separator {
  color: #adb5bd;
}

.rt-status-pagination, .rt-status-records, .rt-status-updated {
  color: #6c757d;
}

.rt-status-pagination {
  color: #28a745;
  font-weight: 500;
  font-size: 12px;
}

.rt-status-records {
  color: #007bff;
  font-size: 12px;
}

.rt-status-updated {
  color: #6c757d;
  font-style: italic;
  font-size: 12px;
}

.rt-page-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 20px;
  background: linear-gradient(180deg, #ffffff 0%, #f8f9fa 100%);
  border: 1px solid #dee2e6;
  border-radius: 2px;
  cursor: pointer;
  transition: all 0.15s;
}

.rt-page-btn:hover:not(:disabled) {
  background: #e3f2fd;
  border-color: #1976d2;
}

.rt-page-btn:active:not(:disabled) {
  background: #bbdefb;
}

.rt-page-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  background: #f8f9fa;
}

.rt-page-info {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #6c757d;
}

.rt-page-input {
  width: 45px;
  padding: 2px 4px;
  background: #ffffff;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  color: #212529;
  font-size: 12px;
  text-align: center;
  outline: none;
}

.rt-page-input:focus {
  border-color: #007bff;
}

.rt-page-separator {
  width: 1px;
  height: 16px;
  background: #dee2e6;
  margin: 0 4px;
}

.rt-page-size {
  padding: 2px 6px;
  background: #ffffff;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  color: #495057;
  font-size: 12px;
  outline: none;
}

.rt-page-label {
  font-size: 11px;
  color: #6c757d;
}

.rt-history-pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 8px 12px;
  background: #f8f9fa;
  border-top: 1px solid #e9ecef;
  font-size: 12px;
  color: #6c757d;
  flex-wrap: wrap;
}

.rt-history-pagination-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.rt-refresh-btn {
  padding: 4px 8px;
  font-size: 16px;
  color: #909399;
  cursor: pointer;
  border: none;
  background: transparent;
  border-radius: 4px;
  transition: all 0.15s;
}

.rt-refresh-btn:hover {
  color: #1976d2;
  background: #e3f2fd;
}

.rt-history-pagination-right {
  display: flex;
  align-items: center;
  gap: 6px;
}

.rt-page-info {
  font-size: 12px;
  color: #6c757d;
}

.rt-page-records {
  font-size: 12px;
  color: #6c757d;
}

.rt-page-updated {
  font-size: 11px;
  color: #6c757d;
}

.rt-message-pane {
  padding: 8px 12px;
  font-family: Consolas, Monaco, monospace;
}

.rt-message-item {
  padding: 6px 0;
  font-size: 12px;
  border-bottom: 1px dashed #e0e0e0;
}
.rt-message-item.INFO { color: #1976d2; }
.rt-message-item.ERROR { color: #c62828; }
.rt-message-item.WARN { color: #ef6c00; }

.rt-message-time { color: #78909c; margin-right: 8px; }
.rt-message-level { color: #546e7a; margin-right: 8px; }
.rt-message-content { color: #37474f; }

.rt-history-sql {
  max-width: 500px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.rt-tag-ok { background: #e8f5e9; color: #2e7d32; padding: 2px 6px; border-radius: 2px; font-size: 11px; }
.rt-tag-err { background: #ffebee; color: #c62828; padding: 2px 6px; border-radius: 2px; font-size: 11px; }

.rt-history-action {
  padding: 2px 8px;
  background: #1976d2;
  color: #ffffff;
  border: none;
  border-radius: 2px;
  font-size: 11px;
  cursor: pointer;
}
.rt-history-action:hover { background: #1565c0; }

.rt-ddl-text {
  padding: 16px;
  background: #263238;
  color: #eceff1;
  font-family: Consolas, Monaco, monospace;
  font-size: 12.5px;
  line-height: 1.6;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}

.rt-fav-sql {
  max-width: 400px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-family: Consolas, Monaco, monospace;
}

.rt-fav-action {
  padding: 2px 8px;
  background: #1976d2;
  color: #ffffff;
  border: none;
  border-radius: 2px;
  font-size: 11px;
  cursor: pointer;
  margin-right: 6px;
}

.rt-fav-action:hover {
  background: #1565c0;
}

.rt-fav-action-danger {
  background: #c62828;
}

.rt-fav-action-danger:hover {
  background: #b71c1c;
}

.rt-shell-wrap {
  max-height: 400px;
  overflow: auto;
  background: #1e1e1e;
}

.rt-shell-output {
  padding: 12px 16px;
  color: #d4d4d4;
  font-family: Consolas, Monaco, 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.4;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}

.rt-explain-item {
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  margin-bottom: 8px;
  overflow: hidden;
}

.rt-explain-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 12px;
  background: #f5f7fa;
  font-size: 12.5px;
}

.rt-explain-title { font-weight: 500; color: #303133; }
.rt-explain-dura { color: #7e57c2; font-size: 11.5px; }
.rt-explain-status {
  padding: 2px 8px;
  border-radius: 2px;
  font-size: 11px;
  font-weight: 500;
}
.rt-explain-status.ok { background: #e8f5e9; color: #2e7d32; }
.rt-explain-status.err { background: #ffebee; color: #c62828; }

.rt-explain-sql {
  padding: 8px 12px;
  background: #fafafa;
  border-top: 1px solid #e8e8e8;
  border-bottom: 1px solid #e8e8e8;
  font-family: Consolas, Monaco, monospace;
  font-size: 12px;
  color: #52c41a;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>