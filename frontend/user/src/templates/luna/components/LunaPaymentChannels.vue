<template>
  <div>
    <!-- 单列紧凑列表 + 单选圆点，渠道名不截断；渠道多时只露前几个，其余折叠 -->
    <div :id="listId" class="grid grid-cols-1 gap-1.5" role="radiogroup" :aria-label="ariaLabel">
      <button
        v-for="channel in visibleChannels"
        :key="channel.id"
        type="button"
        role="radio"
        class="flex w-full items-center gap-2.5 rounded-[10px] border-2 bg-card px-3 py-2 text-left transition-colors"
        :class="[
          isSelected(channel) ? 'border-primary bg-primary/10' : 'border-hairline-strong hover:border-primary/50',
          isDisabled(channel) ? 'cursor-not-allowed opacity-55' : '',
        ]"
        :aria-checked="isSelected(channel)"
        :disabled="isDisabled(channel)"
        :title="isDisabled(channel) ? limitHint(channel) : ''"
        @click="select(channel)"
      >
        <img v-if="channel.icon" :src="getImageUrl(channel.icon)" alt="" loading="lazy" class="h-6 w-6 flex-none rounded-md object-contain" />
        <span v-else class="grid h-6 w-6 flex-none place-items-center rounded-md text-[13px] font-extrabold leading-none text-white" :style="{ background: channelBadge(channel).bg }" aria-hidden="true">
          <component :is="channelBadge(channel).icon" v-if="channelBadge(channel).icon" class="h-3.5 w-3.5" />
          <template v-else>{{ channelBadge(channel).text }}</template>
        </span>
        <span class="min-w-0 flex-1">
          <span class="block break-words text-sm font-semibold leading-snug text-foreground">{{ channel.name }}</span>
          <span v-if="channel.fee_policy === 'customer_surcharge'" class="mt-0.5 block text-[11.5px] text-warning">{{ t('payment.feeLabel') }}：{{ feeDescription(channel) }}</span>
          <span v-if="isDisabled(channel)" class="mt-0.5 block text-[11px] text-warning">{{ limitHint(channel) }}</span>
        </span>
        <span class="grid h-[18px] w-[18px] flex-none place-items-center rounded-full border-2" :class="isSelected(channel) ? 'border-primary' : 'border-hairline-strong'">
          <span v-if="isSelected(channel)" class="h-2 w-2 rounded-full bg-primary"></span>
        </span>
      </button>
    </div>

    <!-- 展开 / 收起：与渠道行同宽的虚线按钮，叠放被收起渠道的小图标，提示里面还有什么 -->
    <button
      v-if="collapsible"
      ref="toggleRef"
      type="button"
      class="mt-1.5 flex min-h-11 w-full items-center justify-center gap-2 rounded-[10px] border-2 border-dashed px-3 py-2 text-[13px] font-semibold text-primary transition-colors"
      :class="expanded ? 'border-hairline-strong hover:border-primary/50' : 'border-primary/40 bg-primary/5 hover:border-primary/70 hover:bg-primary/10'"
      :aria-controls="listId"
      :aria-expanded="expanded"
      @click="toggle"
    >
      <template v-if="expanded">{{ t('luna.collapsePaymentMethods') }} <ChevronUp class="h-4 w-4" /></template>
      <template v-else>
        <span class="flex -space-x-1.5" aria-hidden="true">
          <template v-for="channel in hiddenChannels.slice(0, 4)" :key="channel.id">
            <img v-if="channel.icon" :src="getImageUrl(channel.icon)" alt="" loading="lazy" class="h-5 w-5 rounded-full bg-card object-contain ring-2 ring-card" />
            <span v-else class="grid h-5 w-5 place-items-center rounded-full text-[10px] font-extrabold leading-none text-white ring-2 ring-card" :style="{ background: channelBadge(channel).bg }">
              <component :is="channelBadge(channel).icon" v-if="channelBadge(channel).icon" class="h-3 w-3" />
              <template v-else>{{ channelBadge(channel).text }}</template>
            </span>
          </template>
        </span>
        {{ t('luna.morePaymentMethods', { count: hiddenChannels.length }) }}
        <ChevronDown class="h-4 w-4" />
      </template>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, type Component } from 'vue'
import { useI18n } from 'vue-i18n'
import { ChevronDown, ChevronUp, CreditCard, MessageCircle, Wallet } from 'lucide-vue-next'
import { getImageUrl } from '../../../utils/image'

const props = defineProps<{
  channels: any[]
  modelValue: number | null
  ariaLabel?: string
  isChannelDisabledForAmount?: (channel?: any) => boolean
  channelAmountLimitHint?: (channel?: any) => string
}>()

const emit = defineEmits<{ 'update:modelValue': [value: number] }>()

const { t } = useI18n()

// 同页多个实例时 aria-controls 仍指向自己的列表
const listId = `luna-payment-channels-${Math.random().toString(36).slice(2, 8)}`

const isDisabled = (channel: any) => Boolean(props.isChannelDisabledForAmount?.(channel))
const limitHint = (channel: any) => String(props.channelAmountLimitHint?.(channel) || '')
const isSelected = (channel: any) => Number(props.modelValue) === Number(channel.id) && !isDisabled(channel)

const select = (channel: any) => {
  if (!channel || isDisabled(channel)) return
  const id = Number(channel.id)
  if (Number.isFinite(id) && id > 0) emit('update:modelValue', id)
}

/** 与公共 PaymentChannelSelector 一致：费率% + 固定费 */
const feeDescription = (channel: any) => {
  const parts: string[] = []
  const rate = Number(channel?.fee_rate || 0)
  const fixed = Number(channel?.fixed_fee || 0)
  if (rate > 0) parts.push(`${rate.toFixed(2)}%`)
  if (fixed > 0) parts.push(fixed.toFixed(2))
  return parts.join(' + ') || t('payment.feeFree')
}

/** 没上传图标的渠道按类型给底色 + 图标或字母徽标，方便一眼区分 */
const channelBadge = (channel: any): { bg: string; icon?: Component; text?: string } => {
  const type = String(channel?.channel_type || '').toLowerCase()
  if (type === 'alipay') return { bg: '#1677ff', icon: Wallet }
  if (type === 'wxpay' || type === 'wechat') return { bg: '#07c160', icon: MessageCircle }
  if (type === 'qqpay') return { bg: '#12b7f5', icon: MessageCircle }
  if (type === 'paypal') return { bg: '#003087', text: 'P' }
  if (type === 'stripe') return { bg: '#635bff', text: 'S' }
  if (type.startsWith('usdt') || type === 'bepusdt') return { bg: '#26a17b', text: '₮' }
  if (type.startsWith('usdc')) return { bg: '#2775ca', text: '$' }
  if (type === 'trx') return { bg: '#eb0029', text: 'T' }
  return { bg: 'var(--luna-accent)', icon: CreditCard }
}

// 渠道多时折叠：只露前 4 个；总数不超过 5 个时全部显示，免得出现只藏一个的「+1」
const VISIBLE_LIMIT = 4
const expanded = ref(false)

/** 可用的排前面、因金额限制不可用的排最后（各自保持原顺序），折叠时不可用的自然收进去 */
const ordered = computed(() => [
  ...props.channels.filter((c) => !isDisabled(c)),
  ...props.channels.filter((c) => isDisabled(c)),
])
const collapsible = computed(() => ordered.value.length > VISIBLE_LIMIT + 1)
const visibleChannels = computed(() => {
  const list = ordered.value
  if (expanded.value || !collapsible.value) return list
  const head = list.slice(0, VISIBLE_LIMIT)
  // 选中的渠道在折叠区里时也露出来，保证当前选择始终看得见
  const selected = list.find((c) => isSelected(c))
  return selected && !head.includes(selected) ? [...head, selected] : head
})
const hiddenChannels = computed(() => ordered.value.filter((c) => !visibleChannels.value.includes(c)))

const toggleRef = ref<HTMLElement | null>(null)
const toggle = async () => {
  expanded.value = !expanded.value
  if (expanded.value) return
  // 展开后往下滑再收起，列表变短会把按钮甩到视口上方，拉回来
  await nextTick()
  toggleRef.value?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
}
</script>
