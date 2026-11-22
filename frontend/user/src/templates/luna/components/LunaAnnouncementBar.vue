<script setup lang="ts">
import { computed } from 'vue'
import { Megaphone } from 'lucide-vue-next'
import { useLocalized } from '../../../composables/useProduct'
import { sanitizeAnnouncementHtml } from '../../../utils/announcement'
import { useAppStore } from '../../../stores/app'
import type { HomeAnnouncement } from '../../../composables/useAnnouncement'

const props = defineProps<{
  announcement: HomeAnnouncement
}>()

const appStore = useAppStore()
const { getLocalizedText } = useLocalized()

// 标题只取当前语言，不回退到其它语言：清空中文标题后不该冒出英文标题配中文正文
const title = computed(() => String(props.announcement.title?.[appStore.locale] ?? '').trim())
const content = computed(() => sanitizeAnnouncementHtml(getLocalizedText(props.announcement.content)))
</script>

<template>
  <!-- 普通公告：常驻在分类面板上方的一栏，不弹窗、不可关闭；标题为空时不显示标题和图标，只显示正文 -->
  <section
    v-if="title || content"
    class="luna-panel mb-4 flex gap-3 px-4 py-4 sm:mb-5 sm:gap-4 sm:px-6"
    :aria-label="title || undefined"
  >
    <span v-if="title" class="flex h-9 w-9 flex-none items-center justify-center rounded-[10px] bg-primary/10 text-primary">
      <Megaphone class="h-5 w-5" aria-hidden="true" />
    </span>
    <div class="min-w-0 flex-1">
      <h2 v-if="title" class="text-base font-semibold leading-9 text-foreground">{{ title }}</h2>
      <div
        v-if="content"
        class="prose prose-sm max-w-none text-muted-foreground dark:prose-invert prose-p:my-0.5 prose-a:text-primary prose-img:rounded-md"
        v-html="content"
      ></div>
    </div>
  </section>
</template>
