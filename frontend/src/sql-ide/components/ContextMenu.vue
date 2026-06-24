﻿﻿﻿<template>
  <Teleport to="body">
    <div v-if="ctxMenu.show" class="ctx-menu-overlay" @click="ctxMenu.show = false" @contextmenu.prevent="ctxMenu.show = false">
      <div class="ctx-menu" :style="{ left: ctxMenu.x + 'px', top: ctxMenu.y + 'px' }" @click.stop>
        <template v-if="ctxMenu.datasource">
          <div class="ctx-menu-item" @click="handleClick('ds-connect')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><ellipse cx="8" cy="3" rx="5" ry="2" stroke="#2d7cf0" fill="#4a90e2"/><path d="M3 5v3c0 1.1 2.2 2 5 2s5-0.9 5-2V5" stroke="#2d7cf0" fill="none"/><ellipse cx="8" cy="8" rx="5" ry="2" stroke="#2d7cf0" fill="#4a90e2" opacity="0.7"/></svg>
            </span>
            <span>{{ ctxMenu.datasource.connected ? '刷新连接' : '打开连接' }}</span>
          </div>
          <div class="ctx-menu-item" @click="handleClick('ds-query')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="2" y="3" width="12" height="10" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.5"/><path d="M4 6 h8 M4 9 h6" stroke="#424242" stroke-width="0.8"/></svg>
            </span>
            <span>新建查询</span>
          </div>
          <div class="ctx-menu-sep"></div>
          <div class="ctx-menu-item" @click="handleClick('ds-refresh')">
            <span class="ctx-icon ctx-icon-refresh">⟳</span>
            <span>刷新对象树</span>
          </div>
          <template v-if="ctxMenu.datasource.connected">
            <div class="ctx-menu-item" @click="handleClick('ds-close')">
              <span class="ctx-icon">
                <svg viewBox="0 0 16 16" width="14" height="14"><path d="M4 4 L12 12 M12 4 L4 12" stroke="#c62828" stroke-width="1.2" stroke-linecap="round"/></svg>
              </span>
              <span>关闭连接</span>
            </div>
          </template>
          <div class="ctx-menu-sep"></div>
          <div class="ctx-menu-item" @click="handleClick('ds-copy-name')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="3.5" y="3.5" width="6" height="6" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.5"/><rect x="6" y="6" width="6" height="6" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.8"/></svg>
            </span>
            <span>复制连接名</span>
          </div>
        </template>

        <template v-else-if="ctxMenu.node && (ctxMenu.node as any).type === 'database'">
          <div class="ctx-menu-item" @click="handleClick('db-query')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="2" y="3" width="12" height="10" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.5"/><path d="M4 6 h8 M4 9 h6" stroke="#424242" stroke-width="0.8"/></svg>
            </span>
            <span>新建查询</span>
          </div>
          <div class="ctx-menu-item" @click="handleClick('set-db')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><ellipse cx="8" cy="3" rx="5" ry="1.3" fill="#3f51b5"/><path d="M3 4 v4 a5 1.3 0 0 0 10 0 v-4" fill="#5c6bc0" stroke="#1a237e" stroke-width="0.5"/><ellipse cx="8" cy="10" rx="5" ry="1.3" fill="#3f51b5"/></svg>
            </span>
            <span>设为当前数据库</span>
          </div>
          <div class="ctx-menu-sep"></div>
          <div class="ctx-menu-item" @click="handleClick('db-refresh')">
            <span class="ctx-icon ctx-icon-refresh">⟳</span>
            <span>刷新</span>
          </div>
          <div class="ctx-menu-item" @click="handleClick('copy-name')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="3.5" y="3.5" width="6" height="6" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.5"/><rect x="6" y="6" width="6" height="6" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.8"/></svg>
            </span>
            <span>复制数据库名</span>
          </div>
        </template>

        <template v-else-if="ctxMenu.node && ((ctxMenu.node as any).type === 'table' || (ctxMenu.node as any).type === 'view')">
          <div class="ctx-menu-item" @click="handleClick('open-table')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="2" y="3" width="12" height="10" rx="1" fill="#ffffff" stroke="#2e7d32" stroke-width="0.8"/><path d="M2 6.5 h12 M2 9.5 h12" stroke="#a5d6a7" stroke-width="0.6"/><path d="M5.5 3 v10 M9.5 3 v10" stroke="#a5d6a7" stroke-width="0.6"/></svg>
            </span>
            <span>打开表</span>
          </div>
          <div class="ctx-menu-item" @click="handleClick('select-top')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><polygon points="3,2 14,8 3,14" fill="#2a9d3b"/></svg>
            </span>
            <span>查询前 200 行</span>
          </div>
          <div class="ctx-menu-item" @click="handleClick('select-top-1000')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><polygon points="3,2 14,8 3,14" fill="#2a9d3b"/></svg>
            </span>
            <span>查询前 1000 行</span>
          </div>
          <div class="ctx-menu-item" @click="handleClick('count')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="3" y="3" width="10" height="10" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.8"/></svg>
            </span>
            <span>统计行数</span>
          </div>
          <div class="ctx-menu-sep"></div>
          <div class="ctx-menu-item" @click="handleClick('ddl')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="3" y="2.5" width="10" height="11" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.8"/><path d="M5 5.5 h6 M5 8 h5 M5 10.5 h4" stroke="#424242" stroke-width="0.8"/></svg>
            </span>
            <span>查看 DDL / 建表语句</span>
          </div>
          <div class="ctx-menu-item" @click="handleClick('select-all')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="2" y="2" width="12" height="12" rx="1" fill="#ffffff" stroke="#5c6bc0" stroke-width="0.8"/></svg>
            </span>
            <span>SELECT * FROM</span>
          </div>
          <div class="ctx-menu-item" @click="handleClick('insert-template')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="2" y="2" width="12" height="12" rx="1" fill="#ffffff" stroke="#7e57c2" stroke-width="0.8"/><path d="M7 4 v8 M3 8 h8" stroke="#7e57c2" stroke-width="1.2"/></svg>
            </span>
            <span>生成 INSERT 模板</span>
          </div>
          <div class="ctx-menu-item" @click="handleClick('update-template')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="2" y="2" width="12" height="12" rx="1" fill="#ffffff" stroke="#ff9800" stroke-width="0.8"/><path d="M5 8 h6" stroke="#ff9800" stroke-width="1.2"/></svg>
            </span>
            <span>生成 UPDATE 模板</span>
          </div>
          <div class="ctx-menu-item" @click="handleClick('delete-template')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="2" y="2" width="12" height="12" rx="1" fill="#ffffff" stroke="#f44336" stroke-width="0.8"/><path d="M4 4 L12 12 M12 4 L4 12" stroke="#f44336" stroke-width="1.2" stroke-linecap="round"/></svg>
            </span>
            <span>生成 DELETE 模板</span>
          </div>
          <div class="ctx-menu-sep"></div>
          <div class="ctx-menu-item" @click="handleClick('copy-qualified')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="3.5" y="3.5" width="6" height="6" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.5"/><rect x="6" y="6" width="6" height="6" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.8"/></svg>
            </span>
            <span>复制完整名 (db.table)</span>
          </div>
          <div class="ctx-menu-item" @click="handleClick('copy-name')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="4" y="4" width="8" height="8" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.8"/></svg>
            </span>
            <span>复制名称</span>
          </div>
          <div class="ctx-menu-sep"></div>
          <div class="ctx-menu-item" @click="handleClick('refresh')">
            <span class="ctx-icon ctx-icon-refresh">⟳</span>
            <span>刷新</span>
          </div>
        </template>

        <template v-else-if="ctxMenu.node && ((ctxMenu.node as any).type === 'function' || (ctxMenu.node as any).type === 'procedure')">
          <div class="ctx-menu-item" @click="handleClick('ddl')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="3" y="2.5" width="10" height="11" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.8"/><path d="M5 5.5 h6 M5 8 h5 M5 10.5 h4" stroke="#424242" stroke-width="0.8"/></svg>
            </span>
            <span>查看 DDL / 创建语句</span>
          </div>
          <div class="ctx-menu-item" @click="handleClick('select-top')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><polygon points="3,2 14,8 3,14" fill="#2a9d3b"/></svg>
            </span>
            <span>在编辑器中调用 / 测试</span>
          </div>
          <div class="ctx-menu-sep"></div>
          <div class="ctx-menu-item" @click="handleClick('copy-qualified')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="3.5" y="3.5" width="6" height="6" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.5"/><rect x="6" y="6" width="6" height="6" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.8"/></svg>
            </span>
            <span>复制完整名 (db.name)</span>
          </div>
          <div class="ctx-menu-item" @click="handleClick('copy-name')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="4" y="4" width="8" height="8" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.8"/></svg>
            </span>
            <span>复制名称</span>
          </div>
        </template>

        <template v-else-if="ctxMenu.node && (ctxMenu.node as any).type === 'trigger'">
          <div class="ctx-menu-item" @click="handleClick('ddl')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="3" y="2.5" width="10" height="11" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.8"/><path d="M5 5.5 h6 M5 8 h5 M5 10.5 h4" stroke="#424242" stroke-width="0.8"/></svg>
            </span>
            <span>查看 DDL / 创建语句</span>
          </div>
          <div class="ctx-menu-sep"></div>
          <div class="ctx-menu-item" @click="handleClick('copy-qualified')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="3.5" y="3.5" width="6" height="6" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.5"/><rect x="6" y="6" width="6" height="6" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.8"/></svg>
            </span>
            <span>复制完整名 (db.trigger)</span>
          </div>
          <div class="ctx-menu-item" @click="handleClick('copy-name')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="4" y="4" width="8" height="8" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.8"/></svg>
            </span>
            <span>复制名称</span>
          </div>
        </template>

        <template v-else-if="ctxMenu.node && (ctxMenu.node as any).type === 'event'">
          <div class="ctx-menu-item" @click="handleClick('ddl')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="3" y="2.5" width="10" height="11" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.8"/><path d="M5 5.5 h6 M5 8 h5 M5 10.5 h4" stroke="#424242" stroke-width="0.8"/></svg>
            </span>
            <span>查看 DDL / 创建语句</span>
          </div>
          <div class="ctx-menu-sep"></div>
          <div class="ctx-menu-item" @click="handleClick('copy-qualified')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="3.5" y="3.5" width="6" height="6" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.5"/><rect x="6" y="6" width="6" height="6" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.8"/></svg>
            </span>
            <span>复制完整名 (db.event)</span>
          </div>
          <div class="ctx-menu-item" @click="handleClick('copy-name')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="4" y="4" width="8" height="8" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.8"/></svg>
            </span>
            <span>复制名称</span>
          </div>
        </template>

        <template v-else-if="ctxMenu.node && (ctxMenu.node as any).type === 'column'">
          <div class="ctx-menu-item" @click="handleClick('insert-col')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="2" y="2" width="12" height="12" rx="1" fill="#ffffff" stroke="#7e57c2" stroke-width="0.8"/><path d="M7 4 v8 M3 8 h8" stroke="#7e57c2" stroke-width="1.2"/></svg>
            </span>
            <span>插入列名到编辑器</span>
          </div>
          <div class="ctx-menu-item" @click="handleClick('copy-name')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="4" y="4" width="8" height="8" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.8"/></svg>
            </span>
            <span>复制列名</span>
          </div>
        </template>

        <template v-else-if="ctxMenu.node && (ctxMenu.node as any).type === 'query-save'">
          <div class="ctx-menu-item" @click="handleClick('refresh')">
            <span class="ctx-icon ctx-icon-refresh">⟳</span>
            <span>刷新</span>
          </div>
        </template>

        <template v-else-if="ctxMenu.node && (ctxMenu.node as any).type === 'group'">
          <div class="ctx-menu-item" @click="handleClick('db-query')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="2" y="3" width="12" height="10" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.5"/><path d="M4 6 h8 M4 9 h6" stroke="#424242" stroke-width="0.8"/></svg>
            </span>
            <span>新建查询</span>
          </div>
          <div class="ctx-menu-item" @click="handleClick('refresh')">
            <span class="ctx-icon ctx-icon-refresh">⟳</span>
            <span>刷新</span>
          </div>
          <div class="ctx-menu-item" @click="handleClick('copy-name')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="4" y="4" width="8" height="8" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.8"/></svg>
            </span>
            <span>复制名称</span>
          </div>
        </template>

        <template v-else>
          <div v-if="ctxMenu.node" class="ctx-menu-item" @click="handleClick('copy-name')">
            <span class="ctx-icon">
              <svg viewBox="0 0 16 16" width="14" height="14"><rect x="4" y="4" width="8" height="8" rx="1" fill="#ffffff" stroke="#757575" stroke-width="0.8"/></svg>
            </span>
            <span>复制名称</span>
          </div>
        </template>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { useSqlIdeContextMenu } from '../hooks/useContextMenu'
const { ctxMenu, handleClick } = useSqlIdeContextMenu()
</script>

<style scoped>
.ctx-menu-overlay { position: fixed; inset: 0; z-index: 3000; background: transparent; }
.ctx-menu {
  position: absolute;
  min-width: 210px;
  background: #ffffff;
  border: 1px solid #d0d0d0;
  border-radius: 4px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  padding: 4px 0;
  user-select: none;
}
.ctx-menu-item {
  display: flex; align-items: center; gap: 8px;
  padding: 6px 14px; cursor: pointer;
  font-size: 12.5px; color: #303133;
  transition: background 0.1s;
}
.ctx-menu-item:hover { background: #1976d2; color: #ffffff; }
.ctx-icon {
  width: 16px; height: 16px; flex-shrink: 0;
  display: inline-flex; align-items: center; justify-content: center;
}
.ctx-menu-sep { height: 1px; background: #e0e0e0; margin: 4px 0; }
</style>
