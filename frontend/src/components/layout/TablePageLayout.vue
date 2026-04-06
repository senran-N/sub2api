<template>
  <div class="table-page-layout" :class="{ 'mobile-mode': isMobile }">
    <!-- 固定区域：操作按钮 -->
    <div v-if="$slots.actions" class="layout-section-fixed">
      <slot name="actions" />
    </div>

    <!-- 固定区域：搜索和过滤器 -->
    <div v-if="$slots.filters" class="layout-section-fixed">
      <slot name="filters" />
    </div>

    <!-- 滚动区域：表格 -->
    <div class="layout-section-scrollable">
      <div class="card table-scroll-container">
        <slot name="table" />
      </div>
    </div>

    <!-- 固定区域：分页器 -->
    <div v-if="$slots.pagination" class="layout-section-fixed">
      <slot name="pagination" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const isMobile = ref(false)

const checkMobile = () => {
  isMobile.value = window.innerWidth < 1024
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})
</script>

<style scoped>
/* 桌面端：Flexbox 布局 */
.table-page-layout {
  @apply flex flex-col;
  gap: var(--theme-table-layout-gap);
  height: calc(100vh - var(--theme-shell-header-height) - var(--theme-table-layout-height-offset-mobile));
}

@media (min-width: 768px) {
  .table-page-layout {
    height: calc(100vh - var(--theme-shell-header-height) - var(--theme-table-layout-height-offset-tablet));
  }
}

@media (min-width: 1024px) {
  .table-page-layout {
    gap: var(--theme-table-layout-gap-lg);
    height: calc(100vh - var(--theme-shell-header-height) - var(--theme-table-layout-height-offset-desktop));
  }
}

.layout-section-fixed {
  @apply flex-shrink-0;
}

.layout-section-scrollable {
  @apply flex-1 min-h-0 flex flex-col;
}

/* 表格滚动容器 - 增强版表体滚动方案 */
.table-scroll-container {
  @apply flex flex-col overflow-hidden h-full;
  background: var(--theme-surface);
  border: 1px solid var(--theme-card-border);
  border-radius: var(--theme-surface-radius);
  box-shadow: var(--theme-card-shadow);
}

.table-scroll-container :deep(.table-wrapper) {
  @apply flex-1 overflow-x-auto overflow-y-auto;
  /* 确保横向滚动条显示在最底部 */
  scrollbar-gutter: stable;
}

.table-scroll-container :deep(table) {
  @apply w-full;
  min-width: max-content; /* 关键：确保表格宽度根据内容撑开，从而触发横向滚动 */
  display: table; /* 使用标准 table 布局以支持 sticky 列 */
}

.table-scroll-container :deep(thead) {
  backdrop-filter: blur(var(--theme-table-head-blur));
  background: color-mix(in srgb, var(--theme-table-head-bg) 92%, transparent);
}

.table-scroll-container :deep(tbody) {
  /* 保持默认 table-row-group 显示，不使用 block */
}

.table-scroll-container :deep(th) {
  @apply text-left font-medium;
  padding: var(--theme-table-cell-padding-y) var(--theme-table-cell-padding-x);
  font-size: var(--theme-table-head-font-size);
  letter-spacing: var(--theme-table-head-letter-spacing);
  text-transform: var(--theme-table-head-text-transform);
  color: var(--theme-table-head-text);
  border-bottom: 1px solid var(--theme-page-border);
}

.table-scroll-container :deep(td) {
  @apply text-sm;
  padding: var(--theme-table-cell-padding-y) var(--theme-table-cell-padding-x);
  color: var(--theme-page-text);
  border-bottom: 1px solid color-mix(in srgb, var(--theme-page-border) 72%, transparent);
}

/* 移动端：恢复正常滚动 */
.table-page-layout.mobile-mode .table-scroll-container {
  @apply h-auto overflow-visible border-none shadow-none;
  background: transparent;
}

.table-page-layout.mobile-mode .layout-section-scrollable {
  @apply flex-none min-h-fit;
}

.table-page-layout.mobile-mode .table-scroll-container :deep(.table-wrapper) {
  @apply overflow-visible;
}

.table-page-layout.mobile-mode .table-scroll-container :deep(table) {
  @apply flex-none;
  display: table;
  min-width: 100%;
}
</style>
