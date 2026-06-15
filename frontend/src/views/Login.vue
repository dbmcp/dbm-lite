<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DB老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-title">DBM-Lite</div>
      <div class="login-subtitle">轻量级数据库管控平台</div>
      <el-form :model="form" label-width="0" @submit.prevent="handleLogin">
        <el-form-item>
          <el-input v-model="form.username" placeholder="用户名" size="large" :prefix-icon="UserIcon" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.password" type="password" placeholder="密码" size="large" show-password :prefix-icon="LockIcon" @keyup.enter="handleLogin" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" size="large" class="login-btn" :loading="loading" @click="handleLogin">
            登录
          </el-button>
        </el-form-item>
      </el-form>
      <div style="text-align:center;color:#909399;font-size:12px;margin-top:16px;">
        默认账号: admin / admin123
      </div>
    </div>
    <div class="login-footer">
      © 2026 DBM-Lite v0.1.0 | 作者：DB老王 | 开源协议：Apache-2.0 / MulanPSL-2.0
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { User as UserIcon, Lock as LockIcon } from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const loading = ref(false)

const form = ref({
  username: 'admin',
  password: ''
})

async function handleLogin() {
  if (!form.value.username || !form.value.password) {
    ElMessage.warning('请输入用户名和密码')
    return
  }
  loading.value = true
  try {
    await userStore.login(form.value)
    ElMessage.success('登录成功')
    const redirect = (route.query.redirect as string) || localStorage.getItem('dbm-lite-last-page') || '/dashboard'
    router.push(redirect)
  } catch (e: any) {
    console.error('login failed:', e)
    if (e?.message) {
      ElMessage.error(e.message)
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-btn {
  width: 100%;
}
.login-footer {
  position: fixed;
  bottom: 12px;
  left: 0;
  right: 0;
  text-align: center;
  font-size: 12px;
  color: #909399;
  padding: 8px;
  background: rgba(255, 255, 255, 0.6);
}
</style>

