import test from 'node:test'
import assert from 'node:assert/strict'

import { toMetaDescription } from '../src/utils/metaDescription.ts'

test('plain text is kept', () => {
  assert.equal(toMetaDescription('命运方舟 点卡 50 元'), '命运方舟 点卡 50 元')
})

test('tags are removed and whitespace is collapsed', () => {
  assert.equal(toMetaDescription('<p>第一行</p>\n\n<p>第二&nbsp;行 &amp; 说明</p>'), '第一行 第二 行 & 说明')
})

test('script and style content is dropped entirely', () => {
  assert.equal(toMetaDescription('<script>alert(1)</script><style>.a{}</style>正文'), '正文')
})

test('empty and non-string values give an empty string', () => {
  assert.equal(toMetaDescription(''), '')
  assert.equal(toMetaDescription('   '), '')
  assert.equal(toMetaDescription(undefined), '')
  assert.equal(toMetaDescription({ 'zh-CN': 'x' }), '')
})

test('long text is cut by character without splitting emoji', () => {
  const result = toMetaDescription('😀'.repeat(200), 10)
  assert.equal(Array.from(result).length, 10)
  assert.ok(result.endsWith('…'))
  assert.equal(Array.from(result)[8], '😀')
})

test('hostile long input stays fast and bounded', () => {
  for (const input of ['<'.repeat(500000), '<script'.repeat(70000), '<a '.repeat(170000)]) {
    const start = performance.now()
    const result = toMetaDescription(input)
    assert.ok(performance.now() - start < 300, 'should not block the page')
    assert.ok(Array.from(result).length <= 160)
  }
})

test('text exactly at the limit is not truncated', () => {
  assert.equal(toMetaDescription('a'.repeat(160)), 'a'.repeat(160))
})
