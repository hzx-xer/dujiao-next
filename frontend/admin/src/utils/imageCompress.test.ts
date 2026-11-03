import { describe, expect, it } from 'vitest'

import {
  fitWithinMaxDimension,
  isWorthReplacing,
  shouldCompressImage,
  toWebpFileName,
} from './imageCompress'

describe('shouldCompressImage', () => {
  const large = 500 * 1024

  it('compresses only large png and jpeg files', () => {
    expect(shouldCompressImage('image/png', large)).toBe(true)
    expect(shouldCompressImage('image/jpeg', large)).toBe(true)
    expect(shouldCompressImage('IMAGE/PNG', large)).toBe(true)
    expect(shouldCompressImage('image/png', 200 * 1024)).toBe(true)
    expect(shouldCompressImage('image/png', 199 * 1024)).toBe(false)
  })

  it('leaves animated, vector and already compressed formats alone', () => {
    expect(shouldCompressImage('image/gif', large)).toBe(false)
    expect(shouldCompressImage('image/svg+xml', large)).toBe(false)
    expect(shouldCompressImage('image/webp', large)).toBe(false)
  })
})

describe('fitWithinMaxDimension', () => {
  it('keeps images within the limit', () => {
    expect(fitWithinMaxDimension(1200, 800)).toEqual({ width: 1200, height: 800 })
    expect(fitWithinMaxDimension(2400, 2400)).toEqual({ width: 2400, height: 2400 })
  })

  it('scales larger images keeping the aspect ratio', () => {
    expect(fitWithinMaxDimension(4800, 2400)).toEqual({ width: 2400, height: 1200 })
    expect(fitWithinMaxDimension(1000, 4000)).toEqual({ width: 600, height: 2400 })
  })

  it('never produces a zero dimension', () => {
    expect(fitWithinMaxDimension(10000, 1)).toEqual({ width: 2400, height: 1 })
  })
})

describe('toWebpFileName', () => {
  it('replaces the extension', () => {
    expect(toWebpFileName('商品图.png')).toBe('商品图.webp')
    expect(toWebpFileName('a.b.JPEG')).toBe('a.b.webp')
    expect(toWebpFileName('no-extension')).toBe('no-extension.webp')
    expect(toWebpFileName('.png')).toBe('image.webp')
  })
})

describe('isWorthReplacing', () => {
  it('requires at least a 10 percent saving', () => {
    expect(isWorthReplacing(1000, 500)).toBe(true)
    expect(isWorthReplacing(1000, 899)).toBe(true)
    expect(isWorthReplacing(1000, 900)).toBe(false)
    expect(isWorthReplacing(1000, 1200)).toBe(false)
    expect(isWorthReplacing(1000, 0)).toBe(false)
  })
})
