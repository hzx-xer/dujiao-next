<template>
  <div class="mx-auto w-full max-w-[1100px] px-3 pb-6 sm:px-6">
    <!-- Loading -->
    <div v-if="loading" class="grid grid-cols-1 gap-11 py-2.5 lg:grid-cols-2">
      <div class="h-[180px] rounded-xl bg-secondary opacity-50 sm:h-[380px]"></div>
      <div class="grid grid-cols-1 content-start gap-4 pt-2">
        <div class="h-7 w-3/5 rounded-md border bg-card"></div>
        <div class="h-12 w-2/5 rounded-md border bg-card"></div>
        <div class="h-[120px] rounded-md border bg-card"></div>
      </div>
    </div>

    <!-- Content -->
    <template v-else-if="product">
      <!-- 面包屑：首页 › 所属分类 › 商品（商品名单行截断，完整标题在下面的 h1） -->
      <nav class="flex min-w-0 items-center gap-1.5 py-4 pb-2 text-[13px] font-semibold text-muted-foreground sm:py-5 sm:pb-2 sm:text-[13.5px]">
        <RouterLink to="/" class="flex-none hover:text-primary">{{ t('nav.home') }}</RouterLink>
        <ChevronRight class="h-4 w-4 flex-none" />
        <RouterLink :to="categoryPath" class="max-w-[40%] flex-none truncate hover:text-primary">{{ categoryName || t('nav.products') }}</RouterLink>
        <ChevronRight class="h-4 w-4 flex-none" />
        <span class="min-w-0 truncate text-foreground">{{ getLocalizedText(product.title) }}</span>
      </nav>

      <!-- 宽屏购买按钮统一在右栏：有图 = 左图 / 右标题+购买，详情在下方；无图 = 左标题+详情 / 右吸顶购买卡 -->
      <section class="luna-panel grid grid-cols-1 gap-y-4 p-4 sm:gap-y-5 sm:p-6 lg:gap-x-9" :class="sectionGridClass">
        <div v-if="images.length" class="min-w-0 lg:sticky lg:top-24 lg:row-span-2 lg:self-start">
          <div class="relative grid aspect-[4/3] place-items-center overflow-hidden rounded-xl bg-secondary">
            <img v-if="currentImage" :src="currentImage" :alt="getLocalizedText(product.title)" class="absolute inset-0 h-full w-full object-cover" />
          </div>
          <div v-if="images.length > 1" class="luna-no-scrollbar mt-3 flex gap-2.5 overflow-x-auto p-0.5 sm:mt-3.5 sm:flex-wrap sm:gap-3 sm:overflow-visible">
            <button
              v-for="(img, idx) in images"
              :key="idx"
              type="button"
              class="h-[54px] w-[66px] flex-none overflow-hidden rounded-md border-2 bg-secondary sm:h-[60px] sm:w-[74px]"
              :class="img === currentImage ? 'border-primary' : 'border-border'"
              :aria-pressed="img === currentImage"
              @click="currentImage = img"
            >
              <img :src="img" :alt="`${idx + 1}`" loading="lazy" class="h-full w-full object-cover" />
            </button>
          </div>
        </div>

        <!-- 标题区：有图时在图片右侧上方，无图时在左栏顶部 -->
        <div class="min-w-0" :class="headerPlaceClass">
          <div class="flex items-start gap-3">
            <!-- 无商品图：标题旁放小标识，优先用分类图标 -->
            <span v-if="!images.length" class="relative grid h-14 w-14 flex-none place-items-center overflow-hidden rounded-xl sm:h-16 sm:w-16" :class="categoryIcon ? 'bg-primary/10' : 'bg-[image:var(--luna-gradient)] text-white'">
              <img v-if="categoryIcon" :src="categoryIcon" alt="" class="absolute inset-0 h-full w-full object-contain p-2" @error="categoryIconErrored = true" />
              <Package v-else class="h-7 w-7" />
            </span>
            <div class="min-w-0 flex-1">
              <span v-if="categoryName" class="block truncate text-[13px] font-semibold text-muted-foreground">{{ categoryName }}</span>
              <h1 class="mb-3 mt-1 break-words text-[19px] font-bold leading-snug sm:my-1.5 sm:mb-3 sm:text-[22px] lg:text-[24px]">{{ getLocalizedText(product.title) }}</h1>
            </div>
          </div>

          <div class="flex flex-wrap gap-2" :class="{ 'mb-1.5': product.tags && product.tags.length }">
            <span class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[12.5px] font-semibold" :class="stockPillTone">{{ getStockStatusLabel(product) }}</span>
            <span class="inline-flex items-center gap-1.5 rounded-full bg-success/15 px-2.5 py-1 text-[12.5px] font-semibold text-success">
              <component :is="product.fulfillment_type === 'auto' ? Zap : Pencil" class="h-3.5 w-3.5" />
              {{ getFulfillmentTypeLabel(product.fulfillment_type) }}
            </span>
            <span class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[12.5px] font-semibold" :class="product.purchase_type === 'guest' ? 'bg-warning/15 text-warning' : 'bg-success/15 text-success'">
              <component :is="product.purchase_type === 'guest' ? UserPlus : Lock" class="h-3.5 w-3.5" />
              {{ getPurchaseTypeLabel(product.purchase_type) }}
            </span>
          </div>

          <div v-if="product.tags && product.tags.length" class="flex flex-wrap gap-1.5">
            <span v-for="(tag, i) in product.tags" :key="i" class="inline-flex items-center rounded-full bg-secondary px-2.5 py-1 text-[12.5px] font-semibold text-muted-foreground">{{ tag }}</span>
          </div>

        </div>

        <!-- 购买区：始终在右栏（与结算页右侧结算栏一致）；无图时做成吸顶卡片 -->
        <div class="min-w-0" :class="purchasePlaceClass">
          <!-- 价格 -->
          <div class="mb-4 border-b pb-4 sm:mb-5 sm:pb-5">
            <div class="mb-2.5 flex flex-wrap items-center gap-2">
              <span class="text-[13px] font-semibold text-muted-foreground">{{ t('products.price') }}</span>
              <span v-if="(selectedSku && hasSkuPromotionPrice(selectedSku)) || (!selectedSku && hasPromotionPrice(product))" class="inline-flex items-center rounded-full bg-primary/10 px-2.5 py-1 text-[12.5px] font-semibold text-primary">{{ t('products.promotionTag') }}</span>
              <span v-if="showSelectedSkuMemberBadge" class="inline-flex items-center rounded-full bg-warning/15 px-2.5 py-1 text-[12.5px] font-semibold text-warning">{{ t('products.memberPriceTag') }}</span>
              <span v-if="hasSelectedSkuWholesalePrice" class="inline-flex items-center rounded-full bg-success/15 px-2.5 py-1 text-[12.5px] font-semibold text-success">{{ t('products.wholesaleTag') }}</span>
            </div>

            <!-- 1. SKU 批发价 -->
            <template v-if="selectedSku && hasSelectedSkuWholesalePrice">
              <div class="flex flex-wrap items-baseline gap-3">
                <span class="text-[32px] font-bold tabular-nums" :class="selectedSkuWholesaleFinalIsMember ? 'text-warning' : 'text-success'">{{ formatPrice(selectedSkuWholesaleFinalPrice!, siteCurrency) }}</span>
                <span class="font-semibold text-muted-foreground line-through">{{ formatPrice(selectedSku.price_amount, siteCurrency) }}</span>
              </div>
              <p class="mt-2 text-sm font-semibold" :class="selectedSkuWholesaleFinalIsMember ? 'text-warning' : 'text-success'">
                {{ selectedSkuWholesaleFinalIsMember ? t('products.memberPriceTag') : t('products.wholesaleTag') }} · {{ t('products.saveAmount') }} {{ formatPrice(Number(selectedSku.price_amount) - Number(selectedSkuWholesaleFinalPrice), siteCurrency) }}
              </p>
            </template>
            <!-- 2. SKU 促销价 -->
            <template v-else-if="selectedSku && hasSkuPromotionPrice(selectedSku)">
              <div class="flex flex-wrap items-baseline gap-3">
                <span v-if="selectedSkuPromotionFinalIsMember" class="text-[32px] font-bold tabular-nums text-warning">{{ formatPrice(selectedSkuPromotionFinalPrice!, siteCurrency) }}</span>
                <span v-else class="text-[32px] font-bold tabular-nums text-destructive">{{ formatPrice(selectedSkuPromotionPrice!, siteCurrency) }}</span>
                <span class="font-semibold text-muted-foreground line-through">{{ formatPrice(selectedSku.price_amount, siteCurrency) }}</span>
              </div>
              <p v-if="selectedSkuPromotionFinalIsMember" class="mt-2 text-sm font-semibold text-warning">{{ t('products.memberPriceTag') }} · {{ t('products.saveAmount') }} {{ formatPrice(Number(selectedSku.price_amount) - Number(selectedSkuPromotionFinalPrice), siteCurrency) }}</p>
              <p v-else class="mt-2 text-sm font-semibold text-destructive">{{ t('products.saveAmount') }} {{ formatPrice(getSkuPromotionSaveAmount(selectedSku), siteCurrency) }}</p>
            </template>
            <!-- 3. SKU 会员价 -->
            <template v-else-if="selectedSku && hasMemberPrice">
              <div class="flex flex-wrap items-baseline gap-3">
                <span class="text-[32px] font-bold tabular-nums text-warning">{{ formatPrice(selectedSkuMemberPrice!, siteCurrency) }}</span>
                <span class="font-semibold text-muted-foreground line-through">{{ formatPrice(selectedSku.price_amount, siteCurrency) }}</span>
              </div>
              <p class="mt-2 text-sm font-semibold text-warning">{{ t('products.memberPriceTag') }} · {{ t('products.saveAmount') }} {{ formatPrice(Number(selectedSku.price_amount) - selectedSkuMemberPrice!, siteCurrency) }}</p>
            </template>
            <!-- 4. SKU 原价 -->
            <div v-else-if="selectedSku" class="flex flex-wrap items-baseline gap-3">
              <span class="text-[32px] font-bold tabular-nums text-destructive">{{ formatPrice(selectedSku.price_amount, siteCurrency) }}</span>
            </div>
            <!-- 5. 产品级促销 -->
            <template v-else-if="hasPromotionPrice(product)">
              <div class="flex flex-wrap items-baseline gap-3">
                <span class="text-[32px] font-bold tabular-nums text-destructive">{{ formatPrice(getPromotionPriceAmount(product), siteCurrency) }}</span>
                <span class="font-semibold text-muted-foreground line-through">{{ formatPrice(product.price_amount, siteCurrency) }}</span>
              </div>
              <p class="mt-2 text-sm font-semibold text-destructive">{{ t('products.saveAmount') }} {{ formatPrice(getPromotionSaveAmount(product), siteCurrency) }}</p>
            </template>
            <!-- 6. 产品级原价 -->
            <div v-else class="flex flex-wrap items-baseline gap-3">
              <span class="text-[32px] font-bold tabular-nums text-destructive">{{ formatPrice(product.price_amount, siteCurrency) }}</span>
            </div>
          </div>

          <!-- 批发规则 -->
          <div v-if="selectedSkuWholesaleRules.length" class="mb-[18px] rounded-md bg-success/15 px-[15px] py-3">
            <h4 class="mb-2 text-[13.5px] font-bold text-success">{{ t('products.wholesaleRulesTitle') }}</h4>
            <div class="flex flex-wrap gap-1.5">
              <span v-for="tier in selectedSkuWholesaleRules" :key="`${tier.sku_id || tier.sku_code || 'all'}-${tier.min_quantity}`" class="inline-flex items-center rounded-full bg-card px-2.5 py-1 text-[12.5px] font-semibold text-success">{{ formatWholesaleTier(tier) }}</span>
            </div>
          </div>

          <!-- 活动规则 -->
          <div v-if="hasPromotionRules(product)" class="mb-[18px] rounded-md bg-warning/15 px-[15px] py-3">
            <h4 class="mb-2 flex items-center gap-1.5 text-[13.5px] font-bold text-warning"><Tag class="h-[15px] w-[15px]" /> {{ t('products.promotionRulesTitle') }}</h4>
            <ul class="grid grid-cols-1 gap-1">
              <li v-for="rule in getPromotionRules(product)" :key="rule.id" class="text-[13px] text-warning">{{ formatPromotionRule(rule) }}</li>
            </ul>
          </div>

          <!-- 规格 -->
          <!-- 只有一个默认规格时不展示选择区，数量和价格已经说明一切 -->
          <div v-if="activeSkus.length > 1 || (activeSkus.length === 1 && !isDefaultSku(activeSkus[0]))" class="my-4 sm:my-5">
            <div class="mb-2.5 text-[13px] font-bold uppercase tracking-[0.04em] text-muted-foreground">{{ t('productDetail.skuTitle') }}</div>
            <!-- 窄屏两列等宽，宽屏按内容流式排列；按钮上直接给出该规格的价格 -->
            <div class="grid grid-cols-2 gap-2 sm:flex sm:flex-wrap sm:gap-2.5">
              <button
                v-for="sku in activeSkus"
                :key="sku.id"
                type="button"
                class="flex min-w-0 flex-col items-start gap-0.5 rounded-[10px] border-2 px-3 py-2 text-left text-sm font-semibold transition-colors sm:min-w-[110px] sm:px-4 sm:py-2.5"
                :class="[
                  normalizeSkuId(sku.id) === selectedSkuId ? 'border-primary bg-primary/10 text-primary' : 'border-hairline-strong bg-card text-foreground hover:border-primary/50',
                  !isSkuPurchasable(sku) ? 'cursor-not-allowed opacity-[0.42]' : '',
                ]"
                :disabled="!isSkuPurchasable(sku)"
                :aria-pressed="normalizeSkuId(sku.id) === selectedSkuId"
                @click="selectedSkuId = normalizeSkuId(sku.id)"
              >
                <span class="w-full break-words leading-snug">{{ skuDisplayText(sku) }}</span>
                <span class="flex w-full flex-wrap items-baseline gap-x-1.5 text-xs">
                  <span class="font-bold tabular-nums" :class="hasSkuPromotionPrice(sku) ? 'text-destructive' : (normalizeSkuId(sku.id) === selectedSkuId ? 'text-primary' : 'text-foreground')">{{ formatPrice(hasSkuPromotionPrice(sku) ? getSkuPromotionPriceAmount(sku) : sku.price_amount, siteCurrency) }}</span>
                  <!-- 库存不限时不再每个按钮重复「库存不限」 -->
                  <span v-if="sku.stock_status !== 'unlimited'" class="font-medium" :class="normalizeSkuId(sku.id) === selectedSkuId ? 'text-primary/80' : 'text-muted-foreground'">{{ skuStockText(sku) }}</span>
                </span>
              </button>
            </div>
            <p v-if="requiresSKUSelection" class="mt-2 text-[13px] text-warning">{{ t('productDetail.skuRequired') }}</p>
          </div>

          <!-- 数量：标题和步进器同一行，省一行高度 -->
          <div v-if="!soldOut" class="my-4 flex items-center justify-between gap-3 sm:my-5 sm:block">
            <div class="text-[13px] font-bold uppercase tracking-[0.04em] text-muted-foreground sm:mb-2.5">{{ t('productDetail.quantity') }}</div>
            <div class="inline-flex items-center overflow-hidden rounded-full border-2 border-hairline-strong">
              <button type="button" class="grid h-11 w-[42px] place-items-center bg-card text-foreground disabled:opacity-35" :aria-label="t('productDetail.quantity')" :disabled="quantity <= quantityEffectiveMin" @click="quantity = Math.max(quantityEffectiveMin, quantity - 1)"><Minus class="h-[17px] w-[17px]" /></button>
              <input inputmode="numeric" class="h-11 w-[50px] border-x bg-card text-center font-bold tabular-nums outline-none" :value="quantity" :aria-label="t('productDetail.quantity')" @change="handleQuantityInput($event)" @keydown.enter.prevent="($event.target as HTMLInputElement)?.blur()" />
              <button type="button" class="grid h-11 w-[42px] place-items-center bg-card text-foreground disabled:opacity-35" :aria-label="t('productDetail.quantity')" :disabled="quantityEffectiveLimit !== null && quantity >= quantityEffectiveLimit" @click="quantity = quantity + 1"><Plus class="h-[17px] w-[17px]" /></button>
            </div>
          </div>

          <!-- 人工交付：下单前先告诉买家要准备哪些信息 -->
          <div v-if="manualFormFields.length" class="my-4 rounded-[10px] border border-primary/15 bg-primary/5 px-3.5 py-3 sm:my-5">
            <h4 class="mb-2 flex items-center gap-1.5 text-[13px] font-bold text-primary"><ClipboardList class="h-4 w-4" /> {{ t('luna.formFieldsHint') }}</h4>
            <div class="flex flex-wrap gap-1.5">
              <span v-for="f in manualFormFields" :key="f.key" class="inline-flex items-center gap-1 rounded-full bg-card px-2.5 py-1 text-[12.5px] font-medium text-foreground">
                {{ f.label }}
                <span v-if="!f.required" class="text-muted-foreground">({{ t('luna.optionalField') }})</span>
              </span>
            </div>
          </div>

          <!-- 提示：整件售罄时由下方「已售罄」按钮说明，不再提「最小购买数量」 -->
          <div v-if="!soldOut && cannotPurchaseReason" class="my-3.5 rounded-sm bg-destructive/10 px-3.5 py-2.5 text-sm font-semibold text-destructive">{{ cannotPurchaseReason }}</div>
          <div v-if="purchaseWarning" class="my-3.5 rounded-sm bg-warning/10 px-3.5 py-2.5 text-sm font-semibold text-warning">{{ purchaseWarning }}</div>

          <!-- 操作 -->
          <div ref="purchaseActionsRef" class="mt-[18px] flex flex-wrap gap-3">
            <Button v-if="soldOut" variant="secondary" disabled class="h-12 w-full rounded-full text-[17px] font-bold">{{ t('productDetail.skuStockOut') }}</Button>
            <Button v-else-if="requiresLogin" class="luna-gradient h-12 w-full rounded-full text-[17px] font-bold" @click="goLogin">{{ t('productDetail.loginToBuy') }}</Button>
            <template v-else>
              <Button class="luna-gradient h-12 flex-1 rounded-full text-[17px] font-bold" :disabled="!canPurchase" @click="buyNow"><Zap /> {{ t('productDetail.buyNow') }}</Button>
              <Button variant="outline" class="h-12 rounded-full text-[17px] font-bold" :disabled="!canPurchase" @click="addToCart"><ShoppingCart /> {{ t('productDetail.addToCart') }}</Button>
            </template>
          </div>

          <!-- 交付说明：自动发卡与人工处理分开说，避免人工商品也写「兑换码即时送达」 -->
          <div class="mt-4 flex items-center gap-3 rounded-md px-4 py-3 sm:px-[18px] sm:py-3.5" :class="product.fulfillment_type === 'auto' ? 'bg-success/15 text-success' : 'bg-primary/10 text-primary'">
            <component :is="product.fulfillment_type === 'auto' ? TicketCheck : Headphones" class="h-[22px] w-[22px] flex-none" />
            <span class="text-[13px] font-semibold sm:text-sm">{{ product.fulfillment_type === 'auto' ? t('productDetail.deliveryReassurance') : t('luna.manualDeliveryReassurance') }}</span>
          </div>
        </div>

        <!-- 无图商品的详情：宽屏在左栏标题下方，窄屏接在购买区下方 -->
        <aside v-if="!images.length && hasDetails" class="min-w-0 border-t pt-5 lg:col-start-1 lg:row-start-2">
          <h2 class="mb-3 flex items-center gap-2 text-base font-semibold text-foreground sm:text-lg">
            <FileText class="h-5 w-5 flex-none text-primary" />
            {{ t('productDetail.details') }}
          </h2>
          <div class="prose prose-sm max-w-none dark:prose-invert prose-a:text-primary prose-img:rounded-md sm:prose-base">
            <p v-if="getLocalizedText(product.description)">{{ getLocalizedText(product.description) }}</p>
            <div v-if="product.content" v-html="processHtmlForDisplay(getLocalizedText(product.content))"></div>
          </div>
        </aside>
      </section>

      <!-- 描述 / 详情（有图商品放在下方整宽面板） -->
      <section v-if="images.length && hasDetails" class="luna-panel mt-4 p-4 sm:mt-5 sm:p-6">
        <h2 class="mb-4 flex items-center gap-2 border-b pb-3 text-lg font-semibold text-foreground">
          <FileText class="h-5 w-5 flex-none text-primary" />
          {{ t('productDetail.details') }}
        </h2>
        <div class="prose max-w-none dark:prose-invert prose-a:text-primary prose-img:rounded-md">
          <p v-if="getLocalizedText(product.description)">{{ getLocalizedText(product.description) }}</p>
          <div v-if="product.content" v-html="processHtmlForDisplay(getLocalizedText(product.content))"></div>
        </div>
      </section>

      <!-- 相关文章 -->
      <section v-if="relatedPosts.length" class="pt-8">
        <div class="mb-[22px]"><h2 class="text-2xl font-extrabold">{{ t('productDetail.relatedPosts') }}</h2></div>
        <div class="grid gap-4 grid-cols-[repeat(auto-fill,minmax(228px,1fr))]">
          <RouterLink v-for="rp in relatedPosts" :key="rp.id" class="block h-full rounded-lg border bg-card p-[22px] transition hover:-translate-y-[3px] hover:border-hairline-strong hover:shadow-[var(--luna-shadow)]" :to="`/blog/${rp.slug}`">
            <span class="text-[12.5px] font-semibold text-muted-foreground">{{ formatRelatedPostDate(rp.published_at) }}</span>
            <h3 class="mt-2 line-clamp-2 text-[17px] font-bold">{{ getLocalizedText(rp.title) }}</h3>
            <p v-if="rp.summary" class="mt-2 line-clamp-2 text-sm text-muted-foreground">{{ getLocalizedText(rp.summary) }}</p>
            <span class="mt-3.5 inline-flex items-center gap-1.5 text-sm font-bold text-primary">{{ t('blog.readMore') }} <ChevronRight class="h-4 w-4" /></span>
          </RouterLink>
        </div>
      </section>

      <div class="py-6 text-center">
        <Button as-child variant="ghost" size="sm" class="rounded-full"><RouterLink :to="backToListTo"><ArrowLeft /> {{ t('productDetail.backToProducts') }}</RouterLink></Button>
      </div>

      <!-- 移动端固定购买条 -->
      <LunaProductMobileBar
        :visible="showMobileBar && !!product && !loading"
        :requires-login="requiresLogin"
        :can-purchase="canPurchase"
        :sold-out="soldOut"
        :summary="mobileBarSummary"
        :show-member-price="mobileBarShowMemberPrice"
        :member-price-display="mobileBarMemberPriceDisplay"
        :show-sku-promotion-price="mobileBarShowSkuPromotionPrice"
        :sku-promotion-price-display="mobileBarSkuPromotionPriceDisplay"
        :show-sku-price="mobileBarShowSkuPrice"
        :sku-price-display="mobileBarSkuPriceDisplay"
        :show-product-promotion-price="mobileBarShowProductPromotionPrice"
        :product-promotion-price-display="mobileBarProductPromotionPriceDisplay"
        :product-price-display="mobileBarProductPriceDisplay"
        @add-to-cart="addToCart"
        @buy-now="buyNow"
        @go-login="goLogin"
      />
    </template>

    <!-- 错误 -->
    <div v-else class="my-10 flex flex-col items-center gap-3 rounded-lg border border-dashed py-14 text-center text-muted-foreground">
      <AlertCircle class="h-12 w-12 opacity-60" />
      <p>{{ t('productDetail.notFound') }}</p>
      <div class="mt-2 flex flex-wrap justify-center gap-2.5">
        <Button size="sm" class="rounded-full" @click="loadProduct"><RotateCw /> {{ t('errorBoundary.retry') }}</Button>
        <Button as-child variant="outline" size="sm" class="rounded-full"><RouterLink :to="backToListTo">{{ t('productDetail.backToProducts') }}</RouterLink></Button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  AlertCircle, ArrowLeft, ChevronRight, ClipboardList, FileText, Headphones, Lock, Minus, Package, Pencil, Plus,
  RotateCw, ShoppingCart, Tag, TicketCheck, UserPlus, Zap,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { processHtmlForDisplay } from '../../utils/content'
import { getImageUrl } from '../../utils/image'
import { useProductDetail } from '../../composables/useProductDetail'
import LunaProductMobileBar from './components/LunaProductMobileBar.vue'

const { t } = useI18n()

// 移动端固定购买条：IntersectionObserver 监听桌面购买区是否出屏
const purchaseActionsRef = ref<HTMLElement | null>(null)
const showMobileBar = ref(false)
let observer: IntersectionObserver | null = null

const setupMobileBarObserver = () => {
  if (observer) observer.disconnect()
  if (!purchaseActionsRef.value) return
  observer = new IntersectionObserver(
    (entries) => {
      const entry = entries[0]
      // 按可见比例判断：isIntersecting 只要露出 1px 就为 true，按钮快被遮完时购买条还不出现
      if (entry) showMobileBar.value = entry.intersectionRatio < 0.5
    },
    // 顶部扣掉 64px 的固定顶栏：按钮藏到顶栏后面也算出屏
    { threshold: [0, 0.5, 1], rootMargin: '-64px 0px 0px 0px' },
  )
  observer.observe(purchaseActionsRef.value)
}

// onLoaded 回调触发时 loading 还没置回 false，购买区尚未渲染、ref 为空，监听挂不上；
// 改为等购买区真正出现（或切换商品重新渲染）时再挂
watch(purchaseActionsRef, () => setupMobileBarObserver())

// 购买条出现时给页脚留出底部空间（页脚在 LunaLayout 里，见 styles/luna.css）
watch(showMobileBar, (open) => document.body.classList.toggle('luna-mobilebar-open', open))

const {
  getLocalizedText, siteCurrency, formatPrice,
  getFulfillmentTypeLabel, getPurchaseTypeLabel, getStockStatusLabel, getStockBadgeVariant,
  hasPromotionPrice, getPromotionPriceAmount, getPromotionSaveAmount,
  hasSkuPromotionPrice, getSkuPromotionPriceAmount, getSkuPromotionSaveAmount,
  hasPromotionRules, getPromotionRules,
  formatPromotionRule, formatWholesaleTier, formatRelatedPostDate, normalizeSkuId,
  loading, product, relatedPosts, currentImage, selectedSkuId, quantity, purchaseWarning,
  activeSkus, selectedSku,
  selectedSkuMemberPrice, hasMemberPrice,
  hasSelectedSkuWholesalePrice, selectedSkuWholesaleFinalIsMember, selectedSkuWholesaleFinalPrice,
  selectedSkuWholesaleRules,
  selectedSkuPromotionPrice, selectedSkuPromotionFinalIsMember, selectedSkuPromotionFinalPrice,
  showSelectedSkuMemberBadge,
  isSkuPurchasable, skuDisplayText, skuStockText,
  quantityEffectiveLimit, quantityEffectiveMin, handleQuantityInput,
  requiresLogin, requiresSKUSelection, canPurchase, cannotPurchaseReason,
  categoryName, images,
  backToListTo,
  addToCart, buyNow, goLogin, loadProduct,
  mobileBarShowMemberPrice, mobileBarMemberPriceDisplay,
  mobileBarShowSkuPromotionPrice, mobileBarSkuPromotionPriceDisplay,
  mobileBarShowSkuPrice, mobileBarSkuPriceDisplay,
  mobileBarShowProductPromotionPrice, mobileBarProductPromotionPriceDisplay, mobileBarProductPriceDisplay,
} = useProductDetail({ onLoaded: () => setupMobileBarObserver() })

const stockPillTone = computed(() => {
  const variant = getStockBadgeVariant(product.value?.stock_status)
  if (variant === 'destructive') return 'bg-secondary text-muted-foreground'
  if (variant === 'warning') return 'bg-warning/15 text-warning'
  return 'bg-success/15 text-success'
})

const hasDetails = computed(() => Boolean(getLocalizedText(product.value?.description) || product.value?.content))

/** 无商品图时标题旁的小标识用分类图标（与首页卡片的取图规则一致） */
const categoryIconErrored = ref(false)
watch(() => product.value?.id, () => { categoryIconErrored.value = false })
const categoryIcon = computed(() => {
  const icon = product.value?.category?.icon
  return icon && !categoryIconErrored.value ? getImageUrl(icon) : ''
})

// 宽屏网格定位（窄屏都是单列，按 图片 → 标题 → 购买 → 详情 的文档顺序排）
const sectionGridClass = computed(() => images.value.length
  ? 'lg:grid-cols-[minmax(0,5fr)_minmax(0,7fr)] lg:grid-rows-[auto_1fr]'
  : 'lg:grid-cols-[minmax(0,7fr)_minmax(0,5fr)] lg:grid-rows-[auto_1fr]')
const headerPlaceClass = computed(() => images.value.length ? 'lg:col-start-2 lg:row-start-1' : 'lg:col-start-1 lg:row-start-1')
const purchasePlaceClass = computed(() => images.value.length
  ? 'lg:col-start-2 lg:row-start-2 lg:rounded-xl lg:border lg:bg-secondary/40 lg:p-5'
  : 'lg:col-start-2 lg:row-span-2 lg:row-start-1 lg:sticky lg:top-24 lg:self-start lg:rounded-xl lg:border lg:bg-secondary/40 lg:p-5')

/** 面包屑与「返回」都回到商品所属分类 */
const categoryPath = computed(() => {
  const slug = product.value?.category?.slug
  return slug ? `/categories/${encodeURIComponent(slug)}` : '/products'
})

const soldOut = computed(() => Boolean(product.value?.is_sold_out))

/** 后台没填规格值的「默认规格」，单独一个时不必让买家选 */
const isDefaultSku = (sku: any) => {
  const spec = sku?.spec_values
  return !spec || (typeof spec === 'object' && Object.keys(spec).length === 0)
}

/** 人工交付商品下单时要填的字段（只取名称和是否必填，用来提前告知买家） */
const manualFormFields = computed(() => {
  if (product.value?.fulfillment_type !== 'manual') return []
  const fields = product.value?.manual_form_schema?.fields
  if (!Array.isArray(fields)) return []
  return fields
    .filter((f: any) => f && f.key)
    .map((f: any) => ({ key: String(f.key), label: getLocalizedText(f.label) || String(f.key), required: f.required !== false }))
})

/** 购买条上的已选规格与数量，如「980+110 结晶 × 2」 */
const mobileBarSummary = computed(() => {
  const sku = selectedSku.value
  const label = sku && activeSkus.value.length > 1 ? skuDisplayText(sku) : ''
  return label ? `${label} × ${quantity.value}` : `× ${quantity.value}`
})

onUnmounted(() => {
  document.body.classList.remove('luna-mobilebar-open')
  if (observer) {
    observer.disconnect()
    observer = null
  }
})
</script>
