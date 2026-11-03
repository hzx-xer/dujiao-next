import test from 'node:test'
import assert from 'node:assert/strict'

import {
  formatCryptoCountdown,
  hasCryptoPaymentTarget,
  isWalletScannableQr,
  resolveCryptoCountdownTone,
  resolveCryptoNetworkLabel,
  resolveEarliestExpiryMs,
  splitCryptoAddress,
} from '../src/utils/cryptoPaymentDisplay.ts'

test('network label shows the token standard for tokens and only the chain for native coins', () => {
  assert.equal(resolveCryptoNetworkLabel('tron', 'USDT'), 'TRON (TRC20)')
  assert.equal(resolveCryptoNetworkLabel('tron', 'TRX'), 'TRON')
  assert.equal(resolveCryptoNetworkLabel('ethereum', 'USDC'), 'Ethereum (ERC20)')
  assert.equal(resolveCryptoNetworkLabel('ethereum', 'eth'), 'Ethereum')
  assert.equal(resolveCryptoNetworkLabel('bsc', 'USDT'), 'BNB Smart Chain (BEP20)')
  assert.equal(resolveCryptoNetworkLabel('bsc', 'BNB'), 'BNB Smart Chain')
})

test('network label accepts aliases and never invents a network', () => {
  assert.equal(resolveCryptoNetworkLabel(' TRC20 ', 'USDT'), 'TRON (TRC20)')
  assert.equal(resolveCryptoNetworkLabel('erc20', 'USDT'), 'Ethereum (ERC20)')
  assert.equal(resolveCryptoNetworkLabel('polygon', 'USDT'), 'Polygon')
  assert.equal(resolveCryptoNetworkLabel('newchain', 'USDT'), 'NEWCHAIN')
  assert.equal(resolveCryptoNetworkLabel('', 'USDT'), '')
})

test('address split keeps every character in order', () => {
  const address = 'TPeSRfgaRpkUHhjLBiYjKSsHUA7j4mUeTV'
  const parts = splitCryptoAddress(address)
  assert.equal(parts.head, 'TPeSRf')
  assert.equal(parts.tail, 'j4mUeTV'.slice(-6))
  assert.equal(parts.head + parts.middle + parts.tail, address)
})

test('short address is not split', () => {
  assert.deepEqual(splitCryptoAddress('TAddr'), { head: 'TAddr', middle: '', tail: '' })
  assert.deepEqual(splitCryptoAddress(''), { head: '', middle: '', tail: '' })
})

test('earliest expiry wins so the page never promises more time than the gateway gives', () => {
  const gateway = '2026-11-02T12:20:00Z'
  const order = '2026-11-02T12:30:00Z'
  assert.equal(resolveEarliestExpiryMs(gateway, order), new Date(gateway).getTime())
  assert.equal(resolveEarliestExpiryMs(order, gateway), new Date(gateway).getTime())
})

test('earliest expiry ignores missing and invalid values', () => {
  const order = '2026-11-02T12:30:00Z'
  assert.equal(resolveEarliestExpiryMs(undefined, order), new Date(order).getTime())
  assert.equal(resolveEarliestExpiryMs(null, '', 'not-a-date', order), new Date(order).getTime())
  assert.equal(resolveEarliestExpiryMs(undefined, null, ''), null)
})

test('countdown tone turns urgent in the last three minutes and expired at zero', () => {
  assert.equal(resolveCryptoCountdownTone(null), 'normal')
  assert.equal(resolveCryptoCountdownTone(3 * 60 * 1000 + 1), 'normal')
  assert.equal(resolveCryptoCountdownTone(3 * 60 * 1000), 'urgent')
  assert.equal(resolveCryptoCountdownTone(1), 'urgent')
  assert.equal(resolveCryptoCountdownTone(0), 'expired')
  assert.equal(resolveCryptoCountdownTone(-1000), 'expired')
})

test('countdown text pads minutes and seconds and shows hours when needed', () => {
  assert.equal(formatCryptoCountdown(455_000), '07:35')
  assert.equal(formatCryptoCountdown(999), '00:00')
  assert.equal(formatCryptoCountdown(-5000), '00:00')
  assert.equal(formatCryptoCountdown(3_661_000), '1:01:01')
})

test('web links are not treated as wallet scannable', () => {
  assert.equal(isWalletScannableQr('TPeSRfgaRpkUHhjLBiYjKSsHUA7j4mUeTV'), true)
  assert.equal(isWalletScannableQr('https://gateway.example/pay/checkout-counter/T1'), false)
  assert.equal(isWalletScannableQr('HTTP://gateway.example/pay'), false)
  assert.equal(isWalletScannableQr(''), false)
})

test('full payment view requires both the address and the chain amount', () => {
  assert.equal(hasCryptoPaymentTarget('TAddr', '2.03'), true)
  assert.equal(hasCryptoPaymentTarget('TAddr', ''), false)
  assert.equal(hasCryptoPaymentTarget(' ', '2.03'), false)
})
