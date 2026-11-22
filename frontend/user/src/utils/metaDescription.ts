const META_DESCRIPTION_MAX_LENGTH = 160
// 描述可能来自上游同步，内容不可信。先截断再跑正则，避免超长文本让正则卡住页面
const META_DESCRIPTION_SCAN_LIMIT = 20000

const HTML_ENTITIES: Record<string, string> = {
  '&nbsp;': ' ',
  '&amp;': '&',
  '&lt;': '<',
  '&gt;': '>',
  '&quot;': '"',
  '&#39;': "'",
}

// 以下两个函数用 indexOf 线性扫描，结果分别等价于
//   .replace(/<(script|style)\b[^>]*>[\s\S]*?<\/\1\s*>/gi, ' ')
//   .replace(/<[^>]*>/g, ' ')
// 正则版在「大量 < 却没有 >」「大量开标签却没有闭合标签」时会从每个位置扫到结尾再回溯，
// 是 O(n²)：两万字符即需数亿步，会把页面卡住数百毫秒。

/** 删除 <script>…</script> / <style>…</style> 整段（标签名不区分大小写） */
const stripScriptAndStyleBlocks = (input: string): string => {
  const opener = /<(script|style)\b/gi
  const tagsWithoutCloser = new Set<string>()
  let output = ''
  let copiedUntil = 0
  let match: RegExpExecArray | null
  while ((match = opener.exec(input))) {
    const tag = (match[1] ?? '').toLowerCase()
    if (tagsWithoutCloser.has(tag)) continue
    const openerEnd = input.indexOf('>', opener.lastIndex)
    // 这里没有 >，后面任何开标签也不会有，不可能再匹配
    if (openerEnd < 0) break
    const closer = new RegExp(`</${tag}\\s*>`, 'gi')
    closer.lastIndex = openerEnd + 1
    const closeMatch = closer.exec(input)
    if (!closeMatch) {
      // 之后同名开标签更靠后，同样找不到闭合标签，记下来不再重复扫描
      tagsWithoutCloser.add(tag)
      opener.lastIndex = match.index + 1
      continue
    }
    output += `${input.slice(copiedUntil, match.index)} `
    copiedUntil = closer.lastIndex
    opener.lastIndex = copiedUntil
  }
  return output + input.slice(copiedUntil)
}

/** 把每个 <…> 替换成一个空格；没有对应 > 的 < 原样保留 */
const stripTags = (input: string): string => {
  let output = ''
  let position = 0
  for (;;) {
    const tagStart = input.indexOf('<', position)
    if (tagStart < 0) break
    const tagEnd = input.indexOf('>', tagStart + 1)
    // 这里没有 >，后面的 < 也都不会有，剩余部分原样保留
    if (tagEnd < 0) break
    output += `${input.slice(position, tagStart)} `
    position = tagEnd + 1
  }
  return output + input.slice(position)
}

/** 把商品描述转成适合放进 meta description 的纯文本：去掉标签、合并空白、限制长度 */
export const toMetaDescription = (value: unknown, maxLength = META_DESCRIPTION_MAX_LENGTH): string => {
  if (typeof value !== 'string') return ''
  const text = stripTags(stripScriptAndStyleBlocks(value.slice(0, META_DESCRIPTION_SCAN_LIMIT)))
    .replace(/&(?:nbsp|amp|lt|gt|quot|#39);/g, (entity) => HTML_ENTITIES[entity] ?? entity)
    .replace(/\s+/g, ' ')
    .trim()
  const characters = Array.from(text)
  if (characters.length <= maxLength) return text
  return `${characters.slice(0, maxLength - 1).join('').trimEnd()}…`
}
