/**
 * 订单交付区的展示辅助：把交付内容拆成逐行卡密，把使用说明开头的标题提出来当区块标题。
 * 纯函数，不依赖 DOM，方便在 node 测试里跑。
 */

/** 交付内容按行拆开，去掉首尾空白和空行（每行一条卡密 / 账号）。 */
export const splitFulfillmentLines = (payload: unknown): string[] => {
  if (typeof payload !== 'string') return []
  return payload
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter((line) => line !== '')
}

/** 自动发货时应发出的卡密条数：订单项里自动交付商品的数量之和（未标交付方式的按自动算）。 */
export const expectedAutoSecretCount = (items: unknown): number => {
  if (!Array.isArray(items)) return 0
  return items.reduce((sum: number, item: any) => {
    const type = String(item?.fulfillment_type || 'auto')
    const quantity = Number(item?.quantity)
    return type === 'auto' && Number.isFinite(quantity) && quantity > 0 ? sum + quantity : sum
  }, 0)
}

/**
 * 交付内容能不能按「一行一条」逐条列出。
 * 自动发货把多条卡密用换行拼成一段，单条卡密本身也可能有多行（账号 + 密码 + 2FA…），
 * 两者在 payload 里无法区分。只有行数正好等于应发条数时，才能确定每行就是一条卡密；
 * 其余情况（多行卡密、人工交付的大段文字）整块原样展示，避免把一条卡密拆成好几条。
 */
export const canListFulfillmentByLine = (input: {
  type: unknown
  lineCount: number
  expectedCount: number
  structured: boolean
}): boolean => {
  if (input.structured) return true
  if (input.lineCount <= 1) return true
  return input.type === 'auto' && input.expectedCount > 0 && input.lineCount === input.expectedCount
}

const NAMED_ENTITIES: Record<string, string> = {
  amp: '&',
  lt: '<',
  gt: '>',
  quot: '"',
  apos: "'",
  nbsp: ' ',
}

const decodeEntities = (text: string): string =>
  text.replace(/&(#x[0-9a-f]+|#\d+|[a-z]+);/gi, (whole, code: string) => {
    if (code[0] === '#') {
      const value = code[1] === 'x' || code[1] === 'X' ? parseInt(code.slice(2), 16) : parseInt(code.slice(1), 10)
      return Number.isFinite(value) && value > 0 && value <= 0x10ffff ? String.fromCodePoint(value) : whole
    }
    return NAMED_ENTITIES[code.toLowerCase()] ?? whole
  })

// 超过这个长度的开头标题更像正文，不提出来
const MAX_HEADING_LENGTH = 40

/**
 * 站长写的使用说明常以「<h2>使用说明</h2>」开头，和区块自带的标题重复。
 * 开头是 h1–h4 时把它的纯文本提出来作为区块标题，正文里去掉这一段；否则原样返回、heading 为空。
 * 传入的 html 应已经过净化。
 *
 * 安全：这里是在净化之后按正则切字符串，切出来的 body 会直接进 v-html。
 * 标题段里只要出现属性（= 或引号）就不拆：部分浏览器序列化属性值时不转义 < >，
 * 属性里藏一段「</h2><img onerror=…>」会让正则切到属性中间，把后半段未净化的标签放出来。
 * 不带属性时，文本里的 < > 一定被转义成实体，匹配到的 </hN> 只可能是真正的闭合标签。
 */
export const splitLeadingHeading = (html: string): { heading: string; body: string } => {
  const match = /^\s*<h([1-4])(?:\s[^>]*)?>([\s\S]*?)<\/h\1\s*>/i.exec(html)
  if (!match || /[="']/.test(match[0])) return { heading: '', body: html }
  const heading = decodeEntities((match[2] ?? '').replace(/<[^>]*>/g, ''))
    .replace(/\s+/g, ' ')
    .trim()
  if (!heading || heading.length > MAX_HEADING_LENGTH) return { heading: '', body: html }
  return { heading, body: html.slice(match[0].length).trim() }
}
