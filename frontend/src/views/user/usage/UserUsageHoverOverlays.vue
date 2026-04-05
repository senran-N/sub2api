<template>
  <UserUsageTokenTooltip
    :visible="tokenTooltipVisible"
    :position="tokenTooltipPosition"
    :log="tokenTooltipData"
  />
  <UserUsageCostTooltip
    :visible="tooltipVisible"
    :position="tooltipPosition"
    :log="tooltipData"
  />
</template>

<script setup lang="ts">
import type { UsageLog } from '@/types'
import type { UserUsageHoverTooltipPosition } from './useUserUsageHoverTooltip'
import UserUsageTokenTooltip from './UserUsageTokenTooltip.vue'
import UserUsageCostTooltip from './UserUsageCostTooltip.vue'

defineProps<{
  tokenTooltipVisible: boolean
  tokenTooltipPosition: UserUsageHoverTooltipPosition
  tokenTooltipData: UsageLog | null
  tooltipVisible: boolean
  tooltipPosition: UserUsageHoverTooltipPosition
  tooltipData: UsageLog | null
}>()
</script>
