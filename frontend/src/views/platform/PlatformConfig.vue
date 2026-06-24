<template>
  <div class="page-container">
    <div class="page-header">
      <div class="page-title">平台配置</div>
      <div class="page-sub">统一管理账号、部门、角色、菜单、字典与参数。</div>
    </div>

    <el-row :gutter="16" class="entry-grid">
      <el-col v-for="item in items" :key="item.path" :xs="12" :sm="12" :md="8" :lg="6">
        <el-card shadow="hover" class="entry-card" :body-style="{ padding: '20px' }" @click="go(item.path)">
          <div class="entry-icon" :style="{ background: item.color }">
            <el-icon :size="28" color="#fff"><component :is="item.icon" /></el-icon>
          </div>
          <div class="entry-title">{{ item.title }}</div>
          <div class="entry-desc">{{ item.desc }}</div>
          <el-button type="primary" plain style="margin-top:10px">进入</el-button>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

const items = [
  { title: '账号管理', desc: '管理平台用户账号、权限分配、状态控制与密码维护', icon: 'User', path: '/platform-config/account', color: '#409EFF' },
  { title: '部门管理', desc: '维护组织架构、部门层级与联系信息', icon: 'OfficeBuilding', path: '/platform-config/departments', color: '#67C23A' },
  { title: '角色管理', desc: '定义角色及其权限标识', icon: 'Avatar', path: '/platform-config/roles', color: '#E6A23C' },
  { title: '岗位管理', desc: '维护组织内岗位信息与排序', icon: 'Film', path: '/platform-config/posts', color: '#909399' },
  { title: '菜单管理', desc: '维护平台菜单项、路由与可见性', icon: 'Menu', path: '/platform-config/menus', color: '#67C23A' },
  { title: '字典管理', desc: '维护数据字典类型与字典项', icon: 'Collection', path: '/platform-config/dictionaries', color: '#8E44AD' },
  { title: '参数配置', desc: '调整系统参数、默认值与开关', icon: 'Setting', path: '/platform-config/parameters', color: '#606266' },
  { title: '系统配置', desc: '调整平台全局参数与功能开关', icon: 'Tools', path: '/platform-config/system-settings', color: '#909399' },
  { title: '介质维护', desc: '数据库安装介质、补丁与客户端', icon: 'Upload', path: '/platform-config/medium', color: '#409EFF' },
  { title: '插件管理', desc: '浏览与管理已安装的平台扩展', icon: 'MagicStick', path: '/platform-config/plugins', color: '#8E44AD' }
]

function go(path: string) {
  router.push(path)
}
</script>

<style scoped>
.entry-grid {
  margin-top: 4px;
}
.entry-card {
  cursor: pointer;
  transition: transform 0.15s ease, box-shadow 0.15s ease;
  margin-bottom: 16px;
  border-radius: 8px;
}
.entry-card:hover {
  transform: translateY(-2px);
}
.entry-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 12px;
}
.entry-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}
.entry-desc {
  font-size: 13px;
  color: #909399;
  margin-top: 4px;
  min-height: 36px;
}
</style>
