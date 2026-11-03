import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

const read = (relative: string) => readFileSync(fileURLToPath(new URL(`../src/${relative}`, import.meta.url)), 'utf8')

const views = ['views/Payment.vue', 'templates/luna/Payment.vue', 'templates/vault/Payment.vue']

// 已支付的订单如果落到默认视图，会在跳转详情页之前再展示一遍支付方式列表
for (const view of views) {
  test(`${view} shows the completed state before the result and default views`, () => {
    const source = read(view)
    const completed = source.indexOf('v-else-if="paymentCompleted"')
    const result = source.indexOf('v-else-if="showResultView"')
    assert.ok(completed > 0, 'should render a payment completed state')
    assert.ok(completed < result, 'completed state must be checked before the result view')
    assert.match(source.slice(completed, result), /payment\.redirecting/)
    assert.match(source.slice(source.indexOf('<script')), /\bpaymentCompleted\b/)
  })
}

test('paymentCompleted covers the same statuses that trigger the redirect', () => {
  assert.match(read('composables/usePayment.ts'), /const paymentCompleted = computed\(\(\) => shouldRedirect\(order\.value\?\.status\)\)/)
})
