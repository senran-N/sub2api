<script setup lang="ts">
interface Props {
  fullscreen?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  fullscreen: false
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header (matches OpsDashboardHeader + overview blocks) -->
    <div :class="['ops-dashboard-skeleton__surface', props.fullscreen ? 'ops-dashboard-skeleton__surface--fullscreen' : 'ops-dashboard-skeleton__surface--default']">
      <div class="ops-dashboard-skeleton__divider flex flex-wrap items-center justify-between gap-4 pb-4">
        <div class="space-y-2">
          <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong h-6 w-44 animate-pulse rounded"></div>
          <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--soft h-3 w-80 animate-pulse rounded"></div>
        </div>
        <div v-if="!props.fullscreen" class="flex flex-wrap items-center gap-3">
          <div class="ops-dashboard-skeleton__toolbar-block ops-dashboard-skeleton__toolbar-block--sm ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong animate-pulse"></div>
          <div class="ops-dashboard-skeleton__toolbar-block ops-dashboard-skeleton__toolbar-block--md ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong animate-pulse"></div>
          <div class="ops-dashboard-skeleton__toolbar-block ops-dashboard-skeleton__toolbar-block--base ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong animate-pulse"></div>
          <div class="ops-dashboard-skeleton__toolbar-block ops-dashboard-skeleton__toolbar-block--icon ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong animate-pulse"></div>
          <div class="ops-dashboard-skeleton__toolbar-block ops-dashboard-skeleton__toolbar-block--pill ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong animate-pulse"></div>
          <div class="ops-dashboard-skeleton__toolbar-block ops-dashboard-skeleton__toolbar-block--pill ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong animate-pulse"></div>
          <div class="ops-dashboard-skeleton__toolbar-block ops-dashboard-skeleton__toolbar-block--icon ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong animate-pulse"></div>
        </div>
      </div>

      <div class="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-12">
        <div class="ops-dashboard-skeleton__panel lg:col-span-5">
          <div class="grid h-full grid-cols-1 gap-6 md:grid-cols-[200px_1fr] md:items-center">
            <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--soft ops-dashboard-skeleton__shape-card animate-pulse h-28"></div>
            <div class="space-y-4">
              <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong h-4 w-32 animate-pulse rounded"></div>
              <div class="grid grid-cols-2 gap-3">
                <div
                  v-for="i in 4"
                  :key="i"
                  class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--soft ops-dashboard-skeleton__shape-card h-14 animate-pulse"
                ></div>
              </div>
            </div>
          </div>
        </div>

        <div class="lg:col-span-7">
          <div class="grid h-full grid-cols-1 content-center gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <div
              v-for="i in 6"
              :key="i"
              class="ops-dashboard-skeleton__panel h-20 animate-pulse"
            ></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Row: Concurrency + Throughput (matches OpsDashboard.vue) -->
    <div class="grid grid-cols-1 gap-6 lg:grid-cols-4">
      <div :class="['ops-dashboard-skeleton__surface ops-dashboard-skeleton__surface--min-card lg:col-span-1', props.fullscreen ? 'ops-dashboard-skeleton__surface--fullscreen' : 'ops-dashboard-skeleton__surface--default']">
        <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong h-4 w-44 animate-pulse rounded"></div>
        <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--soft ops-dashboard-skeleton__shape-panel mt-6 h-72 animate-pulse"></div>
      </div>
      <div :class="['ops-dashboard-skeleton__surface ops-dashboard-skeleton__surface--min-card lg:col-span-1', props.fullscreen ? 'ops-dashboard-skeleton__surface--fullscreen' : 'ops-dashboard-skeleton__surface--default']">
        <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong h-4 w-44 animate-pulse rounded"></div>
        <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--soft ops-dashboard-skeleton__shape-panel mt-6 h-72 animate-pulse"></div>
      </div>
      <div :class="['ops-dashboard-skeleton__surface ops-dashboard-skeleton__surface--min-card lg:col-span-2', props.fullscreen ? 'ops-dashboard-skeleton__surface--fullscreen' : 'ops-dashboard-skeleton__surface--default']">
        <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong h-4 w-56 animate-pulse rounded"></div>
        <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--soft ops-dashboard-skeleton__shape-panel mt-6 h-72 animate-pulse"></div>
      </div>
    </div>

    <!-- Row: Visual Analysis (baseline 3-up grid) -->
    <div class="grid grid-cols-1 gap-6 md:grid-cols-3">
      <div
        v-for="i in 3"
        :key="i"
        :class="['ops-dashboard-skeleton__surface', props.fullscreen ? 'ops-dashboard-skeleton__surface--fullscreen' : 'ops-dashboard-skeleton__surface--default']"
      >
        <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong h-4 w-44 animate-pulse rounded"></div>
        <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--soft ops-dashboard-skeleton__shape-panel mt-6 h-56 animate-pulse"></div>
      </div>
    </div>

    <!-- Alert Events -->
    <div :class="['ops-dashboard-skeleton__surface', props.fullscreen ? 'ops-dashboard-skeleton__surface--fullscreen' : 'ops-dashboard-skeleton__surface--default']">
      <div class="flex flex-wrap items-center justify-between gap-4">
        <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong h-4 w-48 animate-pulse rounded"></div>
        <div v-if="!props.fullscreen" class="flex flex-wrap items-center gap-2">
          <div class="ops-dashboard-skeleton__toolbar-block ops-dashboard-skeleton__toolbar-block--sm ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong animate-pulse"></div>
          <div class="ops-dashboard-skeleton__toolbar-block ops-dashboard-skeleton__toolbar-block--xs ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong animate-pulse"></div>
          <div class="ops-dashboard-skeleton__toolbar-block ops-dashboard-skeleton__toolbar-block--xs ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong animate-pulse"></div>
        </div>
      </div>

      <div class="mt-6 space-y-3">
        <div v-for="i in 6" :key="i" class="ops-dashboard-skeleton__panel flex items-center justify-between gap-4">
          <div class="flex-1 space-y-2">
            <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong h-3 w-56 animate-pulse rounded"></div>
            <div class="ops-dashboard-skeleton__block ops-dashboard-skeleton__block--soft h-3 w-80 animate-pulse rounded"></div>
          </div>
          <div class="ops-dashboard-skeleton__toolbar-block ops-dashboard-skeleton__toolbar-block--status ops-dashboard-skeleton__block ops-dashboard-skeleton__block--strong animate-pulse"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
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
  border-bottom: 1px solid color-mix(in srgb, var(--theme-page-border) 78%, transparent);
}

.ops-dashboard-skeleton__panel {
  padding: var(--theme-ops-panel-padding);
  border-radius: var(--theme-select-panel-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 86%, var(--theme-surface));
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

.ops-dashboard-skeleton__shape-panel {
  border-radius: var(--theme-select-panel-radius);
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
</style>
