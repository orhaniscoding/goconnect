import QRCode from 'qrcode'

/**
 * Simple QR Code generator for WireGuard configs
 * Uses HTML Canvas API for client-side generation
 */

/**
 * Generate QR code as Data URL
 * @param text - Text to encode in QR code
 * @param size - Size of the QR code (default: 256)
 * @returns Promise<string> - Data URL of the QR code image
 */
export async function generateQRCode(text: string, size: number = 256): Promise<string> {
  try {
    return await QRCode.toDataURL(text, { width: size, margin: 1 })
  } catch (err: unknown) {
    console.error('QR Code generation failed:', err)
    // Fallback to empty string or error placeholder
    return ''
  }
}

/**
 * Generate QR code on canvas element
 * @param text - Text to encode
 * @param canvasElement - Canvas element to draw on
 */
export function generateQRCodeOnCanvas(text: string, canvasElement: HTMLCanvasElement): void {
  QRCode.toCanvas(canvasElement, text, { width: 256, margin: 1 }, (error?: Error | null) => {
    if (error) console.error('QR Code canvas generation failed:', error)
  })
}

/**
 * Download QR code as image file
 * @param qrDataUrl - Data URL of the QR code
 * @param filename - Filename for download
 */
export function downloadQRCode(qrDataUrl: string, filename: string): void {
  const a = document.createElement('a')
  a.href = qrDataUrl
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}
