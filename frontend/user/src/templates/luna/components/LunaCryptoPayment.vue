<template>
  <section class="luna-panel mb-[18px] min-w-0" data-testid="crypto-payment">
    <!-- 币种 + 状态 + 倒计时：手机上吸顶，滚动到哪都能看到剩余时间 -->
    <div
      class="flex items-center justify-between gap-3 rounded-t-[14px] border-b px-4 py-3 sm:px-6 sm:py-4 max-sm:sticky max-sm:top-[65px] max-sm:z-20"
      :class="toneBarClass"
      data-testid="crypto-status-bar"
    >
      <div class="flex min-w-0 items-center gap-2.5 sm:gap-3">
        <span class="luna-gradient grid h-9 w-9 flex-none place-items-center rounded-full text-[10px] font-extrabold tracking-tight sm:h-11 sm:w-11 sm:text-xs">{{ tokenBadge }}</span>
        <div class="min-w-0">
          <div class="truncate text-sm font-bold text-foreground sm:text-base">{{ token || channelName }}</div>
          <div class="break-words text-xs leading-snug text-muted-foreground sm:text-[13px]" data-testid="crypto-network">{{ networkLabel || channelName }}</div>
        </div>
      </div>
      <!-- 手机上状态叠在倒计时上方，省掉「剩余时间」文字 -->
      <div class="flex flex-none flex-col items-end gap-0.5 sm:flex-row sm:items-center sm:gap-5">
        <span v-if="!isExpired" class="inline-flex items-center gap-1.5 text-[11px] font-semibold text-foreground sm:text-[13px]">
          <span class="relative flex h-2 w-2 flex-none">
            <span v-if="polling" class="absolute inline-flex h-full w-full animate-ping rounded-full bg-success opacity-75 motion-reduce:animate-none"></span>
            <span class="relative inline-flex h-2 w-2 rounded-full bg-success"></span>
          </span>
          {{ t('luna.cryptoWaiting') }}
        </span>
        <div v-if="remainingMs !== null" class="text-right" role="timer">
          <div class="text-xs font-semibold opacity-80 max-sm:hidden">{{ t('payment.countdownLabel') }}</div>
          <div class="text-xl font-extrabold leading-tight tabular-nums sm:text-2xl" data-testid="crypto-countdown">{{ isExpired ? t('payment.countdownExpired') : countdownText }}</div>
        </div>
      </div>
    </div>

    <!-- 已过期：地址和二维码都不再显示，避免继续往失效的订单转账 -->
    <div v-if="isExpired" class="px-4 py-6 sm:px-6 sm:py-8" data-testid="crypto-expired">
      <div class="flex items-start gap-2.5 text-destructive">
        <TriangleAlert class="mt-0.5 h-5 w-5 flex-none" />
        <div class="min-w-0">
          <h2 class="text-base font-bold sm:text-lg">{{ t('luna.cryptoExpiredTitle') }}</h2>
          <p class="mt-1 text-[13px] leading-relaxed text-muted-foreground">{{ t('luna.cryptoExpiredHint') }}</p>
        </div>
      </div>
      <Button v-if="canChangeMethod" class="luna-gradient mt-5 h-11 w-full rounded-full font-bold sm:w-auto sm:px-8" @click="emit('change-method')">{{ t('luna.cryptoRetry') }}</Button>
    </div>

    <!-- 收款信息。窄屏顺序：数量、地址、网络警告、二维码；宽屏二维码放右侧 -->
    <div v-else class="grid grid-cols-1 lg:grid-cols-[minmax(0,1fr)_280px]">
      <div class="min-w-0 divide-y px-4 sm:px-6">
        <div class="py-4 sm:py-6">
          <div class="text-[13px] font-semibold text-muted-foreground">{{ t('luna.cryptoSendAmount') }}</div>
          <div class="mt-1.5 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div class="min-w-0 break-all text-[34px] font-extrabold leading-tight tabular-nums text-primary select-all sm:text-[44px]" data-testid="crypto-amount">{{ amount }}<span v-if="token" class="ml-2 select-none text-lg font-bold sm:text-2xl">{{ token }}</span></div>
            <Button class="h-11 w-full flex-none rounded-full font-bold sm:w-auto sm:px-6" data-testid="crypto-copy-amount" @click="copy('amount', amount)">
              <component :is="copyIcon('amount')" />{{ copyLabel('amount', t('luna.copyAmount')) }}
            </Button>
          </div>
          <p class="mt-2 text-[13px] font-medium text-foreground">{{ t('luna.cryptoExactAmount') }}</p>
        </div>

        <div class="py-4 sm:py-6">
          <div class="text-[13px] font-semibold text-muted-foreground">{{ t('luna.cryptoReceiveAddress') }}</div>
          <div class="mt-1.5 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <code class="min-w-0 break-all font-mono text-[15px] leading-relaxed text-foreground select-all sm:text-base" data-testid="crypto-address"><span class="font-bold text-primary">{{ addressParts.head }}</span>{{ addressParts.middle }}<span class="font-bold text-primary">{{ addressParts.tail }}</span></code>
            <Button variant="outline" class="h-11 w-full flex-none rounded-full font-bold sm:w-auto sm:px-6" data-testid="crypto-copy-address" @click="copy('address', address)">
              <component :is="copyIcon('address')" />{{ copyLabel('address', t('luna.copyAddress')) }}
            </Button>
          </div>
          <p class="mt-2 text-xs text-muted-foreground">{{ t('luna.cryptoAddressCheck') }}</p>
        </div>
      </div>

      <div v-if="qrScannable" class="flex flex-col items-center justify-center gap-2.5 border-t px-4 py-5 sm:px-6 lg:col-start-2 lg:row-start-1 lg:border-l lg:border-t-0 lg:py-6 max-lg:order-1">
        <div class="aspect-square w-full max-w-[200px] overflow-hidden rounded-md bg-white p-2 lg:max-w-[220px]"><img v-if="qrImageUrl" :src="qrImageUrl" alt="QR Code" class="h-full w-full object-contain" data-testid="crypto-qr" /></div>
        <div class="text-center text-xs text-muted-foreground">{{ t('luna.cryptoScanHint') }}</div>
      </div>

      <div v-if="networkLabel" class="flex items-start gap-2 border-t bg-warning-soft px-4 py-3 text-[13px] font-semibold text-warning sm:px-6 lg:col-span-2 lg:row-start-2" data-testid="crypto-network-warning">
        <TriangleAlert class="mt-px h-4 w-4 flex-none" />
        <span>{{ t('luna.cryptoNetworkWarning', { network: networkLabel }) }}</span>
      </div>
    </div>

    <!-- 订单信息：法币金额只作参考，字号要小，避免被当成转账数量 -->
    <div class="grid grid-cols-1 gap-3 border-t px-4 py-4 text-[13px] sm:px-6 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-center">
      <div class="grid min-w-0 grid-cols-1 gap-1.5">
        <div class="flex flex-wrap items-center gap-x-2 gap-y-1 text-muted-foreground">
          <span>{{ orderNoLabel }}</span>
          <span class="break-all font-semibold text-foreground select-all">{{ orderNo }}</span>
          <button type="button" class="inline-flex min-h-6 items-center gap-1 font-semibold text-primary hover:underline" @click="copy('orderNo', orderNo)">
            <component :is="copyIcon('orderNo')" class="h-3.5 w-3.5" />{{ copyLabel('orderNo', t('luna.copy')) }}
          </button>
        </div>
        <div v-if="summary" class="line-clamp-2 text-muted-foreground">{{ summary }}</div>
        <div class="font-medium text-foreground" data-testid="crypto-order-amount">{{ t('luna.cryptoOrderAmount', { amount: orderAmountDisplay }) }}</div>
        <details v-if="$slots.default" class="group">
          <summary class="inline-flex min-h-6 cursor-pointer list-none items-center gap-1 font-semibold text-primary [&::-webkit-details-marker]:hidden">
            {{ t('luna.cryptoOrderDetails') }}<ChevronDown class="h-3.5 w-3.5 transition-transform group-open:rotate-180" />
          </summary>
          <div class="mt-2 max-w-md"><slot /></div>
        </details>
      </div>
      <div class="grid grid-cols-2 gap-2 lg:flex">
        <Button variant="outline" class="h-11 rounded-full font-semibold" :class="{ 'col-span-2': !canChangeMethod || isExpired }" :disabled="refreshing" @click="emit('refresh')">{{ refreshLabel }}</Button>
        <Button v-if="canChangeMethod && !isExpired" variant="outline" class="h-11 rounded-full font-semibold" @click="emit('change-method')">{{ t('payment.changeMethod') }}</Button>
      </div>
    </div>

    <!-- 复制结果给读屏软件播报 -->
    <span class="sr-only" aria-live="polite">{{ copyAnnouncement }}</span>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Check, ChevronDown, Copy, TriangleAlert, X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { copyText } from '../../../utils/clipboard'
import {
  formatCryptoCountdown,
  isWalletScannableQr,
  resolveCryptoCountdownTone,
  resolveCryptoNetworkLabel,
  splitCryptoAddress,
} from '../../../utils/cryptoPaymentDisplay'

const props = withDefaults(defineProps<{
  amount: string
  address: string
  token?: string
  chain?: string
  channelName?: string
  qrContent?: string
  qrImageUrl?: string
  expiresAtMs?: number | null
  orderNo: string
  orderNoLabel: string
  orderAmountDisplay: string
  summary?: string
  polling?: boolean
  refreshing?: boolean
  refreshLabel: string
  canChangeMethod?: boolean
}>(), {
  token: '',
  chain: '',
  channelName: '',
  qrContent: '',
  qrImageUrl: '',
  expiresAtMs: null,
  summary: '',
  polling: false,
  refreshing: false,
  canChangeMethod: false,
})

const emit = defineEmits<{
  refresh: []
  'change-method': []
}>()

const { t } = useI18n()

const networkLabel = computed(() => resolveCryptoNetworkLabel(props.chain, props.token))
const addressParts = computed(() => splitCryptoAddress(props.address))
const tokenBadge = computed(() => (props.token || '?').slice(0, 4))
const qrScannable = computed(() => isWalletScannableQr(props.qrContent))

// ==================== 倒计时 ====================
const nowMs = ref(Date.now())
let tickTimer: number | undefined
onMounted(() => {
  tickTimer = window.setInterval(() => { nowMs.value = Date.now() }, 1000)
})

const remainingMs = computed(() => (props.expiresAtMs === null ? null : props.expiresAtMs - nowMs.value))
const tone = computed(() => resolveCryptoCountdownTone(remainingMs.value))
const isExpired = computed(() => tone.value === 'expired')
const countdownText = computed(() => formatCryptoCountdown(remainingMs.value ?? 0))
const toneBarClass = computed(() => {
  if (tone.value === 'expired') return 'bg-danger-soft text-destructive'
  if (tone.value === 'urgent') return 'bg-warning-soft text-warning'
  return 'bg-card text-primary'
})

// ==================== 复制 ====================
type CopyKey = 'amount' | 'address' | 'orderNo'
type CopyState = 'copied' | 'failed'

const copyStates = ref<Partial<Record<CopyKey, CopyState>>>({})
const copyAnnouncement = ref('')
const copyTimers = new Map<CopyKey, number>()

const copy = async (key: CopyKey, value: string) => {
  let state: CopyState = 'copied'
  try {
    await copyText(value)
  } catch {
    state = 'failed'
  }
  copyStates.value = { ...copyStates.value, [key]: state }
  copyAnnouncement.value = state === 'copied' ? t('payment.copied') : t('luna.copyFailed')
  window.clearTimeout(copyTimers.get(key))
  copyTimers.set(key, window.setTimeout(() => {
    copyStates.value = { ...copyStates.value, [key]: undefined }
  }, 1800))
}

const copyIcon = (key: CopyKey) => {
  if (copyStates.value[key] === 'copied') return Check
  return copyStates.value[key] === 'failed' ? X : Copy
}

const copyLabel = (key: CopyKey, idle: string) => {
  if (copyStates.value[key] === 'copied') return t('payment.copied')
  return copyStates.value[key] === 'failed' ? t('luna.copyFailed') : idle
}

onUnmounted(() => {
  window.clearInterval(tickTimer)
  copyTimers.forEach((timer) => window.clearTimeout(timer))
})
</script>
