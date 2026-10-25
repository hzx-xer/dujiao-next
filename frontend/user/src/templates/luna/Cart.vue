<template>
  <div class="mx-auto w-full max-w-[1100px] px-3 pb-8 sm:px-6">
    <!-- 页头：步骤条已经交代位置，不再重复面包屑 -->
    <header class="mb-4 mt-5 sm:mb-5 sm:mt-7">
      <h1 class="text-2xl font-extrabold sm:text-[30px]">{{ t('cart.title') }}</h1>
      <p class="mt-1 text-sm text-muted-foreground">{{ t('cart.subtitle') }}</p>
    </header>

    <LunaCheckoutSteps current="cart" />

    <!-- 空购物车 -->
    <div v-if="cartItems.length === 0" class="my-8 flex flex-col items-center gap-3 rounded-xl border border-dashed py-16 text-center text-muted-foreground">
      <ShoppingCart class="h-10 w-10 opacity-60" />
      <p>{{ t('cart.empty') }}</p>
      <Button as-child class="mt-2 rounded-full" size="sm"><RouterLink to="/products">{{ t('cart.emptyAction') }}</RouterLink></Button>
    </div>

    <div v-else class="grid grid-cols-1 items-start gap-4 sm:gap-6 lg:grid-cols-[minmax(0,1fr)_340px]">
      <!-- 商品列表：窄屏压成「小图 + 标题 / 规格·单价 / 数量·小计」三行 -->
      <div class="grid min-w-0 grid-cols-1 gap-3 sm:gap-4">
        <article v-for="item in cartItems" :key="cartItemKey(item)" class="luna-panel flex gap-3 p-3 sm:gap-4 sm:p-4">
          <RouterLink :to="`/products/${item.slug}`" class="relative grid h-16 w-16 flex-none place-items-center overflow-hidden rounded-[10px] bg-primary/10 text-primary sm:h-[88px] sm:w-[88px]">
            <img v-if="cartItemImage(item)" :src="cartItemImage(item)" :alt="getLocalizedText(item.title)" loading="lazy" class="absolute inset-0 h-full w-full object-cover" />
            <Package v-else class="h-7 w-7 sm:h-10 sm:w-10" />
          </RouterLink>

          <div class="min-w-0 flex-1">
            <div class="flex items-start justify-between gap-2">
              <RouterLink :to="`/products/${item.slug}`" class="line-clamp-2 min-w-0 break-words text-sm font-semibold leading-snug hover:text-primary sm:text-[15px] sm:font-bold">{{ getLocalizedText(item.title) }}</RouterLink>
              <button type="button" class="-mr-1 -mt-1 grid h-8 w-8 flex-none place-items-center rounded-md text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive sm:h-9 sm:w-9" :aria-label="t('cart.remove')" :title="t('cart.remove')" @click="removeWithUndo(item)">
                <Trash2 class="h-4 w-4" />
              </button>
            </div>
            <!-- 规格 · 单价 合成一行 -->
            <p class="mt-1 flex flex-wrap gap-x-2 text-xs text-muted-foreground sm:text-[13px]">
              <span v-if="skuLabel(item)" class="min-w-0 break-words">{{ skuLabel(item) }}</span>
              <span class="tabular-nums">{{ formatPrice(item.priceAmount, totalCurrency) }}</span>
            </p>
            <p v-if="itemStockHint(item)" class="mt-0.5 text-xs text-muted-foreground">{{ itemStockHint(item) }}</p>
            <!-- 购买身份 / 交付方式标签在窄屏省掉，宽屏保留 -->
            <div class="mt-2 hidden flex-wrap gap-1.5 sm:flex">
              <Badge :variant="item.purchaseType === 'guest' ? 'warning' : 'info'" size="sm" class="rounded-full">{{ item.purchaseType === 'guest' ? t('productPurchase.guest') : t('productPurchase.member') }}</Badge>
              <Badge variant="success" size="sm" class="rounded-full">{{ item.fulfillmentType === 'auto' ? t('products.fulfillmentType.auto') : t('products.fulfillmentType.manual') }}</Badge>
            </div>

            <div class="mt-2 flex items-center justify-between gap-3 sm:mt-3">
              <div class="inline-flex items-center overflow-hidden rounded-full border-2 border-hairline-strong">
                <button type="button" class="grid h-8 w-8 place-items-center bg-card text-foreground disabled:opacity-35 sm:h-9 sm:w-9" :aria-label="t('luna.decreaseQuantity')" :disabled="item.quantity <= itemPurchaseMin(item)" @click="updateQty(item, item.quantity - 1)"><Minus class="h-3.5 w-3.5" /></button>
                <input
                  type="number"
                  inputmode="numeric"
                  class="h-8 w-10 border-x bg-card text-center text-sm font-bold tabular-nums outline-none [appearance:textfield] sm:h-9 sm:w-11 [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
                  :value="item.quantity"
                  :min="itemPurchaseMin(item)"
                  :max="itemMaxQuantity(item)"
                  :aria-label="t('productDetail.quantity')"
                  @change="handleQtyChange(item, $event)"
                />
                <button type="button" class="grid h-8 w-8 place-items-center bg-card text-foreground disabled:opacity-35 sm:h-9 sm:w-9" :aria-label="t('luna.increaseQuantity')" :disabled="item.quantity >= itemMaxQuantity(item)" @click="updateQty(item, item.quantity + 1)"><Plus class="h-3.5 w-3.5" /></button>
              </div>
              <strong class="text-[15px] tabular-nums text-foreground sm:text-base">{{ itemSubtotal(item) }}</strong>
            </div>

            <div v-if="quantityWarning(item)" class="mt-2 rounded-md bg-warning/10 px-2.5 py-1.5 text-xs font-semibold text-warning">{{ quantityWarning(item) }}</div>
          </div>
        </article>
      </div>

      <!-- 汇总 -->
      <aside class="luna-panel p-4 sm:p-[22px] lg:sticky lg:top-[90px]">
        <h2 class="mb-4 text-lg font-bold">{{ t('cart.summaryTitle') }}</h2>
        <div class="grid grid-cols-1 gap-3">
          <div class="flex items-center justify-between text-sm text-muted-foreground">
            <span>{{ t('cart.itemsCount') }}</span>
            <span class="font-bold text-foreground">{{ totalItems }}</span>
          </div>
          <div class="flex items-center justify-between border-t pt-3 font-bold text-foreground">
            <span>{{ t('cart.totalLabel') }}</span>
            <span class="text-[22px] text-primary tabular-nums">{{ formatPrice(totalAmount, totalCurrency) }}</span>
          </div>
        </div>
        <p class="my-4 rounded-sm border bg-secondary px-3 py-2.5 text-xs leading-relaxed text-muted-foreground">{{ t('cart.disclaimer') }}</p>
        <div ref="checkoutActionRef">
          <Button as-child class="luna-gradient h-11 w-full rounded-full font-bold"><RouterLink to="/checkout">{{ t('cart.checkout') }} <ArrowRight /></RouterLink></Button>
        </div>
        <Button as-child variant="outline" class="mt-2.5 h-11 w-full rounded-full font-bold"><RouterLink to="/products">{{ t('cart.emptyAction') }}</RouterLink></Button>
      </aside>
    </div>

    <!-- 窄屏底部固定结算条：汇总区的「去结算」不在屏幕里时出现 -->
    <Transition
      enter-active-class="transition duration-300 ease-out"
      enter-from-class="translate-y-full opacity-0"
      leave-active-class="transition duration-200 ease-in"
      leave-to-class="translate-y-full opacity-0"
    >
      <div v-if="showMobileBar && cartItems.length" class="theme-safe-bottom fixed bottom-0 left-0 right-0 z-40 border-t bg-card/95 shadow-[var(--luna-shadow)] backdrop-blur-xl lg:hidden">
        <div class="flex items-center gap-3 px-4 py-2.5">
          <div class="min-w-0 flex-1">
            <span class="block text-xs text-muted-foreground">{{ t('cart.totalLabel') }} · {{ t('cart.itemsCount') }} {{ totalItems }}</span>
            <span class="block truncate text-xl font-extrabold tabular-nums text-primary">{{ formatPrice(totalAmount, totalCurrency) }}</span>
          </div>
          <Button as-child size="lg" class="luna-gradient flex-none rounded-full px-6 font-bold">
            <RouterLink to="/checkout">{{ t('cart.checkout') }} <ArrowRight /></RouterLink>
          </Button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ArrowRight, Minus, Package, Plus, ShoppingCart, Trash2 } from 'lucide-vue-next'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import LunaCheckoutSteps from './components/LunaCheckoutSteps.vue'
import { useCart } from '../../composables/useCart'

const { t } = useI18n()

const {
  getLocalizedText, formatPrice, totalCurrency,
  cartItems, totalItems, totalAmount,
  cartItemKey, cartItemImage, itemSkuDisplay, itemSubtotal, itemStockHint, quantityWarning,
  itemPurchaseMin, itemMaxQuantity,
  removeWithUndo, updateQty, handleQtyChange,
} = useCart()

/** 没填规格值的「默认规格」不显示，只占地方 */
const skuLabel = (item: Parameters<typeof itemSkuDisplay>[0]) => {
  const label = itemSkuDisplay(item)
  return label === t('productDetail.skuFallback') ? '' : label
}

// ==================== 窄屏底部结算条 ====================
const checkoutActionRef = ref<HTMLElement | null>(null)
const showMobileBar = ref(false)
let observer: IntersectionObserver | null = null

watch(checkoutActionRef, (el) => {
  observer?.disconnect()
  if (!el) {
    showMobileBar.value = false
    return
  }
  observer = new IntersectionObserver(
    (entries) => {
      const entry = entries[0]
      // 按钮露出一半以上才算看得见；顶部扣掉 64px 固定顶栏
      if (entry) showMobileBar.value = entry.intersectionRatio < 0.5
    },
    { threshold: [0, 0.5, 1], rootMargin: '-64px 0px 0px 0px' },
  )
  observer.observe(el)
})

// 结算条出现时给页脚留出底部空间（见 styles/luna.css）
watch(showMobileBar, (open) => document.body.classList.toggle('luna-mobilebar-open', open))

onUnmounted(() => {
  observer?.disconnect()
  observer = null
  document.body.classList.remove('luna-mobilebar-open')
})
</script>
