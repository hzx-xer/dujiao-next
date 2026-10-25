<template>
  <Transition
    enter-active-class="transition duration-300 ease-out"
    enter-from-class="translate-y-full opacity-0"
    enter-to-class="translate-y-0 opacity-100"
    leave-active-class="transition duration-200 ease-in"
    leave-from-class="translate-y-0 opacity-100"
    leave-to-class="translate-y-full opacity-0"
  >
    <div
      v-if="visible"
      class="theme-safe-bottom fixed bottom-0 left-0 right-0 z-40 border-t bg-card/95 shadow-[var(--luna-shadow)] backdrop-blur-xl lg:hidden"
    >
      <div class="flex items-center gap-2.5 px-4 py-2.5">
        <!-- 价格 + 已选规格 / 数量 -->
        <div class="min-w-0 flex-1">
          <span v-if="showMemberPrice" class="block truncate text-xl font-extrabold tabular-nums text-warning">{{ memberPriceDisplay }}</span>
          <span v-else-if="showSkuPromotionPrice" class="block truncate text-xl font-extrabold tabular-nums text-destructive">{{ skuPromotionPriceDisplay }}</span>
          <span v-else-if="showSkuPrice" class="block truncate text-xl font-extrabold tabular-nums text-primary">{{ skuPriceDisplay }}</span>
          <span v-else-if="showProductPromotionPrice" class="block truncate text-xl font-extrabold tabular-nums text-destructive">{{ productPromotionPriceDisplay }}</span>
          <span v-else class="block truncate text-xl font-extrabold tabular-nums text-primary">{{ productPriceDisplay }}</span>
          <span v-if="summary" class="block truncate text-xs text-muted-foreground">{{ summary }}</span>
        </div>
        <!-- 操作：加购只放图标，把宽度留给主按钮 -->
        <Button v-if="soldOut" variant="secondary" size="lg" disabled class="flex-none rounded-full font-bold">{{ t('productDetail.skuStockOut') }}</Button>
        <Button v-else-if="requiresLogin" size="lg" class="luna-gradient flex-none rounded-full font-bold" @click="$emit('goLogin')">{{ t('productDetail.loginToBuy') }}</Button>
        <template v-else>
          <Button variant="outline" size="icon-lg" class="flex-none rounded-full" :aria-label="t('productDetail.addToCart')" :title="t('productDetail.addToCart')" :disabled="!canPurchase" @click="$emit('addToCart')">
            <ShoppingCart />
          </Button>
          <Button size="lg" class="luna-gradient flex-none rounded-full px-6 font-bold" :disabled="!canPurchase" @click="$emit('buyNow')">
            <Zap /> {{ t('productDetail.buyNow') }}
          </Button>
        </template>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { ShoppingCart, Zap } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

const { t } = useI18n()

defineProps<{
  visible: boolean
  requiresLogin: boolean
  canPurchase: boolean
  soldOut?: boolean
  summary?: string
  showMemberPrice: boolean
  memberPriceDisplay: string
  showSkuPromotionPrice: boolean
  skuPromotionPriceDisplay: string
  showSkuPrice: boolean
  skuPriceDisplay: string
  showProductPromotionPrice: boolean
  productPromotionPriceDisplay: string
  productPriceDisplay: string
}>()

defineEmits<{
  addToCart: []
  buyNow: []
  goLogin: []
}>()
</script>
