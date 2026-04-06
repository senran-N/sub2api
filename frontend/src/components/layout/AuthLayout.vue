<template>
  <div class="auth-layout relative flex min-h-screen items-center justify-center overflow-hidden">
    <!-- Background -->
    <div class="auth-layout__backdrop absolute inset-0"></div>

    <!-- Decorative Elements -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <!-- Gradient Orbs -->
      <div class="auth-layout__orb auth-layout__orb--right absolute -right-40 -top-40 h-80 w-80"></div>
      <div class="auth-layout__orb auth-layout__orb--left absolute -bottom-40 -left-40 h-80 w-80"></div>
      <div class="auth-layout__orb auth-layout__orb--center absolute left-1/2 top-1/2 h-96 w-96 -translate-x-1/2 -translate-y-1/2"></div>

      <!-- Grid Pattern -->
      <div class="auth-layout__grid absolute inset-0"></div>
    </div>

    <!-- Content Container -->
    <div class="auth-layout__content relative z-10 w-full">
      <!-- Logo/Brand -->
      <div class="mb-8 text-center">
        <!-- Custom Logo or Default Logo -->
        <template v-if="settingsLoaded">
          <div
            class="auth-layout__logo mb-4 inline-flex items-center justify-center overflow-hidden"
          >
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <h1 class="auth-layout__title text-gradient mb-2">
            {{ siteName }}
          </h1>
          <p class="auth-layout__subtitle text-sm">
            {{ siteSubtitle }}
          </p>
        </template>
      </div>

      <!-- Card Container -->
      <div class="auth-layout__card card-glass shadow-glass">
        <slot />
      </div>

      <!-- Footer Links -->
      <div class="mt-6 text-center text-sm">
        <slot name="footer" />
      </div>

      <!-- Copyright -->
      <div class="auth-layout__copyright mt-8 text-center text-xs">
        &copy; {{ currentYear }} {{ siteName }}. All rights reserved.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'Subscription to API Conversion Platform')
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>

<style scoped>
.auth-layout {
  padding: 1rem;
}

.auth-layout__content {
  max-width: var(--theme-auth-container-max-width);
}

.auth-layout__backdrop {
  background:
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--theme-page-bg) 82%, var(--theme-surface) 18%),
      color-mix(in srgb, var(--theme-accent-soft) 70%, var(--theme-page-bg) 30%) 55%,
      color-mix(in srgb, var(--theme-surface-soft) 80%, var(--theme-page-bg) 20%)
    );
}

.auth-layout__orb {
  border-radius: 9999px;
  filter: blur(72px);
}

.auth-layout__orb--right {
  background: color-mix(in srgb, var(--theme-accent) 22%, transparent);
}

.auth-layout__orb--left {
  background: color-mix(in srgb, var(--theme-accent-strong) 12%, transparent);
  opacity: 0.72;
}

.auth-layout__orb--center {
  background: color-mix(in srgb, var(--theme-surface-emphasis) 8%, transparent);
}

.auth-layout__grid {
  background-image:
    linear-gradient(
      color-mix(in srgb, var(--theme-accent) 10%, transparent) 1px,
      transparent 1px
    ),
    linear-gradient(
      90deg,
      color-mix(in srgb, var(--theme-accent) 10%, transparent) 1px,
      transparent 1px
    );
  background-size: var(--theme-auth-grid-size) var(--theme-auth-grid-size);
  opacity: var(--theme-auth-grid-opacity);
}

.auth-layout__logo {
  height: var(--theme-auth-logo-size);
  width: var(--theme-auth-logo-size);
  border-radius: var(--theme-auth-logo-radius);
  background: color-mix(in srgb, var(--theme-page-backdrop) 88%, var(--theme-surface) 12%);
  border: var(--theme-auth-logo-border);
  box-shadow: var(--theme-auth-logo-shadow);
}

.auth-layout__title {
  font-family: var(--theme-auth-title-font);
  font-size: var(--theme-auth-title-size);
  font-weight: var(--theme-auth-title-weight);
  letter-spacing: var(--theme-auth-title-letter-spacing);
  text-transform: var(--theme-auth-title-transform);
}

.auth-layout__subtitle,
.auth-layout__copyright {
  color: var(--theme-page-muted);
}

.auth-layout__card {
  padding: var(--theme-auth-card-padding);
  border-radius: var(--theme-auth-card-radius);
}

.text-gradient {
  background:
    linear-gradient(
      135deg,
      var(--theme-accent),
      color-mix(in srgb, var(--theme-accent-strong) 28%, var(--theme-accent) 72%)
    );
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}
</style>
