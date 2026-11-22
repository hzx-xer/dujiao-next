import DOMPurify from 'dompurify'
import { filterInstructionStyle } from './fulfillmentDisplay'

// 发货说明专用的 DOMPurify 实例：style 过滤钩子只挂在这里，不影响全局实例的其它净化场景
let instructionsPurifier: ReturnType<typeof DOMPurify> | null = null

const getInstructionsPurifier = () => {
  if (instructionsPurifier) return instructionsPurifier
  const purifier = DOMPurify(window)
  purifier.addHook('uponSanitizeAttribute', (_node, data) => {
    if (data.attrName !== 'style') return
    const filtered = filterInstructionStyle(data.attrValue)
    if (filtered) {
      data.attrValue = filtered
    } else {
      data.keepAttr = false
    }
  })
  instructionsPurifier = purifier
  return purifier
}

/** 净化商品「发货说明」HTML：保留后台编辑器设置的文字颜色 / 对齐等白名单样式，其余样式与危险标签一律剔除 */
export const sanitizeInstructionsHtml = (raw: string) => getInstructionsPurifier().sanitize(raw, {
  ALLOWED_TAGS: ['p', 'br', 'strong', 'em', 'u', 's', 'code', 'pre', 'blockquote', 'ul', 'ol', 'li', 'a', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'span', 'div', 'img', 'hr'],
  ALLOWED_ATTR: ['href', 'target', 'rel', 'src', 'alt', 'title', 'style'],
  FORBID_ATTR: ['class', 'id'],
  ALLOW_DATA_ATTR: false,
  ALLOWED_URI_REGEXP: /^(?:https?:|mailto:|tel:|#|\/(?!\/))/i,
})
