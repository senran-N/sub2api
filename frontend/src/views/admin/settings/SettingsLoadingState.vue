<template>
  <div class="settings-loading-state__root flex items-center justify-center">
    <div class="settings-loading-state__spinner h-8 w-8 animate-spin rounded-full border-b-2"></div>
  </div>
</template>

<style scoped>
.settings-loading-state__spinner {
  border-color: transparent;
  border-bottom-color: var(--theme-accent);
}

.settings-loading-state__root {
  padding-block: var(--theme-settings-card-body-padding);
}
</style>
