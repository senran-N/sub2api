<script setup lang="ts">
interface Props {
  fullscreen?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  fullscreen: false
})
</script>

<template>
  <div class="ops-dashboard-skeleton">
    <!-- Header (matches OpsDashboardHeader + overview blocks) -->
    <div :class="['ops-dashboard-skeleton__surface', props.fullscreen ? 'ops-dashboard-skeleton__surface--fullscreen' : 'ops-dashboard-skeleton__surface--default']">
      <div class="ops-dashboard-skeleton__divider">
        <div class="ops-dashboard-skeleton__header-copy">
          <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong ops-dashboard-skeleton__block--title animate-pulse"></div>
          <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--soft ops-dashboard-skeleton__block--subtitle animate-pulse"></div>
        </div>
        <div v-if="!props.fullscreen" class="ops-dashboard-skeleton__toolbar">
          <div class="ops-dashboard-skeleton__toolbar-block ops-dashboard-skeleton__toolbar-block--sm ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong animate-pulse"></div>
          <div class="ops-dashboard-skeleton__toolbar-block ops-dashboard-skeleton__toolbar-block--md ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong animate-pulse"></div>
          <div class="ops-dashboard-skeleton__toolbar-block ops-dashboard-skeleton__toolbar-block--base ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong animate-pulse"></div>
          <div class="ops-dashboard-skeleton__toolbar-block ops-dashboard-skeleton__toolbar-block--icon ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong animate-pulse"></div>
          <div class="ops-dashboard-skeleton__toolbar-block ops-dashboard-skeleton__toolbar-block--pill ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong animate-pulse"></div>
          <div class="ops-dashboard-skeleton__toolbar-block ops-dashboard-skeleton__toolbar-block--pill ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong animate-pulse"></div>
          <div class="ops-dashboard-skeleton__toolbar-block ops-dashboard-skeleton__toolbar-block--icon ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong animate-pulse"></div>
        </div>
      </div>

      <div class="ops-dashboard-skeleton__overview-grid">
        <div class="ops-dashboard-skeleton__panel ops-dashboard-skeleton__panel--summary">
          <div class="ops-dashboard-skeleton__summary-layout">
            <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--soft ops-dashboard-skeleton__shape-card ops-dashboard-skeleton__shape-card--health animate-pulse"></div>
            <div class="ops-dashboard-skeleton__summary-metrics">
              <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong ops-dashboard-skeleton__block--metric-title animate-pulse"></div>
              <div class="ops-dashboard-skeleton__metric-grid">
                <div
                  v-for="i in 4"
                  :key="i"
                  class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--soft ops-dashboard-skeleton__shape-card ops-dashboard-skeleton__shape-card--metric animate-pulse"
                ></div>
              </div>
            </div>
          </div>
        </div>

        <div class="ops-dashboard-skeleton__overview-cards">
          <div class="ops-dashboard-skeleton__card-grid">
            <div
              v-for="i in 6"
              :key="i"
              class="ops-dashboard-skeleton__panel ops-dashboard-skeleton__panel--stat animate-pulse"
            ></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Row: Concurrency + Throughput (matches OpsDashboard.vue) -->
    <div class="ops-dashboard-skeleton__grid ops-dashboard-skeleton__grid--overview">
      <div :class="['ops-dashboard-skeleton__surface ops-dashboard-skeleton__surface--min-card ops-dashboard-skeleton__surface--single', props.fullscreen ? 'ops-dashboard-skeleton__surface--fullscreen' : 'ops-dashboard-skeleton__surface--default']">
        <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong ops-dashboard-skeleton__block--card-title animate-pulse"></div>
        <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--soft ops-dashboard-skeleton__shape-panel ops-dashboard-skeleton__shape-panel--large animate-pulse"></div>
      </div>
      <div :class="['ops-dashboard-skeleton__surface ops-dashboard-skeleton__surface--min-card ops-dashboard-skeleton__surface--single', props.fullscreen ? 'ops-dashboard-skeleton__surface--fullscreen' : 'ops-dashboard-skeleton__surface--default']">
        <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong ops-dashboard-skeleton__block--card-title animate-pulse"></div>
        <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--soft ops-dashboard-skeleton__shape-panel ops-dashboard-skeleton__shape-panel--large animate-pulse"></div>
      </div>
      <div :class="['ops-dashboard-skeleton__surface ops-dashboard-skeleton__surface--min-card ops-dashboard-skeleton__surface--double', props.fullscreen ? 'ops-dashboard-skeleton__surface--fullscreen' : 'ops-dashboard-skeleton__surface--default']">
        <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong ops-dashboard-skeleton__block--card-title ops-dashboard-skeleton__block--card-title-wide animate-pulse"></div>
        <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--soft ops-dashboard-skeleton__shape-panel ops-dashboard-skeleton__shape-panel--large animate-pulse"></div>
      </div>
    </div>

    <!-- Row: Visual Analysis (baseline 3-up grid) -->
    <div class="ops-dashboard-skeleton__grid ops-dashboard-skeleton__grid--analytics">
      <div
        v-for="i in 3"
        :key="i"
        :class="['ops-dashboard-skeleton__surface', props.fullscreen ? 'ops-dashboard-skeleton__surface--fullscreen' : 'ops-dashboard-skeleton__surface--default']"
      >
        <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong ops-dashboard-skeleton__block--card-title animate-pulse"></div>
        <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--soft ops-dashboard-skeleton__shape-panel ops-dashboard-skeleton__shape-panel--medium animate-pulse"></div>
      </div>
    </div>

    <!-- Alert Events -->
    <div :class="['ops-dashboard-skeleton__surface', props.fullscreen ? 'ops-dashboard-skeleton__surface--fullscreen' : 'ops-dashboard-skeleton__surface--default']">
      <div class="ops-dashboard-skeleton__events-head">
        <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong ops-dashboard-skeleton__block--events-title animate-pulse"></div>
        <div v-if="!props.fullscreen" class="ops-dashboard-skeleton__events-filters">
          <div class="ops-dashboard-skeleton__toolbar-block ops-dashboard-skeleton__toolbar-block--sm ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong animate-pulse"></div>
          <div class="ops-dashboard-skeleton__toolbar-block ops-dashboard-skeleton__toolbar-block--xs ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong animate-pulse"></div>
          <div class="ops-dashboard-skeleton__toolbar-block ops-dashboard-skeleton__toolbar-block--xs ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong animate-pulse"></div>
        </div>
      </div>

      <div class="ops-dashboard-skeleton__events-list">
        <div v-for="i in 6" :key="i" class="ops-dashboard-skeleton__panel ops-dashboard-skeleton__panel--event">
          <div class="ops-dashboard-skeleton__event-copy">
            <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong ops-dashboard-skeleton__block--event-title animate-pulse"></div>
            <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--soft ops-dashboard-skeleton__block--event-subtitle animate-pulse"></div>
          </div>
          <div class="ops-dashboard-skeleton__toolbar-block ops-dashboard-skeleton__toolbar-block--status ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong animate-pulse"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ops-dashboard-skeleton {
  display: flex;
  flex-direction: column;
  gap: var(--theme-ops-dashboard-layout-gap);
}

.ops-dashboard-skeleton__surface {
  border-radius: calc(var(--theme-surface-radius) * 2);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
}

.ops-dashboard-skeleton__surface--default {
  padding: calc(var(--theme-ops-card-padding) * 1.33);
}

.ops-dashboard-skeleton__surface--fullscreen {
  padding: calc(var(--theme-ops-card-padding) * 1.78);
}

.ops-dashboard-skeleton__surface--min-card {
  min-height: calc(var(--theme-ops-table-max-height) * 0.6);
}

.ops-dashboard-skeleton__divider {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--theme-ops-dashboard-header-gap);
  padding-bottom: var(--theme-ops-dashboard-header-gap);
  border-bottom: 1px solid color-mix(in srgb, var(--theme-page-border) 78%, transparent);
}

.ops-dashboard-skeleton__header-copy,
.ops-dashboard-skeleton__summary-metrics,
.ops-dashboard-skeleton__event-copy {
  display: flex;
  flex-direction: column;
}

.ops-dashboard-skeleton__header-copy,
.ops-dashboard-skeleton__event-copy {
  gap: calc(var(--theme-ops-dashboard-header-control-gap) * 0.5);
}

.ops-dashboard-skeleton__toolbar,
.ops-dashboard-skeleton__events-head,
.ops-dashboard-skeleton__events-filters,
.ops-dashboard-skeleton__panel--event {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
}

.ops-dashboard-skeleton__toolbar,
.ops-dashboard-skeleton__events-filters {
  gap: var(--theme-ops-dashboard-header-control-gap);
}

.ops-dashboard-skeleton__events-head,
.ops-dashboard-skeleton__panel--event {
  justify-content: space-between;
  gap: var(--theme-ops-dashboard-header-gap);
}

.ops-dashboard-skeleton__panel {
  padding: var(--theme-ops-panel-padding);
  border-radius: var(--theme-select-panel-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 86%, var(--theme-surface));
}

.ops-dashboard-skeleton__panel--summary {
  grid-column: span 1;
}

.ops-dashboard-skeleton__panel--stat {
  height: 5rem;
}

.ops-dashboard-skeleton__block--strong {
  background: color-mix(in srgb, var(--theme-page-border) 78%, transparent);
}

.ops-dashboard-skeleton__block--soft {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-page-border));
}

.ops-dashboard-skeleton__shape-card {
  border-radius: var(--theme-select-panel-radius);
}

.ops-dashboard-skeleton__shape-card--health {
  height: 7rem;
}

.ops-dashboard-skeleton__shape-card--metric {
  height: 3.5rem;
}

.ops-dashboard-skeleton__shape-panel {
  border-radius: var(--theme-select-panel-radius);
}

.ops-dashboard-skeleton__shape-panel--large {
  margin-top: var(--theme-ops-dashboard-layout-gap);
  height: 18rem;
}

.ops-dashboard-skeleton__shape-panel--medium {
  margin-top: var(--theme-ops-dashboard-layout-gap);
  height: 14rem;
}

.ops-dashboard-skeleton__toolbar-block {
  height: calc(var(--theme-button-padding-y) * 0.9 + 1.5rem);
  border-radius: var(--theme-select-panel-radius);
}

.ops-dashboard-skeleton__toolbar-block--icon {
  width: calc(var(--theme-button-padding-y) * 0.9 + 1.5rem);
}

.ops-dashboard-skeleton__toolbar-block--xs {
  width: 7.5rem;
}

.ops-dashboard-skeleton__toolbar-block--sm {
  width: 8.75rem;
}

.ops-dashboard-skeleton__toolbar-block--base {
  width: 9.375rem;
}

.ops-dashboard-skeleton__toolbar-block--md {
  width: 10rem;
}

.ops-dashboard-skeleton__toolbar-block--pill {
  width: 7rem;
}

.ops-dashboard-skeleton__toolbar-block--status {
  width: 5rem;
  height: 1.75rem;
}

.ops-dashboard-skeleton__block {
  border-radius: var(--theme-button-radius);
}

.ops-dashboard-skeleton__block--title {
  width: 11rem;
  height: 1.5rem;
}

.ops-dashboard-skeleton__block--subtitle {
  width: 20rem;
  height: 0.75rem;
}

.ops-dashboard-skeleton__block--metric-title {
  width: 8rem;
  height: 1rem;
}

.ops-dashboard-skeleton__block--card-title {
  width: 11rem;
  height: 1rem;
}

.ops-dashboard-skeleton__block--card-title-wide {
  width: 14rem;
}

.ops-dashboard-skeleton__block--events-title {
  width: 12rem;
  height: 1rem;
}

.ops-dashboard-skeleton__block--event-title {
  width: 14rem;
  height: 0.75rem;
}

.ops-dashboard-skeleton__block--event-subtitle {
  width: 20rem;
  height: 0.75rem;
}

.ops-dashboard-skeleton__overview-grid,
.ops-dashboard-skeleton__grid,
.ops-dashboard-skeleton__summary-layout,
.ops-dashboard-skeleton__metric-grid,
.ops-dashboard-skeleton__card-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: var(--theme-ops-dashboard-layout-gap);
}

.ops-dashboard-skeleton__overview-grid {
  margin-top: var(--theme-ops-dashboard-layout-gap);
}

.ops-dashboard-skeleton__summary-metrics {
  gap: var(--theme-ops-dashboard-header-gap);
}

.ops-dashboard-skeleton__metric-grid {
  gap: var(--theme-ops-dashboard-header-control-gap);
}

.ops-dashboard-skeleton__card-grid {
  height: 100%;
  align-content: center;
  gap: var(--theme-ops-dashboard-header-gap);
}

.ops-dashboard-skeleton__events-list {
  display: flex;
  flex-direction: column;
  gap: var(--theme-ops-dashboard-header-control-gap);
  margin-top: var(--theme-ops-dashboard-layout-gap);
}

@media (min-width: 640px) {
  .ops-dashboard-skeleton__card-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 768px) {
  .ops-dashboard-skeleton__grid--analytics {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .ops-dashboard-skeleton__summary-layout {
    grid-template-columns: 200px minmax(0, 1fr);
    align-items: center;
  }

  .ops-dashboard-skeleton__metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 1024px) {
  .ops-dashboard-skeleton__overview-grid {
    grid-template-columns: repeat(12, minmax(0, 1fr));
  }

  .ops-dashboard-skeleton__panel--summary {
    grid-column: span 5;
  }

  .ops-dashboard-skeleton__overview-cards {
    grid-column: span 7;
  }

  .ops-dashboard-skeleton__card-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .ops-dashboard-skeleton__grid--overview {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .ops-dashboard-skeleton__surface--single {
    grid-column: span 1;
  }

  .ops-dashboard-skeleton__surface--double {
    grid-column: span 2;
  }
}
</style>
