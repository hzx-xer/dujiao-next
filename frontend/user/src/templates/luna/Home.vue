<template>
  <div class="mx-auto w-full max-w-[1100px] px-3 pt-5 sm:px-6 sm:pt-8">
    <!-- 页面只有「选择分类 / 选择商品」两个二级标题，补一个仅供读屏与搜索引擎使用的一级标题 -->
    <h1 class="sr-only">{{ pageHeading }}</h1>
    <section class="luna-panel px-4 py-5 sm:px-6 sm:py-6">
      <!-- 选择分类：没有「全部商品」，默认选中第一个分类 -->
      <template v-if="categoryGroups.length">
        <h2 class="flex items-center gap-2 text-lg font-semibold text-foreground">
          <LayoutGrid class="h-5 w-5 flex-none text-primary" />
          {{ t('luna.chooseCategory') }}
        </h2>
        <!-- 窄屏三列网格，避免每个按钮独占半行；宽屏按内容宽度流式排列 -->
        <div class="mt-4 grid grid-cols-3 gap-2 sm:flex sm:flex-wrap sm:gap-2.5">
          <button
            v-for="group in categoryGroups"
            :key="group.id"
            type="button"
            class="inline-flex h-11 min-w-0 items-center justify-center gap-1.5 rounded-[10px] px-2 text-[13px] transition sm:h-12 sm:min-w-[108px] sm:max-w-full sm:gap-2 sm:px-5 sm:text-sm"
            :class="activeGroup?.id === group.id ? 'luna-gradient font-semibold' : 'bg-secondary text-foreground hover:bg-primary/10 hover:text-primary'"
            :aria-pressed="activeGroup?.id === group.id"
            @click="selectCategory(group.id)"
          >
            <img v-if="group.icon" :src="getImageUrl(group.icon)" alt="" loading="lazy" class="h-5 w-5 flex-none rounded object-cover" />
            <span class="truncate">{{ catName(group) }}</span>
            <!-- 有子分类的一级分类带个小箭头，提示点开还有下一级 -->
            <ChevronDown v-if="group.children.length" class="h-3.5 w-3.5 flex-none opacity-70 transition-transform" :class="{ 'rotate-180': activeGroup?.id === group.id }" />
          </button>
        </div>

        <!-- 子分类面板：「全部」看父分类及所有子分类的商品，点子分类只看该子分类 -->
        <Transition
          enter-active-class="transition duration-200 ease-out"
          enter-from-class="-translate-y-1 opacity-0"
          leave-active-class="transition duration-100 ease-in"
          leave-to-class="opacity-0"
          mode="out-in"
        >
          <div
            v-if="activeGroup?.children.length"
            :key="activeGroup.id"
            class="mt-3 flex items-center gap-2 rounded-[10px] border border-primary/15 bg-primary/5 p-1.5 sm:p-2"
          >
            <span class="hidden flex-none items-center gap-1 pl-1.5 pr-1 text-[13px] font-semibold text-primary sm:inline-flex">
              <CornerDownRight class="h-3.5 w-3.5" />
              {{ catName(activeGroup) }}
            </span>
            <!-- 窄屏单行横滑，宽屏折行 -->
            <div ref="subcategoryRow" class="luna-no-scrollbar -my-0.5 flex min-w-0 flex-1 gap-1.5 overflow-x-auto p-0.5 sm:flex-wrap sm:overflow-visible" role="group" :aria-label="t('luna.subcategories')">
              <button
                v-for="item in subcategoryItems"
                :key="item.id"
                type="button"
                class="inline-flex h-8 max-w-[70vw] flex-none items-center gap-1.5 rounded-full px-3.5 text-[13px] transition-colors sm:max-w-full"
                :class="selectedCategory === item.id ? 'bg-card font-semibold text-primary shadow-[var(--luna-shadow-sm)] ring-1 ring-primary/40' : 'text-muted-foreground hover:bg-card/70 hover:text-primary'"
                :aria-pressed="selectedCategory === item.id"
                :data-active="selectedCategory === item.id || undefined"
                @click="selectCategory(item.id)"
              >
                <img v-if="item.icon" :src="getImageUrl(item.icon)" alt="" loading="lazy" class="h-4 w-4 flex-none rounded object-cover" />
                <span class="truncate">{{ item.label }}</span>
              </button>
            </div>
          </div>
        </Transition>
      </template>

      <!-- 选择商品 -->
      <div :class="{ 'mt-5 border-t pt-5': categoryGroups.length }">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <h2 class="flex min-w-0 items-center gap-2 text-lg font-semibold text-foreground">
            <ShoppingBag class="h-5 w-5 flex-none text-primary" />
            {{ t('luna.chooseProduct') }}
            <!-- 当前浏览范围，例如「游戏充值 / Steam 钱包」 -->
            <span v-if="selectedCategoryPath" class="truncate text-sm font-normal text-muted-foreground">· {{ selectedCategoryPath }}</span>
          </h2>
          <div class="relative w-full sm:w-64">
            <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <input
              v-model="searchQuery"
              class="h-9 w-full rounded-full border border-border bg-card pl-9 pr-9 text-sm text-foreground transition-colors placeholder:text-muted-foreground hover:border-hairline-strong focus:border-primary focus:outline-none"
              :placeholder="t('products.searchBoxPlaceholder')"
              :aria-label="t('products.searchLabel')"
            />
            <button v-if="searchQuery" type="button" class="absolute right-2 top-1/2 grid h-6 w-6 -translate-y-1/2 place-items-center rounded-full text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground" :aria-label="t('blog.searchClear')" @click="clearSearch"><X class="h-3.5 w-3.5" /></button>
          </div>
        </div>

        <!-- 骨架 -->
        <div v-if="loading" class="mt-4 grid" :class="gridClass">
          <div v-for="i in 6" :key="i" class="animate-pulse rounded-[10px] bg-secondary" :class="productLayout === 'list' ? 'h-[72px]' : 'h-[150px] sm:h-[100px]'"></div>
        </div>

        <div v-else-if="products.length" class="mt-4 grid" :class="gridClass">
          <LunaProductCard v-for="product in products" :key="product.id" :product="product" :layout="productLayout" :category-label="productSubcategoryLabel(product)" />
        </div>

        <!-- 空态 -->
        <div v-else class="mt-4 flex flex-col items-center gap-3 rounded-[10px] border border-dashed py-14 text-center text-sm text-muted-foreground">
          <component :is="searchQuery ? SearchX : PackageOpen" class="h-11 w-11 opacity-60" />
          <p>{{ searchQuery ? t('products.emptyFiltered') : t('products.empty') }}</p>
          <Button v-if="searchQuery" variant="outline" size="sm" class="rounded-full" @click="clearSearch">{{ t('products.clearFilters') }}</Button>
        </div>

        <PaginationNav :current-page="currentPage" :total-pages="totalPages" :loading="loading" :scroll-top="false" compact @change-page="changePage" />
      </div>
    </section>

    <AnnouncementModal
      v-if="activeAnnouncement"
      :announcement="activeAnnouncement"
      :visible="announcementVisible"
      @update:visible="announcementVisible = $event"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ChevronDown, CornerDownRight, LayoutGrid, PackageOpen, Search, SearchX, ShoppingBag, X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { buildCategoryPath, normalizeCategoryParentId, type PublicCategory } from '../../utils/category'
import { getImageUrl } from '../../utils/image'
import { useLocalized } from '../../composables/useProduct'
import { useProductList } from '../../composables/useProductList'
import { usePageSeo } from '../../composables/usePageSeo'
import { useAnnouncement, type HomeAnnouncement } from '../../composables/useAnnouncement'
import { useAppStore } from '../../stores/app'
import LunaProductCard from './components/LunaProductCard.vue'
import PaginationNav from '../../components/PaginationNav.vue'
import AnnouncementModal from '../../components/AnnouncementModal.vue'

const route = useRoute()
const { t } = useI18n()
const { getLocalizedText } = useLocalized()
const appStore = useAppStore()

// luna 的首页、/products、/categories/:slug 是同一个页面：分类 + 该分类下的商品
const {
  loading,
  products,
  selectedCategory,
  searchQuery,
  currentPage,
  totalPages,
  categoryGroups,
  categoryMap,
  selectCategory,
  changePage,
  clearSearch,
  initialize,
  cleanup,
} = useProductList({ pageSize: 24, homeRouteName: 'home', defaultToFirstCategory: true })

const catName = (cat: PublicCategory) => getLocalizedText(cat.name) || cat.slug || ''

/** 当前高亮的顶级分类：选中的就是它，或选中的是它的子分类 */
const activeGroup = computed(() => {
  const selected = selectedCategory.value ? categoryMap.value.get(selectedCategory.value) : undefined
  if (!selected) return undefined
  const parentId = normalizeCategoryParentId(selected.parent_id)
  return categoryGroups.value.find((group) => group.id === (parentId > 0 ? parentId : selected.id))
})

/** 子分类面板的按钮：第一个「全部」就是父分类本身 */
const subcategoryItems = computed(() => {
  const group = activeGroup.value
  if (!group?.children.length) return []
  return [
    { id: group.id, label: t('luna.subcategoryAll'), icon: '' },
    ...group.children.map((child) => ({ id: child.id, label: catName(child), icon: child.icon || '' })),
  ]
})

/** 正在看父分类（全部）时，商品卡片标出各自所属的子分类 */
const productSubcategoryLabel = (product: { category_id?: number }) => {
  const group = activeGroup.value
  if (!group?.children.length || selectedCategory.value !== group.id) return ''
  const cat = product.category_id ? categoryMap.value.get(Number(product.category_id)) : undefined
  return cat && cat.id !== group.id ? catName(cat) : ''
}

// 窄屏子分类是横滑的一行，选中项可能在屏幕外（如直接打开 /categories/xxx），滚到可见处
const subcategoryRow = ref<HTMLElement | null>(null)
watch([selectedCategory, () => subcategoryItems.value.length], async () => {
  await nextTick()
  const row = subcategoryRow.value
  const active = row?.querySelector<HTMLElement>('[data-active]')
  if (!row || !active || row.scrollWidth <= row.clientWidth) return
  const offset = active.getBoundingClientRect().left - row.getBoundingClientRect().left
  row.scrollTo({ left: row.scrollLeft + offset - (row.clientWidth - active.offsetWidth) / 2, behavior: 'smooth' })
})

// 后台「商品布局」（grid-cols-N 的轨道都是 minmax(0,1fr)，长商品名不会把列撑宽）
const productLayout = computed<'card' | 'list'>(() => appStore.config?.template_mode === 'list' ? 'list' : 'card')
// 卡片模式窄屏也两列（卡片自身在窄屏切成竖版）；列表模式各断点单列紧凑行
// 窄屏不拉伸行高：有图卡和无图卡同排时，无图卡保持自身高度，空白留在卡片外而不是卡片中间
const gridClass = computed(() => productLayout.value === 'list'
  ? 'grid-cols-1 gap-2.5'
  : 'grid-cols-2 items-start gap-2.5 sm:items-stretch sm:gap-3 lg:grid-cols-3')

// ==================== 公告弹窗（版本化可关闭） ====================
const { shouldShow } = useAnnouncement()
const activeAnnouncement = ref<HomeAnnouncement | null>(null)
const announcementVisible = ref(false)
const showAnnouncementIfNeeded = () => {
  const announcement = appStore.config?.announcement as HomeAnnouncement | undefined
  if (announcement && shouldShow(announcement)) {
    activeAnnouncement.value = announcement
    announcementVisible.value = true
  }
}

// ==================== SEO ====================
const selectedCategoryName = computed(() => {
  const cat = selectedCategory.value ? categoryMap.value.get(selectedCategory.value) : undefined
  return cat ? catName(cat) : ''
})

const selectedCategoryPath = computed(() => {
  const cat = selectedCategory.value ? categoryMap.value.get(selectedCategory.value) : undefined
  return buildCategoryPath(cat, categoryMap.value, (name) => getLocalizedText(name))
})

/** 一级标题：分类页用分类名，首页用站点名 */
const pageHeading = computed(() => {
  if (route.name === 'category-products' && selectedCategoryName.value) return selectedCategoryName.value
  return String(appStore.config?.brand?.site_name || '').trim() || t('nav.products')
})

usePageSeo({
  canonicalPath: () => route.path,
  title: () => {
    if (route.name === 'category-products') return selectedCategoryName.value || t('nav.products')
    if (route.name === 'products') return t('nav.products')
    return undefined
  },
})

onMounted(async () => {
  await appStore.loadConfig()
  await initialize()
  showAnnouncementIfNeeded()
})

onUnmounted(() => cleanup())
</script>
