<template>
  <div class="space-y-6">
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4 lg:grid-cols-4">
      <div v-for="item in 4" :key="`core-${item}`" class="card dashboard-loading-skeleton__card">
        <div class="flex items-center gap-3">
          <div class="dashboard-loading-skeleton__block dashboard-loading-skeleton__icon animate-pulse"></div>
          <div class="flex-1 space-y-2">
            <div class="dashboard-loading-skeleton__block h-3 w-16 animate-pulse rounded"></div>
            <div class="dashboard-loading-skeleton__block h-6 w-24 animate-pulse rounded"></div>
          </div>
        </div>
      </div>
    </div>
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4 lg:grid-cols-4">
      <div v-for="item in 4" :key="`token-${item}`" class="card dashboard-loading-skeleton__card">
        <div class="flex items-center gap-3">
          <div class="dashboard-loading-skeleton__block dashboard-loading-skeleton__icon animate-pulse"></div>
          <div class="flex-1 space-y-2">
            <div class="dashboard-loading-skeleton__block h-3 w-16 animate-pulse rounded"></div>
            <div class="dashboard-loading-skeleton__block h-6 w-24 animate-pulse rounded"></div>
          </div>
        </div>
      </div>
    </div>
    <div class="card dashboard-loading-skeleton__chart-card">
      <div class="dashboard-loading-skeleton__block mb-4 h-5 w-40 animate-pulse rounded"></div>
      <div class="dashboard-loading-skeleton__block dashboard-loading-skeleton__chart animate-pulse"></div>
    </div>
  </div>
</template>

<style scoped>
.dashboard-loading-skeleton__card {
  padding: var(--theme-dashboard-skeleton-card-padding);
}

.dashboard-loading-skeleton__chart-card {
  padding: var(--theme-dashboard-skeleton-chart-padding);
}

.dashboard-loading-skeleton__icon {
  width: 2.5rem;
  height: 2.5rem;
  border-radius: var(--theme-dashboard-skeleton-icon-radius);
}

.dashboard-loading-skeleton__chart {
  height: 18rem;
  border-radius: var(--theme-dashboard-skeleton-chart-radius);
}

.dashboard-loading-skeleton__block {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}
</style>
