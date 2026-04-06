<template>
  <div
    class="pagination pagination__container flex items-center justify-between"
  >
    <div class="flex flex-1 items-center justify-between sm:hidden">
      <!-- Mobile pagination -->
      <button
        @click="goToPage(page - 1)"
        :disabled="page === 1"
        class="pagination__control pagination__control--mobile relative inline-flex items-center text-sm font-medium"
      >
        {{ t('pagination.previous') }}
      </button>
      <span class="pagination__meta text-sm">
        {{ t('pagination.pageOf', { page, total: totalPages }) }}
      </span>
      <button
        @click="goToPage(page + 1)"
        :disabled="page === totalPages"
        class="pagination__control pagination__control--mobile relative ml-3 inline-flex items-center text-sm font-medium"
      >
        {{ t('pagination.next') }}
      </button>
    </div>

    <div class="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
      <!-- Desktop pagination info -->
      <div class="flex flex-wrap items-center gap-x-4 gap-y-2">
        <p class="pagination__meta text-sm">
          {{ t('pagination.showing') }}
          <span class="font-medium">{{ fromItem }}</span>
          {{ t('pagination.to') }}
          <span class="font-medium">{{ toItem }}</span>
          {{ t('pagination.of') }}
          <span class="font-medium">{{ total }}</span>
          {{ t('pagination.results') }}
        </p>

        <!-- Page size selector (hidden on sm, visible from md) -->
        <div v-if="showPageSizeSelector" class="hidden md:flex items-center space-x-2">
          <span class="pagination__meta text-sm"
            >{{ t('pagination.perPage') }}:</span
          >
          <div class="page-size-select w-20">
            <Select
              :model-value="pageSize"
              :options="pageSizeSelectOptions"
              @update:model-value="handlePageSizeChange"
            />
          </div>
        </div>

        <!-- Jump to page (hidden on sm/md, visible from lg) -->
        <div v-if="showJump" class="hidden lg:flex items-center space-x-2">
          <span class="pagination__meta text-sm">{{ t('pagination.jumpTo') }}</span>
          <input
            v-model="jumpPage"
            type="number"
            min="1"
            :max="totalPages"
            class="input w-20 text-sm"
            :placeholder="t('pagination.jumpPlaceholder')"
            @keyup.enter="submitJump"
          />
          <button type="button" class="btn btn-ghost btn-sm" @click="submitJump">
            {{ t('pagination.jumpAction') }}
          </button>
        </div>
      </div>

      <!-- Desktop pagination buttons -->
      <nav
        class="pagination__nav relative z-0 inline-flex"
        aria-label="Pagination"
      >
        <!-- Previous button -->
        <button
          @click="goToPage(page - 1)"
          :disabled="page === 1"
          class="pagination__control pagination__control--edge pagination__control--prev relative inline-flex items-center text-sm font-medium"
          :aria-label="t('pagination.previous')"
        >
          <Icon name="chevronLeft" size="md" />
        </button>

        <!-- Page numbers -->
        <button
          v-for="(pageNum, index) in visiblePages"
          :key="`${pageNum}-${index}`"
          @click="typeof pageNum === 'number' && goToPage(pageNum)"
          :disabled="typeof pageNum !== 'number'"
          :class="[
            'pagination__page relative inline-flex items-center text-sm font-medium',
            pageNum === page
              ? 'pagination__page--active z-10'
              : 'pagination__page--idle',
            typeof pageNum !== 'number' && 'cursor-default'
          ]"
          :aria-label="
            typeof pageNum === 'number' ? t('pagination.goToPage', { page: pageNum }) : undefined
          "
          :aria-current="pageNum === page ? 'page' : undefined"
        >
          {{ pageNum }}
        </button>

        <!-- Next button -->
        <button
          @click="goToPage(page + 1)"
          :disabled="page === totalPages"
          class="pagination__control pagination__control--edge pagination__control--next relative inline-flex items-center text-sm font-medium"
          :aria-label="t('pagination.next')"
        >
          <Icon name="chevronRight" size="md" />
        </button>
      </nav>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import Select from './Select.vue'
import { setPersistedPageSize } from '@/composables/usePersistedPageSize'

const { t } = useI18n()

interface Props {
  total: number
  page: number
  pageSize: number
  pageSizeOptions?: number[]
  showPageSizeSelector?: boolean
  showJump?: boolean
}

interface Emits {
  (e: 'update:page', page: number): void
  (e: 'update:pageSize', pageSize: number): void
}

const props = withDefaults(defineProps<Props>(), {
  pageSizeOptions: () => [10, 20, 50, 100],
  showPageSizeSelector: true,
  showJump: false
})

const emit = defineEmits<Emits>()

const totalPages = computed(() => Math.ceil(props.total / props.pageSize))

const fromItem = computed(() => {
  if (props.total === 0) return 0
  return (props.page - 1) * props.pageSize + 1
})

const toItem = computed(() => {
  const to = props.page * props.pageSize
  return to > props.total ? props.total : to
})

const pageSizeSelectOptions = computed(() => {
  return props.pageSizeOptions.map((size) => ({
    value: size,
    label: String(size)
  }))
})

const jumpPage = ref('')

const visiblePages = computed(() => {
  const pages: (number | string)[] = []
  const maxVisible = 7
  const total = totalPages.value

  if (total <= maxVisible) {
    // Show all pages if total is small
    for (let i = 1; i <= total; i++) {
      pages.push(i)
    }
  } else {
    // Always show first page
    pages.push(1)

    const start = Math.max(2, props.page - 2)
    const end = Math.min(total - 1, props.page + 2)

    // Add ellipsis before if needed
    if (start > 2) {
      pages.push('...')
    }

    // Add middle pages
    for (let i = start; i <= end; i++) {
      pages.push(i)
    }

    // Add ellipsis after if needed
    if (end < total - 1) {
      pages.push('...')
    }

    // Always show last page
    pages.push(total)
  }

  return pages
})

const goToPage = (newPage: number) => {
  if (newPage >= 1 && newPage <= totalPages.value && newPage !== props.page) {
    emit('update:page', newPage)
  }
}

const handlePageSizeChange = (value: string | number | boolean | null) => {
  if (value === null || typeof value === 'boolean') return
  const newPageSize = typeof value === 'string' ? parseInt(value) : value
  setPersistedPageSize(newPageSize)
  emit('update:pageSize', newPageSize)
}

const submitJump = () => {
  const value = jumpPage.value.trim()
  if (!value) return
  const pageNum = Number.parseInt(value, 10)
  if (Number.isNaN(pageNum)) return
  const nextPage = Math.min(Math.max(pageNum, 1), totalPages.value)
  jumpPage.value = ''
  goToPage(nextPage)
}
</script>

<style scoped>
.pagination {
  border-top: 1px solid var(--theme-page-border);
  background: var(--theme-surface);
}

.pagination__container {
  padding: var(--theme-pagination-padding-y) var(--theme-pagination-padding-x);
}

@media (min-width: 640px) {
  .pagination__container {
    padding-inline: var(--theme-pagination-padding-x-sm);
  }
}

.pagination__meta {
  color: var(--theme-page-muted);
}

.pagination__nav {
  border-radius: calc(var(--theme-button-radius) + 2px);
  box-shadow: var(--theme-card-shadow);
}

.pagination__control,
.pagination__page {
  border: 1px solid var(--theme-card-border);
  background: var(--theme-button-secondary-bg);
  color: var(--theme-button-secondary-text);
  transition:
    background 0.2s ease,
    color 0.2s ease,
    border-color 0.2s ease,
    box-shadow 0.2s ease;
}

.pagination__control--mobile {
  padding: var(--theme-pagination-mobile-control-padding-y)
    var(--theme-pagination-mobile-control-padding-x);
}

.pagination__control--edge {
  padding: var(--theme-pagination-edge-control-padding-y)
    var(--theme-pagination-edge-control-padding-x);
}

.pagination__page {
  padding: var(--theme-pagination-page-padding-y) var(--theme-pagination-page-padding-x);
}

.pagination__control:hover,
.pagination__page--idle:hover {
  background: var(--theme-button-secondary-hover-bg);
}

.pagination__control--mobile,
.pagination__control--edge {
  border-radius: calc(var(--theme-button-radius) + 2px);
}

.pagination__control--prev {
  border-top-left-radius: calc(var(--theme-button-radius) + 2px);
  border-bottom-left-radius: calc(var(--theme-button-radius) + 2px);
}

.pagination__control--next {
  border-top-right-radius: calc(var(--theme-button-radius) + 2px);
  border-bottom-right-radius: calc(var(--theme-button-radius) + 2px);
}

.pagination__page + .pagination__page,
.pagination__page + .pagination__control,
.pagination__control + .pagination__page {
  margin-left: -1px;
}

.pagination__page--active {
  border-color: color-mix(in srgb, var(--theme-accent) 52%, var(--theme-card-border));
  background: color-mix(in srgb, var(--theme-accent-soft) 78%, var(--theme-surface));
  color: var(--theme-accent);
}

.page-size-select :deep(.select-trigger) {
  padding: var(--theme-pagination-page-size-trigger-padding-y)
    var(--theme-pagination-page-size-trigger-padding-x);
  font-size: 0.875rem;
}
</style>
