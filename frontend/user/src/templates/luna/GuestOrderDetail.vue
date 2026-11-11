<template>
  <div class="mx-auto w-full max-w-[1100px] px-4 pb-8 sm:px-6">
    <!-- 页头：手机上压成一行，把首屏留给卡密和使用说明 -->
    <div class="my-3 flex flex-wrap items-center justify-between gap-x-3 gap-y-2 sm:my-5 sm:items-start">
      <div class="min-w-0">
        <h1 class="text-xl font-extrabold sm:mb-1.5 sm:text-3xl">{{ t('guestOrderDetail.title') }}</h1>
        <p class="hidden text-muted-foreground sm:block">{{ t('guestOrderDetail.subtitle') }}</p>
      </div>
      <div class="flex min-w-0 max-w-full flex-wrap items-center gap-3">
        <RouterLink class="text-[13px] font-semibold text-muted-foreground transition-colors hover:text-primary" to="/guest/orders">{{ t('guestOrderDetail.backSearch') }}</RouterLink>
        <Button as-child variant="outline" class="h-9 flex-none rounded-full border-primary/40 bg-card px-3.5 font-bold text-primary hover:bg-primary/10 hover:text-primary sm:h-10 sm:px-4" data-testid="continue-shopping">
          <RouterLink to="/products"><ShoppingBag /> {{ t('orders.continueShopping') }}</RouterLink>
        </Button>
      </div>
    </div>

    <!-- 游客验证 -->
    <div v-if="viewState === 'auth'" class="mb-[18px] luna-panel p-4 sm:p-[22px]">
      <h2 class="mb-1.5 text-lg font-bold">{{ t('guestOrderDetail.authTitle') }}</h2>
      <p class="mb-3.5 text-[13px] text-muted-foreground">{{ t('guestOrderDetail.authHint') }}</p>
      <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <Input v-model="auth.email" type="email" class="h-11" :placeholder="t('guestOrders.emailPlaceholder')" />
        <Input v-model="auth.order_password" type="password" class="h-11" :placeholder="t('guestOrders.passwordPlaceholder')" />
      </div>
      <div v-if="authError" class="mt-3.5 rounded-sm bg-destructive/10 px-3 py-2.5 text-[13px] font-semibold text-destructive">{{ authError }}</div>
      <div class="mt-3.5 flex flex-wrap gap-3">
        <Button type="button" size="sm" class="rounded-full" @click="handleAuthSubmit">{{ t('guestOrderDetail.authSubmit') }}</Button>
        <Button type="button" variant="ghost" size="sm" class="rounded-full" @click="clearAuth">{{ t('guestOrderDetail.authClear') }}</Button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="viewState === 'loading'" class="luna-panel p-4 sm:p-[22px]">
      <div class="mb-4 h-5 w-[35%] rounded bg-secondary"></div>
      <div class="h-[200px] rounded-md bg-secondary"></div>
    </div>

    <!-- 不存在 -->
    <div v-else-if="viewState === 'empty'" class="my-6 flex flex-col items-center gap-3 rounded-xl border border-dashed py-16 text-center text-muted-foreground">
      <AlertCircle class="h-10 w-10 opacity-60" />
      <p>{{ t('guestOrderDetail.notFound') }}</p>
    </div>

    <template v-else-if="viewState === 'detail' && order">
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
          <Button v-if="order.status === 'pending_payment'" as-child size="sm" class="ml-auto rounded-full">
            <RouterLink :to="`/pay?guest=1&order_no=${order.order_no}`">{{ t('orders.payNow') }}</RouterLink>
          </Button>
        </div>
      </div>

      <LunaOrderBody
        :order="order"
        variant="guest"
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
import { Input } from '@/components/ui/input'
import LunaOrderBody from './components/LunaOrderBody.vue'
import { useGuestOrderDetail } from '../../composables/useGuestOrderDetail'

const { t } = useI18n()

const {
  order, authError, auth, viewState, handleAuthSubmit, clearAuth,
  fulfillmentDownloading, handleDownloadFulfillment,
  statusLabel, statusVariant, formatDate, formatMoney,
} = useGuestOrderDetail()
</script>
