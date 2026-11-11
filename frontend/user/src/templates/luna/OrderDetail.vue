<template>
  <div class="mx-auto w-full max-w-[1100px] px-4 pb-8 sm:px-6">
    <nav class="my-3 flex flex-wrap items-center gap-2 text-[13px] text-muted-foreground sm:my-3.5 sm:mt-[18px]">
      <RouterLink to="/" class="hover:text-primary">{{ t('nav.home') }}</RouterLink>
      <span class="text-hairline-strong">/</span>
      <RouterLink to="/me/orders" class="hover:text-primary">{{ t('orders.title') }}</RouterLink>
      <span class="text-hairline-strong">/</span>
      <span class="font-semibold text-foreground">{{ t('orderDetail.title') }}</span>
    </nav>

    <!-- 页头：手机上压成一行，把首屏留给卡密和使用说明 -->
    <header class="mb-3 flex flex-wrap items-center justify-between gap-3 sm:mb-[18px] sm:items-start">
      <div class="min-w-0">
        <h1 class="text-xl font-extrabold sm:mb-1.5 sm:text-3xl">{{ t('orderDetail.title') }}</h1>
        <p class="hidden text-muted-foreground sm:block">{{ t('orderDetail.subtitle') }}</p>
      </div>
      <Button as-child variant="outline" class="h-9 flex-none rounded-full border-primary/40 bg-card px-3.5 font-bold text-primary hover:bg-primary/10 hover:text-primary sm:h-10 sm:px-4" data-testid="continue-shopping">
        <RouterLink to="/products"><ShoppingBag /> {{ t('orders.continueShopping') }}</RouterLink>
      </Button>
    </header>

    <!-- Loading -->
    <div v-if="loading" class="luna-panel p-4 sm:p-[22px]">
      <div class="mb-4 h-5 w-[35%] rounded bg-secondary"></div>
      <div class="h-[200px] rounded-md bg-secondary"></div>
    </div>

    <!-- 不存在 -->
    <div v-else-if="!order" class="my-6 flex flex-col items-center gap-3 rounded-xl border border-dashed py-16 text-center text-muted-foreground">
      <AlertCircle class="h-10 w-10 opacity-60" />
      <p>{{ t('orderDetail.notFound') }}</p>
      <Button class="mt-2 rounded-full" size="sm" @click="debouncedLoadOrder()">{{ t('errorBoundary.retry') }}</Button>
    </div>

    <template v-else>
      <!-- 订单概要：一张小条，订单号 + 状态、金额 + 下单时间，不和卡密抢位置 -->
      <div class="mb-3 luna-panel px-4 py-3 sm:mb-4 sm:px-5 sm:py-3.5">
        <div class="flex items-center justify-between gap-3">
          <div class="min-w-0 text-[13px] sm:text-sm">
            <span class="text-muted-foreground">{{ t('orders.orderNo') }}：</span><span class="font-bold [overflow-wrap:anywhere]">{{ order.order_no }}</span>
          </div>
          <Badge :variant="statusVariant(order.status)" class="flex-none rounded-full">{{ statusLabel(order.status) }}</Badge>
        </div>
        <div class="mt-1.5 flex flex-wrap items-center gap-x-4 gap-y-1.5 text-[13px] text-muted-foreground">
          <span>{{ t('orderDetail.amountTotal') }}：<span class="font-bold tabular-nums text-foreground">{{ formatMoney(order.total_amount, order.currency) }}</span></span>
          <span class="tabular-nums">{{ t('orderDetail.createdAtLabel') }}：{{ formatDate(order.created_at) }}</span>
          <div v-if="order.status === 'pending_payment'" class="ml-auto flex gap-2">
            <Button as-child size="sm" class="rounded-full">
              <RouterLink :to="`/pay?order_no=${order.order_no}`">{{ t('orderDetail.payNow') }}</RouterLink>
            </Button>
            <Button type="button" variant="outline" size="sm" class="rounded-full border-destructive text-destructive hover:bg-destructive/10" @click="cancelOrder">{{ t('orderDetail.cancel') }}</Button>
          </div>
        </div>
      </div>

      <LunaOrderBody
        :order="order"
        variant="user"
        :fulfillment-downloading="fulfillmentDownloading"
        @download="handleDownloadFulfillment"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { AlertCircle, ShoppingBag } from 'lucide-vue-next'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import LunaOrderBody from './components/LunaOrderBody.vue'
import { useOrderDetail } from '../../composables/useOrderDetail'

const { t } = useI18n()

const {
  loading, order, debouncedLoadOrder, cancelOrder, fulfillmentDownloading, handleDownloadFulfillment,
  statusLabel, statusVariant, formatDate, formatMoney,
} = useOrderDetail()
</script>
