/**
 * 浏览器禁用存储时（隐私设置、沙箱 iframe、部分 WebView），读取 localStorage 会直接抛错或得到 null，
 * 而很多模块在加载时就会读取，结果是整页白屏。这里换成内存实现，功能退化为「刷新后不记住」，页面照常工作。
 *
 * main.ts 必须最先 import 本文件，才能赶在其它模块读取存储之前生效。
 */

type StorageName = 'localStorage' | 'sessionStorage'

export const createMemoryStorage = (): Storage => {
  const items = new Map<string, string>()
  return {
    get length() {
      return items.size
    },
    clear: () => items.clear(),
    getItem: (key) => items.get(String(key)) ?? null,
    key: (index) => [...items.keys()][index] ?? null,
    removeItem: (key) => {
      items.delete(String(key))
    },
    setItem: (key, value) => {
      items.set(String(key), String(value))
    },
  }
}

const isStorageUsable = (target: Record<string, any>, name: StorageName): boolean => {
  try {
    const storage = target[name]
    if (!storage) return false
    storage.getItem('__dujiao_storage_probe__')
    return true
  } catch {
    return false
  }
}

/** 存储不可用时替换成内存实现，返回是否发生了替换。存储可用时不做任何改动，已保存的数据不受影响。 */
export const ensureUsableStorage = (target: Record<string, any>, name: StorageName): boolean => {
  if (isStorageUsable(target, name)) return false
  try {
    Object.defineProperty(target, name, { value: createMemoryStorage(), configurable: true, writable: true })
    return true
  } catch {
    return false
  }
}

if (typeof window !== 'undefined') {
  ensureUsableStorage(window as unknown as Record<string, any>, 'localStorage')
  ensureUsableStorage(window as unknown as Record<string, any>, 'sessionStorage')
}
