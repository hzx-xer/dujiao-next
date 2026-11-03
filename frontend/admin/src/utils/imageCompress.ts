/**
 * 上传前把较大的 PNG / JPEG 转成 WebP，并限制最长边。在浏览器里完成，不需要后端新增依赖。
 * 任何一步失败或没有变小，都原样上传原图，压缩不应该让上传失败。
 */

const COMPRESSIBLE_TYPES = new Set(['image/png', 'image/jpeg'])
const MIN_COMPRESS_BYTES = 200 * 1024
const MAX_DIMENSION = 2400
const WEBP_QUALITY = 0.85
const MIN_SAVING_RATIO = 0.9

export const shouldCompressImage = (type: string, size: number): boolean => {
  return COMPRESSIBLE_TYPES.has(type.toLowerCase()) && size >= MIN_COMPRESS_BYTES
}

export const fitWithinMaxDimension = (
  width: number,
  height: number,
  maxDimension = MAX_DIMENSION,
): { width: number; height: number } => {
  const longest = Math.max(width, height)
  if (longest <= maxDimension) return { width, height }
  const scale = maxDimension / longest
  return {
    width: Math.max(1, Math.round(width * scale)),
    height: Math.max(1, Math.round(height * scale)),
  }
}

export const toWebpFileName = (name: string): string => {
  const base = name.replace(/\.[^./\\]+$/, '')
  return `${base || 'image'}.webp`
}

export const isWorthReplacing = (originalSize: number, compressedSize: number): boolean => {
  return compressedSize > 0 && compressedSize < originalSize * MIN_SAVING_RATIO
}

const loadImage = (file: File): Promise<HTMLImageElement> => new Promise((resolve, reject) => {
  const url = URL.createObjectURL(file)
  const image = new Image()
  image.onload = () => {
    URL.revokeObjectURL(url)
    resolve(image)
  }
  image.onerror = () => {
    URL.revokeObjectURL(url)
    reject(new Error('image decode failed'))
  }
  image.src = url
})

const encodeWebp = (canvas: HTMLCanvasElement): Promise<Blob | null> => new Promise((resolve) => {
  canvas.toBlob(resolve, 'image/webp', WEBP_QUALITY)
})

export const compressImageForUpload = async (file: File): Promise<File> => {
  if (!shouldCompressImage(file.type, file.size)) return file
  try {
    const image = await loadImage(file)
    const { width, height } = fitWithinMaxDimension(image.naturalWidth, image.naturalHeight)
    const canvas = document.createElement('canvas')
    canvas.width = width
    canvas.height = height
    const context = canvas.getContext('2d')
    if (!context) return file
    context.drawImage(image, 0, 0, width, height)
    const blob = await encodeWebp(canvas)
    // 部分浏览器不支持编码 WebP，toBlob 会静默退回 PNG，这时必须保留原图
    if (!blob || blob.type !== 'image/webp' || !isWorthReplacing(file.size, blob.size)) return file
    return new File([blob], toWebpFileName(file.name), { type: 'image/webp', lastModified: file.lastModified })
  } catch {
    return file
  }
}
