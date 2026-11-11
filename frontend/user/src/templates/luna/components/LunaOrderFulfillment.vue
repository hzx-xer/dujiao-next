<template>
  <div class="grid grid-cols-1 gap-3 sm:gap-4">
    <!-- 使用说明：只在已交付后展示，放在卡密上方并标「必看」，先读说明再用卡密 -->
    <template v-if="delivered">
      <section
        v-for="(block, bi) in instructions"
        :key="bi"
        class="overflow-hidden rounded-xl border-2 border-warning bg-card shadow-lg shadow-warning/15"
        data-testid="fulfillment-instructions"
      >
        <!-- 说明头用实心警示色，和下方绿色的交付内容区分开，一眼能看到 -->
        <div class="flex items-center gap-2.5 bg-warning px-3.5 py-3 text-ink-on-accent sm:px-5">
          <BookOpen class="h-5 w-5 flex-none" />
          <div class="min-w-0 flex-1">
            <h3 class="flex flex-wrap items-center gap-2 text-[17px] font-extrabold sm:text-lg">
              <span>{{ block.heading || t('orderDetail.instructionsTitle') }}</span>
              <span class="rounded-full bg-card px-2.5 py-0.5 text-xs font-extrabold leading-4 text-danger shadow-sm">{{ t('orderDetail.instructionsMustRead') }}</span>
            </h3>
            <div v-if="instructions.length > 1 && block.title" class="truncate text-xs opacity-85">{{ block.title }}</div>
            <p class="mt-0.5 text-xs font-semibold opacity-90 sm:text-[13px]">{{ t('orderDetail.instructionsReadFirst') }}</p>
          </div>
        </div>
        <div class="luna-prose prose max-w-none px-3.5 py-3 text-sm leading-[1.75] sm:px-5 sm:py-4 sm:text-[15px] sm:leading-7 prose-headings:font-extrabold prose-h2:mb-2 prose-h2:mt-5 prose-h2:text-base prose-h3:mb-2 prose-h3:mt-4 prose-h3:text-[15px] prose-p:my-2 prose-ol:my-2 prose-ul:my-2 prose-li:my-1 prose-img:rounded-md [&>:first-child]:mt-0 [&>:last-child]:mb-0" v-html="block.body"></div>
      </section>
    </template>

    <!-- 交付内容（卡密等）：页面上最需要一眼看到的部分，已交付时用成功色整块强调 -->
    <div
      class="overflow-hidden rounded-xl border"
      :class="delivered ? 'border-success/40 bg-success-soft' : 'border-dashed border-hairline-strong bg-secondary'"
      data-testid="fulfillment-content"
    >
      <div class="flex flex-wrap items-center justify-between gap-2.5 px-3.5 pb-2.5 pt-3 sm:gap-3 sm:px-5 sm:pb-3 sm:pt-3.5">
        <div class="flex min-w-0 items-center gap-3">
          <span
            class="grid h-9 w-9 flex-none place-items-center rounded-full sm:h-10 sm:w-10"
            :class="delivered ? 'bg-success text-ink-on-accent' : 'bg-card text-muted-foreground'"
          >
            <KeyRound class="h-5 w-5" />
          </span>
          <div class="min-w-0">
            <h3 class="text-base font-extrabold text-foreground">{{ title || t('orderDetail.fulfillmentContentTitle') }}</h3>
            <div class="mt-1 flex flex-wrap items-center gap-1.5 text-xs font-semibold">
              <span
                class="rounded-full px-2 py-0.5"
                :class="delivered ? 'bg-success text-ink-on-accent' : 'bg-warning-soft text-warning'"
              >{{ fulfillmentStatusLabelText(fulfillment.status) }}</span>
              <span class="rounded-full bg-card px-2 py-0.5 text-muted-foreground">{{ fulfillmentTypeLabelText(fulfillment.type) }}</span>
              <span v-if="listByLine && lineCount > 1" class="rounded-full bg-card px-2 py-0.5 tabular-nums text-muted-foreground">{{ t('orderDetail.fulfillmentTotalLines', { count: lineCount }) }}</span>
              <span v-else-if="!listByLine && expectedCount > 1" class="rounded-full bg-card px-2 py-0.5 tabular-nums text-muted-foreground">{{ t('orderDetail.fulfillmentSecretCount', { count: expectedCount }) }}</span>
            </div>
          </div>
        </div>
        <div class="flex flex-wrap gap-2 max-sm:w-full">
          <Button
            v-if="truncated"
            type="button"
            size="sm"
            class="h-9 rounded-full px-4 font-bold max-sm:h-10 max-sm:flex-1"
            :disabled="downloading"
            @click="emit('download', orderNo)"
          >
            <Download /> {{ downloading ? t('orderDetail.fulfillmentDownloading') : t('orderDetail.fulfillmentDownload') }}
          </Button>
          <Button
            v-if="delivered && !truncated && (listByLine ? lines.length > 1 : lines.length > 0)"
            type="button"
            size="sm"
            class="h-9 rounded-full bg-success px-4 font-bold text-ink-on-accent hover:bg-success/90 max-sm:h-10 max-sm:flex-1"
            data-testid="fulfillment-copy-all"
            @click="handleCopyFulfillment(fulfillment)"
          >
            <component :is="fulfillmentCopied ? Check : Copy" /> {{ fulfillmentCopied ? t('orderDetail.fulfillmentCopied') : (listByLine || expectedCount > 1 ? t('orderDetail.fulfillmentCopyAll') : t('orderDetail.fulfillmentCopy')) }}
          </Button>
        </div>
      </div>

      <div class="px-2.5 pb-2.5 sm:px-4 sm:pb-4">
        <div v-if="truncated" class="mb-2.5 rounded-md bg-warning-soft px-3 py-2 text-xs font-semibold text-warning">{{ t('orderDetail.fulfillmentTruncatedHint') }}</div>

        <ol
          v-if="listByLine && lines.length > 0"
          class="grid grid-cols-1 gap-2"
          :class="lines.length > 8 ? 'max-h-[30rem] overflow-y-auto overscroll-contain pr-1' : ''"
        >
          <li
            v-for="(line, index) in lines"
            :key="index"
            class="flex items-center gap-2.5 rounded-lg border bg-card py-2 pl-3 pr-1.5 shadow-sm sm:gap-3 sm:pl-4 sm:pr-2"
          >
            <span v-if="lines.length > 1" class="w-5 flex-none text-right text-xs font-bold tabular-nums text-muted-foreground">{{ index + 1 }}</span>
            <code class="luna-code min-w-0 flex-1 select-all py-1 [overflow-wrap:anywhere] text-[15px] font-bold leading-relaxed text-foreground sm:text-lg">{{ line }}</code>
            <!-- 只有一条：行内就是唯一的复制按钮，做成醒目的实心按钮；多条：表头有「复制全部」，行内只留图标按钮 -->
            <button
              v-if="delivered && !truncated && lines.length === 1"
              type="button"
              class="inline-flex h-10 flex-none items-center justify-center gap-1.5 rounded-full px-4 text-sm font-bold transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-success/40"
              :class="copiedLine === index ? 'bg-success-soft text-success' : 'bg-success text-ink-on-accent hover:bg-success/90'"
              data-testid="fulfillment-copy-line"
              @click="copyLine(line, index)"
            >
              <component :is="copiedLine === index ? Check : Copy" class="h-4 w-4" />
              {{ copiedLine === index ? t('orderDetail.fulfillmentCopied') : t('orderDetail.fulfillmentLineCopy') }}
            </button>
            <button
              v-else-if="delivered && !truncated"
              type="button"
              class="inline-flex h-9 w-9 flex-none items-center justify-center rounded-full border transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/40"
              :class="copiedLine === index ? 'border-success/40 bg-success-soft text-success' : 'border-primary/30 text-primary hover:bg-primary-soft'"
              :aria-label="`${t('orderDetail.fulfillmentLineCopy')} ${index + 1}`"
              :title="copiedLine === index ? t('orderDetail.fulfillmentCopied') : t('orderDetail.fulfillmentLineCopy')"
              data-testid="fulfillment-copy-line"
              @click="copyLine(line, index)"
            >
              <component :is="copiedLine === index ? Check : Copy" class="h-4 w-4" />
            </button>
          </li>
        </ol>
        <!-- 多行卡密 / 人工交付的大段文字：分不清哪几行属于同一条，整块原样展示，保留换行 -->
        <pre
          v-else-if="lines.length > 0"
          class="luna-code max-h-[32rem] overflow-y-auto overscroll-contain whitespace-pre-wrap rounded-lg border bg-card px-3.5 py-3 text-sm font-semibold leading-7 [overflow-wrap:anywhere] text-foreground shadow-sm sm:px-4 sm:text-base"
          data-testid="fulfillment-block"
        >{{ blockText }}</pre>
        <div v-else class="rounded-lg bg-card px-4 py-3.5 text-sm text-muted-foreground">{{ t('orderDetail.fulfillmentPendingHint') }}</div>

        <!-- 交付声明：提醒尽快保存，内容不会长期保留 -->
        <div v-if="delivered && lines.length > 0" class="mt-2.5 flex items-start gap-2 rounded-lg border border-danger/40 bg-danger-soft px-3 py-2 text-xs sm:mt-3 sm:py-2.5 sm:text-[13px] font-semibold leading-relaxed text-danger">
          <ShieldAlert class="mt-0.5 h-4 w-4 flex-none" />
          <span>{{ t('orderDetail.fulfillmentSaveNotice') }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { BookOpen, Check, Copy, Download, KeyRound, ShieldAlert } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { useOrderDisplayHelpers } from '../../../composables/useOrderDisplayHelpers'
import { copyText } from '../../../utils/clipboard'
import {
  canListFulfillmentByLine,
  expectedAutoSecretCount,
  splitFulfillmentLines,
  splitLeadingHeading,
} from '../../../utils/fulfillmentDisplay'

const props = defineProps<{
  title?: string
  fulfillment: any
  items?: any[]
  orderNo: string
  downloading: boolean
}>()

const emit = defineEmits<{
  (e: 'download', orderNo: string): void
}>()

const { t } = useI18n()

const {
  isFulfillmentTruncated, fulfillmentDeliveryLines, instructionBlocks,
  fulfillmentTypeLabelText, fulfillmentStatusLabelText, fulfillmentCopied, handleCopyFulfillment,
} = useOrderDisplayHelpers(ref<any>(null))

const delivered = computed(() => props.fulfillment?.status === 'delivered')
const truncated = computed(() => isFulfillmentTruncated(props.fulfillment))

// 有结构化发货信息（备注 + 键值）时逐条展示它，否则把原始内容按行拆成一条条卡密
const lines = computed(() => {
  const deliveryLines = fulfillmentDeliveryLines(props.fulfillment)
  return deliveryLines.length > 0 ? deliveryLines : splitFulfillmentLines(props.fulfillment?.payload)
})

const lineCount = computed(() => Number(props.fulfillment?.payload_line_count) || lines.value.length)

const expectedCount = computed(() => expectedAutoSecretCount(props.items))
const structured = computed(() => fulfillmentDeliveryLines(props.fulfillment).length > 0)
// 行数正好等于应发条数（或只有一行）才逐条列出，否则整块展示
const listByLine = computed(() => canListFulfillmentByLine({
  type: props.fulfillment?.type,
  lineCount: lineCount.value,
  expectedCount: expectedCount.value,
  structured: structured.value,
}))
// 整块展示时保留原始换行（含卡密内部的空行），只去掉首尾空白
const blockText = computed(() => String(props.fulfillment?.payload || '').replace(/\r\n/g, '\n').trim())

const instructions = computed(() =>
  instructionBlocks(props.items).map((block) => ({ title: block.title, ...splitLeadingHeading(block.html) })),
)

const copiedLine = ref<number | null>(null)
let copiedLineTimer: ReturnType<typeof setTimeout> | null = null

const copyLine = async (line: string, index: number) => {
  try {
    await copyText(line)
  } catch {
    return
  }
  copiedLine.value = index
  if (copiedLineTimer) clearTimeout(copiedLineTimer)
  copiedLineTimer = setTimeout(() => { copiedLine.value = null }, 1500)
}

onBeforeUnmount(() => {
  if (copiedLineTimer) clearTimeout(copiedLineTimer)
})
</script>
