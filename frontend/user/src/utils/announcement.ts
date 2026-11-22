import DOMPurify from 'dompurify'
import { processHtmlForDisplay } from './content'

// 首页公告富文本白名单：弹窗与横栏共用，保证两处渲染结果一致
export const sanitizeAnnouncementHtml = (raw: unknown): string => {
  const withImages = processHtmlForDisplay(String(raw || ''))
  return DOMPurify.sanitize(withImages, {
    ALLOWED_TAGS: ['p', 'br', 'strong', 'em', 'u', 's', 'code', 'pre', 'blockquote', 'ul', 'ol', 'li', 'a', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'span', 'div', 'img', 'hr', 'table', 'thead', 'tbody', 'tr', 'td', 'th', 'colgroup', 'col'],
    ALLOWED_ATTR: ['href', 'target', 'rel', 'src', 'alt', 'title', 'style', 'colspan', 'rowspan', 'width'],
    ALLOW_DATA_ATTR: false,
    ALLOWED_URI_REGEXP: /^(?:https?:|mailto:|tel:|#|\/(?!\/))/i,
  })
}
