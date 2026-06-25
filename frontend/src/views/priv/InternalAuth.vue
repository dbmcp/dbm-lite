<template>
  <div class="page-container">
    <div class="page-header">
      <div class="title-row">
        <el-icon :size="22" color="#409EFF"><Lock /></el-icon>
        <span class="title-text">实例权限管控</span>
      </div>
      <div class="header-actions">
        <el-select
          v-model="selectedDsId"
          placeholder="请选择数据源"
          style="width:260px;"
          filterable
          @change="handleDatasourceChange"
          size="default"
        >
          <el-option
            v-for="ds in datasourceList"
            :key="ds.datasourceId"
            :label="`${ds.name} (${ds.dbType})`"
            :value="ds.datasourceId"
          />
        </el-select>
        <el-button
          v-if="selectedDs"
          type="success"
          size="default"
          @click="syncUsers"
          style="margin-left: 8px;"
        >
          同步用户
        </el-button>
        <el-button @click="loadDatasources" size="default" style="margin-left: 8px;"><span class="refresh-icon">⟳</span>刷新</el-button>
      </div>
    </div>

    <div class="tabs-wrapper">
      <el-tabs v-model="activeTab" class="main-tabs">
        <el-tab-pane label="内部用户" name="users">
        <div class="tab-content">
          <div class="toolbar">
            <el-input v-model="userKeyword" placeholder="用户名" clearable style="width:180px;" size="small">
              <template #prefix><el-icon><Search /></el-icon></template>
            </el-input>
            <el-select v-model="userStatus" placeholder="状态" clearable style="width:120px;" size="small">
              <el-option label="启用" value="active" />
              <el-option label="禁用" value="inactive" />
            </el-select>
            <el-button type="primary" @click="loadUsers" size="small">查询</el-button>
            <el-button @click="showUserDialog = true" size="small">新建用户</el-button>
            <div class="toolbar-right">
              <el-dropdown @command="toggleUserColumn" trigger="click">
                <el-button size="small"><el-icon><Menu /></el-icon>列选择</el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item :command="'username'">
                      <span :class="{ checked: userColumnsVisible.username }">{{ userColumnsVisible.username ? '✓' : '○' }} 用户名</span>
                    </el-dropdown-item>
                    <el-dropdown-item :command="'host'" v-if="selectedDs?.dbType !== 'sqlite'">
                      <span :class="{ checked: userColumnsVisible.host }">{{ userColumnsVisible.host ? '✓' : '○' }} 主机</span>
                    </el-dropdown-item>
                    <el-dropdown-item :command="'status'">
                      <span :class="{ checked: userColumnsVisible.status }">{{ userColumnsVisible.status ? '✓' : '○' }} 状态</span>
                    </el-dropdown-item>
                    <el-dropdown-item :command="'isBuiltIn'">
                      <span :class="{ checked: userColumnsVisible.isBuiltIn }">{{ userColumnsVisible.isBuiltIn ? '✓' : '○' }} 内置</span>
                    </el-dropdown-item>
                    <el-dropdown-item :command="'permissions'">
                      <span :class="{ checked: userColumnsVisible.permissions }">{{ userColumnsVisible.permissions ? '✓' : '○' }} 权限</span>
                    </el-dropdown-item>
                    <el-dropdown-item :command="'createdAt'">
                      <span :class="{ checked: userColumnsVisible.createdAt }">{{ userColumnsVisible.createdAt ? '✓' : '○' }} 创建时间</span>
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>
          <AuthTable
            :data="userList"
            :columns="userTableColumns"
            title="用户列表"
            :total="userTotal"
            :current-page="userPage"
            :page-size="userPageSize"
            :show-checkbox="true"
            :status-text="`当前数据源: ${selectedDs?.name || '-'}`"
            :cell-renderer="renderUserCell"
            @page-change="userPage = $event; loadUsers()"
            @page-size-change="userPageSize = $event; loadUsers()"
            @selection-change="handleUserSelectionChange"
            @cell-click="handleUserCellClick"
          />
        </div>
      </el-tab-pane>

      <el-tab-pane label="角色管理" name="roles">
        <div class="tab-content">
          <div class="toolbar">
            <el-input v-model="roleKeyword" placeholder="角色名称" clearable style="width:180px;" size="small">
              <template #prefix><el-icon><Search /></el-icon></template>
            </el-input>
            <el-button type="primary" @click="loadRoles" size="small">查询</el-button>
            <el-button @click="showRoleDialog = true" size="small">新建角色</el-button>
            <div class="toolbar-right">
              <el-dropdown @command="toggleRoleColumn" trigger="click">
                <el-button size="small"><el-icon><Menu /></el-icon>列选择</el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item :command="'name'">
                      <span :class="{ checked: roleColumnsVisible.name }">{{ roleColumnsVisible.name ? '✓' : '○' }} 角色名称</span>
                    </el-dropdown-item>
                    <el-dropdown-item :command="'description'">
                      <span :class="{ checked: roleColumnsVisible.description }">{{ roleColumnsVisible.description ? '✓' : '○' }} 描述</span>
                    </el-dropdown-item>
                    <el-dropdown-item :command="'userCount'">
                      <span :class="{ checked: roleColumnsVisible.userCount }">{{ roleColumnsVisible.userCount ? '✓' : '○' }} 关联用户数</span>
                    </el-dropdown-item>
                    <el-dropdown-item :command="'permissions'">
                      <span :class="{ checked: roleColumnsVisible.permissions }">{{ roleColumnsVisible.permissions ? '✓' : '○' }} 权限</span>
                    </el-dropdown-item>
                    <el-dropdown-item :command="'createdAt'">
                      <span :class="{ checked: roleColumnsVisible.createdAt }">{{ roleColumnsVisible.createdAt ? '✓' : '○' }} 创建时间</span>
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>
          <AuthTable
            :data="roleList"
            :columns="roleTableColumns"
            title="角色列表"
            :total="roleTotal"
            :current-page="rolePage"
            :page-size="rolePageSize"
            :show-checkbox="false"
            :status-text="`当前数据源: ${selectedDs?.name || '-'}`"
            :cell-renderer="renderRoleCell"
            @page-change="rolePage = $event; loadRoles()"
            @page-size-change="rolePageSize = $event; loadRoles()"
            @cell-click="handleRoleCellClick"
          />
        </div>
      </el-tab-pane>

      <el-tab-pane label="权限规则" name="permissions">
        <div class="tab-content">
          <div class="toolbar">
            <el-select v-model="permPrincipalType" placeholder="授权主体" clearable style="width:120px;" size="small">
              <el-option label="用户" value="user" />
              <el-option label="角色" value="role" />
            </el-select>
            <el-select v-model="permType" placeholder="权限类型" clearable style="width:120px;" size="small">
              <el-option label="只读" value="readonly" />
              <el-option label="DML" value="dml" />
              <el-option label="DDL" value="ddl" />
            </el-select>
            <el-select v-model="permLevel" placeholder="对象层级" clearable style="width:120px;" size="small">
              <el-option label="数据库" value="database" />
              <el-option label="表" value="table" />
              <el-option label="列" value="column" />
              <el-option label="视图" value="view" />
              <el-option label="触发器" value="trigger" />
            </el-select>
            <el-button type="primary" @click="loadPermissions" size="small">查询</el-button>
            <el-button @click="showGrantDialog = true" size="small">授权</el-button>
            <div class="toolbar-right">
              <el-dropdown @command="togglePermColumn" trigger="click">
                <el-button size="small"><el-icon><Menu /></el-icon>列选择</el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item :command="'principalName'">
                      <span :class="{ checked: permColumnsVisible.principalName }">{{ permColumnsVisible.principalName ? '✓' : '○' }} 授权主体</span>
                    </el-dropdown-item>
                    <el-dropdown-item :command="'principalType'">
                      <span :class="{ checked: permColumnsVisible.principalType }">{{ permColumnsVisible.principalType ? '✓' : '○' }} 主体类型</span>
                    </el-dropdown-item>
                    <el-dropdown-item :command="'privilegeType'">
                      <span :class="{ checked: permColumnsVisible.privilegeType }">{{ permColumnsVisible.privilegeType ? '✓' : '○' }} 权限类型</span>
                    </el-dropdown-item>
                    <el-dropdown-item :command="'objectLevel'">
                      <span :class="{ checked: permColumnsVisible.objectLevel }">{{ permColumnsVisible.objectLevel ? '✓' : '○' }} 对象层级</span>
                    </el-dropdown-item>
                    <el-dropdown-item :command="'scope'">
                      <span :class="{ checked: permColumnsVisible.scope }">{{ permColumnsVisible.scope ? '✓' : '○' }} 授权范围</span>
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>
          <div class="perm-footer">
            <el-button type="danger" @click="batchRevokePerm" :disabled="selectedPerms.length === 0" size="small">
              批量回收
            </el-button>
          </div>
          <AuthTable
            :data="permList"
            :columns="permTableColumns"
            title="权限规则"
            :total="permTotal"
            :current-page="permPage"
            :page-size="permPageSize"
            :show-checkbox="true"
            :status-text="`当前数据源: ${selectedDs?.name || '-'}`"
            :cell-renderer="renderPermCell"
            @page-change="permPage = $event; loadPermissions()"
            @page-size-change="permPageSize = $event; loadPermissions()"
            @selection-change="handlePermSelectionChange"
            @cell-click="handlePermCellClick"
          />
        </div>
      </el-tab-pane>
    </el-tabs>
    </div>

    <el-dialog v-model="showUserDialog" :title="editingUser ? '编辑用户' : '新建用户'" width="480px" class="compact-dialog">
      <el-form :model="userForm" :rules="userRules" ref="userFormRef" label-width="80px" :inline="true">
        <template v-if="!editingUser">
          <el-form-item label="用户名" prop="username">
            <el-input v-model="userForm.username" style="width:200px;" />
          </el-form-item>
          <el-form-item label="访问主机" prop="host" v-if="selectedDs?.dbType !== 'sqlite'">
            <el-input v-model="userForm.host" placeholder="默认 %" style="width:200px;" />
          </el-form-item>
          <el-form-item label="密码" prop="password">
            <el-input v-model="userForm.password" type="password" placeholder="不填则自动生成" style="width:200px;" />
          </el-form-item>
        </template>
        <template v-else>
          <el-form-item label="用户名">
            <el-tag size="small">{{ userForm.username }}</el-tag>
          </el-form-item>
          <el-form-item label="访问主机" v-if="selectedDs?.dbType !== 'sqlite'">
            <el-tag size="small">{{ userForm.host || '%' }}</el-tag>
          </el-form-item>
        </template>
        <el-form-item label="备注" prop="remark" style="width:100%;">
          <el-input v-model="userForm.remark" type="textarea" :rows="2" style="width:100%;" placeholder="请输入描述信息" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showUserDialog = false" size="small">取消</el-button>
        <el-button type="primary" @click="saveUser" size="small">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showRoleDialog" :title="editingRole ? '编辑角色' : '新建角色'" width="700px" class="compact-dialog">
      <el-tabs v-model="editRoleActiveTab" size="small">
        <el-tab-pane label="基本信息" name="basic">
          <el-form :model="roleForm" :rules="roleRules" ref="roleFormRef" label-width="80px" :inline="true">
            <el-form-item label="角色名称" prop="name">
              <el-input v-model="roleForm.name" style="width:200px;" />
            </el-form-item>
            <el-form-item label="描述" prop="description" style="width:100%;">
              <el-input v-model="roleForm.description" type="textarea" :rows="2" style="width:100%;" />
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="权限管理" name="permissions" v-if="editingRole">
          <div class="permission-panel">
            <div class="permission-header">
              <span class="permission-title">权限列表</span>
              <el-button size="small" type="primary" @click="openRoleGrantSubDialog">+ 添加权限</el-button>
            </div>
            <div class="permission-list">
              <div v-if="editingRolePermissions.length === 0" class="permission-empty">
                暂无权限，点击上方按钮添加
              </div>
              <div v-else class="permission-items">
                <div v-for="perm in editingRolePermissions" :key="perm.id" class="permission-item">
                  <div class="perm-info">
                    <el-tag :type="getPrivTagType(perm.privilegeType)" size="small">{{ getPrivLabel(perm.privilegeType) }}</el-tag>
                    <span class="perm-scope">{{ perm.databaseName }}{{ perm.table ? '.' + perm.table : '' }}</span>
                    <span class="perm-level">{{ getLevelLabel(perm.objectLevel) }}</span>
                  </div>
                  <el-button size="small" type="danger" @click="revokePermissionDirect(perm)">回收</el-button>
                </div>
              </div>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="showRoleDialog = false" size="small">取消</el-button>
        <el-button type="primary" @click="saveRole" size="small">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showGrantSubDialog" title="添加权限" width="480px" class="compact-dialog">
      <el-form :model="grantSubForm" label-width="80px" :inline="true">
        <el-form-item label="权限类型">
          <el-select v-model="grantSubForm.privilegeType" style="width:100%" size="small">
            <el-option label="只读 (SELECT)" value="readonly" />
            <el-option label="DML (SELECT/INSERT/UPDATE/DELETE)" value="dml" />
            <el-option label="DDL (全部权限)" value="ddl" />
          </el-select>
        </el-form-item>
        <el-form-item label="对象层级">
          <el-select v-model="grantSubForm.objectLevel" style="width:100%" @change="onGrantSubObjectLevelChange" size="small">
            <el-option label="数据库" value="database" />
            <el-option label="表" value="table" />
            <el-option label="列" value="column" />
          </el-select>
        </el-form-item>
        <el-form-item label="数据库">
          <el-select v-model="grantSubForm.databaseName" style="width:100%" filterable @change="onGrantSubDatabaseChange" size="small">
            <el-option v-for="db in editingPermDatabases" :key="db" :label="db" :value="db" />
          </el-select>
        </el-form-item>
        <el-form-item label="表" v-if="grantSubForm.objectLevel !== 'database'">
          <el-select v-model="grantSubForm.tableName" style="width:100%" filterable @change="onGrantSubTableChange" size="small">
            <el-option v-for="t in editingPermTables" :key="t" :label="t" :value="t" />
          </el-select>
        </el-form-item>
        <el-form-item label="列" v-if="grantSubForm.objectLevel === 'column'">
          <el-select v-model="grantSubForm.columns" multiple style="width:100%" filterable size="small">
            <el-option v-for="c in editingPermColumns" :key="c" :label="c" :value="c" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showGrantSubDialog = false" size="small">取消</el-button>
        <el-button type="primary" @click="submitGrantSub" size="small">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showGrantDialog" title="授权" width="560px" class="compact-dialog">
      <el-form :model="grantForm" label-width="80px" :inline="true">
        <el-form-item label="授权主体">
          <el-radio-group v-model="grantForm.principalType">
            <el-radio value="user">用户</el-radio>
            <el-radio value="role">角色</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="选择主体">
          <el-select v-model="grantForm.principalId" style="width:100%" filterable size="small">
            <el-option v-for="item in principalOptions" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="权限类型">
          <el-select v-model="grantForm.privilegeType" style="width:100%" size="small">
            <el-option label="只读 (SELECT)" value="readonly" />
            <el-option label="DML (SELECT/INSERT/UPDATE/DELETE)" value="dml" />
            <el-option label="DDL (全部权限)" value="ddl" />
          </el-select>
        </el-form-item>
        <el-form-item label="对象层级">
          <el-select v-model="grantForm.objectLevel" style="width:100%" @change="onObjectLevelChange" size="small">
            <el-option label="数据库" value="database" />
            <el-option label="表" value="table" />
            <el-option label="列" value="column" />
            <el-option label="视图" value="view" />
            <el-option label="触发器" value="trigger" />
          </el-select>
        </el-form-item>
        <el-form-item label="数据库">
          <el-select v-model="grantForm.databaseName" style="width:100%" filterable @change="onDatabaseChange" size="small">
            <el-option v-for="db in databases" :key="db" :label="db" :value="db" />
          </el-select>
        </el-form-item>
        <el-form-item label="表" v-if="grantForm.objectLevel !== 'database'">
          <el-select v-model="grantForm.tableName" style="width:100%" filterable @change="onTableChange" size="small">
            <el-option v-for="t in tables" :key="t" :label="t" :value="t" />
          </el-select>
        </el-form-item>
        <el-form-item label="列" v-if="grantForm.objectLevel === 'column'">
          <el-select v-model="grantForm.columns" multiple style="width:100%" filterable size="small">
            <el-option v-for="c in columns" :key="c" :label="c" :value="c" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showGrantDialog = false" size="small">取消</el-button>
        <el-button type="primary" @click="submitGrant" size="small">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showPermissionsDialog" :title="permissionsDialogTitle" width="700px" class="compact-dialog">
      <div v-if="permissionsLoading" class="permissions-loading">
        <el-spinner />
      </div>
      <div v-else-if="grantsOutput.length === 0" class="permissions-empty">
        该{{ permissionsDialogType === 'user' ? '用户' : '角色' }}暂无权限
      </div>
      <div v-else class="grants-output">
        <pre class="sql-code">{{ grantsOutput }}</pre>
      </div>
      <template #footer>
        <el-button @click="showPermissionsDialog = false" size="small">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Lock, Menu } from '@element-plus/icons-vue'
import * as internalAuthApi from '@/api/internalAuth'
import * as datasourceApi from '@/api/datasource'
import { listColumns } from '@/api/sql'
import AuthTable from './components/AuthTable.vue'

const activeTab = ref('users')
const selectedDsId = ref('')
const datasourceList = ref<any[]>([])
const selectedDs = computed(() => datasourceList.value.find(ds => ds.datasourceId === selectedDsId.value))

const userTableColumns = computed(() => {
  const cols: any[] = [
    { key: 'username', label: '用户名', width: '120px' },
    { key: 'host', label: '主机', width: '100px' },
    { key: 'status', label: '状态', width: '80px' },
    { key: 'isBuiltIn', label: '内置', width: '60px' },
    { key: 'permissions', label: '权限', width: '80px' },
    { key: 'createdAt', label: '创建时间', width: '160px' },
    { key: 'actions', label: '操作', width: '280px' }
  ]
  return cols.filter(c => userColumnsVisible.value[c.key] !== false || c.key === 'actions')
})

const roleTableColumns = computed(() => {
  const cols: any[] = [
    { key: 'name', label: '角色名称', width: '150px' },
    { key: 'description', label: '描述', width: '200px' },
    { key: 'userCount', label: '关联用户数', width: '100px' },
    { key: 'permissions', label: '权限', width: '80px' },
    { key: 'createdAt', label: '创建时间', width: '160px' },
    { key: 'actions', label: '操作', width: '120px' }
  ]
  return cols.filter(c => roleColumnsVisible.value[c.key] !== false || c.key === 'actions')
})

const permTableColumns = computed(() => {
  const cols: any[] = [
    { key: 'principalName', label: '授权主体', width: '120px' },
    { key: 'principalType', label: '主体类型', width: '80px' },
    { key: 'privilegeType', label: '权限类型', width: '80px' },
    { key: 'objectLevel', label: '对象层级', width: '80px' },
    { key: 'scope', label: '授权范围', width: '200px' },
    { key: 'actions', label: '操作', width: '80px' }
  ]
  return cols.filter(c => permColumnsVisible.value[c.key] !== false || c.key === 'actions')
})

const userKeyword = ref('')
const userStatus = ref('')
const userList = ref<any[]>([])
const userLoading = ref(false)
const userPage = ref(1)
const userPageSize = ref(50)
const userTotal = ref(0)
const selectedUsers = ref<any[]>([])

const userColumnsVisible = ref({
  username: true,
  host: true,
  status: true,
  isBuiltIn: true,
  permissions: true,
  createdAt: true
})

const userColumnWidths = ref<Record<string, string>>({})

const roleColumnsVisible = ref({
  name: true,
  description: true,
  userCount: true,
  permissions: true,
  createdAt: true
})

const roleColumnWidths = ref<Record<string, string>>({})

const permColumnsVisible = ref({
  principalName: true,
  principalType: true,
  privilegeType: true,
  objectLevel: true,
  scope: true
})

const permColumnWidths = ref<Record<string, string>>({})

const resizingColumn = ref<string | null>(null)
const resizeStartX = ref(0)
const resizeStartWidth = ref(0)

const roleKeyword = ref('')
const roleList = ref<any[]>([])
const roleLoading = ref(false)
const rolePage = ref(1)
const rolePageSize = ref(50)
const roleTotal = ref(0)

const permPrincipalType = ref('')
const permType = ref('')
const permLevel = ref('')
const permList = ref<any[]>([])
const permLoading = ref(false)
const permPage = ref(1)
const permPageSize = ref(50)
const permTotal = ref(0)
const selectedPerms = ref<any[]>([])

const showUserDialog = ref(false)
const editingUser = ref<any>(null)
const userForm = ref({
  username: '',
  host: '',
  password: '',
  remark: ''
})
const userFormRef = ref<any>(null)
const userRules = {
  username: [{ required: true, message: '用户名必填', trigger: 'blur' }]
}

const showRoleDialog = ref(false)
const editingRole = ref<any>(null)
const editRoleActiveTab = ref('basic')
const editingRolePermissions = ref<any[]>([])
const roleForm = ref({
  name: '',
  description: ''
})

const showGrantSubDialog = ref(false)
const grantSubForm = ref({
  privilegeType: 'readonly',
  objectLevel: 'database',
  databaseName: '',
  tableName: '',
  columns: [] as string[]
})
const editingPermDatabases = ref<string[]>([])
const editingPermTables = ref<string[]>([])
const editingPermColumns = ref<string[]>([])
const grantSubPrincipalType = ref('user')
const grantSubPrincipalId = ref('')
const roleFormRef = ref<any>(null)
const roleRules = {
  name: [{ required: true, message: '角色名称必填', trigger: 'blur' }]
}

const showGrantDialog = ref(false)
const grantForm = ref({
  principalType: 'user',
  principalId: '',
  privilegeType: 'readonly',
  objectLevel: 'database',
  databaseName: '',
  tableName: '',
  columns: [] as string[]
})

const showPermissionsDialog = ref(false)
const permissionsDialogTitle = ref('')
const permissionsDialogType = ref('user')
const permissionsLoading = ref(false)
const grantsOutput = ref('')

const databases = ref<string[]>([])
const tables = ref<string[]>([])
const columns = ref<string[]>([])

const principalOptions = computed(() => {
  if (grantForm.value.principalType === 'user') {
    return userList.value.map(u => ({ id: u.id, name: u.username }))
  }
  return roleList.value.map(r => ({ id: r.id, name: r.name }))
})

function loadSavedColumnSettings() {
  try {
    const saved = localStorage.getItem('internalAuthColumnSettings')
    if (saved) {
      const settings = JSON.parse(saved)
      if (settings.userColumnsVisible) {
        userColumnsVisible.value = { ...userColumnsVisible.value, ...settings.userColumnsVisible }
      }
      if (settings.userColumnWidths) {
        userColumnWidths.value = settings.userColumnWidths
      }
      if (settings.roleColumnsVisible) {
        roleColumnsVisible.value = { ...roleColumnsVisible.value, ...settings.roleColumnsVisible }
      }
      if (settings.roleColumnWidths) {
        roleColumnWidths.value = settings.roleColumnWidths
      }
      if (settings.permColumnsVisible) {
        permColumnsVisible.value = { ...permColumnsVisible.value, ...settings.permColumnsVisible }
      }
      if (settings.permColumnWidths) {
        permColumnWidths.value = settings.permColumnWidths
      }
    }
  } catch {}
}

function saveColumnSettings() {
  try {
    const settings = {
      userColumnsVisible: userColumnsVisible.value,
      userColumnWidths: userColumnWidths.value,
      roleColumnsVisible: roleColumnsVisible.value,
      roleColumnWidths: roleColumnWidths.value,
      permColumnsVisible: permColumnsVisible.value,
      permColumnWidths: permColumnWidths.value
    }
    localStorage.setItem('internalAuthColumnSettings', JSON.stringify(settings))
  } catch {}
}

function startResize(colKey: string, event: MouseEvent, tableType: string) {
  resizingColumn.value = `${tableType}_${colKey}`
  resizeStartX.value = event.clientX
  
  const widths = tableType === 'user' ? userColumnWidths.value : tableType === 'role' ? roleColumnWidths.value : permColumnWidths.value
  const currentWidth = widths[colKey]
  if (currentWidth) {
    const match = currentWidth.match(/(\d+)/)
    resizeStartWidth.value = match ? parseInt(match[1]) : 100
  } else {
    resizeStartWidth.value = 100
  }
  
  document.addEventListener('mousemove', onResize)
  document.addEventListener('mouseup', stopResize)
  event.preventDefault()
}

function onResize(event: MouseEvent) {
  if (!resizingColumn.value) return
  
  const deltaX = event.clientX - resizeStartX.value
  const newWidth = Math.max(50, resizeStartWidth.value + deltaX)
  
  const [tableType, colKey] = resizingColumn.value.split('_')
  const widths = tableType === 'user' ? userColumnWidths.value : tableType === 'role' ? roleColumnWidths.value : permColumnWidths.value
  widths[colKey] = `${newWidth}px`
  
  saveColumnSettings()
}

function stopResize() {
  resizingColumn.value = null
  document.removeEventListener('mousemove', onResize)
  document.removeEventListener('mouseup', stopResize)
}

async function initTableWidths() {
  await nextTick()
  
  const initWidths = (tableSelector: string, widthsRef: any, columnsVisible: any, tableType: string) => {
    const table = document.querySelector(tableSelector) as HTMLTableElement
    if (!table) return
    
    const thead = table.querySelector('thead')
    if (!thead) return
    
    const ths = thead.querySelectorAll('th')
    let colIndex = 0
    
    const colOrder: string[] = []
    if (tableType === 'user') {
      colOrder.push('selection')
      if (columnsVisible.username) colOrder.push('username')
      if (columnsVisible.host) colOrder.push('host')
      if (columnsVisible.status) colOrder.push('status')
      if (columnsVisible.isBuiltIn) colOrder.push('isBuiltIn')
      if (columnsVisible.permissions) colOrder.push('permissions')
      colOrder.push('action')
      if (columnsVisible.createdAt) colOrder.push('createdAt')
    } else if (tableType === 'role') {
      if (columnsVisible.name) colOrder.push('name')
      if (columnsVisible.description) colOrder.push('description')
      if (columnsVisible.userCount) colOrder.push('userCount')
      if (columnsVisible.permissions) colOrder.push('permissions')
      colOrder.push('action')
      if (columnsVisible.createdAt) colOrder.push('createdAt')
    } else if (tableType === 'perm') {
      colOrder.push('selection')
      if (columnsVisible.principalName) colOrder.push('principalName')
      if (columnsVisible.principalType) colOrder.push('principalType')
      if (columnsVisible.privilegeType) colOrder.push('privilegeType')
      if (columnsVisible.objectLevel) colOrder.push('objectLevel')
      if (columnsVisible.scope) colOrder.push('scope')
      colOrder.push('action')
    }
    
    ths.forEach((th) => {
      if (colIndex < colOrder.length && colOrder[colIndex] !== 'selection' && colOrder[colIndex] !== 'action') {
        const colKey = colOrder[colIndex]
        if (!widthsRef[colKey]) {
          const rect = th.getBoundingClientRect()
          widthsRef[colKey] = `${rect.width}px`
        }
      }
      colIndex++
    })
    
    saveColumnSettings()
  }
  
  initWidths('.user-table', userColumnWidths.value, userColumnsVisible.value, 'user')
  initWidths('.role-table', roleColumnWidths.value, roleColumnsVisible.value, 'role')
  initWidths('.perm-table', permColumnWidths.value, permColumnsVisible.value, 'perm')
}

onUnmounted(() => {
  document.removeEventListener('mousemove', onResize)
  document.removeEventListener('mouseup', stopResize)
})

async function loadDatasources() {
  try {
    const res = await datasourceApi.listAllDatasources()
    datasourceList.value = res
  } catch {}
}

function handleDatasourceChange() {
  databases.value = []
  tables.value = []
  columns.value = []
  grantForm.value.databaseName = ''
  grantForm.value.tableName = ''
  grantForm.value.columns = []
  
  if (selectedDsId.value) {
    loadUsers()
    loadRoles()
    loadPermissions()
    datasourceApi.listDatabases(selectedDsId.value).then((res: any) => {
      databases.value = res.map((d: any) => d.name)
    }).catch(() => {})
  } else {
    userList.value = []
    userTotal.value = 0
    roleList.value = []
    roleTotal.value = 0
    permList.value = []
    permTotal.value = 0
  }
}

async function syncUsers() {
  if (!selectedDsId.value) return
  try {
    await internalAuthApi.syncDBUsers(selectedDsId.value)
    ElMessage.success('同步成功')
    loadUsers()
  } catch (e: any) {
    ElMessage.error(e.message || '同步失败')
  }
}

async function loadUsers() {
  if (!selectedDsId.value) return
  userLoading.value = true
  try {
    const res = await internalAuthApi.listInternalUsers({
      datasourceId: selectedDsId.value,
      keyword: userKeyword.value,
      status: userStatus.value,
      page: userPage.value,
      pageSize: userPageSize.value
    })
    userList.value = res.list || []
    userTotal.value = res.total || 0
  } catch {}
  userLoading.value = false
}

function renderUserCell(row: any, colKey: string): string {
  switch (colKey) {
    case 'host':
      return row.host || '%'
    case 'status':
      const type = row.status === 'active' ? 'success' : 'danger'
      const text = row.status === 'active' ? '启用' : '禁用'
      return `<span style="display:inline-block;padding:2px 8px;border-radius:4px;font-size:11px;font-weight:500;background:${type === 'success' ? '#d4edda' : '#f8d7da'};color:${type === 'success' ? '#155724' : '#721c24'}">${text}</span>`
    case 'isBuiltIn':
      return row.isBuiltIn ? '是' : '-'
    case 'permissions':
      return `<button class="rt-action-btn rt-action-btn-view" style="color:#007bff" title="查看权限">查看权限</button>`
    case 'actions':
      const btnStyle = 'background:none;border:none;font-size:12px;cursor:pointer;padding:2px 6px;margin:0 4px;border-radius:4px;transition:background 0.15s;'
      const editBtn = `<button style="${btnStyle}color:#1976d2;" title="编辑">编辑</button>`
      const resetBtn = `<button style="${btnStyle}color:#6c757d;" title="重置密码">重置密码</button>`
      const toggleBtn = `<button style="${btnStyle}color:#ffc107;" ${row.isBuiltIn ? 'disabled style="opacity:0.4;cursor:not-allowed;"' : ''} title="${row.status === 'active' ? '禁用' : '启用'}">${row.status === 'active' ? '禁用' : '启用'}</button>`
      const deleteBtn = `<button style="${btnStyle}color:#dc3545;" ${row.isBuiltIn ? 'disabled style="opacity:0.4;cursor:not-allowed;"' : ''} title="删除">删除</button>`
      return editBtn + resetBtn + toggleBtn + deleteBtn
    default:
      return ''
  }
}

function handleUserCellClick(row: any, colKey: string, event: any) {
  const target = event.target || event.srcElement
  if (target.tagName === 'BUTTON' || target.closest('button')) {
    const btnText = target.textContent.trim()
    switch (btnText) {
      case '编辑':
        editUser(row)
        break
      case '重置密码':
        resetPassword(row)
        break
      case '禁用':
      case '启用':
        toggleStatus(row)
        break
      case '删除':
        deleteUser(row)
        break
      case '查看权限':
        showUserPermissions(row)
        break
    }
  }
}

function renderRoleCell(row: any, colKey: string): string {
  switch (colKey) {
    case 'permissions':
      return `<button class="rt-action-btn rt-action-btn-view" style="color:#007bff" title="查看权限">查看权限</button>`
    case 'actions':
      const btnStyle = 'background:none;border:none;font-size:12px;cursor:pointer;padding:2px 6px;margin:0 4px;border-radius:4px;transition:background 0.15s;'
      const editBtn = `<button style="${btnStyle}color:#1976d2;" title="编辑">编辑</button>`
      const deleteBtn = `<button style="${btnStyle}color:#dc3545;" title="删除">删除</button>`
      return editBtn + deleteBtn
    default:
      return ''
  }
}

function handleRoleCellClick(row: any, colKey: string, event: any) {
  const target = event.target || event.srcElement
  if (target.tagName === 'BUTTON' || target.closest('button')) {
    const btnText = target.textContent.trim()
    switch (btnText) {
      case '编辑':
        editRole(row)
        break
      case '删除':
        deleteRole(row)
        break
      case '查看权限':
        showRolePermissions(row)
        break
    }
  }
}

function renderPermCell(row: any, colKey: string): string {
  switch (colKey) {
    case 'principalType':
      const text = row.principalType === 'user' ? '用户' : '角色'
      return `<span style="display:inline-block;padding:2px 6px;border-radius:4px;font-size:11px;background:#e9ecef;color:#495057">${text}</span>`
    case 'privilegeType':
      const types: Record<string, { bg: string; color: string; label: string }> = {
        readonly: { bg: '#d1ecf1', color: '#0c5460', label: '只读' },
        dml: { bg: '#fff3cd', color: '#856404', label: 'DML' },
        ddl: { bg: '#f8d7da', color: '#721c24', label: 'DDL' }
      }
      const t = types[row.privilegeType] || { bg: '#e9ecef', color: '#495057', label: row.privilegeType }
      return `<span style="display:inline-block;padding:2px 6px;border-radius:4px;font-size:11px;font-weight:500;background:${t.bg};color:${t.color}">${t.label}</span>`
    case 'objectLevel':
      const levels: Record<string, string> = {
        database: '数据库',
        table: '表',
        column: '列',
        view: '视图',
        trigger: '触发器'
      }
      return levels[row.objectLevel] || row.objectLevel
    case 'actions':
      return `<button class="rt-action-btn rt-action-btn-delete" title="回收">回收</button>`
    default:
      return ''
  }
}

function handlePermCellClick(row: any, colKey: string, event: any) {
  const target = event.target || event.srcElement
  if (target.tagName === 'BUTTON' || target.closest('button')) {
    const btnText = target.textContent.trim()
    if (btnText === '回收') {
      revokePerm(row)
    }
  }
}

function handleUserSelectionChange(rows: any[]) {
  selectedUsers.value = rows
}

function toggleUserColumn(col: string) {
  userColumnsVisible.value[col as keyof typeof userColumnsVisible.value] = !userColumnsVisible.value[col as keyof typeof userColumnsVisible.value]
  saveColumnSettings()
  nextTick(() => initTableWidths())
}

function toggleRoleColumn(col: string) {
  roleColumnsVisible.value[col as keyof typeof roleColumnsVisible.value] = !roleColumnsVisible.value[col as keyof typeof roleColumnsVisible.value]
  saveColumnSettings()
  nextTick(() => initTableWidths())
}

function togglePermColumn(col: string) {
  permColumnsVisible.value[col as keyof typeof permColumnsVisible.value] = !permColumnsVisible.value[col as keyof typeof permColumnsVisible.value]
  saveColumnSettings()
  nextTick(() => initTableWidths())
}

async function showUserPermissions(user: any) {
  permissionsDialogType.value = 'user'
  permissionsDialogTitle.value = `${user.username} 的权限列表`
  permissionsLoading.value = true
  showPermissionsDialog.value = true
  
  try {
    const res = await internalAuthApi.getUserGrants(user.id)
    grantsOutput.value = res.grants || ''
  } catch {
    grantsOutput.value = ''
  }
  permissionsLoading.value = false
}

async function showRolePermissions(role: any) {
  permissionsDialogType.value = 'role'
  permissionsDialogTitle.value = `${role.name} 的权限列表`
  permissionsLoading.value = true
  showPermissionsDialog.value = true
  
  try {
    const res = await internalAuthApi.getRoleGrants(role.id)
    grantsOutput.value = res.grants || ''
  } catch {
    grantsOutput.value = ''
  }
  permissionsLoading.value = false
}

function editUser(row: any) {
  editingUser.value = row
  userForm.value = {
    username: row.username,
    host: row.host || '',
    password: '',
    remark: row.remark || ''
  }
  showUserDialog.value = true
}

function resetPassword(row: any) {
  ElMessageBox.prompt('输入新密码（不填则自动生成）', '重置密码', {
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(({ value }) => {
    internalAuthApi.resetUserPassword(row.id, { password: value || '' }).then(() => {
      ElMessage.success('密码已重置')
    }).catch(() => {})
  }).catch(() => {})
}

function toggleStatus(row: any) {
  const enable = row.status === 'inactive'
  internalAuthApi.toggleUserStatus(row.id, enable).then(() => {
    row.status = enable ? 'active' : 'inactive'
    ElMessage.success(enable ? '已启用' : '已禁用')
  }).catch(() => {})
}

function deleteUser(row: any) {
  ElMessageBox.confirm('确定删除该用户？', '提示', {
    type: 'warning'
  }).then(() => {
    internalAuthApi.deleteInternalUser(row.id).then(() => {
      ElMessage.success('删除成功')
      loadUsers()
    }).catch(() => {})
  }).catch(() => {})
}

async function saveUser() {
  if (!userFormRef.value) return
  if (!selectedDsId.value) {
    ElMessage.warning('请先选择数据源')
    return
  }
  await userFormRef.value.validate()
  try {
    if (editingUser.value) {
      await internalAuthApi.updateInternalUser(editingUser.value.id, {
        host: userForm.value.host,
        remark: userForm.value.remark
      })
    } else {
      await internalAuthApi.createInternalUser({
        datasourceId: selectedDsId.value,
        username: userForm.value.username,
        host: userForm.value.host || '%',
        password: userForm.value.password,
        remark: userForm.value.remark
      })
    }
    showUserDialog.value = false
    ElMessage.success('保存成功')
    loadUsers()
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  }
}

async function loadRoles() {
  if (!selectedDsId.value) return
  roleLoading.value = true
  try {
    const res = await internalAuthApi.listInternalRoles({
      datasourceId: selectedDsId.value,
      keyword: roleKeyword.value,
      page: rolePage.value,
      pageSize: rolePageSize.value
    })
    roleList.value = res.list || []
    roleTotal.value = res.total || 0
    for (const role of roleList.value) {
      const countRes = await internalAuthApi.getRoleUserCount(role.id)
      role.userCount = countRes.count || 0
    }
  } catch {}
  roleLoading.value = false
}

async function editRole(row: any) {
  editingRole.value = row
  roleForm.value = {
    name: row.name,
    description: row.description || ''
  }
  editRoleActiveTab.value = 'basic'
  await loadEditingRolePermissions(row.id)
  showRoleDialog.value = true
}

async function loadEditingRolePermissions(roleId: string) {
  try {
    const res = await internalAuthApi.getRolePermissions(roleId)
    editingRolePermissions.value = res || []
  } catch {
    editingRolePermissions.value = []
  }
}

function deleteRole(row: any) {
  ElMessageBox.confirm('确定删除该角色？', '提示', {
    type: 'warning'
  }).then(() => {
    internalAuthApi.deleteInternalRole(row.id).then(() => {
      ElMessage.success('删除成功')
      loadRoles()
    }).catch(() => {})
  }).catch(() => {})
}

async function saveRole() {
  if (!roleFormRef.value) return
  if (!selectedDsId.value) {
    ElMessage.warning('请先选择数据源')
    return
  }
  await roleFormRef.value.validate()
  try {
    if (editingRole.value) {
      await internalAuthApi.updateInternalRole(editingRole.value.id, roleForm.value)
    } else {
      await internalAuthApi.createInternalRole({
        datasourceId: selectedDsId.value,
        name: roleForm.value.name,
        description: roleForm.value.description
      })
    }
    showRoleDialog.value = false
    ElMessage.success('保存成功')
    loadRoles()
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  }
}

async function loadPermissions() {
  if (!selectedDsId.value) return
  permLoading.value = true
  try {
    const res = await internalAuthApi.listPermissionRules({
      datasourceId: selectedDsId.value,
      principalType: permPrincipalType.value,
      privilegeType: permType.value,
      objectLevel: permLevel.value,
      page: permPage.value,
      pageSize: permPageSize.value
    })
    const list = res.list || []
    for (const rule of list) {
      if (rule.principalType === 'user') {
        const user = userList.value.find(u => u.id === rule.principalId)
        rule.principalName = user?.username || rule.principalId
      } else {
        const role = roleList.value.find(r => r.id === rule.principalId)
        rule.principalName = role?.name || rule.principalId
      }
      let scope = rule.databaseName
      if (rule.table) scope += '.' + rule.table
      rule.scope = scope
    }
    permList.value = list
    permTotal.value = res.total || 0
  } catch {}
  permLoading.value = false
}

function handlePermSelectionChange(rows: any[]) {
  selectedPerms.value = rows
}

function revokePerm(row: any) {
  ElMessageBox.confirm('确定回收该权限？', '提示', {
    type: 'warning'
  }).then(() => {
    internalAuthApi.revokePermission(row.id).then(() => {
      ElMessage.success('回收成功')
      loadPermissions()
    }).catch(() => {})
  }).catch(() => {})
}

function batchRevokePerm() {
  if (selectedPerms.value.length === 0) return
  ElMessageBox.confirm(`确定回收选中的 ${selectedPerms.value.length} 条权限？`, '提示', {
    type: 'warning'
  }).then(() => {
    const ids = selectedPerms.value.map((p: any) => p.id)
    internalAuthApi.batchRevokePermission({ ids }).then(() => {
      ElMessage.success('批量回收成功')
      loadPermissions()
    }).catch(() => {})
  }).catch(() => {})
}

function getPrivTagType(type: string) {
  switch (type) {
    case 'readonly': return 'info'
    case 'dml': return 'warning'
    case 'ddl': return 'danger'
    default: return ''
  }
}

function getPrivLabel(type: string) {
  switch (type) {
    case 'readonly': return '只读'
    case 'dml': return 'DML'
    case 'ddl': return 'DDL'
    default: return type
  }
}

function getLevelLabel(level: string) {
  switch (level) {
    case 'database': return '数据库'
    case 'table': return '表'
    case 'column': return '列'
    case 'view': return '视图'
    case 'trigger': return '触发器'
    default: return level
  }
}

async function onObjectLevelChange() {
  tables.value = []
  columns.value = []
  grantForm.value.tableName = ''
  grantForm.value.columns = []
}

async function onDatabaseChange() {
  tables.value = []
  columns.value = []
  grantForm.value.tableName = ''
  grantForm.value.columns = []
  if (!grantForm.value.databaseName || !selectedDsId.value) return
  try {
    const res = await datasourceApi.listTables(selectedDsId.value, grantForm.value.databaseName)
    tables.value = res.map((t: any) => t.name)
  } catch {}
}

async function onTableChange() {
  columns.value = []
  grantForm.value.columns = []
  if (!grantForm.value.tableName || !selectedDsId.value) return
  try {
    const res = await listColumns(selectedDsId.value, grantForm.value.databaseName, grantForm.value.tableName)
    columns.value = res.map((c: any) => c.name)
  } catch {}
}

function openGrantSubDialog() {
  grantSubPrincipalType.value = 'user'
  grantSubPrincipalId.value = editingUser.value?.id || editingRole.value?.id || ''
  grantSubForm.value = {
    privilegeType: 'readonly',
    objectLevel: 'database',
    databaseName: '',
    tableName: '',
    columns: []
  }
  editingPermTables.value = []
  editingPermColumns.value = []
  datasourceApi.listDatabases(selectedDsId.value).then((res: any) => {
    editingPermDatabases.value = res.map((d: any) => d.name)
  }).catch(() => {})
  showGrantSubDialog.value = true
}

function openRoleGrantSubDialog() {
  grantSubPrincipalType.value = 'role'
  grantSubPrincipalId.value = editingRole.value?.id || ''
  grantSubForm.value = {
    privilegeType: 'readonly',
    objectLevel: 'database',
    databaseName: '',
    tableName: '',
    columns: []
  }
  editingPermTables.value = []
  editingPermColumns.value = []
  datasourceApi.listDatabases(selectedDsId.value).then((res: any) => {
    editingPermDatabases.value = res.map((d: any) => d.name)
  }).catch(() => {})
  showGrantSubDialog.value = true
}

function onGrantSubObjectLevelChange() {
  editingPermTables.value = []
  editingPermColumns.value = []
  grantSubForm.value.tableName = ''
  grantSubForm.value.columns = []
}

async function onGrantSubDatabaseChange() {
  editingPermTables.value = []
  editingPermColumns.value = []
  grantSubForm.value.tableName = ''
  grantSubForm.value.columns = []
  if (!grantSubForm.value.databaseName || !selectedDsId.value) return
  try {
    const res = await datasourceApi.listTables(selectedDsId.value, grantSubForm.value.databaseName)
    editingPermTables.value = res.map((t: any) => t.name)
  } catch {}
}

async function onGrantSubTableChange() {
  editingPermColumns.value = []
  grantSubForm.value.columns = []
  if (!grantSubForm.value.tableName || !selectedDsId.value) return
  try {
    const res = await listColumns(selectedDsId.value, grantSubForm.value.databaseName, grantSubForm.value.tableName)
    editingPermColumns.value = res.map((c: any) => c.name)
  } catch {}
}

async function submitGrantSub() {
  if (!selectedDsId.value) {
    ElMessage.warning('请先选择数据源')
    return
  }
  if (!grantSubPrincipalId.value || !grantSubForm.value.databaseName) {
    ElMessage.warning('请填写完整信息')
    return
  }
  try {
    await internalAuthApi.grantPermission({
      datasourceId: selectedDsId.value,
      principalType: grantSubPrincipalType.value,
      principalId: grantSubPrincipalId.value,
      privilegeType: grantSubForm.value.privilegeType,
      objectLevel: grantSubForm.value.objectLevel,
      databaseName: grantSubForm.value.databaseName,
      tableName: grantSubForm.value.tableName,
      columns: grantSubForm.value.columns
    })
    showGrantSubDialog.value = false
    ElMessage.success('授权成功')
    if (grantSubPrincipalType.value === 'user' && editingUser.value) {
      await loadEditingUserPermissions(editingUser.value.id)
    } else if (grantSubPrincipalType.value === 'role' && editingRole.value) {
      await loadEditingRolePermissions(editingRole.value.id)
    }
  } catch (e: any) {
    ElMessage.error(e.message || '授权失败')
  }
}

async function revokePermissionDirect(perm: any) {
  ElMessageBox.confirm(`确定回收权限: ${perm.databaseName}`, '提示', {
    type: 'warning'
  }).then(async () => {
    try {
      await internalAuthApi.revokePermission(perm.id)
      ElMessage.success('回收成功')
      if (editingUser.value) {
        await loadEditingUserPermissions(editingUser.value.id)
      } else if (editingRole.value) {
        await loadEditingRolePermissions(editingRole.value.id)
      }
    } catch (e: any) {
      ElMessage.error(e.message || '回收失败')
    }
  }).catch(() => {})
}

async function submitGrant() {
  if (!selectedDsId.value) {
    ElMessage.warning('请先选择数据源')
    return
  }
  if (!grantForm.value.principalId || !grantForm.value.databaseName) {
    ElMessage.warning('请填写完整信息')
    return
  }
  try {
    await internalAuthApi.grantPermission({
      datasourceId: selectedDsId.value,
      principalType: grantForm.value.principalType,
      principalId: grantForm.value.principalId,
      privilegeType: grantForm.value.privilegeType,
      objectLevel: grantForm.value.objectLevel,
      databaseName: grantForm.value.databaseName,
      tableName: grantForm.value.tableName,
      columns: grantForm.value.columns
    })
    showGrantDialog.value = false
    ElMessage.success('授权成功')
    loadPermissions()
  } catch (e: any) {
    ElMessage.error(e.message || '授权失败')
  }
}

watch(activeTab, (tab) => {
  if (tab === 'users') {
    loadUsers()
    nextTick(() => initTableWidths())
  } else if (tab === 'roles') {
    loadRoles()
    nextTick(() => initTableWidths())
  } else if (tab === 'permissions') {
    loadPermissions()
    nextTick(() => initTableWidths())
  }
})

onMounted(() => {
  loadSavedColumnSettings()
  loadDatasources()
})
</script>

<style scoped>
.page-container {
  padding: 12px;
  height: calc(100vh - 60px);
  display: flex;
  flex-direction: column;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #e8e8e8;
}

.title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.title-text {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.refresh-icon {
  margin-right: 4px;
}

.tabs-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.main-tabs {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.main-tabs :deep(.el-tabs__content) {
  flex: 1;
  overflow: hidden;
  padding: 0;
}

.tab-content {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  flex-wrap: wrap;
  align-items: center;
}

.toolbar-right {
  margin-left: auto;
}

:deep(.el-table) {
  flex: 1;
  overflow-y: auto;
}

:deep(.el-table__row) {
  height: 40px;
}

:deep(.el-table th.el-table__cell) {
  padding: 8px 12px;
}

:deep(.el-table td.el-table__cell) {
  padding: 6px 12px;
}

:deep(.el-pagination) {
  margin-top: 8px;
}

.perm-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 8px;
}

.compact-dialog :deep(.el-dialog__body) {
  padding: 16px;
}

.compact-dialog :deep(.el-dialog__header) {
  padding: 12px 16px;
}

.compact-dialog :deep(.el-dialog__footer) {
  padding: 12px 16px;
}

.compact-dialog :deep(.el-form-item) {
  margin-bottom: 12px;
}

.permissions-loading {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 40px;
}

.permissions-empty {
  text-align: center;
  color: #909399;
  padding: 40px;
}

.grants-output {
  max-height: 400px;
  overflow-y: auto;
}

.sql-code {
  background-color: #1e1e1e;
  color: #d4d4d4;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  padding: 16px;
  border-radius: 4px;
  white-space: pre-wrap;
  word-break: break-all;
  line-height: 1.6;
}

:deep(.el-dropdown-menu__item) .checked {
  font-weight: 600;
  color: #409EFF;
}

:deep(.el-dropdown-menu__item span) {
  display: inline-flex;
  align-items: center;
}

.permission-panel {
  padding: 8px 0;
}

.permission-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #e8e8e8;
}

.permission-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.permission-list {
  min-height: 200px;
}

.permission-empty {
  text-align: center;
  color: #909399;
  padding: 40px;
  font-size: 14px;
}

.permission-items {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.permission-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background-color: #fafafa;
  border-radius: 4px;
  border: 1px solid #e8e8e8;
}

.perm-info {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.perm-scope {
  font-size: 13px;
  color: #606266;
  font-weight: 500;
}

.perm-level {
  font-size: 12px;
  color: #909399;
  padding: 2px 6px;
  background-color: #f0f0f0;
  border-radius: 4px;
}

.auto-width-table {
  table-layout: auto !important;
}

.auto-width-table :deep(.el-table__header-wrapper),
.auto-width-table :deep(.el-table__body-wrapper) {
  overflow-x: auto;
}

.auto-width-table :deep(.el-table__column-resize-proxy) {
  position: absolute;
  z-index: 10;
  width: 2px;
  top: 0;
  bottom: 0;
  cursor: col-resize;
  background-color: #409EFF;
}

.auto-width-table :deep(.el-table__cell) {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.auto-width-table :deep(.el-table__header .el-table__cell) {
  white-space: nowrap;
}

.resize-handle {
  display: inline-block;
  width: 8px;
  height: 100%;
  position: absolute;
  right: 0;
  top: 0;
  cursor: col-resize;
  z-index: 10;
}

.resize-handle:hover {
  background-color: #409EFF;
}

:deep(.el-table th.el-table__cell) {
  position: relative;
}

.el-table__column--selection {
  width: 46px !important;
  min-width: 46px !important;
}
</style>
