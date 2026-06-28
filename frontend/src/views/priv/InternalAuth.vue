<template>
  <div class="page-container">
    <div class="page-header">
      <div class="title-row">
        <el-icon :size="20" color="#409EFF"><Lock /></el-icon>
        <span class="title-text">实例权限管控</span>
      </div>
      <div class="header-actions">
        <el-select
          v-model="selectedDsId"
          placeholder="请选择数据源"
          style="width:240px;"
          filterable
          @change="handleDatasourceChange"
          size="small"
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
          size="small"
          @click="syncUsers"
          style="margin-left: 6px;"
        >
          同步用户
        </el-button>
        <el-button @click="loadDatasources" size="small" style="margin-left: 6px;"><span class="refresh-icon">⟳</span>刷新</el-button>
      </div>
    </div>

    <div class="tabs-wrapper">
      <el-tabs v-model="activeTab" class="main-tabs">
        <el-tab-pane label="内部用户" name="users">
          <div class="tab-content">
            <div class="toolbar">
              <el-input v-model="userKeyword" placeholder="用户名" clearable style="width:160px;" size="small">
                <template #prefix><el-icon><Search /></el-icon></template>
              </el-input>
              <el-select v-model="userStatus" placeholder="状态" clearable style="width:100px;" size="small">
                <el-option label="启用" value="active" />
                <el-option label="禁用" value="inactive" />
              </el-select>
              <el-button type="primary" @click="loadUsers" size="small">查询</el-button>
              <el-button @click="showUserDialog = true" size="small">新建用户</el-button>
              <div class="toolbar-right">
                <el-button size="small" @click="showUserColumnDialog = true">列选择</el-button>
                <el-button size="small" @click="refreshUsers">刷新</el-button>
              </div>
            </div>
            <el-table :data="userList" border class="auth-table" stripe :height="tableHeight">
              <el-table-column v-if="userColumnConfigs.find(c => c.key === 'index')?.visible" type="index" label="#" min-width="40" />
              <el-table-column v-if="userColumnConfigs.find(c => c.key === 'username')?.visible" prop="username" label="用户名" min-width="90" />
              <el-table-column v-if="userColumnConfigs.find(c => c.key === 'host')?.visible && selectedDs?.dbType !== 'sqlite'" prop="host" label="主机" min-width="70" />
              <el-table-column v-if="userColumnConfigs.find(c => c.key === 'status')?.visible" prop="status" label="状态" min-width="60">
                <template #default="{ row }">
                  <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small">
                    {{ row.status === 'active' ? '启用' : '禁用' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column v-if="userColumnConfigs.find(c => c.key === 'permissions')?.visible" label="权限" min-width="150">
                <template #default="{ row }">
                  <el-button size="small" type="text" @click="showUserPermissions(row)">查看权限 <span class="refresh-icon" @click.stop="refreshUserPermissions(row)" style="cursor:pointer;">⟳</span></el-button>
                </template>
              </el-table-column>
              <el-table-column v-if="userColumnConfigs.find(c => c.key === 'isBuiltIn')?.visible" prop="isBuiltIn" label="内置" min-width="50">
                <template #default="{ row }">{{ row.isBuiltIn ? '是' : '否' }}</template>
              </el-table-column>
              <el-table-column v-if="userColumnConfigs.find(c => c.key === 'createdAt')?.visible" prop="createdAt" label="创建时间" min-width="140">
                <template #default="{ row }">{{ formatDateTime(row.createdAt) }}</template>
              </el-table-column>
              <el-table-column v-if="userColumnConfigs.find(c => c.key === 'actions')?.visible" label="操作" min-width="280">
                <template #default="{ row }">
                  <el-button size="small" @click="editUser(row)">编辑</el-button>
                  <el-button size="small" @click="resetPassword(row)">重置密码</el-button>
                  <el-button size="small" :type="row.status === 'active' ? 'warning' : 'success'" @click="toggleStatus(row)">
                    {{ row.status === 'active' ? '禁用' : '启用' }}
                  </el-button>
                  <el-button size="small" type="success" @click="openUserRoleDialog(row)" :disabled="row.isBuiltIn">分配角色</el-button>
                  <el-button size="small" type="primary" @click="editUserPermissions(row)" :disabled="row.isBuiltIn">编辑权限</el-button>
                  <el-button size="small" type="danger" @click="deleteUser(row)" v-if="!row.isBuiltIn">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
            <div class="pagination-wrapper">
              <el-pagination
                v-model:current-page="userPage"
                v-model:page-size="userPageSize"
                :total="userTotal"
                :page-sizes="[10, 20, 50, 100]"
                layout="total, sizes, prev, pager, next, jumper"
                @size-change="loadUsers"
                @current-change="loadUsers"
                background
              />
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="角色管理" name="roles">
          <div class="tab-content">
            <div class="toolbar">
              <el-input v-model="roleKeyword" placeholder="角色名称" clearable style="width:160px;" size="small">
                <template #prefix><el-icon><Search /></el-icon></template>
              </el-input>
              <el-button type="primary" @click="loadRoles" size="small">查询</el-button>
              <el-button @click="showRoleDialog = true" size="small">新建角色</el-button>
              <div class="toolbar-right">
                <el-button size="small" @click="showRoleColumnDialog = true">列选择</el-button>
                <el-button size="small" @click="refreshRoles">刷新</el-button>
              </div>
            </div>
            <el-table :data="roleList" border class="auth-table" stripe :height="tableHeight">
              <el-table-column v-if="roleColumnConfigs.find(c => c.key === 'index')?.visible" type="index" label="#" min-width="40" />
              <el-table-column v-if="roleColumnConfigs.find(c => c.key === 'name')?.visible" prop="name" label="角色名称" min-width="110" />
              <el-table-column v-if="roleColumnConfigs.find(c => c.key === 'description')?.visible" prop="description" label="描述" min-width="180" />
              <el-table-column v-if="roleColumnConfigs.find(c => c.key === 'userCount')?.visible" prop="userCount" label="关联用户数" min-width="80" />
              <el-table-column v-if="roleColumnConfigs.find(c => c.key === 'createdAt')?.visible" prop="createdAt" label="创建时间" min-width="140">
                <template #default="{ row }">{{ formatDateTime(row.createdAt) }}</template>
              </el-table-column>
              <el-table-column v-if="roleColumnConfigs.find(c => c.key === 'actions')?.visible" label="操作" min-width="120">
                <template #default="{ row }">
                  <el-button size="small" @click="editRole(row)">编辑</el-button>
                  <el-button size="small" type="danger" @click="deleteRole(row)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
            <div class="pagination-wrapper">
              <el-pagination
                v-model:current-page="rolePage"
                v-model:page-size="rolePageSize"
                :total="roleTotal"
                :page-sizes="[10, 20, 50, 100]"
                layout="total, sizes, prev, pager, next, jumper"
                @size-change="loadRoles"
                @current-change="loadRoles"
                background
              />
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="审计日志" name="audit">
          <div class="tab-content">
            <div class="toolbar">
              <el-input v-model="auditOperator" placeholder="操作人" clearable style="width:160px;" size="small">
                <template #prefix><el-icon><Search /></el-icon></template>
              </el-input>
              <el-select v-model="auditOperType" placeholder="操作类型" clearable style="width:120px;" size="small">
                <el-option label="创建用户" value="create_user" />
                <el-option label="修改用户" value="update_user" />
                <el-option label="删除用户" value="delete_user" />
                <el-option label="重置密码" value="reset_password" />
                <el-option label="创建角色" value="create_role" />
                <el-option label="修改角色" value="update_role" />
                <el-option label="删除角色" value="delete_role" />
                <el-option label="授权" value="grant_permission" />
                <el-option label="回收权限" value="revoke_permission" />
              </el-select>
              <el-select v-model="auditResult" placeholder="操作结果" clearable style="width:90px;" size="small">
                <el-option label="成功" value="success" />
                <el-option label="失败" value="failed" />
              </el-select>
              <el-button type="primary" @click="loadAuditLogs" size="small">查询</el-button>
              <div class="toolbar-right">
                <el-button size="small" @click="refreshAuditLogs">刷新</el-button>
              </div>
            </div>
            <el-table :data="auditLogList" border class="auth-table" stripe :height="tableHeight">
              <el-table-column type="index" label="#" min-width="40" />
              <el-table-column prop="operator" label="操作人" min-width="90" />
              <el-table-column prop="operType" label="操作类型" min-width="100">
                <template #default="{ row }">{{ getOperTypeLabel(row.operType) }}</template>
              </el-table-column>
              <el-table-column prop="operObject" label="操作对象" min-width="100" />
              <el-table-column prop="result" label="结果" min-width="70">
                <template #default="{ row }">
                  <el-tag :type="row.result === 'success' ? 'success' : 'danger'" size="small">
                    {{ row.result === 'success' ? '成功' : '失败' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="detail" label="详情" min-width="180" />
              <el-table-column prop="operTime" label="操作时间" min-width="140">
                <template #default="{ row }">{{ formatDateTime(row.operTime) }}</template>
              </el-table-column>
            </el-table>
            <div class="pagination-wrapper">
              <el-pagination
                v-model:current-page="auditPage"
                v-model:page-size="auditPageSize"
                :total="auditTotal"
                :page-sizes="[10, 20, 50, 100]"
                layout="total, sizes, prev, pager, next, jumper"
                @size-change="loadAuditLogs"
                @current-change="loadAuditLogs"
                background
              />
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>

    <el-dialog v-model="showUserDialog" :title="editingUser ? '编辑用户' : '新建用户'" width="540px" class="compact-dialog">
      <el-form :model="userForm" :rules="userRules" ref="userFormRef" label-width="70px" :inline="true">
        <template v-if="!editingUser">
          <el-form-item label="用户名" prop="username">
            <el-input v-model="userForm.username" style="width:180px;" />
          </el-form-item>
          <el-form-item label="主机" prop="host" v-if="selectedDs?.dbType !== 'sqlite'">
            <el-input v-model="userForm.host" placeholder="默认 %" style="width:120px;" />
          </el-form-item>
          <el-form-item label="密码" prop="password">
            <el-input v-model="userForm.password" type="password" placeholder="不填则自动生成" style="width:180px;" />
          </el-form-item>
        </template>
        <template v-else>
          <el-form-item label="用户名">
            <el-tag size="small">{{ userForm.username }}</el-tag>
          </el-form-item>
          <el-form-item label="主机" v-if="selectedDs?.dbType !== 'sqlite'">
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

    <el-dialog v-model="showRoleDialog" :title="editingRole ? '编辑角色' : '新建角色'" width="680px" class="compact-dialog">
      <el-tabs v-model="editRoleActiveTab" size="small">
        <el-tab-pane label="基本信息" name="basic">
          <el-form :model="roleForm" :rules="roleRules" ref="roleFormRef" label-width="70px" :inline="true">
            <el-form-item label="角色名称" prop="name">
              <el-input v-model="roleForm.name" style="width:200px;" />
            </el-form-item>
            <el-form-item label="描述" prop="description" style="width:100%;">
              <el-input v-model="roleForm.description" type="textarea" :rows="2" style="width:100%;" />
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="权限配置" name="permissions" v-if="editingRole">
          <div class="permission-panel">
            <div class="permission-header">
              <span class="permission-title">对象权限</span>
              <el-button size="small" type="primary" @click="openGrantSubDialog('role')">+ 添加对象权限</el-button>
            </div>
            <div class="permission-list">
              <div v-if="editingRoleObjectPermissions.length === 0" class="permission-empty">暂无对象权限</div>
              <div v-else class="permission-items">
                <div v-for="perm in editingRoleObjectPermissions" :key="perm.id" class="permission-item">
                  <div class="perm-info">
                    <el-tag :type="getPrivTagType(perm.privilegeType)" size="small">{{ getPrivLabel(perm.privilegeType) }}</el-tag>
                    <span class="perm-scope">{{ perm.databaseName }}{{ perm.tableName ? '.' + perm.tableName : '' }}</span>
                    <span class="perm-level">{{ getLevelLabel(perm.objectLevel) }}</span>
                  </div>
                  <el-button size="small" type="danger" @click="revokePermissionDirect(perm)">回收</el-button>
                </div>
              </div>
            </div>
            <div class="permission-header" v-if="selectedDs?.dbType !== 'sqlite'">
              <span class="permission-title">系统权限</span>
              <el-button size="small" type="primary" @click="openSystemPrivDialog('role')">+ 配置系统权限</el-button>
            </div>
            <div class="permission-list" v-if="selectedDs?.dbType !== 'sqlite'">
              <div v-if="editingRoleSystemPermissions.length === 0" class="permission-empty">暂无系统权限</div>
              <div v-else class="permission-items">
                <div v-for="perm in editingRoleSystemPermissions" :key="perm.id" class="permission-item">
                  <div class="perm-info">
                    <el-tag type="warning" size="small">系统权限</el-tag>
                    <span class="perm-scope">{{ parseSystemPrivs(perm.systemPrivileges) }}</span>
                  </div>
                  <el-button size="small" type="danger" @click="revokeSystemPermissionDirect(perm)">回收</el-button>
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

    <el-dialog v-model="showGrantSubDialog" title="添加对象权限" width="460px" class="compact-dialog">
      <el-form :model="grantSubForm" label-width="70px" :inline="true">
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
            <el-option label="视图" value="view" />
            <el-option label="存储过程" value="procedure" />
            <el-option label="函数" value="function" />
            <el-option label="触发器" value="trigger" />
            <el-option label="事件" value="event" />
          </el-select>
        </el-form-item>
        <el-form-item label="数据库">
          <el-select v-model="grantSubForm.databaseName" style="width:100%" filterable @change="onGrantSubDatabaseChange" size="small">
            <el-option v-for="db in editingPermDatabases" :key="db" :label="db" :value="db" />
          </el-select>
        </el-form-item>
        <el-form-item label="对象" v-if="grantSubForm.objectLevel !== 'database'">
          <el-select v-model="grantSubForm.tableName" style="width:100%" filterable @change="onGrantSubTableChange" size="small">
            <el-option v-for="t in editingPermObjects" :key="t" :label="t" :value="t" />
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

    <el-dialog v-model="showSystemPrivDialog" title="配置系统权限" width="560px" class="compact-dialog">
      <div v-if="systemPrivilegesLoading" class="permissions-loading">加载中...</div>
      <div v-else>
        <div v-for="group in groupedSystemPrivileges" :key="group.category" class="system-priv-group">
          <div class="system-priv-group-title">{{ group.category }}</div>
          <div class="system-priv-items">
            <label v-for="priv in group.privileges" :key="priv.name" class="system-priv-item">
              <el-checkbox :checked="selectedSystemPrivileges.includes(priv.name)" @change="toggleSystemPrivilege(priv.name)" />
              <span>{{ priv.name }}</span>
              <span style="color:#909399;font-size:11px;">({{ priv.description }})</span>
            </label>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="showSystemPrivDialog = false" size="small">取消</el-button>
        <el-button type="primary" @click="submitSystemPriv" size="small">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showGrantDialog" title="授权" width="440px" class="compact-dialog">
      <el-form :model="grantForm" label-width="70px">
        <el-form-item label="授权主体">
          <el-radio-group v-model="grantForm.principalType">
            <el-radio value="user">用户</el-radio>
            <el-radio value="role">角色</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="选择主体">
          <el-select v-model="grantForm.principalId" style="width:100%" filterable>
            <el-option v-for="item in principalOptions" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="授权类型">
          <el-radio-group v-model="grantForm.grantType">
            <el-radio value="object">对象权限</el-radio>
            <el-radio value="system" :disabled="selectedDs?.dbType === 'sqlite'">系统权限</el-radio>
          </el-radio-group>
        </el-form-item>
        <template v-if="grantForm.grantType === 'object'">
          <el-form-item label="权限类型">
            <el-select v-model="grantForm.privilegeType" style="width:100%">
              <el-option label="只读" value="readonly" />
              <el-option label="DML" value="dml" />
              <el-option label="DDL" value="ddl" />
            </el-select>
          </el-form-item>
          <el-form-item label="对象层级">
            <el-select v-model="grantForm.objectLevel" style="width:100%" @change="onObjectLevelChange">
              <el-option label="数据库" value="database" />
              <el-option label="表" value="table" />
              <el-option label="列" value="column" />
              <el-option label="视图" value="view" />
              <el-option label="存储过程" value="procedure" />
              <el-option label="函数" value="function" />
              <el-option label="触发器" value="trigger" />
              <el-option label="事件" value="event" />
            </el-select>
          </el-form-item>
          <el-form-item label="数据库">
            <el-select v-model="grantForm.databaseName" style="width:100%" filterable @change="onDatabaseChange">
              <el-option v-for="db in databases" :key="db" :label="db" :value="db" />
            </el-select>
          </el-form-item>
          <el-form-item label="对象" v-if="grantForm.objectLevel !== 'database'">
            <el-select v-model="grantForm.tableName" style="width:100%" filterable @change="onTableChange">
              <el-option v-for="t in objects" :key="t" :label="t" :value="t" />
            </el-select>
          </el-form-item>
          <el-form-item label="列" v-if="grantForm.objectLevel === 'column'">
            <el-select v-model="grantForm.columns" multiple style="width:100%" filterable>
              <el-option v-for="c in columns" :key="c" :label="c" :value="c" />
            </el-select>
          </el-form-item>
        </template>
        <template v-if="grantForm.grantType === 'system' && selectedDs?.dbType !== 'sqlite'">
          <div v-if="grantSystemPrivilegesLoading" class="permissions-loading">加载中...</div>
          <div v-else>
            <div v-for="group in groupedGrantSystemPrivileges" :key="group.category" class="system-priv-group">
              <div class="system-priv-group-title">{{ group.category }}</div>
              <div class="system-priv-items">
                <label v-for="priv in group.privileges" :key="priv.name" class="system-priv-item">
                  <el-checkbox :checked="grantForm.systemPrivileges.includes(priv.name)" @change="toggleGrantSystemPrivilege(priv.name)" />
                  <span>{{ priv.name }}</span>
                  <span style="color:#909399;font-size:11px;">({{ priv.description }})</span>
                </label>
              </div>
            </div>
          </div>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="showGrantDialog = false">取消</el-button>
        <el-button type="primary" @click="submitGrant">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showUserColumnDialog" title="列选择" width="340px" class="compact-dialog">
      <div v-for="col in userColumnConfigs" :key="col.key" class="column-checkbox-item">
        <el-checkbox :checked="col.visible" @change="toggleColumnVisibility(userColumnConfigs, col.key)">
          {{ col.label }}
        </el-checkbox>
      </div>
      <template #footer>
        <el-button @click="resetColumnSettings" size="small">重置</el-button>
        <el-button @click="showUserColumnDialog = false" size="small">关闭</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showRoleColumnDialog" title="列选择" width="340px" class="compact-dialog">
      <div v-for="col in roleColumnConfigs" :key="col.key" class="column-checkbox-item">
        <el-checkbox :checked="col.visible" @change="toggleColumnVisibility(roleColumnConfigs, col.key)">
          {{ col.label }}
        </el-checkbox>
      </div>
      <template #footer>
        <el-button @click="resetColumnSettings" size="small">重置</el-button>
        <el-button @click="showRoleColumnDialog = false" size="small">关闭</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showUserPermDetailDialog" :title="`${currentUserDetail?.username} 的权限详情`" width="720px" class="compact-dialog">
      <div v-if="userPermDetailLoading" class="permissions-loading">加载中...</div>
      <div v-else>
        <el-tabs v-model="userPermDetailTab" size="small">
          <el-tab-pane label="有效权限" name="effective" v-if="selectedDs?.dbType !== 'sqlite'">
            <div v-if="userEffectiveGrantsLoading" class="permissions-loading">加载中...</div>
            <div v-else-if="userEffectiveGrants.length === 0" class="permission-empty">暂无有效权限数据</div>
            <div v-else>
              <div class="grant-list">
                <div v-for="(grant, index) in userEffectiveGrants" :key="index" class="grant-item">
                  <pre class="grant-text">{{ typeof grant === 'string' ? grant : grant.grant || JSON.stringify(grant) }}</pre>
                </div>
              </div>
            </div>
          </el-tab-pane>
          <el-tab-pane label="对象权限" name="object">
            <div v-if="userPermDetailObjects.length === 0" class="permission-empty">暂无对象权限</div>
            <div v-else class="permission-items">
              <div v-for="perm in userPermDetailObjects" :key="perm.id" class="permission-item">
                <div class="perm-info">
                  <el-tag :type="getPrivTagType(perm.privilegeType)" size="small">{{ getPrivLabel(perm.privilegeType) }}</el-tag>
                  <span class="perm-scope">{{ perm.databaseName }}{{ perm.tableName ? '.' + perm.tableName : '' }}</span>
                  <span class="perm-level">{{ getLevelLabel(perm.objectLevel) }}</span>
                </div>
                <el-button size="small" type="danger" @click="revokePermissionDirect(perm)">回收</el-button>
              </div>
            </div>
          </el-tab-pane>
          <el-tab-pane label="系统权限" name="system" v-if="selectedDs?.dbType !== 'sqlite'">
            <div v-if="userPermDetailSystem.length === 0" class="permission-empty">暂无系统权限</div>
            <div v-else class="permission-items">
              <div v-for="perm in userPermDetailSystem" :key="perm.id" class="permission-item">
                <div class="perm-info">
                  <el-tag type="warning" size="small">系统权限</el-tag>
                  <span class="perm-scope">{{ parseSystemPrivs(perm.systemPrivileges) }}</span>
                </div>
                <el-button size="small" type="danger" @click="revokeSystemPermissionDirect(perm)">回收</el-button>
              </div>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
      <template #footer>
        <el-button @click="showUserPermDetailDialog = false" size="small">关闭</el-button>
        <el-button type="primary" @click="editUserPermissions(currentUserDetail)" size="small">编辑权限</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showEditPermDialog" :title="`编辑用户权限 - ${editingPermUser?.username}`" width="780px" class="compact-dialog" @open="loadEditPermData" @close="resetEditPermForm">
      <div v-if="editPermLoading" class="permissions-loading">加载中...</div>
      <div v-else>
        <el-tabs v-model="editPermActiveTab" size="small">
          <el-tab-pane label="对象权限" name="object">
            <div class="permission-panel">
              <div style="display: flex; gap: 8px; align-items: center; margin-bottom: 12px;">
                <el-select v-model="editPermSelectedObjType" placeholder="对象类型" style="width:100px;" size="small">
                  <el-option label="表" value="table" />
                  <el-option label="视图" value="view" />
                  <el-option label="存储过程" value="procedure" v-if="selectedDs?.dbType !== 'sqlite'" />
                  <el-option label="函数" value="function" v-if="selectedDs?.dbType !== 'sqlite'" />
                  <el-option label="触发器" value="trigger" v-if="selectedDs?.dbType !== 'sqlite'" />
                  <el-option label="事件" value="event" v-if="selectedDs?.dbType !== 'sqlite'" />
                </el-select>
                <el-cascader
                  v-model="editPermSelectedCascader"
                  :options="editPermCascaderOptions"
                  :props="{ checkStrictly: true, multiple: true, emitPath: false, lazy: true, lazyLoad: loadEditPermCascaderChildren }"
                  placeholder="选择库/表/列"
                  style="flex: 1;"
                  size="small"
                  @change="handleEditPermCascaderChange"
                />
                <el-button type="primary" @click="addEditPermRule" size="small" :disabled="editPermSelectedCascader.length === 0">添加规则</el-button>
              </div>

              <div style="display: flex; gap: 16px; margin-bottom: 12px;">
                <div style="flex: 1;">
                  <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px;">
                    <span style="font-size: 12px; font-weight: 600; color: #606266;">只读权限</span>
                    <el-checkbox :checked="isPrivGroupAllSelected('readonly')" @change="togglePrivGroup('readonly')">全选</el-checkbox>
                  </div>
                  <div style="display: flex; flex-wrap: wrap; gap: 6px;">
                    <el-checkbox :checked="editPermSelectedPrivs.includes('SELECT')" @change="togglePrivItem('SELECT')" label="SELECT" />
                    <el-checkbox :checked="editPermSelectedPrivs.includes('SHOW VIEW')" @change="togglePrivItem('SHOW VIEW')" label="SHOW VIEW" />
                  </div>
                </div>
                <div style="flex: 1;">
                  <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px;">
                    <span style="font-size: 12px; font-weight: 600; color: #606266;">DML权限</span>
                    <el-checkbox :checked="isPrivGroupAllSelected('dml')" @change="togglePrivGroup('dml')">全选</el-checkbox>
                  </div>
                  <div style="display: flex; flex-wrap: wrap; gap: 6px;">
                    <el-checkbox :checked="editPermSelectedPrivs.includes('INSERT')" @change="togglePrivItem('INSERT')" label="INSERT" />
                    <el-checkbox :checked="editPermSelectedPrivs.includes('UPDATE')" @change="togglePrivItem('UPDATE')" label="UPDATE" />
                    <el-checkbox :checked="editPermSelectedPrivs.includes('DELETE')" @change="togglePrivItem('DELETE')" label="DELETE" />
                    <el-checkbox :checked="editPermSelectedPrivs.includes('EXECUTE')" @change="togglePrivItem('EXECUTE')" label="EXECUTE" />
                  </div>
                </div>
                <div style="flex: 1;">
                  <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px;">
                    <span style="font-size: 12px; font-weight: 600; color: #606266;">DDL权限</span>
                    <el-checkbox :checked="isPrivGroupAllSelected('ddl')" @change="togglePrivGroup('ddl')">全选</el-checkbox>
                  </div>
                  <div style="display: flex; flex-wrap: wrap; gap: 6px;">
                    <el-checkbox :checked="editPermSelectedPrivs.includes('CREATE')" @change="togglePrivItem('CREATE')" label="CREATE" />
                    <el-checkbox :checked="editPermSelectedPrivs.includes('ALTER')" @change="togglePrivItem('ALTER')" label="ALTER" />
                    <el-checkbox :checked="editPermSelectedPrivs.includes('DROP')" @change="togglePrivItem('DROP')" label="DROP" />
                    <el-checkbox :checked="editPermSelectedPrivs.includes('INDEX')" @change="togglePrivItem('INDEX')" label="INDEX" />
                    <el-checkbox :checked="editPermSelectedPrivs.includes('TRIGGER')" @change="togglePrivItem('TRIGGER')" label="TRIGGER" />
                    <el-checkbox :checked="editPermSelectedPrivs.includes('CREATE ROUTINE')" @change="togglePrivItem('CREATE ROUTINE')" label="CREATE ROUTINE" />
                    <el-checkbox :checked="editPermSelectedPrivs.includes('ALTER ROUTINE')" @change="togglePrivItem('ALTER ROUTINE')" label="ALTER ROUTINE" />
                  </div>
                </div>
              </div>

              <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;">
                <span style="font-size: 12px; font-weight: 600; color: #606266;">已授权规则</span>
                <el-button size="small" type="danger" @click="batchRemoveEditPermRules" :disabled="editPermSelectedRules.length === 0">批量删除</el-button>
              </div>
              <el-table :data="editPermRules" border size="small" :height="200" @selection-change="handleEditPermRuleSelectionChange">
                <el-table-column type="selection" width="40" />
                <el-table-column prop="objectType" label="对象类型" width="80">
                  <template #default="{ row }">{{ getObjectLevelLabel(row.objectType) }}</template>
                </el-table-column>
                <el-table-column prop="scope" label="授权范围" min-width="200" />
                <el-table-column prop="privilegeType" label="权限类型" width="60">
                  <template #default="{ row }">
                    <el-tag :type="getPrivilegeTagType(row.privilegeType)" size="small">{{ getPrivilegeLabel(row.privilegeType) }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="50">
                  <template #default="{ row }">
                    <el-button size="small" type="danger" @click="removeEditPermRule(row)">删除</el-button>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </el-tab-pane>

          <el-tab-pane label="系统权限" name="system" v-if="selectedDs?.dbType !== 'sqlite'">
            <div v-for="group in editPermSystemPrivGroups" :key="group.category" 
                 v-show="!(group.isTiDBExclusive && selectedDs?.dbType !== 'tidb')"
                 style="margin-bottom: 12px; padding-bottom: 12px; border-bottom: 1px solid #f0f0f0;">
              <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px;">
                <span style="font-size: 13px; font-weight: 600; color: #303133; padding-left: 4px; border-left: 3px solid #409EFF;">{{ group.category }}</span>
                <el-checkbox @change="(val: boolean) => toggleSystemPrivGroup(group.category, val)">全选</el-checkbox>
              </div>
              <div style="display: flex; flex-wrap: wrap; gap: 8px;">
                <el-checkbox v-for="priv in group.privileges" :key="priv.name" 
                            :checked="editPermSelectedSystemPrivs.includes(priv.name)"
                            @change="(val: boolean) => val ? editPermSelectedSystemPrivs.push(priv.name) : editPermSelectedSystemPrivs.splice(editPermSelectedSystemPrivs.indexOf(priv.name), 1)">
                  <span style="font-size: 12px;">{{ priv.name }} <span style="color: #909399;">({{ priv.description }})</span></span>
                </el-checkbox>
              </div>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
      <template #footer>
        <el-button @click="showEditPermDialog = false" size="small">取消</el-button>
        <el-button type="primary" @click="saveEditPermissions" size="small">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showUserRoleDialog" :title="`分配角色 - ${currentRoleUser?.username}`" width="480px" class="compact-dialog" @open="loadUserRoleDialogData">
      <div v-if="availableRoleList.length === 0" style="text-align:center;color:#909399;padding:20px;">暂无可用角色</div>
      <div v-else>
        <div v-for="role in availableRoleList" :key="role.id" style="display:flex;align-items:center;padding:8px 0;border-bottom:1px solid #f5f7fa;">
          <el-checkbox :checked="selectedRoleIds.includes(role.id)" @change="(val: boolean) => toggleRoleSelection(role.id, val)" />
          <div style="flex:1;margin-left:10px;">
            <div style="font-size:14px;font-weight:500;color:#303133;">{{ role.name }}</div>
            <div style="font-size:12px;color:#909399;">{{ role.description || '无描述' }}</div>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="showUserRoleDialog = false" size="small">取消</el-button>
        <el-button type="primary" @click="saveUserRoles" size="small">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Lock } from '@element-plus/icons-vue'
import * as internalAuthApi from '@/api/internalAuth'
import * as datasourceApi from '@/api/datasource'
import { listColumns } from '@/api/sql'

const activeTab = ref('users')
const selectedDsId = ref('')
const datasourceList = ref<any[]>([])
const selectedDs = computed(() => datasourceList.value.find(ds => ds.datasourceId === selectedDsId.value))

const tableHeight = ref(400)

function updateTableHeight() {
  const containerHeight = window.innerHeight - 180
  tableHeight.value = Math.max(280, containerHeight)
}

onMounted(() => {
  updateTableHeight()
  window.addEventListener('resize', updateTableHeight)
  loadColumnSettings()
})

interface ColumnConfig {
  key: string
  label: string
  visible: boolean
}

const USER_COLUMNS_KEY = 'internal_auth_user_columns'
const ROLE_COLUMNS_KEY = 'internal_auth_role_columns'
const PERM_COLUMNS_KEY = 'internal_auth_perm_columns'

const userColumnConfigs = ref<ColumnConfig[]>([
  { key: 'index', label: '#', visible: true },
  { key: 'username', label: '用户名', visible: true },
  { key: 'host', label: '主机', visible: true },
  { key: 'status', label: '状态', visible: true },
  { key: 'permissions', label: '权限', visible: true },
  { key: 'isBuiltIn', label: '内置', visible: true },
  { key: 'createdAt', label: '创建时间', visible: true },
  { key: 'actions', label: '操作', visible: true }
])

const roleColumnConfigs = ref<ColumnConfig[]>([
  { key: 'index', label: '#', visible: true },
  { key: 'name', label: '角色名称', visible: true },
  { key: 'description', label: '描述', visible: true },
  { key: 'userCount', label: '关联用户数', visible: true },
  { key: 'createdAt', label: '创建时间', visible: true },
  { key: 'actions', label: '操作', visible: true }
])

const permColumnConfigs = ref<ColumnConfig[]>([
  { key: 'index', label: '#', visible: true },
  { key: 'principalName', label: '授权主体', visible: true },
  { key: 'principalType', label: '主体类型', visible: true },
  { key: 'privilegeCategory', label: '权限分类', visible: true },
  { key: 'privilegeType', label: '权限类型', visible: true },
  { key: 'objectLevel', label: '对象层级', visible: true },
  { key: 'scope', label: '授权范围', visible: true },
  { key: 'createdAt', label: '创建时间', visible: true },
  { key: 'actions', label: '操作', visible: true }
])

function loadColumnSettings() {
  const userSettings = localStorage.getItem(USER_COLUMNS_KEY)
  if (userSettings) {
    try {
      const saved = JSON.parse(userSettings)
      userColumnConfigs.value = userColumnConfigs.value.map(col => {
        const savedCol = saved.find((c: ColumnConfig) => c.key === col.key)
        return savedCol ? { ...col, visible: savedCol.visible } : col
      })
    } catch {}
  }

  const roleSettings = localStorage.getItem(ROLE_COLUMNS_KEY)
  if (roleSettings) {
    try {
      const saved = JSON.parse(roleSettings)
      roleColumnConfigs.value = roleColumnConfigs.value.map(col => {
        const savedCol = saved.find((c: ColumnConfig) => c.key === col.key)
        return savedCol ? { ...col, visible: savedCol.visible } : col
      })
    } catch {}
  }

  const permSettings = localStorage.getItem(PERM_COLUMNS_KEY)
  if (permSettings) {
    try {
      const saved = JSON.parse(permSettings)
      permColumnConfigs.value = permColumnConfigs.value.map(col => {
        const savedCol = saved.find((c: ColumnConfig) => c.key === col.key)
        return savedCol ? { ...col, visible: savedCol.visible } : col
      })
    } catch {}
  }
}

function saveColumnSettings() {
  localStorage.setItem(USER_COLUMNS_KEY, JSON.stringify(userColumnConfigs.value))
  localStorage.setItem(ROLE_COLUMNS_KEY, JSON.stringify(roleColumnConfigs.value))
  localStorage.setItem(PERM_COLUMNS_KEY, JSON.stringify(permColumnConfigs.value))
}

function toggleColumnVisibility(configs: ColumnConfig[], key: string) {
  const col = configs.find(c => c.key === key)
  if (col) {
    col.visible = !col.visible
    saveColumnSettings()
  }
}

function resetColumnSettings() {
  userColumnConfigs.value.forEach(col => col.visible = true)
  roleColumnConfigs.value.forEach(col => col.visible = true)
  permColumnConfigs.value.forEach(col => col.visible = true)
  saveColumnSettings()
}

const showUserColumnDialog = ref(false)
const showRoleColumnDialog = ref(false)
const showPermColumnDialog = ref(false)

const showUserPermDetailDialog = ref(false)
const userPermDetailLoading = ref(false)
const userPermDetailTab = ref('effective')
const currentUserDetail = ref<any>(null)
const userPermDetailObjects = ref<any[]>([])
const userPermDetailSystem = ref<any[]>([])
const userEffectiveGrants = ref<any[]>([])
const userEffectiveGrantsLoading = ref(false)
const grantsCache = ref<Record<string, any[]>>({})
const currentGrantUser = ref<any>(null)

const showEditPermDialog = ref(false)
const editPermActiveTab = ref('object')
const editingPermUser = ref<any>(null)
const editPermLoading = ref(false)
const editPermObjTypes = ref<string[]>([])
const editPermDatabases = ref<string[]>([])
const editPermTables = ref<string[]>([])
const editPermColumns = ref<string[]>([])
const editPermSelectedDb = ref('')
const editPermSelectedTables = ref<string[]>([])
const editPermSelectedColumns = ref<string[]>([])
const editPermSelectedObjType = ref('table')
const editPermRules = ref<any[]>([])
const editPermSelectedPrivs = ref<string[]>([])
const editPermSystemPrivs = ref<any[]>([])
const editPermSelectedSystemPrivs = ref<string[]>([])
const editPermSystemPrivGroups = ref<any[]>([])
const editPermCascaderOptions = ref<any[]>([])
const editPermSelectedCascader = ref<any[]>([])
const editPermSelectedRules = ref<any[]>([])

const showUserRoleDialog = ref(false)
const currentRoleUser = ref<any>(null)
const availableRoleList = ref<any[]>([])
const selectedRoleIds = ref<string[]>([])

const userKeyword = ref('')
const userStatus = ref('')
const userList = ref<any[]>([])
const userLoading = ref(false)
const userPage = ref(1)
const userPageSize = ref(50)
const userTotal = ref(0)

const roleKeyword = ref('')
const roleList = ref<any[]>([])
const roleLoading = ref(false)
const rolePage = ref(1)
const rolePageSize = ref(50)
const roleTotal = ref(0)

const permPrincipalType = ref('')
const permCategory = ref('')
const permType = ref('')
const permLevel = ref('')
const permList = ref<any[]>([])
const permLoading = ref(false)
const permPage = ref(1)
const permPageSize = ref(50)
const permTotal = ref(0)
const selectedPerms = ref<any[]>([])

const auditOperator = ref('')
const auditOperType = ref('')
const auditResult = ref('')
const auditLogList = ref<any[]>([])
const auditLoading = ref(false)
const auditPage = ref(1)
const auditPageSize = ref(50)
const auditTotal = ref(0)

const showUserDialog = ref(false)
const editingUser = ref<any>(null)
const editUserActiveTab = ref('basic')
const availableRoles = ref<any[]>([])
const editingUserRoles = ref<string[]>([])
const editingUserObjectPermissions = ref<any[]>([])
const editingUserSystemPermissions = ref<any[]>([])
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
const editingRoleObjectPermissions = ref<any[]>([])
const editingRoleSystemPermissions = ref<any[]>([])
const roleForm = ref({
  name: '',
  description: ''
})
const roleFormRef = ref<any>(null)
const roleRules = {
  name: [{ required: true, message: '角色名称必填', trigger: 'blur' }]
}

const showGrantSubDialog = ref(false)
const grantSubForm = ref({
  privilegeType: 'readonly',
  objectLevel: 'database',
  databaseName: '',
  tableName: '',
  columns: [] as string[]
})
const editingPermDatabases = ref<string[]>([])
const editingPermObjects = ref<string[]>([])
const editingPermColumns = ref<string[]>([])
const grantSubPrincipalType = ref('user')
const grantSubPrincipalId = ref('')

const showSystemPrivDialog = ref(false)
const systemPrivileges = ref<any[]>([])
const systemPrivilegesLoading = ref(false)
const selectedSystemPrivileges = ref<string[]>([])
const systemPrivPrincipalType = ref('user')
const systemPrivPrincipalId = ref('')

const showGrantDialog = ref(false)
const grantForm = ref({
  principalType: 'user',
  principalId: '',
  grantType: 'object',
  privilegeType: 'readonly',
  objectLevel: 'database',
  databaseName: '',
  tableName: '',
  columns: [] as string[],
  systemPrivileges: [] as string[]
})

const databases = ref<string[]>([])
const objects = ref<string[]>([])
const columns = ref<string[]>([])

const grantSystemPrivileges = ref<any[]>([])
const grantSystemPrivilegesLoading = ref(false)

const principalOptions = computed(() => {
  if (grantForm.value.principalType === 'user') {
    return userList.value.map(u => ({ id: u.id, name: u.username }))
  }
  return roleList.value.map(r => ({ id: r.id, name: r.name }))
})

const groupedSystemPrivileges = computed(() => {
  const groups: Record<string, any[]> = {}
  systemPrivileges.value.forEach(priv => {
    if (!groups[priv.category]) {
      groups[priv.category] = []
    }
    groups[priv.category].push(priv)
  })
  return Object.keys(groups).map(category => ({
    category,
    privileges: groups[category]
  }))
})

const groupedGrantSystemPrivileges = computed(() => {
  const groups: Record<string, any[]> = {}
  grantSystemPrivileges.value.forEach(priv => {
    if (!groups[priv.category]) {
      groups[priv.category] = []
    }
    groups[priv.category].push(priv)
  })
  return Object.keys(groups).map(category => ({
    category,
    privileges: groups[category]
  }))
})

async function loadDatasources() {
  try {
    const res = await datasourceApi.listAllDatasources()
    datasourceList.value = res
  } catch {}
}

function handleDatasourceChange() {
  databases.value = []
  objects.value = []
  columns.value = []
  grantForm.value.databaseName = ''
  grantForm.value.tableName = ''
  grantForm.value.columns = []
  grantForm.value.systemPrivileges = []

  if (selectedDsId.value) {
    loadUsers()
    loadRoles()
    loadPermissions()
    datasourceApi.listDatabases(selectedDsId.value).then((res: any) => {
      databases.value = Array.isArray(res) ? res.map((d: any) => typeof d === 'object' && d.name ? d.name : d) : []
    }).catch(() => {})
  } else {
    userList.value = []
    userTotal.value = 0
    roleList.value = []
    roleTotal.value = 0
    permList.value = []
    permTotal.value = 0
    auditLogList.value = []
    auditTotal.value = 0
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

function refreshUsers() { loadUsers() }
function refreshRoles() { loadRoles() }
function refreshPermissions() { loadPermissions() }
function refreshAuditLogs() { loadAuditLogs() }

async function showUserPermissions(row: any) {
  currentUserDetail.value = row
  await loadUserPermDetail(row.id)
  await loadUserEffectiveGrants(row.id)
  showUserPermDetailDialog.value = true
}

async function refreshUserPermissions(row: any) {
  if (!selectedDsId.value || !row) return
  try {
    await loadUserEffectiveGrants(row.id, true)
    await loadUserPermDetail(row.id)
    ElMessage.success('权限数据已刷新')
  } catch (e: any) {
    ElMessage.error(e.message || '刷新失败')
  }
}

async function loadUserEffectiveGrants(userId: string, forceRefresh = false) {
  const cacheKey = `grants_${selectedDsId.value}_${userId}`
  if (!forceRefresh && grantsCache.value[cacheKey]) {
    userEffectiveGrants.value = grantsCache.value[cacheKey]
    return
  }

  userEffectiveGrantsLoading.value = true
  try {
    const res = await internalAuthApi.getUserEffectiveGrants(userId)
    userEffectiveGrants.value = res.grants || []
    grantsCache.value[cacheKey] = userEffectiveGrants.value
  } catch (e: any) {
    userEffectiveGrants.value = []
    const msg = e.message || '加载有效权限失败'
    if (msg.includes('password is empty') || msg.includes('decrypt')) {
      ElMessage.warning('数据源密码未配置或无效，无法获取数据库实时权限信息')
    } else {
      console.error('加载有效权限失败:', e)
    }
  }
  userEffectiveGrantsLoading.value = false
}

function editUserPermissions(row: any) {
  editingPermUser.value = row
  showEditPermDialog.value = true
}

function openUserRoleDialog(row: any) {
  currentRoleUser.value = row
  selectedRoleIds.value = []
  showUserRoleDialog.value = true
}

async function loadUserRoleDialogData() {
  if (!selectedDsId.value || !currentRoleUser.value) return
  try {
    const [rolesRes, userRolesRes] = await Promise.all([
      internalAuthApi.listInternalRoles({ datasourceId: selectedDsId.value, pageSize: 100 }),
      internalAuthApi.getUserRoles(currentRoleUser.value.id)
    ])
    availableRoleList.value = rolesRes.list || []
    selectedRoleIds.value = (userRolesRes || []).map((r: any) => r.id)
  } catch {
    availableRoleList.value = []
    selectedRoleIds.value = []
  }
}

function toggleRoleSelection(roleId: string, checked: boolean) {
  const idx = selectedRoleIds.value.indexOf(roleId)
  if (checked && idx === -1) {
    selectedRoleIds.value.push(roleId)
  } else if (!checked && idx > -1) {
    selectedRoleIds.value.splice(idx, 1)
  }
}

async function saveUserRoles() {
  if (!currentRoleUser.value) return
  try {
    await internalAuthApi.assignUserRoles(currentRoleUser.value.id, { roleIds: selectedRoleIds.value })
    ElMessage.success('角色分配成功')
    showUserRoleDialog.value = false
    loadUsers()
  } catch (e: any) {
    ElMessage.error(e.message || '分配失败')
  }
}

async function loadEditPermData() {
  if (!selectedDsId.value || !editingPermUser.value) return
  editPermLoading.value = true
  try {
    const [detailRes, dbRes, sysPrivsRes] = await Promise.all([
      internalAuthApi.getUserPermissionDetail({ datasourceId: selectedDsId.value, userId: editingPermUser.value.id }),
      datasourceApi.listDatabases(selectedDsId.value),
      internalAuthApi.getSystemPrivileges(selectedDsId.value)
    ])

    editPermDatabases.value = Array.isArray(dbRes) ? dbRes.map((d: any) => typeof d === 'object' && d.name ? d.name : d) : []
    editPermCascaderOptions.value = editPermDatabases.value.map(db => ({
      value: { type: 'database', name: db },
      label: db,
      children: []
    }))

    const objectPerms = detailRes?.objectPermissions || []
    editPermRules.value = objectPerms.map((p: any) => ({
      id: p.id,
      objectType: p.objectType,
      privilegeType: p.privilegeType,
      databaseName: p.databaseName,
      tableName: p.tableName || '',
      columns: p.columns ? (Array.isArray(p.columns) ? p.columns : JSON.parse(p.columns)) : [],
      scope: `${p.databaseName}${p.tableName ? '.' + p.tableName : ''}${p.columns ? '(' + (Array.isArray(p.columns) ? p.columns.join(',') : JSON.parse(p.columns).join(',')) + ')' : ''}`
    }))

    editPermSystemPrivs.value = sysPrivsRes || []
    editPermSelectedSystemPrivs.value = detailRes?.systemPermissions || []

    const groups: Record<string, any[]> = {}
    editPermSystemPrivs.value.forEach((priv: any) => {
      const category = priv.category || '其他'
      if (!groups[category]) {
        groups[category] = []
      }
      groups[category].push(priv)
    })
    editPermSystemPrivGroups.value = Object.keys(groups).map(category => ({
      category,
      privileges: groups[category],
      isTiDBExclusive: isTiDBExclusiveCategory(category)
    }))
  } catch (e) {
    console.error('加载权限数据失败:', e)
  }
  editPermLoading.value = false
}

function isTiDBExclusiveCategory(category: string): boolean {
  const tiDBExclusiveCategories = ['备份恢复', 'TiDB专属', 'PLACEMENT管理', 'Dashboard访问', '安全', 'SQL优化']
  return tiDBExclusiveCategories.includes(category) || 
         category.includes('ADMIN') || 
         category.includes('TiDB')
}

function resetEditPermForm() {
  editPermActiveTab.value = 'object'
  editPermSelectedDb.value = ''
  editPermSelectedTables.value = []
  editPermSelectedColumns.value = []
  editPermSelectedObjType.value = 'table'
  editPermSelectedPrivs.value = []
  editPermSelectedCascader.value = []
  editPermCascaderOptions.value = []
  editPermSelectedRules.value = []
  editPermTables.value = []
  editPermColumns.value = []
}

async function handleEditPermDbChange(dbName: string) {
  editPermSelectedTables.value = []
  editPermSelectedColumns.value = []
  editPermTables.value = []
  editPermColumns.value = []
  if (!dbName || !selectedDsId.value) return
  try {
    const res = await datasourceApi.listTables(selectedDsId.value, dbName)
    editPermTables.value = Array.isArray(res) ? res.map((t: any) => typeof t === 'object' && t.name ? t.name : t) : []
  } catch {}
}

async function handleEditPermTableChange(tables: string[]) {
  editPermSelectedColumns.value = []
  editPermColumns.value = []
  if (tables.length === 0 || !selectedDsId.value || !editPermSelectedDb.value) return
  try {
    const res = await listColumns(selectedDsId.value, editPermSelectedDb.value, tables[0])
    editPermColumns.value = Array.isArray(res) ? res.map((c: any) => c.name || c) : []
  } catch {}
}

function isPrivGroupAllSelected(group: string): boolean {
  const privs: Record<string, string[]> = {
    readonly: ['SELECT', 'SHOW VIEW'],
    dml: ['INSERT', 'UPDATE', 'DELETE', 'EXECUTE'],
    ddl: ['CREATE', 'ALTER', 'DROP', 'INDEX', 'TRIGGER', 'CREATE ROUTINE', 'ALTER ROUTINE']
  }
  const groupPrivs = privs[group] || []
  return groupPrivs.length > 0 && groupPrivs.every(p => editPermSelectedPrivs.value.includes(p))
}

function togglePrivItem(priv: string) {
  const idx = editPermSelectedPrivs.value.indexOf(priv)
  if (idx > -1) {
    editPermSelectedPrivs.value.splice(idx, 1)
  } else {
    editPermSelectedPrivs.value.push(priv)
  }
}

function togglePrivGroup(group: string) {
  const privs: Record<string, string[]> = {
    readonly: ['SELECT', 'SHOW VIEW'],
    dml: ['INSERT', 'UPDATE', 'DELETE', 'EXECUTE'],
    ddl: ['CREATE', 'ALTER', 'DROP', 'INDEX', 'TRIGGER', 'CREATE ROUTINE', 'ALTER ROUTINE']
  }
  const groupPrivs = privs[group] || []
  const allSelected = groupPrivs.every(p => editPermSelectedPrivs.value.includes(p))
  if (allSelected) {
    editPermSelectedPrivs.value = editPermSelectedPrivs.value.filter(p => !groupPrivs.includes(p))
  } else {
    groupPrivs.forEach(p => {
      if (!editPermSelectedPrivs.value.includes(p)) {
        editPermSelectedPrivs.value.push(p)
      }
    })
  }
}

function toggleSystemPrivGroup(category: string, checked: boolean) {
  const group = editPermSystemPrivGroups.value.find(g => g.category === category)
  if (!group) return
  group.privileges.forEach((priv: any) => {
    const idx = editPermSelectedSystemPrivs.value.indexOf(priv.name)
    if (checked && idx === -1) {
      editPermSelectedSystemPrivs.value.push(priv.name)
    } else if (!checked && idx > -1) {
      editPermSelectedSystemPrivs.value.splice(idx, 1)
    }
  })
}

function addEditPermRule() {
  if (editPermSelectedCascader.value.length === 0) {
    ElMessage.warning('请选择库/表/列')
    return
  }

  const privilegeType = determinePrivilegeType(editPermSelectedPrivs.value)
  if (!privilegeType) {
    ElMessage.warning('请至少选择一项权限')
    return
  }

  const dbMap = new Map<string, Map<string, string[]>>()
  editPermSelectedCascader.value.forEach(item => {
    if (item.type === 'column') {
      const tableMap = dbMap.get(item.databaseName) || new Map<string, string[]>()
      const cols = tableMap.get(item.tableName) || []
      cols.push(item.name)
      tableMap.set(item.tableName, cols)
      dbMap.set(item.databaseName, tableMap)
    } else if (item.type === 'table') {
      const tableMap = dbMap.get(item.databaseName) || new Map<string, string[]>()
      tableMap.set(item.name, [])
      dbMap.set(item.databaseName, tableMap)
    } else if (item.type === 'database') {
      dbMap.set(item.name, new Map<string, string[]>())
    }
  })

  dbMap.forEach((tableMap, dbName) => {
    if (tableMap.size === 0) {
      editPermRules.value.push({
        id: '',
        objectType: 'database',
        privilegeType,
        databaseName: dbName,
        tableName: '',
        columns: [],
        scope: dbName
      })
    } else {
      tableMap.forEach((cols, tableName) => {
        const scope = cols.length > 0 ? `${dbName}.${tableName}(${cols.join(',')})` : `${dbName}.${tableName}`
        editPermRules.value.push({
          id: '',
          objectType: cols.length > 0 ? 'column' : 'table',
          privilegeType,
          databaseName: dbName,
          tableName,
          columns: cols,
          scope
        })
      })
    }
  })

  editPermSelectedCascader.value = []
  editPermSelectedPrivs.value = []
  ElMessage.success('权限规则添加成功')
}

async function loadEditPermCascaderChildren(node: any, resolve: (data: any[]) => void) {
  if (!selectedDsId.value) return
  try {
    if (node.data.value.type === 'database') {
      const dbName = node.data.value.name
      const res = await datasourceApi.listTables(selectedDsId.value, dbName)
      const tables = Array.isArray(res) ? res.map((t: any) => typeof t === 'object' && t.name ? t.name : t) : []
      resolve(tables.map(table => ({
        value: { type: 'table', databaseName: dbName, name: table },
        label: table,
        children: []
      })))
    } else if (node.data.value.type === 'table') {
      const { databaseName, name: tableName } = node.data.value
      const res = await listColumns(selectedDsId.value, databaseName, tableName)
      const cols = Array.isArray(res) ? res.map((c: any) => c.name || c) : []
      resolve(cols.map(col => ({
        value: { type: 'column', databaseName, tableName, name: col },
        label: col,
        children: null
      })))
    }
  } catch {
    resolve([])
  }
}

function handleEditPermCascaderChange() {
}

function handleEditPermRuleSelectionChange(rows: any[]) {
  editPermSelectedRules.value = rows
}

function batchRemoveEditPermRules() {
  if (editPermSelectedRules.value.length === 0) return
  editPermSelectedRules.value.forEach(row => {
    removeEditPermRule(row)
  })
  editPermSelectedRules.value = []
  ElMessage.success('批量删除成功')
}

function determinePrivilegeType(privs: string[]): string {
  const ddlPrivs = ['CREATE', 'ALTER', 'DROP', 'INDEX', 'TRIGGER', 'CREATE ROUTINE', 'ALTER ROUTINE']
  const hasDDL = privs.some(p => ddlPrivs.includes(p))
  if (hasDDL) return 'ddl'

  const dmlPrivs = ['INSERT', 'UPDATE', 'DELETE', 'EXECUTE']
  const hasDML = privs.some(p => dmlPrivs.includes(p))
  if (hasDML) return 'dml'

  return 'readonly'
}

function removeEditPermRule(row: any) {
  const idx = editPermRules.value.findIndex(r => r.id === row.id && r.scope === row.scope)
  if (idx > -1) {
    editPermRules.value.splice(idx, 1)
  }
}

async function saveEditPermissions() {
  if (!selectedDsId.value || !editingPermUser.value) return

  if (editPermRules.value.length === 0 && editPermSelectedSystemPrivs.value.length === 0) {
    ElMessage.warning('请至少配置一项权限')
    return
  }

  try {
    await internalAuthApi.saveUserPermissions({
      datasourceId: selectedDsId.value,
      userId: editingPermUser.value.id,
      objectPermissions: editPermRules.value.map(r => ({
        objectType: r.objectType,
        privilegeType: r.privilegeType,
        databaseName: r.databaseName,
        tableName: r.tableName,
        columns: r.columns
      })),
      systemPermissions: editPermSelectedSystemPrivs.value
    })

    ElMessage.success('权限保存成功')
    showEditPermDialog.value = false
    loadPermissions()
    loadUsers()
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  }
}

async function loadUserPermDetail(userId: string) {
  userPermDetailLoading.value = true
  try {
    const res = await internalAuthApi.getUserPermissions(userId)
    const perms = res || []
    userPermDetailObjects.value = perms.filter((p: any) => !p.privilegeCategory || p.privilegeCategory === 'object')
    userPermDetailSystem.value = perms.filter((p: any) => p.privilegeCategory === 'system')
  } catch {
    userPermDetailObjects.value = []
    userPermDetailSystem.value = []
  }
  userPermDetailLoading.value = false
}

function formatDateTime(dateStr: any): string {
  if (!dateStr) return ''
  try {
    const date = new Date(dateStr)
    return date.toLocaleString('zh-CN', {
      year: 'numeric', month: '2-digit', day: '2-digit',
      hour: '2-digit', minute: '2-digit', second: '2-digit'
    })
  } catch {
    return String(dateStr)
  }
}

function getPrivilegeTagType(type: string): string {
  const types: Record<string, string> = { readonly: 'info', dml: 'warning', ddl: 'danger' }
  return types[type] || ''
}

function getPrivilegeLabel(type: string): string {
  const types: Record<string, string> = { readonly: '只读', dml: 'DML', ddl: 'DDL' }
  return types[type] || type
}

function getObjectLevelLabel(level: string): string {
  const levels: Record<string, string> = {
    database: '数据库', table: '表', column: '列', view: '视图',
    procedure: '存储过程', function: '函数', trigger: '触发器', event: '事件'
  }
  return levels[level] || level
}

function getOperTypeLabel(type: string): string {
  const types: Record<string, string> = {
    create_user: '创建用户', update_user: '修改用户', delete_user: '删除用户',
    reset_password: '重置密码', create_role: '创建角色', update_role: '修改角色',
    delete_role: '删除角色', grant_permission: '授权', revoke_permission: '回收权限'
  }
  return types[type] || type
}

function getPrivTagType(type: string) {
  switch (type) { case 'readonly': return 'info'; case 'dml': return 'warning'; case 'ddl': return 'danger'; default: return '' }
}

function getPrivLabel(type: string) {
  switch (type) { case 'readonly': return '只读'; case 'dml': return 'DML'; case 'ddl': return 'DDL'; default: return type }
}

function getLevelLabel(level: string) {
  switch (level) {
    case 'database': return '数据库'; case 'table': return '表'; case 'column': return '列';
    case 'view': return '视图'; case 'procedure': return '存储过程'; case 'function': return '函数';
    case 'trigger': return '触发器'; case 'event': return '事件'; default: return level
  }
}

function parseColumns(cols: any): string {
  if (!cols) return '-'
  try {
    const arr = JSON.parse(cols)
    return Array.isArray(arr) && arr.length > 0 ? arr.join(', ') : '-'
  } catch { return '-' }
}

function parseSystemPrivs(privs: any): string {
  if (!privs) return '-'
  try {
    const arr = JSON.parse(privs)
    return Array.isArray(arr) && arr.length > 0 ? arr.join(', ') : '-'
  } catch { return '-' }
}

function isUserInRole(roleId: string): boolean {
  return editingUserRoles.value.includes(roleId)
}

function toggleRoleAssignment(roleId: string) {
  const index = editingUserRoles.value.indexOf(roleId)
  if (index > -1) {
    editingUserRoles.value.splice(index, 1)
  } else {
    editingUserRoles.value.push(roleId)
  }
}

function resetPassword(row: any) {
  ElMessageBox.prompt('输入新密码（不填则自动生成）', '重置密码', {
    confirmButtonText: '确定', cancelButtonText: '取消'
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
  ElMessageBox.confirm('确定删除该用户？', '提示', { type: 'warning' }).then(() => {
    internalAuthApi.deleteInternalUser(row.id).then(() => {
      ElMessage.success('删除成功')
      loadUsers()
    }).catch(() => {})
  }).catch(() => {})
}

async function saveUser() {
  if (!userFormRef.value || !selectedDsId.value) {
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
  roleForm.value = { name: row.name, description: row.description || '' }
  editRoleActiveTab.value = 'basic'
  await loadEditingRolePermissions(row.id)
  showRoleDialog.value = true
}

async function loadEditingRolePermissions(roleId: string) {
  try {
    const res = await internalAuthApi.getRolePermissions(roleId)
    const perms = res || []
    editingRoleObjectPermissions.value = perms.filter((p: any) => p.privilegeCategory === 'object')
    editingRoleSystemPermissions.value = perms.filter((p: any) => p.privilegeCategory === 'system')
  } catch {
    editingRoleObjectPermissions.value = []
    editingRoleSystemPermissions.value = []
  }
}

function deleteRole(row: any) {
  ElMessageBox.confirm('确定删除该角色？', '提示', { type: 'warning' }).then(() => {
    internalAuthApi.deleteInternalRole(row.id).then(() => {
      ElMessage.success('删除成功')
      loadRoles()
    }).catch(() => {})
  }).catch(() => {})
}

async function saveRole() {
  if (!roleFormRef.value || !selectedDsId.value) {
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
      privilegeCategory: permCategory.value,
      privilegeType: permType.value,
      objectLevel: permLevel.value,
      page: permPage.value,
      pageSize: permPageSize.value
    })
    const list = res.list || []
    for (const rule of list) {
      if (rule.privilegeCategory === 'object') {
        let scope = rule.databaseName || ''
        if (rule.tableName) {
          scope += '.' + rule.tableName
        }
        if (rule.columns) {
          try {
            const cols = JSON.parse(rule.columns)
            if (Array.isArray(cols) && cols.length > 0) {
              scope += '(' + cols.join(',') + ')'
            }
          } catch {}
        }
        rule.scope = scope || '-'
      } else {
        rule.scope = '全局'
      }
    }
    permList.value = list
    permTotal.value = res.total || 0
  } catch {}
  permLoading.value = false
}

function handlePermSelectionChange(rows: any[]) {
  selectedPerms.value = rows
}

async function loadAuditLogs() {
  if (!selectedDsId.value) return
  auditLoading.value = true
  try {
    const res = await internalAuthApi.listAuditLogs({
      datasourceId: selectedDsId.value,
      operator: auditOperator.value,
      operType: auditOperType.value,
      result: auditResult.value,
      page: auditPage.value,
      pageSize: auditPageSize.value
    })
    auditLogList.value = res.list || []
    auditTotal.value = res.total || 0
  } catch {}
  auditLoading.value = false
}

function revokePerm(row: any) {
  ElMessageBox.confirm('确定回收该权限？', '提示', { type: 'warning' }).then(() => {
    if (row.privilegeCategory === 'system') {
      internalAuthApi.revokeSystemPrivileges(row.id).then(() => {
        ElMessage.success('回收成功')
        loadPermissions()
      }).catch(() => {})
    } else {
      internalAuthApi.revokePermission(row.id).then(() => {
        ElMessage.success('回收成功')
        loadPermissions()
      }).catch(() => {})
    }
  }).catch(() => {})
}

function batchRevokePerm() {
  if (selectedPerms.value.length === 0) return
  ElMessageBox.confirm(`确定回收选中的 ${selectedPerms.value.length} 条权限？`, '提示', { type: 'warning' }).then(() => {
    const ids = selectedPerms.value.map((p: any) => p.id)
    const systemIds = ids.filter((id: string) => {
      const perm = permList.value.find((p: any) => p.id === id)
      return perm && perm.privilegeCategory === 'system'
    })
    const objectIds = ids.filter((id: string) => !systemIds.includes(id))
    const promises: Promise<any>[] = []
    if (objectIds.length > 0) {
      promises.push(internalAuthApi.batchRevokePermission({ ids: objectIds }))
    }
    systemIds.forEach(id => promises.push(internalAuthApi.revokeSystemPrivileges(id)))
    Promise.all(promises).then(() => {
      ElMessage.success('批量回收成功')
      loadPermissions()
    }).catch(() => {})
  }).catch(() => {})
}

async function editUser(row: any) {
  editingUser.value = row
  userForm.value = {
    username: row.username,
    host: row.host || '',
    password: '',
    remark: row.remark || ''
  }
  editUserActiveTab.value = 'basic'
  await loadAvailableRoles()
  await loadEditingUserRoles(row.id)
  await loadEditingUserPermissions(row.id)
  showUserDialog.value = true
}

async function loadAvailableRoles() {
  if (!selectedDsId.value) return
  try {
    const res = await internalAuthApi.listInternalRoles({ datasourceId: selectedDsId.value, pageSize: 100 })
    availableRoles.value = res.list || []
  } catch { availableRoles.value = [] }
}

async function loadEditingUserRoles(userId: string) {
  try {
    const res = await internalAuthApi.getUserRoles(userId)
    editingUserRoles.value = res.map((r: any) => r.id)
  } catch { editingUserRoles.value = [] }
}

async function loadEditingUserPermissions(userId: string) {
  try {
    const res = await internalAuthApi.getUserPermissions(userId)
    const perms = res || []
    editingUserObjectPermissions.value = perms.filter((p: any) => !p.privilegeCategory || p.privilegeCategory === 'object')
    editingUserSystemPermissions.value = perms.filter((p: any) => p.privilegeCategory === 'system')
  } catch {
    editingUserObjectPermissions.value = []
    editingUserSystemPermissions.value = []
  }
}

async function onObjectLevelChange() {
  objects.value = []
  columns.value = []
  grantForm.value.tableName = ''
  grantForm.value.columns = []
}

async function onDatabaseChange() {
  objects.value = []
  columns.value = []
  grantForm.value.tableName = ''
  grantForm.value.columns = []
  if (!grantForm.value.databaseName || !selectedDsId.value) return
  try {
    const res = await internalAuthApi.listObjects(selectedDsId.value, grantForm.value.databaseName, grantForm.value.objectLevel)
    objects.value = res || []
  } catch {}
}

async function onTableChange() {
  columns.value = []
  grantForm.value.columns = []
  if (!grantForm.value.tableName || !selectedDsId.value) return
  try {
    const res = await internalAuthApi.listColumns(selectedDsId.value, grantForm.value.databaseName, grantForm.value.tableName)
    columns.value = res || []
  } catch {}
}

function openGrantSubDialog(type: string) {
  grantSubPrincipalType.value = type
  grantSubPrincipalId.value = type === 'user' ? editingUser.value?.id : editingRole.value?.id
  grantSubForm.value = {
    privilegeType: 'readonly',
    objectLevel: 'database',
    databaseName: '',
    tableName: '',
    columns: []
  }
  editingPermObjects.value = []
  editingPermColumns.value = []
  datasourceApi.listDatabases(selectedDsId.value).then((res: any) => {
    editingPermDatabases.value = Array.isArray(res) ? res.map((d: any) => typeof d === 'object' && d.name ? d.name : d) : []
  }).catch(() => {})
  showGrantSubDialog.value = true
}

function onGrantSubObjectLevelChange() {
  editingPermObjects.value = []
  editingPermColumns.value = []
  grantSubForm.value.tableName = ''
  grantSubForm.value.columns = []
}

async function onGrantSubDatabaseChange() {
  editingPermObjects.value = []
  editingPermColumns.value = []
  grantSubForm.value.tableName = ''
  grantSubForm.value.columns = []
  if (!grantSubForm.value.databaseName || !selectedDsId.value) return
  try {
    const res = await internalAuthApi.listObjects(selectedDsId.value, grantSubForm.value.databaseName, grantSubForm.value.objectLevel)
    editingPermObjects.value = res || []
  } catch {}
}

async function onGrantSubTableChange() {
  editingPermColumns.value = []
  grantSubForm.value.columns = []
  if (!grantSubForm.value.tableName || !selectedDsId.value) return
  try {
    const res = await internalAuthApi.listColumns(selectedDsId.value, grantSubForm.value.databaseName, grantSubForm.value.tableName)
    editingPermColumns.value = res || []
  } catch {}
}

async function submitGrantSub() {
  if (!grantSubForm.value.databaseName) {
    ElMessage.warning('请选择数据库')
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
    if (grantSubPrincipalType.value === 'user') {
      await loadEditingUserPermissions(grantSubPrincipalId.value)
    } else {
      await loadEditingRolePermissions(grantSubPrincipalId.value)
    }
  } catch (e: any) {
    ElMessage.error(e.message || '授权失败')
  }
}

function openSystemPrivDialog(type: string) {
  systemPrivPrincipalType.value = type
  systemPrivPrincipalId.value = type === 'user' ? editingUser.value?.id : editingRole.value?.id
  selectedSystemPrivileges.value = []
  loadSystemPrivileges()
  showSystemPrivDialog.value = true
}

async function loadSystemPrivileges() {
  systemPrivilegesLoading.value = true
  try {
    const res = await internalAuthApi.getSystemPrivileges(selectedDsId.value)
    systemPrivileges.value = res || []
    const currentPrivs = await internalAuthApi.getUserSystemPrivileges(systemPrivPrincipalId.value)
    selectedSystemPrivileges.value = currentPrivs || []
  } catch {
    systemPrivileges.value = []
  }
  systemPrivilegesLoading.value = false
}

function toggleSystemPrivilege(name: string) {
  const index = selectedSystemPrivileges.value.indexOf(name)
  if (index > -1) {
    selectedSystemPrivileges.value.splice(index, 1)
  } else {
    selectedSystemPrivileges.value.push(name)
  }
}

async function submitSystemPriv() {
  try {
    await internalAuthApi.grantSystemPrivileges({
      datasourceId: selectedDsId.value,
      principalType: systemPrivPrincipalType.value,
      principalId: systemPrivPrincipalId.value,
      systemPrivileges: selectedSystemPrivileges.value
    })
    showSystemPrivDialog.value = false
    ElMessage.success('配置成功')
    if (systemPrivPrincipalType.value === 'user') {
      await loadEditingUserPermissions(systemPrivPrincipalId.value)
    } else {
      await loadEditingRolePermissions(systemPrivPrincipalId.value)
    }
  } catch (e: any) {
    ElMessage.error(e.message || '配置失败')
  }
}

async function submitGrant() {
  if (!grantForm.value.principalId) {
    ElMessage.warning('请选择授权主体')
    return
  }
  try {
    if (grantForm.value.grantType === 'object') {
      if (!grantForm.value.databaseName) {
        ElMessage.warning('请选择数据库')
        return
      }
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
    } else {
      if (grantForm.value.systemPrivileges.length === 0) {
        ElMessage.warning('请选择系统权限')
        return
      }
      await internalAuthApi.grantSystemPrivileges({
        datasourceId: selectedDsId.value,
        principalType: grantForm.value.principalType,
        principalId: grantForm.value.principalId,
        systemPrivileges: grantForm.value.systemPrivileges
      })
    }
    showGrantDialog.value = false
    ElMessage.success('授权成功')
    loadPermissions()
  } catch (e: any) {
    ElMessage.error(e.message || '授权失败')
  }
}

async function revokePermissionDirect(perm: any) {
  try {
    await internalAuthApi.revokePermission(perm.id)
    ElMessage.success('回收成功')
    if (editingUser.value) {
      await loadEditingUserPermissions(editingUser.value.id)
    } else if (editingRole.value) {
      await loadEditingRolePermissions(editingRole.value.id)
    }
    if (showUserPermDetailDialog.value && currentUserDetail.value) {
      await loadUserPermDetail(currentUserDetail.value.id)
    }
  } catch (e: any) {
    ElMessage.error(e.message || '回收失败')
  }
}

async function revokeSystemPermissionDirect(perm: any) {
  try {
    await internalAuthApi.revokeSystemPrivileges(perm.id)
    ElMessage.success('回收成功')
    if (editingUser.value) {
      await loadEditingUserPermissions(editingUser.value.id)
    } else if (editingRole.value) {
      await loadEditingRolePermissions(editingRole.value.id)
    }
    if (showUserPermDetailDialog.value && currentUserDetail.value) {
      await loadUserPermDetail(currentUserDetail.value.id)
    }
  } catch (e: any) {
    ElMessage.error(e.message || '回收失败')
  }
}

watch(grantForm, (val) => {
  if (val.grantType === 'system') {
    loadGrantSystemPrivileges()
  }
}, { immediate: false })

async function loadGrantSystemPrivileges() {
  grantSystemPrivilegesLoading.value = true
  try {
    const res = await internalAuthApi.getSystemPrivileges(selectedDsId.value)
    grantSystemPrivileges.value = res || []
    const currentPrivs = await internalAuthApi.getUserSystemPrivileges(grantForm.value.principalId)
    grantForm.value.systemPrivileges = currentPrivs || []
  } catch {
    grantSystemPrivileges.value = []
  }
  grantSystemPrivilegesLoading.value = false
}

function toggleGrantSystemPrivilege(name: string) {
  const index = grantForm.value.systemPrivileges.indexOf(name)
  if (index > -1) {
    grantForm.value.systemPrivileges.splice(index, 1)
  } else {
    grantForm.value.systemPrivileges.push(name)
  }
}

onMounted(() => {
  loadDatasources()
})
</script>

<style scoped>
.page-container {
  padding: 8px;
  height: calc(100vh - 56px);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #fff;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  padding-bottom: 6px;
  border-bottom: 1px solid #f0f0f0;
}

.title-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.title-text {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
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
  min-height: 0;
  overflow: auto;
}

.toolbar {
  display: flex;
  gap: 6px;
  margin-bottom: 6px;
  flex-wrap: wrap;
  align-items: center;
  padding: 4px 0;
}

.toolbar-right {
  margin-left: auto;
}

.perm-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 6px;
}

.compact-dialog :deep(.el-dialog__body) {
  padding: 12px;
}

.compact-dialog :deep(.el-dialog__header) {
  padding: 10px 12px;
}

.compact-dialog :deep(.el-dialog__footer) {
  padding: 10px 12px;
}

.compact-dialog :deep(.el-form-item) {
  margin-bottom: 10px;
}

.permissions-loading {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 30px;
}

.permissions-empty {
  text-align: center;
  color: #909399;
  padding: 30px;
}

.grants-output {
  max-height: 350px;
  overflow-y: auto;
}

.sql-code {
  background-color: #1e1e1e;
  color: #d4d4d4;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  padding: 12px;
  border-radius: 4px;
  white-space: pre-wrap;
  word-break: break-all;
  line-height: 1.5;
}

.permission-panel {
  padding: 6px 0;
}

.permission-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  padding-bottom: 6px;
  border-bottom: 1px solid #f0f0f0;
}

.permission-title {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
}

.permission-list {
  min-height: 160px;
}

.permission-empty {
  text-align: center;
  color: #909399;
  padding: 30px;
  font-size: 13px;
}

.permission-items {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.permission-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 10px;
  background-color: #fafafa;
  border-radius: 3px;
  border: 1px solid #f0f0f0;
}

.perm-info {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
}

.perm-scope {
  font-size: 12px;
  color: #606266;
  font-weight: 500;
}

.perm-level {
  font-size: 11px;
  color: #909399;
  padding: 1px 5px;
  background-color: #f5f5f5;
  border-radius: 3px;
}

.auth-table {
  width: 100%;
  flex: 1;
  min-height: 280px;
}

.auth-table :deep(.el-table__header-wrapper),
.auth-table :deep(.el-table__body-wrapper) {
  overflow-x: auto;
}

.auth-table :deep(.el-table__cell) {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  padding: 6px 8px;
  font-size: 12px;
  line-height: 1.4;
}

.auth-table :deep(.el-table__header .el-table__cell) {
  white-space: nowrap;
  padding: 8px 8px;
  font-size: 12px;
  font-weight: 500;
}

.auth-table :deep(.el-table--border .el-table__cell) {
  border-right: 1px solid #f0f0f0;
}

.auth-table :deep(.el-table--border) {
  border: none;
}

.auth-table :deep(.el-table__body tr) {
  height: 36px;
}

.role-assign-panel {
  padding: 4px 0;
}

.role-assign-header {
  margin-bottom: 8px;
}

.role-assign-title {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
}

.role-assign-list {
  max-height: 280px;
  overflow-y: auto;
}

.role-assign-empty {
  text-align: center;
  color: #909399;
  padding: 20px;
  font-size: 13px;
}

.role-assign-items {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.role-assign-item {
  display: flex;
  align-items: center;
  padding: 6px 10px;
  background-color: #fafafa;
  border-radius: 3px;
}

.role-name {
  font-size: 12px;
  font-weight: 500;
  color: #303133;
  margin-right: 8px;
}

.role-desc {
  font-size: 11px;
  color: #909399;
}

.system-priv-group {
  margin-bottom: 12px;
}

.system-priv-group-title {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 6px;
  padding-left: 4px;
  border-left: 3px solid #409EFF;
}

.system-priv-items {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 4px 8px;
  background-color: #fafafa;
  border-radius: 4px;
}

.system-priv-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #606266;
}

.object-priv-section {
  margin-bottom: 12px;
}

.object-priv-section-title {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 8px;
}

.object-priv-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}

.object-priv-item {
  flex: 1;
  min-width: 200px;
}

.column-checkbox-item {
  display: flex;
  align-items: center;
  padding: 4px 0;
  font-size: 13px;
}

.pagination-wrapper {
  padding: 6px 0;
  display: flex;
  justify-content: flex-end;
}

.grant-list {
  max-height: 350px;
  overflow-y: auto;
  padding: 4px;
}

.grant-item {
  margin-bottom: 8px;
  padding: 8px 10px;
  background-color: #fafafa;
  border-radius: 4px;
  border: 1px solid #f0f0f0;
}

.grant-text {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  color: #606266;
  white-space: pre-wrap;
  word-break: break-all;
  line-height: 1.5;
  margin: 0;
}
</style>
