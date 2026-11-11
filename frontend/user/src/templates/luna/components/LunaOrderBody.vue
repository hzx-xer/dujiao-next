<template>
  <div class="grid grid-cols-1 gap-[18px]">
    <!-- 只有一个子订单：它的交付内容就是整单的交付内容，不再套一层子订单卡片，也不加「交付信息」标题，使用说明直接置顶；商品放到下方订单明细 -->
    <section v-if="singleChild" class="luna-panel p-3 sm:p-[22px]">
      <LunaOrderFulfillment
        v-if="singleChild.fulfillment"
        :fulfillment="singleChild.fulfillment"
        :items="singleChild.items"
        :order-no="singleChild.order_no || order.order_no"
        :downloading="fulfillmentDownloading"
        @download="emit('download', $event)"
      />
      <div v-else class="rounded-lg border border-dashed bg-secondary px-4 py-3.5 text-sm text-muted-foreground">{{ t('orderDetail.childFulfillmentEmpty') }}</div>
    </section>

    <!-- 多个子订单：每个子订单先给交付内容，再列它包含的商品，方便对上哪条卡密属于哪件商品 -->
    <section v-else-if="order.children && order.children.length > 1" class="luna-panel p-3 sm:p-[22px]">
      <h2 class="mb-3 text-base font-bold sm:mb-4 sm:text-lg">{{ t('orderDetail.childOrdersTitle') }}</h2>
      <div class="grid grid-cols-1 gap-4">
        <div v-for="child in order.children" :key="child.id" class="rounded-xl border p-3 sm:p-4">
          <div class="mb-3 flex flex-wrap items-start justify-between gap-2">
            <div class="min-w-0 text-[13px] text-muted-foreground">
              <div class="[overflow-wrap:anywhere]">{{ t('orderDetail.childOrderNo') }}：<span class="font-semibold text-foreground">{{ child.order_no }}</span></div>
              <div>{{ t('orderDetail.childOrderAmount') }}：<span class="font-semibold tabular-nums text-foreground">{{ formatMoney(child.total_amount, child.currency || order.currency) }}</span></div>
            </div>
            <Badge :variant="statusVariant(resolvedChildStatus(child))" class="rounded-full">{{ statusLabel(resolvedChildStatus(child)) }}</Badge>
          </div>

          <LunaOrderFulfillment
            v-if="child.fulfillment"
            :fulfillment="child.fulfillment"
            :items="child.items"
            :order-no="child.order_no || order.order_no"
            :downloading="fulfillmentDownloading"
            @download="emit('download', $event)"
          />
          <div v-else class="rounded-lg border border-dashed bg-secondary px-4 py-3.5 text-sm text-muted-foreground">{{ t('orderDetail.childFulfillmentEmpty') }}</div>

          <h3 class="mb-2.5 mt-4 text-sm font-bold">{{ t('orderDetail.childItemsTitle') }}</h3>
          <div v-if="child.items && child.items.length" class="grid grid-cols-1">
            <LunaOrderItem v-for="(item, cidx) in child.items" :key="cidx" :item="item" :currency="child.currency || order.currency" />
          </div>
          <div v-else class="text-[13px] text-muted-foreground">{{ t('orderDetail.noItems') }}</div>
        </div>
      </div>
    </section>

    <!-- 主订单发货（含卡密） -->
    <section v-if="order.fulfillment" class="luna-panel p-3 sm:p-[22px]">
      <LunaOrderFulfillment
        :fulfillment="order.fulfillment"
        :items="order.items"
        :order-no="order.order_no"
        :downloading="fulfillmentDownloading"
        @download="emit('download', $event)"
      />
    </section>

    <!-- 商品（多个子订单时已在各子订单里列出，这里不再重复） -->
    <section v-if="showOrderItems" class="luna-panel p-3 sm:p-[22px]">
      <h2 class="mb-3 text-base font-bold sm:mb-4 sm:text-lg">{{ t('orderDetail.itemsTitle') }}</h2>
      <div v-if="order.items && order.items.length > 0" class="grid grid-cols-1">
        <LunaOrderItem v-for="(item, idx) in order.items" :key="idx" :item="item" :currency="order.currency" />
      </div>
      <div v-else class="text-[13px] text-muted-foreground">{{ t('orderDetail.noItems') }}</div>
    </section>

    <!-- 金额与时间：放在最后的小明细表，字号和行距都压小，宽屏左金额右时间，窄屏上下排 -->
    <section class="luna-panel px-4 py-3 sm:px-5 sm:py-4">
      <h2 class="mb-1 text-sm font-bold">{{ t('orderDetail.summaryTitle') }}</h2>
      <div class="grid grid-cols-1 gap-x-10 text-xs sm:grid-cols-2 sm:text-[13px]">
        <dl class="divide-y divide-dashed">
          <div class="flex items-baseline justify-between gap-3 py-1.5"><dt class="text-muted-foreground">{{ t('orderDetail.amountOriginal') }}</dt><dd class="font-semibold tabular-nums">{{ formatMoney(order.original_amount, order.currency) }}</dd></div>
          <div v-if="hasDiscountAmount(order.discount_amount)" class="flex items-baseline justify-between gap-3 py-1.5"><dt class="text-muted-foreground">{{ t('orderDetail.amountDiscount') }}</dt><dd class="font-semibold tabular-nums text-destructive">{{ formatDiscountMoney(order.discount_amount, order.currency) }}</dd></div>
          <div v-if="hasDiscountAmount(order.member_discount_amount)" class="flex items-baseline justify-between gap-3 py-1.5"><dt class="text-muted-foreground">{{ t('orderDetail.amountMemberDiscount') }}</dt><dd class="font-semibold tabular-nums text-warning">{{ formatDiscountMoney(order.member_discount_amount, order.currency) }}</dd></div>
          <div v-if="hasDiscountAmount(order.wholesale_discount_amount)" class="flex items-baseline justify-between gap-3 py-1.5"><dt class="text-muted-foreground">{{ t('orderDetail.amountWholesaleDiscount') }}</dt><dd class="font-semibold tabular-nums text-success">{{ formatDiscountMoney(order.wholesale_discount_amount, order.currency) }}</dd></div>
          <template v-if="variant === 'user'">
            <div v-if="hasAmount(order.wallet_paid_amount)" class="flex items-baseline justify-between gap-3 py-1.5"><dt class="text-muted-foreground">{{ t('orderDetail.amountWalletPaid') }}</dt><dd class="font-semibold tabular-nums">{{ formatMoney(order.wallet_paid_amount, order.currency) }}</dd></div>
            <div v-if="hasAmount(order.online_paid_amount)" class="flex items-baseline justify-between gap-3 py-1.5"><dt class="text-muted-foreground">{{ t('orderDetail.amountOnlinePaid') }}</dt><dd class="font-semibold tabular-nums">{{ formatMoney(order.online_paid_amount, order.currency) }}</dd></div>
            <div v-if="hasAmount(order.refunded_amount)" class="flex items-baseline justify-between gap-3 py-1.5"><dt class="text-muted-foreground">{{ t('orderDetail.amountRefunded') }}</dt><dd class="font-semibold tabular-nums">{{ formatMoney(order.refunded_amount, order.currency) }}</dd></div>
          </template>
          <div class="flex items-baseline justify-between gap-3 py-1.5"><dt class="font-bold text-foreground">{{ t('orderDetail.amountTotal') }}</dt><dd class="font-extrabold tabular-nums text-primary">{{ formatMoney(order.total_amount, order.currency) }}</dd></div>
        </dl>
        <dl class="divide-y divide-dashed border-t border-dashed sm:border-t-0">
          <div class="flex items-baseline justify-between gap-3 py-1.5"><dt class="text-muted-foreground">{{ t('orderDetail.createdAtLabel') }}</dt><dd class="font-semibold tabular-nums">{{ formatDate(order.created_at) }}</dd></div>
          <div v-if="order.paid_at" class="flex items-baseline justify-between gap-3 py-1.5"><dt class="text-muted-foreground">{{ t('orderDetail.paidAtLabel') }}</dt><dd class="font-semibold tabular-nums">{{ formatDate(order.paid_at) }}</dd></div>
          <!-- 过期时间是支付截止时间，付款后没有意义，只在未付款时显示 -->
          <div v-if="order.expires_at && !order.paid_at" class="flex items-baseline justify-between gap-3 py-1.5"><dt class="text-muted-foreground">{{ t('orderDetail.expiresAtLabel') }}</dt><dd class="font-semibold tabular-nums">{{ formatDate(order.expires_at) }}</dd></div>
          <div v-if="order.canceled_at" class="flex items-baseline justify-between gap-3 py-1.5"><dt class="text-muted-foreground">{{ t('orderDetail.canceledAtLabel') }}</dt><dd class="font-semibold tabular-nums">{{ formatDate(order.canceled_at) }}</dd></div>
        </dl>
      </div>
    </section>

    <!-- 退款记录 -->
    <section v-if="showRefundRecordsCard" class="luna-panel p-3 sm:p-[22px]">
      <h2 class="mb-3 text-base font-bold sm:mb-4 sm:text-lg">{{ t('orderDetail.refundRecordsTitle') }}</h2>
      <div v-if="refundRecords.length > 0" class="overflow-hidden rounded-md border">
        <div class="grid grid-cols-[1.2fr_1fr_1.6fr] gap-3 bg-secondary px-3.5 py-2.5 text-xs font-bold text-muted-foreground max-[640px]:hidden">
          <span>{{ t('orderDetail.refundRecordTime') }}</span>
          <span>{{ t('orderDetail.refundRecordAmount') }}</span>
          <span>{{ t('orderDetail.refundRecordReason') }}</span>
        </div>
        <div v-for="(record, idx) in refundRecords" :key="`refund-${idx}`" class="grid grid-cols-[1.2fr_1fr_1.6fr] gap-3 border-t px-3.5 py-2.5 text-[13px] max-[640px]:grid-cols-1 max-[640px]:gap-1">
          <span class="text-muted-foreground">{{ formatDate(record.created_at) }}</span>
          <span class="tabular-nums">{{ formatMoney(record.amount, record.currency || order.currency) }}</span>
          <span class="whitespace-pre-wrap break-words text-muted-foreground">{{ refundReasonText(record.remark) }}</span>
        </div>
      </div>
      <div v-else class="text-[13px] text-muted-foreground">{{ t('orderDetail.refundRecordsEmpty') }}</div>
    </section>

  </div>
</template>

<script setup lang="ts">
import { computed, toRef } from 'vue'
import { useI18n } from 'vue-i18n'
import { Badge } from '@/components/ui/badge'
import LunaOrderItem from './LunaOrderItem.vue'
import LunaOrderFulfillment from './LunaOrderFulfillment.vue'
import { useOrderDisplayHelpers } from '../../../composables/useOrderDisplayHelpers'

const props = defineProps<{
  order: any
  variant: 'user' | 'guest'
  fulfillmentDownloading: boolean
}>()

const emit = defineEmits<{
  (e: 'download', orderNo: string): void
}>()

const { t } = useI18n()

const {
  showRefundRecordsCard, refundRecords, resolvedChildStatus,
  statusLabel, statusVariant, formatDate, refundReasonText,
  formatMoney, formatDiscountMoney, hasDiscountAmount, hasAmount,
} = useOrderDisplayHelpers(toRef(props, 'order'))

const children = computed<any[]>(() => (Array.isArray(props.order?.children) ? props.order.children : []))
const singleChild = computed(() => (children.value.length === 1 ? children.value[0] : null))
// 多个子订单且每个都带了商品时，商品已在各子订单里列出，不再单独渲染订单明细
const showOrderItems = computed(() => !(children.value.length > 1 && children.value.every((child) => child?.items?.length)))
</script>
