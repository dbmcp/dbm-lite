<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DBA老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <div class="page-container">
    <div class="page-header">个人中心</div>
    <div class="card" style="max-width:600px;">
      <el-form label-width="100px">
        <el-form-item label="用户名">
          <span>{{ userInfo?.username }}</span>
        </el-form-item>
        <el-form-item label="显示名">
          <span>{{ userInfo?.displayName }}</span>
        </el-form-item>
        <el-form-item label="邮箱">
          <span>{{ userInfo?.email || '-' }}</span>
        </el-form-item>
        <el-form-item label="角色">
          <el-tag v-if="userInfo?.role === 'admin'" type="danger">管理员</el-tag>
          <el-tag v-else>成员</el-tag>
        </el-form-item>
        <el-form-item label="创建时间">
          <span>{{ userInfo?.createdAt }}</span>
        </el-form-item>
      </el-form>

      <el-divider />

      <div style="font-weight:600;margin-bottom:12px;">修改密码</div>
      <el-form label-width="100px">
        <el-form-item label="原密码">
          <el-input v-model="oldPwd" type="password" show-password />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="newPwd" type="password" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="doChange">保存修改</el-button>
        </el-form-item>
      </el-form>

      <el-divider />

      <div style="font-weight:600;margin-bottom:12px;">关于本项目</div>
      <el-form label-width="100px">
        <el-form-item label="项目名称">
          <span>DBM-Lite 轻量级全域数据库管控平台</span>
        </el-form-item>
        <el-form-item label="版本号">
          <span>v0.1.0</span>
        </el-form-item>
        <el-form-item label="作者">
          <span>DBA老王</span>
        </el-form-item>
        <el-form-item label="开源协议">
          <span>Apache License 2.0 / 木兰宽松许可证第2版</span>
        </el-form-item>
        <el-form-item>
          <el-button type="success" @click="showAbout = true">查看完整版权声明</el-button>
        </el-form-item>
      </el-form>
    </div>

    <el-dialog v-model="showAbout" title="关于 DBM-Lite" width="520px">
      <div style="line-height:1.8;color:#303133;">
        <p style="margin:0 0 8px 0;"><b>项目全称：</b>DBM-Lite 轻量级全域数据库管控平台</p>
        <p style="margin:0 0 8px 0;"><b>版本号：</b>v0.1.0</p>
        <p style="margin:0 0 8px 0;"><b>作者：</b>DBA老王</p>
        <p style="margin:0 0 8px 0;"><b>协议：</b>本软件采用 <code>Apache-2.0 OR MulanPSL-2.0</code> 双许可模式，使用者可自由选择其中一种协议使用</p>
        <p style="margin:0 0 8px 0;"><b>技术栈：</b>Go + Gin + GORM + SQLite · Vue3 + Element Plus</p>
        <p style="margin:0 0 8px 0;"><b>发布日期：</b>2026</p>
        <el-divider style="margin:12px 0;" />
        <p style="margin:0;color:#909399;font-size:12px;">© 2026 DBA老王 版权所有</p>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getMe, changePassword, UserInfo } from '@/api/auth'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const userInfo = ref<UserInfo | null>(userStore.userInfo)
const oldPwd = ref('')
const newPwd = ref('')
const showAbout = ref(false)

async function loadMe() {
  try {
    const u: any = await getMe()
    userInfo.value = u
    userStore.setUserInfo(u)
  } catch (e: any) {
    // ignore
  }
}

async function doChange() {
  if (!oldPwd.value || newPwd.value.length < 6) {
    ElMessage.warning('请输入原密码和至少6位的新密码')
    return
  }
  try {
    await changePassword(oldPwd.value, newPwd.value)
    ElMessage.success('密码修改成功，请重新登录')
    oldPwd.value = ''
    newPwd.value = ''
    setTimeout(() => {
      userStore.logout()
      location.href = '#/login'
    }, 1000)
  } catch (e: any) {
    ElMessage.error(e?.message || '密码修改失败')
  }
}

onMounted(loadMe)
</script>

