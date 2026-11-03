/** 让手机浏览器地址栏的颜色和页面底色一致。页面配色生效后才读得到准确的值。 */
export const syncThemeColorMeta = (): void => {
  const background = getComputedStyle(document.body).backgroundColor
  const meta = document.querySelector('meta[name="theme-color"]')
  if (meta && background && background !== 'rgba(0, 0, 0, 0)' && background !== 'transparent') {
    meta.setAttribute('content', background)
  }
}

/**
 * index.html 的内联脚本在 JS 加载前就按主题铺好背景（html[data-boot]），避免先白屏再变色。
 * 背景要一直保留到模板外壳挂载、页面配色生效，否则中间会闪一下默认底色。
 */
export const finishBootScreen = (): void => {
  document.documentElement.removeAttribute('data-boot')
  syncThemeColorMeta()
}
