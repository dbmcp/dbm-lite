<template>
  <el-dialog v-model="dialogVisible" title="编辑数据源" width="720px" :close-on-click-modal="false" @closed="$emit('close')">
    <el-form :model="form" label-width="110px" ref="formRef">
      <div class="form-section-title">基本信息</div>
      <el-form-item label="数据源名称">
        <el-input v-model="form.name" />
      </el-form-item>
      <el-form-item label="数据库类型">
        <el-radio-group v-model="form.dbType" @change="onDbTypeChange">
          <el-radio-button value="mysql">🗄️ MySQL</el-radio-button>
          <el-radio-button value="tidb">⚡ TiDB</el-radio-button>
          <el-radio-button value="sqlite">📁 SQLite</el-radio-button>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="颜色标签">
        <div class="color-picker">
          <div v-for="c in colorOptions" :key="c.value" class="color-option"
            :class="{ active: form.colorLabel === c.value }"
            :style="{ background: c.color }"
            @click="form.colorLabel = c.value">
            <span v-if="form.colorLabel === c.value" class="check">✓</span>
          </div>
        </div>
      </el-form-item>
      <el-form-item label="环境">
        <el-select v-model="form.env" style="width:100%">
          <el-option label="生产" value="prod" />
          <el-option label="预发" value="stage" />
          <el-option label="测试" value="test" />
          <el-option label="开发" value="dev" />
        </el-select>
      </el-form-item>

      <div class="form-section-title">连接配置</div>
      <template v-if="form.dbType !== 'sqlite'">
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="主机/IP">
              <el-input v-model="form.host" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="端口">
              <el-input-number v-model="form.port" :min="1" :max="65535" style="width:100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="用户名">
              <el-input v-model="form.username" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="密码">
              <el-input v-model="form.password" type="password" show-password placeholder="不修改请留空" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="初始数据库">
              <el-input v-model="form.defaultDatabase" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="字符集">
              <el-select v-model="form.charset" style="width:100%" placeholder="utf8mb4">
                <el-option label="utf8mb4" value="utf8mb4" />
                <el-option label="utf8" value="utf8" />
                <el-option label="latin1" value="latin1" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item>
          <el-checkbox v-model="form.sslMode" :true-value="'true'" :false-value="'false'">启用 SSL 连接</el-checkbox>
          <el-checkbox v-model="form.readOnly" style="margin-left:16px">只读模式</el-checkbox>
        </el-form-item>
      </template>
      <template v-if="form.dbType === 'sqlite'">
        <el-form-item label="数据库文件路径">
          <el-input v-model="form.filePath" placeholder="如 test.db 或 /data/test.db；留空使用内存库" />
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="form.readOnly">只读模式</el-checkbox>
        </el-form-item>
        <el-form-item label="读写模式">
          <el-select v-model="form.openMode" style="width:100%">
            <el-option label="读写（rw）" value="rw" />
            <el-option label="只读（ro）" value="ro" />
          </el-select>
        </el-form-item>
      </template>

      <div class="form-section-title">组织与描述</div>
      <el-form-item label="标签">
        <el-input v-model="form.tags" placeholder="逗号分隔" />
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="form.remark" type="textarea" :rows="2" />
      </el-form-item>
    </el-form>
    <div class="dialog-footer">
      <span></span>
      <div>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save" :loading="saving">保存</el-button>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { getDatasource, updateDatasource } from '@/api/datasource'

const props = defineProps<{ datasourceId: string }>()
const emit = defineEmits<{
  (e: 'close'): void
  (e: 'saved'): void
}>()

const dialogVisible = ref(true)
const saving = ref(false)
const formRef = ref()

const colorOptions = [
  { value: 'blue', color: '#409eff' },
  { value: 'green', color: '#67c23a' },
  { value: 'red', color: '#f56c6c' },
  { value: 'yellow', color: '#e6a23c' },
  { value: 'purple', color: '#8e44ad' },
  { value: 'orange', color: '#e67e22' },
  { value: 'gray', color: '#909399' }
]

const form = reactive<any>({
  name: '', dbType: 'mysql', host: '', port: 3306, username: '', password: '',
  defaultDatabase: '', filePath: '', openMode: 'rw', charset: 'utf8mb4',
  timezone: 'Local', sslMode: 'false', readOnly: false, colorLabel: 'blue',
  tags: '', env: 'dev', remark: ''
})

watch(dialogVisible, (v) => {
  if (v && props.datasourceId) load()
})

onMounted(() => {
  if (props.datasourceId) load()
})

async function load() {
  try {
    const ds: any = await getDatasource(props.datasourceId)
    Object.assign(form, {
      name: ds.name, dbType: ds.dbType || 'mysql', host: ds.host || '',
      port: ds.port || 3306, username: ds.username || '', password: '',
      defaultDatabase: ds.defaultDatabase || '', filePath: ds.filePath || '',
      openMode: ds.openMode || 'rw', charset: ds.charset || 'utf8mb4',
      timezone: ds.timezone || 'Local', sslMode: ds.sslMode || 'false',
      readOnly: !!ds.readOnly, colorLabel: ds.colorLabel || 'blue',
      tags: ds.tags || '', env: ds.env || 'dev', remark: ds.remark || ''
    })
  } catch (e: any) {
    ElMessage.error(e?.message || '加载数据源失败')
  }
}

function onDbTypeChange(type: string) {
  if (type === 'mysql') {
    form.port = 3306
  } else if (type === 'tidb') {
    form.port = 4000
  } else if (type === 'sqlite') {
    form.defaultDatabase = 'main'
    form.host = ''
    form.username = ''
    form.port = 0
  }
}

async function save() {
  saving.value = true
  try {
    await updateDatasource(props.datasourceId, form)
    ElMessage.success('更新成功')
    dialogVisible.value = false
    emit('saved')
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.form-section-title {
  width: 100%;
  padding: 6px 0 8px;
  font-size: 13px;
  font-weight: 600;
  color: #303133;
  border-left: 3px solid #409eff;
  padding-left: 10px;
  background: #f4f8ff;
  margin: 6px 0 8px 0;
}
.dialog-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 8px;
}
.color-picker { display: flex; gap: 10px; align-items: center; padding: 4px 0; }
.color-option {
  width: 28px; height: 28px; border-radius: 50%; cursor: pointer;
  border: 2px solid transparent; transition: all .2s;
  display: flex; align-items: center; justify-content: center;
  color: #fff; font-size: 14px; font-weight: bold;
}
.color-option:hover { transform: scale(1.1); }
.color-option.active { border-color: #303133; box-shadow: 0 0 0 2px rgba(64,158,255,.18); }
</style>
