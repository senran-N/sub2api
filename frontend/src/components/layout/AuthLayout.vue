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

      <!-- Grid / paper overlays: each theme opts in via CSS. -->
      <div class="auth-layout__grid absolute inset-0"></div>
      <div class="auth-layout__paper absolute inset-0"></div>

      <!-- Factory-only corner rivets. -->
      <span class="auth-layout__rivet auth-layout__rivet--tl"></span>
      <span class="auth-layout__rivet auth-layout__rivet--tr"></span>
      <span class="auth-layout__rivet auth-layout__rivet--bl"></span>
      <span class="auth-layout__rivet auth-layout__rivet--br"></span>
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
        <!-- Factory-only rivets inside the card. -->
        <span class="auth-layout__card-rivet auth-layout__card-rivet--tl"></span>
        <span class="auth-layout__card-rivet auth-layout__card-rivet--tr"></span>
        <span class="auth-layout__card-rivet auth-layout__card-rivet--bl"></span>
        <span class="auth-layout__card-rivet auth-layout__card-rivet--br"></span>
        <slot />
      </div>

      <!-- Footer Links -->
      <div class="mt-6 text-center text-sm">
        <slot name="footer" />
      </div>

      <!-- Copyright -->
      <div class="auth-layout__copyright mt-8 text-center text-xs">
        &copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const { t } = useI18n()

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
  background: var(--theme-auth-backdrop);
}

.auth-layout__orb {
  border-radius: 9999px;
  filter: blur(72px);
}

.auth-layout__orb--right {
  background: var(--theme-auth-orb-right-bg);
}

.auth-layout__orb--left {
  background: var(--theme-auth-orb-left-bg);
  opacity: var(--theme-auth-orb-left-opacity);
}

.auth-layout__orb--center {
  background: var(--theme-auth-orb-center-bg);
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
  position: relative;
  padding: var(--theme-auth-card-padding);
  border-radius: var(--theme-auth-card-radius);
  border: var(--theme-auth-card-border);
  box-shadow: var(--theme-auth-card-shadow);
  background: var(--theme-auth-card-bg);
  backdrop-filter: var(--theme-auth-card-backdrop-filter);
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

.auth-layout__grid {
  display: var(--theme-auth-grid-display);
  background-image: var(--theme-auth-grid-image);
  background-size: var(--theme-auth-grid-background-size);
  mask-image: var(--theme-auth-grid-mask);
  -webkit-mask-image: var(--theme-auth-grid-mask);
  opacity: var(--theme-auth-grid-opacity);
}

.auth-layout__paper {
  display: var(--theme-auth-paper-display);
  background-image: var(--theme-auth-paper-image);
  background-size: var(--theme-auth-paper-size);
  opacity: var(--theme-auth-paper-opacity);
  mix-blend-mode: var(--theme-auth-paper-blend);
}

.auth-layout__rivet {
  display: var(--theme-auth-rivet-display);
  position: absolute;
  width: var(--theme-auth-rivet-size);
  height: var(--theme-auth-rivet-size);
  background: var(--theme-auth-rivet-color);
  border-radius: var(--theme-auth-rivet-radius);
}

.auth-layout__rivet--tl { top: var(--theme-auth-rivet-offset); left: var(--theme-auth-rivet-offset); }
.auth-layout__rivet--tr { top: var(--theme-auth-rivet-offset); right: var(--theme-auth-rivet-offset); }
.auth-layout__rivet--bl { bottom: var(--theme-auth-rivet-offset); left: var(--theme-auth-rivet-offset); }
.auth-layout__rivet--br { bottom: var(--theme-auth-rivet-offset); right: var(--theme-auth-rivet-offset); }

.auth-layout__card-rivet {
  display: var(--theme-auth-card-rivet-display);
  position: absolute;
  width: var(--theme-auth-card-rivet-size);
  height: var(--theme-auth-card-rivet-size);
  background: var(--theme-auth-card-rivet-color);
  border-radius: var(--theme-auth-card-rivet-radius);
}

.auth-layout__card-rivet--tl { top: var(--theme-auth-card-rivet-offset); left: var(--theme-auth-card-rivet-offset); }
.auth-layout__card-rivet--tr { top: var(--theme-auth-card-rivet-offset); right: var(--theme-auth-card-rivet-offset); }
.auth-layout__card-rivet--bl { bottom: var(--theme-auth-card-rivet-offset); left: var(--theme-auth-card-rivet-offset); }
.auth-layout__card-rivet--br { bottom: var(--theme-auth-card-rivet-offset); right: var(--theme-auth-card-rivet-offset); }
</style>
