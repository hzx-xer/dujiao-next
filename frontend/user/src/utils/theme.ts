import { ref, watch } from 'vue'
import { syncThemeColorMeta } from './bootScreen'

// 旧版本会把自动判断的结果写进 dujiao_theme，分不清是不是用户选的，所以换一个只在手动切换时写入的键。
// 这个键名和默认值在 index.html 的内联脚本里也有一份，修改时要同步。
const THEME_CHOICE_KEY = 'dujiao_theme_choice'
const DEFAULT_THEME: Theme = 'dark'

type Theme = 'light' | 'dark'

const getSavedTheme = (): Theme | null => {
    try {
        const saved = localStorage.getItem(THEME_CHOICE_KEY)
        return saved === 'dark' || saved === 'light' ? saved : null
    } catch {
        return null
    }
}

const theme = ref<Theme>(getSavedTheme() || DEFAULT_THEME)

const applyTheme = (newTheme: Theme) => {
    const root = document.documentElement
    if (newTheme === 'dark') {
        root.classList.add('dark')
    } else {
        root.classList.remove('dark')
    }
    // 让滚动条、表单控件等浏览器原生部分也跟随主题
    root.style.colorScheme = newTheme
    // 启动背景还在时 body 是透明的，地址栏颜色由内联脚本负责，撤掉背景时再同步
    if (!root.hasAttribute('data-boot')) syncThemeColorMeta()
}

watch(theme, (newVal) => {
    applyTheme(newVal)
}, { immediate: true, flush: 'post' })

export const useTheme = () => {
    const toggleTheme = () => {
        theme.value = theme.value === 'dark' ? 'light' : 'dark'
        try {
            localStorage.setItem(THEME_CHOICE_KEY, theme.value)
        } catch {
            // 存储不可用时本次访问仍然生效，只是刷新后不记住
        }
    }

    return {
        theme,
        toggleTheme
    }
}
