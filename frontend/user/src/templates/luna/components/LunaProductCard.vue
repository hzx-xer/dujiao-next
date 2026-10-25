<template>
  <!-- 列表模式：各断点都是单列紧凑行 -->
  <RouterLink
    v-if="layout === 'list'"
    :to="`/products/${product.slug}`"
    class="group flex min-w-0 items-center gap-3 rounded-[10px] border-2 border-border bg-card px-3 py-2.5 text-left shadow-[var(--luna-shadow-sm)] transition hover:border-primary/45 sm:px-3.5"
    :class="{ 'opacity-70': soldOut }"
  >
    <!-- 没有图就不放缩略图，文字直接靠左 -->
    <span v-if="hasImage" class="relative block h-12 w-12 flex-none overflow-hidden rounded-lg bg-primary/10 sm:h-14 sm:w-14">
      <img :src="coverImage" :alt="title" loading="lazy" class="absolute inset-0 h-full w-full" :class="coverIsCategoryIcon ? 'object-contain p-2' : 'object-cover'" @error="imageErrored = true" />
    </span>

    <div class="min-w-0 flex-1">
      <h3 class="line-clamp-2 break-words text-sm font-medium leading-snug text-foreground">{{ title }}</h3>
      <!-- 窄屏价格放到标题下方，右侧只留购买按钮，给标题腾出宽度 -->
      <div class="mt-1 flex flex-wrap items-baseline gap-x-1.5 sm:hidden">
        <template v-if="promo">
          <span class="text-[15px] font-bold tabular-nums text-destructive">{{ formatPrice(getPromotionPriceAmount(product), siteCurrency) }}</span>
          <span class="text-[11px] tabular-nums text-muted-foreground line-through">{{ formatPrice(product.price_amount, siteCurrency) }}</span>
        </template>
        <span v-else class="text-[15px] font-bold tabular-nums text-primary">{{ formatPrice(product.price_amount, siteCurrency) }}</span>
        <span v-if="priceSignal" class="rounded px-1.5 py-px text-[11px] font-medium" :class="priceSignal.tone">{{ priceSignal.label }}</span>
      </div>
      <div class="mt-1 flex flex-wrap items-center gap-1.5">
        <span v-if="categoryLabel" class="max-w-full truncate rounded bg-secondary px-1.5 py-px text-[11px] font-medium text-muted-foreground">{{ categoryLabel }}</span>
        <span class="inline-flex items-center gap-1 rounded px-1.5 py-px text-[11px] font-medium" :class="stockPill.tone">
          <component :is="stockPill.icon" class="h-3 w-3" />
          {{ stockPill.label }}
        </span>
        <span class="hidden rounded px-1.5 py-px text-[11px] font-medium sm:inline" :class="fulfillmentTone">{{ getFulfillmentTypeLabel(product.fulfillment_type) }}</span>
        <template v-if="!soldOut && product.tags && product.tags.length">
          <span v-for="(tag, i) in product.tags.slice(0, 2)" :key="i" class="hidden max-w-full truncate rounded bg-destructive/10 px-1.5 py-px text-[11px] font-medium text-destructive sm:inline">{{ tag }}</span>
        </template>
      </div>
    </div>

    <div class="hidden flex-none text-right sm:block">
      <template v-if="promo">
        <span class="block text-[15px] font-bold tabular-nums text-destructive sm:text-base">{{ formatPrice(getPromotionPriceAmount(product), siteCurrency) }}</span>
        <span class="block text-[11px] tabular-nums text-muted-foreground line-through">{{ formatPrice(product.price_amount, siteCurrency) }}</span>
      </template>
      <span v-else class="block text-[15px] font-bold tabular-nums text-primary sm:text-base">{{ formatPrice(product.price_amount, siteCurrency) }}</span>
      <span v-if="priceSignal" class="mt-0.5 inline-block rounded px-1.5 py-px text-[11px] font-medium" :class="priceSignal.tone">{{ priceSignal.label }}</span>
    </div>

    <template v-if="!soldOut">
      <span aria-hidden="true" class="luna-gradient inline-flex h-8 flex-none items-center gap-1 rounded-full px-3 text-[13px] font-semibold transition group-hover:brightness-110 group-active:scale-[0.98] sm:hidden">
        <ShoppingCart class="h-3.5 w-3.5" />
        {{ t('products.quickBuy') }}
      </span>
      <span aria-hidden="true" class="luna-gradient hidden h-8 flex-none items-center rounded-full px-3.5 text-[13px] font-semibold transition group-hover:brightness-110 sm:inline-flex">{{ t('products.quickBuy') }}</span>
    </template>
  </RouterLink>

  <!-- 卡片模式：窄屏竖版（首页一行两个），sm 起横版 -->
  <RouterLink
    v-else
    :to="`/products/${product.slug}`"
    class="group flex min-w-0 flex-col overflow-hidden rounded-[10px] border-2 border-border bg-card text-left shadow-[var(--luna-shadow-sm)] transition hover:-translate-y-0.5 hover:border-primary/45 sm:h-full sm:flex-row sm:items-center sm:gap-3 sm:p-3.5"
    :class="{ 'opacity-70': soldOut }"
  >
    <!-- 没有图就不放图片区（不再用渐变占位图） -->
    <span v-if="hasImage" class="relative block aspect-[4/3] w-full flex-none overflow-hidden bg-primary/10 sm:aspect-auto sm:h-[72px] sm:w-[72px] sm:rounded-[10px]">
      <img :src="coverImage" :alt="title" loading="lazy" class="absolute inset-0 h-full w-full" :class="coverIsCategoryIcon ? 'object-contain p-4 sm:p-2' : 'object-cover'" @error="imageErrored = true" />
    </span>

    <div class="flex min-w-0 flex-1 flex-col gap-1.5 p-2.5 sm:p-0">
      <!-- 窄屏标题固定占两行高，同为有图（或同为无图）的一排卡片底部仍对齐 -->
      <h3 class="line-clamp-2 min-h-[2.75em] break-words text-[13px] font-medium leading-snug text-foreground sm:min-h-0 sm:text-sm">{{ title }}</h3>

      <!-- 价格 + 购买：窄屏压到卡片底部，购买是通栏按钮；sm 起价格回到标题下方，按钮改用右侧文字版 -->
      <div class="order-last mt-auto sm:order-none sm:mt-0">
        <div class="flex min-w-0 flex-wrap items-baseline gap-x-1.5 gap-y-0.5">
          <template v-if="promo">
            <span class="text-[15px] font-bold tabular-nums text-destructive sm:text-base">{{ formatPrice(getPromotionPriceAmount(product), siteCurrency) }}</span>
            <span class="text-xs tabular-nums text-muted-foreground line-through">{{ formatPrice(product.price_amount, siteCurrency) }}</span>
          </template>
          <span v-else class="text-[15px] font-bold tabular-nums text-primary sm:text-base">{{ formatPrice(product.price_amount, siteCurrency) }}</span>
          <span v-if="priceSignal" class="rounded px-1.5 py-px text-[11px] font-medium" :class="priceSignal.tone">{{ priceSignal.label }}</span>
        </div>
        <span v-if="!soldOut" aria-hidden="true" class="luna-gradient mt-2 inline-flex h-9 w-full items-center justify-center gap-1.5 rounded-full text-[13px] font-semibold transition group-hover:brightness-110 group-active:scale-[0.98] sm:hidden">
          <ShoppingCart class="h-4 w-4" />
          {{ t('products.quickBuy') }}
        </span>
        <!-- 售罄占同样的位置，同一排卡片高度不跳 -->
        <span v-else class="mt-2 flex h-9 w-full items-center justify-center rounded-full bg-secondary text-[13px] font-semibold text-muted-foreground sm:hidden">{{ t('products.stockStatus.outOfStock') }}</span>
      </div>

      <div class="flex flex-wrap items-center gap-1.5">
        <!-- 首页看父分类（全部）时标出所属子分类 -->
        <span v-if="categoryLabel" class="max-w-full truncate rounded bg-secondary px-1.5 py-px text-[11px] font-medium text-muted-foreground">{{ categoryLabel }}</span>
        <span class="inline-flex items-center gap-1 rounded px-1.5 py-px text-[11px] font-medium" :class="stockPill.tone">
          <component :is="stockPill.icon" class="h-3 w-3" />
          {{ stockPill.label }}
        </span>
        <!-- 交付方式与商家标签：窄屏两列太挤，只在 sm 起显示 -->
        <span class="hidden rounded px-1.5 py-px text-[11px] font-medium sm:inline" :class="fulfillmentTone">{{ getFulfillmentTypeLabel(product.fulfillment_type) }}</span>
        <template v-if="!soldOut && product.tags && product.tags.length">
          <span v-for="(tag, i) in product.tags.slice(0, 2)" :key="i" class="hidden max-w-full truncate rounded bg-destructive/10 px-1.5 py-px text-[11px] font-medium text-destructive sm:inline">{{ tag }}</span>
        </template>
      </div>
    </div>

    <span
      v-if="!soldOut"
      aria-hidden="true"
      class="luna-gradient hidden h-8 flex-none items-center rounded-full px-3.5 text-[13px] font-semibold transition group-hover:brightness-110 sm:inline-flex"
    >{{ t('products.quickBuy') }}</span>
  </RouterLink>
</template>

<script setup lang="ts">
import { computed, ref, type Component } from 'vue'
import { useI18n } from 'vue-i18n'
import { AlarmClock, CircleCheck, ShoppingCart, XCircle } from 'lucide-vue-next'
import { getFirstImageUrl, getImageUrl } from '../../../utils/image'
import { useLocalized, useProductLabels } from '../../../composables/useProduct'

const props = withDefaults(defineProps<{ product: any; categoryLabel?: string; layout?: 'card' | 'list' }>(), {
  categoryLabel: '',
  layout: 'card',
})

const { t } = useI18n()
const { getLocalizedText, siteCurrency, formatPrice } = useLocalized()
const {
  getStockStatusLabel, getFulfillmentTypeLabel,
  isSoldOut, hasPromotionPrice, getPromotionPriceAmount, hasWholesalePrices, hasPromotionRules,
} = useProductLabels()

const title = computed(() => getLocalizedText(props.product?.title))
const soldOut = computed(() => isSoldOut(props.product))
const promo = computed(() => hasPromotionPrice(props.product))

const imageErrored = ref(false)
const productImage = computed(() => getFirstImageUrl(props.product?.images))
const coverImage = computed(() => {
  if (imageErrored.value) return ''
  if (productImage.value) return productImage.value
  const icon = props.product?.category?.icon
  return icon ? getImageUrl(icon) : ''
})
/** 没有商品图、用的是分类图标：等比放进去，不裁切 */
const coverIsCategoryIcon = computed(() => !productImage.value && Boolean(coverImage.value))
const hasImage = computed(() => Boolean(coverImage.value))

const fulfillmentTone = computed(() => props.product?.fulfillment_type === 'auto' ? 'bg-success/15 text-success' : 'bg-warning/15 text-warning')

const stockPill = computed<{ tone: string; icon: Component; label: string }>(() => {
  if (soldOut.value) {
    return { tone: 'bg-secondary text-muted-foreground', icon: XCircle, label: t('products.stockStatus.outOfStock') }
  }
  if (props.product?.stock_status === 'low_stock') {
    return { tone: 'bg-warning/15 text-warning', icon: AlarmClock, label: getStockStatusLabel(props.product) }
  }
  return { tone: 'bg-primary/10 text-primary', icon: CircleCheck, label: getStockStatusLabel(props.product) }
})

// 价签徽章：促销 / 批发 / 活动（择一，对齐 classic 优先级）
const priceSignal = computed<{ tone: string; label: string } | null>(() => {
  if (promo.value) return { tone: 'bg-destructive/10 text-destructive', label: t('products.promotionTag') }
  if (hasWholesalePrices(props.product)) return { tone: 'bg-destructive/10 text-destructive', label: t('products.wholesaleTag') }
  if (hasPromotionRules(props.product)) return { tone: 'bg-destructive/10 text-destructive', label: t('products.promotionBadge') }
  return null
})
</script>
