<template>
  <span class="db-icon" :title="label">
    <!-- MySQL Logo (橙色海豚风格) -->
    <svg v-if="type === 'mysql'" viewBox="0 0 32 32" width="18" height="18" xmlns="http://www.w3.org/2000/svg">
      <defs>
        <linearGradient id="mysqlgrad" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stop-color="#00758F"/>
          <stop offset="100%" stop-color="#F29111"/>
        </linearGradient>
      </defs>
      <circle cx="16" cy="16" r="15" fill="url(#mysqlgrad)" />
      <path d="M9 20 C 10 17, 14 15, 18 14 C 20 13.5, 22 14, 23 15 C 22 15.5, 21 16, 20 16 L 19 18 L 22 18 L 21 20 C 19 21, 15 22, 12 22 Z" fill="#ffffff"/>
      <circle cx="20" cy="15" r="1" fill="#ffffff"/>
      <text x="6" y="27" font-family="Arial, sans-serif" font-size="6" fill="#ffffff" font-weight="bold">MySQL</text>
    </svg>
    <!-- TiDB Logo (TiDB的蓝色闪电+叶子) -->
    <svg v-else-if="type === 'tidb'" viewBox="0 0 32 32" width="18" height="18" xmlns="http://www.w3.org/2000/svg">
      <rect width="32" height="32" rx="4" fill="#5C9EE8"/>
      <path d="M 12 8 L 18 16 L 14 16 L 19 24 L 13 18 L 17 18 Z" fill="#ffffff"/>
      <text x="5" y="30" font-family="Arial, sans-serif" font-size="5" fill="#ffffff" font-weight="bold">TiDB</text>
    </svg>
    <!-- SQLite Logo (白色铅笔+灰色) -->
    <svg v-else-if="type === 'sqlite'" viewBox="0 0 32 32" width="18" height="18" xmlns="http://www.w3.org/2000/svg">
      <circle cx="16" cy="16" r="15" fill="#003B57"/>
      <path d="M 10 11 L 21 22 L 19 24 L 8 13 Z" fill="#F29111"/>
      <path d="M 19 24 L 21 22 L 23 24 L 21 26 Z" fill="#cccccc"/>
      <text x="5" y="30" font-family="Arial, sans-serif" font-size="5" fill="#ffffff" font-weight="bold">SQLite</text>
    </svg>
    <!-- 默认：通用数据库图标 -->
    <svg v-else viewBox="0 0 32 32" width="18" height="18" xmlns="http://www.w3.org/2000/svg">
      <ellipse cx="16" cy="8" rx="12" ry="3" fill="#5C9EE8"/>
      <rect x="4" y="8" width="24" height="16" fill="#5C9EE8"/>
      <ellipse cx="16" cy="24" rx="12" ry="3" fill="#00758F"/>
      <line x1="4" y1="12" x2="28" y2="12" stroke="#ffffff" stroke-width="0.6"/>
      <line x1="4" y1="17" x2="28" y2="17" stroke="#ffffff" stroke-width="0.6"/>
    </svg>
  </span>
</template>

<script setup lang="ts">
defineProps<{ type: string; label?: string }>()
</script>

<style scoped>
.db-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  vertical-align: middle;
  line-height: 1;
}
</style>
