/**
 * 列表页滚动恢复信号。
 *
 * useProductList 在数据加载完成后 resolve 此 Promise，
 * router 的 scrollBehavior 在后退到列表页时 await 它，
 * 确保页面高度撑开后再恢复 savedPosition，避免截断。
 *
 * 每次 loadProducts 开始前通过 resetListScrollSignal() 重置；scrollBehavior 在等待前也会先重置，
 * 否则后退时等到的是上一次加载留下的已完成信号，页面还没撑开就恢复位置。
 */

let _resolve: (() => void) | null = null
let _promise: Promise<void> = Promise.resolve()

export function resetListScrollSignal() {
  // 已有未完成的信号时沿用它，避免先挂上的等待方（scrollBehavior）永远等不到 resolve
  if (_resolve) return
  _promise = new Promise<void>((resolve) => {
    _resolve = resolve
  })
}

export function resolveListScrollSignal() {
  _resolve?.()
  _resolve = null
}

export function waitForListScrollSignal(timeoutMs = 2000): Promise<void> {
  return Promise.race([
    _promise,
    new Promise<void>((resolve) => setTimeout(resolve, timeoutMs)),
  ])
}
