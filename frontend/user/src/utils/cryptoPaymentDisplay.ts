/** 加密货币付款页的展示规则。只负责显示，不参与金额和支付状态判断。 */

interface CryptoNetwork {
  name: string
  /** 代币标准，例如 TRC20。链的原生币不是代币，不显示标准 */
  standard?: string
  nativeTokens?: string[]
}

const CRYPTO_NETWORKS: Record<string, CryptoNetwork> = {
  tron: { name: 'TRON', standard: 'TRC20', nativeTokens: ['TRX'] },
  ethereum: { name: 'Ethereum', standard: 'ERC20', nativeTokens: ['ETH'] },
  bsc: { name: 'BNB Smart Chain', standard: 'BEP20', nativeTokens: ['BNB'] },
  polygon: { name: 'Polygon' },
  arbitrum: { name: 'Arbitrum One' },
  optimism: { name: 'Optimism' },
  base: { name: 'Base' },
  solana: { name: 'Solana' },
  aptos: { name: 'Aptos' },
  xlayer: { name: 'X Layer' },
}

const CRYPTO_NETWORK_ALIASES: Record<string, string> = {
  trc20: 'tron',
  erc20: 'ethereum',
  eth: 'ethereum',
  bep20: 'bsc',
  'bnb smart chain': 'bsc',
  op: 'optimism',
  arb: 'arbitrum',
}

/** 网络显示名：代币显示「TRON (TRC20)」，原生币只显示「TRON」，未收录的网络原样显示 */
export const resolveCryptoNetworkLabel = (chain: string, token: string): string => {
  const raw = String(chain || '').trim()
  if (!raw) return ''
  const key = raw.toLowerCase()
  const network = CRYPTO_NETWORKS[CRYPTO_NETWORK_ALIASES[key] || key]
  if (!network) return raw.toUpperCase()
  const isNative = network.nativeTokens?.includes(String(token || '').trim().toUpperCase())
  return network.standard && !isNative ? `${network.name} (${network.standard})` : network.name
}

/** 把地址拆成首、中、尾三段，方便高亮首尾和钱包里的地址比对。短地址不拆 */
export const splitCryptoAddress = (address: string, edge = 6): { head: string; middle: string; tail: string } => {
  const value = String(address || '').trim()
  if (value.length <= edge * 2) return { head: value, middle: '', tail: '' }
  return { head: value.slice(0, edge), middle: value.slice(edge, -edge), tail: value.slice(-edge) }
}

/** 取多个截止时间里最早的一个。网关可能比订单更早停止监听，必须按较早的时间提示用户 */
export const resolveEarliestExpiryMs = (...candidates: unknown[]): number | null => {
  const times = candidates
    .filter((value) => value !== null && value !== undefined && value !== '')
    .map((value) => new Date(value as string).getTime())
    .filter((value) => Number.isFinite(value))
  return times.length > 0 ? Math.min(...times) : null
}

export type CryptoCountdownTone = 'normal' | 'urgent' | 'expired'

const CRYPTO_URGENT_MS = 3 * 60 * 1000

export const resolveCryptoCountdownTone = (remainingMs: number | null): CryptoCountdownTone => {
  if (remainingMs === null) return 'normal'
  if (remainingMs <= 0) return 'expired'
  return remainingMs <= CRYPTO_URGENT_MS ? 'urgent' : 'normal'
}

export const formatCryptoCountdown = (remainingMs: number): string => {
  const totalSeconds = Math.max(0, Math.floor(remainingMs / 1000))
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60
  const pad = (value: number) => String(value).padStart(2, '0')
  return hours > 0 ? `${hours}:${pad(minutes)}:${pad(seconds)}` : `${pad(minutes)}:${pad(seconds)}`
}

/** 二维码内容是网页链接时，钱包扫出来不能直接转账，不能提示用户用钱包扫码 */
export const isWalletScannableQr = (content: string): boolean => {
  const value = String(content || '').trim()
  return Boolean(value) && !/^https?:\/\//i.test(value)
}

/** 地址和链上数量都有才用整页收款界面；少一项就没法照着付款，要退回通用支付块 */
export const hasCryptoPaymentTarget = (address: string, chainAmount: string): boolean => {
  return Boolean(String(address || '').trim() && String(chainAmount || '').trim())
}
