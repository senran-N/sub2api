<template>
  <div class="space-y-6">
    <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
      <div class="key-usage-loading-state__card">
        <div class="skeleton mb-6 h-5 w-24"></div>
        <div class="flex justify-center"><div class="skeleton h-44 w-44 rounded-full"></div></div>
      </div>
      <div class="key-usage-loading-state__card">
        <div class="skeleton mb-6 h-5 w-24"></div>
        <div class="flex justify-center"><div class="skeleton h-44 w-44 rounded-full"></div></div>
      </div>
    </div>
    <div class="key-usage-loading-state__card">
      <div class="skeleton mb-6 h-5 w-32"></div>
      <div class="space-y-4">
        <div class="skeleton h-4 w-full"></div>
        <div class="skeleton h-4 w-3/4"></div>
        <div class="skeleton h-4 w-5/6"></div>
        <div class="skeleton h-4 w-2/3"></div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.key-usage-loading-state__card {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 84%, transparent);
  border-radius: var(--theme-key-usage-card-radius);
  padding: var(--theme-key-usage-loading-card-padding);
  background: color-mix(in srgb, var(--theme-surface) 90%, transparent);
}
</style>
