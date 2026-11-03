import { ref, watch } from 'vue'
import { syncThemeColorMeta } from './bootScreen'

// 主题优先级：用户手动选择 > 系统设置 > 默认黑夜。
// 旧版本会把自动判断的结果写进 dujiao_theme，分不清是不是用户选的，所以换一个只在手动切换时写入的键。
// 这段规则在 index.html 的内联脚本里也有一份，修改时要同步。
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

/** 读取不到系统设置（浏览器不支持、或用户没有偏好）时返回 null */
const getSystemTheme = (): Theme | null => {
    try {
        if (window.matchMedia('(prefers-color-scheme: dark)').matches) return 'dark'
        if (window.matchMedia('(prefers-color-scheme: light)').matches) return 'light'
    } catch {
        // 不支持 matchMedia 的环境按读取不到处理
    }
    return null
}

let hasManualChoice = getSavedTheme() !== null

const theme = ref<Theme>(getSavedTheme() || getSystemTheme() || DEFAULT_THEME)

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

// 没手动选过时，页面开着期间系统主题变化也跟随，和重新打开页面的结果保持一致
try {
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
        if (!hasManualChoice) theme.value = getSystemTheme() || DEFAULT_THEME
    })
} catch {
    // 不支持时只是不会实时跟随，打开页面时的判断不受影响
}

export const useTheme = () => {
    const toggleTheme = () => {
        hasManualChoice = true
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
