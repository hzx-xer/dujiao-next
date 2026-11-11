import test from 'node:test'
import assert from 'node:assert/strict'
import {
  canListFulfillmentByLine,
  expectedAutoSecretCount,
  splitFulfillmentLines,
  splitLeadingHeading,
} from '../src/utils/fulfillmentDisplay.ts'

test('expected auto secret count sums quantities of auto items only', () => {
  assert.equal(expectedAutoSecretCount([{ quantity: 2, fulfillment_type: 'auto' }, { quantity: 3 }, { quantity: 1, fulfillment_type: 'manual' }]), 5)
  assert.equal(expectedAutoSecretCount(undefined), 0)
  assert.equal(expectedAutoSecretCount([{ quantity: 'x' }]), 0)
})

test('auto payload is listed line by line only when lines match the purchased quantity', () => {
  // 买 3 件、3 行：每行一条卡密
  assert.equal(canListFulfillmentByLine({ type: 'auto', lineCount: 3, expectedCount: 3, structured: false }), true)
  // 买 2 件、12 行：卡密本身是多行，整块展示
  assert.equal(canListFulfillmentByLine({ type: 'auto', lineCount: 12, expectedCount: 2, structured: false }), false)
  // 买 1 件、6 行
  assert.equal(canListFulfillmentByLine({ type: 'auto', lineCount: 6, expectedCount: 1, structured: false }), false)
  // 单行总是逐条
  assert.equal(canListFulfillmentByLine({ type: 'manual', lineCount: 1, expectedCount: 0, structured: false }), true)
  // 人工交付的多行文字整块展示，结构化键值逐条
  assert.equal(canListFulfillmentByLine({ type: 'manual', lineCount: 4, expectedCount: 0, structured: false }), false)
  assert.equal(canListFulfillmentByLine({ type: 'manual', lineCount: 4, expectedCount: 0, structured: true }), true)
})

test('fulfillment payload splits into trimmed non-empty lines', () => {
  assert.deepEqual(splitFulfillmentLines('  CODE-1 \r\n\nCODE-2\n   \n'), ['CODE-1', 'CODE-2'])
  assert.deepEqual(splitFulfillmentLines(''), [])
  assert.deepEqual(splitFulfillmentLines(null), [])
})

test('leading heading of instructions becomes the block title', () => {
  assert.deepEqual(splitLeadingHeading('<h2>使用说明</h2><ol><li>登录</li></ol>'), {
    heading: '使用说明',
    body: '<ol><li>登录</li></ol>',
  })
  assert.deepEqual(splitLeadingHeading('\n<h3 class="x"><b>How</b> to&nbsp;use &amp; tips</h3>\n<p>a</p>'), {
    heading: 'How to use & tips',
    body: '<p>a</p>',
  })
})

test('instructions without a short leading heading stay unchanged', () => {
  const plain = '<p>先登录</p><h2>步骤</h2>'
  assert.deepEqual(splitLeadingHeading(plain), { heading: '', body: plain })
  const empty = '<h2> </h2><p>a</p>'
  assert.deepEqual(splitLeadingHeading(empty), { heading: '', body: empty })
  const long = `<h2>${'很长'.repeat(30)}</h2><p>a</p>`
  assert.deepEqual(splitLeadingHeading(long), { heading: '', body: long })
  // h5/h6 和不配对的标签不算
  const h5 = '<h5>小标题</h5><p>a</p>'
  assert.deepEqual(splitLeadingHeading(h5), { heading: '', body: h5 })
})
