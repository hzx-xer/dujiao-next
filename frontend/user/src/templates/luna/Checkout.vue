<template>
  <div class="mx-auto w-full max-w-[1100px] px-3 pb-8 sm:px-6">
    <!-- 页头：步骤条已经交代位置，不再重复面包屑 -->
    <header class="mb-4 mt-5 sm:mb-5 sm:mt-7">
      <h1 class="text-2xl font-extrabold sm:text-[30px]">{{ t('checkout.title') }}</h1>
      <p class="mt-1 text-sm text-muted-foreground">{{ t('checkout.subtitle') }}</p>
    </header>

    <LunaCheckoutSteps current="checkout" :skip-cart="isBuyNowMode" />

    <!-- 空 -->
    <div v-if="cartItems.length === 0" class="my-8 flex flex-col items-center gap-3 rounded-xl border border-dashed py-16 text-center text-muted-foreground">
      <ShoppingCart class="h-10 w-10 opacity-60" />
      <p>{{ t('checkout.empty') }}</p>
      <Button as-child class="mt-2 rounded-full" size="sm"><RouterLink to="/products">{{ t('checkout.emptyAction') }}</RouterLink></Button>
    </div>

    <div v-else class="grid grid-cols-1 items-start gap-4 sm:gap-6 lg:grid-cols-[minmax(0,1fr)_360px]">
      <div ref="mainColumnRef" class="grid min-w-0 grid-cols-1 gap-4">
        <!-- 商品确认 -->
        <section class="luna-panel p-4 sm:p-5">
          <h2 class="mb-3 text-base font-bold sm:text-lg">
            {{ t('checkout.itemsTitle') }}
          </h2>
          <div class="grid grid-cols-1 divide-y">
            <div
              v-for="item in cartItems"
              :key="cartItemKey(item)"
              class="flex gap-3 py-3 first:pt-0 last:pb-0"
              :class="{ 'rounded-md bg-warning/10 px-2.5 !py-2.5': itemStockExceeded(item) }"
            >
              <div class="relative grid h-14 w-14 flex-none place-items-center overflow-hidden rounded-[10px] bg-primary/10 text-primary sm:h-16 sm:w-16">
                <img v-if="checkoutItemImage(item)" :src="checkoutItemImage(item)" :alt="getLocalizedText(item.title)" loading="lazy" class="absolute inset-0 h-full w-full object-cover" />
                <Package v-else class="h-6 w-6" />
              </div>
              <div class="min-w-0 flex-1">
                <RouterLink :to="`/products/${item.slug}`" class="line-clamp-2 break-words text-sm font-semibold leading-snug hover:text-primary sm:text-[15px]">{{ getLocalizedText(item.title) }}</RouterLink>
                <div class="mt-1 flex flex-wrap gap-x-2.5 gap-y-0.5 text-xs text-muted-foreground">
                  <span v-if="skuLabel(item)">{{ skuLabel(item) }}</span>
                  <span>× {{ item.quantity }}</span>
                </div>
                <div v-if="itemStockHint(item)" class="mt-0.5 text-xs" :class="itemStockExceeded(item) ? 'text-warning' : 'text-muted-foreground'">{{ itemStockHint(item) }}</div>
              </div>
              <div class="flex-none text-right">
                <span class="inline-flex items-baseline font-bold" :class="checkoutItemHasPriceDiscount(item) ? 'text-primary' : 'text-foreground'">
                  <b class="text-[17px] sm:text-[19px]">{{ checkoutItemPriceParts(item).integer }}</b><small class="text-xs font-bold">{{ checkoutItemPriceParts(item).decimal }}</small><em class="ml-0.5 text-[11px] font-bold not-italic">{{ checkoutItemCurrency }}</em>
                </span>
                <span v-if="checkoutItemHasPriceDiscount(item)" class="mt-0.5 block text-[11px] text-muted-foreground line-through">
                  {{ checkoutItemOriginalPriceParts(item).integer }}{{ checkoutItemOriginalPriceParts(item).decimal }} {{ checkoutItemCurrency }}
                </span>
              </div>
            </div>
          </div>
        </section>

        <!-- 交付信息（复用公共表单组件） -->
        <!-- 公共组件自带 border + p-6，这里覆盖成与其它区块一致的 luna 面板 -->
        <CheckoutManualForm
          v-if="manualFormProducts.length"
          class="luna-panel !rounded-[14px] !border-0 !p-4 sm:!p-5"
          :manual-form-products="manualFormProducts"
          v-model="manualFormData"
          :submit-attempted="submitAttempted"
          :get-manual-field-label="getManualFieldLabel"
          :get-manual-field-placeholder="getManualFieldPlaceholder"
          :manual-field-error="manualFieldError"
        />

        <!-- 联系方式（游客） -->
        <!-- 点过提交但邮箱 / 密码还不完整：整块描红 + 抖一下 + 顶部提示条 -->
        <section
          v-if="!userAuthStore.isAuthenticated"
          ref="contactRef"
          class="luna-panel scroll-mt-24 p-4 transition-shadow sm:p-5"
          :class="{ 'ring-2 ring-destructive/60': showGuestErrors, 'luna-shake': contactShake }"
        >
          <div class="mb-3.5 flex flex-wrap items-center justify-between gap-2.5">
            <h2 class="flex items-center gap-2 text-base font-bold sm:text-lg">
              {{ t('checkout.modeTitle') }}
              <!-- 游客信息完成度：没填完时一直挂着，填好变绿 -->
              <span v-if="checkoutMode === 'guest' && !guestInfoComplete" class="inline-flex items-center gap-1 rounded-full bg-warning/15 px-2 py-0.5 text-[11.5px] font-semibold text-warning">
                <CircleAlert class="h-3.5 w-3.5" /> {{ t('luna.guestInfoPending') }}
              </span>
              <span v-else-if="checkoutMode === 'guest'" class="inline-flex items-center gap-1 rounded-full bg-success/15 px-2 py-0.5 text-[11.5px] font-semibold text-success">
                <CircleCheck class="h-3.5 w-3.5" /> {{ t('luna.guestInfoDone') }}
              </span>
            </h2>
            <!-- 分段切换：游客 / 登录会员 -->
            <div class="inline-flex rounded-full bg-secondary p-1 text-[13px] font-semibold">
              <button type="button" class="rounded-full px-3.5 py-1.5 transition-colors" :class="checkoutMode === 'guest' ? 'bg-card text-primary shadow-[var(--luna-shadow-sm)]' : 'text-muted-foreground hover:text-foreground'" :aria-pressed="checkoutMode === 'guest'" @click="checkoutMode = 'guest'">{{ t('checkout.guestPurchase') }}</button>
              <RouterLink :to="loginPath" class="rounded-full px-3.5 py-1.5 text-muted-foreground transition-colors hover:text-foreground">{{ t('checkout.memberPurchase') }}</RouterLink>
            </div>
          </div>

          <template v-if="checkoutMode === 'guest'">
            <div v-if="showGuestErrors" class="mb-3.5 flex items-start gap-2 rounded-md bg-destructive/10 px-3 py-2.5 text-[13px] font-semibold text-destructive" role="alert">
              <CircleAlert class="mt-px h-4 w-4 flex-none" />
              <span>{{ t('luna.guestInfoRequired') }}</span>
            </div>
            <div class="grid grid-cols-1 gap-3.5 sm:grid-cols-2">
              <div>
                <label for="luna-guest-email" class="mb-1.5 block text-[13px] font-semibold text-foreground">{{ t('luna.guestEmailLabel') }}<span class="ml-0.5 text-destructive">*</span></label>
                <Input
                  id="luna-guest-email"
                  v-model="guestEmail"
                  type="email"
                  inputmode="email"
                  autocomplete="email"
                  class="h-11"
                  :class="{ 'border-destructive bg-destructive/5 focus-visible:ring-destructive/30': guestEmailError }"
                  :aria-invalid="Boolean(guestEmailError)"
                  aria-describedby="luna-guest-email-hint"
                  :placeholder="t('checkout.guestEmailPlaceholder')"
                />
                <p v-if="guestEmailError" id="luna-guest-email-hint" class="mt-1.5 flex items-center gap-1 text-xs text-destructive"><CircleAlert class="h-3.5 w-3.5 flex-none" />{{ guestEmailError }}</p>
                <p v-else id="luna-guest-email-hint" class="mt-1.5 text-xs text-muted-foreground">{{ t('luna.guestEmailHelp') }}</p>
              </div>
              <div>
                <label for="luna-guest-password" class="mb-1.5 block text-[13px] font-semibold text-foreground">{{ t('luna.guestPasswordLabel') }}<span class="ml-0.5 text-destructive">*</span></label>
                <Input
                  id="luna-guest-password"
                  v-model="guestPassword"
                  type="password"
                  autocomplete="new-password"
                  class="h-11"
                  :class="{ 'border-destructive bg-destructive/5 focus-visible:ring-destructive/30': guestPasswordError }"
                  :aria-invalid="Boolean(guestPasswordError)"
                  aria-describedby="luna-guest-password-hint"
                  :placeholder="t('checkout.guestPasswordPlaceholder')"
                />
                <p v-if="guestPasswordError" id="luna-guest-password-hint" class="mt-1.5 flex items-center gap-1 text-xs text-destructive"><CircleAlert class="h-3.5 w-3.5 flex-none" />{{ guestPasswordError }}</p>
                <p v-else id="luna-guest-password-hint" class="mt-1.5 text-xs text-muted-foreground">{{ t('luna.guestPasswordHelp') }}</p>
              </div>
            </div>

            <div v-if="guestCaptchaEnabled" class="mt-3.5">
              <p class="mb-2 text-[11px] font-bold uppercase tracking-[0.1em] text-muted-foreground">{{ t('auth.common.captchaLabel') }}</p>
              <ImageCaptcha
                v-if="captchaProvider === 'image'"
                ref="guestImageCaptchaRef"
                v-model="guestCaptchaPayload"
                :disabled="submitting"
                @config-stale="handleGuestCaptchaConfigStale"
              />
              <TurnstileCaptcha
                v-else-if="captchaProvider === 'turnstile'"
                ref="guestTurnstileRef"
                v-model="guestTurnstileToken"
                :site-key="guestTurnstileSiteKey"
              />
            </div>
          </template>
        </section>
      </div>

      <!-- 右栏：支付方式 + 金额 + 提交 -->
      <aside ref="summaryRef" class="luna-panel min-w-0 scroll-mt-24 p-4 sm:p-[22px] lg:sticky lg:top-[90px]">
        <!-- 支付方式 -->
        <h2 class="mb-3 text-base font-bold sm:text-lg">{{ t('checkout.paymentMethod') }}</h2>

        <div v-if="showBalanceOption" class="mb-3 rounded-[10px] border bg-secondary p-3.5">
          <div class="flex items-start justify-between gap-2.5">
            <div>
              <div class="text-xs text-muted-foreground">{{ t('payment.walletBalanceLabel') }}</div>
              <div class="mt-0.5 font-bold text-foreground">{{ walletLoading ? t('common.loading') : formatPrice(walletBalance, previewCurrency) }}</div>
            </div>
            <label class="inline-flex items-center gap-1.5 text-xs text-muted-foreground">
              <input v-model="useBalance" type="checkbox" class="h-4 w-4 accent-[var(--ui-accent)]" :disabled="walletOnlyPayment" />
              <span>{{ t('payment.useBalance') }}</span>
            </label>
          </div>
          <div v-if="walletOnlyPayment" class="mt-2 text-xs text-warning">{{ t('payment.walletOnlyHint') }}</div>
          <div v-if="useBalance" class="mt-2.5 grid grid-cols-1 gap-0.5 text-xs text-muted-foreground">
            <div>{{ t('payment.walletDeductLabel') }}：{{ expectedWalletPaidDisplay }}</div>
            <div v-if="!walletOnlyPayment">{{ t('payment.onlinePayLabel') }}：{{ expectedOnlinePayDisplay }}</div>
            <div v-if="walletOnlyPayment && expectedOnlinePayCents > 0" class="text-warning">{{ t('payment.walletInsufficientHint') }}</div>
          </div>
        </div>

        <template v-if="!walletOnlyPayment">
          <LunaPaymentChannels
            v-if="requiresOnlineChannel && paymentChannels.length > 0"
            :channels="paymentChannels"
            :model-value="selectedChannelId"
            :aria-label="t('checkout.paymentMethod')"
            :is-channel-disabled-for-amount="isChannelDisabledForAmount"
            :channel-amount-limit-hint="channelAmountLimitHint"
            @update:model-value="onSelectChannelId"
          />
          <div v-else-if="requiresOnlineChannel && paymentChannels.length === 0" class="text-[13px] text-muted-foreground">{{ t('checkout.noPaymentChannels') }}</div>
        </template>
        <div v-if="!requiresOnlineChannel" class="text-[13px] text-success">{{ t('checkout.walletCoversAll') }}</div>

        <!-- 优惠码：默认收起，填过才展开 -->
        <div v-if="!isResellerTenant" class="mt-4 border-t pt-3.5">
          <button v-if="!couponOpen" type="button" class="inline-flex items-center gap-1.5 text-[13px] font-semibold text-primary hover:underline" @click="couponOpen = true">
            <TicketPercent class="h-4 w-4" /> {{ t('luna.useCoupon') }}
          </button>
          <div v-else>
            <label for="luna-coupon" class="mb-1.5 block text-[13px] font-semibold text-foreground">{{ t('checkout.couponTitle') }}</label>
            <Input id="luna-coupon" v-model="couponCode" type="text" class="h-10" autocomplete="off" :placeholder="t('checkout.couponPlaceholder')" />
          </div>
        </div>

        <!-- 金额：只列出实际有优惠的行 -->
        <div class="mt-4 grid grid-cols-1 gap-2 border-t pt-3.5 text-[13.5px]">
          <div class="flex items-center justify-between"><span class="text-muted-foreground">{{ t('checkout.previewOriginal') }}<span class="ml-1 text-xs">({{ totalItems }})</span></span><span class="font-semibold tabular-nums text-foreground">{{ formatPrice(previewOriginal, previewCurrency) }}</span></div>
          <template v-if="!isResellerTenant">
            <div v-if="hasPositiveAmount(previewCoupon)" class="flex items-center justify-between"><span class="text-muted-foreground">{{ t('checkout.previewCoupon') }}</span><span class="font-semibold tabular-nums text-primary">{{ formatDiscountPrice(previewCoupon, previewCurrency) }}</span></div>
            <div v-if="hasPositiveAmount(previewPromotion)" class="flex items-center justify-between"><span class="text-muted-foreground">{{ t('checkout.previewPromotion') }}</span><span class="font-semibold tabular-nums text-primary">{{ formatDiscountPrice(previewPromotion, previewCurrency) }}</span></div>
            <div v-if="hasPositiveAmount(previewWholesale)" class="flex items-center justify-between"><span class="text-muted-foreground">{{ t('checkout.previewWholesale') }}</span><span class="font-semibold tabular-nums text-success">{{ formatDiscountPrice(previewWholesale, previewCurrency) }}</span></div>
          </template>
          <div v-if="Number(previewMemberDiscount) > 0" class="flex items-center justify-between"><span class="text-muted-foreground">{{ t('checkout.previewMemberDiscount') }}</span><span class="font-semibold tabular-nums text-warning">-{{ formatPrice(previewMemberDiscount, previewCurrency) }}</span></div>
          <div class="mt-1 flex items-baseline justify-between border-t pt-3 font-bold text-foreground">
            <span>{{ t('checkout.previewTotal') }}</span>
            <span class="text-[24px] text-primary tabular-nums">{{ formatPrice(previewTotal, previewCurrency) }}</span>
          </div>
          <p class="text-[11.5px] leading-relaxed text-muted-foreground">{{ t('checkout.submitHint') }}</p>
        </div>

        <div v-if="previewLoading || couponRefreshing" class="mt-2 text-xs text-muted-foreground">{{ previewStatusText }}</div>
        <div v-if="checkoutAlert" class="mt-3 flex items-start gap-2 rounded-md px-3 py-2.5 text-[13px] font-semibold" :class="checkoutAlert.level === 'error' ? 'bg-destructive/10 text-destructive' : 'bg-warning/10 text-warning'">
          <AlertCircle class="mt-px h-4 w-4 flex-none" />
          <span>{{ checkoutAlert.message }}</span>
        </div>

        <!-- 游客信息没填完时给个直达入口（右栏在窄屏排在最下面，离联系方式较远） -->
        <button v-if="guestInfoNeeded" type="button" class="mt-2.5 inline-flex items-center gap-1 text-[13px] font-semibold text-primary hover:underline" @click="goToContact">
          {{ t('luna.guestInfoGoFill') }} <ArrowUp class="h-3.5 w-3.5" />
        </button>

        <!-- 不因「没填完」而禁用：点了才能告诉用户缺什么 -->
        <Button ref="submitButtonRef" class="luna-gradient mt-4 h-12 w-full rounded-full text-base font-bold" :class="{ 'opacity-80': !canSubmit && !submitting }" :disabled="submitting" @click="onSubmit">
          <LoaderCircle v-if="submitting" class="animate-spin" />
          {{ submitting ? t('checkout.submitting') : t('checkout.submitButton') }}
        </Button>
      </aside>
    </div>

    <!-- 窄屏底部固定提交条：右栏提交按钮不在屏幕里时出现；还不能提交时点它滚到第一个要处理的地方 -->
    <Transition
      enter-active-class="transition duration-300 ease-out"
      enter-from-class="translate-y-full opacity-0"
      leave-active-class="transition duration-200 ease-in"
      leave-to-class="translate-y-full opacity-0"
    >
      <div v-if="showMobileBar && cartItems.length" class="theme-safe-bottom fixed bottom-0 left-0 right-0 z-40 border-t bg-card/95 shadow-[var(--luna-shadow)] backdrop-blur-xl lg:hidden">
        <div class="flex items-center gap-3 px-4 py-2.5">
          <div class="min-w-0 flex-1">
            <span class="block text-xs text-muted-foreground">{{ t('checkout.previewTotal') }}</span>
            <span class="block truncate text-xl font-extrabold tabular-nums text-primary">{{ formatPrice(previewTotal, previewCurrency) }}</span>
          </div>
          <Button size="lg" class="luna-gradient flex-none rounded-full px-6 font-bold" :disabled="submitting" @click="onSubmit">
            <LoaderCircle v-if="submitting" class="animate-spin" />
            {{ submitting ? t('checkout.submitting') : t('checkout.submitButton') }}
          </Button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AlertCircle, ArrowUp, CircleAlert, CircleCheck, LoaderCircle, Package, ShoppingCart, TicketPercent } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import ImageCaptcha from '../../components/captcha/ImageCaptcha.vue'
import TurnstileCaptcha from '../../components/captcha/TurnstileCaptcha.vue'
import CheckoutManualForm from '../../components/checkout/CheckoutManualForm.vue'
import LunaCheckoutSteps from './components/LunaCheckoutSteps.vue'
import LunaPaymentChannels from './components/LunaPaymentChannels.vue'
import { useCheckout } from '../../composables/useCheckout'

const { t } = useI18n()
const route = useRoute()

const {
  userAuthStore, getLocalizedText, formatPrice,
  isBuyNowMode, cartItems, totalItems, cartItemKey, checkoutItemImage, itemSkuDisplay,
  itemStockExceeded, itemStockHint,
  checkoutItemCurrency, checkoutItemPriceParts, checkoutItemOriginalPriceParts, checkoutItemHasPriceDiscount,
  manualFormProducts, manualFormData, submitAttempted, getManualFieldLabel, getManualFieldPlaceholder, manualFieldError,
  couponCode, isResellerTenant,
  checkoutMode, guestEmail, guestPassword, guestEmailValid,
  guestCaptchaEnabled, captchaProvider, guestCaptchaPayload, guestTurnstileToken, guestTurnstileSiteKey,
  guestImageCaptchaRef, guestTurnstileRef, handleGuestCaptchaConfigStale,
  previewCurrency, previewOriginal, previewCoupon, previewPromotion, previewWholesale, previewMemberDiscount, previewTotal,
  previewLoading, couponRefreshing, previewStatusText, hasPositiveAmount, formatDiscountPrice, checkoutAlert,
  showBalanceOption, walletLoading, walletBalance, useBalance, walletOnlyPayment,
  expectedWalletPaidDisplay, expectedOnlinePayDisplay, expectedOnlinePayCents,
  requiresOnlineChannel, paymentChannels, selectedChannelId, isChannelDisabledForAmount, channelAmountLimitHint,
  handleSelectChannel,
  submitting, canSubmit, handleSubmit,
} = useCheckout()

// 字符串模板 ref，逻辑在 composable 内，显式标记避免 noUnusedLocals 误报。
void guestImageCaptchaRef
void guestTurnstileRef

/** 没填规格值的「默认规格」不显示，只占地方 */
const skuLabel = (item: Parameters<typeof itemSkuDisplay>[0]) => {
  const label = itemSkuDisplay(item)
  return label === t('productDetail.skuFallback') ? '' : label
}

/** 登录后回到结算页 */
const loginPath = computed(() => `/auth/login?redirect=${encodeURIComponent(route.fullPath)}`)

const couponOpen = ref(Boolean(couponCode.value))

// ==================== 游客信息校验提示 ====================
// 与 useCheckout 里的 GUEST_PASSWORD_MIN_LENGTH 保持一致（该常量未导出）
const GUEST_PASSWORD_MIN_LENGTH = 6
const mainColumnRef = ref<HTMLElement | null>(null)
const contactRef = ref<HTMLElement | null>(null)
/** 点过提交后才把没填的字段标红，避免一进页面就满屏红字 */
const guestAttempted = ref(false)
const contactShake = ref(false)

const isGuestCheckout = computed(() => !userAuthStore.isAuthenticated && checkoutMode.value === 'guest')
const guestInfoComplete = computed(() => guestEmailValid.value && guestPassword.value.trim().length >= GUEST_PASSWORD_MIN_LENGTH)
const guestInfoNeeded = computed(() => isGuestCheckout.value && !guestInfoComplete.value)

const guestEmailError = computed(() => {
  if (!isGuestCheckout.value) return ''
  const email = guestEmail.value.trim()
  if (!email) return guestAttempted.value ? t('luna.guestEmailRequired') : ''
  // 填了但格式不对，不等提交就提示
  return guestEmailValid.value ? '' : t('error.email_invalid')
})
const guestPasswordError = computed(() => {
  if (!isGuestCheckout.value || !guestAttempted.value) return ''
  const password = guestPassword.value.trim()
  if (!password) return t('luna.guestPasswordRequired')
  return password.length < GUEST_PASSWORD_MIN_LENGTH ? t('checkout.errors.guestPasswordTooShort') : ''
})
const showGuestErrors = computed(() => guestAttempted.value && guestInfoNeeded.value)

const shakeContact = async () => {
  contactShake.value = false
  await nextTick()
  contactShake.value = true
  window.setTimeout(() => { contactShake.value = false }, 500)
}

const goToContact = () => {
  guestAttempted.value = true
  contactRef.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  const target = !guestEmailValid.value ? '#luna-guest-email' : '#luna-guest-password'
  window.setTimeout(() => contactRef.value?.querySelector<HTMLInputElement>(target)?.focus({ preventScroll: true }), 350)
  shakeContact()
}

/** 两个提交按钮共用：能提交就提交；不能提交就把用户带到第一个要处理的地方 */
const onSubmit = async () => {
  if (submitting.value) return
  const ready = canSubmit.value
  if (isGuestCheckout.value) guestAttempted.value = true
  await handleSubmit()
  if (ready) return

  await nextTick()
  // 左栏按页面顺序找第一个报错（交付信息表单在联系方式前面）
  const firstError = mainColumnRef.value?.querySelector<HTMLElement>('.text-xs.text-destructive')
  if (firstError && contactRef.value?.contains(firstError)) {
    goToContact()
    return
  }
  if (firstError) {
    firstError.scrollIntoView({ behavior: 'smooth', block: 'center' })
    const field = firstError.parentElement?.querySelector<HTMLElement>('input, select, textarea')
    window.setTimeout(() => field?.focus({ preventScroll: true }), 350)
    return
  }
  // 其余原因（支付方式、库存等）都在右栏提示里
  summaryRef.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

// ==================== 支付渠道 ====================
const onSelectChannelId = (id: number) => {
  const channel = paymentChannels.value.find((c: any) => Number(c.id) === id)
  if (channel) handleSelectChannel(channel)
}

// 还没选渠道时默认选中第一个可用的，少点一下
watch(
  () => [paymentChannels.value, selectedChannelId.value, requiresOnlineChannel.value] as const,
  () => {
    if (selectedChannelId.value || !requiresOnlineChannel.value) return
    const first = paymentChannels.value.find((channel: any) => !isChannelDisabledForAmount(channel))
    if (first) handleSelectChannel(first)
  },
  { immediate: true, deep: true },
)

// ==================== 窄屏底部提交条 ====================
const summaryRef = ref<HTMLElement | null>(null)
const submitButtonRef = ref<{ $el?: HTMLElement } | HTMLElement | null>(null)
const showMobileBar = ref(false)
let observer: IntersectionObserver | null = null

const resolveEl = (target: typeof submitButtonRef.value): HTMLElement | null => {
  if (!target) return null
  if (target instanceof HTMLElement) return target
  return target.$el instanceof HTMLElement ? target.$el : null
}

watch(submitButtonRef, (target) => {
  observer?.disconnect()
  const el = resolveEl(target)
  if (!el) return
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

watch(showMobileBar, (open) => document.body.classList.toggle('luna-mobilebar-open', open))

onUnmounted(() => {
  observer?.disconnect()
  observer = null
  document.body.classList.remove('luna-mobilebar-open')
})
</script>
