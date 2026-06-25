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
          <el-checkbox v-model="props.modelValue[col.key]" @change="onChange">
            {{ col.label }}
          </el-checkbox>
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { Setting as SettingIcon } from '@element-plus/icons-vue'

const props = defineProps<{
  columns: { key: string; label: string }[]
  modelValue: Record<string, boolean>
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', v: Record<string, boolean>): void
}>()

for (const c of props.columns) {
  if (props.modelValue[c.key] === undefined) {
    props.modelValue[c.key] = true
  }
}

watch(
  () => props.columns,
  (cols) => {
    for (const c of cols) {
      if (props.modelValue[c.key] === undefined) {
        props.modelValue[c.key] = true
      }
    }
  },
  { immediate: true }
)

const isAllSelected = computed(() => {
  return props.columns.every((c) => props.modelValue[c.key] !== false)
})
const isIndeterminate = computed(() => {
  const shown = props.columns.filter((c) => props.modelValue[c.key] !== false).length
  return shown > 0 && shown < props.columns.length
})

function toggleAll(val: boolean) {
  for (const c of props.columns) {
    props.modelValue[c.key] = val
  }
}

function onChange() {}

function onVisibleChange(_v: boolean) {}
</script>

<style scoped>
.col-toggle-item :deep(.el-checkbox) {
  width: 100%;
}
</style>
