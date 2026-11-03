import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

const source = readFileSync(fileURLToPath(new URL('../src/composables/useCheckout.ts', import.meta.url)), 'utf8')
const start = source.indexOf('const handleSubmit = async')
const handleSubmit = source.slice(start, source.indexOf('watch(', start))
const pushIndex = handleSubmit.indexOf('router.push(`/pay?')

test('the source store is not cleared before leaving the checkout page', () => {
  assert.ok(pushIndex > 0, 'handleSubmit should redirect to the payment page')
  assert.equal(
    /\b(?:buyNowStore|cartStore)\.clear\(\)/.test(handleSubmit.slice(0, pushIndex)),
    false,
    'clearing before navigation renders the empty cart while the page fades out',
  )
  assert.match(source, /onUnmounted\(\(\) => \{\s*clearSourceOnLeave\?\.\(\)/)
})

test('the source is decided at submit time because the route has already changed when leaving', () => {
  assert.match(handleSubmit.slice(0, pushIndex), /isBuyNowMode\.value \? buyNowStore\.clear : cartStore\.clear/)
})

test('a failed navigation still clears the source store so the order cannot be repeated', () => {
  assert.match(
    handleSubmit.slice(pushIndex),
    /if \(navigationFailure\) \{[\s\S]*?clearSourceOnLeave = null[\s\S]*?clearSource\(\)/,
  )
})

test('submitting stays true after the order is placed to block repeat clicks during the fade', () => {
  assert.match(handleSubmit, /if \(!clearSourceOnLeave\) submitting\.value = false/)
})
