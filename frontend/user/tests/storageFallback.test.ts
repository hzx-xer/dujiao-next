import test from 'node:test'
import assert from 'node:assert/strict'

import { createMemoryStorage, ensureUsableStorage } from '../src/utils/storageFallback.ts'

test('memory storage behaves like Storage', () => {
  const storage = createMemoryStorage()
  assert.equal(storage.getItem('missing'), null)
  storage.setItem('a', '1')
  storage.setItem('b', '2')
  assert.equal(storage.getItem('a'), '1')
  assert.equal(storage.length, 2)
  assert.equal(storage.key(1), 'b')
  storage.removeItem('a')
  assert.equal(storage.getItem('a'), null)
  storage.clear()
  assert.equal(storage.length, 0)
})

test('working storage is left untouched so saved data survives', () => {
  const real = createMemoryStorage()
  real.setItem('saved', 'yes')
  const target: Record<string, any> = { localStorage: real }
  assert.equal(ensureUsableStorage(target, 'localStorage'), false)
  assert.equal(target.localStorage, real)
  assert.equal(target.localStorage.getItem('saved'), 'yes')
})

test('storage access that throws is replaced', () => {
  const target: Record<string, any> = {}
  Object.defineProperty(target, 'localStorage', {
    configurable: true,
    get() {
      throw new Error('SecurityError')
    },
  })
  assert.equal(ensureUsableStorage(target, 'localStorage'), true)
  target.localStorage.setItem('k', 'v')
  assert.equal(target.localStorage.getItem('k'), 'v')
})

test('missing or null storage is replaced', () => {
  const target: Record<string, any> = { localStorage: null }
  assert.equal(ensureUsableStorage(target, 'localStorage'), true)
  assert.equal(target.localStorage.getItem('k'), null)
  const empty: Record<string, any> = {}
  assert.equal(ensureUsableStorage(empty, 'sessionStorage'), true)
  assert.equal(empty.sessionStorage.length, 0)
})

test('a storage whose getItem throws is replaced', () => {
  const target: Record<string, any> = {
    localStorage: {
      getItem() {
        throw new Error('denied')
      },
    },
  }
  assert.equal(ensureUsableStorage(target, 'localStorage'), true)
  assert.equal(target.localStorage.getItem('k'), null)
})

test('a property that cannot be replaced does not throw', () => {
  const target: Record<string, any> = {}
  Object.defineProperty(target, 'localStorage', {
    configurable: false,
    get() {
      throw new Error('SecurityError')
    },
  })
  assert.equal(ensureUsableStorage(target, 'localStorage'), false)
})
