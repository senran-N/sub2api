<template>
  <div class="announcement-bell">
    <!-- 铃铛按钮 -->
    <button
      @click="openModal"
      class="announcement-bell__trigger relative flex h-9 w-9 items-center justify-center transition-all"
      :class="{ 'announcement-bell__trigger--active': unreadCount > 0 }"
      :aria-label="t('announcements.title')"
    >
      <Icon name="bell" size="md" />
      <!-- 未读红点 -->
      <span
        v-if="unreadCount > 0"
        class="absolute right-1 top-1 flex h-2 w-2"
      >
        <span class="announcement-bell__dot-ping absolute inline-flex h-full w-full animate-ping rounded-full opacity-75"></span>
        <span class="announcement-bell__dot relative inline-flex h-2 w-2 rounded-full"></span>
      </span>
    </button>

    <!-- 公告列表 Modal -->
    <Teleport to="body">
      <Transition name="modal-fade">
        <div
          v-if="isModalOpen"
          class="announcement-bell__overlay announcement-bell__overlay--list fixed inset-0 z-[100] flex items-start justify-center overflow-y-auto"
          @click="closeModal"
        >
          <div
            class="announcement-bell__panel announcement-bell__panel--list w-full overflow-hidden"
            @click.stop
          >
            <!-- Header with Gradient -->
            <div class="announcement-bell__hero announcement-bell__hero--list relative overflow-hidden">
              <div class="relative z-10 flex items-start justify-between">
                <div>
                  <div class="flex items-center gap-2">
                    <div class="announcement-bell__hero-icon flex h-8 w-8 items-center justify-center">
                      <Icon name="bell" size="sm" />
                    </div>
                    <h2 class="announcement-bell__hero-title text-lg font-semibold">
                      {{ t('announcements.title') }}
                    </h2>
                  </div>
                  <p v-if="unreadCount > 0" class="announcement-bell__hero-subtext mt-2 text-sm">
                    <span class="announcement-bell__hero-count font-medium">{{ unreadCount }}</span>
                    {{ t('announcements.unread') }}
                  </p>
                </div>
                <div class="flex items-center gap-2">
                  <button
                    v-if="unreadCount > 0"
                    @click="markAllAsRead"
                    :disabled="loading"
                    class="announcement-bell__primary-action announcement-bell__primary-action--compact text-xs font-medium"
                  >
                    {{ t('announcements.markAllRead') }}
                  </button>
                  <button
                    @click="closeModal"
                    class="announcement-bell__ghost-icon flex h-9 w-9 items-center justify-center transition-all"
                    :aria-label="t('common.close')"
                  >
                    <Icon name="x" size="sm" />
                  </button>
                </div>
              </div>
              <!-- Decorative gradient -->
              <div class="announcement-bell__hero-sheen absolute right-0 top-0 h-full"></div>
            </div>

            <!-- Body -->
            <div class="announcement-bell__list-body overflow-y-auto">
              <!-- Loading -->
              <div v-if="loading" class="announcement-bell__state announcement-bell__state--loading flex items-center justify-center">
                <div class="announcement-bell__loader relative">
                  <div class="announcement-bell__loader-core h-12 w-12 animate-spin rounded-full border-4"></div>
                  <div class="announcement-bell__loader-ring absolute inset-0 h-12 w-12 animate-pulse rounded-full border-4"></div>
                </div>
              </div>

              <!-- Announcements List -->
              <div v-else-if="announcements.length > 0">
                <div
                  v-for="item in announcements"
                  :key="item.id"
                  class="announcement-bell__item group relative flex items-center gap-4 transition-all"
                  :class="{ 'announcement-bell__item--unread': !item.read_at }"
                  @click="openDetail(item)"
                >
                  <!-- Status Indicator -->
                  <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center">
                    <div
                      v-if="!item.read_at"
                      class="announcement-bell__item-icon announcement-bell__item-icon--unread relative flex h-10 w-10 items-center justify-center"
                    >
                      <!-- Pulse ring -->
                      <span class="announcement-bell__item-ping absolute inline-flex h-full w-full animate-ping opacity-75"></span>
                      <!-- Icon -->
                      <svg class="relative z-10 h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </div>
                    <div
                      v-else
                      class="announcement-bell__item-icon announcement-bell__item-icon--read flex h-10 w-10 items-center justify-center"
                    >
                      <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </div>
                  </div>

                  <!-- Content -->
                  <div class="flex min-w-0 flex-1 items-center justify-between gap-4">
                    <div class="min-w-0 flex-1">
                      <h3 class="announcement-bell__item-title truncate text-sm font-medium">
                        {{ item.title }}
                      </h3>
                      <div class="mt-1 flex items-center gap-2">
                        <time class="announcement-bell__item-time text-xs">
                          {{ formatRelativeTime(item.created_at) }}
                        </time>
                        <span
                          v-if="!item.read_at"
                          class="announcement-bell__item-pill inline-flex items-center gap-1 text-xs font-medium"
                        >
                          <span class="relative flex h-1.5 w-1.5">
                            <span class="announcement-bell__dot-ping absolute inline-flex h-full w-full animate-ping rounded-full opacity-75"></span>
                            <span class="announcement-bell__dot relative inline-flex h-1.5 w-1.5 rounded-full"></span>
                          </span>
                          {{ t('announcements.unread') }}
                        </span>
                      </div>
                    </div>

                    <!-- Arrow -->
                    <div class="flex-shrink-0">
                      <svg
                        class="announcement-bell__item-arrow h-5 w-5 transition-transform group-hover:translate-x-1"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                        stroke-width="2"
                      >
                        <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
                      </svg>
                    </div>
                  </div>

                  <!-- Unread indicator bar -->
                  <div
                    v-if="!item.read_at"
                    class="announcement-bell__item-bar absolute left-0 top-0 h-full w-1"
                  ></div>
                </div>
              </div>

              <!-- Empty State -->
              <div v-else class="announcement-bell__state announcement-bell__state--empty flex flex-col items-center justify-center">
                <div class="relative mb-4">
                  <div class="announcement-bell__empty-icon flex h-20 w-20 items-center justify-center">
                    <Icon name="inbox" size="xl" class="announcement-bell__empty-symbol" />
                  </div>
                  <div class="announcement-bell__empty-check absolute -right-1 -top-1 flex h-6 w-6 items-center justify-center">
                    <svg class="h-3.5 w-3.5" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                    </svg>
                  </div>
                </div>
                <p class="announcement-bell__empty-title text-sm font-medium">{{ t('announcements.empty') }}</p>
                <p class="announcement-bell__empty-description mt-1 text-xs">{{ t('announcements.emptyDescription') }}</p>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- 公告详情 Modal -->
    <Teleport to="body">
      <Transition name="modal-fade">
        <div
          v-if="detailModalOpen && selectedAnnouncement"
          class="announcement-bell__overlay announcement-bell__overlay--detail fixed inset-0 z-[110] flex items-start justify-center overflow-y-auto"
          @click="closeDetail"
        >
          <div
            class="announcement-bell__panel announcement-bell__panel--detail w-full overflow-hidden"
            @click.stop
          >
            <!-- Header with Decorative Elements -->
            <div class="announcement-bell__hero announcement-bell__hero--detail relative overflow-hidden">
              <!-- Decorative background elements -->
              <div class="announcement-bell__hero-sheen announcement-bell__hero-sheen--detail absolute right-0 top-0 h-full"></div>
              <div class="announcement-bell__hero-orb announcement-bell__hero-orb--right absolute"></div>
              <div class="announcement-bell__hero-orb announcement-bell__hero-orb--left absolute"></div>

              <div class="relative z-10 flex items-start justify-between gap-4">
                <div class="flex-1 min-w-0">
                  <!-- Icon and Category -->
                  <div class="mb-3 flex items-center gap-2">
                    <div class="announcement-bell__hero-icon announcement-bell__hero-icon--detail flex h-10 w-10 items-center justify-center">
                      <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </div>
                    <div class="flex items-center gap-2">
                      <span class="announcement-bell__category-badge text-xs font-medium">
                        {{ t('announcements.title') }}
                      </span>
                      <span
                        v-if="!selectedAnnouncement.read_at"
                        class="announcement-bell__hero-pill inline-flex items-center gap-1.5 text-xs font-medium"
                      >
                        <span class="relative flex h-2 w-2">
                          <span class="announcement-bell__hero-pill-ping absolute inline-flex h-full w-full animate-ping rounded-full opacity-75"></span>
                          <span class="announcement-bell__hero-pill-dot relative inline-flex h-2 w-2 rounded-full"></span>
                        </span>
                        {{ t('announcements.unread') }}
                      </span>
                    </div>
                  </div>

                  <!-- Title -->
                  <h2 class="announcement-bell__detail-title mb-3 leading-tight">
                    {{ selectedAnnouncement.title }}
                  </h2>

                  <!-- Meta Info -->
                  <div class="announcement-bell__meta flex items-center gap-4 text-sm">
                    <div class="flex items-center gap-1.5">
                      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                      <time>{{ formatRelativeWithDateTime(selectedAnnouncement.created_at) }}</time>
                    </div>
                    <div class="flex items-center gap-1.5">
                      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        <path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                      </svg>
                      <span>{{ selectedAnnouncement.read_at ? t('announcements.read') : t('announcements.unread') }}</span>
                    </div>
                  </div>
                </div>

                <!-- Close button -->
                <button
                  @click="closeDetail"
                  class="announcement-bell__ghost-icon flex h-10 w-10 flex-shrink-0 items-center justify-center transition-all"
                  :aria-label="t('common.close')"
                >
                  <Icon name="x" size="md" />
                </button>
              </div>
            </div>

            <!-- Body with Enhanced Markdown -->
            <div class="announcement-bell__detail-body overflow-y-auto">
              <!-- Content with decorative border -->
              <div class="relative">
                <!-- Decorative left border -->
                <div class="announcement-bell__content-rail absolute left-0 top-0 bottom-0 w-1"></div>

                <div class="announcement-bell__content-body">
                  <div
                    class="announcement-bell__markdown markdown-body max-w-none"
                    v-html="renderMarkdown(selectedAnnouncement.content)"
                  ></div>
                </div>
              </div>
            </div>

            <!-- Footer with Actions -->
            <div class="announcement-bell__footer">
              <div class="flex items-center justify-between">
                <div class="announcement-bell__footer-hint flex items-center gap-2 text-xs">
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <span>{{ selectedAnnouncement.read_at ? t('announcements.readStatus') : t('announcements.markReadHint') }}</span>
                </div>
                <div class="flex items-center gap-3">
                  <button
                    @click="closeDetail"
                    class="announcement-bell__secondary-action announcement-bell__secondary-action--detail text-sm font-medium"
                  >
                    {{ t('common.close') }}
                  </button>
                  <button
                    v-if="!selectedAnnouncement.read_at"
                    @click="markAsReadAndClose(selectedAnnouncement.id)"
                    class="announcement-bell__primary-action announcement-bell__primary-action--detail text-sm font-medium"
                  >
                    <span class="flex items-center gap-2">
                      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                      </svg>
                      {{ t('announcements.markRead') }}
                    </span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { storeToRefs } from 'pinia'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useAppStore } from '@/stores/app'
import { useAnnouncementStore } from '@/stores/announcements'
import { formatRelativeTime, formatRelativeWithDateTime } from '@/utils/format'
import type { UserAnnouncement } from '@/types'
import Icon from '@/components/icons/Icon.vue'
import { lockBodyScroll as acquireBodyScrollLock, unlockBodyScroll as releaseBodyScrollLock } from '@/utils/bodyScrollLock'
import { resolveRequestErrorMessage } from '@/utils/requestError'

const { t } = useI18n()
const appStore = useAppStore()
const announcementStore = useAnnouncementStore()

// Configure marked
marked.setOptions({
  breaks: true,
  gfm: true,
})

// Use store state (storeToRefs for reactivity)
const { announcements, loading } = storeToRefs(announcementStore)
const unreadCount = computed(() => announcementStore.unreadCount)

// Local modal state
const isModalOpen = ref(false)
const detailModalOpen = ref(false)
const selectedAnnouncement = ref<UserAnnouncement | null>(null)
let bodyScrollLocked = false

// Methods
function renderMarkdown(content: string): string {
  if (!content) return ''
  const html = marked.parse(content) as string
  return DOMPurify.sanitize(html)
}

function openModal() {
  isModalOpen.value = true
}

function closeModal() {
  isModalOpen.value = false
}

function openDetail(announcement: UserAnnouncement) {
  selectedAnnouncement.value = announcement
  detailModalOpen.value = true
  if (!announcement.read_at) {
    markAsRead(announcement.id)
  }
}

function closeDetail() {
  detailModalOpen.value = false
  selectedAnnouncement.value = null
}

async function markAsRead(id: number) {
  try {
    await announcementStore.markAsRead(id)
  } catch (err: any) {
    appStore.showError(resolveRequestErrorMessage(err, t('common.unknownError')))
  }
}

async function markAsReadAndClose(id: number) {
  await markAsRead(id)
  appStore.showSuccess(t('announcements.markedAsRead'))
  closeDetail()
}

async function markAllAsRead() {
  try {
    await announcementStore.markAllAsRead()
    appStore.showSuccess(t('announcements.allMarkedAsRead'))
  } catch (err: any) {
    appStore.showError(resolveRequestErrorMessage(err, t('common.unknownError')))
  }
}

function handleEscape(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    if (detailModalOpen.value) {
      closeDetail()
    } else if (isModalOpen.value) {
      closeModal()
    }
  }
}

function syncBodyScrollLock(shouldLock: boolean) {
  if (shouldLock && !bodyScrollLocked) {
    bodyScrollLocked = true
    acquireBodyScrollLock()
    return
  }

  if (!shouldLock && bodyScrollLocked) {
    bodyScrollLocked = false
    releaseBodyScrollLock()
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleEscape)
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleEscape)
  syncBodyScrollLock(false)
})

watch(
  [isModalOpen, detailModalOpen, () => announcementStore.currentPopup],
  ([modal, detail, popup]) => {
    syncBodyScrollLock(Boolean(modal || detail || popup))
  },
  { immediate: true }
)
</script>

<style scoped>
/* Modal Animations */
.modal-fade-enter-active {
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.modal-fade-leave-active {
  transition: all 0.2s cubic-bezier(0.4, 0, 1, 1);
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}

.modal-fade-enter-from > div {
  transform: scale(0.94) translateY(-12px);
  opacity: 0;
}

.modal-fade-leave-to > div {
  transform: scale(0.96) translateY(-8px);
  opacity: 0;
}

.announcement-bell {
  --announcement-danger-rgb: var(--theme-danger-rgb);
  --announcement-success-rgb: var(--theme-success-rgb);
  --announcement-overlay-padding: max(var(--theme-floating-panel-viewport-padding), 1rem);
  --announcement-overlay-top-list: 8vh;
  --announcement-overlay-top-detail: 6vh;
  --announcement-hero-list-padding-x: calc(var(--theme-markdown-block-padding) + 0.5rem);
  --announcement-hero-list-padding-y: calc(var(--theme-markdown-block-padding) + 0.25rem);
  --announcement-hero-detail-padding-x: calc(var(--theme-markdown-block-padding) + 1rem);
  --announcement-hero-detail-padding-y: calc(var(--theme-markdown-block-padding) + 0.5rem);
  --announcement-list-body-max-height: 65vh;
  --announcement-detail-body-max-height: 60vh;
  --announcement-detail-body-padding: calc(var(--theme-markdown-block-padding) * 2);
  --announcement-content-padding-left: calc(var(--theme-markdown-block-padding) + 0.5rem);
  --announcement-footer-padding-x: calc(var(--theme-markdown-block-padding) * 2);
  --announcement-footer-padding-y: calc(var(--theme-markdown-block-padding) + 0.25rem);
  --announcement-item-padding-x: calc(var(--theme-markdown-block-padding) + 0.5rem);
  --announcement-item-padding-y: var(--theme-markdown-block-padding);
  --announcement-state-padding-y: calc(var(--theme-table-mobile-empty-padding) * 0.67);
  --announcement-pill-padding-y: 0.125rem;
  --announcement-pill-padding-x: 0.375rem;
  --announcement-category-padding-y: 0.25rem;
  --announcement-category-padding-x: 0.625rem;
  --announcement-action-padding-compact-y: 0.5rem;
  --announcement-action-padding-compact-x: 1rem;
  --announcement-action-padding-detail-y: 0.625rem;
  --announcement-action-padding-detail-x: 1.25rem;
}

.announcement-bell__trigger,
.announcement-bell__ghost-icon,
.announcement-bell__item,
.announcement-bell__primary-action,
.announcement-bell__secondary-action,
.announcement-bell__item-pill,
.announcement-bell__hero-pill,
.announcement-bell__category-badge,
.announcement-bell__panel {
  border-radius: calc(var(--theme-button-radius) + 4px);
}

.announcement-bell__trigger {
  color: var(--theme-page-muted);
}

.announcement-bell__trigger:hover {
  background: var(--theme-button-ghost-hover-bg);
  color: var(--theme-page-text);
  transform: scale(1.05);
}

.announcement-bell__trigger--active {
  color: var(--theme-accent);
}

.announcement-bell__dot-ping {
  background: color-mix(in srgb, rgb(var(--announcement-danger-rgb)) 70%, transparent);
}

.announcement-bell__dot {
  background: rgb(var(--announcement-danger-rgb));
}

.announcement-bell__overlay {
  padding: var(--announcement-overlay-padding);
  background:
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--theme-overlay-strong) 100%, transparent),
      color-mix(in srgb, var(--theme-overlay-soft) 100%, transparent),
      color-mix(in srgb, var(--theme-overlay-strong) 92%, transparent)
    );
  backdrop-filter: blur(18px);
}

.announcement-bell__overlay--list {
  padding-top: var(--announcement-overlay-top-list);
}

.announcement-bell__overlay--detail {
  padding-top: var(--announcement-overlay-top-detail);
}

.announcement-bell__panel {
  background: var(--theme-surface);
  border: var(--theme-card-border-width) solid
    color-mix(in srgb, var(--theme-card-border) 78%, transparent);
  box-shadow: var(--theme-card-shadow-hover);
}

.announcement-bell__panel--list {
  width: min(calc(100vw - 2rem), var(--theme-announcement-list-width));
  border-radius: var(--theme-announcement-panel-radius-list);
}

.announcement-bell__panel--detail {
  width: min(calc(100vw - 2rem), var(--theme-announcement-detail-width));
  border-radius: var(--theme-announcement-panel-radius-detail);
}

.announcement-bell__hero {
  padding: var(--announcement-hero-list-padding-y) var(--announcement-hero-list-padding-x);
  border-bottom: 1px solid color-mix(in srgb, var(--theme-page-border) 80%, transparent);
  background:
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--theme-accent-soft) 68%, var(--theme-surface) 32%),
      color-mix(in srgb, var(--theme-surface-soft) 84%, transparent)
    );
}

.announcement-bell__hero--detail {
  padding: var(--announcement-hero-detail-padding-y) var(--announcement-hero-detail-padding-x);
  background:
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--theme-accent-soft) 82%, var(--theme-surface) 18%),
      color-mix(in srgb, var(--theme-surface-soft) 82%, transparent)
    );
}

.announcement-bell__hero-icon,
.announcement-bell__hero-pill,
.announcement-bell__primary-action {
  color: var(--theme-filled-text);
  background:
    linear-gradient(
      135deg,
      var(--theme-accent),
      color-mix(in srgb, var(--theme-accent-strong) 22%, var(--theme-accent) 78%)
    );
  box-shadow: 0 14px 28px color-mix(in srgb, var(--theme-accent) 24%, transparent);
}

.announcement-bell__hero-icon--detail {
  border-radius: calc(var(--theme-button-radius) + 6px);
}

.announcement-bell__hero-title,
.announcement-bell__item-title,
.announcement-bell__detail-title,
.announcement-bell__empty-title {
  font-family: var(--theme-announcement-title-font);
  color: var(--theme-page-text);
}

.announcement-bell__detail-title {
  font-size: var(--theme-announcement-title-size);
  font-weight: 700;
  letter-spacing: var(--theme-announcement-title-letter-spacing);
}

.announcement-bell__hero-subtext,
.announcement-bell__item-time,
.announcement-bell__empty-description,
.announcement-bell__meta,
.announcement-bell__footer-hint {
  color: var(--theme-page-muted);
}

.announcement-bell__hero-count {
  color: var(--theme-accent);
}

.announcement-bell__hero-sheen {
  width: min(28vw, var(--theme-announcement-sheen-width));
  background: linear-gradient(
    270deg,
    color-mix(in srgb, var(--theme-accent) 10%, transparent),
    transparent
  );
}

.announcement-bell__hero-sheen--detail {
  width: min(34vw, var(--theme-announcement-sheen-width));
  background: linear-gradient(
    270deg,
    color-mix(in srgb, var(--theme-accent) 14%, transparent),
    transparent
  );
}

.announcement-bell__hero-orb {
  border-radius: 999px;
  filter: blur(var(--theme-announcement-hero-orb-blur));
}

.announcement-bell__hero-orb--right {
  top: var(--theme-announcement-hero-orb-top-offset);
  right: var(--theme-announcement-hero-orb-right-offset);
  width: var(--theme-announcement-hero-orb-large-size);
  height: var(--theme-announcement-hero-orb-large-size);
}

.announcement-bell__hero-orb--right {
  background: color-mix(in srgb, var(--theme-accent) 18%, transparent);
}

.announcement-bell__hero-orb--left {
  left: var(--theme-announcement-hero-orb-left-offset);
  bottom: var(--theme-announcement-hero-orb-bottom-offset);
  width: var(--theme-announcement-hero-orb-small-size);
  height: var(--theme-announcement-hero-orb-small-size);
}

.announcement-bell__hero-orb--left {
  background: color-mix(in srgb, var(--theme-surface-emphasis) 12%, transparent);
}

.announcement-bell__ghost-icon {
  background: color-mix(in srgb, var(--theme-page-backdrop) 86%, transparent);
  color: var(--theme-page-muted);
  backdrop-filter: blur(12px);
}

.announcement-bell__ghost-icon:hover {
  background: var(--theme-surface);
  color: var(--theme-page-text);
  box-shadow: var(--theme-card-shadow);
}

.announcement-bell__loader-core {
  border-color: color-mix(in srgb, var(--theme-page-border) 90%, transparent);
  border-top-color: var(--theme-accent);
}

.announcement-bell__loader-ring {
  border-color: color-mix(in srgb, var(--theme-accent) 24%, transparent);
}

.announcement-bell__list-body {
  max-height: var(--announcement-list-body-max-height);
}

.announcement-bell__state {
  padding-top: var(--announcement-state-padding-y);
  padding-bottom: var(--announcement-state-padding-y);
}

.announcement-bell__item {
  padding: var(--announcement-item-padding-y) var(--announcement-item-padding-x);
  min-height: var(--theme-announcement-item-min-height);
  border-bottom: 1px solid color-mix(in srgb, var(--theme-page-border) 70%, transparent);
}

.announcement-bell__item:hover {
  background: var(--theme-table-row-hover);
}

.announcement-bell__item--unread {
  background: color-mix(in srgb, var(--theme-accent-soft) 48%, transparent);
}

.announcement-bell__item-icon {
  border-radius: calc(var(--theme-button-radius) + 6px);
}

.announcement-bell__item-icon--unread {
  color: var(--theme-filled-text);
  background:
    linear-gradient(
      135deg,
      var(--theme-accent),
      color-mix(in srgb, var(--theme-accent-strong) 22%, var(--theme-accent) 78%)
    );
  box-shadow: 0 12px 24px color-mix(in srgb, var(--theme-accent) 22%, transparent);
}

.announcement-bell__item-ping {
  border-radius: calc(var(--theme-button-radius) + 6px);
  background: color-mix(in srgb, var(--theme-accent) 50%, transparent);
}

.announcement-bell__item-icon--read {
  background: color-mix(in srgb, var(--theme-surface-soft) 92%, transparent);
  color: color-mix(in srgb, var(--theme-page-muted) 70%, transparent);
}

.announcement-bell__item-pill,
.announcement-bell__category-badge {
  padding: var(--announcement-pill-padding-y) var(--announcement-pill-padding-x);
  background: color-mix(in srgb, var(--theme-accent) 14%, var(--theme-surface));
  color: color-mix(in srgb, var(--theme-accent) 82%, var(--theme-page-text));
}

.announcement-bell__category-badge,
.announcement-bell__hero-pill {
  padding: var(--announcement-category-padding-y) var(--announcement-category-padding-x);
}

.announcement-bell__item-arrow {
  color: color-mix(in srgb, var(--theme-page-muted) 62%, transparent);
}

.announcement-bell__item-bar,
.announcement-bell__content-rail {
  border-radius: var(--theme-version-icon-radius);
  background:
    linear-gradient(
      180deg,
      var(--theme-accent),
      color-mix(in srgb, var(--theme-accent-strong) 28%, var(--theme-accent) 72%)
    );
}

.announcement-bell__empty-icon {
  border-radius: var(--theme-version-icon-radius);
  background:
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--theme-surface-soft) 92%, transparent),
      color-mix(in srgb, var(--theme-page-border) 62%, transparent)
    );
}

.announcement-bell__empty-symbol {
  color: color-mix(in srgb, var(--theme-page-muted) 68%, transparent);
}

.announcement-bell__empty-check {
  border-radius: var(--theme-version-icon-radius);
  color: var(--theme-filled-text);
  background: rgb(var(--announcement-success-rgb));
  box-shadow: 0 10px 18px color-mix(in srgb, rgb(var(--announcement-success-rgb)) 30%, transparent);
}

.announcement-bell__hero-pill-ping,
.announcement-bell__hero-pill-dot {
  background: var(--theme-filled-text);
}

.announcement-bell__detail-body {
  max-height: var(--announcement-detail-body-max-height);
  padding: var(--announcement-detail-body-padding);
  background: var(--theme-surface);
}

.announcement-bell__content-body {
  padding-left: var(--announcement-content-padding-left);
}

.announcement-bell__footer {
  padding: var(--announcement-footer-padding-y) var(--announcement-footer-padding-x);
  border-top: 1px solid color-mix(in srgb, var(--theme-page-border) 80%, transparent);
  background: color-mix(in srgb, var(--theme-surface-soft) 82%, transparent);
}

.announcement-bell__secondary-action {
  border: 1px solid var(--theme-card-border);
  background: var(--theme-button-secondary-bg);
  color: var(--theme-button-secondary-text);
  box-shadow: var(--theme-card-shadow);
}

.announcement-bell__secondary-action:hover {
  background: var(--theme-button-secondary-hover-bg);
  box-shadow: var(--theme-card-shadow-hover);
}

.announcement-bell__primary-action {
  padding: var(--announcement-action-padding-compact-y) var(--announcement-action-padding-compact-x);
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease,
    background 0.2s ease;
}

.announcement-bell__primary-action--detail {
  padding: var(--announcement-action-padding-detail-y) var(--announcement-action-padding-detail-x);
}

.announcement-bell__secondary-action--detail {
  padding: var(--announcement-action-padding-detail-y) var(--announcement-action-padding-detail-x);
}

.announcement-bell__primary-action:hover {
  transform: translateY(-1px);
  box-shadow: 0 18px 32px color-mix(in srgb, var(--theme-accent) 28%, transparent);
}

.announcement-bell__primary-action:disabled {
  opacity: 0.5;
}

/* Scrollbar Styling */
.overflow-y-auto::-webkit-scrollbar {
  width: 8px;
}

.overflow-y-auto::-webkit-scrollbar-track {
  background: transparent;
}

.overflow-y-auto::-webkit-scrollbar-thumb {
  background: linear-gradient(
    to bottom,
    color-mix(in srgb, var(--theme-scrollbar-thumb) 78%, var(--theme-surface)),
    var(--theme-scrollbar-thumb)
  );
  border-radius: 4px;
}

.overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(
    to bottom,
    color-mix(in srgb, var(--theme-scrollbar-thumb-hover) 72%, var(--theme-surface)),
    var(--theme-scrollbar-thumb-hover)
  );
}
</style>

<style>
/* Enhanced Markdown Styles */
.markdown-body {
  font-size: 0.9375rem;
  line-height: 1.75;
  color: var(--theme-page-text);
}

.markdown-body h1 {
  @apply mb-6 mt-8 pb-3 font-bold;
  font-family: var(--theme-markdown-heading-font);
  font-size: var(--theme-markdown-heading-1-size);
  color: var(--theme-page-text);
  border-bottom: 1px solid var(--theme-page-border);
}

.markdown-body h2 {
  @apply mb-4 mt-7 pb-2 font-bold;
  font-family: var(--theme-markdown-heading-font);
  font-size: var(--theme-markdown-heading-2-size);
  color: var(--theme-page-text);
  border-bottom: 1px solid color-mix(in srgb, var(--theme-page-border) 70%, transparent);
}

.markdown-body h3 {
  @apply mb-3 mt-6 text-xl font-semibold;
  color: var(--theme-page-text);
}

.markdown-body h4 {
  @apply mb-2 mt-5 text-lg font-semibold;
  color: var(--theme-page-text);
}

.markdown-body p {
  @apply mb-4 leading-relaxed;
}

.markdown-body a {
  @apply font-medium underline decoration-2 underline-offset-2 transition-all;
  color: var(--theme-accent);
  text-decoration-color: color-mix(in srgb, var(--theme-accent) 30%, transparent);
}

.markdown-body a:hover {
  text-decoration-color: var(--theme-accent);
}

.markdown-body ul,
.markdown-body ol {
  @apply mb-4 ml-6 space-y-2;
}

.markdown-body ul {
  @apply list-disc;
}

.markdown-body ol {
  @apply list-decimal;
}

.markdown-body li {
  @apply leading-relaxed;
  padding-left: 0.5rem;
}

.markdown-body li::marker {
  color: var(--theme-accent);
}

.markdown-body blockquote {
  @apply relative my-5 border-l-4 italic;
  padding: calc(var(--theme-markdown-block-padding) * 0.75)
    calc(var(--theme-markdown-block-padding) * 0.75)
    calc(var(--theme-markdown-block-padding) * 0.75)
    calc(var(--theme-markdown-block-padding) + 0.25rem);
  color: var(--theme-page-text);
  border-left-color: var(--theme-accent);
  background: color-mix(in srgb, var(--theme-accent-soft) 58%, transparent);
}

.markdown-body blockquote::before {
  content: '"';
  @apply absolute -left-1 top-0 text-5xl font-serif;
  color: color-mix(in srgb, var(--theme-accent) 22%, transparent);
}

.markdown-body code {
  @apply font-mono;
  padding: 0.25rem 0.5rem;
  font-size: 0.8125rem;
  border-radius: var(--theme-markdown-code-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 92%, transparent);
  color: color-mix(in srgb, var(--theme-accent) 78%, var(--theme-page-text));
}

.markdown-body pre {
  @apply my-5 overflow-x-auto;
  padding: var(--theme-markdown-block-padding);
  border: 1px solid var(--theme-card-border);
  border-radius: var(--theme-markdown-block-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 84%, transparent);
}

.markdown-body pre code {
  @apply bg-transparent;
  padding: 0;
  font-size: 0.8125rem;
  color: var(--theme-page-text);
}

.markdown-body hr {
  @apply my-8 border-0 border-t-2;
  border-top-color: var(--theme-page-border);
}

.markdown-body table {
  @apply mb-5 w-full overflow-hidden;
  border: 1px solid var(--theme-card-border);
  border-radius: var(--theme-markdown-block-radius);
}

.markdown-body th,
.markdown-body td {
  @apply border-r border-b text-left;
  padding: var(--theme-table-cell-padding-y) var(--theme-table-cell-padding-x);
  border-color: var(--theme-page-border);
}

.markdown-body th:last-child,
.markdown-body td:last-child {
  @apply border-r-0;
}

.markdown-body tr:last-child td {
  @apply border-b-0;
}

.markdown-body th {
  @apply font-semibold;
  color: var(--theme-page-text);
  background:
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--theme-accent-soft) 68%, var(--theme-surface) 32%),
      color-mix(in srgb, var(--theme-surface-soft) 88%, transparent)
    );
}

.markdown-body tbody tr {
  @apply transition-colors;
}

.markdown-body tbody tr:hover {
  background: var(--theme-table-row-hover);
}

.markdown-body img {
  @apply my-5 max-w-full;
  border: 1px solid var(--theme-card-border);
  border-radius: var(--theme-markdown-block-radius);
  box-shadow: var(--theme-card-shadow);
}

.markdown-body strong {
  @apply font-semibold;
  color: var(--theme-page-text);
}

.markdown-body em {
  @apply italic;
  color: var(--theme-page-muted);
}
</style>
