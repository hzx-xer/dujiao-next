const META_DESCRIPTION_MAX_LENGTH = 160

const HTML_ENTITIES: Record<string, string> = {
  '&nbsp;': ' ',
  '&amp;': '&',
  '&lt;': '<',
  '&gt;': '>',
  '&quot;': '"',
  '&#39;': "'",
}

/** 把商品描述转成适合放进 meta description 的纯文本：去掉标签、合并空白、限制长度 */
export const toMetaDescription = (value: unknown, maxLength = META_DESCRIPTION_MAX_LENGTH): string => {
  if (typeof value !== 'string') return ''
  const text = value
    .replace(/<(script|style)\b[^>]*>[\s\S]*?<\/\1\s*>/gi, ' ')
    .replace(/<[^>]*>/g, ' ')
    .replace(/&(?:nbsp|amp|lt|gt|quot|#39);/g, (entity) => HTML_ENTITIES[entity] ?? entity)
    .replace(/\s+/g, ' ')
    .trim()
  const characters = Array.from(text)
  if (characters.length <= maxLength) return text
  return `${characters.slice(0, maxLength - 1).join('').trimEnd()}…`
}
