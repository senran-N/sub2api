<template>
  <footer class="public-page-footer">
    <div class="public-page-footer__inner">
      <p class="public-page-footer__text text-sm">
        &copy; {{ currentYear }} {{ siteName }}. {{ rightsLabel }}
      </p>
      <div class="flex items-center gap-4">
        <a
          v-if="docUrl"
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="public-page-footer__link text-sm transition-colors"
        >
          {{ docsLabel }}
        </a>
        <a
          :href="githubUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="public-page-footer__link text-sm transition-colors"
        >
          GitHub
        </a>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
defineProps<{
  currentYear: number
  siteName: string
  docUrl?: string
  githubUrl: string
  rightsLabel: string
  docsLabel: string
}>()
</script>

<style scoped>
.public-page-footer {
  position: relative;
  z-index: 10;
  border-top-width: var(--theme-public-footer-border-width);
  border-top-style: solid;
  border-color: color-mix(in srgb, var(--theme-page-border) 68%, transparent);
  padding: 2rem 1.5rem;
}

.public-page-footer__inner {
  display: flex;
  width: 100%;
  max-width: var(--theme-public-shell-max-width);
  margin: 0 auto;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  text-align: center;
}

.public-page-footer__text,
.public-page-footer__link {
  color: var(--theme-page-muted);
}

.public-page-footer__link:hover {
  color: var(--theme-page-text);
}

@media (min-width: 640px) {
  .public-page-footer__inner {
    flex-direction: row;
    text-align: left;
  }
}

@media (max-width: 767px) {
  .public-page-footer {
    padding-left: 1rem;
    padding-right: 1rem;
  }
}
</style>
