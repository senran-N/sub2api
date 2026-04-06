<template>
  <Teleport to="body">
    <Transition name="popup-fade">
      <div
        v-if="announcementStore.currentPopup"
        class="announcement-popup fixed inset-0 z-[120] flex items-start justify-center overflow-y-auto"
      >
        <div
          class="announcement-popup__panel w-full overflow-hidden"
          @click.stop
        >
          <!-- Header with warm gradient -->
          <div class="announcement-popup__hero relative overflow-hidden">
            <!-- Decorative background -->
            <div class="announcement-popup__hero-sheen absolute right-0 top-0 h-full"></div>
            <div class="announcement-popup__hero-orb announcement-popup__hero-orb--right absolute"></div>
            <div class="announcement-popup__hero-orb announcement-popup__hero-orb--left absolute"></div>

            <div class="relative z-10">
              <!-- Icon and badge -->
              <div class="mb-3 flex items-center gap-2">
                <div class="announcement-popup__hero-icon flex h-10 w-10 items-center justify-center">
                  <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                  </svg>
                </div>
                <span class="announcement-popup__pill inline-flex items-center gap-1.5 text-xs font-medium">
                  <span class="relative flex h-2 w-2">
                    <span class="announcement-popup__pill-ping absolute inline-flex h-full w-full animate-ping rounded-full opacity-75"></span>
                    <span class="announcement-popup__pill-dot relative inline-flex h-2 w-2 rounded-full"></span>
                  </span>
                  {{ t('announcements.unread') }}
                </span>
              </div>

              <!-- Title -->
              <h2 class="announcement-popup__title mb-2 leading-tight">
                {{ announcementStore.currentPopup.title }}
              </h2>

              <!-- Time -->
              <div class="announcement-popup__meta flex items-center gap-1.5 text-sm">
                <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <time>{{ formatRelativeWithDateTime(announcementStore.currentPopup.created_at) }}</time>
              </div>
            </div>
          </div>

          <!-- Body -->
          <div class="announcement-popup__body overflow-y-auto">
            <div class="relative">
              <div class="announcement-popup__rail absolute left-0 top-0 bottom-0 w-1"></div>
              <div class="pl-6">
                <div
                  class="announcement-popup__markdown markdown-body max-w-none"
                  v-html="renderedContent"
                ></div>
              </div>
            </div>
          </div>

          <!-- Footer -->
          <div class="announcement-popup__footer">
            <div class="flex items-center justify-end">
              <button
                @click="handleDismiss"
                class="announcement-popup__action text-sm font-medium"
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
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useAnnouncementStore } from '@/stores/announcements'
import { formatRelativeWithDateTime } from '@/utils/format'
import { lockBodyScroll as acquireBodyScrollLock, unlockBodyScroll as releaseBodyScrollLock } from '@/utils/bodyScrollLock'

const { t } = useI18n()
const announcementStore = useAnnouncementStore()
let bodyScrollLocked = false

marked.setOptions({
  breaks: true,
  gfm: true,
})

const renderedContent = computed(() => {
  const content = announcementStore.currentPopup?.content
  if (!content) return ''
  const html = marked.parse(content) as string
  return DOMPurify.sanitize(html)
})

function handleDismiss() {
  announcementStore.dismissPopup()
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

watch(
  () => announcementStore.currentPopup,
  (popup) => {
    syncBodyScrollLock(Boolean(popup))
  },
  { immediate: true }
)

onBeforeUnmount(() => {
  syncBodyScrollLock(false)
})
</script>

<style scoped>
.popup-fade-enter-active {
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.popup-fade-leave-active {
  transition: all 0.2s cubic-bezier(0.4, 0, 1, 1);
}

.popup-fade-enter-from,
.popup-fade-leave-to {
  opacity: 0;
}

.popup-fade-enter-from > div {
  transform: scale(0.94) translateY(-12px);
  opacity: 0;
}

.popup-fade-leave-to > div {
  transform: scale(0.96) translateY(-8px);
  opacity: 0;
}

.announcement-popup {
  padding: var(--theme-announcement-popup-viewport-padding);
  padding-top: var(--theme-announcement-popup-top-offset);
  background:
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--theme-overlay-strong) 100%, transparent),
      color-mix(in srgb, var(--theme-overlay-soft) 100%, transparent),
      color-mix(in srgb, var(--theme-overlay-strong) 92%, transparent)
    );
  backdrop-filter: blur(18px);
}

.announcement-popup__panel,
.announcement-popup__hero-icon,
.announcement-popup__pill,
.announcement-popup__action {
  border-radius: calc(var(--theme-button-radius) + 6px);
}

.announcement-popup__panel {
  width: min(calc(100vw - 2rem), var(--theme-announcement-popup-width));
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 78%, transparent);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow-hover);
}

.announcement-popup__hero {
  padding: var(--theme-announcement-popup-hero-padding-y)
    var(--theme-announcement-popup-hero-padding-x);
  border-bottom: 1px solid color-mix(in srgb, var(--theme-page-border) 80%, transparent);
  background:
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--theme-accent-soft) 72%, var(--theme-surface)),
      color-mix(in srgb, var(--theme-surface-soft) 84%, transparent)
    );
}

.announcement-popup__hero-sheen {
  width: min(34vw, var(--theme-announcement-sheen-width));
  background: linear-gradient(
    270deg,
    color-mix(in srgb, var(--theme-accent) 12%, transparent),
    transparent
  );
}

.announcement-popup__hero-orb {
  border-radius: 999px;
  filter: blur(var(--theme-announcement-hero-orb-blur));
}

.announcement-popup__hero-orb--right {
  top: var(--theme-announcement-hero-orb-top-offset);
  right: var(--theme-announcement-hero-orb-right-offset);
  width: var(--theme-announcement-hero-orb-large-size);
  height: var(--theme-announcement-hero-orb-large-size);
}

.announcement-popup__hero-orb--right {
  background: color-mix(in srgb, var(--theme-accent) 18%, transparent);
}

.announcement-popup__hero-orb--left {
  left: var(--theme-announcement-hero-orb-left-offset);
  bottom: var(--theme-announcement-hero-orb-bottom-offset);
  width: var(--theme-announcement-hero-orb-small-size);
  height: var(--theme-announcement-hero-orb-small-size);
}

.announcement-popup__hero-orb--left {
  background: color-mix(in srgb, var(--theme-surface-emphasis) 12%, transparent);
}

.announcement-popup__hero-icon,
.announcement-popup__pill,
.announcement-popup__action {
  color: var(--theme-filled-text);
  background:
    linear-gradient(
      135deg,
      var(--theme-accent),
      color-mix(in srgb, var(--theme-accent-strong) 22%, var(--theme-accent) 78%)
    );
  box-shadow: 0 16px 30px color-mix(in srgb, var(--theme-accent) 24%, transparent);
}

.announcement-popup__pill {
  padding: var(--theme-announcement-popup-pill-padding-y)
    var(--theme-announcement-popup-pill-padding-x);
}

.announcement-popup__title {
  font-family: var(--theme-announcement-title-font);
  font-size: var(--theme-announcement-title-size);
  font-weight: 700;
  letter-spacing: var(--theme-announcement-title-letter-spacing);
  color: var(--theme-page-text);
}

.announcement-popup__pill-ping,
.announcement-popup__pill-dot {
  background: var(--theme-filled-text);
}

.announcement-popup__meta {
  color: var(--theme-page-muted);
}

.announcement-popup__body {
  max-height: var(--theme-announcement-popup-body-max-height);
  padding: var(--theme-announcement-popup-body-padding-y)
    var(--theme-announcement-popup-body-padding-x);
  background: var(--theme-surface);
}

.announcement-popup__rail {
  border-radius: var(--theme-version-icon-radius);
  background:
    linear-gradient(
      180deg,
      var(--theme-accent),
      color-mix(in srgb, var(--theme-accent-strong) 26%, var(--theme-accent) 74%)
    );
}

.announcement-popup__footer {
  padding: var(--theme-announcement-popup-footer-padding-y)
    var(--theme-announcement-popup-footer-padding-x);
  border-top: 1px solid color-mix(in srgb, var(--theme-page-border) 80%, transparent);
  background: color-mix(in srgb, var(--theme-surface-soft) 82%, transparent);
}

.announcement-popup__action {
  padding: var(--theme-announcement-popup-action-padding-y)
    var(--theme-announcement-popup-action-padding-x);
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease,
    background 0.2s ease;
}

.announcement-popup__action:hover {
  transform: translateY(-1px);
  box-shadow: 0 20px 36px color-mix(in srgb, var(--theme-accent) 28%, transparent);
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
</style>
