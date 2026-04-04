<template>
  <span :class="['badge', getPromoCodeStatusClass(code)]">
    {{ t(getPromoCodeStatusLabelKey(code)) }}
  </span>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { PromoCode } from '@/types'
import {
  getPromoCodeStatusClass,
  getPromoCodeStatusLabelKey
} from '../promoCodeForm'

defineProps<{
  code: PromoCode
}>()

const { t } = useI18n()
</script>
