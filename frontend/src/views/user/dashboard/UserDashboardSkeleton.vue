<template>
  <div class="user-dashboard-skeleton">
    <div class="user-dashboard-skeleton__stats-grid">
      <div v-for="index in 4" :key="index" class="card user-dashboard-skeleton__card">
        <div class="flex items-start gap-4">
          <div class="user-dashboard-skeleton__block user-dashboard-skeleton__icon-block animate-pulse"></div>
          <div class="flex-1 space-y-2">
            <div class="user-dashboard-skeleton__block h-4 w-20 animate-pulse rounded"></div>
            <div class="user-dashboard-skeleton__block h-7 w-28 animate-pulse rounded"></div>
          </div>
        </div>
      </div>
    </div>

    <div class="card user-dashboard-skeleton__section-card">
        <div class="user-dashboard-skeleton__section-header flex items-center justify-between">
          <div class="user-dashboard-skeleton__block h-5 w-32 animate-pulse rounded"></div>
        <div class="flex gap-2">
          <div class="user-dashboard-skeleton__block user-dashboard-skeleton__control-block animate-pulse"></div>
          <div class="user-dashboard-skeleton__block user-dashboard-skeleton__control-block animate-pulse"></div>
        </div>
      </div>
      <div class="user-dashboard-skeleton__block user-dashboard-skeleton__chart-block animate-pulse"></div>
    </div>

    <div class="user-dashboard-skeleton__detail-grid">
      <div class="card user-dashboard-skeleton__section-card lg:col-span-2">
        <div class="user-dashboard-skeleton__block user-dashboard-skeleton__section-title animate-pulse rounded"></div>
        <div class="space-y-3">
          <div v-for="index in 4" :key="index" class="flex items-center gap-4">
            <div class="user-dashboard-skeleton__block h-4 w-full animate-pulse rounded"></div>
          </div>
        </div>
      </div>

      <div class="card user-dashboard-skeleton__section-card lg:col-span-1">
        <div class="user-dashboard-skeleton__block user-dashboard-skeleton__section-title user-dashboard-skeleton__section-title--narrow animate-pulse rounded"></div>
        <div class="space-y-3">
          <div v-for="index in 3" :key="index" class="user-dashboard-skeleton__block user-dashboard-skeleton__list-item animate-pulse"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.user-dashboard-skeleton {
  display: flex;
  flex-direction: column;
  gap: var(--theme-table-layout-gap-lg);
}

.user-dashboard-skeleton__stats-grid {
  display: grid;
  grid-template-columns: repeat(1, minmax(0, 1fr));
  gap: var(--theme-table-layout-gap);
}

.user-dashboard-skeleton__detail-grid {
  display: grid;
  grid-template-columns: repeat(1, minmax(0, 1fr));
  gap: var(--theme-table-layout-gap-lg);
}

.user-dashboard-skeleton__card {
  padding: var(--theme-stat-card-padding);
}

.user-dashboard-skeleton__section-card {
  padding: calc(var(--theme-stat-card-padding) + 4px);
}

.user-dashboard-skeleton__section-header {
  margin-bottom: var(--theme-table-layout-gap);
}

.user-dashboard-skeleton__section-title {
  width: 10rem;
  height: 1.25rem;
  margin-bottom: var(--theme-table-layout-gap);
}

.user-dashboard-skeleton__section-title--narrow {
  width: 8rem;
}

.user-dashboard-skeleton__icon-block {
  width: var(--theme-stat-icon-size);
  height: var(--theme-stat-icon-size);
  border-radius: var(--theme-stat-icon-radius);
}

.user-dashboard-skeleton__control-block {
  width: 6rem;
  height: 2rem;
  border-radius: var(--theme-user-dashboard-skeleton-control-radius);
}

.user-dashboard-skeleton__chart-block {
  height: 16rem;
  border-radius: var(--theme-user-dashboard-skeleton-chart-radius);
}

.user-dashboard-skeleton__list-item {
  height: 2.5rem;
  border-radius: var(--theme-user-dashboard-skeleton-item-radius);
}

.user-dashboard-skeleton__block {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

@media (min-width: 640px) {
  .user-dashboard-skeleton__stats-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 1024px) {
  .user-dashboard-skeleton__stats-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .user-dashboard-skeleton__detail-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
</style>
