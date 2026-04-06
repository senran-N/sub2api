<template>
  <div class="data-table-mobile lg:hidden">
    <template v-if="loading">
      <div v-for="i in 5" :key="i" class="data-table-mobile__card">
        <div class="space-y-3">
          <div v-for="column in dataColumns" :key="column.key" class="flex justify-between">
            <div class="data-table-mobile__skeleton data-table-mobile__skeleton--label"></div>
            <div class="data-table-mobile__skeleton data-table-mobile__skeleton--value"></div>
          </div>
          <div v-if="hasActionsColumn" class="data-table-mobile__actions">
            <div class="data-table-mobile__skeleton data-table-mobile__skeleton--action"></div>
          </div>
        </div>
      </div>
    </template>

    <template v-else-if="!data || data.length === 0">
      <div class="data-table-mobile__empty">
        <slot name="empty">
          <div class="flex flex-col items-center">
            <Icon
              name="inbox"
              size="xl"
              class="data-table-mobile__empty-icon"
            />
            <p class="data-table-mobile__empty-title">
              {{ t('empty.noData') }}
            </p>
          </div>
        </slot>
      </div>
    </template>

    <template v-else>
      <div
        v-for="(row, index) in sortedData"
        :key="resolveRowKey(row, index)"
        class="data-table-mobile__card"
      >
        <div class="space-y-3">
          <div
            v-for="column in dataColumns"
            :key="column.key"
            class="flex items-start justify-between gap-4"
          >
            <span class="data-table-mobile__label">
              {{ column.label }}
            </span>
            <div class="data-table-mobile__value">
              <slot :name="`cell-${column.key}`" :row="row" :value="row[column.key]" :expanded="actionsExpanded">
                {{ column.formatter ? column.formatter(row[column.key], row) : row[column.key] }}
              </slot>
            </div>
          </div>
          <div v-if="hasActionsColumn" class="data-table-mobile__actions">
            <div class="data-table-mobile__actions-scroll overflow-x-auto scrollbar-hide">
              <div class="data-table-mobile__actions-row">
                <slot name="cell-actions" :row="row" :value="row['actions']" :expanded="actionsExpanded"></slot>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>

  <div
    ref="tableWrapperRef"
    class="table-wrapper hidden lg:block"
    :class="{
      'actions-expanded': actionsExpanded,
      'is-scrollable': isScrollable
    }"
    :style="tableLayoutVars"
  >
    <table class="data-table-desktop">
      <thead class="table-header">
        <tr>
          <th
            v-for="(column, index) in columns"
            :key="column.key"
            scope="col"
            :class="[
              'data-table-desktop__head-cell sticky-header-cell text-left text-xs font-medium uppercase tracking-wider',
              { 'data-table-desktop__head-cell--sortable cursor-pointer': column.sortable },
              getStickyColumnClass(column, index),
              column.class
            ]"
            @click="column.sortable && handleSort(column.key)"
          >
            <slot
              :name="`header-${column.key}`"
              :column="column"
              :sort-key="sortKey"
              :sort-order="sortOrder"
            >
              <div class="flex items-center space-x-1">
                <span>{{ column.label }}</span>
                <span v-if="column.sortable" class="data-table-desktop__sort-indicator">
                  <svg
                    v-if="sortKey === column.key"
                    class="h-4 w-4"
                    :class="{ 'rotate-180 transform': sortOrder === 'desc' }"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fill-rule="evenodd"
                      d="M14.707 12.707a1 1 0 01-1.414 0L10 9.414l-3.293 3.293a1 1 0 01-1.414-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 010 1.414z"
                      clip-rule="evenodd"
                    />
                  </svg>
                  <svg v-else class="h-4 w-4" fill="currentColor" viewBox="0 0 20 20">
                    <path
                      d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
                    />
                  </svg>
                </span>
              </div>
            </slot>
          </th>
        </tr>
      </thead>
      <tbody class="table-body data-table-desktop__body">
        <!-- Loading skeleton -->
        <tr v-if="loading" v-for="i in 5" :key="i">
          <td
            v-for="column in columns"
            :key="column.key"
            :class="['data-table-desktop__cell data-table-desktop__cell--loading whitespace-nowrap']"
          >
            <div class="animate-pulse">
              <div class="data-table-mobile__skeleton"></div>
            </div>
          </td>
        </tr>

        <!-- Empty state -->
        <tr v-else-if="!data || data.length === 0">
          <td
            :colspan="columns.length"
            :class="['data-table-desktop__empty data-table-desktop__empty-cell']"
          >
            <slot name="empty">
              <div class="flex flex-col items-center">
                <Icon
                  name="inbox"
                  size="xl"
                  class="data-table-mobile__empty-icon"
                />
                <p class="data-table-mobile__empty-title">
                  {{ t('empty.noData') }}
                </p>
              </div>
            </slot>
          </td>
        </tr>

        <!-- Data rows (virtual scroll) -->
        <template v-else>
          <tr v-if="virtualPaddingTop > 0" aria-hidden="true">
            <td :colspan="columns.length"
                :style="{ height: virtualPaddingTop + 'px', padding: 0, border: 'none' }">
            </td>
          </tr>
          <tr
            v-for="virtualRow in virtualItems"
            :key="resolveRowKey(sortedData[virtualRow.index], virtualRow.index)"
            :data-row-id="resolveRowKey(sortedData[virtualRow.index], virtualRow.index)"
            :data-index="virtualRow.index"
            :ref="measureElement"
            class="data-table-desktop__row"
          >
            <td
              v-for="(column, colIndex) in columns"
              :key="column.key"
              :class="[
                'data-table-desktop__cell whitespace-nowrap text-sm',
                getStickyColumnClass(column, colIndex),
                column.class
              ]"
            >
              <slot :name="`cell-${column.key}`"
                    :row="sortedData[virtualRow.index]"
                    :value="sortedData[virtualRow.index][column.key]"
                    :expanded="actionsExpanded">
                {{ column.formatter
                   ? column.formatter(sortedData[virtualRow.index][column.key], sortedData[virtualRow.index])
                   : sortedData[virtualRow.index][column.key] }}
              </slot>
            </td>
          </tr>
          <tr v-if="virtualPaddingBottom > 0" aria-hidden="true">
            <td :colspan="columns.length"
                :style="{ height: virtualPaddingBottom + 'px', padding: 0, border: 'none' }">
            </td>
          </tr>
        </template>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useVirtualizer } from '@tanstack/vue-virtual'
import { useI18n } from 'vue-i18n'
import type { Column } from './types'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const emit = defineEmits<{
  sort: [key: string, order: 'asc' | 'desc']
}>()

// 表格容器引用
const tableWrapperRef = ref<HTMLElement | null>(null)
const isScrollable = ref(false)
const actionsColumnNeedsExpanding = ref(false)

// 检查是否可滚动
const checkScrollable = () => {
  if (tableWrapperRef.value) {
    isScrollable.value = tableWrapperRef.value.scrollWidth > tableWrapperRef.value.clientWidth
  }
}

// 检查操作列是否需要展开
const checkActionsColumnWidth = () => {
  if (!tableWrapperRef.value) return

  // 查找第一行的操作列单元格
  const firstActionCell = tableWrapperRef.value.querySelector('tbody tr:first-child td:last-child')
  if (!firstActionCell) return

  // 查找操作列内容的容器div
  const actionsContainer = firstActionCell.querySelector('div')
  if (!actionsContainer) return

  // 临时展开以测量完整宽度
  const wasExpanded = actionsExpanded.value
  actionsExpanded.value = true

  // 等待DOM更新
  nextTick(() => {
    // 测量所有按钮的总宽度
    const actionItems = actionsContainer.querySelectorAll('button, a, [role="button"]')
    if (actionItems.length <= 2) {
      actionsColumnNeedsExpanding.value = false
      actionsExpanded.value = wasExpanded
      return
    }

    // 计算所有按钮的总宽度（包括gap）
    let totalWidth = 0
    actionItems.forEach((item, index) => {
      totalWidth += (item as HTMLElement).offsetWidth
      if (index < actionItems.length - 1) {
        totalWidth += 4 // gap-1 = 4px
      }
    })

    // 获取单元格可用宽度（减去padding）
    const cellWidth = (firstActionCell as HTMLElement).clientWidth - 32 // 减去左右padding

    // 如果总宽度超过可用宽度，需要展开功能
    actionsColumnNeedsExpanding.value = totalWidth > cellWidth

    // 恢复原来的展开状态
    actionsExpanded.value = wasExpanded
  })
}

// 监听尺寸变化
let resizeObserver: ResizeObserver | null = null
let resizeHandler: (() => void) | null = null

onMounted(() => {
  checkScrollable()
  checkActionsColumnWidth()
  if (tableWrapperRef.value && typeof ResizeObserver !== 'undefined') {
    resizeObserver = new ResizeObserver(() => {
      checkScrollable()
      checkActionsColumnWidth()
    })
    resizeObserver.observe(tableWrapperRef.value)
  } else {
    // 降级方案：不支持 ResizeObserver 时使用 window resize
    resizeHandler = () => {
      checkScrollable()
      checkActionsColumnWidth()
    }
    window.addEventListener('resize', resizeHandler)
  }
})

onUnmounted(() => {
  resizeObserver?.disconnect()
  if (resizeHandler) {
    window.removeEventListener('resize', resizeHandler)
    resizeHandler = null
  }
})

interface Props {
  columns: Column[]
  data: any[]
  loading?: boolean
  stickyFirstColumn?: boolean
  stickyActionsColumn?: boolean
  expandableActions?: boolean
  actionsCount?: number // 操作按钮总数，用于判断是否需要展开功能
  rowKey?: string | ((row: any) => string | number)
  /**
   * Default sort configuration (only applied when there is no persisted sort state)
   */
  defaultSortKey?: string
  defaultSortOrder?: 'asc' | 'desc'
  /**
   * Persist sort state (key + order) to localStorage using this key.
   * If provided, DataTable will load the stored sort state on mount.
   */
  sortStorageKey?: string
  /**
   * Enable server-side sorting mode. When true, clicking sort headers
   * will emit 'sort' events instead of performing client-side sorting.
   */
  serverSideSort?: boolean
  /** Estimated row height in px for the virtualizer (default 56) */
  estimateRowHeight?: number
  /** Number of rows to render beyond the visible area (default 5) */
  overscan?: number
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  stickyFirstColumn: true,
  stickyActionsColumn: true,
  expandableActions: true,
  defaultSortOrder: 'asc',
  serverSideSort: false
})

const sortKey = ref<string>('')
const sortOrder = ref<'asc' | 'desc'>('asc')
const actionsExpanded = ref(false)

type PersistedSortState = {
  key: string
  order: 'asc' | 'desc'
}

const collator = new Intl.Collator(undefined, {
  numeric: true,
  sensitivity: 'base'
})

const getSortableKeys = () => {
  const keys = new Set<string>()
  for (const col of props.columns) {
    if (col.sortable) keys.add(col.key)
  }
  return keys
}

const normalizeSortKey = (candidate: string) => {
  if (!candidate) return ''
  const sortableKeys = getSortableKeys()
  return sortableKeys.has(candidate) ? candidate : ''
}

const normalizeSortOrder = (candidate: any): 'asc' | 'desc' => {
  return candidate === 'desc' ? 'desc' : 'asc'
}

const readPersistedSortState = (): PersistedSortState | null => {
  if (!props.sortStorageKey) return null
  try {
    const raw = localStorage.getItem(props.sortStorageKey)
    if (!raw) return null
    const parsed = JSON.parse(raw) as Partial<PersistedSortState>
    const key = normalizeSortKey(typeof parsed.key === 'string' ? parsed.key : '')
    if (!key) return null
    return { key, order: normalizeSortOrder(parsed.order) }
  } catch (e) {
    console.error('[DataTable] Failed to read persisted sort state:', e)
    return null
  }
}

const writePersistedSortState = (state: PersistedSortState) => {
  if (!props.sortStorageKey) return
  try {
    localStorage.setItem(props.sortStorageKey, JSON.stringify(state))
  } catch (e) {
    console.error('[DataTable] Failed to persist sort state:', e)
  }
}

const resolveInitialSortState = (): PersistedSortState | null => {
  const persisted = readPersistedSortState()
  if (persisted) return persisted

  const key = normalizeSortKey(props.defaultSortKey || '')
  if (!key) return null
  return { key, order: normalizeSortOrder(props.defaultSortOrder) }
}

const applySortState = (state: PersistedSortState | null) => {
  if (!state) return
  sortKey.value = state.key
  sortOrder.value = state.order
}

const isNullishOrEmpty = (value: any) => value === null || value === undefined || value === ''

const toFiniteNumberOrNull = (value: any): number | null => {
  if (typeof value === 'number') return Number.isFinite(value) ? value : null
  if (typeof value === 'boolean') return value ? 1 : 0
  if (typeof value === 'string') {
    const trimmed = value.trim()
    if (!trimmed) return null
    const n = Number(trimmed)
    return Number.isFinite(n) ? n : null
  }
  return null
}

const toSortableString = (value: any): string => {
  if (value === null || value === undefined) return ''
  if (typeof value === 'string') return value
  if (typeof value === 'number' || typeof value === 'boolean') return String(value)
  if (value instanceof Date) return value.toISOString()
  try {
    return JSON.stringify(value)
  } catch {
    return String(value)
  }
}

const compareSortValues = (a: any, b: any): number => {
  const aEmpty = isNullishOrEmpty(a)
  const bEmpty = isNullishOrEmpty(b)
  if (aEmpty && bEmpty) return 0
  if (aEmpty) return 1
  if (bEmpty) return -1

  const aNum = toFiniteNumberOrNull(a)
  const bNum = toFiniteNumberOrNull(b)
  if (aNum !== null && bNum !== null) {
    if (aNum === bNum) return 0
    return aNum < bNum ? -1 : 1
  }

  const aStr = toSortableString(a)
  const bStr = toSortableString(b)
  const res = collator.compare(aStr, bStr)
  if (res === 0) return 0
  return res < 0 ? -1 : 1
}
const resolveRowKey = (row: any, index: number) => {
  if (typeof props.rowKey === 'function') {
    const key = props.rowKey(row)
    return key ?? index
  }
  if (typeof props.rowKey === 'string' && props.rowKey) {
    const key = row?.[props.rowKey]
    return key ?? index
  }
  const key = row?.id
  return key ?? index
}

const dataColumns = computed(() => props.columns.filter((column) => column.key !== 'actions'))
const columnsSignature = computed(() =>
  props.columns.map((column) => `${column.key}:${column.sortable ? '1' : '0'}`).join('|')
)

// 数据/列变化时重新检查滚动状态
// 注意：不能监听 actionsExpanded，因为 checkActionsColumnWidth 会临时修改它，会导致无限循环
watch(
  [() => props.data.length, columnsSignature],
  async () => {
    await nextTick()
    checkScrollable()
    checkActionsColumnWidth()
  },
  { flush: 'post' }
)

// 单独监听展开状态变化，只更新滚动状态
watch(actionsExpanded, async () => {
  await nextTick()
  checkScrollable()
})

const handleSort = (key: string) => {
  let newOrder: 'asc' | 'desc' = 'asc'
  if (sortKey.value === key) {
    newOrder = sortOrder.value === 'asc' ? 'desc' : 'asc'
  }

  if (props.serverSideSort) {
    // Server-side sort mode: emit event and update internal state for UI feedback
    sortKey.value = key
    sortOrder.value = newOrder
    emit('sort', key, newOrder)
  } else {
    // Client-side sort mode: just update internal state
    sortKey.value = key
    sortOrder.value = newOrder
  }
}

const sortedData = computed(() => {
  // Server-side sort mode: return data as-is (server handles sorting)
  if (props.serverSideSort || !sortKey.value || !props.data) return props.data

  const key = sortKey.value
  const order = sortOrder.value

  // Stable sort (tie-break with original index) to avoid jitter when values are equal.
  return props.data
    .map((row, index) => ({ row, index }))
    .sort((a, b) => {
      const cmp = compareSortValues(a.row?.[key], b.row?.[key])
      if (cmp !== 0) return order === 'asc' ? cmp : -cmp
      return a.index - b.index
    })
    .map(item => item.row)
})

// --- Virtual scrolling ---
const rowVirtualizer = useVirtualizer(computed(() => ({
  count: sortedData.value?.length ?? 0,
  getScrollElement: () => tableWrapperRef.value,
  estimateSize: () => props.estimateRowHeight ?? 56,
  overscan: props.overscan ?? 5,
})))

const virtualItems = computed(() => rowVirtualizer.value.getVirtualItems())

const virtualPaddingTop = computed(() => {
  const items = virtualItems.value
  return items.length > 0 ? items[0].start : 0
})

const virtualPaddingBottom = computed(() => {
  const items = virtualItems.value
  if (items.length === 0) return 0
  return rowVirtualizer.value.getTotalSize() - items[items.length - 1].end
})

const measureElement = (el: any) => {
  if (el) {
    rowVirtualizer.value.measureElement(el as Element)
  }
}

const hasActionsColumn = computed(() => {
  return props.columns.some(column => column.key === 'actions')
})

const hasSelectColumn = computed(() => {
  return props.columns.length > 0 && props.columns[0].key === 'select'
})

// 生成固定列的 CSS 类
const getStickyColumnClass = (column: Column, index: number) => {
  const classes: string[] = []

  if (props.stickyFirstColumn) {
    // 如果第一列是勾选列，固定前两列（勾选+名称）
    if (hasSelectColumn.value) {
      if (index === 0) {
        classes.push('sticky-col sticky-col-left-first')
      } else if (index === 1) {
        classes.push('sticky-col sticky-col-left-second')
      }
    } else {
      // 否则只固定第一列
      if (index === 0) {
        classes.push('sticky-col sticky-col-left')
      }
    }
  }

  // 操作列固定（最后一列）
  if (props.stickyActionsColumn && column.key === 'actions') {
    classes.push('sticky-col sticky-col-right')
  }

  return classes.join(' ')
}

const getAdaptiveCellPaddingX = () => {
  const columnCount = props.columns.length

  if (columnCount >= 10) {
    return '0.5rem'
  } else if (columnCount >= 7) {
    return '0.75rem'
  } else if (columnCount >= 5) {
    return '1rem'
  } else {
    return '1.5rem'
  }
}

const getAdaptiveSelectColumnWidth = () => {
  const columnCount = props.columns.length

  if (columnCount >= 10) {
    return '2rem'
  } else if (columnCount >= 7) {
    return '2.5rem'
  } else if (columnCount >= 5) {
    return '3rem'
  } else {
    return '4rem'
  }
}

const tableLayoutVars = computed(() => ({
  '--data-table-cell-padding-x': getAdaptiveCellPaddingX(),
  '--data-table-select-col-width': getAdaptiveSelectColumnWidth(),
}))

// Init + keep persisted sort state consistent with current columns
const didInitSort = ref(false)

onMounted(() => {
  const initial = resolveInitialSortState()
  applySortState(initial)
  didInitSort.value = true
})

watch(
  columnsSignature,
  () => {
    // If current sort key is no longer sortable/visible, fall back to default/persisted.
    const normalized = normalizeSortKey(sortKey.value)
    if (!sortKey.value) {
      const initial = resolveInitialSortState()
      applySortState(initial)
      return
    }

    if (!normalized) {
      const fallback = resolveInitialSortState()
      if (fallback) {
        applySortState(fallback)
      } else {
        sortKey.value = ''
        sortOrder.value = 'asc'
      }
    }
  },
  { flush: 'post' }
)

watch(
  [sortKey, sortOrder],
  ([nextKey, nextOrder]) => {
    if (!didInitSort.value) return
    if (!props.sortStorageKey) return
    const key = normalizeSortKey(nextKey)
    if (!key) return
    writePersistedSortState({ key, order: normalizeSortOrder(nextOrder) })
  },
  { flush: 'post' }
)

defineExpose({
  virtualizer: rowVirtualizer,
  sortedData,
  resolveRowKey,
  tableWrapperEl: tableWrapperRef,
})
</script>

<style scoped>
.data-table-mobile {
  @apply space-y-3;
}

.data-table-mobile__card,
.data-table-mobile__empty {
  border: var(--theme-card-border-width) solid var(--theme-card-border);
  border-radius: var(--theme-surface-radius);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
}

.data-table-mobile__card {
  padding: var(--theme-table-mobile-card-padding);
}

.data-table-mobile__empty {
  @apply text-center;
  padding: var(--theme-table-mobile-empty-padding);
}

.data-table-mobile__label,
.data-table-desktop__head-cell,
.data-table-desktop__empty {
  color: var(--theme-page-muted);
}

.data-table-mobile__label {
  @apply font-medium;
  font-size: var(--theme-table-head-font-size);
  letter-spacing: var(--theme-table-head-letter-spacing);
  text-transform: var(--theme-table-head-text-transform);
}

.data-table-mobile__value,
.data-table-mobile__empty-title,
.data-table-desktop__cell {
  color: var(--theme-page-text);
}

.data-table-mobile__value {
  @apply text-right text-sm;
}

.data-table-mobile__empty-icon {
  @apply mb-4 h-12 w-12;
  color: var(--theme-input-placeholder);
}

.data-table-mobile__empty-title {
  @apply font-medium;
  font-family: var(--theme-empty-title-font);
  font-size: var(--theme-empty-title-size);
  letter-spacing: var(--theme-empty-title-letter-spacing);
}

.data-table-mobile__actions {
  @apply pt-3;
  border-top-width: var(--theme-card-divider-width);
  border-top-style: solid;
  border-color: var(--theme-page-border);
}

.data-table-mobile__actions-scroll {
  margin-inline: calc(var(--theme-table-mobile-actions-strip-padding-x) * -1);
}

.data-table-mobile__actions-row {
  @apply flex min-w-max items-center;
  gap: var(--theme-table-mobile-actions-gap);
  padding-inline: var(--theme-table-mobile-actions-strip-padding-x);
}

.data-table-mobile__skeleton {
  @apply h-4 animate-pulse rounded;
  background: color-mix(in srgb, var(--theme-page-border) 78%, transparent);
}

.data-table-mobile__skeleton--label {
  @apply w-20;
}

.data-table-mobile__skeleton--value {
  @apply w-32;
}

.data-table-mobile__skeleton--action {
  @apply h-8 w-full;
}

.data-table-desktop {
  @apply min-w-full;
  border-collapse: separate;
  border-spacing: 0;
}

.data-table-desktop__head-cell {
  background: var(--theme-table-head-bg);
  border-bottom: var(--theme-card-divider-width) solid var(--theme-page-border);
  font-size: var(--theme-table-head-font-size);
  letter-spacing: var(--theme-table-head-letter-spacing);
  text-transform: var(--theme-table-head-text-transform);
  padding-block: var(--theme-table-head-padding-y);
  padding-inline: var(--data-table-cell-padding-x, var(--theme-table-cell-padding-x));
}

.data-table-desktop__head-cell--sortable:hover {
  background: var(--theme-dropdown-item-hover-bg);
}

.data-table-desktop__sort-indicator {
  color: var(--theme-input-placeholder);
}

.data-table-desktop__body {
  background: var(--theme-surface);
}

.data-table-desktop__row {
  transition: background-color 0.15s ease;
}

.data-table-desktop__row:hover {
  background: var(--theme-table-row-hover);
}

.data-table-desktop__cell {
  border-bottom: var(--theme-card-divider-width) solid
    color-mix(in srgb, var(--theme-page-border) 72%, transparent);
  padding-block: var(--theme-table-cell-padding-y);
  padding-inline: var(--data-table-cell-padding-x, var(--theme-table-cell-padding-x));
}

.data-table-desktop__empty-cell {
  @apply text-center;
  padding-block: var(--theme-table-empty-padding-y);
}
</style>

<style scoped>
/* 表格横向滚动 */
.table-wrapper {
  --select-col-width: var(--data-table-select-col-width, 4rem);
  position: relative;
  overflow-x: auto;
  overflow-y: auto;
  flex: 1;
  min-height: 0;
  isolation: isolate;
}

/* 表头容器，确保在滚动时覆盖表体内容 */
.table-wrapper .table-header {
  position: sticky;
  top: 0;
  z-index: 200;
  background-color: var(--theme-table-head-bg);
}

/* 表体保持在表头下方 */
.table-body {
  position: relative;
  z-index: 0;
}

/* 所有表头单元格固定在顶部 */
.sticky-header-cell {
  position: sticky;
  top: 0;
  z-index: 210; /* 必须高于所有表体内容 */
  background-color: var(--theme-table-head-bg);
}

/* Sticky 列基础样式 */
.sticky-col {
  position: sticky;
  z-index: 20; /* 表体固定列 */
}

/* 单列固定（无勾选列时） */
.sticky-col-left {
  left: 0;
}

/* 双列固定（有勾选列时）：第一列（勾选） */
.sticky-col-left-first {
  left: 0;
}

/* 双列固定（有勾选列时）：第二列（名称） */
.sticky-col-left-second {
  left: var(--select-col-width);
}

/* 操作列固定 */
.sticky-col-right {
  right: 0;
}

/* 表头 sticky 列 - 需要比普通表头单元格更高的 z-index */
.sticky-header-cell.sticky-col {
  z-index: 220; /* 高于普通表头单元格和表体固定列 */
}

/* 表体 sticky 列背景 */
tbody .sticky-col {
  background-color: var(--theme-surface);
}

/* hover 状态保持 */
tbody tr:hover .sticky-col {
  background-color: color-mix(in srgb, var(--theme-table-row-hover) 64%, var(--theme-surface));
}

/* 阴影只在可滚动时显示 */
/* 单列固定右侧阴影 */
.is-scrollable .sticky-col-left::after {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  width: 10px;
  transform: translateX(100%);
  background: linear-gradient(
    to right,
    color-mix(in srgb, var(--theme-surface-contrast) 10%, transparent),
    transparent
  );
  pointer-events: none;
}

/* 双列固定：只在第二列显示阴影 */
.is-scrollable .sticky-col-left-second::after {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  width: 10px;
  transform: translateX(100%);
  background: linear-gradient(
    to right,
    color-mix(in srgb, var(--theme-surface-contrast) 10%, transparent),
    transparent
  );
  pointer-events: none;
}

/* 操作列左侧阴影 */
.is-scrollable .sticky-col-right::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  bottom: 0;
  width: 10px;
  transform: translateX(-100%);
  background: linear-gradient(
    to left,
    color-mix(in srgb, var(--theme-surface-contrast) 10%, transparent),
    transparent
  );
  pointer-events: none;
}
</style>
