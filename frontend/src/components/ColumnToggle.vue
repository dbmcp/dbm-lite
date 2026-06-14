<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DB老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <el-dropdown trigger="click" @visible-change="onVisibleChange" :teleported="false">
    <el-button :icon="SettingIcon">列选择</el-button>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item disabled style="font-weight:600;color:#606266;">
          选择需要显示的列
        </el-dropdown-item>
        <el-dropdown-item divided>
          <el-checkbox
            v-model="isAllSelected"
            :indeterminate="isIndeterminate"
            @change="toggleAll"
            style="width:100%;"
          >全选 / 反选</el-checkbox>
        </el-dropdown-item>
        <el-dropdown-item
          v-for="col in columns"
          :key="col.key"
          class="col-toggle-item"
        >
          <el-checkbox v-model="localVisible[col.key]" @change="onChange">
            {{ col.label }}
          </el-checkbox>
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup lang="ts">
import { reactive, computed, watch } from 'vue'
import { Setting as SettingIcon } from '@element-plus/icons-vue'

const props = defineProps<{
  columns: { key: string; label: string }[]
  modelValue: Record<string, boolean>
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', v: Record<string, boolean>): void
}>()

const localVisible = reactive<Record<string, boolean>>({})

for (const c of props.columns) {
  localVisible[c.key] = props.modelValue?.[c.key] !== false
}

watch(
  () => props.modelValue,
  (val) => {
    for (const c of props.columns) {
      localVisible[c.key] = val?.[c.key] !== false
    }
  },
  { deep: true }
)

const isAllSelected = computed(() => {
  return props.columns.every((c) => localVisible[c.key] !== false)
})
const isIndeterminate = computed(() => {
  const shown = props.columns.filter((c) => localVisible[c.key] !== false).length
  return shown > 0 && shown < props.columns.length
})

function toggleAll(val: boolean) {
  for (const c of props.columns) {
    localVisible[c.key] = val
  }
  emit('update:modelValue', { ...localVisible })
}

function onChange() {
  emit('update:modelValue', { ...localVisible })
}

function onVisibleChange(_v: boolean) {}
</script>

<style scoped>
.col-toggle-item :deep(.el-checkbox) {
  width: 100%;
}
</style>
