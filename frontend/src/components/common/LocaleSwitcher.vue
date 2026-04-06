<template>
  <div class="relative" ref="dropdownRef">
    <button
      @click="toggleDropdown"
      :disabled="switching"
      class="locale-switcher__trigger flex items-center gap-1.5 text-sm font-medium transition-colors"
      :title="currentLocale?.name"
    >
      <span class="text-base">{{ currentLocale?.flag }}</span>
      <span class="hidden sm:inline">{{ currentLocale?.code.toUpperCase() }}</span>
      <Icon
        name="chevronDown"
        size="xs"
        class="locale-switcher__chevron transition-transform duration-200"
        :class="{ 'rotate-180': isOpen }"
      />
    </button>

    <transition name="dropdown">
      <div
        v-if="isOpen"
        class="locale-switcher__panel absolute right-0 z-50 mt-1 w-32 overflow-hidden"
      >
        <button
          v-for="locale in availableLocales"
          :key="locale.code"
          :disabled="switching"
          @click="selectLocale(locale.code)"
          class="locale-switcher__option flex w-full items-center gap-2 text-sm transition-colors"
          :class="{
            'locale-switcher__option--active': locale.code === currentLocaleCode
          }"
        >
          <span class="text-base">{{ locale.flag }}</span>
          <span>{{ locale.name }}</span>
          <Icon v-if="locale.code === currentLocaleCode" name="check" size="sm" class="locale-switcher__check ml-auto" />
        </button>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { setLocale, availableLocales } from '@/i18n'

const { locale } = useI18n()

const isOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)
const switching = ref(false)

const currentLocaleCode = computed(() => locale.value)
const currentLocale = computed(() => availableLocales.find((l) => l.code === locale.value))

function toggleDropdown() {
  isOpen.value = !isOpen.value
}

async function selectLocale(code: string) {
  if (switching.value || code === currentLocaleCode.value) {
    isOpen.value = false
    return
  }
  switching.value = true
  try {
    await setLocale(code)
    isOpen.value = false
  } finally {
    switching.value = false
  }
}

function handleClickOutside(event: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) {
    isOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.15s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(-4px);
}

.locale-switcher__trigger,
.locale-switcher__panel {
  border-radius: calc(var(--theme-button-radius) + 2px);
}

.locale-switcher__trigger {
  padding:
    calc(var(--theme-dropdown-item-padding-y) + 0.125rem)
    calc(var(--theme-dropdown-item-padding-x) - 0.5rem);
  color: var(--theme-page-muted);
}

.locale-switcher__trigger:hover {
  background: var(--theme-button-ghost-hover-bg);
  color: var(--theme-page-text);
}

.locale-switcher__chevron {
  color: color-mix(in srgb, var(--theme-page-muted) 70%, transparent);
}

.locale-switcher__panel {
  border: 1px solid var(--theme-dropdown-border);
  background: var(--theme-dropdown-bg);
  box-shadow: var(--theme-dropdown-shadow);
}

.locale-switcher__option {
  padding: var(--theme-dropdown-item-padding-y) var(--theme-dropdown-item-padding-x);
  color: var(--theme-page-text);
}

.locale-switcher__option:hover {
  background: var(--theme-dropdown-item-hover-bg);
}

.locale-switcher__option--active {
  background: color-mix(in srgb, var(--theme-accent-soft) 72%, var(--theme-surface));
  color: var(--theme-accent);
}

.locale-switcher__check {
  color: var(--theme-accent);
}
</style>
